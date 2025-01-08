package handlers

import (
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
		loggedIn, userID := isLoggedIn(r)
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

		// Decode the request body into a PostRequest object
		var postReq models.PostRequest
		if err := json.NewDecoder(r.Body).Decode(&postReq); err != nil {
			logger.Error("Failed to decode post request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid input",
			})
			return
		}

		// Validate required fields
		if postReq.Title == "" || postReq.Content == "" || len(postReq.Categories) == 0 {
			logger.Warning("Invalid post creation request: missing or empty required fields - remote_addr: %s, method: %s, path: %s, missing_fields: %v",
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				getMissingFields(postReq),
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Title, content, and categories are required",
			})
			return
		}

		userName := controllers.GetUsernameByID(pc.DB, userID)

		// Create a Post object from the PostRequest
		createPost := models.Post{
			Title:     postReq.Title,
			Author:    userName,
			UserID:    userID,
			Category:  postReq.Categories,
			Content:   postReq.Content,
			Timestamp: time.Now(),
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

// Helper function to get missing fields
func getMissingFields(postReq models.PostRequest) []string {
	var missingFields []string
	if postReq.Title == "" {
		missingFields = append(missingFields, "title")
	}
	if postReq.Content == "" {
		missingFields = append(missingFields, "content")
	}
	if len(postReq.Categories) == 0 {
		missingFields = append(missingFields, "categories")
	}
	return missingFields
}
