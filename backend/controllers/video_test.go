package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
	"nextpitch.com/backend/test/helpers"
)

func TestNewVideoController(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Create services
	userService := services.NewUserService(testDB)
	emailService := services.NewEmailService()

	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful initialization",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
				"AWS_S3_BUCKET":         "test-bucket",
			},
			wantErr: false,
		},
		{
			name:        "missing AWS credentials",
			envVars:     map[string]string{},
			wantErr:     true,
			errContains: "error loading .env file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create a mock environment loader
			mockEnvLoader := func(filenames ...string) error {
				if len(tt.envVars) == 0 {
					return fmt.Errorf("error loading .env file")
				}
				return nil
			}

			// Test initialization
			got, err := NewVideoController(testDB, userService, emailService, mockEnvLoader)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, testDB, got.db)
			assert.Equal(t, userService, got.userService)
			assert.Equal(t, emailService, got.emailService)
		})
	}
}

func TestVideoController_UploadVideo(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Set up test environment variables
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	os.Setenv("AWS_S3_BUCKET", "test-bucket")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_S3_BUCKET")
	}()

	// Create services
	userService := services.NewUserService(testDB)
	emailService := services.NewEmailService()

	// Create test user
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		IsAdmin: false,
	}
	err := userService.CreateUser(user)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		file           []byte
		filename       string
		userEmail      string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "successful upload with db verification",
			file:           []byte("test video content"),
			filename:       "test.mp4",
			userEmail:      user.Email,
			expectedStatus: http.StatusOK,
			expectedFields: []string{"link", "message", "path", "upload_id"},
		},
		{
			name:           "invalid file",
			file:           []byte{},
			filename:       "test.mp4",
			userEmail:      user.Email,
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "unauthorized",
			file:           []byte("test video content"),
			filename:       "test.mp4",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
			expectedFields: []string{"error"},
		},
		{
			name:           "invalid file type",
			file:           []byte("test video content"),
			filename:       "test.txt",
			userEmail:      user.Email,
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:           "file too large",
			file:           make([]byte, maxFileSize+1),
			filename:       "test.mp4",
			userEmail:      user.Email,
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create controller
			vc, err := NewVideoController(testDB, userService, emailService, func(filenames ...string) error { return nil })
			assert.NoError(t, err)

			// Setup Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user email in context
			if tt.userEmail != "" {
				c.Set("user_email", tt.userEmail)
			}

			// Create multipart form
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			if len(tt.file) > 0 {
				part, err := writer.CreateFormFile("video", tt.filename)
				assert.NoError(t, err)
				_, err = part.Write(tt.file)
				assert.NoError(t, err)
			}
			writer.Close()

			// Create request
			req := httptest.NewRequest("POST", "/api/video/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			c.Request = req

			// Test upload
			vc.UploadVideo(c)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			// For successful uploads, verify database record
			if tt.expectedStatus == http.StatusOK {
				var count int
				err = testDB.QueryRow("SELECT COUNT(*) FROM video_uploads WHERE user_id = $1", user.ID).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestVideoController_GetVideos(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Create services
	userService := services.NewUserService(testDB)
	emailService := services.NewEmailService()

	// Create test user
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		IsAdmin: false,
	}
	err := userService.CreateUser(user)
	assert.NoError(t, err)

	// Insert test video uploads
	_, err = testDB.Exec(`
		INSERT INTO video_uploads (user_id, s3_url, file_name, status)
		VALUES ($1, $2, $3, $4)
	`, user.ID, "test/path1.mp4", "test1.mp4", models.VideoUploadStatusUploaded)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userEmail      string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "successful retrieval",
			userEmail:      user.Email,
			expectedStatus: http.StatusOK,
			expectedFields: []string{"videos"},
		},
		{
			name:           "unauthorized",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
			expectedFields: []string{"error"},
		},
		{
			name:           "user not found",
			userEmail:      "nonexistent@example.com",
			expectedStatus: http.StatusUnauthorized,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create controller
			vc, err := NewVideoController(testDB, userService, emailService, func(filenames ...string) error { return nil })
			assert.NoError(t, err)

			// Setup Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user email in context
			if tt.userEmail != "" {
				c.Set("user_email", tt.userEmail)
			}

			// Test get videos
			vc.GetVideos(c)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			// For successful retrieval, verify videos array
			if tt.expectedStatus == http.StatusOK {
				videos, ok := response["videos"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, 1, len(videos))

				video := videos[0].(map[string]interface{})
				assert.Equal(t, "test1.mp4", video["file_name"])
				assert.Equal(t, "test/path1.mp4", video["s3_url"])
				assert.Equal(t, string(models.VideoUploadStatusUploaded), video["status"])
			}
		})
	}
}
