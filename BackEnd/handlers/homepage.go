package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type HomePageHandler struct {
	db *sql.DB
}

func NewHomePageHandler(db *sql.DB) http.HandlerFunc {
	return (&HomePageHandler{db: db}).ServeHTTP
}

func (h *HomePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Check if user is logged in
	loggedIn, userID := isLoggedIn(h.db, r)

	var csrfToken string

	if loggedIn {
		sessionToken, err := controllers.GetSessionToken(r)
		if err != nil {
			logger.Error("Error getting session token: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Generate CSRF token for the session
		csrfToken, err = controllers.GenerateCSRFToken(h.db, sessionToken)
		if err != nil {
			logger.Error("Error generating CSRF token: %V", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Create a PostController instance using the handler's db
	postController := controllers.NewPostController(h.db)

	// Fetch posts from the database using the controller
	posts, err := postController.GetAllPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Add IsAuthor field to each post
	for i := range posts {
		posts[i].IsAuthor = loggedIn && posts[i].UserID == userID
	}

    // Fetch comments for each post
    commentController := controllers.NewCommentController(h.db)
    for i := range posts {
        comments, err := commentController.GetCommentsByPostID(strconv.Itoa(posts[i].ID))
        if err != nil {
            logger.Error("Failed to fetch comments for post %d: %v", posts[i].ID, err)
            http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
            return
        }
        posts[i].Comments = comments
    }

	// Create template function map
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 02, 2006 at 15:04")
		},
	}

	// Create template with function map
	tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(
		"./FrontEnd/templates/layout.html",
		"./FrontEnd/templates/homepage.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		IsAuthenticated bool
		CSRFToken       string
		Posts           []models.Post
	}{
		IsAuthenticated: loggedIn,
		CSRFToken:       csrfToken,
		Posts:           posts,
	}

	// Execute template with data
	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
