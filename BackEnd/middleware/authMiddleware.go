package middleware

import (
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/auth"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

// Middleware to check if the user is authenticated
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated
		sessionCookie, err := r.Cookie("session_token")
		if err != nil || sessionCookie == nil || sessionCookie.Value == "" {
			logger.Warning("Unauthorized attempt  nil sessionCookie - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			http.Redirect(w, r, "/login_Page", http.StatusSeeOther)
			return
		}

		// Validating the session token
		isValid := auth.IsValidSession(sessionCookie.Value)
		if !isValid {
			logger.Warning("Unauthorized attempt  Invalid Session - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			http.Redirect(w, r, "/login_Page", http.StatusSeeOther)
			return
		}

		// User is authenticated, call the next handler
		next.ServeHTTP(w, r)
	})
}

// Middleware chain to apply multiple middleware functions
func ApplyMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
