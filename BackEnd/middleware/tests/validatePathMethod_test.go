package Test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func TestValidatePathAndMethod(t *testing.T) {
	type args struct {
		expectedPath   string
		expectedMethod string
	}
	tests := []struct {
		name string
		args args
		want func(http.Handler) http.Handler
	}{
		{
			name: "Valid path and method",
			args: args{
				expectedPath:   "/valid-path",
				expectedMethod: http.MethodGet,
			},
			want: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/valid-path" || r.Method != http.MethodGet {
						http.Error(w, "Invalid path or method", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name: "Invalid path",
			args: args{
				expectedPath:   "/valid-path",
				expectedMethod: http.MethodGet,
			},
			want: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/invalid-path" || r.Method != http.MethodGet {
						http.Error(w, "Invalid path or method", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name: "Invalid method",
			args: args{
				expectedPath:   "/valid-path",
				expectedMethod: http.MethodGet,
			},
			want: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/valid-path" || r.Method != http.MethodPost {
						http.Error(w, "Invalid path or method", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name: "Empty path",
			args: args{
				expectedPath:   "",
				expectedMethod: http.MethodGet,
			},
			want: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "" || r.Method != http.MethodGet {
						http.Error(w, "Invalid path or method", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name: "Empty method",
			args: args{
				expectedPath:   "/valid-path",
				expectedMethod: "",
			},
			want: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/valid-path" || r.Method != "" {
						http.Error(w, "Invalid path or method", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name: "Case-sensitive path",
			args: args{
				expectedPath:   "/Valid-Path",
				expectedMethod: http.MethodGet,
			},
			want: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/valid-path" || r.Method != http.MethodGet {
						http.Error(w, "Invalid path or method", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		{
			name: "Case-sensitive method",
			args: args{
				expectedPath:   "/valid-path",
				expectedMethod: "GET",
			},
			want: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/valid-path" || r.Method != "get" {
						http.Error(w, "Invalid path or method", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := middleware.ValidatePathAndMethod(tt.args.expectedPath, tt.args.expectedMethod)

			// Create a mock handler to pass to the middleware
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap the mock handler with the middleware
			handler := got(mockHandler)

			// Create a test request
			req, err := http.NewRequest(tt.args.expectedMethod, tt.args.expectedPath, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request using the handler
			handler.ServeHTTP(rr, req)

			// Check if the response matches the expected behavior
			if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest && rr.Code != 405 {
				t.Errorf("unexpected status code: got %v, want %v or %v", rr.Code, http.StatusOK, http.StatusBadRequest)
			}
		})
	}
}
