package tests

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func init() {
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

func setupLoginTemplates(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "forum-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	templatesDir := filepath.Join(tempDir, "FrontEnd", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	loginContent := `
        <!DOCTYPE html>
        <html>
        <head><title>Login Page</title></head>
        <body>
            <form class="login-form">
                <input type="email" name="email" required>
                <input type="password" name="password" required>
                <button type="submit">Login</button>
            </form>
        </body>
        </html>
    `
	err = os.WriteFile(filepath.Join(templatesDir, "login.html"), []byte(loginContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create login.html: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestLoginPageHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	tempDir, cleanupTemplates := setupLoginTemplates(t)
	defer cleanupTemplates()

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	database.GloabalDB = db

	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create test user and session for logged-in test case
	userID, err := controllers.NewAuthController(db).RegisterUser("test@example.com", "testuser", "Password123!")
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
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Not Logged In - Show Login Page",
			sessionToken:   "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if !strings.Contains(body, "login-form") {
					t.Error("Response should contain login form")
				}
				if !strings.Contains(body, `type="email"`) {
					t.Error("Response should contain email input")
				}
				if !strings.Contains(body, `type="password"`) {
					t.Error("Response should contain password input")
				}
			},
		},
		{
			name:           "Already Logged In - Redirect to Home",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				if location != "/" {
					t.Errorf("Expected redirect to /, got %s", location)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/login_Page", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tt.sessionToken,
				})
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handlers.LoginPageHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
