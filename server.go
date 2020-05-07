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
	"log"
	"net/http"
	"time"
)

var rootDir string = "."

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

// ShowFile writes the contents of a file based on the
// request's path. The path is relative to the root dir of
// the application. Only GET requests are allowed.
func showFile(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	if method != "GET" {
		status := http.StatusMethodNotAllowed
		http.Error(w, "Method not allowed", status)
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
func SetupServer(addr string, rdir string, timeout int) *http.Server {
	// read this for new implementation
	// https://github.com/golang/go/issues/35626
	setRootDir(rdir)
	handler := logRequest(http.HandlerFunc(showFile))
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}

	return server
}
