package main

import "net/http"

func Serve(r *http.Request, domain string, conf *map[string]any) (bool, int) {
	return true, 200
}
