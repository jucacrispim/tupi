package main

import "net/http"

func Authenticate(r *http.Request, domain string, conf *map[string]any) bool {
	if r.Host == "test.localhost" {
		return true
	}
	return false
}
