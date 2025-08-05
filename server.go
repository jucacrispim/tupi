// Copyright 2020, 2023-2025 Juca Crispim <juca@poraodojuca.dev>

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
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const UPLOAD_CONTENT_TYPE = "multipart/form-data"
const indexFile = "index.html"

var config Config
var certsCache map[string]tls.Certificate = make(map[string]tls.Certificate, 0)

// StatusedResponseWriter is a respose writer that holds the status code.
// Used for log purposes.
type StatusedResponseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader writes the status code header for a request and stores the
// status in the writer.
func (w *StatusedResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

type RequestError struct {
	StatusCode int
	Err        error
}

// TupiServer is the struct that holds the config and a slice of the http.Server.
// One http.Server for each port we listen
type TupiServer struct {
	Conf    Config
	Servers []TupiPortServer
}

func (s *TupiServer) LoadPlugins() {
	for domain, conf := range s.Conf.Domains {
		if conf.AuthPlugin != "" {
			err := LoadAuthPlugin(conf.AuthPlugin, domain, &conf.AuthPluginConf)
			if err != nil {
				Errorf("Error loading auth plugin %s", err.Error())
			}
		}

		if conf.ServePlugin != "" {
			err := LoadServePlugin(conf.ServePlugin, domain, &conf.ServePluginConf)
			if err != nil {
				Errorf("Error loading serve plugin %s", err.Error())
			}
		}
	}
}

func (s *TupiServer) Run() {
	wg := new(sync.WaitGroup)
	for _, serv := range s.Servers {
		serv := serv
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := serv.Run()
			if err != nil {
				// notest
				Errorf("server on %s failed: %s", serv.Server.Addr, err.Error())
			}
		}()
	}
	wg.Wait()
}

// SetupServer creates a new instance of the tupi
// http server. You can start it using “TupiServer.Run“
func SetupServer(conf Config) TupiServer {

	// read this for new implementation
	// https://github.com/golang/go/issues/35626

	setConfig(conf)
	loglevel := conf.Domains["default"].LogLevel
	SetLogLevelStr(loglevel)
	handler := logRequest(http.HandlerFunc(route))
	s := TupiServer{
		Conf: conf,
	}
	servers := make([]TupiPortServer, 0)
	host := conf.Domains["default"].Host
	timeout := conf.Domains["default"].Timeout

	portsConf := conf.GetPortsConfig()
	for _, portConf := range portsConf {
		addr := fmt.Sprintf(
			"%s:%s",
			host,
			strconv.FormatInt(int64(portConf.Port), 10))
		Debugf("new server config: %s", addr)
		server := &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  time.Duration(timeout) * time.Second,
			WriteTimeout: time.Duration(timeout) * time.Second,
		}

		portServer := TupiPortServer{
			Server: server,
			UseSSL: portConf.UseSSL,
		}
		servers = append(servers, portServer)
	}

	s.Servers = servers
	s.LoadPlugins()
	return s
}

type startServerFn func(server *http.Server, use_ssl bool) error

type TupiPortServer struct {
	Server *http.Server
	UseSSL bool
}

func (s *TupiPortServer) Run() error {
	startFn := getStartServerFn()
	return startFn(s.Server, s.UseSSL)
}

// Call the default tupi actions or a pluging based
// in the domain config
func route(w http.ResponseWriter, req *http.Request) {
	c := getConfigForRequest(req)
	if shouldAuthenticate(req, c) {
		ok, status := authenticate(req, c)
		if !ok {
			if c.AuthPlugin == "" {
				w.Header().Set("WWW-Authenticate", "Basic realm=xZsd234-1M82sa")
			}
			http.Error(w, "Bad auth", status)
			return
		}
	}
	if c.ServePlugin == "" {
		serveDefaultTupi(w, req, c)
		return
	}
	wr := w.(*StatusedResponseWriter)
	servePlugin(wr.ResponseWriter, req, c)
}

// Does the default tupi actions, serve and receive files.
func serveDefaultTupi(w http.ResponseWriter, req *http.Request, c *DomainConfig) {
	if req.URL.Path == c.UploadPath {
		recieveFile(w, req, c)
	} else if req.URL.Path == c.ExtractPath {
		recieveAndExtract(w, req, c)
	} else {
		showFile(w, req, c)
	}
}

