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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestShowFile_SingleFile(t *testing.T) {
	var tests = []struct {
		path   string
		method string
		status int
	}{
		{"/badfile.txt", "GET", 404},
		{"/impossible.txt", "GET", 403},
		{"/file.txt", "GET", 200},
		{"/file.txt", "POST", 405},
		{"/", "GET", 200},
		{"/../server.go", "GET", 400},
		{"/.../file.txt", "GET", 404},
	}
	dconf := DomainConfig{
		Host:           "0.0.0.0",
		Port:           8000,
		RootDir:        "./testdata",
		Timeout:        300,
		HtpasswdFile:   "",
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: true,
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	server := SetupServer(conf)
	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		w := httptest.NewRecorder()
		server.Servers[0].Handler.ServeHTTP(w, req)
		status := w.Code
		if status != test.status {
			t.Errorf("got %d, expected %d", status, test.status)
		}
	}
}

func TestShowFile_ListDir(t *testing.T) {
	var tests = []struct {
		path    string
		method  string
		status  int
		checkFn func(*httptest.ResponseRecorder)
		headers map[string]string
	}{
		{"/", "GET", 200, func(r *httptest.ResponseRecorder) {
			body := string(r.Body.Bytes())
			if strings.Index(body, "index.html") < 0 {
				t.Fatalf("No index.html on list dir")
			}
		}, nil},
		{"", "GET", 301, nil, nil},
		{"/", "GET", 304, nil,
			map[string]string{"If-Modified-Since": time.Now().Add(time.Hour * 3).Format(http.TimeFormat)}},
		{"/", "GET", 200, nil,
			map[string]string{"If-Modified-Since": time.Time{}.Format(http.TimeFormat)}},
		{"/", "GET", 200, nil,
			map[string]string{"If-Modified-Since": "xx"}},
	}
	dconf := DomainConfig{
		Host:           "0.0.0.0",
		Port:           8000,
		RootDir:        "./testdata",
		Timeout:        300,
		HtpasswdFile:   "",
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: false,
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	server := SetupServer(conf)
	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		if test.headers != nil {
			for k, v := range test.headers {
				req.Header.Set(k, v)
			}
		}
		w := httptest.NewRecorder()
		server.Servers[0].Handler.ServeHTTP(w, req)
		status := w.Code
		if status != test.status {
			t.Errorf("got %d, expected %d", status, test.status)
		}
		if test.checkFn != nil {
			test.checkFn(w)
		}
	}
}

func TestShowFile_Authenticated(t *testing.T) {
	fpath := "./testdata/htpasswd"
	var tests = []struct {
		path     string
		status   int
		username string
		password string
	}{
		{"/file.txt", 200, "test", "123"},
		{"/file.txt", 401, "", ""},
	}
	dconf := DomainConfig{
		Host:           "0.0.0.0",
		Port:           8000,
		RootDir:        "./testdata",
		Timeout:        300,
		HtpasswdFile:   fpath,
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: true,
		AuthDownloads:  true,
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	server := SetupServer(conf)
	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.path, nil)
		if test.username != "" && test.password != "" {
			req.SetBasicAuth(test.username, test.password)
		}
		w := httptest.NewRecorder()
		server.Servers[0].Handler.ServeHTTP(w, req)
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
		prefix string
	}{
		{"PUT", UPLOAD_CONTENT_TYPE, 405, "test", "123", ""},
		{"POST", "application/json", 400, "test", "123", ""},
		{"POST", UPLOAD_CONTENT_TYPE, 401, "test", "456", ""},
		{"POST", UPLOAD_CONTENT_TYPE, 201, "test", "123", ""},
		{"POST", UPLOAD_CONTENT_TYPE, 400, "test", "123", "../invalid"},
		{"POST", UPLOAD_CONTENT_TYPE, 201, "test", "123", "good-prefix"},
	}

	rdir := "/tmp/tupitest"
	os.MkdirAll(rdir, 0755)
	defer os.RemoveAll(rdir)

	dconf := DomainConfig{
		Host:             "0.0.0.0",
		Port:             8000,
		RootDir:          rdir,
		Timeout:          300,
		HtpasswdFile:     fpath,
		UploadPath:       "/u/",
		ExtractPath:      "/e/",
		MaxUploadSize:    10 << 20,
		DefaultToIndex:   true,
		PreventOverwrite: true,
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	server := SetupServer(conf)

	for _, test := range tests {
		pr, boundary, err := createBufferMultipartReader("file.txt", "test", test.prefix)
		if err != nil {
			t.Errorf("error creating reader")
		}

		req, _ := http.NewRequest(test.method, "/u/", pr)
		req.SetBasicAuth(test.user, test.passwd)
		req.Header.Set("Content-Type", test.ctype+"; boundary="+boundary)
		w := httptest.NewRecorder()
		server.Servers[0].Handler.ServeHTTP(w, req)
		status := w.Code
		if status != test.status {
			t.Errorf("got %d, expected %d", status, test.status)
		}
	}
}

