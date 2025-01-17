package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func CreateUserVoteHandler(lc *controllers.LikesController) http.HandlerFunc {
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

		// Handle the vote
		err = lc.HandleVote(postID, userID, userVote)
		if err != nil {
			logger.Error("Failed to handle vote: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		// After updating the post votes, fetch the updated likes count
		likesCount, dislikesCount, err := lc.GetPostVotes(postID)
		if err != nil {
			logger.Error("Failed to fetch updated votes for post %d: %v", postID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to fetch updated votes",
			})
			return
		}

		// Return the updated likes and dislikes count
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{
			"likes":    likesCount,
			"dislikes": dislikesCount,
		})
	}
}

func GetUserVotesHandler(lc *controllers.LikesController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is logged in
		loggedIn, userID := isLoggedIn(lc.DB, r)
		if !loggedIn {
			logger.Warning("Unauthorized attempt to GetUserVotesHandler - remote_addr: %s, method: %s, path: %s, user_id: %d",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				userID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Must be logged in to GetUserVotes",
			})
			return
		}

		userVote, err := lc.GetUserVotes(userID)
		if err != nil {
			logger.Error("Failed to get user votes: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to get user votes",
			})
			return
		}

		// Return the user's votes as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userVote)
	}
}
