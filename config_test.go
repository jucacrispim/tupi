// Copyright 2023 Juca Crispim <juca@poraodojuca.dev>

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
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestGetConfig_FromCommandLine(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{"-host", "1.1.1.1", "-root", "/some/dir", "-auth-methods", "POST,GET"}
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()
	conf, err := GetConfig()

	if err != nil {
		t.Fatalf("Error GetConfigFromCommandLine %s", err.Error())
	}

	if conf.Domains["default"].Host != "1.1.1.1" {
		t.Fatalf("Bad host GetConfigFromCommandLine %s", conf.Domains["default"].Host)
	}
	if conf.Domains["default"].RootDir != "/some/dir" {
		t.Fatalf("Bad root dir GetConfigFromCommandLine %s", conf.Domains["default"].RootDir)
	}

	if !reflect.DeepEqual(conf.Domains["default"].AuthMethods, []string{"POST", "GET"}) {
		t.Fatalf("Bad auth methods %s", conf.Domains["default"].AuthMethods)
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

	if conf.Domains["default"].Host != "2.2.2.2" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["default"].Host)
	}

	if conf.Domains["default"].Port != 1234 {
		t.Fatalf("Bad port GetConfigFromFile %d", conf.Domains["default"].Port)
	}
	if conf.Domains["default"].RootDir != "/some/dir" {
		t.Fatalf("Bad root dir GetConfigFromFile %s", conf.Domains["default"].RootDir)
	}
}

func TestGetConfig_FromFile_InlineTable(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{"-host", "1.1.1.1", "-conf", "./testdata/conf_inline_table.toml"}
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()

	conf, err := GetConfig()
	if err != nil {
		t.Fatalf("Errr GetConfigFromFile %s", err.Error())
	}

	if conf.Domains["default"].Host != "2.2.2.2" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["default"].Host)
	}

	if conf.Domains["default"].Port != 1234 {
		t.Fatalf("Bad port GetConfigFromFile %d", conf.Domains["default"].Port)
	}
	if conf.Domains["default"].RootDir != "/some/dir" {
		t.Fatalf("Bad root dir GetConfigFromFile %s", conf.Domains["default"].RootDir)
	}

	cnf := conf.Domains["default"].AuthPluginConf
	val, _ := cnf["somekey"].(string)
	if val != "the value" {
		t.Fatalf("Bad plugin conf %s", val)
	}
	other, _ := cnf["other"].(string)
	if other != "strange {value\n}" {
		t.Fatalf("Bad plugin conf %s", other)
	}
}

func TestGetConfig_FromFile_MultipleDomains(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{"-host", "1.1.1.1", "-conf", "./testdata/vdomains_conf.toml"}
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()

	conf, err := GetConfig()
	if err != nil {
		t.Fatalf("Errr GetConfigFromFile %s", err.Error())
	}

	if conf.Domains["default"].Host != "2.2.2.2" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["default"].Host)
	}

	if conf.Domains["domain"].Host != "3.3.3.3" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["domain"].Host)
	}

	if conf.Domains["other.domain"].Host != "4.4.4.4" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["other.domain"].Host)
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

	if conf.Domains["default"].Host != "2.2.2.2" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["default"].Host)
	}

	if conf.Domains["default"].Port != 1234 {
		t.Fatalf("Bad port GetConfigFromFile %d", conf.Domains["default"].Port)
	}
	if conf.Domains["default"].RootDir != "/some/dir" {
		t.Fatalf("Bad root dir GetConfigFromFile %s", conf.Domains["default"].RootDir)
	}
}

func TestValidate_MissingCertFile(t *testing.T) {
	config := DomainConfig{
		KeyFilePath: "some path",
	}
	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = config
	if c.Validate() == nil {
		t.Fatalf("It says the config is valid missing certfile")
	}
}

func TestValidate_MissingKeyFile(t *testing.T) {
	config := DomainConfig{
		CertFilePath: "some path",
	}
	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = config
	if c.Validate() == nil {
		t.Fatalf("It says the config is valid missing keyfile")
	}
}

func TestValidate_WithCertAndKeyFile(t *testing.T) {
	config := DomainConfig{
		CertFilePath: "some-path",
		KeyFilePath:  "the other",
	}
	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = config
	if c.Validate() != nil {
		t.Fatalf("Invalid config with cert and key file!")
	}
}

func TestHasSSL_True(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{}
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()
	fpath := "./testdata/conf.toml"
	os.Setenv("TUPI_CONFIG_FILE", fpath)
	defer os.Unsetenv("TUPI_CONFIG_FILE")
	conf, err := GetConfig()
	if err != nil {
		t.Fatalf("Error reading config file for ssl false %s", err.Error())
	}
	if !conf.HasSSL() {
		t.Fatalf("dont have ssl for config with ssl")
	}
}

func TestHasSSL_False(t *testing.T) {
	fpath := "./testdata/vdomains_conf.toml"
	os.Setenv("TUPI_CONFIG_FILE", fpath)
	defer os.Unsetenv("TUPI_CONFIG_FILE")
	old_command := flag.CommandLine
	testCommandLine = []string{}
	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()

	conf, err := GetConfig()
	if err != nil {
		t.Fatalf("Error reading config file for ssl false %s", err.Error())
	}
	if conf.HasSSL() {
		t.Fatalf("do have ssl for config without ssl")
	}

}
