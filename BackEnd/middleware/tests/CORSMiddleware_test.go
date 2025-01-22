package Test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

// TestCORSMiddleware tests the CORSMiddleware function
func TestCORSMiddleware(t *testing.T) {
	// Test cases
	tests := []struct {
		name            string
		method          string            // HTTP method for the request
		expectedStatus  int               // Expected HTTP status code
		expectedHeaders map[string]string // Expected headers in the response
	}{
		{
			name:           "Preflight OPTIONS request",
			method:         http.MethodOptions,
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Methods":     "POST, GET, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, X-CSRF-Token",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:           "Regular GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:           "Regular POST request",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Credentials": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock handler to pass to the middleware
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply the CORSMiddleware to the mock handler
			handler := middleware.CORSMiddleware(mockHandler)

			// Create a request with the specified method
			req, err := http.NewRequest(tt.method, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check the headers
			for key, value := range tt.expectedHeaders {
				if actualValue := rr.Header().Get(key); actualValue != value {
					t.Errorf("handler returned wrong header for %s: got %v want %v", key, actualValue, value)
				}
			}
		})
	}
}
