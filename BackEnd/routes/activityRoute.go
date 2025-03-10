package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func ActivityRoutes(db *sql.DB) {
	http.Handle("/activity", middleware.ApplyMiddleware(
		handlers.NewActivityPageHandler(db),
		middleware.SetCSPHeaders,
		middleware.AuthMiddleware,
		middleware.CORSMiddleware,
		middleware.ErrorHandler(handlers.ServeErrorPage),
		middleware.ValidatePathAndMethod("/activity", http.MethodGet),
	))
} 