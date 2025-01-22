package Test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

// TestSetCSPHeaders tests the SetCSPHeaders function
func TestSetCSPHeaders(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		expectedCSP string // Expected Content-Security-Policy header value
	}{
		{
			name: "Set CSP headers",
			expectedCSP: "default-src 'self'; " +
				"script-src 'self' https://cdnjs.cloudflare.com 'unsafe-inline' 'unsafe-eval'; " +
				"style-src 'self' https://cdnjs.cloudflare.com 'unsafe-inline'; " +
				"font-src 'self' data: https://cdnjs.cloudflare.com/ajax/libs/font-awesome/; " +
				"img-src 'self' data: blob:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock handler to pass to the middleware
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply the SetCSPHeaders middleware to the mock handler
			handler := middleware.SetCSPHeaders(mockHandler)

			// Create a request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the Content-Security-Policy header
			if csp := rr.Header().Get("Content-Security-Policy"); csp != tt.expectedCSP {
				t.Errorf("handler returned wrong Content-Security-Policy header: got %v want %v", csp, tt.expectedCSP)
			}
		})
	}
}
