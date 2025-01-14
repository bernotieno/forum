package middleware

import (
	"context"
	"net/http"
	"time"
)

func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			done := make(chan bool)

			go func() {
				next.ServeHTTP(w, r)
				done <- true
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				http.Error(w, "Request timeout", http.StatusGatewayTimeout)
				return
			}
		})
	}
}
