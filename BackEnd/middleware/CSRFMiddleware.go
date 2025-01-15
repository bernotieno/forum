package middleware

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
)

// VerifyCSRFMiddleware is a middleware function to verify CSRF tokens
func VerifyCSRFMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF verification for safe methods (GET, HEAD, OPTIONS)
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Verify the CSRF token
			if !controllers.VerifyCSRFToken(db, r) {
				log.Printf("Invalid CSRF token in request - remote_addr: %s, method: %s, path: %s",
					r.RemoteAddr,
					r.Method,
					r.URL.Path,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Invalid CSRF token",
				})
				return
			}

			// If the CSRF token is valid, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
