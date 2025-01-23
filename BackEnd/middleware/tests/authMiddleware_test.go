package Test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

// Helper function to insert a session into the database
func insertSession(db *sql.DB, sessionToken string, userID int, expiresAt time.Time) error {
	_, err := db.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)",
		sessionToken, userID, expiresAt)
	return err
}

// Helper function to delete a session from the database
func deleteSession(db *sql.DB, sessionToken string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE session_token = ?", sessionToken)
	return err
}

// Mock controller function for testing
func IsValidSession(db *sql.DB, sessionToken string) (int, bool) {
	userID, expiresAt, err := GetSession(db, sessionToken)
	if err != nil {
		return userID, false
	}

	// Check if the session has expired
	if time.Now().After(expiresAt) {
		// Delete the expired session
		_ = deleteSession(db, sessionToken)
		return userID, false
	}

	return userID, true
}

// GetSession retrieves session data from the database
func GetSession(db *sql.DB, sessionToken string) (int, time.Time, error) {
	var userID int
	var expiresAt time.Time
	err := db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE session_token = ?", sessionToken).
		Scan(&userID, &expiresAt)
	return userID, expiresAt, err
}

// TestAuthMiddleware tests the AuthMiddleware function
func TestAuthMiddleware(t *testing.T) {
	// Initialize the logger (if required)
	logger.Init()

	// Setup test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear all tables before running tests
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Test cases
	tests := []struct {
		name           string
		setup          func(db *sql.DB) // Setup function to prepare the database
		sessionCookie  *http.Cookie     // Session cookie for the request
		expectedStatus int              // Expected HTTP status code
		expectedPath   string           // Expected redirect path (if any)
	}{
		{
			name: "No session cookie",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			sessionCookie:  nil,
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/login_Page",
		},
		{
			name: "Empty session cookie",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			sessionCookie:  &http.Cookie{Name: "session_token", Value: ""},
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/login_Page",
		},
		{
			name: "Invalid session token",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			sessionCookie:  &http.Cookie{Name: "session_token", Value: "invalid_token"},
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/login_Page",
		},
		{
			name: "Valid session token",
			setup: func(db *sql.DB) {
				// Insert a valid session into the database
				expiresAt := time.Now().Add(24 * time.Hour) // Session expires in 24 hours
				err := insertSession(db, "valid_token", 1, expiresAt)
				if err != nil {
					t.Fatalf("Failed to insert session: %v", err)
				}
			},
			sessionCookie:  &http.Cookie{Name: "session_token", Value: "valid_token"},
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function
			tt.setup(db)

			// Create a request with the session cookie
			req, err := http.NewRequest("GET", "/protected", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.sessionCookie != nil {
				req.AddCookie(tt.sessionCookie)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create a mock handler to pass to the middleware
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply the AuthMiddleware to the mock handler
			handler := middleware.AuthMiddleware(mockHandler)

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check the redirect path if applicable
			if tt.expectedPath != "" {
				location, err := rr.Result().Location()
				if err != nil {
					t.Errorf("expected redirect to %v, but got no location header", tt.expectedPath)
				} else if location.Path != tt.expectedPath {
					t.Errorf("handler returned wrong redirect path: got %v want %v", location.Path, tt.expectedPath)
				}
			}
		})
	}
}

// Add cleanup helper function
func cleanupTestResources() {
	// Clean up log files
	os.RemoveAll("logs")
	// Clean up uploads directory if it exists
	os.RemoveAll("uploads")
	// Clean up entire storage directory
	os.RemoveAll("BackEnd/database/storage")
}

// Add this new function after cleanupTestResources
func clearDatabaseTables(db *sql.DB) error {
	// List of tables to clear
	tables := []string{"users", "posts", "comments", "likes", "sessions"}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear table %s: %v", table, err)
		}
	}
	return nil
}
