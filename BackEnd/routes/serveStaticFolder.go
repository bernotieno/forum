package routes

import (
	"net/http"
)

func ServeStaticFolder() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./FrontEnd/static"))))
}
