package routes

import (
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/handlers"
)

func HomeRoute() {
	http.HandleFunc("/", handlers.HomePageHandler)
}
