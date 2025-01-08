package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

func CommentRoute(db *sql.DB) {
	commentController := controllers.NewCommentController(db)

	http.HandleFunc("/comment/", handlers.CommentHandler(commentController))
}
