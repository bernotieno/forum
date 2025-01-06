package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

func UserRegAndLogin(db *sql.DB) {
	AuthController := controllers.NewAuthController(db)

	http.HandleFunc("/check_login", handlers.CheckLoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler(AuthController))
	http.HandleFunc("/login", handlers.LoginHandler(AuthController))
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/login_Page", handlers.LoginPageHandler)
}
