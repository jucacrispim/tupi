package main

import (
	"net/http"
	"strings"
)

func Serve(w http.ResponseWriter, r *http.Request, conf *map[string]any) {
	domain := strings.Split(r.Host, ":")[0]
	if domain == "error.req" {
		w.WriteHeader(400)
		w.Write([]byte("something went wrong"))
	}
	w.WriteHeader(200)
	w.Write([]byte("serve plugin"))
}
