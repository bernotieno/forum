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
	// Create a LikesController instance
	LikesController := controllers.NewLikesController(db)

	// Rate limiter for likes
	likesLimiter := middleware.NewRateLimiter(30, time.Minute) // 30 likes per minute

	http.Handle("/likePost", middleware.ApplyMiddleware(
		handlers.CreateLikeHandler(LikesController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		likesLimiter.RateLimit,
		middleware.VerifyCSRFMiddleware(db),
	))
}
