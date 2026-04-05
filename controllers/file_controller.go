package controllers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dansonserge/DaFileService/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileController struct {
	minioService  *services.MinioService
	defaultBucket string
}

func NewFileController(minioService *services.MinioService, defaultBucket string) *FileController {
	return &FileController{
		minioService:  minioService,
		defaultBucket: defaultBucket,
	}
}

// Upload handles multipart file streams
func (ctrl *FileController) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	bucket := c.DefaultPostForm("bucket", ctrl.defaultBucket)

	// Create unique object name - allow override from form
	objectName := c.PostForm("object_name")
	if objectName == "" {
		objectName = uuid.New().String() + "-" + file.Filename
	}

	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	ctx := context.Background()
	err = ctrl.minioService.UploadFile(ctx, bucket, objectName, fileContent, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to storage: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Upload successful",
		"bucket":      bucket,
		"object_name": objectName,
		"filename":    file.Filename,
	})
}

// Download handles streaming retrieval
func (ctrl *FileController) Download(c *gin.Context) {
	bucket := c.Param("bucket")
	// For keys with slashes
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Object key is required"})
		return
	}

	ctx := context.Background()
	reader, info, err := ctrl.minioService.DownloadFile(ctx, bucket, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer reader.Close()

	contentType := info.ContentType
	if strings.HasSuffix(strings.ToLower(key), ".pdf") {
		contentType = "application/pdf"
	}

	if contentType == "application/pdf" {
		c.Header("Content-Disposition", "inline")
	} else {
		c.Header("Content-Disposition", "inline; filename=\""+key+"\"")
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(info.Size, 10))

	http.ServeContent(c.Writer, c.Request, key, info.LastModified, reader.(io.ReadSeeker))
}

// GetPreview generates a temporal presigned URL for secure visualization
func (ctrl *FileController) GetPreview(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	ctx := context.Background()
	// Default 30 minutes
	url, err := ctrl.minioService.GetPresignedURL(ctx, bucket, key, 30*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate preview link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url.String()})
}

// Delete removes an asset from storage
func (ctrl *FileController) Delete(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	ctx := context.Background()
	err := ctrl.minioService.DeleteFile(ctx, bucket, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete asset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset deleted successfully"})
}

// ListBuckets list available storage containers
func (ctrl *FileController) ListBuckets(c *gin.Context) {
	ctx := context.Background()
	buckets, err := ctrl.minioService.ListBuckets(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve buckets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": buckets})
}
