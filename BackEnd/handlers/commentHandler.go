package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

// CommentHandler handles requests for creating comments
func CommentHandler(cCtrl *controllers.CommentController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			log.Printf("Invalid method %s for comment creation attempt", r.Method)
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(cCtrl.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to create comment - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to create a comment",
			})
			return
		}

		// Extract postID from the URL path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 || pathParts[1] != "comment" {
			logger.Error("Invalid URL path: %s", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid URL path",
			})
			return
		}

		postIDStr := pathParts[2]
		postId, err := strconv.Atoi(postIDStr)
		if err != nil {
			logger.Error("Invalid postID: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid postID",
			})
			return
		}

		// Decode the request body into a CommentRequest object
		var commentReq models.CommentRequest
		if err := json.NewDecoder(r.Body).Decode(&commentReq); err != nil {
			logger.Error("Failed to decode comment request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input",
			})
			return
		}

		// Validate required fields
		if commentReq.Content == "" {
			logger.Warning("Invalid comment creation request: missing or empty content - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Content is required",
			})
			return
		}

		// Get the username for the logged-in user
		username := controllers.GetUsernameByID(cCtrl.DB, userID)

		// Create a Comment object from the CommentRequest
		comment := models.Comment{
			PostID:    postId,
			UserID:    userID,
			Author:    username,
			Content:   commentReq.Content,
			Likes:     0,
			Dislikes:  0,
			UserVote:  sql.NullString{String: "", Valid: false},
			Timestamp: time.Now(),
			ParentID:  sql.NullInt64{Int64: int64(commentReq.ParentID), Valid: commentReq.ParentID != 0},
		}

		// Insert the comment into the database
		commentID, err := cCtrl.InsertComment(comment)
		if err != nil {
			logger.Error("Failed to insert comment: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create comment",
			})
			return
		}

		// Return the created comment ID in the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{
			"commentID": commentID,
		})
	}
}

// DeleteCommentHandler handles requests for deleting comments
func DeleteCommentHandler(cCtrl *controllers.CommentController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is DELETE
		if r.Method != http.MethodDelete {
			log.Printf("Invalid method %s for comment deletion attempt", r.Method)
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(cCtrl.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to delete comment - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to delete a comment",
			})
			return
		}

		commentIDStr := r.URL.Query().Get("id")
		if commentIDStr == "" {
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

		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			logger.Error("Invalid commentID: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid commentID",
			})
			return
		}

		// Verify that the user is the author of the comment
		isAuthor, err := cCtrl.IsCommentAuthor(commentID, userID)
		if err != nil {
			logger.Error("Failed to verify comment author: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to verify comment author",
			})
			return
		}

		if !isAuthor {
			logger.Warning("Unauthorized attempt to delete comment - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "You are not authorized to delete this comment",
			})
			return
		}

		logger.Info("User %d is authorized to delete comment %d", userID, commentID)
		// Delete the comment
		err = cCtrl.DeleteComment(commentID)
		if err != nil {
			logger.Error("Failed to delete comment: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to delete comment",
			})
			return
		}

		logger.Info("Comment %d deleted successfully by user %d", commentID, userID)
		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Comment deleted successfully",
		})
	}
}

func UpdateCommentHandler(cc *controllers.CommentController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get comment ID from query parameters
		commentID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}

		// Check if user is logged in
		loggedIn, userID := isLoggedIn(cc.DB, r)
		if !loggedIn {
			http.Error(w, "Must be logged in to edit comment", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var updateReq struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify user is the comment author
		isAuthor, err := cc.IsCommentAuthor(commentID, userID)
		if err != nil {
			logger.Error("Failed to verify comment author: %v", err)
			http.Error(w, "Failed to verify comment author", http.StatusInternalServerError)
			return
		}

		if !isAuthor {
			http.Error(w, "Not authorized to edit this comment", http.StatusForbidden)
			return
		}

		// Update the comment
		err = cc.UpdateComment(commentID, updateReq.Content)
		if err != nil {
			logger.Error("Failed to update comment: %v", err)
			http.Error(w, "Failed to update comment", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Comment updated successfully",
		})
	}
}
