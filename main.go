package main

import (
	"log"
	"net/http"

	"flarecloud/internal/database"
	"flarecloud/internal/handlers"
	"flarecloud/internal/middleware"
)

func main() {
	// env.LoadEnv()
	client := database.ConnectMongoDB()
	defer client.Disconnect(nil)

	http.Handle("/health", middleware.LoggingMiddleware(http.HandlerFunc(handlers.HealthHandler)))
	// http.Handle("/articles", middleware.LoggingMiddleware(http.HandlerFunc(handlers.PostArticleHandler)))
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}