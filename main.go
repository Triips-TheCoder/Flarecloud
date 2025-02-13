package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"flarecloud/internal/database"
	"flarecloud/internal/env"
	"flarecloud/internal/handlers"
	"flarecloud/internal/middleware"
	service "flarecloud/internal/services"
)

func main() {
	env.LoadEnv()
	client := database.ConnectMongoDB()
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	userCollection := client.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")
	authService := service.NewAuthService(userCollection)
	authHandler := handlers.NewAuthHandler(authService)

	http.Handle("/health", middleware.LoggingMiddleware(http.HandlerFunc(handlers.HealthHandler)))
	http.Handle("/signup", middleware.LoggingMiddleware(http.HandlerFunc(authHandler.SignUp)))
	http.Handle("/login", middleware.LoggingMiddleware(http.HandlerFunc(authHandler.Login)))

	http.HandleFunc("/upload", handlers.UploadFileHandler)
	http.HandleFunc("/download", handlers.DownloadFileHandler)
	http.HandleFunc("/delete", handlers.DeleteFileHandler)
	http.HandleFunc("/list", handlers.ListFilesHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
