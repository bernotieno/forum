package middleware

import (
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		if rw.status >= 400 {
			switch rw.status {
			case http.StatusNotFound:
				handlers.ServeErrorPage(w, rw.status, "Page Not Found", "Oops! The page you're looking for doesn't exist.")
			case http.StatusInternalServerError:
				handlers.ServeErrorPage(w, rw.status, "Internal Server Error", "Something went wrong.")
			default:
				handlers.ServeErrorPage(w, rw.status, "Error", "An unexpected error occurred.")
			}
		}
	})
}
