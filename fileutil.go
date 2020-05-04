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
	"io/ioutil"
	"os"
)

type File struct {
	Path    string
	Content []byte
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFile opens the file at `path` and returns a poiter to a File struct.
// Note that the path is relative to the path in the `GIMME_ROOT_DIR` envvar.
func GetFile(path string) (*File, error) {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return &File{Path: path, Content: content}, nil
}
