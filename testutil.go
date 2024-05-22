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
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
)

func createBufferMultipartReader(fname string, content string, prefix string) (*bytes.Buffer, string, error) {
	buf := new(bytes.Buffer)
	bw := multipart.NewWriter(buf)

	file, err := bw.CreateFormFile("file", "file.txt")
	if err != nil {
		return nil, "", err
	}
	file.Write([]byte(content))

	field, err := bw.CreateFormField("prefix")
	if err != nil {
		return nil, "", err
	}
	field.Write([]byte(prefix))

	bw.Close()
	return buf, bw.Boundary(), nil
}
func createMultipartPipeReader(fname string, content []byte) (
	*io.PipeReader, string, error) {
	// https://stackoverflow.com/questions/43904974/
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() error {
		defer writer.Close()
		part, err := writer.CreateFormFile("file", "file.txt")
		if err != nil {
			return err
		}
		part.Write([]byte(content))
		return nil
	}()

	return pr, writer.Boundary(), nil
}

func createMultipartReader(fname string, content []byte) (
	*multipart.Reader, error) {
	pr, boundary, err := createMultipartPipeReader(fname, content)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("GET", "/e/", pr)
	req.Header.Set("Content-Type", UPLOAD_CONTENT_TYPE+"; boundary="+boundary)
	reader, err := req.MultipartReader()
	if err != nil {
		return nil, err
	}
	return reader, nil
}
