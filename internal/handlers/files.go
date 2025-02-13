package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"flarecloud/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uploadPath = "uploads/"

func init() {
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		log.Fatalf("Erreur création du dossier de stockage : %v", err)
	}
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB max 

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Erreur récupération du fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := filepath.Join(uploadPath, handler.Filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Erreur de stockage du fichier", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Erreur lors de la copie", http.StatusInternalServerError)
		return
	}

	collection := database.Client.Database("cdn").Collection("files")
	_, err = collection.InsertOne(context.Background(), bson.M{
		"name":      handler.Filename,
		"path":      filePath,
		"size":      handler.Size,
		"uploaded":  time.Now(),
	})
	if err != nil {
		http.Error(w, "Erreur sauvegarde metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Fichier %s uploadé avec succès\n", handler.Filename)
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "Nom du fichier manquant", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadPath, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Erreur ouverture fichier", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "Nom du fichier manquant", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadPath, fileName)

	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Erreur suppression fichier", http.StatusInternalServerError)
		return
	}

	collection := database.Client.Database("cdn").Collection("files")
	_, err := collection.DeleteOne(context.Background(), bson.M{"name": fileName})
	if err != nil {
		http.Error(w, "Erreur suppression metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Fichier %s supprimé avec succès\n", fileName)
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	collection := database.Client.Database("cdn").Collection("files")

	cursor, err := collection.Find(context.Background(), bson.M{}, options.Find())
	if err != nil {
		http.Error(w, "Erreur récupération fichiers", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var files []bson.M
	if err = cursor.All(context.Background(), &files); err != nil {
		http.Error(w, "Erreur traitement fichiers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%v", files)
}
