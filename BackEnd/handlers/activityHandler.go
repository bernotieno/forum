package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type ActivityPageHandler struct {
	db *sql.DB
}

func NewActivityPageHandler(db *sql.DB) http.HandlerFunc {
	return (&ActivityPageHandler{db: db}).ServeHTTP
}

func (h *ActivityPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	// Check if user is logged in
	loggedIn, userID := isLoggedIn(h.db, r)
	if !loggedIn {
		// Redirect to login page if not logged in
		http.Redirect(w, r, "/login_Page", http.StatusSeeOther)
		return
	}
	
	// Generate CSRF token
	var csrfToken string
	sessionToken, err := controllers.GetSessionToken(r)
	if err != nil {
		logger.Error("Error getting session token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	csrfToken, err = controllers.GenerateCSRFToken(h.db, sessionToken)
	if err != nil {
		logger.Error("Error generating CSRF token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	// Create activity controller
	activityController := controllers.NewActivityController(h.db)
	
	// Get user's posts
	userPosts, err := activityController.GetUserPosts(userID)
	if err != nil {
		logger.Error("Failed to fetch user posts: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	// Get posts the user has voted on
	votedPosts, err := activityController.GetUserVotedPosts(userID)
	if err != nil {
		logger.Error("Failed to fetch user voted posts: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	// Get user's comments
	userComments, err := activityController.GetUserComments(userID)
	if err != nil {
		logger.Error("Failed to fetch user comments: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	// Add comment counts to posts
	commentController := controllers.NewCommentController(h.db)
	
	// Process user posts
	for i := range userPosts {
		userPosts[i].IsAuthor = true // User is always the author of their own posts
		
		// Fetch total comment count
		commentCount, err := commentController.GetCommentCountByPostID(userPosts[i].ID)
		if err != nil {
			logger.Error("Failed to fetch comment count for post %d: %v", userPosts[i].ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userPosts[i].Comments = make([]models.Comment, 0)
		userPosts[i].CommentCount = commentCount
	}
	
	// Process voted posts
	for i := range votedPosts {
		votedPosts[i].IsAuthor = votedPosts[i].UserID == userID
		
		// Fetch total comment count
		commentCount, err := commentController.GetCommentCountByPostID(votedPosts[i].ID)
		if err != nil {
			logger.Error("Failed to fetch comment count for post %d: %v", votedPosts[i].ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		votedPosts[i].Comments = make([]models.Comment, 0)
		votedPosts[i].CommentCount = commentCount
	}
	
	// Create template function map
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 02, 2006 at 15:04")
		},
		"split": strings.Split,
		"trim":  strings.TrimSpace,
		"truncate": func(s string, maxLen int) string {
			if len(s) <= maxLen {
				return s
			}
			return s[:maxLen] + "..."
		},
	}
	
	// Create template with function map
	tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(
		"./FrontEnd/templates/layout.html",
		"./FrontEnd/templates/activity.html",
	)
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	// Prepare data for the template
	data := struct {
		IsAuthenticated bool
		CSRFToken       string
		UserID          int
		UserPosts       []models.Post
		VotedPosts      []models.Post
		UserComments    []models.UserCommentActivity
	}{
		IsAuthenticated: loggedIn,
		CSRFToken:       csrfToken,
		UserID:          userID,
		UserPosts:       userPosts,
		VotedPosts:      votedPosts,
		UserComments:    userComments,
	}
	
	// Execute template with data
	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		logger.Error("Failed to execute template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
} 