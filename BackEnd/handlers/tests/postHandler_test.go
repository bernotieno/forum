package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func init() {
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

func createMultipartFormData(t *testing.T, fields map[string]string, file []byte) ([]byte, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		err := writer.WriteField(key, value)
		if err != nil {
			t.Fatalf("Failed to write form field: %v", err)
		}
	}

	if file != nil {
		part, err := writer.CreateFormFile("post-file", "test.jpg")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		if _, err := part.Write(file); err != nil {
			t.Fatalf("Failed to write file content: %v", err)
		}
	}

	writer.Close()
	return body.Bytes(), writer.FormDataContentType()
}

func TestCreatePostHandler(t *testing.T) {
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

	pc := controllers.NewPostController(db)
	handler := handlers.CreatePostHandler(pc)

	testFile := []byte("test image content")

	tests := []struct {
		name           string
		fields         map[string]string
		file           []byte
		sessionToken   string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid Post Creation",
			fields: map[string]string{
				"title":    "Test Post",
				"content":  "Test content",
				"category": "test",
			},
			file:           testFile,
			sessionToken:   sessionToken,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Unauthorized - No Session",
			fields: map[string]string{
				"title":    "Test Post",
				"content":  "Test content",
				"category": "test",
			},
			sessionToken:   "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Must be logged in to create post",
		},
		{
			name: "Missing Title",
			fields: map[string]string{
				"content":  "Test content",
				"category": "test",
			},
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Title and categories are required",
		},
		{
			name: "Missing Content and File",
			fields: map[string]string{
				"title":    "Test Post",
				"category": "test",
			},
			sessionToken:   sessionToken,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing content and image  fields  at least one is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, contentType := createMultipartFormData(t, tt.fields, tt.file)
			req, err := http.NewRequest("POST", "/create_post", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", contentType)

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
