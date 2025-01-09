package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

func PostRoutes(db *sql.DB) {
	PostController := controllers.NewPostController(db)

	http.HandleFunc("/create-post", handlers.CreatePostPageHandler)
	http.HandleFunc("/createPost", handlers.PostHandler(PostController))
}
