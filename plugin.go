// Copyright 2023, 2024 Juca Crispim <juca@poraodojuca.dev>

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
	"fmt"
	"net/http"
	"plugin"
)

type AuthFn func(*http.Request, string, *map[string]any) (bool, int)
type ServeFn func(http.ResponseWriter, *http.Request, *map[string]any)

var authPluginsCache map[string]AuthFn = make(map[string]AuthFn)
var servePluginsCache map[string]ServeFn = make(map[string]ServeFn)

// InitPlugin tries to run the “Init“ function of a plugin. As it is
// optinal, if not found returns without error.
// The “Init“ function get a domain and a config map for the domain as
// parameters. The plugin “Init“ function signature is as follows:
//
// func(string, map[string]any) error
//
// InitPlugin is intended to be run as part of the server start process.
func InitPlugin(fpath string, domain string, conf *map[string]any) (*plugin.Plugin, error) {
	p, err := plugin.Open(fpath)
	if err != nil {
		return nil, err
	}

	s, err := p.Lookup("Init")
	if err != nil {
		return p, nil
	}

	fn, ok := s.(func(string, *map[string]any) error)
	if !ok {
		return nil, errors.New("Invalid Init symbol for plugin: " + fpath)
	}

	inner := func() {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("Error loading plugin %s", fpath)
				err = errors.New(msg)
			}
		}()
		err = fn(domain, conf)
	}
	inner()
	return p, err
}

// LoadAuthPlugin loads a authentication plugin looking for an “Authenticate“ function.
// The “Authenticate“ function gets a referece to http.Request, a domain and a
// config map for the domain as parameters.
// The signature of the “Authenticate“ function is as follows:
//
//	func(*http.Request, string, map[string]any)
//
// LoadAuthPlugin is intended to be run as part of the server start process.
func LoadAuthPlugin(fpath string, domain string, conf *map[string]any) error {
	p, err := InitPlugin(fpath, domain, conf)
	if err != nil {
		return err
	}

	s, err := p.Lookup("Authenticate")
	if err != nil {
		return err
	}
	fn, ok := s.(func(*http.Request, string, *map[string]any) (bool, int))
	if !ok {
		return errors.New("Invalid Authenticate symbol for plugin: " + fpath)
	}

	authPluginsCache[fpath] = fn
	return nil
}

// LoadServePlugin loads a authentication plugin looking for an “Serve“ function.
// The “Serve“ function gets a referece to http.Request, a domain and a
// config map for the domain as parameters.
// The signature of the “Serve“ function is as follows:
//
//	func(*http.Request, string, map[string]any)
//
// LoadServePlugin is intended to be run as part of the server start process.
func LoadServePlugin(fpath string, domain string, conf *map[string]any) error {
	p, err := InitPlugin(fpath, domain, conf)
	if err != nil {
		return err
	}

	s, err := p.Lookup("Serve")
	if err != nil {
		return err
	}
	fn, ok := s.(func(http.ResponseWriter, *http.Request, *map[string]any))
	if !ok {
		return errors.New("Invalid Serve symbol for plugin: " + fpath)
	}

	if err != nil {
		return err
	}
	servePluginsCache[fpath] = fn
	return nil
}

// Returns an already loaded “Autenticate“ function of an auth plugin.
func GetAuthPlugin(fpath string) (AuthFn, error) {
	if fn, exists := authPluginsCache[fpath]; exists {
		return fn, nil
	}
	return nil, errors.New("Auth plugin " + fpath + " not loaded")
}

// Returns an already loaded "Serve" function of a serve plugin
func GetServePlugin(fpath string) (ServeFn, error) {
	if fn, exists := servePluginsCache[fpath]; exists {
		return fn, nil
	}

	return nil, errors.New("Serve plugin " + fpath + " not loaded")
}
