package handlers

import (
	"database/sql"
	"net/http"
	"text/template"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
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
	loggedIn, userID := isLoggedIn(r)

	// Generate CSRF token for the session
	csrfToken := controllers.GenerateCSRFToken(userID)

	// Create a PostController instance using the handler's db
	postController := controllers.NewPostController(h.db)

	// Fetch posts from the database using the controller
	posts, err := postController.GetAllPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
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
