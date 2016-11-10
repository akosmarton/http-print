package main

import (
	"net/http"
	"strings"

	"github.com/akosmarton/simplejwt"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const BEARER = "Bearer "
		a := r.Header.Get("Authorization")
		if strings.Contains(a, BEARER) == false {
			http.Error(w, "403 Forbidden", 403)
			return
		}
		f, valid, _ := simplejwt.ParseToken(a[len(BEARER):], []byte(config.JWT.Secret))
		if valid == false {
			http.Error(w, "403 Forbidden (Invalid token)", 403)
			return
		}

		if simplejwt.VerifyFields(f) == false {
			http.Error(w, "403 Forbidden (Expired token)", 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
