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
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func init() {
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

func setupHomePageTemplates(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "forum-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	templatesDir := filepath.Join(tempDir, "FrontEnd", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

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

	homepageContent := `
        {{define "content"}}
        {{if .IsAuthenticated}}
            <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
        {{end}}
        {{range .Posts}}
            <div class="post">
                <h2>{{.Title}}</h2>
                <p>{{.Content}}</p>
                <span>Comments: {{.CommentCount}}</span>
            </div>
        {{end}}
        {{end}}
    `
	err = os.WriteFile(filepath.Join(templatesDir, "homepage.html"), []byte(homepageContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create homepage.html: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestHomePageHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	tempDir, cleanupTemplates := setupHomePageTemplates(t)
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

	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create test user and session
	userID, err := controllers.NewAuthController(db).RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	sessionToken := "test_session"
	err = controllers.AddSession(db, sessionToken, int(userID), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Create test posts
	pc := controllers.NewPostController(db)
	testPost := models.Post{
		Title:     "Test Post",
		Content:   "Test Content",
		UserID:    int(userID),
		Author:    "testuser",
		Timestamp: time.Now(),
	}
	_, err = pc.InsertPost(testPost)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Authenticated User View",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if !strings.Contains(body, "csrf_token") {
					t.Error("Response should contain CSRF token for authenticated user")
				}
				if !strings.Contains(body, "Test Post") {
					t.Error("Response should contain test post title")
				}
				if !strings.Contains(body, "Test Content") {
					t.Error("Response should contain test post content")
				}
			},
		},
		{
			name:           "Unauthenticated User View",
			sessionToken:   "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if strings.Contains(body, "csrf_token") {
					t.Error("Response should not contain CSRF token for unauthenticated user")
				}
				if !strings.Contains(body, "Test Post") {
					t.Error("Response should contain test post title")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
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
			handler := handlers.NewHomePageHandler(db)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if contentType := rr.Header().Get("Content-Type"); contentType != "text/html" {
				t.Errorf("Expected Content-Type text/html, got %s", contentType)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
