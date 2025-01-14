package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func CommentRoute(db *sql.DB) {
	commentController := controllers.NewCommentController(db)

	http.Handle("/comment/", middleware.ApplyMiddleware(
		handlers.CommentHandler(commentController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
	))
}
