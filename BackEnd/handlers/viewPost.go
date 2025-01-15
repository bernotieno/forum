package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type ViewPostHandler struct {
	db *sql.DB
}

func NewViewPostHandler(db *sql.DB) http.HandlerFunc {
	return (&ViewPostHandler{db: db}).ServeHTTP
}

func (h *ViewPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Extract post ID from URL
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Check if user is logged in
	loggedIn, userID := isLoggedIn(h.db, r) // Assume `isLoggedIn` returns user details if authenticated.

	// Generate CSRF token if user is logged in
	var csrfToken string
	if loggedIn {
		sessionToken, err := controllers.GetSessionToken(r)
		if err != nil {
			logger.Error("Error getting session token: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		csrfToken, _ = controllers.GenerateCSRFToken(h.db, sessionToken)
	}

	// Create a PostController instance
	postController := controllers.NewPostController(h.db)

	// Fetch the post from the database
	post, err := postController.GetPostByID(postID)
	if err != nil {
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	// Determine if the logged-in user is the post author
	isAuthor := loggedIn && userID == post.UserID

	// Create CommentController and fetch comments
	commentController := controllers.NewCommentController(h.db)
	comments, err := commentController.GetCommentsByPostID(postID)
	if err != nil {
		logger.Error("Failed to fetch comments: %v", err)
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	// Create template function map
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 02, 2006 at 15:04")
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	// Create template with function map
	tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(
		"./FrontEnd/templates/layout.html",
		"./FrontEnd/templates/viewPost.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		IsAuthenticated bool
		IsAuthor        bool
		CSRFToken       string
		Post            models.Post
		Comments        []models.Comment
	}{
		IsAuthenticated: loggedIn,
		IsAuthor:        isAuthor,
		CSRFToken:       csrfToken,
		Post:            post,
		Comments:        comments,
	}

	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
