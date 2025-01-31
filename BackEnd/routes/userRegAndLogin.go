package routes

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func UserRegAndLogin(db *sql.DB) {
	AuthController := controllers.NewAuthController(db)

	// Strict rate limit for authentication attempts
	authLimiter := middleware.NewRateLimiter(5, time.Minute) // 5 attempts per minute

	// Less strict rate limit for page views
	pageLimiter := middleware.NewRateLimiter(30, time.Minute) // 30 requests per minute

	http.Handle("/login", middleware.ApplyMiddleware(
		handlers.LoginHandler(AuthController),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		authLimiter.RateLimit,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/login", http.MethodPost),
	))

	http.Handle("/register", middleware.ApplyMiddleware(
		handlers.RegisterHandler(AuthController),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		authLimiter.RateLimit,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/register", http.MethodPost),
	))

	http.Handle("/login_Page", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.LoginPageHandler),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		pageLimiter.RateLimit,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/login_Page", http.MethodGet),
	))

	http.Handle("/checkLoginStatus", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.CheckLoginHandler),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		pageLimiter.RateLimit,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/checkLoginStatus", http.MethodGet),
	))

	http.Handle("/logout", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.LogoutHandler),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
		pageLimiter.RateLimit,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.VerifyCSRFMiddleware(db),
		middleware.ValidatePathAndMethod("/logout", http.MethodPost),
	))

	// http.Handle("/googleLogin", middleware.ApplyMiddleware(
	// 	http.HandlerFunc(handlers.GoogleHandler),
	// 	middleware.SetCSPHeaders,
	// 	middleware.CORSMiddleware,
	// 	// authLimiter.RateLimit,
	// 	middleware.ErrorHandler(handlers.ServeErrorPage),
	// ))

	http.HandleFunc("/googleLogin", handlers.GoogleHandler)
	http.HandleFunc("/auth/google/callback", handlers.CallbackHandler)
	http.HandleFunc("/githubLogin", handlers.GitHubHandler)
	http.HandleFunc("/auth/github/callback", handlers.GitHubCallbackHandler)

	// http.Handle("/auth/google/callback", middleware.ApplyMiddleware(
	// 	http.HandlerFunc(handlers.CallbackHandler),
	// 	middleware.SetCSPHeaders,
	// 	middleware.CORSMiddleware,
	// 	middleware.ErrorHandler(handlers.ServeErrorPage),
	// ))
}
