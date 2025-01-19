package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func LikesRoutes(db *sql.DB) {
	// Create controllers
	LikesController := controllers.NewLikesController(db)
	CommentVotesController := controllers.NewCommentVotesController(db)

	// Rate limiter for likes
	likesLimiter := middleware.NewRateLimiter(30, time.Minute) // 30 likes per minute

	// Post vote routes
	http.Handle("/likePost", middleware.ApplyMiddleware(
		handlers.CreateUserVoteHandler(LikesController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		likesLimiter.RateLimit,
		middleware.VerifyCSRFMiddleware(db),
	))

	http.Handle("/getUserVotes", middleware.ApplyMiddleware(
		handlers.GetUserVotesHandler(LikesController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		middleware.VerifyCSRFMiddleware(db),
	))

	// Comment vote routes
	http.Handle("/commentVote", middleware.ApplyMiddleware(
		handlers.CreateCommentVoteHandler(CommentVotesController),
		middleware.AuthMiddleware,
		middleware.VerifyCSRFMiddleware(db),
	))

	http.Handle("/getUserCommentVotes", middleware.ApplyMiddleware(
		handlers.GetUserCommentVotesHandler(CommentVotesController),
		middleware.AuthMiddleware,
	))
}
