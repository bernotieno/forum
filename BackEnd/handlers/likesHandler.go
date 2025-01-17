package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func CreateLikeHandler(lc *controllers.LikesController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			logger.Error("Invalid method %s for Like attempt in Like Handler", r.Method)
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(lc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to create like - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to like a post",
			})
			return
		}

		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			logger.Error("Failed to parse form data: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to parse form data",
			})
			return
		}

		// Extract form fields
		postIDStr := r.FormValue("post_id")
		userVote := r.FormValue("vote")

		// Validate required fields
		if postIDStr == "" || userVote == "" {
			logger.Warning("Invalid like creation request: missing or empty required fields - remote_addr: %s, method: %s, path: %s",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Post ID and vote are required",
			})
			return
		}

		// Convert post_id to an integer
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			logger.Error("Invalid post_id: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid post ID format",
			})
			return
		}

		// Validate the vote value
		if userVote != "like" && userVote != "dislike" {
			logger.Warning("Invalid vote value: %s - remote_addr: %s, method: %s, path: %s",
				userVote,
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Vote must be either 'like' or 'dislike'",
			})
			return
		}

		// Create a Like object
		newLike := models.Likes{
			PostId:   postID,
			UserId:   userID,
			UserVote: userVote,
		}

		// Insert the like into the database
		likeID, err := lc.InsertLikes(newLike)
		if err != nil {
			logger.Error("Failed to insert like: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create like",
			})
			return
		}

		// Update the likes and dislikes in the posts table
		err = lc.UpdatePostVotes(postID)
		if err != nil {
			logger.Error("Failed to update votes for post %d: %v", postID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to update post votes",
			})
			return
		}

		// Return the created like ID in the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{
			"likeID": likeID,
		})
	}
}
