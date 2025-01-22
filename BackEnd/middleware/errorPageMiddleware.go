package middleware

import (
	"net/http"
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
			next.ServeHTTP(rw, r)

			if rw.status >= 400 {
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
