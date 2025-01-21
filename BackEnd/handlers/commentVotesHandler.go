package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
)

func CreateCommentVoteHandler(cc *controllers.CommentVotesController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedIn, userID := isLoggedIn(cc.DB, r)
		if !loggedIn {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err := r.ParseForm()
		if err != nil {
			logger.Error("Error while Parsing Form %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		commentID, err := strconv.Atoi(r.FormValue("comment_id"))
		if err != nil {
			logger.Error("Invalid Comment Id %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		voteType := r.FormValue("vote")
		if voteType != "like" && voteType != "dislike" {
			logger.Warning("Invalid Vote Type request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = cc.HandleCommentVote(commentID, userID, voteType)
		if err != nil {
			logger.Error("Failed to handle comment vote: %v", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		likes, dislikes, err := cc.GetCommentVotes(commentID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{
			"likes":    likes,
			"dislikes": dislikes,
		})
	}
}

func GetUserCommentVotesHandler(cc *controllers.CommentVotesController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is logged in
		loggedIn, userID := isLoggedIn(cc.DB, r)
		if !loggedIn {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Query to get all user's comment votes
		query := `
			SELECT comment_id, vote_type 
			FROM comment_votes 
			WHERE user_id = ?
		`
		rows, err := cc.DB.Query(query, userID)
		if err != nil {
			logger.Error("Failed to fetch user comment votes: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Map to store comment votes
		userVotes := make(map[string]string)

		// Iterate through results
		for rows.Next() {
			var commentID int
			var voteType string
			if err := rows.Scan(&commentID, &voteType); err != nil {
				logger.Error("Error scanning vote row: %v", err)
				continue
			}
			userVotes[strconv.Itoa(commentID)] = voteType
		}

		// Return the votes as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userVotes)
	}
}
