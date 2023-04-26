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

//go:build !test
// +build !test

package main

// notest

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func daemonize(stdout string, stderr string, pidfile string) error {
	r, _, errno := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if errno != 0 {
		return errors.New("Error forking")
	}

	if r != 0 {
		os.Exit(0)
	}

	pid, err := syscall.Setsid()
	if err != nil {
		return errors.New("Error setsid")
	}

	err = redirIO(os.Stdout, stdout)
	if err != nil {
		return err
	}

	err = redirIO(os.Stderr, stderr)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(pidfile, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%d\n", pid))
	f.Sync()

	return nil
}

func redirIO(stream *os.File, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)

	if err != nil {
		return errors.New("Error opening " + path)
	}
	syscall.Dup2(int(f.Fd()), int(stream.Fd()))
	return nil
}
