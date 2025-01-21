package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

func CreateUserVoteHandler(lc *controllers.LikesController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

func GetUserPostLikesHandler(lc *controllers.LikesController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		// Check if user is logged in
		loggedIn, userID := isLoggedIn(lc.DB, r)
		var csrfToken string
		if loggedIn {
			sessionToken, err := controllers.GetSessionToken(r)
			if err != nil {
				logger.Error("Error getting session token: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Generate CSRF token for the session
			csrfToken, err = controllers.GenerateCSRFToken(lc.DB, sessionToken)
			if err != nil {
				logger.Error("Error generating CSRF token: %V", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		// Fetch posts created by the logged-in user
		userPosts, err := lc.GetUserLikesPosts(userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Add IsAuthor field to each post and fetch comment count
		commentController := controllers.NewCommentController(lc.DB)
		for i := range userPosts {
			userPosts[i].IsAuthor = loggedIn && userPosts[i].UserID == userID

			// Fetch total comment count including replies
			commentCount, err := commentController.GetCommentCountByPostID(userPosts[i].ID)
			if err != nil {
				logger.Error("Failed to fetch comment count for post %d: %v", userPosts[i].ID, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			userPosts[i].Comments = make([]models.Comment, 0)
			userPosts[i].CommentCount = commentCount
		}

		// Create template function map
		funcMap := template.FuncMap{
			"formatTime": func(t time.Time) string {
				return t.Format("Jan 02, 2006 at 15:04")
			},
			"split": strings.Split,
			"trim":  strings.TrimSpace,
		}

		// Create template with function map
		tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(
			"./FrontEnd/templates/layout.html",
			"./FrontEnd/templates/homepage.html",
		)
		if err != nil {
			logger.Error("An Error Occured while Rendering template %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Prepare data for the template
		data := struct {
			IsAuthenticated bool
			CSRFToken       string
			Posts           []models.Post
			UserID          int
		}{
			IsAuthenticated: loggedIn,
			CSRFToken:       csrfToken,
			Posts:           userPosts, // Only the user's posts
			UserID:          userID,
		}

		// Execute template with data
		err = tmpl.ExecuteTemplate(w, "layout.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
