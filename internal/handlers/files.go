package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	endpoint        = "localhost:9000"
	accessKeyID     = os.Getenv("MINIO_ROOT_USER")
	secretAccessKey = os.Getenv("MINIO_ROOT_PASSWORD")
	bucketName      = "cdn-bucket"
	useSSL          = false
)

var minioClient *minio.Client

func InitMinio() {
	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Impossible d'initialiser MinIO : %v", err)
	}

	log.Println("✅ MinIO connecté avec succès")
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Erreur lors de l'analyse du formulaire multipart : "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Erreur récupération du fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()

	objectName := filepath.Base(handler.Filename)
	contentType := handler.Header.Get("Content-Type")

	_, err = minioClient.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		handler.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur upload vers MinIO : %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "✅ Fichier %s uploadé avec succès dans MinIO\n", objectName)
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "Nom du fichier manquant", http.StatusBadRequest)
		return
	}

	object, err := minioClient.GetObject(context.Background(), bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}
	defer object.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, object); err != nil {
		http.Error(w, "Erreur de téléchargement", http.StatusInternalServerError)
	}
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "Nom du fichier manquant", http.StatusBadRequest)
		return
	}

	err := minioClient.RemoveObject(context.Background(), bucketName, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		http.Error(w, "Erreur suppression fichier", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "🗑️ Fichier %s supprimé avec succès\n", fileName)
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Recursive: true})

	files := []string{}
	for object := range objectCh {
		if object.Err != nil {
			http.Error(w, fmt.Sprintf("Erreur listing : %v", object.Err), http.StatusInternalServerError)
			return
		}
		files = append(files, object.Key)
	}

	w.WriteHeader(http.StatusOK)
	for _, file := range files {
		fmt.Fprintf(w, "📂 %s\n", file)
	}
}
