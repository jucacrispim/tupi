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

package tupi

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	auth "github.com/abbot/go-http-auth"
)

// username => hashed password
type credentials map[string]string

// fpath => credentials
var credsCache map[string]credentials = make(map[string]credentials)

// parseCredentialsFile parses a htpasswd style file and returns
// a map of username => hashed password
func parseCredentialsFile(fpath string) (credentials, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(b), "\n")
	creds := make(map[string]string)
	for _, line := range lines {
		line = strings.Trim(line, " ")
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return nil, errors.New("Invalid line: " + line)

		}
		creds[strings.Trim(parts[0], " ")] = strings.Trim(parts[1], " ")
	}
	return creds, nil
}

// authCredentials returns the credentials in a given file. Uses
// an in-memory cache for the file parsing results.
func authCredentials(fpath string) (credentials, error) {
	var err error
	creds, cached := credsCache[fpath]
	if !cached {
		creds, err = parseCredentialsFile(fpath)
		if err != nil {
			return nil, err
		}
		credsCache[fpath] = creds
	}
	return creds, nil
}

// userSecret returns the hashed password of a given user using
// a given htpasswd file.
func userSecret(username string, fpath string) (string, error) {
	creds, err := authCredentials(fpath)
	if err != nil {
		return "", err
	}
	pwd, exists := creds[username]
	err = nil
	if !exists {
		err = errors.New("User does not exist")
	}
	return pwd, err
}

func basicAuth(r *http.Request, fpath string) (bool, int) {

	if fpath == "" {
		return false, http.StatusUnauthorized
	}

	var ret bool = false
	var status int = http.StatusUnauthorized
	realm := "*"
	sprovider := func(user, realm string) string {
		pwd, _ := userSecret(user, fpath)
		return pwd
	}
	a := &auth.BasicAuth{Realm: realm, Secrets: sprovider}

	if username := a.CheckAuth(r); username != "" {
		ret = true
		status = http.StatusOK
	}

	return ret, status
}

func authenticate(r *http.Request, conf *DomainConfig) (bool, int) {
	if conf.AuthPlugin == "" {
		Debugf("Loading basicAuth")
		r.Header.Set("AUTH_TYPE", "Basic")
		return basicAuth(r, conf.HtpasswdFile)
	}
	Debugf("Loading auth plugin")
	auth := r.Header.Get("Authorization")
	Debugf("Auth %s", auth)
	p, err := GetAuthPlugin(conf.AuthPlugin)
	if err != nil {
		Errorf(
			"Error gettting auth plugin %s. Not authenticating. %s",
			conf.AuthPlugin, err.Error())
		return false, 500
	}
	ok := false
	defer func() {
		if err := recover(); err != nil {
			Errorf("Error authenticating with %s", conf.AuthPlugin)
		}
	}()
	domain := getDomainForRequest(r)
	Debugf("Got domain %s for request", domain)
	ok, status := p(r, domain, &conf.AuthPluginConf)
	return ok, status
}
