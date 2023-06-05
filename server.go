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

package tupi

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const UPLOAD_CONTENT_TYPE = "multipart/form-data"
const indexFile = "index.html"

var config Config

type TupiServer struct {
	Conf Config
	// We have one server for each port we listen
	Servers []*http.Server
}

func (s *TupiServer) LoadPlugins() {
	for domain, conf := range s.Conf.Domains {
		if conf.AuthPlugin != "" {
			err := LoadAuthPlugin(conf.AuthPlugin, domain, &conf.AuthPluginConf)
			if err != nil {
				Errorf("Error loading plugin %s", err.Error())
			}
		}
	}
}

func (s *TupiServer) Run() {
	startServer := getStartServerFn(s)
	use_ssl := s.Conf.HasSSL()
	if len(s.Servers) == 1 {
		startServer(s.Servers[0], use_ssl)
	} else {
		server := s.Servers[0]
		for _, serv := range s.Servers[1:] {
			go startServer(serv, use_ssl)
		}
		startServer(server, use_ssl)
	}
}

// SetupServer creates a new instance of the tupi
// http server. You can start it using “HTTPServer.Run“
func SetupServer(conf Config) TupiServer {

	// read this for new implementation
	// https://github.com/golang/go/issues/35626

	setConfig(conf)
	handler := logRequest(http.HandlerFunc(route))
	s := TupiServer{
		Conf: conf,
	}
	servers := make([]*http.Server, 0)
	addr := fmt.Sprintf(
		"%s:%s",
		conf.Domains["default"].Host,
		strconv.FormatInt(int64(conf.Domains["default"].Port), 10))
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(conf.Domains["default"].Timeout) * time.Second,
		WriteTimeout: time.Duration(conf.Domains["default"].Timeout) * time.Second,
	}
	servers = append(servers, server)
	s.Servers = servers
	s.LoadPlugins()
	return s
}

var certsCache map[string]tls.Certificate = make(map[string]tls.Certificate, 0)

// Returns a certificate based on the host config.
func getCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := info.ServerName
	if cert, exists := certsCache[domain]; exists {
		return &cert, nil
	}
	conf, exists := config.Domains[domain]
	if !exists {
		conf = config.Domains["default"]
	}
	AcquireLock(domain)
	defer ReleaseLock(domain)
	// check if the cert was created while waiting for the lock
	if cert, exists := certsCache[domain]; exists {
		// notest
		return &cert, nil
	}
	cert, err := tls.LoadX509KeyPair(conf.CertFilePath, conf.KeyFilePath)
	certsCache[domain] = cert
	return &cert, err
}

type statusedResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusedResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func setConfig(conf Config) {
	config = conf
}

func getDomainForRequest(req *http.Request) string {
	domain := strings.Split(req.Host, ":")[0]
	domain = strings.ToLower(domain)
	return domain
}
func getConfigForRequest(req *http.Request) *DomainConfig {
	domain := getDomainForRequest(req)
	if conf, exists := config.Domains[domain]; exists {
		return &conf
	}
	default_confg := config.Domains["default"]
	return &default_confg
}

// route is responsible for calling the proper handler based in the
// request path.
func route(w http.ResponseWriter, req *http.Request) {
	c := getConfigForRequest(req)
	if req.URL.Path == c.UploadPath {
		recieveFile(w, req)
	} else if req.URL.Path == c.ExtractPath {
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
	c := getConfigForRequest(req)
	ok := authenticate(req, c)
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

	if !strings.HasPrefix(req.Header.Get("Content-Type"), UPLOAD_CONTENT_TYPE) {
		msg := "Bad request. Use Content-Type: " + UPLOAD_CONTENT_TYPE
		err.StatusCode = http.StatusBadRequest
		err.Err = errors.New(msg)
		return nil, err
	}

	req.Body = http.MaxBytesReader(w, req.Body, c.MaxUploadSize)
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
	c := getConfigForRequest(req)
	fname, err := writeFile(c.RootDir, reader, false)
	if err != nil && err != io.EOF {
		// notest
		Errorf("%s\n", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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
	c := getConfigForRequest(req)
	fname, err := writeFile(c.RootDir, reader, true)
	if err != nil && err != io.EOF {
		// notest
		Errorf("%s\n", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fpath := filepath.Join(c.RootDir, fname)

	defer os.RemoveAll(fpath)
	file, err := os.Open(fpath)
	if err != nil {
		// notest
		Errorf("%s\n", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	files, err := extractFiles(file, c.RootDir)
	if err != nil {
		// notest
		Errorf("%s\n", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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
	if containsDotDot(req.URL.Path) {
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}
	c := getConfigForRequest(req)
	fpath := req.URL.Path
	if strings.HasSuffix(fpath, "/") && c.DefaultToIndex {
		fpath += indexFile
	}
	path := c.RootDir + fpath
	dir, file := filepath.Split(path)
	serveFile(w, req, http.Dir(dir), file)
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func logRequest(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		sw := &statusedResponseWriter{w, http.StatusOK}
		h.ServeHTTP(sw, req)
		remote := getIp(req)
		path := req.URL.Path
		method := req.Method
		ua := req.Header.Get("User-Agent")
		Infof("%s %s %s %d %s\n", remote, method, path, sw.status, ua)
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
type startServerFn func(server *http.Server, use_ssl bool)

var startServerTestFn startServerFn = nil

func getStartServerFn(s *TupiServer) startServerFn {
	// notest
	if startServerTestFn != nil {
		return startServerTestFn
	}
	startServer := func(server *http.Server, use_ssl bool) {
		if use_ssl {
			if server.TLSConfig == nil {
				server.TLSConfig = &tls.Config{}
			}
			tls_conf := server.TLSConfig
			tls_conf.GetCertificate = getCertificate
			err := server.ListenAndServeTLS("", "")
			if err != nil {
				panic(err.Error())
			}
		} else {
			err := server.ListenAndServe()
			if err != nil {
				panic(err.Error())
			}
		}

	}
	return startServer
}
