package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func UploadFile(r *http.Request, formName string, userID int) (string, error) {
	var filePath string

	file, handler, err := r.FormFile(formName)
	if err != nil && err != http.ErrMissingFile {
		logger.Error("Failed to retrieve file %v", err)
		return "", err
	}

	if file != nil {
		defer file.Close()

		// Generate a unique filename
		timestamp := time.Now().Unix()
		fileExt := filepath.Ext(handler.Filename)
		newFilename := fmt.Sprintf("User%s_%d%s", string(userID), timestamp, fileExt)

		// Define the upload directory
		uploadDir := "uploads"
		filePath = filepath.Join(uploadDir, newFilename)

		// Create the upload directory if it doesn't exist
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			logger.Error("Failed to create upload directory: %v", err)
			return "", err
		}

		// Save the file to the server's filesystem
		dst, err := os.Create(filePath)
		if err != nil {
			logger.Error("Failed to create file on server: %v", err)
			return "", err
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			logger.Error("Failed to save file content: %v", err)
			return "", err
		}
	}
	return filePath, nil
}
