package Test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

// VerifyCSRFToken is the function to verify CSRF tokens
func VerifyCSRFToken(db *sql.DB, r *http.Request) bool {
	// Get the CSRF token from the request
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		// If not found in header, get from form
		token = r.FormValue("csrf_token")
		if token == "" {
			return false
		}
	}

	// Get userID from session
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}

	_, exists := IsValidSession(db, cookie.Value)
	if !exists {
		return false
	}

	// Retrieve the stored token from the database
	storedToken, expiresAt, err := controllers.GetCSRFToken(db, cookie.Value)
	if err != nil {
		return false
	}

	// Check if the token matches and is not expired
	if storedToken != token || time.Now().After(expiresAt) {
		// Delete the expired or invalid token
		_ = controllers.DeleteCSRFToken(db, cookie.Value)
		return false
	}

	return true
}

// Helper function to insert a CSRF token into the database
func insertCSRFToken(db *sql.DB, sessionToken string, csrfToken string, expiresAt time.Time) error {
	_, err := db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
		sessionToken, csrfToken, expiresAt)
	return err
}

// TestVerifyCSRFMiddleware tests the VerifyCSRFMiddleware function
func TestVerifyCSRFMiddleware(t *testing.T) {
	logger.Init()
	// Setup test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test cases
	tests := []struct {
		name           string
		method         string           // HTTP method for the request
		sessionCookie  *http.Cookie     // Session cookie for the request
		csrfToken      string           // CSRF token in the request
		setup          func(db *sql.DB) // Setup function to prepare the database
		expectedStatus int              // Expected HTTP status code
		expectedError  string           // Expected error message in the response
	}{
		{
			name:          "Safe method (GET) - bypass CSRF verification",
			method:        http.MethodGet,
			sessionCookie: &http.Cookie{Name: "session_token", Value: "valid_session"},
			csrfToken:     "",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:          "Invalid CSRF token",
			method:        http.MethodPost,
			sessionCookie: &http.Cookie{Name: "session_token", Value: "valid_session"},
			csrfToken:     "invalid_token",
			setup: func(db *sql.DB) {
				// Insert a valid session
				expiresAt := time.Now().Add(24 * time.Hour)
				err := insertSession(db, "valid_session", 1, expiresAt)
				if err != nil {
					t.Fatalf("Failed to insert session: %v", err)
				}
				// Insert a valid CSRF token
				err = insertCSRFToken(db, "valid_session", "valid_token", expiresAt)
				if err != nil {
					t.Fatalf("Failed to insert CSRF token: %v", err)
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Invalid CSRF token",
		},
		{
			name:          "Valid CSRF token",
			method:        http.MethodPost,
			sessionCookie: &http.Cookie{Name: "session_token", Value: "valid_session1"},
			csrfToken:     "valid_token1",
			setup: func(db *sql.DB) {
				// Insert a valid session
				expiresAt := time.Now().Add(24 * time.Hour)
				err := insertSession(db, "valid_session1", 1, expiresAt)
				if err != nil {
					t.Fatalf("Failed to insert session: %v", err)
				}
				// Insert a valid CSRF token
				err = insertCSRFToken(db, "valid_session1", "valid_token1", expiresAt)
				if err != nil {
					t.Fatalf("Failed to insert CSRF token: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function
			tt.setup(db)

			// Create a request with the specified method and CSRF token
			req, err := http.NewRequest(tt.method, "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.sessionCookie != nil {
				req.AddCookie(tt.sessionCookie)
			}
			if tt.csrfToken != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfToken)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create a mock handler to pass to the middleware
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply the VerifyCSRFMiddleware to the mock handler
			handler := middleware.VerifyCSRFMiddleware(db)(mockHandler)

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check the error message in the response (if applicable)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("handler returned wrong error message: got %v want %v", response["error"], tt.expectedError)
				}
			}
		})
	}
}
