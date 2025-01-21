package Test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
)

func TestUploadFile(t *testing.T) {
	// Create a temporary directory for uploads
	uploadDir := "test_uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create upload directory: %v", err)
	}
	defer os.RemoveAll(uploadDir) // Clean up after the test

	// Test cases
	tests := []struct {
		name         string
		formName     string
		fileName     string
		fileContent  string
		userID       int
		wantFilePath string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Successful File Upload",
			formName:     "file",
			fileName:     "test.txt",
			fileContent:  "Hello, World!",
			userID:       1,
			wantFilePath: "/uploads/User1_", // Partial match for timestamp
			wantErr:      false,
		},
		{
			name:         "No File Uploaded",
			formName:     "file",
			fileName:     "",
			fileContent:  "",
			userID:       1,
			wantFilePath: "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a multipart form with a file
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if tt.fileName != "" {
				part, err := writer.CreateFormFile(tt.formName, tt.fileName)
				if err != nil {
					t.Fatalf("Failed to create form file: %v", err)
				}
				_, err = io.WriteString(part, tt.fileContent)
				if err != nil {
					t.Fatalf("Failed to write file content: %v", err)
				}
			}

			writer.Close()

			// Create an HTTP request with the multipart form
			req := httptest.NewRequest(http.MethodPost, "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Call the UploadFile function
			filePath, err := controllers.UploadFile(req, tt.formName, tt.userID)

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadFile() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("UploadFile() error = %v, wantErrMsg = %v", err, tt.errMsg)
				return
			}

			// Verify the file path
			if !tt.wantErr && tt.wantFilePath != "" {
				if !strings.Contains(filePath, tt.wantFilePath) {
					t.Errorf("UploadFile() filePath = %v, wantFilePath containing %v", filePath, tt.wantFilePath)
				}

				// Verify the file was saved correctly
				cleanedPath := strings.TrimPrefix(filePath, "/")
				_, err := os.Stat(cleanedPath)
				if err != nil {
					t.Errorf("UploadFile() failed to save file: %v", err)
				}
			}
		})
	}
}

func TestRemoveImages(t *testing.T) {
	// Create a temporary directory for test files
	testDir := "test_images"
	if err := os.MkdirAll(testDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir) // Clean up after the test

	// Create test files
	file1 := filepath.Join(testDir, "image1.jpg")
	file2 := filepath.Join(testDir, "image2.jpg")
	if err := os.WriteFile(file1, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test cases
	tests := []struct {
		name       string
		imagePaths []string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Successful File Deletion",
			imagePaths: []string{file1, file2},
			wantErr:    false,
		},
		{
			name:       "Non-Existent File",
			imagePaths: []string{"non_existent_file.jpg"},
			wantErr:    false, // No error expected for non-existent files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controllers.RemoveImages(tt.imagePaths)

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("removeImages() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("removeImages() error = %v, wantErrMsg = %v", err, tt.errMsg)
				return
			}

			// Verify files were deleted
			if !tt.wantErr {
				for _, path := range tt.imagePaths {
					cleanedPath := strings.TrimPrefix(path, "/")
					if _, err := os.Stat(cleanedPath); !os.IsNotExist(err) {
						t.Errorf("removeImages() failed to delete file: %v", cleanedPath)
					}
				}
			}
		})
	}
}
