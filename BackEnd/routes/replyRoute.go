package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func ReplyRoute(db *sql.DB) {
	replyController := controllers.NewReplyController(db)

	replyLimiter := middleware.NewRateLimiter(10, time.Minute) // 10 replies per minute

	http.Handle("/reply/", middleware.ApplyMiddleware(
		handlers.ReplyHandler(replyController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		replyLimiter.RateLimit,
		middleware.VerifyCSRFMiddleware(db),
	))
}
