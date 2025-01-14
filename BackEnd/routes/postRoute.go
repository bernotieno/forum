package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func PostRoutes(db *sql.DB) {
	PostController := controllers.NewPostController(db)

	http.Handle("/viewPost", middleware.ApplyMiddleware(
		handlers.NewViewPostHandler(db),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
	))

	http.Handle("/create-post", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.CreatePostPageHandler),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
	))

	http.Handle("/createPost", middleware.ApplyMiddleware(
		handlers.PostHandler(PostController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
	))
}
