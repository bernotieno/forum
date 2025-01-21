package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func CreatePostHandler(pc *controllers.PostController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

// UpdatePostHandler handles PUT requests for updating a post
func UpdatePostHandler(pc *controllers.PostController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(pc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to update post - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to update post",
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

		// Extract post ID from URL
		postID := r.URL.Query().Get("id")
		if postID == "" {
			logger.Warning("Post Id Is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Extract form fields
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.FormValue("category")

		// Validate required fields
		if postID == "" || title == "" || categories == "" {
			logger.Error("Invalid post update request: missing or empty required fields - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Post ID, title, and categories are required",
			})
			return
		}

		// Convert postID to int
		postIDInt, err := strconv.Atoi(postID)
		if err != nil {
			logger.Error("Invalid post ID: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid post ID",
			})
			return
		}

		// Handle file upload (if a new file is provided)
		filePath, err := controllers.UploadFile(r, "post-file", userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to save file",
			})
			return
		}

		// Ensure at least one of content or file is provided
		if content == "" && filePath == "" {
			logger.Warning("Invalid post update request: missing content and image fields - at least one is required - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Missing content and image fields - at least one is required",
			})
			return
		}

		// Get the username of the logged-in user
		userName := controllers.GetUsernameByID(pc.DB, userID)

		// Create a Post object from the form data
		updatePost := models.Post{
			ID:        postIDInt,
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

		// Update the post in the database
		err = pc.UpdatePost(updatePost)
		if err != nil {
			logger.Error("Failed to update post: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to update post",
			})
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Post updated successfully",
		})
	}
}

// DeletePostHandler handles DELETE requests for deleting a post
func DeletePostHandler(pc *controllers.PostController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(pc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to delete post - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to delete post",
			})
			return
		}

		// Extract the post ID from the query parameters
		postIDStr := r.URL.Query().Get("id")
		if postIDStr == "" {
			logger.Warning("Missing post ID in delete request - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Post ID is required",
			})
			return
		}

		// Convert post ID to an integer
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			logger.Error("Invalid post ID: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid post ID",
			})
			return
		}

		// Verify that the user is the author of the Post
		isAuthor, err := pc.IsPostAuthor(postID, userID)
		if err != nil {
			logger.Error("Failed to verify Post author: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to verify Post author",
			})
			return
		}

		if !isAuthor {
			logger.Warning("Unauthorized attempt to delete Post - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "You are not authorized to delete this Post",
			})
			return
		}

		logger.Info("User %d is authorized to delete Post %d", userID, postID)
		// Call the controller to delete the post
		err = pc.DeletePost(postID, userID)
		if err != nil {
			logger.Error("Failed to delete post: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to delete post",
			})
			return
		}

		logger.Info("Post %d deleted successfully by user %d", postID, userID)
		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Post deleted successfully",
		})
	}
}
