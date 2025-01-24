package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

// setupErrorTemplate creates a temporary error template file for testing
func setupErrorTemplate(t *testing.T) (string, func()) {
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

	// Create test errorPage.html
	errorContent := `
		<!DOCTYPE html>
		<html>
		<head><title>Error Page</title></head>
		<body>
			<h1>{{.ErrorTitle}}</h1>
			<p>{{.ErrorMessage}}</p>
			<p>Error Code: {{.ErrorCode}}</p>
		</body>
		</html>
	`
	err = os.WriteFile(filepath.Join(templatesDir, "errorPage.html"), []byte(errorContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create errorPage.html: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestServeErrorPage(t *testing.T) {
	// Setup test templates
	tempDir, cleanupTemplates := setupErrorTemplate(t)
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

	tests := []struct {
		name           string
		errorCode      int
		errorTitle     string
		errorMessage   string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Not Found Error",
			errorCode:      http.StatusNotFound,
			errorTitle:     "Page Not Found",
			errorMessage:   "The requested page could not be found.",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if !strings.Contains(body, "Page Not Found") {
					t.Error("Response should contain error title")
				}
				if !strings.Contains(body, "The requested page could not be found") {
					t.Error("Response should contain error message")
				}
				if !strings.Contains(body, "404") {
					t.Error("Response should contain error code")
				}
			},
		},
		{
			name:           "Internal Server Error",
			errorCode:      http.StatusInternalServerError,
			errorTitle:     "Internal Server Error",
			errorMessage:   "Something went wrong on our end.",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				body := rr.Body.String()
				if !strings.Contains(body, "Internal Server Error") {
					t.Error("Response should contain error title")
				}
				if !strings.Contains(body, "Something went wrong on our end") {
					t.Error("Response should contain error message")
				}
				if !strings.Contains(body, "500") {
					t.Error("Response should contain error code")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handlers.ServeErrorPage(rr, tt.errorCode, tt.errorTitle, tt.errorMessage)

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
