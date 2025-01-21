package middleware

import (
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

// ValidatePathAndMethod is a middleware factory that returns a middleware function
// to validate the path and method of incoming requests.
func ValidatePathAndMethod(expectedPath string, expectedMethod string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the path and method match the expected values
			if r.URL.Path != expectedPath {
				w.WriteHeader(http.StatusNotFound)
				handlers.ServeErrorPage(w, http.StatusNotFound, "Page Not Found", "Oops! The page you're looking for doesn't exist.")
				return
			} else if r.Method != expectedMethod {
				w.WriteHeader(http.StatusMethodNotAllowed)
				handlers.ServeErrorPage(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Oops! The Method Used is Not Allowed")
			}

			// Call the next handler if validation passes
			next.ServeHTTP(w, r)
		})
	}
}
