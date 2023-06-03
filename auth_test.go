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
	"net/http"
	"testing"
)

func TestCredentials(t *testing.T) {

	fpath := "./testdata/htpasswd"
	bad := "./testdata/badhtpasswd"
	pwd := "$2y$05$aaD9jwzs9TImqvelCOe3K.S7bdR.UBgG71yqo0j0vZ0PaBpVZaDKO"
	var tests = []struct {
		fpath   string
		user    string
		pwd     string
		has_err bool
	}{
		{fpath, "test", pwd, false},
		{fpath, "chico", pwd, false},
		{bad, "chico", "", true},
	}

	for _, test := range tests {
		creds, err := authCredentials(test.fpath)
		if err != nil && !test.has_err {
			t.Errorf("%s", err)
			continue
		}

		if creds[test.user] != test.pwd {
			t.Errorf("Got a bad password: %s", creds["test"])
		}

	}
}

func TestUserSecret(t *testing.T) {
	fpath := "./testdata/htpasswd"
	var tests = []struct {
		username  string
		fpath     string
		has_error bool
	}{
		{"test", fpath, false},
		{"zé", fpath, true},
	}

	for _, test := range tests {
		_, err := userSecret(test.username, test.fpath)

		if err != nil && !test.has_error {
			t.Errorf("%s", err)
		}
	}
}

func TestAuthenticate_BasicAuth(t *testing.T) {
	req, _ := http.NewRequest("GET", "/u/", nil)
	fpath := "./testdata/htpasswd"
	var tests = []struct {
		user     string
		password string
		ok       bool
		fpath    string
	}{
		{"test", "123", true, fpath},
		{"test", "345", false, fpath},
		{"missing", "123", false, fpath},
		{"fpath", "123", false, ""},
	}
	for _, test := range tests {
		conf := DomainConfig{}
		conf.HtpasswdFile = test.fpath

		req.SetBasicAuth(test.user, test.password)
		r := authenticate(req, conf)

		if r != test.ok {
			t.Errorf("error in %s %s: %t", test.user, test.password, r)
		}
	}
}

func TestAuthenticate_Plugin(t *testing.T) {
	req, _ := http.NewRequest("GET", "/u/", nil)
	fpath := "./build/auth_plugin.so"
	fpath_bad := "./build/auth_plugin_bad.so"
	fpath_panic := "./build/auth_plugin_panic.so"
	var tests = []struct {
		host  string
		ok    bool
		fpath string
	}{
		{"test.localhost", true, fpath},
		{"bla", false, fpath},
		{"bla", false, "error.so"},
		{"bla", false, fpath_bad},
		{"bla", false, fpath_panic},
	}
	for _, test := range tests {
		conf := DomainConfig{}
		conf.AuthPlugin = test.fpath
		req.Host = test.host
		r := authenticate(req, conf)

		if r != test.ok {
			t.Errorf("error in %s: %t", test.host, r)
		}
	}
}
