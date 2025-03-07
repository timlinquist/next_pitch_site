package controllers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"nextpitch.com/backend/models"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB in bytes

// VideoUploader defines the interface for video upload operations
type VideoUploader interface {
	Upload(file io.Reader, filename string) (string, string, error)
}

// VideoController handles video-related HTTP requests
type VideoController struct {
	video VideoUploader
}

func NewVideoController() (*VideoController, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Initialize video model
	video, err := models.NewVideo()
	if err != nil {
		return nil, fmt.Errorf("error initializing video model: %v", err)
	}

	return &VideoController{
		video: video,
	}, nil
}

func (c *VideoController) UploadVideo(ctx *gin.Context) {
	// Get the file from the request
	file, err := ctx.FormFile("video")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}

	// Validate file size
	if file.Size > maxFileSize {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File too large. Maximum size is %dMB", maxFileSize/1024/1024)})
		return
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	if ext != ".mp4" && ext != ".mov" && ext != ".avi" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only MP4, MOV, and AVI files are allowed"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Upload the file
	path, link, err := c.video.Upload(src, file.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Video uploaded successfully",
		"path":    path,
		"link":    link,
	})
}
