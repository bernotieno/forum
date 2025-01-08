package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	if db == nil {
		fmt.Println("An error occured while initializing Database")
		os.Exit(1)
	}
	logger.Info("Database initialized successfully")

	routes.HomeRoute()
	routes.ServeStaticFolder()
	routes.UserRegAndLogin(db)
	routes.PostRoutes(db)
	routes.CommentRoute(db)
	routes.ReplyRoute(db)

	fmt.Println("Server running at http://localhost:8080")
	logger.Info("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Server failed to start: %v", err)
		log.Fatal(err)
	}
}
