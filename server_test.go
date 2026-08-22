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
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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
	defaultToIndex := true
	dconf := DomainConfig{
		Host:           "0.0.0.0",
		Port:           8000,
		RootDir:        "./testdata",
		Timeout:        300,
		HtpasswdFile:   "",
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: &defaultToIndex,
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	server := SetupServer(conf)
	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.path, nil)
		w := httptest.NewRecorder()
		server.Servers[0].Server.Handler.ServeHTTP(w, req)
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
	defaultToIndex := false
	dconf := DomainConfig{
		Host:           "0.0.0.0",
		Port:           8000,
		RootDir:        "./testdata",
		Timeout:        300,
		HtpasswdFile:   "",
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: &defaultToIndex,
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
		server.Servers[0].Server.Handler.ServeHTTP(w, req)
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
	defaultToIndex := true
	dconf := DomainConfig{
		Host:           "0.0.0.0",
		Port:           8000,
		RootDir:        "./testdata",
		Timeout:        300,
		HtpasswdFile:   fpath,
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: &defaultToIndex,
		AuthMethods:    []string{"POST", "GET"},
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
		server.Servers[0].Server.Handler.ServeHTTP(w, req)
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
	defaultToIndex := true
	dconf := DomainConfig{
		Host:             "0.0.0.0",
		Port:             8000,
		RootDir:          rdir,
		Timeout:          300,
		HtpasswdFile:     fpath,
		UploadPath:       "/u/",
		ExtractPath:      "/e/",
		MaxUploadSize:    10 << 20,
		DefaultToIndex:   &defaultToIndex,
		PreventOverwrite: true,
		AuthMethods:      []string{"POST"},
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
		server.Servers[0].Server.Handler.ServeHTTP(w, req)
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

	b, _ := os.ReadFile("./testdata/test.tar.gz")
	pr, boundary, err := createMultipartPipeReader("test.tar.gz", b)
	if err != nil {
		t.Errorf("error creating reader")
	}

	dconf := DomainConfig{
		Host: "0.0.0.0",
	}
	defaultToIndex := true
	vconf := DomainConfig{
		Port:           8080,
		RootDir:        rdir,
		Timeout:        300,
		HtpasswdFile:   fpath,
		UploadPath:     "/u/",
		ExtractPath:    "/e/",
		MaxUploadSize:  10 << 20,
		DefaultToIndex: &defaultToIndex,
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	conf.Domains["localhost"] = vconf
	server := SetupServer(conf)
	for _, test := range tests {

		req, _ := http.NewRequest(test.method, "/e/", pr)
		req.Host = "localhost:8080"
		req.SetBasicAuth(test.user, test.passwd)
		req.Header.Set("Content-Type", test.ctype+"; boundary="+boundary)
		w := httptest.NewRecorder()
		server.Servers[0].Server.Handler.ServeHTTP(w, req)
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
	startServerTestFn = func(s *http.Server, use_ssl bool) error {
		called = true
		return nil
	}
	defer func() {
		startServerTestFn = nil
	}()
	dconf := DomainConfig{Port: 8080}
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
	startServerTestFn = func(s *http.Server, use_ssl bool) error {
		called = true
		return nil
	}
	defer func() {
		startServerTestFn = nil
	}()
	dconf := DomainConfig{Port: 8080}
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

func TestGetCertificate_CacheHit(t *testing.T) {
	setupCertTest(t)

	domain := "cachehit.example.com"
	dir := t.TempDir()
	certPath, keyPath, _ := writeTestCertificate(t, dir, domain)

	config.Domains = map[string]DomainConfig{
		domain: {
			CertFilePath: certPath,
			KeyFilePath:  keyPath,
		},
	}

	info := &tls.ClientHelloInfo{ServerName: domain}

	// First call loads from disk and populates the cache
	_, err := getCertificate(info)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should find the certificate in cache (covers the if exists)
	cert, err := getCertificate(info)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if cert == nil {
		t.Fatal("certificate returned from cache is nil")
	}
}

func TestTupiServer_LoadPlugins(t *testing.T) {
	aconf := DomainConfig{Port: 8080}
	aconf.AuthPlugin = "./build/auth_plugin.so"
	aconf.ServePlugin = "./build/serve_plugin.so"
	otherconf := DomainConfig{}
	otherconf.AuthPlugin = "./build/init_plugin_bad.so"
	otherconf.ServePlugin = "./build/init_plugin_bad.so"
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

func TestTupiServer_ServePlugin(t *testing.T) {
	dconf := DomainConfig{Port: 8080}
	dconf.ServePlugin = "./build/serve_plugin.so"

	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	server := SetupServer(conf)
	url := "/serve"
	req, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	server.Servers[0].Server.Handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Wrong status for serve plugin %d", w.Code)
	}
}

func TestTupiServer_ServePlugin_error(t *testing.T) {
	dconf := DomainConfig{Port: 8080}
	dconf.ServePlugin = "./build/serve_plugin.so"

	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf
	server := SetupServer(conf)
	url := "/serve"
	req, _ := http.NewRequest("GET", url, nil)
	req.Host = "error.req"
	w := httptest.NewRecorder()
	server.Servers[0].Server.Handler.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("Wrong status for serve plugin %d", w.Code)
	}
}

func TestGetPortForRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		expected int
	}{
		{
			name: "Host with explicit port",
			request: &http.Request{
				Host: "example.com:8080",
			},
			expected: 8080,
		},
		{
			name: "No port in Host, fallback to LocalAddr",
			request: func() *http.Request {
				req := &http.Request{
					Host: "example.com",
				}
				ctx := context.WithValue(req.Context(), http.LocalAddrContextKey, &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 9090,
				})
				return req.WithContext(ctx)
			}(),
			expected: 9090,
		},
		{
			name: "No Host, fallback to LocalAddr",
			request: func() *http.Request {
				req := &http.Request{}
				ctx := context.WithValue(req.Context(), http.LocalAddrContextKey, &net.TCPAddr{
					IP:   net.IPv4(0, 0, 0, 0),
					Port: 1234,
				})
				return req.WithContext(ctx)
			}(),
			expected: 1234,
		},
		{
			name: "No Host or LocalAddr, with TLS",
			request: &http.Request{
				TLS: &tls.ConnectionState{},
			},
			expected: 443,
		},
		{
			name:     "No Host, no LocalAddr, no TLS",
			request:  &http.Request{},
			expected: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := getPortForRequest(tt.request)
			if err != nil {
				t.Fatalf("Error getting port for request %s", err.Error())
			}
			if port != tt.expected {
				t.Errorf("expected port %q, got %q", tt.expected, port)
			}
		})
	}
}

