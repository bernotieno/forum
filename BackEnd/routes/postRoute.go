package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func PostRoutes(db *sql.DB) {
	PostController := controllers.NewPostController(db)

	// Rate limit for post creation
	postLimiter := middleware.NewRateLimiter(10, time.Minute)  // 10 posts per minute

	// Less strict limit for viewing
	viewLimiter := middleware.NewRateLimiter(60, time.Minute)  // 60 views per minute

	http.Handle("/viewPost", middleware.ApplyMiddleware(
		handlers.NewViewPostHandler(db),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		viewLimiter.RateLimit,  // Less strict for viewing
	))

	http.Handle("/create-post", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.CreatePostPageHandler),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		postLimiter.RateLimit,
	))

	http.Handle("/createPost", middleware.ApplyMiddleware(
		handlers.PostHandler(PostController),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		postLimiter.RateLimit,  // Strict limit for post creation
	))
}
