package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func setupViewPostTemplates(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "forum-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	templatesDir := filepath.Join(tempDir, "FrontEnd", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create layout.html
	layoutContent := `
        <!DOCTYPE html>
        <html>
        <head><title>View Post</title></head>
        <body>
            {{template "content" .}}
        </body>
        </html>
    `
	err = os.WriteFile(filepath.Join(templatesDir, "layout.html"), []byte(layoutContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create layout.html: %v", err)
	}

	// Create viewPost.html
	viewPostContent := `
        {{define "content"}}
        <div class="post">
            <h1>{{.Post.Title}}</h1>
            <p>{{.Post.Content}}</p>
            <p>By {{.Post.Author}}</p>
            {{if .IsAuthor}}
            <div class="author-controls">Edit Delete</div>
            {{end}}
            {{range .Comments}}
            <div class="comment">{{.Content}}</div>
            {{end}}
        </div>
        {{end}}
    `
	err = os.WriteFile(filepath.Join(templatesDir, "viewPost.html"), []byte(viewPostContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create viewPost.html: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestViewPostHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	tempDir, cleanupTemplates := setupViewPostTemplates(t)
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

	// Create test user and post
	userID, err := controllers.NewAuthController(db).RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	sessionToken := "test_session"
	err = controllers.AddSession(db, sessionToken, int(userID), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	pc := controllers.NewPostController(db)
	testPost := models.Post{
		Title:     "Test Post",
		Content:   "Test Content",
		UserID:    int(userID),
		Author:    "testuser",
		Timestamp: time.Now(),
	}
	postID, err := pc.InsertPost(testPost)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Add test comment
	cc := controllers.NewCommentController(db)
	testComment := models.Comment{
		Content:   "Test Comment",
		PostID:    postID,
		UserID:    int(userID),
		Author:    "testuser",
		Timestamp: time.Now(),
	}
	_, err = cc.InsertComment(testComment)
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	handler := handlers.NewViewPostHandler(db)

	tests := []struct {
		name           string
		postID         string
		sessionToken   string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid Post View - Authenticated Author",
			postID:         strconv.Itoa(postID),
			sessionToken:   sessionToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if !strings.Contains(body, "Test Post") {
					t.Error("Response should contain post title")
				}
				if !strings.Contains(body, "Test Content") {
					t.Error("Response should contain post content")
				}
				if !strings.Contains(body, "author-controls") {
					t.Error("Response should contain author controls")
				}
				if !strings.Contains(body, "Test Comment") {
					t.Error("Response should contain comment")
				}
			},
		},
		{
			name:           "Valid Post View - Unauthenticated",
			postID:         strconv.Itoa(postID),
			sessionToken:   "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if !strings.Contains(body, "Test Post") {
					t.Error("Response should contain post title")
				}
				if strings.Contains(body, "author-controls") {
					t.Error("Response should not contain author controls")
				}
			},
		},
		{
			name:           "Invalid Post ID",
			postID:         "",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/view_post?id="+tt.postID, nil)
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
