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
	"mime/multipart"
	"os"
)

var chunkSize int64 = 5 << 20

// writeFile writes do contents of an uploaded file into a file in the
// local fs.
func writeFile(dir string, r *multipart.Reader) (string, error) {

	part, err := r.NextPart()
	if err != nil {
		return "", err
	}
	fname := part.FileName()
	fpath := dir + string(os.PathSeparator) + fname
	AcquireLock(fpath)
	defer RelaseLock(fpath)
	f, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var off int64 = 0
	for {
		c, err := ioutil.ReadAll(part)
		if err != nil {
			return "", err
		}
		f.WriteAt([]byte(c), off)
		err = f.Sync()
		if err != nil {
			return "", nil
		}
		off += int64(len(c))
		if err == io.EOF {
			return fname, nil
		}

		part, err = r.NextPart()
		if err != nil {
			return "", err
		}
	}
}
