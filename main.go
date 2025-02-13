package main

import (
	"context"
	"log"
	"net/http"

	"flarecloud/internal/database"
	"flarecloud/internal/env"
	"flarecloud/internal/handlers"
	"flarecloud/internal/middleware"
)

func main() {
	env.LoadEnv()
	client := database.ConnectMongoDB()
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	handlers.InitMinio()

	http.Handle("/health", middleware.LoggingMiddleware(http.HandlerFunc(handlers.HealthHandler)))

	http.HandleFunc("/upload", handlers.UploadFileHandler)
	http.HandleFunc("/download", handlers.DownloadFileHandler)
	http.HandleFunc("/delete", handlers.DeleteFileHandler)
	http.HandleFunc("/list", handlers.ListFilesHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
