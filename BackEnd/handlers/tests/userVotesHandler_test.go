package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestCreateUserVoteHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	userID, err := controllers.NewAuthController(db).RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	sessionToken := "test_session"
	err = controllers.AddSession(db, sessionToken, int(userID), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Create a test post
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

	lc := controllers.NewLikesController(db)
	handler := handlers.CreateUserVoteHandler(lc)

	tests := []struct {
		name             string
		postID           string
		voteType         string
		sessionToken     string
		expectedStatus   int
		expectedLikes    int
		expectedDislikes int
		checkResponse    func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:             "Valid Like Vote",
			postID:           strconv.Itoa(postID),
			voteType:         "like",
			sessionToken:     sessionToken,
			expectedStatus:   http.StatusCreated,
			expectedLikes:    1,
			expectedDislikes: 0,
		},
		{
			name:             "Valid Dislike Vote",
			postID:           strconv.Itoa(postID),
			voteType:         "dislike",
			sessionToken:     sessionToken,
			expectedStatus:   http.StatusCreated,
			expectedLikes:    0,
			expectedDislikes: 1,
		},
		{
			name:           "Unauthorized - No Session",
			postID:         strconv.Itoa(postID),
			voteType:       "like",
			sessionToken:   "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Vote Type",
			postID:         strconv.Itoa(postID),
			voteType:       "invalid",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("post_id", tt.postID)
			form.Add("vote", tt.voteType)

			req, err := http.NewRequest("POST", "/post/vote", strings.NewReader(form.Encode()))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]int
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response["likes"] != tt.expectedLikes {
					t.Errorf("Expected %d likes, got %d", tt.expectedLikes, response["likes"])
				}
				if response["dislikes"] != tt.expectedDislikes {
					t.Errorf("Expected %d dislikes, got %d", tt.expectedDislikes, response["dislikes"])
				}
			}
		})
	}
}

func setupUserLikesTemplates(t *testing.T) (string, func()) {
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
		<head><title>User Likes</title></head>
		<body>
			{{template "content" .}}
		</body>
		</html>
	`
	err = os.WriteFile(filepath.Join(templatesDir, "layout.html"), []byte(layoutContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create layout.html: %v", err)
	}

	// Create homepage.html
	homepageContent := `
		{{define "content"}}
		{{range .Posts}}
			<div class="post">
				<h2>{{.Title}}</h2>
				<p>{{.Content}}</p>
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

func TestGetUserPostLikesHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Setup templates
	tempDir, cleanupTemplates := setupUserLikesTemplates(t)
	defer cleanupTemplates()

	// Set working directory
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
	lc := controllers.NewLikesController(db)

	post1 := models.Post{
		Title:     "Test Post 1",
		Content:   "Test Content 1",
		UserID:    int(userID),
		Author:    "testuser",
		Timestamp: time.Now(),
	}
	post1ID, err := pc.InsertPost(post1)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	err = lc.HandleVote(post1ID, int(userID), "like")
	if err != nil {
		t.Fatalf("Failed to add test vote: %v", err)
	}

	handler := handlers.GetUserPostLikesHandler(lc)

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid Request",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if !strings.Contains(body, "Test Post 1") {
					t.Error("Response should contain liked post title")
				}
				if !strings.Contains(body, "Test Content 1") {
					t.Error("Response should contain liked post content")
				}
			},
		},
		{
			name:           "Unauthorized - No Session",
			sessionToken:   "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if strings.Contains(body, "Test Post 1") {
					t.Error("Response should not contain liked post title for unauthorized user")
				}
				if strings.Contains(body, "Test Content 1") {
					t.Error("Response should not contain liked post content for unauthorized user")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/user/likes", nil)
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
