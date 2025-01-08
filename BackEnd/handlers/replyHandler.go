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

// ReplyHandler handles requests for creating replies to comments
func ReplyHandler(rCtrl *controllers.ReplyController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			log.Printf("Invalid method %s for reply creation attempt", r.Method)
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to create reply - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to create a reply",
			})
			return
		}

		// Extract commentID from the URL path
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 || pathParts[1] != "reply" {
			logger.Error("Invalid URL path: %s", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid URL path",
			})
			return
		}

		commentIDStr := pathParts[2]
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

		// Decode the request body into a ReplyRequest object
		var replyReq models.ReplyRequest
		if err := json.NewDecoder(r.Body).Decode(&replyReq); err != nil {
			logger.Error("Failed to decode reply request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input",
			})
			return
		}

		// Validate required fields
		if replyReq.Content == "" {
			logger.Warning("Invalid reply creation request: missing or empty content - remote_addr: %s, method: %s, path: %s",
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
		username := controllers.GetUsernameByID(rCtrl.DB, userID)

		// Create a Reply object from the ReplyRequest
		reply := models.Reply{
			CommentID: commentID,
			UserID:    userID,
			Author:    username,
			Content:   replyReq.Content,
			Likes:     0,
			Dislikes:  0,
			UserVote:  sql.NullString{String: "", Valid: false}, // Default to no vote
			Timestamp: time.Now(),
		}

		// Insert the reply into the database
		replyID, err := rCtrl.InsertReply(reply)
		if err != nil {
			logger.Error("Failed to insert reply: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create reply",
			})
			return
		}

		// Return the created reply ID in the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{
			"replyID": replyID,
		})
	}
}
