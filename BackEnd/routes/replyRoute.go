package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func ReplyRoute(db *sql.DB) {
	replyController := controllers.NewReplyController(db)

	http.Handle("/reply/", middleware.ApplyMiddleware(
		handlers.ReplyHandler(replyController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
	))
}
