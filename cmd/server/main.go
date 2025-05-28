// cmd/server/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reddit-ingestion/internal/app"
	_ "reddit-ingestion/docs"
)

// @title Reddit Ingestion API
// @version 1.0
// @description This API provides endpoints to ingest data from Reddit, including posts, comments, user information, and search functionality.
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.email support@example.com
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host 192.168.10.69:8080
// @BasePath /

func main() {
	application, err := app.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	go func() {
		if err := application.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	
	log.Println("Server started successfully")
	log.Println("Swagger documentation available at http://localhost:8080/swagger/index.html")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Echo.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}