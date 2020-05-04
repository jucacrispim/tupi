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

// ShowFile prints the contents of a file based in the
// request's path. The path is relative to the root dir of
// the application. Returns the http status for the request
func showFile(w http.ResponseWriter, req *http.Request) {
	path := rootDir + req.URL.Path
	if !FileExists(path) {
		status := http.StatusNotFound
		http.Error(w, "File not found", status)
		return
	}
	f, err := GetFile(path)

	if err != nil {
		status := http.StatusInternalServerError
		http.Error(w, err.Error(), status)
		return
	}
	status := http.StatusOK
	w.WriteHeader(status)
	w.Write(f.Content)
}

func logRequest(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		sw := &statusedResponseWriter{w, http.StatusOK}
		h.ServeHTTP(sw, req)
		path := req.URL.Path
		log.Printf("%s %d\n", path, sw.status)
	}
	return http.HandlerFunc(handler)
}

// SetupServer creates a new instance of the tupi
// http server. You can start it using `ListenAndServe`
func SetupServer(rdir string) http.Handler {
	setRootDir(rdir)
	handler := logRequest(http.HandlerFunc(showFile))
	http.Handle("/", handler)
	return handler
}
