// Copyright 2023-2025 Juca Crispim <juca@poraodojuca.dev>

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
	"strings"
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

	if conf.Domains["default"].Host != "1.1.1.1" {
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
		t.Fatalf("Err GetConfigFromFile %s", err.Error())
	}

	if conf.Domains["default"].Host != "1.1.1.1" {
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

	ports := conf.Domains["default"].Ports
	if len(ports) != 2 {
		t.Fatalf("Bad ports %T", ports)
	}
}

func TestGetConfig_FromFile_MultipleDomains(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{"-conf", "./testdata/vdomains_conf.toml"}
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

	if conf.Domains["default"].DefaultToIndex == nil {
		t.Fatalf("nil defaultToIndex GetConfigFromFile")
	}

	if !*conf.Domains["default"].DefaultToIndex {
		t.Fatalf("bad defaultToIndex GetConfigFromFile %t", *conf.Domains["default"].DefaultToIndex)
	}

	if conf.Domains["domain"].DefaultToIndex == nil {
		t.Fatalf("nil defaultToIndex GetConfigFromFile")
	}

	if !*conf.Domains["domain"].DefaultToIndex {
		t.Fatalf("bad defaultToIndex GetConfigFromFile %t", *conf.Domains["domain"].DefaultToIndex)
	}

	if conf.Domains["domain"].Host != "3.3.3.3" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["domain"].Host)
	}

	if conf.Domains["other.domain"].DefaultToIndex == nil {
		t.Fatalf("nil defaultToIndex GetConfigFromFile")
	}

	if *conf.Domains["other.domain"].DefaultToIndex {
		t.Fatalf("bad defaultToIndex GetConfigFromFile %t", *conf.Domains["other.domain"].DefaultToIndex)
	}

	if conf.Domains["other.domain"].Host != "4.4.4.4" {
		t.Fatalf("Bad host GetConfigFromFile %s", conf.Domains["other.domain"].Host)
	}
}

