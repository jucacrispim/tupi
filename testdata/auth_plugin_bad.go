package main

import "net/http"

func Authenticate(r *http.Request) bool {
	if r.Host == "test.localhost" {
		return true
	}
	return false
}
