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

package tupi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	dir := "/tmp/tupitest"
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)

	var tests = []struct {
		content           []byte
		randfname         bool
		prevent_overwrite bool
		has_err           bool
	}{
		{[]byte("oi"), false, false, false},
		{[]byte("oi"), true, false, false},
		{[]byte("oi"), false, true, true},
	}

	for _, test := range tests {
		r, err := createMultipartReader("file.txt", test.content)
		if err != nil {
			t.Errorf("Error creating reader %s", err)

		}

		fname, err := writeFile(dir, r, test.randfname, test.prevent_overwrite)
		if err != nil && !test.has_err {
			t.Errorf("Error writing file: %s", err)
		}

		if fname != "file.txt" && !test.randfname && !test.has_err {
			t.Errorf("File %s not present", fname)
		}

	}

}

func TestExtractFiles(t *testing.T) {
	f, _ := os.Open("./testdata/test.tar.gz")
	root_dir := "/tmp/xx"
	defer os.RemoveAll(root_dir)
	fl, err := extractFiles(f, root_dir, false)

	if err != nil {
		t.Errorf("error extracting files %s", err)
	}

	bad_links := make(map[string]bool, 0)
	bad_links["bla/ble/bad.txt"] = true

	for _, fname := range fl {
		path := filepath.Join(root_dir, fname)
		_, err = os.Stat(path)
		is_bad := bad_links[fname]
		if err != nil && !is_bad {
			t.Errorf("error extracting file %s: %s", path, err)
		}
	}

	f, _ = os.Open("./testdata/test.tar.gz")
	_, err = extractFiles(f, root_dir, true)

	if err == nil {
		t.Errorf("Error preventing overwrite")
	}
}
