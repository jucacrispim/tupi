// Copyright 2015-2019 Juca Crispim <juca@poraodojuca.net>

// This file is part of toxicbuild.

// toxicbuild is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// toxicbuild is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with toxicbuild. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"testing"
)

func TestFileExists(t *testing.T) {
	var tests = []struct {
		path   string
		exists bool
	}{
		{"./testdata/badfile.txt", false},
		{"./testdata/file.txt", true},
	}

	for _, test := range tests {
		r := FileExists(test.path)
		if r != test.exists {
			t.Errorf("got %t, expected %t", r, test.exists)
		}
	}

}

func TestGetFile(t *testing.T) {

	var tests = []struct {
		path string
		err  bool
	}{
		{"./testdata/badfile.txt", true},
		{"./testdata/file.txt", false},
	}

	for _, test := range tests {
		_, err := GetFile(test.path)
		has_err := err != nil
		if has_err != test.err {
			t.Errorf("got %t, expected %t", has_err, test.err)
		}
	}

}
