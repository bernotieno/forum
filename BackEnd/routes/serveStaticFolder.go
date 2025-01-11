package routes

import (
	"net/http"
)

func ServeStaticFolder() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./FrontEnd/static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
}
