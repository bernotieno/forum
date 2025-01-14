package middleware

import "net/http"

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS requests (preflight CORS requests)
		if r.Method == http.MethodOptions {
			// w.Header().Set("Access-Control-Allow-Origin", "http://your-frontend-domain.com")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Set CORS headers for all responses
		// w.Header().Set("Access-Control-Allow-Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
