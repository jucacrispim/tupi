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

//go:build !test
// +build !test

package main

// notest

import (
	"fmt"
	"strconv"
)

func main() {
	conf, err := GetConfig()
	if err != nil {
		panic("Bad config " + err.Error())
	}
	has_cert := conf.HasCert()
	has_key := conf.HasKey()

	if (has_cert || has_key) && !(has_cert && has_key) {
		panic("To use HTTPS you must pass certfile and keyfile")
	}

	ports := strconv.FormatInt(int64(conf.Port), 10)
	host := fmt.Sprintf("%s:%s", conf.Host, ports)
	fmt.Println("Tupi is serving at " + host)

	server := SetupServer(conf)
	server.Run()
}
