package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
}

// ErrorHandler takes a ServeErrorPageFunc as a dependency
func ErrorHandler(serveErrorPageFunc func(w http.ResponseWriter, status int, title string, message string)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			
			// Check if it's an API/AJAX request
			isAPIRequest := r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
						  r.Header.Get("Accept") == "application/json" ||
						  strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")

			next.ServeHTTP(rw, r)

			if rw.status >= 400 {
				if isAPIRequest {
					// For API requests, ensure we're sending JSON
					w.Header().Set("Content-Type", "application/json")
					if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
						// Convert error to JSON format if it's not already
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(map[string]string{
							"error": http.StatusText(rw.status),
						})
					}
					return
				}

				// For regular web requests, serve the error page
				switch rw.status {
				case http.StatusNotFound:
					serveErrorPageFunc(w, rw.status, "Page Not Found", "Oops! The page you're looking for doesn't exist.")
				case http.StatusInternalServerError:
					serveErrorPageFunc(w, rw.status, "Internal Server Error", "Something went wrong.")
				default:
					serveErrorPageFunc(w, rw.status, "Error", "An unexpected error occurred.")
				}
			}
		})
	}
}
