package middleware

import "net/http"

// Middleware to set Content Security Policy headers
func SetCSPHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' https://cdnjs.cloudflare.com 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' https://cdnjs.cloudflare.com 'unsafe-inline'; "+
				"font-src 'self' data: https://cdnjs.cloudflare.com/ajax/libs/font-awesome/; "+
				"img-src 'self' data: blob:",
		)
		next.ServeHTTP(w, r)
	})
}
