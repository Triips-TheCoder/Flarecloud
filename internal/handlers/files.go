package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"flarecloud/internal/database"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	// "go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	endpoint        = "localhost:9000"
	accessKeyID     = "minioadmin"
	secretAccessKey = "minioadmin"
	bucketName      = "cdn-bucket"
	useSSL          = false
)

// create Folder type for mongodb
type Folder struct {
	ID  primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"name"`
	Path string        `bson:"path"`
	CreatedAt time.Time `bson:"created_at"`
	ParentID  primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
}

// create File type for mongodb
type File struct {
	ID  primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"name"`
	Size int64 `bson:"size"`
	Type string `bson:"type"`
	Path string `bson:"path"`
	CreatedAt time.Time `bson:"created_at"`
	ParentID primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
}

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

// func findOrCreateFolder(collection *mongo.Collection, folderPath string) (primitive.ObjectID, error) {
// 	segments := strings.Split(folderPath, "/")
// 	var parentID primitive.ObjectID

// 	for _, segment := range segments {
// 		if segment == "" {
// 			continue
// 		}

// 		var folder Folder
// 		filter := bson.M{"name": segment, "parent_id": parentID}
// 		err := collection.FindOne(context.Background(), filter).Decode(&folder)

// 		if err == mongo.ErrNoDocuments {
// 			newFolder := Folder{
// 				ID:        primitive.NewObjectID(),
// 				Name:      segment,
// 				CreatedAt: time.Now(),
// 				ParentID:  parentID,
// 			}
// 			_, err := collection.InsertOne(context.Background(), newFolder)
// 			if err != nil {
// 				return primitive.NilObjectID, err
// 			}
// 			parentID = newFolder.ID
// 		} else if err != nil {
// 			return primitive.NilObjectID, err
// 		} else {
// 			parentID = folder.ID
// 		}
// 	}

// 	return parentID, nil
// }


func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
    collection := database.Client.Database("flarecloud").Collection("files")

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

    parentID := r.FormValue("parentId")

    var parentObjectID primitive.ObjectID
    if parentID != "" {
        var err error
        parentObjectID, err = primitive.ObjectIDFromHex(parentID)
        if err != nil {
            http.Error(w, "Invalid parent ID", http.StatusBadRequest)
            return
        }
    }

    contentType := handler.Header.Get("Content-Type")

    // Generate a unique path using UUID
    uniquePath := uuid.New().String() + "/" + handler.Filename

    _, err = minioClient.PutObject(
        context.Background(),
        bucketName,
        uniquePath,
        file,
        handler.Size,
        minio.PutObjectOptions{ContentType: contentType},
    )
    if err != nil {
        http.Error(w, fmt.Sprintf("Erreur upload vers MinIO : %v", err), http.StatusInternalServerError)
        return
    }

    fileDoc := File{
        Name:      handler.Filename,
        Size:      handler.Size,
        Type:      contentType,
        Path:      uniquePath,
        CreatedAt: time.Now(),
        ParentID:  parentObjectID,
    }

    _, err = collection.InsertOne(context.Background(), fileDoc)
    if err != nil {
        http.Error(w, "Erreur lors de l'enregistrement du fichier dans la base de données", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json")
    // message created
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Fichier créé avec succès",
    }); err != nil {
        http.Error(w, "Erreur lors de l'encodage de la réponse JSON", http.StatusInternalServerError)
        return
    }
}

func CreateFolderHandler(w http.ResponseWriter, r *http.Request) {
	collection := database.Client.Database("flarecloud").Collection("folders")

	folder := r.URL.Query().Get("folder")
	if folder == "" {
		http.Error(w, "Nom du dossier manquant", http.StatusBadRequest)
		return
	}

	parentID := r.FormValue("parentId")

	var parentObjectID primitive.ObjectID
	if parentID != "" {
		var err error
		parentObjectID, err = primitive.ObjectIDFromHex(parentID)
		if err != nil {
			http.Error(w, "Invalid parent ID", http.StatusBadRequest)
			return
		}
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

	folderDoc := Folder{
		Name:      folder,
		CreatedAt: time.Now(),
		ParentID:  parentObjectID,
	}
	_, err = collection.InsertOne(context.Background(), folderDoc)
	if err != nil {
		http.Error(w, "Erreur lors de l'enregistrement du dossier dans la base de données", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json")
    // message created
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Dossier créé avec succès",
    }); err != nil {
        http.Error(w, "Erreur lors de l'encodage de la réponse JSON", http.StatusInternalServerError)
        return
    }
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

    // Delete from Minio
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

    // Delete from Database
    filesCollection := database.Client.Database("flarecloud").Collection("files")
    _, err := filesCollection.DeleteOne(ctx, bson.M{"path": path})
    if err != nil {
        http.Error(w, fmt.Sprintf("Erreur suppression de la base de données : %v", err), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    // message created
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Fichier supprimé avec succès",
    }); err != nil {
        http.Error(w, "Erreur lors de l'encodage de la réponse JSON", http.StatusInternalServerError)
        return
    }
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
    parentID := r.FormValue("parentId")
    var filter bson.M

    if parentID != "" {
        parentObjectID, err := primitive.ObjectIDFromHex(parentID)
        if err != nil {
            http.Error(w, "Invalid parent ID", http.StatusBadRequest)
            return
        }
        filter = bson.M{"parent_id": parentObjectID}
    } else {
        filter = bson.M{"parent_id": nil}
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    filesCollection := database.Client.Database("flarecloud").Collection("files")
    foldersCollection := database.Client.Database("flarecloud").Collection("folders")

    var files []File
    cursor, err := filesCollection.Find(ctx, filter)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error retrieving files: %v", err), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var file File
        if err := cursor.Decode(&file); err != nil {
            http.Error(w, fmt.Sprintf("Error decoding file record: %v", err), http.StatusInternalServerError)
            return
        }
        files = append(files, file)
    }

    var folders []Folder
    cursor, err = foldersCollection.Find(ctx, filter)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error retrieving folders: %v", err), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var folder Folder
        if err := cursor.Decode(&folder); err != nil {
            http.Error(w, fmt.Sprintf("Error decoding folder record: %v", err), http.StatusInternalServerError)
            return
        }
        folders = append(folders, folder)
    }

    var history []map[string]interface{}
    if parentID != "" {
        currentID, _ := primitive.ObjectIDFromHex(parentID)
        for {
            var folder Folder
            err := foldersCollection.FindOne(ctx, bson.M{"_id": currentID}).Decode(&folder)
            if err != nil {
                break
            }
            history = append([]map[string]interface{}{{"name": folder.Name, "parentID": folder.ParentID}}, history...)
            if folder.ParentID == primitive.NilObjectID {
                break
            }
            currentID = folder.ParentID
        }
    }

    for i := range folders {
        if folders[i].ParentID == primitive.NilObjectID {
            folders[i].ParentID = primitive.ObjectID{} // or any other default value
        }
    }

    response := map[string]interface{}{
        "files":   files,
        "folders": folders,
    }

    if parentID != "" {
        response["history"] = history
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Erreur lors de l'encodage de la réponse JSON", http.StatusInternalServerError)
        return
    }
}