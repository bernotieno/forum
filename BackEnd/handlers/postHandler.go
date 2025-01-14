package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func PostHandler(pc *controllers.PostController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			logger.Error("Invalid method %s for Post attempt in Post Handler", r.Method)
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(pc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to create post - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to create post",
			})
			return
		}

		if !controllers.VerifyCSRFToken(pc.DB, r) {
			logger.Warning("Invalid CSRF token in post attempt")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid CSRF token",
			})
			return
		}

		// Parse the multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
		if err != nil {
			logger.Error("Failed to parse multipart form: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to parse form data",
			})
			return
		}

		// Extract form fields
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.FormValue("category")

		// Validate required fields
		if title == "" || categories == "" {
			logger.Warning("Invalid post creation request: missing or empty required fields - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Title and categories are required",
			})
			return
		}

		// Handle file upload
		filePath, err := controllers.UploadFile(r, "post-file", userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to save file",
			})
			return
		}

		if content == "" && filePath == "" {
			logger.Warning("Invalid post creation request: missing content and image  fields  at least one is required - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Missing content and image  fields  at least one is required",
			})
			return
		}

		userName := controllers.GetUsernameByID(pc.DB, userID)

		// Create a Post object from the form data
		createPost := models.Post{
			Title:     title,
			Author:    userName,
			UserID:    userID,
			Category:  categories,
			Content:   content,
			Timestamp: time.Now(),
			ImageUrl: sql.NullString{
				String: filePath,
				Valid:  filePath != "",
			},
		}

		// Insert the post into the database
		postID, err := pc.InsertPost(createPost)
		if err != nil {
			logger.Error("Failed to insert post: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create post",
			})
			return
		}

		// Return the created post ID in the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{
			"postID": postID,
		})
	}
}
