// Copyright 2020 Juca Crispim <juca@poraodojuca.net>

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
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var UPLOADCONTENTTYPE string = "multipart/form-data"

var rootDir string = "."
var uploadPath string = "/u/"
var maxUpload int64 = 10 << 20
var maxFileMemory int64 = 10 << 20
var htpasswdFile string = ""

type statusedResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusedResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func setRootDir(rdir string) {
	rootDir = rdir
}

func setUploadPath(upath string) {
	uploadPath = upath
}

func setMaxUpload(mupload int64) {
	maxUpload = mupload
}

func setMaxFileMemory(mfmemory int64) {
	maxFileMemory = mfmemory
}

func setHtpasswordFile(fpath string) {
	htpasswdFile = fpath
}

// route is responsible for calling the proper handler based in the
// request path.
func route(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == uploadPath {
		recieveFile(w, req)
	} else {
		showFile(w, req)
	}
}

func recieveFile(w http.ResponseWriter, req *http.Request) {
	ok := authenticate(req, htpasswdFile)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(req.Header.Get("Content-Type"), UPLOADCONTENTTYPE) {
		msg := "Bad request. Use Content-Type: " + UPLOADCONTENTTYPE
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	req.Body = http.MaxBytesReader(w, req.Body, maxUpload)
	reader, err := req.MultipartReader()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	fname, err := writeFile(rootDir, reader)
	if err != nil && err != io.EOF {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fname + "\n"))
}

func showFile(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := rootDir + req.URL.Path
	http.ServeFile(w, req, path)
}

func logRequest(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		sw := &statusedResponseWriter{w, http.StatusOK}
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

// SetupServer creates a new instance of the tupi
// http server. You can start it using `ListenAndServe`
func SetupServer(
	addr string, rdir string, timeout int, htpasswd string) *http.Server {
	// read this for new implementation
	// https://github.com/golang/go/issues/35626
	setRootDir(rdir)
	setHtpasswordFile(htpasswd)
	handler := logRequest(http.HandlerFunc(route))
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}
	return server
}
