package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	folder := r.FormValue("folder")
	if folder != "" && !strings.HasSuffix(folder, "/") {
		folder += "/"
	}

	objectName := folder + filepath.Base(handler.Filename)
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
	fmt.Fprintf(w, "✅ Fichier %s uploadé dans %s\n", handler.Filename, folder)
}

func CreateFolderHandler(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	if folder == "" {
		http.Error(w, "Nom du dossier manquant", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(folder, "/") {
		folder += "/"
	}

	_, err := minioClient.PutObject(
		context.Background(),
		bucketName,
		folder,
		strings.NewReader(""),
		0,
		minio.PutObjectOptions{ContentType: "application/x-directory"},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur création dossier : %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "📂 Dossier '%s' créé avec succès\n", folder)
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

	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, object); err != nil {
		http.Error(w, "Erreur de téléchargement", http.StatusInternalServerError)
	}
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Chemin manquant", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	opts := minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}

	objectsCh := minioClient.ListObjects(ctx, bucketName, opts)
	for obj := range objectsCh {
		if obj.Err != nil {
			http.Error(w, fmt.Sprintf("Erreur suppression : %v", obj.Err), http.StatusInternalServerError)
			return
		}
		err := minioClient.RemoveObject(ctx, bucketName, obj.Key, minio.RemoveObjectOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("Impossible de supprimer %s : %v", obj.Key, err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "🗑️ Fichier/dossier '%s' supprimé avec succès\n", path)
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	prefix := r.URL.Query().Get("prefix")
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	}

	files := []string{}
	objectsCh := minioClient.ListObjects(ctx, bucketName, opts)
	for obj := range objectsCh {
		if obj.Err != nil {
			http.Error(w, fmt.Sprintf("Erreur listing : %v", obj.Err), http.StatusInternalServerError)
			return
		}
		files = append(files, obj.Key)
	}

	w.WriteHeader(http.StatusOK)
	for _, file := range files {
		if strings.HasSuffix(file, "/") {
			fmt.Fprintf(w, "📂 %s\n", file)
		} else {
			fmt.Fprintf(w, "📄 %s\n", file)
		}
	}
}
