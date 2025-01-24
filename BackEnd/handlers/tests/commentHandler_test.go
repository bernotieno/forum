package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func init() {
	// Initialize the logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}
func TestCommentHandler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Clear all tables before running tests
	if err := clearDatabaseTables(db); err != nil {
		t.Fatalf("Failed to clear database tables: %v", err)
	}

	// Create a test user and session
	userID, err := controllers.NewAuthController(db).RegisterUser("test@example.com", "testuser", "Password123!")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test session
	sessionToken := "test_session"
	err = controllers.AddSession(db, sessionToken, int(userID), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	cCtrl := controllers.NewCommentController(db)
	handler := handlers.CommentHandler(cCtrl)

	tests := []struct {
		name           string
		postID         string
		payload        models.CommentRequest
		sessionToken   string
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "Valid Comment Creation",
			postID: "1",
			payload: models.CommentRequest{
				Content: "Test comment content",
			},
			sessionToken:   sessionToken,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Unauthorized - No Session",
			postID: "1",
			payload: models.CommentRequest{
				Content: "Test comment content",
			},
			sessionToken:   "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Must be logged in to create a comment",
		},
		{
			name:   "Invalid Post ID",
			postID: "invalid",
			payload: models.CommentRequest{
				Content: "Test comment content",
			},
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid postID",
		},
		{
			name:   "Empty Content",
			postID: "1",
			payload: models.CommentRequest{
				Content: "",
			},
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			req, err := http.NewRequest("POST", "/comment/"+tt.postID, bytes.NewBuffer(payloadBytes))
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

			if tt.expectedError != "" {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if message, exists := response["message"]; exists && message != tt.expectedError {
					t.Errorf("Expected error message %s, got %s", tt.expectedError, message)
				}
				if errMsg, exists := response["error"]; exists && errMsg != tt.expectedError {
					t.Errorf("Expected error message %s, got %s", tt.expectedError, errMsg)
				}
			}
		})
	}
}

func TestDeleteCommentHandler(t *testing.T) {
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

	handler := handlers.DeleteCommentHandler(cCtrl)

	tests := []struct {
		name           string
		commentID      string
		sessionToken   string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid Comment Deletion",
			commentID:      strconv.Itoa(commentID),
			sessionToken:   sessionToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized - No Session",
			commentID:      strconv.Itoa(commentID),
			sessionToken:   "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Must be logged in to delete a comment",
		},
		{
			name:           "Invalid Comment ID",
			commentID:      "invalid",
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid commentID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("DELETE", "/comment/delete?id="+tt.commentID, nil)
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

			if tt.expectedError != "" {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if message, exists := response["message"]; exists && message != tt.expectedError {
					t.Errorf("Expected error message %s, got %s", tt.expectedError, message)
				}
			}
		})
	}
}
