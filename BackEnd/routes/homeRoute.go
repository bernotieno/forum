package routes

import (
	"database/sql"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

func HomeRoute(db *sql.DB) {
	http.HandleFunc("/", handlers.NewHomePageHandler(db))
}
