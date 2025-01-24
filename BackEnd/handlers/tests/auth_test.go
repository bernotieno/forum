package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func TestRegisterHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	ac := controllers.NewAuthController(db)
	handler := handlers.RegisterHandler(ac)

	tests := []struct {
		name           string
		request        models.RegisterRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid Registration",
			request: models.RegisterRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "Password123!",
			},
			expectedStatus: http.StatusFound, // 302
		},
		{
			name: "Invalid Email",
			request: models.RegisterRequest{
				Email:    "invalid-email",
				Username: "testuser",
				Password: "Password123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name: "Invalid Username",
			request: models.RegisterRequest{
				Email:    "test@example.com",
				Username: "t", // too short
				Password: "Password123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Username must be between 3 and 20 characters",
		},
		{
			name: "Invalid Password",
			request: models.RegisterRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tt.expectedError != "" {
				if errMsg, exists := response["error"]; !exists || !strings.Contains(errMsg, tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, errMsg)
				}
			} else if redirect, exists := response["redirect"]; !exists || redirect != "/" {
				t.Errorf("Expected redirect to /, got %q", redirect)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	ac := controllers.NewAuthController(db)
	handler := handlers.LoginHandler(ac)

	// Create a test user first
	_, err := ac.RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid Login",
			username:       "testuser",
			password:       "Password123!",
			expectedStatus: http.StatusFound,
		},
		{
			name:           "Invalid Username",
			username:       "nonexistent",
			password:       "Password123!",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid username or password",
		},
		{
			name:           "Invalid Password",
			username:       "testuser",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid username or password",
		},
		{
			name:           "Missing Fields",
			username:       "",
			password:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(map[string]string{
				"username": tt.username,
				"password": tt.password,
			})
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tt.expectedError != "" {
				if errMsg, exists := response["error"]; !exists || errMsg != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, errMsg)
				}
			} else if redirect, exists := response["redirect"]; !exists || redirect != "/" {
				t.Errorf("Expected redirect to /, got %q", redirect)
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	database.GloabalDB = db

	// Create a test user and session
	ac := controllers.NewAuthController(db)
	userID, err := ac.RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	sessionToken := "test_session"
	err = controllers.AddSession(db, sessionToken, int(userID), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
	}{
		{
			name:           "Valid Logout",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No Session",
			sessionToken:   "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/logout", nil)
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tt.sessionToken,
				})
			}

			rr := httptest.NewRecorder()
			handlers.LogoutHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				// Check if cookie was cleared
				cookies := rr.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == "session_token" {
						if cookie.MaxAge != -1 {
							t.Error("Session cookie was not properly expired")
						}
						found = true
						break
					}
				}
				if !found {
					t.Error("No session cookie found in response")
				}
			}
		})
	}
}
