package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"flarecloud/internal/database"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	
	log.Printf("Pinging MongoDB with client: %v", database.Client.Ping(ctx, nil))
	if err := database.Client.Ping(ctx, nil); err != nil {
		http.Error(w, "MongoDB not available", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Service is available")); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
