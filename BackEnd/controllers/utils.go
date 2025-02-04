package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func UploadFile(r *http.Request, fieldName string, userID int) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // No file uploaded, not an error
		}
		return "", fmt.Errorf("error retrieving file: %v", err)
	}
	defer file.Close()

	// Check file size (20MB limit)
	if header.Size > 20 * 1024 * 1024 { 
		return "", fmt.Errorf("file size exceeds 20MB limit")
	}

	// Read first 512 bytes to detect content type
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return "", fmt.Errorf("error reading file header: %v", err)
	}

	// Reset file pointer
	file.Seek(0, 0)

	// Check file type
	contentType := http.DetectContentType(buff)
	allowedTypes := map[string]string{
		"image/jpeg":    ".jpg",
		"image/png":     ".png",
		"image/gif":     ".gif",
		"image/svg+xml": ".svg",
	}

	extension, allowed := allowedTypes[contentType]
	if !allowed {
		return "", fmt.Errorf("invalid file type. Only JPEG, PNG, GIF and SVG files are allowed")
	}

	// Create unique filename
	filename := fmt.Sprintf("%d_%s%s", userID, time.Now().Format("20060102150405"), extension)
	uploadDir := "uploads/posts" // Configure your upload directory

	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", fmt.Errorf("error creating upload directory: %v", err)
	}

	// Create new file
	filepath := path.Join(uploadDir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("error creating file: %v", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("error saving file: %v", err)
	}

	return filepath, nil
}

func RemoveImages(imagePaths []string) error {
	for _, imagePath := range imagePaths {
		// Remove the "uploads/" prefix if it exists in the imagePath
		cleanedPath := strings.TrimPrefix(imagePath, "/")

		// Check if the file exists before attempting to delete it
		if _, err := os.Stat(cleanedPath); os.IsNotExist(err) {
			return nil
		}

		// Delete the file
		if err := os.Remove(cleanedPath); err != nil {
			return fmt.Errorf("failed to delete image file %s: %w", imagePath, err)
		}

	}
	return nil
}