func TestRecieveAndExtract(t *testing.T) {
	fpath := "./testdata/htpasswd"
	var tests = []struct {
		method string
		ctype  string
		status int
		user   string
		passwd string
	}{
		{"POST", UPLOAD_CONTENT_TYPE, 201, "test", "123"},
		{"GET", UPLOAD_CONTENT_TYPE, 405, "test", "123"},
	}

	rdir := "/tmp/tupitest"
	os.MkdirAll(rdir, 0755)
	defer os.RemoveAll(rdir)

	b, _ := ioutil.ReadFile("./testdata/test.tar.gz")
	pr, boundary, err := createMultipartPipeReader("test.tar.gz", b)
	if err != nil {
		t.Errorf("error creating reader")
	}

	dconf := DomainConfig{
		Host: "0.0.0.0",
		Port: 8000,
	}
	vconf := DomainConfig{
		RootDir:        rdir,
		Timeout:        300,
		HtpasswdFile:   fpath,
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: true,
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	conf.Domains["localhost"] = vconf
	server := SetupServer(conf)
	for _, test := range tests {

		req, _ := http.NewRequest(test.method, "/e/", pr)
		req.Host = "localhost"
		req.SetBasicAuth(test.user, test.passwd)
		req.Header.Set("Content-Type", test.ctype+"; boundary="+boundary)
		w := httptest.NewRecorder()
		server.Servers[0].Handler.ServeHTTP(w, req)
		status := w.Code
		if status != test.status {
			t.Errorf("got %d, expected %d", status, test.status)
		}
	}

	// the directory should be ok
	_, err = os.Stat(filepath.Join(rdir, "bla"))
	if err != nil {
		t.Errorf("Error extracting file: %s", err)
	}

	// a normal file should be ok
	_, err = os.Stat(filepath.Join(rdir, "bla", "one.txt"))
	if err != nil {
		t.Errorf("Error extracting file: %s", err)
	}

	// a link to a file inside the root dir should be ok
	_, err = os.Stat(filepath.Join(rdir, "bla", "ble", "four.txt"))
	if err != nil {
		t.Errorf("Error extracting file: %s", err)
	}

	// a link to a file outside the root dir should not be ok
	_, err = os.Stat(filepath.Join(rdir, "bla", "ble", "bad.txt"))
	if err == nil {
		t.Errorf("Error extracting file: %s", err)
	}
}

func TestHTTPServer_RunOneServer(t *testing.T) {
	called := false
	startServerTestFn = func(s *http.Server, use_ssl bool) {
		called = true
	}
	defer func() {
		startServerTestFn = nil
	}()
	dconf := DomainConfig{}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	s := SetupServer(conf)
	s.Run()

	if !called {
		t.Fatalf("startServerFn was not called!")
	}
}

func TestHTTPServer_RunMultipleServers(t *testing.T) {
	called := false
	startServerTestFn = func(s *http.Server, use_ssl bool) {
		called = true
	}
	defer func() {
		startServerTestFn = nil
	}()
	dconf := DomainConfig{}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	s := SetupServer(conf)
	s.Servers = append(s.Servers, s.Servers[0])
	s.Run()

	if !called {
		t.Fatalf("startServerFn not called")
	}
}

func TestGetCertificate_Default(t *testing.T) {
	info := tls.ClientHelloInfo{ServerName: "somewhere.com"}
	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = DomainConfig{
		CertFilePath: "./testdata/test.cert",
		KeyFilePath:  "./testdata/test.key",
	}
	oldconf := config
	config = c
	defer func() {
		config = oldconf
	}()
	cert, err := getCertificate(&info)
	if err != nil {
		t.Fatalf("Error getCertificate default %s", err.Error())
	}
	if cert == nil {
		t.Fatalf("Bad certificate")
	}
}

func TestGetCertificate_VirtualDomain(t *testing.T) {
	info := tls.ClientHelloInfo{ServerName: "somewhere.com"}
	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["somewhere.com"] = DomainConfig{
		CertFilePath: "./testdata/test.cert",
		KeyFilePath:  "./testdata/test.key",
	}
	oldconf := config
	config = c
	defer func() {
		config = oldconf
	}()
	cert, err := getCertificate(&info)
	if err != nil {
		t.Fatalf("Error getCertificate default %s", err.Error())
	}
	if cert == nil {
		t.Fatalf("Bad certificate")
	}
}

func TestTupiServer_LoadPlugins(t *testing.T) {
	aconf := DomainConfig{}
	aconf.AuthPlugin = "./build/auth_plugin.so"
	otherconf := DomainConfig{}
	otherconf.AuthPlugin = "./build/init_plugin_bad.so"
	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = aconf
	c.Domains["other"] = otherconf
	SetupServer(c)
	_, err := GetAuthPlugin(aconf.AuthPlugin)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
