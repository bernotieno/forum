package Test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func TestSessionFunctions(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Insert test data into the users table (if needed)
	_, err = db.Exec("INSERT INTO users (id, username) VALUES (1, 'user1'), (2, 'user2')")
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	// Test data
	sessionToken := "test_session_token"
	userID := 1
	expiresAt := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name    string
		setup   func(db *sql.DB)
		test    func(db *sql.DB) error
		wantErr bool
		errMsg  string
	}{
		{
			name: "AddSession - Success",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			test: func(db *sql.DB) error {
				return controllers.AddSession(db, sessionToken, userID, expiresAt)
			},
			wantErr: false,
		},
		{
			name: "AddSession - Duplicate Session Token",
			setup: func(db *sql.DB) {
				// Add a session with the same token
				_ = controllers.AddSession(db, sessionToken, userID, expiresAt)
			},
			test: func(db *sql.DB) error {
				return controllers.AddSession(db, sessionToken, userID, expiresAt)
			},
			wantErr: true,
			errMsg:  "UNIQUE constraint failed", // SQLite error for duplicate session token
		},
		{
			name: "GetSession - Success",
			setup: func(db *sql.DB) {
				// Add a session
				_ = controllers.AddSession(db, sessionToken, userID, expiresAt)
			},
			test: func(db *sql.DB) error {
				_, _, err := controllers.GetSession(db, sessionToken)
				return err
			},
			wantErr: false,
		},
		{
			name: "GetSession - Non-Existent Session",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			test: func(db *sql.DB) error {
				_, _, err := controllers.GetSession(db, "non_existent_token")
				return err
			},
			wantErr: true,
			errMsg:  "sql: no rows in result set",
		},
		{
			name: "IsValidSession - Valid Session",
			setup: func(db *sql.DB) {
				// Add a session
				_ = controllers.AddSession(db, sessionToken, userID, expiresAt)
			},
			test: func(db *sql.DB) error {
				_, valid := controllers.IsValidSession(db, sessionToken)
				if !valid {
					return errors.New("session is not valid")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name: "IsValidSession - Expired Session",
			setup: func(db *sql.DB) {
				// Add an expired session
				_ = controllers.AddSession(db, sessionToken, userID, time.Now().Add(-1*time.Hour))
			},
			test: func(db *sql.DB) error {
				_, valid := controllers.IsValidSession(db, sessionToken)
				if valid {
					return errors.New("session should be expired")
				}
				return nil
			},
			wantErr: true,
		},
		{
			name: "DeleteSession - Success",
			setup: func(db *sql.DB) {
				// Add a session
				_ = controllers.AddSession(db, sessionToken, userID, expiresAt)
			},
			test: func(db *sql.DB) error {
				return controllers.DeleteSession(db, sessionToken)
			},
			wantErr: false,
		},
		{
			name: "DeleteSession - Non-Existent Session",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			test: func(db *sql.DB) error {
				return controllers.DeleteSession(db, "non_existent_token")
			},
			wantErr: false, // Deleting a non-existent session should not return an error
		},
		{
			name: "DeleteExpiredSessions - Success",
			setup: func(db *sql.DB) {
				// Add an expired session
				_ = controllers.AddSession(db, sessionToken, userID, time.Now().Add(-1*time.Hour))
			},
			test: func(db *sql.DB) error {
				return controllers.DeleteExpiredSessions(db)
			},
			wantErr: false,
		},
		{
			name: "DeleteUserSessions - Success",
			setup: func(db *sql.DB) {
				// Add a session for the user
				_ = controllers.AddSession(db, sessionToken, userID, expiresAt)
			},
			test: func(db *sql.DB) error {
				return controllers.DeleteUserSessions(db, userID)
			},
			wantErr: false,
		},
		{
			name: "GetSessionToken - Success",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			test: func(db *sql.DB) error {
				// Create a mock HTTP request with a session cookie
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				cookie := &http.Cookie{Name: "session_token", Value: sessionToken}
				req.AddCookie(cookie)

				_, err := controllers.GetSessionToken(req)
				return err
			},
			wantErr: false,
		},
		{
			name: "GetSessionToken - No Cookie",
			setup: func(db *sql.DB) {
				// No setup needed
			},
			test: func(db *sql.DB) error {
				// Create a mock HTTP request without a session cookie
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				_, err := controllers.GetSessionToken(req)
				return err
			},
			wantErr: true,
			errMsg:  "http: named cookie not present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			// Run the test function
			err := tt.test(db)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: error = %v, wantErr = %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("%s: error = %v, wantErrMsg = %v", tt.name, err, tt.errMsg)
			}
		})
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	// Initialize the logger (if required)
	logger.Init()

	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear tables before test
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Insert expired and non-expired sessions into the database
	_, err = db.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)",
		"expired-session", 1, time.Now().Add(-1*time.Hour)) // Expired 1 hour ago
	if err != nil {
		t.Fatalf("Failed to insert expired session: %v", err)
	}
	_, err = db.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)",
		"valid-session", 2, time.Now().Add(1*time.Hour)) // Expires in 1 hour
	if err != nil {
		t.Fatalf("Failed to insert valid session: %v", err)
	}

	// Create a context with a timeout for the cleanup task
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run the cleanup task
	go controllers.CleanupExpiredSessions(ctx, db)

	// Wait for the cleanup task to run
	time.Sleep(1 * time.Second)

	// Verify that only the expired session was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_token = 'expired-session'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query expired session: %v", err)
	}
	if count != 0 {
		t.Errorf("CleanupExpiredSessions() expired session count = %v, want 0", count)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_token = 'valid-session'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query valid session: %v", err)
	}
	if count != 1 {
		t.Errorf("CleanupExpiredSessions() valid session count = %v, want 1", count)
	}
}
