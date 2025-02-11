package main

import (
	"context"
	"log"
	"net/http"

	"flarecloud/internal/database"
	"flarecloud/internal/handlers"
	"flarecloud/internal/middleware"
)

func main() {
	client := database.ConnectMongoDB()
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()	

	http.Handle("/health", middleware.LoggingMiddleware(http.HandlerFunc(handlers.HealthHandler)))
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
