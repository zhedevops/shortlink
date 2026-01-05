package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

func RequireMethod(method string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				http.Error(w, fmt.Sprintf("expected %s method", method), http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireContentType(rct string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ct := r.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, rct) {
				http.Error(w, fmt.Sprintf("unsupported content type, require: %s", rct), http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
