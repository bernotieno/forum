package main

import (
	"log"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/routes"
)

func main() {
	db := database.Init()

	routes.ServeStaticFolder()
	routes.UserRegAndLogin(db)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
