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

// setupTestTemplates creates temporary template files for testing
func setupTestTemplates(t *testing.T) (string, func()) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "forum-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create FrontEnd/templates directory structure
	templatesDir := filepath.Join(tempDir, "FrontEnd", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create test layout.html
	layoutContent := `
		<!DOCTYPE html>
		<html>
		<head><title>Test Layout</title></head>
		<body>
			{{template "content" .}}
		</body>
		</html>
	`
	err = os.WriteFile(filepath.Join(templatesDir, "layout.html"), []byte(layoutContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create layout.html: %v", err)
	}

	// Create test post.html
	postContent := `
		{{define "content"}}
		<form>
			<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
			<!-- Add other form elements -->
		</form>
		{{end}}
	`
	err = os.WriteFile(filepath.Join(templatesDir, "post.html"), []byte(postContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create post.html: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestCreatePostPageHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup test templates
	tempDir, cleanupTemplates := setupTestTemplates(t)
	defer cleanupTemplates()

	// Set the working directory to the temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}

	// Set the global DB for the handler
	database.GloabalDB = db

	// Clear all tables before running tests
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a test user and session
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
			name:           "Successful Page Load",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				// Check if the response contains expected template elements
				if !strings.Contains(body, "csrf_token") {
					t.Error("Response should contain CSRF token")
				}
				if !strings.Contains(body, "form") {
					t.Error("Response should contain a form")
				}
			},
		},
		{
			name:           "Unauthorized - No Session",
			sessionToken:   "",
			expectedStatus: http.StatusSeeOther, // 303 redirect to login page
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				if location != "/login_Page" {
					t.Errorf("Expected redirect to /login_Page, got %s", location)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/create_post", nil)
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
			handler := http.HandlerFunc(handlers.CreatePostPageHandler)
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
