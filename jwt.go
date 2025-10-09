package main

import (
	"net/http"
	"strings"
)

func getToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	if token != "" {
		return token
	}
	token = r.FormValue("token")
	return token
}
