package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func HomeRoute(db *sql.DB) {
	// Less strict rate limit for home page views
	homePageLimiter := middleware.NewRateLimiter(60, time.Minute) // 60 requests per minute

	http.Handle("/", middleware.ApplyMiddleware(
		handlers.NewHomePageHandler(db),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		homePageLimiter.RateLimit,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/", http.MethodGet),
	))

	// http.HandleFunc("/",handlers.NewHomePageHandler(db))
}
