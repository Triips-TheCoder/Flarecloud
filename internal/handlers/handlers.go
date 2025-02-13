package handlers

import (
	"context"
	"net/http"
	"time"

	"flarecloud/internal/database"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := database.Client.Ping(ctx, nil); err != nil {
		http.Error(w, "MongoDB not available", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is available"))
}