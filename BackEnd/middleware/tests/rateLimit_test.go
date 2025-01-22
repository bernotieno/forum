package Test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

// TestRateLimiter tests the RateLimiter middleware
func TestRateLimiter(t *testing.T) {
	// Create a RateLimiter with a rate of 2 requests per second
	rl := middleware.NewRateLimiter(2, time.Second)

	// Create a mock handler to pass to the RateLimiter middleware
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply the RateLimiter middleware to the mock handler
	handler := rl.RateLimit(mockHandler)

	// Test cases
	tests := []struct {
		name           string
		ip             string // IP address for the request
		requests       int    // Number of requests to simulate
		expectedStatus int    // Expected HTTP status code
	}{
		{
			name:           "Single request within rate limit",
			ip:             "192.168.1.1",
			requests:       1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Multiple requests within rate limit",
			ip:             "192.168.1.2",
			requests:       2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Exceed rate limit",
			ip:             "192.168.1.3",
			requests:       3,
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.requests; i++ {
				// Create a request with the specified IP address
				req, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = tt.ip

				// Create a ResponseRecorder to record the response
				rr := httptest.NewRecorder()

				// Serve the request
				handler.ServeHTTP(rr, req)

				// Check the status code for the last request
				if i == tt.requests-1 {
					if status := rr.Code; status != tt.expectedStatus {
						t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
					}
				}
			}
		})
	}
}

func TestRateLimiter_cleanupVisitors(t *testing.T) {
	type fields struct {
		visitors map[string]*middleware.Visitor
		mu       *sync.RWMutex
		rate     int
		interval time.Duration
	}
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedVisits int // Expected number of visitors after cleanup
	}{
		{
			name: "No visitors",
			fields: fields{
				visitors: make(map[string]*middleware.Visitor),
				mu:       &sync.RWMutex{},
				rate:     2,
				interval: time.Second,
			},
			args: args{
				interval: time.Second,
			},
			expectedVisits: 0,
		},
		{
			name: "All visitors expired",
			fields: fields{
				visitors: map[string]*middleware.Visitor{
					"192.168.1.1": {LastSeen: time.Now().Add(-2 * time.Second), Count: 1},
					"192.168.1.2": {LastSeen: time.Now().Add(-3 * time.Second), Count: 2},
				},
				mu:       &sync.RWMutex{},
				rate:     2,
				interval: time.Second,
			},
			args: args{
				interval: time.Second,
			},
			expectedVisits: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := &middleware.RateLimiter{
				Visitors: tt.fields.visitors,
				Rate:     tt.fields.rate,
				Interval: tt.fields.interval,
			}

			// Create a context for the cleanup goroutine
			ctx, cancel := context.WithCancel(context.Background())
			rl.Ctx = ctx
			rl.Cancel = cancel

			// Run the cleanupVisitors function
			go rl.CleanupVisitors(tt.args.interval)

			// Wait for the cleanup to complete
			time.Sleep(2 * time.Second)

			// Verify the number of visitors after cleanup
			rl.Mu.RLock()
			if len(rl.Visitors) != tt.expectedVisits {
				t.Errorf("expected %d visitors after cleanup, got %d", tt.expectedVisits, len(rl.Visitors))
			}
			rl.Mu.RUnlock()

			// Cancel the context to stop the cleanup goroutine
			cancel()
		})
	}
}
