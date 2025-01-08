package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

func ReplyRoute(db *sql.DB) {
	replyController := controllers.NewReplyController(db)

	http.HandleFunc("/reply/", handlers.ReplyHandler(replyController))
}
