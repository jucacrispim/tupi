// Copyright 2020, 2023 Juca Crispim <juca@poraodojuca.net>

// This file is part of tupi.

// tupi is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// tupi is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with tupi. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var UPLOADCONTENTTYPE string = "multipart/form-data"

var rootDir string = "."
var uploadPath string = "/u/"
var extractPath string = "/e/"
var maxUpload int64 = 10 << 20
var maxFileMemory int64 = 10 << 20
var htpasswdFile string = ""
var defaultToIndex bool = false

const indexFile = "index.html"

type HTTPServer struct {
	Conf Config
	// We have one server for each port we are listen
	Servers []*http.Server
}

func (s HTTPServer) Run() {
	startServer := getStartServerFn(s)

	if len(s.Servers) == 1 {
		startServer(s.Servers[0])
	} else {
		server := s.Servers[0]
		for _, serv := range s.Servers[1:] {
			go startServer(serv)
		}
		startServer(server)
	}
}

// SetupServer creates a new instance of the tupi
// http server. You can start it using “HTTPServer.Run“
func SetupServer(conf Config) HTTPServer {

	// read this for new implementation
	// https://github.com/golang/go/issues/35626

	setRootDir(conf.RootDir)
	setHtpasswordFile(conf.HtpasswdFile)
	setUploadPath(conf.UploadPath)
	setExtractPath(conf.ExtractPath)
	setMaxUpload(conf.MaxUploadSize)
	setDefaultToIndex(conf.DefaultToIndex)

	handler := logRequest(http.HandlerFunc(route))
	s := HTTPServer{
		Conf: conf,
	}
	servers := make([]*http.Server, 0)
	addr := fmt.Sprintf("%s:%s", conf.Host, strconv.FormatInt(int64(conf.Port), 10))
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(conf.Timeout) * time.Second,
		WriteTimeout: time.Duration(conf.Timeout) * time.Second,
	}
	servers = append(servers, server)
	s.Servers = servers
	return s
}

type StatusedResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *StatusedResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func setRootDir(rdir string) {
	rootDir = rdir
}

func setUploadPath(upath string) {
	uploadPath = upath
}

func setExtractPath(epath string) {
	extractPath = epath
}

func setMaxUpload(mupload int64) {
	maxUpload = mupload
}

func setHtpasswordFile(fpath string) {
	htpasswdFile = fpath
}

func setDefaultToIndex(def bool) {
	defaultToIndex = def
}

// route is responsible for calling the proper handler based in the
// request path.
func route(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == uploadPath {
		recieveFile(w, req)
	} else if req.URL.Path == extractPath {
		recieveAndExtract(w, req)
	} else {
		showFile(w, req)
	}
}

type requestError struct {
	StatusCode int
	Err        error
}

func (r *requestError) Error() string {
	return fmt.Sprintf("%s", r.Err)
}

func checkUploadRequest(
	w http.ResponseWriter, req *http.Request) (*multipart.Reader, error) {
	err := &requestError{}

	ok := authenticate(req, htpasswdFile)
	if !ok {
		err.StatusCode = http.StatusUnauthorized
		err.Err = errors.New("Unauthorized")
		return nil, err
	}

	if req.Method != "POST" {
		err.StatusCode = http.StatusMethodNotAllowed
		err.Err = errors.New("Method not allowed")
		return nil, err
	}

	if !strings.HasPrefix(req.Header.Get("Content-Type"), UPLOADCONTENTTYPE) {
		msg := "Bad request. Use Content-Type: " + UPLOADCONTENTTYPE
		err.StatusCode = http.StatusBadRequest
		err.Err = errors.New(msg)
		return nil, err
	}

	req.Body = http.MaxBytesReader(w, req.Body, maxUpload)
	reader, mperr := req.MultipartReader()
	if mperr != nil {
		// notest
		err.StatusCode = http.StatusBadRequest
		err.Err = errors.New("Bad request")
		return nil, err
	}
	return reader, nil
}

func recieveFile(w http.ResponseWriter, req *http.Request) {

	reader, err := checkUploadRequest(w, req)
	if err != nil {
		e, _ := err.(*requestError)
		http.Error(w, string(err.Error()), e.StatusCode)
		return
	}
	fname, err := writeFile(rootDir, reader, false)
	if err != nil && err != io.EOF {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fname + "\n"))
}

func recieveAndExtract(w http.ResponseWriter, req *http.Request) {
	reader, err := checkUploadRequest(w, req)
	if err != nil {
		e, _ := err.(*requestError)
		http.Error(w, string(err.Error()), e.StatusCode)
		return
	}

	fname, err := writeFile(rootDir, reader, true)
	if err != nil && err != io.EOF {
		panic(err)
	}
	fpath := filepath.Join(rootDir, fname)

	defer os.RemoveAll(fpath)
	file, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	files, err := extractFiles(file, rootDir)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	for _, f := range files {
		w.Write([]byte(f + "\n"))
	}

}

func showFile(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fpath := req.URL.Path
	if strings.HasSuffix(fpath, "/") && defaultToIndex {
		fpath += indexFile
	}
	path := rootDir + fpath
	http.ServeFile(w, req, path)
}

func logRequest(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		sw := &StatusedResponseWriter{w, http.StatusOK}
		h.ServeHTTP(sw, req)
		remote := getIp(req)
		path := req.URL.Path
		method := req.Method
		ua := req.Header.Get("User-Agent")
		log.Printf("%s %s %s %d %s\n", remote, method, path, sw.status, ua)
	}
	return http.HandlerFunc(handler)
}

func getIp(req *http.Request) string {
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = req.RemoteAddr
	}
	return ip
}

// for tests
type startServerFn func(server *http.Server)

var startServerTestFn startServerFn = nil

func getStartServerFn(s HTTPServer) startServerFn {
	// notest
	if startServerTestFn != nil {
		return startServerTestFn
	}
	startServer := func(server *http.Server) {
		if s.Conf.HasCert() && s.Conf.HasKey() {
			server.ListenAndServeTLS(s.Conf.CertFilePath, s.Conf.KeyFilePath)
		} else {
			server.ListenAndServe()
		}

	}
	return startServer
}
