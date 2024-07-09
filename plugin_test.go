// Copyright 2023 Juca Crispim <juca@poraodojuca.net>

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
	"errors"
	"testing"
)

func TestInitPlugin(t *testing.T) {
	fpath := "./build/init_plugin.so"
	fpath_bad := "./build/init_plugin_bad.so"
	fpath_panic := "./build/init_plugin_panic.so"
	fpath_dont_exist := "./build/dont_exist.so"

	var tests = []struct {
		fpath string
		err   error
	}{
		{fpath, nil},
		{fpath_bad, errors.New("Invalid Init symbol for plugin: " + fpath_bad)},
		{fpath_panic, errors.New("Error loading plugin " + fpath_panic)},
		{fpath_dont_exist, errors.New("plugin.Open(\"./build/dont_exist.so\"): realpath failed")},
	}

	for _, test := range tests {
		_, err := InitPlugin(test.fpath, "some.domain", nil)
		if !compareErr(test.err, err) {
			t.Fatalf("Invalid error %s", err.Error())
		}
	}

}

func TestLoadAuthPlugin(t *testing.T) {
	fpath := "./build/auth_plugin.so"
	fpath_bad := "./build/auth_plugin_bad.so"
	var tests = []struct {
		fpath string
		ok    bool
	}{
		{fpath, true},
		{"error.so", false},
		{fpath_bad, false},
	}
	for _, test := range tests {
		err := LoadAuthPlugin(test.fpath, "domain", nil)
		if err != nil && test.ok {
			t.Fatalf(err.Error())
		}
	}
}

func TestLoadServePlugin(t *testing.T) {
	fpath := "./build/serve_plugin.so"
	fpath_bad := "./build/serve_plugin_bad.so"
	var tests = []struct {
		fpath string
		ok    bool
	}{
		{fpath, true},
		{fpath_bad, false},
	}
	for _, test := range tests {
		err := LoadServePlugin(test.fpath, "domain", nil)
		if err != nil && test.ok {
			t.Fatalf(err.Error())
		}
	}
}

func TestGetAuthPluign(t *testing.T) {
	fpath := "./build/auth_plugin.so"
	LoadAuthPlugin(fpath, "domain", nil)

	var tests = []struct {
		fpath string
		ok    bool
	}{
		{fpath, true},
		{"error.so", false},
	}

	for _, test := range tests {
		_, err := GetAuthPlugin(test.fpath)
		if err != nil && test.ok {
			t.Fatalf(err.Error())
		}
	}
}

func TestGetServePluign(t *testing.T) {
	fpath := "./build/serve_plugin.so"
	LoadServePlugin(fpath, "domain", nil)

	var tests = []struct {
		fpath string
		ok    bool
	}{
		{fpath, true},
		{"error.so", false},
	}

	for _, test := range tests {
		_, err := GetServePlugin(test.fpath)
		if err != nil && test.ok {
			t.Fatalf(err.Error())
		}
	}
}

func compareErr(err1 error, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}
