package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func CommentRoute(db *sql.DB) {
	commentController := controllers.NewCommentController(db)

	// Rate limit for comments
	commentLimiter := middleware.NewRateLimiter(10, time.Minute) // 10 comments per minute

	http.Handle("/comment/", middleware.ApplyMiddleware(
		handlers.CommentHandler(commentController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		commentLimiter.RateLimit,
		middleware.VerifyCSRFMiddleware(db),
	))

	http.Handle("/deleteComment", middleware.ApplyMiddleware(
		handlers.DeleteCommentHandler(commentController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		commentLimiter.RateLimit,
		middleware.VerifyCSRFMiddleware(db),
	))

	http.Handle("/updateComment", middleware.ApplyMiddleware(
		handlers.UpdateCommentHandler(commentController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		commentLimiter.RateLimit,
		middleware.VerifyCSRFMiddleware(db),
	))
}
