package main

import "net/http"

func Serve(r *http.Request, domain string, conf *map[string]any) (bool, int, []byte) {
	if domain == "error.req" {
		return false, 400, []byte("something went wrong")
	}
	return true, 200, []byte("serve plugin")
}
