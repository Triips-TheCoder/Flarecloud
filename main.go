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

	userCollection := client.Database(os.Getenv("MONGODB_DATABASE")).Collection("users")
	authService := service.NewAuthService(userCollection)
	authHandler := handlers.NewAuthHandler(authService)


	http.Handle("/health", applyMiddlewares(http.HandlerFunc(handlers.HealthHandler)))
	http.Handle("/captcha", applyMiddlewares(http.HandlerFunc(handlers.CaptchaHandler)))
	http.Handle("/upload", applyMiddlewares(http.HandlerFunc(handlers.UploadFileHandler)))
	http.Handle("/create-folder", applyMiddlewares(http.HandlerFunc(handlers.CreateFolderHandler)))
	http.Handle("/download", applyMiddlewares(http.HandlerFunc(handlers.DownloadFileHandler)))
	http.Handle("/delete", applyMiddlewares(http.HandlerFunc(handlers.DeleteFileHandler)))
	http.Handle("/delete-folder", applyMiddlewares(http.HandlerFunc(handlers.DeleteFolderHandler)))
	http.Handle("/list", applyMiddlewares(http.HandlerFunc(handlers.ListFilesHandler)))
	http.Handle("/update-file", applyMiddlewares(http.HandlerFunc(handlers.UpdateFileHandler)))
	http.Handle("/signup", applyMiddlewares(http.HandlerFunc(authHandler.SignUp)))
	http.Handle("/login", applyMiddlewares(http.HandlerFunc(authHandler.Login)))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
