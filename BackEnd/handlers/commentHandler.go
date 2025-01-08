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
		loggedIn, userID := isLoggedIn(r)
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
			UserVote:  sql.NullString{String: "", Valid: false}, // Default to no vote
			Timestamp: time.Now(),
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
