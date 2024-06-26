package v1

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"tarkib.uz/pkg/logger"
)

type fileRoutes struct {
	l logger.Interface
}

func newFileRoutes(handler *gin.RouterGroup, l logger.Interface) {
	r := &fileRoutes{l}

	h := handler.Group("/file")
	{
		h.POST("/upload", r.upload)
	}
}

type File struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// @Summary 		Image upload
// @Description 	Api for image upload
// @Tags 			file-upload
// @Accept 			json
// @Produce 		json
// @Param 			file formData file true "Image"
// @Success 		200 {object} string
// @Failure 		400 {object} string
// @Failure 		500 {object} string
// @Router 			/file/upload [post]
func (f *fileRoutes) upload(c *gin.Context) {

	endpoint := "localhost:9000"
	accessKeyID := "nodirbek"
	secretAccessKey := "nodirbek"
	bucketName := "avatars"
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error connecting to MiniIO server")
		log.Println("Error connecting to MiniIO server:", err)
		return
	}

	var file File
	if err := c.ShouldBind(&file); err != nil {
		c.JSON(http.StatusBadRequest, "Error uploading file")
		log.Println("Error uploading file:", err)
		return
	}

	ext := filepath.Ext(file.File.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".svg" && ext != ".jpeg" {
		c.JSON(http.StatusBadRequest, "Bad request of file image")
		log.Println("Bad request of file image")
		return
	}

	id := uuid.New().String()
	objectName := id + ext
	contentType := "image/jpeg"

	fileReader, err := file.File.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error opening file")
		log.Println("Error opening file:", err)
		return
	}
	defer fileReader.Close()

	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, fileReader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error uploading file to MinIO server")
		log.Println("Error uploading file to MinIO server:", err)
		return
	}

	minioURL := fmt.Sprintf("http://%s/%s/%s", "localhost:9000", bucketName, objectName)

	c.JSON(http.StatusOK, gin.H{
		"url": minioURL,
	})
}
