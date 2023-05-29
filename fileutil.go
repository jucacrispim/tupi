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
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

var chunkSize int64 = 10 << 20

func genRandFname(fname string) (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-", b) + fname, nil
}

// writeFile writes the contents of an uploaded file into a file in the
// local fs.
func writeFile(dir string, r *multipart.Reader, randfname bool) (string, error) {

	part, err := r.NextPart()
	if err != nil {
		return "", err
	}
	var fname string
	if randfname {
		fname, err = genRandFname(part.FileName())
		if err != nil {
			return "", err
		}
	} else {
		fname = part.FileName()
	}

	fpath := dir + string(os.PathSeparator) + fname
	AcquireLock(fpath)
	defer ReleaseLock(fpath)

	f, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var off int64 = 0
	for {
		c, err := ioutil.ReadAll(part)
		if err != nil {
			return fname, err
		}
		f.WriteAt([]byte(c), off)
		err = f.Sync()
		if err != nil {
			return fname, err
		}
		off += int64(len(c))

		part, err = r.NextPart()

		if err == io.EOF {
			break
		}

	}
	return fname, nil
}

// extractFiles extract the contents of a tar.gz file to the local
// file system. All files will be extracted inside `root_dir`
func extractFiles(file io.Reader, root_dir string) ([]string, error) {
	buf, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer buf.Close()
	tr := tar.NewReader(buf)
	files := make([]string, 0)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		fname := hdr.Name
		path := filepath.Join(root_dir, fname)
		switch hdr.Typeflag {
		case tar.TypeDir:
			// for a directory we hold the lock till the end of the function
			// to avoid someome messing with a directory we are working inside
			AcquireLock(path)
			defer ReleaseLock(path)
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return nil, err
			}
			files = append(files, fname)

		case tar.TypeReg:
			AcquireLock(path)
			out, err := os.Create(path)
			if err != nil {
				ReleaseLock(path)
				return nil, err
			}
			_, err = io.Copy(out, tr)
			ReleaseLock(path)
			if err != nil {
				return nil, err
			}
			files = append(files, fname)

		case tar.TypeSymlink:
			target := filepath.Join(filepath.Dir(path), hdr.Linkname)
			if !strings.HasPrefix(target, root_dir) {
				target = filepath.Join(root_dir, strings.TrimLeft(target, "/"))
			}

			AcquireLock(path)
			err := os.Symlink(target, path)
			ReleaseLock(path)
			if err != nil {
				return nil, err
			}

			files = append(files, fname)

		default:
			// notest
			log.Printf("Unknown type %d for %s", hdr.Typeflag, path)
		}

	}
	return files, nil
}