func TestGetConfigForRequest(t *testing.T) {
	// Setup global config for the test
	oldConfig := config
	defer func() { config = oldConfig }()

	c := Config{}
	c.Domains = make(map[string]DomainConfig)

	// Default config
	c.Domains["default"] = DomainConfig{
		Host:    "0.0.0.0",
		RootDir: "/default_root",
	}

	// Domain with path configs
	c.Domains["example.com"] = DomainConfig{
		Host:    "example.com",
		RootDir: "/example_root",
		PathConfigs: map[string]DomainConfig{
			"/api": {
				RootDir: "/api_root",
			},
			"/api/v1": {
				RootDir: "/v1_root",
			},
		},
	}

	// Domain with specific port to test port mismatch fallback
	c.Domains["port-test.com"] = DomainConfig{
		Host:    "port-test.com",
		Port:    8080,
		RootDir: "/port_root",
	}

	config = c

	tests := []struct {
		name           string
		host           string
		path           string
		expectedRootDir string
	}{
		{
			name:           "Fallback to default domain",
			host:           "unknown.com",
			path:           "/any",
			expectedRootDir: "/default_root",
		},
		{
			name:           "Exact domain match, no path match",
			host:           "example.com",
			path:           "/home",
			expectedRootDir: "/example_root",
		},
		{
			name:           "Path match - shortest prefix",
			host:           "example.com",
			path:           "/api/test",
			expectedRootDir: "/api_root",
		},
		{
			name:           "Path match - longest prefix (best match)",
			host:           "example.com",
			path:           "/api/v1/users",
			expectedRootDir: "/v1_root",
		},
		{
			name:           "Domain match, port match",
			host:           "port-test.com:8080",
			path:           "/any",
			expectedRootDir: "/port_root",
		},
		{
			name:           "Domain match, port mismatch (fallback to default)",
			host:           "port-test.com:9090",
			path:           "/any",
			expectedRootDir: "/default_root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			req.Host = tt.host
			
			resConf := getConfigForRequest(req)
			if resConf.RootDir != tt.expectedRootDir {
				t.Errorf("expected RootDir %s, got %s", tt.expectedRootDir, resConf.RootDir)
			}
		})
	}
}

