package Test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func TestTimeoutMiddleware(t *testing.T) {
	// Define a timeout duration for the middleware
	timeout := 100 * time.Millisecond

	// Create a handler that simulates a long-running process
	longRunningHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Simulate a delay longer than the timeout
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the long-running handler with the TimeoutMiddleware
	handler := middleware.TimeoutMiddleware(timeout)(longRunningHandler)

	// Create a test request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the request using the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status code %d, got %d", http.StatusGatewayTimeout, rr.Code)
	}

	// Check the response body
	expectedBody := "Request timeout\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestTimeoutMiddleware_NoTimeout(t *testing.T) {
	// Define a timeout duration for the middleware
	timeout := 200 * time.Millisecond

	// Create a handler that simulates a short-running process
	shortRunningHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate a delay shorter than the timeout
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the short-running handler with the TimeoutMiddleware
	handler := middleware.TimeoutMiddleware(timeout)(shortRunningHandler)

	// Create a test request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the request using the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}
}
