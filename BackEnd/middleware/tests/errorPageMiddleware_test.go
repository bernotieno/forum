package Test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

// MockServeErrorPageFunc is a mock implementation of ServeErrorPageFunc
func mockServeErrorPageFunc(w http.ResponseWriter, status int, title string, message string) {
	w.WriteHeader(status)
	w.Write([]byte(title + ": " + message))
}

func TestErrorHandler(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		statusCode     int    // Simulated status code from the next handler
		expectedStatus int    // Expected status code in the response
		expectedBody   string // Expected response body
	}{
		{
			name:           "No error (status 200)",
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Not Found error (status 404)",
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Page Not Found: Oops! The page you're looking for doesn't exist.",
		},
		{
			name:           "Internal Server Error (status 500)",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error: Something went wrong.",
		},
		{
			name:           "Generic error (status 400)",
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Error: An unexpected error occurred.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create a mock handler to simulate the next handler in the chain
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			// Apply the ErrorHandler middleware with the mock ServeErrorPageFunc
			handler := middleware.ErrorHandler(mockServeErrorPageFunc)(mockHandler)

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check the response body
			if body := rr.Body.String(); body != tt.expectedBody {
				t.Errorf("handler returned wrong response body: got %v want %v", body, tt.expectedBody)
			}
		})
	}
}
