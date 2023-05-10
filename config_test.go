// Copyright 2023 Juca Crispim <juca@poraodojuca.net>

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
	"flag"
	"os"
	"testing"
)

func TestGetConfig_FromCommandLine(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{"-host", "1.1.1.1", "-root", "/some/dir"}
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()
	conf, err := GetConfig()

	if err != nil {
		t.Fatalf("Error GetConfigFromCommandLine %s", err.Error())
	}

	if conf.Host != "1.1.1.1" {
		t.Fatalf("Bad host GetConfigFromCommandLine %s", conf.Host)
	}
	if conf.RootDir != "/some/dir" {
		t.Fatalf("Bad root dir GetConfigFromCommandLine %s", conf.RootDir)
	}
}

func TestGetConfig_FromFile(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{"-host", "1.1.1.1", "-conf", "./testdata/conf.toml"}
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()

	conf, err := GetConfig()
	if err != nil {
		t.Fatalf("Errr GetConfigFromFile %s", err.Error())
	}

	if conf.Host != "2.2.2.2" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Host)
	}

	if conf.Port != 1234 {
		t.Fatalf("Bad port GetConfigFromFile %d", conf.Port)
	}
	if conf.RootDir != "/some/dir" {
		t.Fatalf("Bad root dir GetConfigFromFile %s", conf.RootDir)
	}
}

func TestGetConfig_FromFile_EnvVar(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{"-host", "1.1.1.1"}
	os.Setenv("TUPI_CONFIG_FILE", "./testdata/conf.toml")
	defer os.Unsetenv("TUPI_CONFIG_FILE")
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()

	conf, err := GetConfig()
	if err != nil {
		t.Fatalf("Errr GetConfigFromFile %s", err.Error())
	}

	if conf.Host != "2.2.2.2" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Host)
	}

	if conf.Port != 1234 {
		t.Fatalf("Bad port GetConfigFromFile %d", conf.Port)
	}
	if conf.RootDir != "/some/dir" {
		t.Fatalf("Bad root dir GetConfigFromFile %s", conf.RootDir)
	}
}

func TestIsValid_MissingCertFile(t *testing.T) {
	config := Config{
		KeyFilePath: "some path",
	}

	if config.IsValid() {
		t.Fatalf("It says the config is valid missing certfile")
	}
}

func TestIsValid_MissingKeyFile(t *testing.T) {
	config := Config{
		CertFilePath: "some path",
	}

	if config.IsValid() {
		t.Fatalf("It says the config is valid missing keyfile")
	}
}

func TestIsValid_WithCertAndKeyFile(t *testing.T) {
	config := Config{
		CertFilePath: "some-path",
		KeyFilePath:  "the other",
	}

	if !config.IsValid() {
		t.Fatalf("Invalid config with cert and key file!")
	}
}
