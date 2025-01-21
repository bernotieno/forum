package Test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
)

func TestGenerateCSRFToken(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a session token for testing
	sessionToken := "test-session-token"

	tests := []struct {
		name         string
		sessionToken string
		wantToken    bool
		wantErr      bool
		errMsg       string
		setup        func(db *sql.DB)
	}{
		{
			name:         "Generate New Token",
			sessionToken: sessionToken,
			wantToken:    true,
			wantErr:      false,
		},
		{
			name:         "Existing Valid Token",
			sessionToken: sessionToken,
			wantToken:    true,
			wantErr:      false,
			setup: func(db *sql.DB) {
				// Insert a valid token into the database
				_, _ = db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
					sessionToken, "existing-token", time.Now().Add(1*time.Hour))
			},
		},
		{
			name:         "Expired Token",
			sessionToken: sessionToken,
			wantToken:    true,
			wantErr:      false,
			setup: func(db *sql.DB) {
				// Insert an expired token into the database
				_, _ = db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
					sessionToken, "expired-token", time.Now().Add(-1*time.Hour))
			},
		},
		{
			name:         "Database Error",
			sessionToken: sessionToken,
			wantToken:    false,
			wantErr:      true,
			errMsg:       "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			token, err := controllers.GenerateCSRFToken(db, tt.sessionToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCSRFToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GenerateCSRFToken() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if tt.wantToken && token == "" {
				t.Errorf("GenerateCSRFToken() token = %v, want non-empty token", token)
			}
		})
	}
}

func TestVerifyCSRFToken(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a session token and CSRF token for testing
	sessionToken := "test-session-token"
	csrfToken := "test-csrf-token"

	// Insert a valid session and CSRF token into the database
	_, err = db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
		sessionToken, csrfToken, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert test CSRF token: %v", err)
	}

	tests := []struct {
		name         string
		csrfToken    string
		sessionToken string
		wantValid    bool
		setup        func(db *sql.DB, r *http.Request)
	}{
		{
			name:         "Valid CSRF Token",
			csrfToken:    csrfToken,
			sessionToken: sessionToken,
			wantValid:    true,
			setup: func(db *sql.DB, r *http.Request) {
				// Insert a valid session into the sessions table
				_, err := db.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)",
					sessionToken, 1, time.Now().Add(1*time.Hour)) // Ensure the session is valid
				if err != nil {
					t.Fatalf("Failed to insert test session: %v", err)
				}

				// Set the CSRF token in the request header
				r.Header.Set("X-CSRF-Token", csrfToken)

				// Set the session token in the request cookie
				r.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
			},
		},
		{
			name:         "Invalid CSRF Token",
			csrfToken:    "invalid-token",
			sessionToken: sessionToken,
			wantValid:    false,
			setup: func(db *sql.DB, r *http.Request) {
				// Set an invalid CSRF token in the request header
				r.Header.Set("X-CSRF-Token", "invalid-token")
				// Set the session token in the request cookie
				r.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
			},
		},
		{
			name:         "Expired CSRF Token",
			csrfToken:    csrfToken,
			sessionToken: sessionToken,
			wantValid:    false,
			setup: func(db *sql.DB, r *http.Request) {
				// Insert an expired CSRF token into the database
				_, _ = db.Exec("INSERT OR REPLACE INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
					sessionToken, csrfToken, time.Now().Add(-1*time.Hour))
				// Set the CSRF token in the request header
				r.Header.Set("X-CSRF-Token", csrfToken)
				// Set the session token in the request cookie
				r.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
			},
		},
		{
			name:         "Missing CSRF Token",
			csrfToken:    "",
			sessionToken: sessionToken,
			wantValid:    false,
			setup: func(db *sql.DB, r *http.Request) {
				// Do not set the CSRF token in the request
				// Set the session token in the request cookie
				r.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
			},
		},
		{
			name:         "Missing Session Token",
			csrfToken:    csrfToken,
			sessionToken: "",
			wantValid:    false,
			setup: func(db *sql.DB, r *http.Request) {
				// Set the CSRF token in the request header
				r.Header.Set("X-CSRF-Token", csrfToken)
				// Do not set the session token in the request
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request
			r := httptest.NewRequest("POST", "/", nil)
			if tt.setup != nil {
				tt.setup(db, r)
			}

			valid := controllers.VerifyCSRFToken(db, r)
			if valid != tt.wantValid {
				t.Errorf("VerifyCSRFToken() valid = %v, wantValid %v", valid, tt.wantValid)
			}
		})
	}
}

func TestCleanupExpiredCSRFTokens(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Insert expired and non-expired CSRF tokens into the database
	_, err = db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
		"expired-session", "expired-token", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert expired CSRF token: %v", err)
	}
	_, err = db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
		"valid-session", "valid-token", time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert valid CSRF token: %v", err)
	}

	// Create a context with a timeout for the cleanup task
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run the cleanup task
	go controllers.CleanupExpiredCSRFTokens(ctx, db)

	// Wait for the cleanup task to run
	time.Sleep(1 * time.Second)

	// Verify that only the expired token was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM csrf_tokens WHERE session_token = 'expired-session'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query expired CSRF token: %v", err)
	}
	if count != 0 {
		t.Errorf("CleanupExpiredCSRFTokens() expired token count = %v, want 0", count)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM csrf_tokens WHERE session_token = 'valid-session'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query valid CSRF token: %v", err)
	}
	if count != 1 {
		t.Errorf("CleanupExpiredCSRFTokens() valid token count = %v, want 1", count)
	}
}

