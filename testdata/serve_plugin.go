package main

import "net/http"

func Serve(r *http.Request, domain string, conf *map[string]any) (bool, int, []byte) {
	return true, 200, []byte("serve plugin")
}
