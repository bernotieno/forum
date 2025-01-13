package handlers

import (
	"database/sql"
	"net/http"
	"text/template"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/models"
)

type ViewPostHandler struct {
    db *sql.DB
}

func NewViewPostHandler(db *sql.DB) http.HandlerFunc {
    return (&ViewPostHandler{db: db}).ServeHTTP
}

func (h *ViewPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")

    // Extract post ID from URL
    postID := r.URL.Query().Get("id")
    if postID == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }

    // Check if user is logged in
    loggedIn, _ := isLoggedIn(r)

    // Create a PostController instance using the handler's db
    postController := controllers.NewPostController(h.db)

    // Fetch the post from the database using the controller
    post, err := postController.GetPostByID(postID)
    if err != nil {
        http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
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
        "./FrontEnd/templates/viewPost.html",
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    data := struct {
        IsAuthenticated bool
        Post            models.Post
    }{
        IsAuthenticated: loggedIn,
        Post:            post,
    }

    // Execute template with data
    err = tmpl.ExecuteTemplate(w, "layout.html", data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}