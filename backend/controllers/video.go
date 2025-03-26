package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
)

const maxFileSize = 250 * 1024 * 1024 // 250MB in bytes

// VideoController handles video-related HTTP requests
type VideoController struct {
	db           *sql.DB
	userService  *services.UserService
	emailService services.EmailServiceInterface
}

// EnvLoader is a function type for loading environment variables
type EnvLoader func(filenames ...string) error

// NewVideoController creates a new video controller
func NewVideoController(db *sql.DB, userService *services.UserService, emailService services.EmailServiceInterface, envLoader ...EnvLoader) (*VideoController, error) {
	// Load environment variables
	loader := godotenv.Load
	if len(envLoader) > 0 {
		loader = envLoader[0]
	}
	if err := loader(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	return &VideoController{
		db:           db,
		userService:  userService,
		emailService: emailService,
	}, nil
}

func (c *VideoController) UploadVideo(ctx *gin.Context) {
	// Get the user email from the context (set by auth middleware)
	userEmail, exists := ctx.Get("user_email")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user by email
	user, err := c.userService.GetUserByEmail(userEmail.(string))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

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

	// Upload to S3
	key, link, err := models.UploadVideo(src, file.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload video: %v", err)})
		return
	}

	// Create video upload record
	upload := &models.VideoUpload{
		UserID:   user.ID,
		S3URL:    key,
		FileName: file.Filename,
		Status:   models.VideoUploadStatusUploaded,
	}

	if err := models.CreateVideoUpload(c.db, upload); err != nil {
		// If we fail to create the record, try to delete the uploaded file
		if delErr := models.DeleteVideo(key); delErr != nil {
			log.Printf("[Video] Failed to delete S3 file after record creation failure: %v", delErr)
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload record"})
		return
	}

	// Send email notification
	if err := c.emailService.SendVideoUploadNotification(user, file.Filename); err != nil {
		log.Printf("[Video] Failed to send upload notification: %v", err)
		// Don't fail the request if email fails
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Video uploaded successfully",
		"path":      key,
		"link":      link,
		"upload_id": upload.ID,
	})
}

func (c *VideoController) GetVideos(ctx *gin.Context) {
	// Get the user email from the context (set by auth middleware)
	userEmail, exists := ctx.Get("user_email")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user by email
	user, err := c.userService.GetUserByEmail(userEmail.(string))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get user's video uploads
	rows, err := c.db.Query(`
		SELECT id, s3_url, file_name, status, created_at
		FROM video_uploads
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video uploads"})
		return
	}
	defer rows.Close()

	var videos []map[string]interface{}
	for rows.Next() {
		var upload models.VideoUpload
		err := rows.Scan(&upload.ID, &upload.S3URL, &upload.FileName, &upload.Status, &upload.CreatedAt)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan video upload"})
			return
		}

		videos = append(videos, map[string]interface{}{
			"id":         upload.ID,
			"s3_url":     upload.S3URL,
			"file_name":  upload.FileName,
			"status":     upload.Status,
			"created_at": upload.CreatedAt,
		})
	}

	if err = rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating video uploads"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"videos": videos,
	})
}
