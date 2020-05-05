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
	"io/ioutil"
	"log"
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
		{"/testdata/badfile.txt", "GET", 404},
		{"/testdata/impossible.txt", "GET", 500},
		{"/testdata/file.txt", "GET", 200},
		{"/testdata/file.txt", "POST", 405},
	}
	handler := SetupServer(".")
	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		status := w.Code
		if status != test.status {
			t.Errorf("got %d, expected %d", status, test.status)
		}
	}
}
