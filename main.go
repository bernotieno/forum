package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Raymond9734/forum.git/BackEnd/controllers"
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
	log.Println("Database initialized successfully")

	// Create a context that cancels on interrupt signals (e.g., Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a WaitGroup to wait for cleanup goroutines to finish
	var wg sync.WaitGroup

	// Start cleanup tasks in separate goroutines
	wg.Add(1)
	go func() {
		defer wg.Done()
		controllers.CleanupExpiredCSRFTokens(ctx, db)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		controllers.CleanupExpiredSessions(ctx, db)
	}()

	routes.HomeRoute(db)
	routes.ServeStaticFolder()
	routes.UserRegAndLogin(db)
	routes.PostRoutes(db)
	routes.CommentRoute(db)
	routes.ReplyRoute(db)

	// Start the HTTP server
	server := &http.Server{
		Addr: ":8080",
	}

	// Run the server in a goroutine
	go func() {
		log.Println("Server running at http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed to start: %v\n", err)
		}
	}()

	// Wait for interrupt signal (e.g., Ctrl+C) to gracefully shut down the server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Shutdown the server gracefully
	log.Println("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}

	// Cancel the context to signal cleanup tasks to stop
	cancel()

	// Wait for cleanup tasks to finish
	wg.Wait()

	log.Println("Application stopped gracefully")
}