func TestAddCSRFToken(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test data
	sessionToken := "test-session-token"
	csrfToken := "test-csrf-token"
	expiresAt := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name         string
		sessionToken string
		csrfToken    string
		expiresAt    time.Time
		wantErr      bool
		errMsg       string
		setup        func(db *sql.DB)
	}{
		{
			name:         "Valid CSRF Token",
			sessionToken: sessionToken,
			csrfToken:    csrfToken,
			expiresAt:    expiresAt,
			wantErr:      false,
		},
		{
			name:         "Empty Session Token",
			sessionToken: "",
			csrfToken:    csrfToken,
			expiresAt:    expiresAt,
			wantErr:      true,
			errMsg:       "session token is empty",
		},
		{
			name:         "Database Error",
			sessionToken: sessionToken,
			csrfToken:    csrfToken,
			expiresAt:    expiresAt,
			wantErr:      true,
			errMsg:       "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := controllers.AddCSRFToken(db, tt.sessionToken, tt.csrfToken, tt.expiresAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCSRFToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("AddCSRFToken() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the token was added
			if !tt.wantErr {
				var storedToken string
				var storedExpiresAt time.Time
				err := db.QueryRow("SELECT csrf_token, expires_at FROM csrf_tokens WHERE session_token = ?", tt.sessionToken).
					Scan(&storedToken, &storedExpiresAt)
				if err != nil {
					t.Errorf("Failed to verify CSRF token addition: %v", err)
				}
				if storedToken != tt.csrfToken || !storedExpiresAt.Equal(tt.expiresAt) {
					t.Errorf("AddCSRFToken() storedToken = %v, storedExpiresAt = %v, want %v, %v",
						storedToken, storedExpiresAt, tt.csrfToken, tt.expiresAt)
				}
			}
		})
	}
}

func TestGetCSRFToken(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test data
	sessionToken := "test-session-token"
	csrfToken := "test-csrf-token"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Insert test data into the database
	_, err = db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
		sessionToken, csrfToken, expiresAt)
	if err != nil {
		t.Fatalf("Failed to insert test CSRF token: %v", err)
	}

	tests := []struct {
		name          string
		sessionToken  string
		wantToken     string
		wantExpiresAt time.Time
		wantErr       bool
		errMsg        string
		setup         func(db *sql.DB)
	}{
		{
			name:          "Valid Session Token",
			sessionToken:  sessionToken,
			wantToken:     csrfToken,
			wantExpiresAt: expiresAt,
			wantErr:       false,
		},
		{
			name:         "Invalid Session Token",
			sessionToken: "invalid-session-token",
			wantToken:    "",
			wantErr:      true,
			errMsg:       "sql: no rows in result set",
		},
		{
			name:         "Database Error",
			sessionToken: sessionToken,
			wantToken:    "",
			wantErr:      true,
			errMsg:       "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			token, expiresAt, err := controllers.GetCSRFToken(db, tt.sessionToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCSRFToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetCSRFToken() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}
			if token != tt.wantToken || !expiresAt.Equal(tt.wantExpiresAt) {
				t.Errorf("GetCSRFToken() token = %v, expiresAt = %v, want %v, %v",
					token, expiresAt, tt.wantToken, tt.wantExpiresAt)
			}
		})
	}
}

func TestDeleteCSRFToken(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Test data
	sessionToken := "test-session-token"
	csrfToken := "test-csrf-token"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Insert test data into the database
	_, err = db.Exec("INSERT INTO csrf_tokens (session_token, csrf_token, expires_at) VALUES (?, ?, ?)",
		sessionToken, csrfToken, expiresAt)
	if err != nil {
		t.Fatalf("Failed to insert test CSRF token: %v", err)
	}

	tests := []struct {
		name         string
		sessionToken string
		wantErr      bool
		errMsg       string
		setup        func(db *sql.DB)
	}{
		{
			name:         "Valid Session Token",
			sessionToken: sessionToken,
			wantErr:      false,
		},
		{
			name:         "Invalid Session Token",
			sessionToken: "invalid-session-token",
			wantErr:      false, // Deleting a non-existent token is not an error
		},
		{
			name:         "Database Error",
			sessionToken: sessionToken,
			wantErr:      true,
			errMsg:       "sql: database is closed",
			setup: func(db *sql.DB) {
				// Close the database connection to simulate a database error
				db.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the setup function if provided
			if tt.setup != nil {
				tt.setup(db)
			}

			err := controllers.DeleteCSRFToken(db, tt.sessionToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCSRFToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("DeleteCSRFToken() error = %v, wantErrMsg %v", err, tt.errMsg)
				}
				return
			}

			// Verify the token was deleted
			if !tt.wantErr {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM csrf_tokens WHERE session_token = ?", tt.sessionToken).
					Scan(&count)
				if err != nil {
					t.Errorf("Failed to verify CSRF token deletion: %v", err)
				}
				if count != 0 {
					t.Errorf("DeleteCSRFToken() count = %v, want 0", count)
				}
			}
		})
	}
}
