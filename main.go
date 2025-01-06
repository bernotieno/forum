package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Raymond9734/forum.git/BackEnd/database"
	"github.com/Raymond9734/forum.git/BackEnd/logger"
	"github.com/Raymond9734/forum.git/BackEnd/routes"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatal(err)
	}

	logger.Info("Starting application...")

	db := database.Init()
	logger.Info("Database initialized successfully")

	routes.ServeStaticFolder()
	routes.UserRegAndLogin(db)

	fmt.Println("Server running at http://localhost:8080")
	logger.Info("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Server failed to start: %v", err)
		log.Fatal(err)
	}
}
