package main

import "net/http"

func Authenticate(r *http.Request, domain string, conf *map[string]any) (bool, int) {
	if r.Host == "test.localhost" {
		return true, 200
	}
	return false, 403
}
