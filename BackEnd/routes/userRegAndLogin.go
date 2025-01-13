package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func UserRegAndLogin(db *sql.DB) {
	AuthController := controllers.NewAuthController(db)

	http.Handle("/check_login", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.CheckLoginHandler),
		middleware.SetCSPHeaders,
	))

	http.Handle("/register", middleware.ApplyMiddleware(
		handlers.RegisterHandler(AuthController),
		middleware.SetCSPHeaders,
	))

	http.Handle("/login", middleware.ApplyMiddleware(
		handlers.LoginHandler(AuthController),
		middleware.SetCSPHeaders,
	))

	http.Handle("/logout", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.LogoutHandler),
		middleware.SetCSPHeaders,
	))

	http.Handle("/login_Page", middleware.ApplyMiddleware(
		http.HandlerFunc(handlers.LoginPageHandler),
		middleware.SetCSPHeaders,
	))
}
