package main

import "net/http"

func Authenticate(r *http.Request, conf map[string]interface{}) bool {
	if r.Host == "test.localhost" {
		return true
	}
	return false
}
