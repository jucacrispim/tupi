//go:build !test

/*
Tupi serves files within a directory and files can also be uploaded to a directory.

Usage:

	tupi [params]

The params are:

	 -certfile string
		 Path for the tls certificate file

	 -conf string
		 Path for the configuration file

	 -default-to-index
		 Returns the index.html instead of listing a directory

	 -epath string
		 Path to extract files (default "/e/")

	 -host string
		 host to listen. (default "0.0.0.0")

	 -htpasswd string
		 Full path for a htpasswd file used for authentication

	 -keyfile string
		 Path for the tls key file

	 -maxupload int
		 Max size for uploaded files (default 10485760)

	 -port int
		 port to listen. (default 8080)

	 -root string
		 The directory to serve files from (default ".")

	 -timeout int
		 Timeout in seconds for read/write (default 240)

	 -upath string
		 Path to upload files (default "/u/")

Only authenticated uploads are allowed. To upload a file one need to use
an authentication method. Check the [online documentation] for details.

[online documentation]: https://tupi.poraodojuca.dev
*/
package main

// notest

import (
	"fmt"

	"github.com/jucacrispim/tupi"
)

func main() {
	conf, err := tupi.GetConfig()
	if err != nil {
		panic("Bad config " + err.Error())
	}
	if err := conf.Validate(); err != nil {
		panic("Invalid configuration! " + err.Error())
	}

	fmt.Println("Tupi is running! ")

	server := tupi.SetupServer(conf)
	server.Run()
}
