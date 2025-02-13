package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectMongoDB() *mongo.Client {
	mongoURI := os.Getenv("MONGO_URI") // Read from env variable
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // Default to local MongoDB
	}

	log.Printf("Connecting to MongoDB at %s", mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	cli, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")
	Client = cli
	return cli
}