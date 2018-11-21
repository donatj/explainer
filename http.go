package main

import (
	"net/http"
)

func BasicAuth(handler http.Handler, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); !ok || !(u == username && p == password) {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Zork\"")
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	}
}