// setupCertTest isolates the global state used by certificate tests.
func setupCertTest(t *testing.T) {
	t.Helper()

	oldCache := certsCache
	oldConfig := config
	oldLevel := currentLogLevel

	certsCache = make(map[string]tls.Certificate)
	config = Config{}
	SetLogLevel(LevelError)

	t.Cleanup(func() {
		certsCache = oldCache
		config = oldConfig
		currentLogLevel = oldLevel
	})
}

// writeTestCertificate generates a self-signed certificate and writes the
// PEM files to the given directory. It returns the paths and the certificate DER.
func writeTestCertificate(t *testing.T, dir, commonName string) (certPath, keyPath string, der []byte) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	der, err = x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to encode key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	certPath = filepath.Join(dir, "cert.pem")
	keyPath = filepath.Join(dir, "key.pem")

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("failed to write certificate: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	return certPath, keyPath, der
}

func TestReloadCertificatesClearsCache(t *testing.T) {
	setupCertTest(t)

	certsCacheMu.Lock()
	certsCache["example.com"] = tls.Certificate{}
	certsCacheMu.Unlock()

	ReloadCertificates()

	certsCacheMu.RLock()
	_, exists := certsCache["example.com"]
	certsCacheMu.RUnlock()

	if exists {
		t.Fatal("ReloadCertificates did not clear certificate cache")
	}
}

func TestGetCertificateReload(t *testing.T) {
	setupCertTest(t)

	domain := "example.com"
	dir := t.TempDir()

	certPath, keyPath, firstDER := writeTestCertificate(t, dir, "first")

	config.Domains = map[string]DomainConfig{
		domain: {
			CertFilePath: certPath,
			KeyFilePath:  keyPath,
		},
	}

	info := &tls.ClientHelloInfo{ServerName: domain}

	cert1, err := getCertificate(info)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}
	if !bytes.Equal(cert1.Certificate[0], firstDER) {
		t.Fatal("first loaded certificate does not match generated one")
	}

	// Overwrite files with a new certificate.
	_, _, secondDER := writeTestCertificate(t, dir, "second")

	// Should still come from the cache, therefore remains the old certificate.
	cert2, err := getCertificate(info)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}
	if !bytes.Equal(cert2.Certificate[0], firstDER) {
		t.Fatal("certificate was reloaded without calling ReloadCertificates")
	}

	ReloadCertificates()

	cert3, err := getCertificate(info)
	if err != nil {
		t.Fatalf("third load failed: %v", err)
	}
	if bytes.Equal(cert3.Certificate[0], firstDER) {
		t.Fatal("certificate was not updated after ReloadCertificates")
	}
	if !bytes.Equal(cert3.Certificate[0], secondDER) {
		t.Fatal("reloaded certificate does not match the new certificate on disk")
	}
}

func TestGetCertificate_CacheHitAfterLock(t *testing.T) {
	setupCertTest(t)

	domain := "cache-after-lock.example.com"
	dir := t.TempDir()
	certPath, keyPath, _ := writeTestCertificate(t, dir, domain)

	config.Domains = map[string]DomainConfig{
		domain: {
			CertFilePath: certPath,
			KeyFilePath:  keyPath,
		},
	}

	// Ensure the cache is empty for this domain.
	certsCacheMu.Lock()
	delete(certsCache, domain)
	certsCacheMu.Unlock()

	// Hold the domain lock; this makes the next call to getCertificate
	// wait in AcquireLock after passing the first cache check.
	AcquireLock(domain)

	done := make(chan struct{})
	go func() {
		defer close(done)
		cert, err := getCertificate(&tls.ClientHelloInfo{ServerName: domain})
		if err != nil {
			t.Errorf("getCertificate returned error: %v", err)
			return
		}
		if cert == nil {
			t.Error("getCertificate returned a nil certificate")
		}
	}()

	// Small wait to give the goroutine time to pass the first
	// cache check and block on AcquireLock.
	time.Sleep(50 * time.Millisecond)
	runtime.Gosched()

	// Populate the cache while the goroutine is still blocked.
	certsCacheMu.Lock()
	certsCache[domain] = tls.Certificate{}
	certsCacheMu.Unlock()

	// Release the lock so the goroutine can proceed and find the
	// populated cache in the second check.
	ReleaseLock(domain)

	<-done
}