func servePlugin(w http.ResponseWriter, req *http.Request, c *DomainConfig) {
	fn, err := GetServePlugin(c.ServePlugin)
	if err != nil {
		// notest
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fn(w, req, &c.ServePluginConf)
}

func recieveFile(w http.ResponseWriter, req *http.Request, c *DomainConfig) {

	reader, err := checkUploadRequest(w, req, c)
	if err != nil {
		e, _ := err.(*RequestError)
		http.Error(w, string(err.Error()), e.StatusCode)
		return
	}
	fname, err := writeFile(c.RootDir, reader, false, c.PreventOverwrite)
	if err != nil && err != io.EOF {
		if isBadRequest(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// notest
		Errorf("%s\n", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fname + "\n"))
}

func recieveAndExtract(w http.ResponseWriter, req *http.Request, c *DomainConfig) {
	reader, err := checkUploadRequest(w, req, c)
	if err != nil {
		e, _ := err.(*RequestError)
		http.Error(w, string(err.Error()), e.StatusCode)
		return
	}
	f, err := getFileFromRequest(reader)
	if err != nil {
		// notest
		Errorf("%s\n", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	freader := bytes.NewBuffer(f.content)
	files, err := extractFiles(freader, c.RootDir, c.PreventOverwrite)
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

func showFile(w http.ResponseWriter, req *http.Request, c *DomainConfig) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if containsDotDot(req.URL.Path) {
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}

	fpath := req.URL.Path
	if strings.HasSuffix(fpath, "/") && c.DefaultToIndex {
		fpath += indexFile
	}
	path := c.RootDir + fpath
	dir, file := filepath.Split(path)
	serveFile(w, req, http.Dir(dir), file)
}

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

func setConfig(conf Config) {
	config = conf
}

func getDomainForRequest(req *http.Request) string {
	domain := strings.Split(req.Host, ":")[0]
	domain = strings.ToLower(domain)
	return domain
}

func getPortForRequest(req *http.Request) (int, error) {
	if host := req.Host; host != "" {
		if _, port, err := net.SplitHostPort(host); err == nil {
			p, err := strconv.Atoi(port)
			return p, err
		}
	}
	if addr, ok := req.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
		if _, port, err := net.SplitHostPort(addr.String()); err == nil {
			p, err := strconv.Atoi(port)
			return p, err
		}
	}
	if req.TLS != nil || req.URL != nil && req.URL.Scheme == "https" {
		return 443, nil
	}
	return 80, nil
}

func getConfigForRequest(req *http.Request) *DomainConfig {
	domain := getDomainForRequest(req)
	port, err := getPortForRequest(req)
	default_confg := config.Domains["default"]
	if err != nil {
		// notest
		Errorf("could not get port for request: %s", err.Error())
		return &default_confg

	}
	if conf, exists := config.Domains[domain]; exists {
		if conf.HasPortConf(port) {
			return &conf
		}
	}
	return &default_confg
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("%s", r.Err)
}

func shouldAuthenticate(req *http.Request, c *DomainConfig) bool {
	for _, meth := range c.AuthMethods {
		if strings.ToUpper(meth) == strings.ToUpper(req.Method) {
			return true
		}
	}
	return false
}

func checkUploadRequest(
	w http.ResponseWriter, req *http.Request,
	c *DomainConfig) (*multipart.Reader, error) {
	err := &RequestError{}

	if req.Method != "POST" {
		err.StatusCode = http.StatusMethodNotAllowed
		err.Err = errors.New("Method not allowed")
		return nil, err
	}

	ctype := req.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, UPLOAD_CONTENT_TYPE) {
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

func logRequest(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		sw := &StatusedResponseWriter{w, http.StatusOK}
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

func isBadRequest(err error) bool {
	msg := err.Error()
	return msg == INVALID_PREFIX_MSG || strings.Contains(msg, "already exists")
}

// for tests

var startServerTestFn startServerFn = nil

func getStartServerFn() startServerFn {
	// notest
	if startServerTestFn != nil {
		return startServerTestFn
	}
	startServer := func(server *http.Server, use_ssl bool) error {
		if use_ssl {
			if server.TLSConfig == nil {
				server.TLSConfig = &tls.Config{}
			}
			tls_conf := server.TLSConfig
			tls_conf.GetCertificate = getCertificate
			return server.ListenAndServeTLS("", "")

		} else {
			return server.ListenAndServe()
		}

	}
	return startServer
}
