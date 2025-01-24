// auth/session.go
package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/database"
)

// Add cleanup helper function
func cleanupTestResources() {
	// Clean up log files
	os.RemoveAll("logs")
	// Clean up uploads directory if it exists
	os.RemoveAll("uploads")
	// Clean up entire storage directory with correct path, ensuring recursive removal
	storageDir := "./BackEnd/database/storage"
	if err := os.RemoveAll(storageDir); err != nil {
		// Log the error but don't fail the test
		fmt.Printf("Warning: Failed to remove storage directory: %v\n", err)
	}
}

// Add this new function after cleanupTestResources
func clearDatabaseTables(db *sql.DB) error {
	// List of tables to clear
	tables := []string{"users", "posts", "comments", "likes"}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear table %s: %v", table, err)
		}
	}
	return nil
}

func TestCreateSession(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear all tables before running tests
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a separate database connection for the error test
	errorDB, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create error test database: %v", err)
	}
	// Close it immediately to simulate database error
	errorDB.Close()

	type args struct {
		db     *sql.DB
		w      http.ResponseWriter
		userID int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid Session Creation",
			args: args{
				db:     db,
				w:      httptest.NewRecorder(),
				userID: 1,
			},
			wantErr: false,
		},
		{
			name: "Invalid User ID",
			args: args{
				db:     db,
				w:      httptest.NewRecorder(),
				userID: -1,
			},
			wantErr: true,
		},
		{
			name: "Database Error",
			args: args{
				db:     errorDB, // Use the closed database connection
				w:      httptest.NewRecorder(),
				userID: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateSession(tt.args.db, tt.args.w, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify cookie was set
				recorder := tt.args.w.(*httptest.ResponseRecorder)
				cookies := recorder.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == "session_token" {
						found = true
						break
					}
				}
				if !found {
					t.Error("CreateSession() did not set session cookie")
				}
			}
		})
	}
}

func TestDeleteSession(t *testing.T) {
	// Create a test database
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		cleanupTestResources()
	}()

	// Clear all tables before running tests
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a separate database connection for the error test
	errorDB, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to create error test database: %v", err)
	}
	// Close it immediately to simulate database error
	errorDB.Close()

	// Create a test session first
	w := httptest.NewRecorder()
	if err := CreateSession(db, w, 1); err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Get the session cookie
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session_token" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Failed to create session cookie for test")
	}

	type args struct {
		db     *sql.DB
		w      http.ResponseWriter
		cookie *http.Cookie
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid Session Deletion",
			args: args{
				db:     db,
				w:      httptest.NewRecorder(),
				cookie: sessionCookie,
			},
			wantErr: false,
		},
		{
			name: "Invalid Cookie",
			args: args{
				db:     db,
				w:      httptest.NewRecorder(),
				cookie: &http.Cookie{Name: "session_token", Value: "invalid_token"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteSession(tt.args.db, tt.args.w, tt.args.cookie)
			// Remove error checking since DeleteSession doesn't return an error

			if !tt.wantErr {
				// Verify cookie was invalidated
				recorder := tt.args.w.(*httptest.ResponseRecorder)
				cookies := recorder.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == "session_token" && cookie.MaxAge < 0 {
						found = true
						break
					}
				}
				if !found {
					t.Error("DeleteSession() did not properly invalidate session cookie")
				}
			}
		})
	}
}
