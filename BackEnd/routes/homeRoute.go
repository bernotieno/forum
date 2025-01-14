package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func HomeRoute(db *sql.DB) {
	http.Handle("/", middleware.ApplyMiddleware(
		handlers.NewHomePageHandler(db),
		middleware.SetCSPHeaders,
		middleware.CORSMiddleware,
	))

	// http.HandleFunc("/",handlers.NewHomePageHandler(db))
}
