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

package tupi

import (
	"errors"
	"flag"
	"os"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

// DomainConfig values known to tupi
type DomainConfig struct {
	Host           string
	Port           int
	RootDir        string
	Timeout        int
	HtpasswdFile   string
	UploadPath     string
	ExtractPath    string
	MaxUploadSize  int64
	CertFilePath   string
	KeyFilePath    string
	DefaultToIndex bool
	ConfigFile     string
	AuthPlugin     string
	AuthPluginConf map[string]interface{}
	LogLevel       string
}

func (c *DomainConfig) HasCert() bool {
	return c.CertFilePath != ""
}

func (c *DomainConfig) HasKey() bool {
	return c.KeyFilePath != ""
}

func (c *DomainConfig) Validate() error {
	has_cert := c.HasCert()
	has_key := c.HasKey()

	if (has_cert || has_key) && !(has_cert && has_key) {
		return errors.New("You must pass certfile and certkey to use ssl")
	}
	return nil
}

type Config struct {
	Domains map[string]DomainConfig
}

func (c *Config) Validate() error {
	for _, v := range c.Domains {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) HasSSL() bool {
	for _, v := range c.Domains {
		if v.HasCert() && v.HasKey() {
			return true
		}
	}
	return false
}

// GetConfig returns the config struct for the server by reading
// the confs passed in the command line and optionally in a config file.
// Config file values have precedence over command line values
func GetConfig() (Config, error) {
	cmdConf := GetConfigFromCommandLine()
	envfile := os.Getenv("TUPI_CONFIG_FILE")
	if cmdConf.ConfigFile == "" && envfile == "" {
		c := Config{}
		c.Domains = make(map[string]DomainConfig, 0)
		c.Domains["default"] = cmdConf
		return c, nil
	}
	// cmd line conffile has precedence over envvar config file
	cfg := ""
	if cmdConf.ConfigFile != "" {
		cfg = cmdConf.ConfigFile
	} else {
		cfg = envfile
	}
	fileConf, err := GetConfigFromFile(cfg)
	if err != nil {
		return Config{}, err
	}
	defaultConf := fileConf.Domains["default"]
	defaultConf = mergeConfs(defaultConf, cmdConf)
	fileConf.Domains["default"] = defaultConf
	for k, v := range fileConf.Domains {
		if k == "default" {
			continue
		}
		conf := mergeConfs(v, defaultConf)
		fileConf.Domains[k] = conf
	}
	return fileConf, nil
}

func GetConfigFromCommandLine() DomainConfig {
	host := flag.String("host", "", "host to listen.")
	port := flag.Int("port", 8080, "port to listen.")
	rdir := flag.String("root", ".", "The directory to serve files from")
	timeout := flag.Int("timeout", 240, "Timeout in seconds for read/write")
	htpasswdFile := flag.String(
		"htpasswd",
		"",
		"Full path for a htpasswd file used for authentication")
	upath := flag.String("upath", "/u/", "Path to upload files")
	epath := flag.String("epath", "/e/", "Path to extract files")
	maxUpload := flag.Int64("maxupload", 10<<20, "Max size for uploaded files")
	certfile := flag.String("certfile", "", "Path for the tls certificate file")
	keyfile := flag.String("keyfile", "", "Path for the tls key file")
	defaultToIndex := flag.Bool(
		"default-to-index",
		false,
		"Returns the index.html instead of listing a directory")
	conf_path := flag.String("conf", "", "Path for the configuration file")
	log_level := flag.String("loglevel", "info", "Log level")

	args := getCmdlineArgs()
	flag.CommandLine.Parse(args)
	conf := DomainConfig{
		Host:           *host,
		Port:           *port,
		RootDir:        *rdir,
		Timeout:        *timeout,
		HtpasswdFile:   *htpasswdFile,
		UploadPath:     *upath,
		ExtractPath:    *epath,
		MaxUploadSize:  *maxUpload,
		CertFilePath:   *certfile,
		KeyFilePath:    *keyfile,
		DefaultToIndex: *defaultToIndex,
		ConfigFile:     *conf_path,
		LogLevel:       *log_level,
	}
	return conf
}

func GetConfigFromFile(fpath string) (Config, error) {
	bytes, error := os.ReadFile(fpath)
	if error != nil {
		return Config{}, error
	}
	rawConf := string(bytes)
	c := Config{}
	c.Domains = make(map[string]DomainConfig)
	confs := getDomainRawConfs(rawConf)
	for domain, raw := range confs {
		conf := DomainConfig{}
		_, err := toml.Decode(raw, &conf)
		if err != nil {
			return Config{}, err
		}
		c.Domains[domain] = conf
	}
	return c, nil
}

// merge two confs together. confA has precedence over confB
func mergeConfs(confA DomainConfig, confB DomainConfig) DomainConfig {
	valA := reflect.ValueOf(confA)
	valB := reflect.ValueOf(&confB).Elem()
	for i := 0; i < valA.NumField(); i++ {
		aField := valA.Field(i)
		if !aField.IsZero() {
			name := valA.Type().Field(i).Name
			f := valB.FieldByName(name)
			val := aField
			f.Set(val)
		}
	}
	return confB
}

// what we do here is a kind of hack before sending
// the config to the toml parser.
// First we split the contents in different sections (marked by a
// line begining with `[`) and then send each section to the parser.
// This is done so I can use each section as a key for the domains
// map in the config.
// After that I remove new lines inside inline tables. This is done
// because toml does not accept new lines inside inline tables but
// I want to be able to write them with new lines.
// Check the documentation for the conf syntax.
func getDomainRawConfs(rawConf string) map[string]string {
	confs := make(map[string]string, 0)
	conf := ""
	domain := "default"
	// boring stuff to remove lf cr from inline tables
	lf := '\n'
	cr := '\r'
	sQuote := '\''
	dQuote := '"'
	escapeChr := '\\'
	insideString := false
	var openedChr rune = 0
	openCurly := '{'
	closeCurly := '}'
	insideTable := false
	openedTables := 0

	for _, line := range strings.Split(rawConf, "\n") {
		if strings.HasPrefix(line, "[") {
			if conf != "" {
				confs[domain] = conf
				conf = ""
			}
			domain = strings.Trim(line, "[]")
		} else {
			newLineBytes := make([]byte, 0)
			for i, chr := range line {
				// are we inside a string?
				isQuote := chr == sQuote || chr == dQuote
				isRightQuote := openedChr == chr || !insideString
				isLineStart := i == 0
				var escaped bool
				if isLineStart {
					escaped = false
				} else if len(line) > 0 {
					escaped = line[i-1] == byte(escapeChr)
				}
				if isQuote && isRightQuote && (!escaped || isLineStart) {
					insideString = !insideString
					if insideString {
						openedChr = chr
					} else {
						openedChr = 0
					}
				}
				if insideString {
					newLineBytes = append(newLineBytes, byte(chr))
					continue
				}

				// are we inside a inline table?
				if chr == openCurly {
					openedTables += 1
					insideTable = true
				} else if (chr == closeCurly) && insideTable {
					openedTables -= 1
					if openedTables <= 0 {
						openedTables = 0
						insideTable = false
					}
				}
				if !insideTable || (chr != lf && chr != cr) {
					newLineBytes = append(newLineBytes, byte(chr))
					continue
				}
				// replace line feed or carriage return by nothing
				// newLineBytes = append(newLineBytes, byte(0))
			}
			conf += string(newLineBytes)
			if !insideTable {
				conf += string(lf)
			}
		}
	}
	confs[domain] = conf
	return confs
}

// help tests
var testCommandLine []string = nil

func getCmdlineArgs() []string {
	// notest
	var args []string = nil
	if testCommandLine != nil {
		args = testCommandLine
	} else {
		args = os.Args[1:]
	}
	return args
}
