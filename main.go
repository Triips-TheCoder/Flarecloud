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

func applyMiddlewares(handler http.Handler) http.Handler {
	return middleware.EnableCORS(middleware.LoggingMiddleware(middleware.LimitMiddleware(handler)))
}

func main() {
	env.LoadEnv()
	client := database.ConnectMongoDB()
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	handlers.InitMinio()

	

	http.Handle("/health", applyMiddlewares(http.HandlerFunc(handlers.HealthHandler)))
	http.Handle("/captcha", applyMiddlewares(http.HandlerFunc(handlers.CaptchaHandler)))
	http.Handle("/upload", applyMiddlewares(http.HandlerFunc(handlers.UploadFileHandler)))
	http.Handle("/create-folder", applyMiddlewares(http.HandlerFunc(handlers.CreateFolderHandler)))
	http.Handle("/download", applyMiddlewares(http.HandlerFunc(handlers.DownloadFileHandler)))
	http.Handle("/delete", applyMiddlewares(http.HandlerFunc(handlers.DeleteFileHandler)))
	http.Handle("/list", applyMiddlewares(http.HandlerFunc(handlers.ListFilesHandler)))
	

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
