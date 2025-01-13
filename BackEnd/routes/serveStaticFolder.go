package routes

import (
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/middleware"
)

func ServeStaticFolder() {
	http.Handle("/static/", middleware.ApplyMiddleware(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./FrontEnd/static"))),
		middleware.SetCSPHeaders,
	))
	http.Handle("/uploads/", middleware.ApplyMiddleware(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))),
		middleware.SetCSPHeaders,
	))
}
