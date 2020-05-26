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
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Chmod("./testdata/impossible.txt", 0000)
	status := m.Run()
	os.Chmod("./testdata/impossible.txt", 0644)
	os.Exit(status)
}

func TestShowFile(t *testing.T) {
	var tests = []struct {
		path   string
		method string
		status int
	}{
		{"/badfile.txt", "GET", 404},
		{"/impossible.txt", "GET", 403},
		{"/file.txt", "GET", 200},
		{"/file.txt", "POST", 405},
		{"/../server.go", "GET", 400},
	}
	server := SetupServer(":8000", "./testdata", 300, "", "/u/", 10<<20)
	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		w := httptest.NewRecorder()
		server.Handler.ServeHTTP(w, req)
		status := w.Code
		if status != test.status {
			t.Errorf("got %d, expected %d", status, test.status)
		}
	}
}

func TestGetIp(t *testing.T) {

	var tests = []struct {
		header string
		value  string
	}{
		{"X-Real-Ip", "1.2.3.4"},
		{"X-Forwarded-For", "1.2.3.5"},
		// ths will return the value of req.Remoteaddr
		{"Does-Not-Matter", "1.2.3.6"},
	}
	for _, test := range tests {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set(test.header, test.value)
		req.RemoteAddr = "1.2.3.6"
		ip := getIp(req)
		if ip != test.value {
			t.Errorf("got %s, expected %s", ip, test.value)
		}
	}
}

func TestRecieveFile(t *testing.T) {
	fpath := "./testdata/htpasswd"
	var tests = []struct {
		method string
		ctype  string
		status int
		user   string
		passwd string
	}{
		{"PUT", UPLOADCONTENTTYPE, 405, "test", "123"},
		{"POST", "application/json", 400, "test", "123"},
		{"POST", UPLOADCONTENTTYPE, 401, "test", "456"},
		{"POST", UPLOADCONTENTTYPE, 201, "test", "123"},
	}

	// https://stackoverflow.com/questions/43904974/
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	server := SetupServer(":8000", "/tmp", 300, fpath, "/u/", 10<<20)
	for _, test := range tests {
		go func() {
			defer writer.Close()
			_, err := writer.CreateFormFile("file", "file.txt")
			if err != nil {
				t.Error(err)
			}
		}()

		req, _ := http.NewRequest(test.method, "/u/", pr)
		req.SetBasicAuth(test.user, test.passwd)
		req.Header.Set("Content-Type", test.ctype+"; boundary="+writer.Boundary())
		w := httptest.NewRecorder()
		server.Handler.ServeHTTP(w, req)
		status := w.Code
		if status != test.status {
			t.Errorf("got %d, expected %d", status, test.status)
		}
	}
}