func TestGetConfig_FromFile_EnvVar(t *testing.T) {
	old_command := flag.CommandLine
	testCommandLine = []string{}
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

func TestValidate_DuplicatedPortConfig(t *testing.T) {
	config := DomainConfig{
		Ports: []PortConfig{
			{
				Port:   1234,
				UseSSL: false,
			},
			{
				Port:   1234,
				UseSSL: false,
			},
		},
	}

	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = config
	if c.Validate() == nil {
		t.Fatalf("It says the config is valid with duplicated ports conf")
	}
}

func TestValidate_DuplicatedPortConfigDefaultPort(t *testing.T) {
	config := DomainConfig{
		Port: 1234,
		Ports: []PortConfig{
			{
				Port:   1234,
				UseSSL: false,
			},
		},
	}

	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = config
	if c.Validate() == nil {
		t.Fatalf("It says the config is valid with duplicated ports conf")
	}
}

func TestValidate_PortConfigOk(t *testing.T) {
	config := DomainConfig{
		Ports: []PortConfig{
			{
				Port:   1233,
				UseSSL: false,
			},
			{
				Port:   1234,
				UseSSL: false,
			},
		},
	}

	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = config
	if c.Validate() != nil {
		t.Fatalf("invalid config with distinct ports")
	}
}

func TestValidate_PortConfigSSLMissingCert(t *testing.T) {
	config := DomainConfig{
		Ports: []PortConfig{
			{
				Port:   1233,
				UseSSL: false,
			},
			{
				Port:   1234,
				UseSSL: true,
			},
		},
	}

	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	c.Domains["default"] = config
	if c.Validate() == nil {
		t.Fatalf("It says the config is valid with ports ssl conf without cert or key")
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

func TestValidate_ConflictingPortConfs(t *testing.T) {
	dconf1 := DomainConfig{
		CertFilePath: "some-path",
		KeyFilePath:  "the other",
		Ports: []PortConfig{
			{
				Port:   1234,
				UseSSL: false,
			},
		},
	}
	dconf2 := DomainConfig{
		CertFilePath: "some-path",
		KeyFilePath:  "the other",
		Ports: []PortConfig{
			{
				Port:   1234,
				UseSSL: true,
			},
		},
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf1
	conf.Domains["other"] = dconf2
	err := conf.Validate()
	if err == nil {
		t.Fatalf("Bad validate for conflicting port confs")
	}
}

func TestGetPortsConfig(t *testing.T) {
	dconf1 := DomainConfig{
		CertFilePath: "some-path",
		KeyFilePath:  "the other",
		Ports: []PortConfig{
			{
				Port:   1234,
				UseSSL: false,
			},
		},
	}
	dconf2 := DomainConfig{
		Port:         2233,
		CertFilePath: "some-path",
		KeyFilePath:  "the other",
		Ports: []PortConfig{
			{
				Port:   1235,
				UseSSL: true,
			},
		},
	}
	conf := Config{}
	conf.Domains = make(map[string]DomainConfig)
	conf.Domains["default"] = dconf1
	conf.Domains["other"] = dconf2
	portsConf := conf.GetPortsConfig()
	if len(portsConf) != 3 {
		t.Fatalf("bad confs for GetPortsConfig %+v", portsConf)
	}
}

func TestHasPortConf(t *testing.T) {
	type getConfFn func() DomainConfig
	var tests = []struct {
		testName string
		getConf  getConfFn
		hasConf  bool
	}{
		{
			"conf with default port",
			func() DomainConfig {
				dconf := DomainConfig{Port: 8080}
				return dconf
			},
			true,
		},
		{
			"conf with ports conf",
			func() DomainConfig {
				dconf := DomainConfig{
					Ports: []PortConfig{{Port: 8080, UseSSL: false}},
				}
				return dconf
			},
			true,
		},
		{
			"conf without port conf",
			func() DomainConfig {
				dconf := DomainConfig{Port: 9080}
				return dconf
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			conf := test.getConf()
			r := conf.HasPortConf(8080)
			if r != test.hasConf {
				t.Fatalf("bad has port conf %t", test.hasConf)
			}
		})
	}
}

func TestGetConfig_CommandLineOverride(t *testing.T) {

	defer os.Unsetenv("TUPI_CONFIG_FILE")

	old_command := flag.CommandLine
	testCommandLine = []string{
		"-host=127.0.0.1",
		"-port=443",
		"-root=/cli",
		"-timeout=456",
		"-htpasswd=/cli/.htpasswd",
		"-upath=/cli/up",
		"-epath=/cli/ex",
		"-maxupload=900000",
		"-certfile=/cli/cert.pem",
		"-keyfile=/cli/key.pem",
		"-default-to-index",
		"-loglevel=debug",
		"-conf=./testdata/override_conf.toml",
		"-prevent-overwrite=false",
		"-auth-methods=GET,HEAD",
	}

	flag.CommandLine = flag.NewFlagSet("tupi", flag.ExitOnError)
	defer func() {
		testCommandLine = nil
		flag.CommandLine = old_command
	}()

	cfg, err := GetConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	conf := cfg.Domains["default"]

	assertEqual(t, conf.Host, "127.0.0.1", "Host")
	assertEqual(t, conf.Port, 443, "Port")
	assertEqual(t, conf.RootDir, "/cli", "RootDir")
	assertEqual(t, conf.Timeout, 456, "Timeout")
	assertEqual(t, conf.HtpasswdFile, "/cli/.htpasswd", "HtpasswdFile")
	assertEqual(t, conf.UploadPath, "/cli/up", "UploadPath")
	assertEqual(t, conf.ExtractPath, "/cli/ex", "ExtractPath")
	assertEqual(t, conf.MaxUploadSize, int64(900000), "MaxUploadSize")
	assertEqual(t, conf.CertFilePath, "/cli/cert.pem", "CertFilePath")
	assertEqual(t, conf.KeyFilePath, "/cli/key.pem", "KeyFilePath")
	assertEqual(t, *conf.DefaultToIndex, true, "DefaultToIndex")
	assertEqual(t, conf.ConfigFile, "./testdata/override_conf.toml", "ConfigFile")
	assertEqual(t, conf.LogLevel, "debug", "LogLevel")
	assertEqual(t, conf.PreventOverwrite, false, "PreventOverwrite")
	assertStringSliceEqual(t, conf.AuthMethods, []string{"GET", "HEAD"}, "AuthMethods")
}

func assertEqual[T comparable](t *testing.T, got, want T, name string) {
	if got != want {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}

func assertStringSliceEqual(t *testing.T, got, want []string, name string) {
	if len(got) != len(want) {
		t.Errorf("%s: slice length mismatch: got %d, want %d", name, len(got), len(want))
		return
	}
	for i := range got {
		if strings.TrimSpace(got[i]) != strings.TrimSpace(want[i]) {
			t.Errorf("%s[%d]: got %q, want %q", name, i, got[i], want[i])
		}
	}
}
