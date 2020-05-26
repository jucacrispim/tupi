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

//+build !test

package main

import (
	"flag"
	"fmt"
)

func main() {
	host := flag.String("host", "0.0.0.0:8080", "host:port to listen.")
	rdir := flag.String("root", ".", "The directory to serve files from")
	timeout := flag.Int("timeout", 240, "Timeout in seconds for read/write")
	htpasswdFile := flag.String(
		"htpasswd",
		"",
		"Full path for a htpasswd file used for authentication")
	upath := flag.String("upath", "/u/", "Path to upload files")
	maxUpload := flag.Int64("maxupload", 10<<20, "Max size for uploaded files")
	certfile := flag.String("certfile", "", "Path for the tls certificate file")
	keyfile := flag.String("keyfile", "", "Path for the tls key file")

	flag.Parse()

	server := SetupServer(*host, *rdir, *timeout, *htpasswdFile, *upath,
		*maxUpload)

	has_cert := *certfile != ""
	has_key := *keyfile != ""

	if (has_cert || has_key) && !(has_cert && has_key) {
		panic("To use HTTPS you must pass certfile and keyfile")
	}

	fmt.Println("Tupi is serving at " + server.Addr)

	var err error = nil
	if has_cert && has_key {
		server.ListenAndServeTLS(*certfile, *keyfile)
	} else {
		server.ListenAndServe()
	}

	if err != nil {
		panic(err)
	}
}
