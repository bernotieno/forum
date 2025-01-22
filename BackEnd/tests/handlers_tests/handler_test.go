package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
	_ "github.com/mattn/go-sqlite3"
)

func TestMain(m *testing.M) {
	// Initialize the logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Run the tests
	code := m.Run()
	os.Exit(code)
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := database.Init("Test")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return db
}

func teardownTestDB(t *testing.T, db *sql.DB) {
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close test database: %v", err)
	}
}

func TestRegisterHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ac := controllers.NewAuthController(db)
	handler := handlers.RegisterHandler(ac)

	tests := []struct {
		name           string
		payload        models.RegisterRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid Registration",
			payload: models.RegisterRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "Password123!",
			},
			expectedStatus: http.StatusFound,
			expectedError:  "",
		},
		// Add other test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(payloadBytes))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error %s, got %s", tt.expectedError, response["error"])
				}
			}
		})
	}
}
func TestLoginHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ac := controllers.NewAuthController(db)
	handler := handlers.LoginHandler(ac)

	// Register a user first
	_, err := ac.RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid Login",
			payload: map[string]string{
				"username": "testuser",
				"password": "Password123!",
			},
			expectedStatus: http.StatusFound,
			expectedError:  "",
		},
		{
			name: "Invalid Username",
			payload: map[string]string{
				"username": "wronguser",
				"password": "Password123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid username",
		},
		{
			name: "Invalid Password",
			payload: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid password",
		},
		{
			name: "Missing Fields",
			payload: map[string]string{
				"username": "",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(payloadBytes))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedError != "" {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error %s, got %s", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestCheckLoginHandler(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    ac := controllers.NewAuthController(db)
    handler := http.HandlerFunc(handlers.CheckLoginHandler)

    // Register and login a user first
    userID, err := ac.RegisterUser("test@example.com", "testuser", "Password123!")
    if err != nil {
        t.Fatalf("Failed to register user: %v", err)
    }

    // Create a session for the user
    sessionToken := "test-session-token"
    expiresAt := "2023-12-31T23:59:59Z"
    _, err = db.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)", sessionToken, userID, expiresAt)
    if err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }

    // Debug: Check if the session was inserted
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_token = ?", sessionToken).Scan(&count)
    if err != nil {
        t.Fatalf("Failed to query session: %v", err)
    }
    if count == 0 {
        t.Fatalf("Session was not inserted into the database")
    }

    // Set the session cookie
    req, err := http.NewRequest("GET", "/check-login", nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }
    req.AddCookie(&http.Cookie{
        Name:  "session_token",
        Value: sessionToken,
    })

    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
    }

    var response map[string]bool
    if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
        t.Fatalf("Failed to unmarshal response: %v", err)
    }

    if !response["loggedIn"] {
        t.Errorf("Expected loggedIn to be true, got false")
    }
}
func TestLogoutHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ac := controllers.NewAuthController(db)
	handler := http.HandlerFunc(handlers.LogoutHandler)

	// Register and login a user first
	userID, err := ac.RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a session for the user
	sessionToken := "test-session-token"
	expiresAt := "2023-12-31T23:59:59Z"
	_, err = db.Exec("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)", sessionToken, userID, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Set the session cookie
	req.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify that the session was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_token = ?", sessionToken).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected session to be deleted, but it still exists")
	}
}