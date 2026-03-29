package handler

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// BasicAuth middleware that checks for username and password from environment variables.
func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		adminUser := os.Getenv("ADMIN_USER")
		adminPass := os.Getenv("ADMIN_PASSWORD")

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(adminUser)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(adminPass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
