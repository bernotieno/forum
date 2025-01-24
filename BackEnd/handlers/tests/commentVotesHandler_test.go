package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func TestCreateCommentVoteHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Clear all tables before running tests
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create test user and session
	userID, err := controllers.NewAuthController(db).RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test session
	sessionToken := "test_session"
	err = controllers.AddSession(db, sessionToken, int(userID), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Create a test comment
	cCtrl := controllers.NewCommentController(db)
	comment := models.Comment{
		PostID:    1,
		UserID:    int(userID),
		Author:    "testuser",
		Content:   "Test comment",
		Timestamp: time.Now(),
	}
	commentID, err := cCtrl.InsertComment(comment)
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	cvCtrl := controllers.NewCommentVotesController(db)
	handler := handlers.CreateCommentVoteHandler(cvCtrl)

	tests := []struct {
		name             string
		commentID        string
		voteType         string
		sessionToken     string
		expectedStatus   int
		expectedLikes    int
		expectedDislikes int
	}{
		{
			name:             "Valid Like Vote",
			commentID:        strconv.Itoa(commentID),
			voteType:         "like",
			sessionToken:     sessionToken,
			expectedStatus:   http.StatusOK,
			expectedLikes:    1,
			expectedDislikes: 0,
		},
		{
			name:             "Valid Dislike Vote",
			commentID:        strconv.Itoa(commentID),
			voteType:         "dislike",
			sessionToken:     sessionToken,
			expectedStatus:   http.StatusOK,
			expectedLikes:    0,
			expectedDislikes: 1,
		},
		{
			name:           "Invalid Vote Type",
			commentID:      strconv.Itoa(commentID),
			voteType:       "invalid",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized - No Session",
			commentID:      strconv.Itoa(commentID),
			voteType:       "like",
			sessionToken:   "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Comment ID",
			commentID:      "invalid",
			voteType:       "like",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("comment_id", tt.commentID)
			form.Add("vote", tt.voteType)

			req, err := http.NewRequest("POST", "/comment/vote", strings.NewReader(form.Encode()))
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

			if tt.expectedStatus == http.StatusOK {
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

func TestGetUserCommentVotesHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Clear all tables before running tests
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

	// Create test comments and votes
	cCtrl := controllers.NewCommentController(db)
	cvCtrl := controllers.NewCommentVotesController(db)

	comment1 := models.Comment{
		PostID:    1,
		UserID:    int(userID),
		Author:    "testuser",
		Content:   "Test comment 1",
		Timestamp: time.Now(),
	}
	comment1ID, err := cCtrl.InsertComment(comment1)
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	// Add some votes
	err = cvCtrl.HandleCommentVote(comment1ID, int(userID), "like")
	if err != nil {
		t.Fatalf("Failed to add test vote: %v", err)
	}

	handler := handlers.GetUserCommentVotesHandler(cvCtrl)

	tests := []struct {
		name           string
		sessionToken   string
		expectedStatus int
		expectedVotes  map[string]string
	}{
		{
			name:           "Valid Request",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusOK,
			expectedVotes: map[string]string{
				strconv.Itoa(comment1ID): "like",
			},
		},
		{
			name:           "Unauthorized - No Session",
			sessionToken:   "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/comment/votes", nil)
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

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !reflect.DeepEqual(response, tt.expectedVotes) {
					t.Errorf("Expected votes %v, got %v", tt.expectedVotes, response)
				}
			}
		})
	}
}
