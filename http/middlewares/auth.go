package middlewares

import (
	"net/http"
	"site/security"
)

type Middleware func(next http.Handler) http.Handler

const AuthorizationHeader = "Authorization"

func AuthMiddleware(s security.TokenSecurity) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			header := r.Header.Get(AuthorizationHeader)
			if header == "" || !s.IsValid(header) {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(rw, r)
		})
	}
}
