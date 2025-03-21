package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// MockVideo implements VideoUploader interface for testing
type MockVideo struct {
	uploadFunc func(file io.Reader, filename string) (string, string, error)
}

func (m *MockVideo) Upload(file io.Reader, filename string) (string, string, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(file, filename)
	}
	return "/test/path", "https://test.com/video.mp4", nil
}

func TestNewVideoController(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Create services
	userService := services.NewUserService(testDB)
	emailService := services.NewEmailService()

	tests := []struct {
		name        string
		envToken    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful initialization",
			envToken: "test-token",
			wantErr:  false,
		},
		{
			name:        "missing token",
			envToken:    "",
			wantErr:     true,
			errContains: "error loading .env file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("DROPBOX_ACCESS_TOKEN", tt.envToken)
			defer os.Unsetenv("DROPBOX_ACCESS_TOKEN")

			// Create a mock environment loader
			mockEnvLoader := func(filenames ...string) error {
				if tt.envToken == "" {
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
			assert.NotNil(t, got.video)
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
	os.Setenv("DROPBOX_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("DROPBOX_ACCESS_TOKEN")

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
		mockUpload     func(file io.Reader, filename string) (string, string, error)
		expectedStatus int
		expectedFields []string
	}{
		{
			name:      "successful upload with db verification",
			file:      []byte("test video content"),
			filename:  "test.mp4",
			userEmail: user.Email,
			mockUpload: func(file io.Reader, filename string) (string, string, error) {
				return "/test/path", "https://test.com/video.mp4", nil
			},
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
			name:      "invalid file type",
			file:      []byte("test video content"),
			filename:  "test.txt",
			userEmail: user.Email,
			mockUpload: func(file io.Reader, filename string) (string, string, error) {
				return "", "", fmt.Errorf("Invalid file type")
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name:      "upload failure",
			file:      []byte("test video content"),
			filename:  "test.mp4",
			userEmail: user.Email,
			mockUpload: func(file io.Reader, filename string) (string, string, error) {
				return "", "", fmt.Errorf("upload failed")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock video uploader
			mockVideo := &MockVideo{
				uploadFunc: tt.mockUpload,
			}

			// Create controller with mock
			vc := &VideoController{
				video:        mockVideo,
				db:           testDB,
				userService:  userService,
				emailService: emailService,
			}

			// Setup Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user email in context
			if tt.userEmail != "" {
				c.Set("user_email", tt.userEmail)
			}

			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if len(tt.file) > 0 {
				part, err := writer.CreateFormFile("video", tt.filename)
				assert.NoError(t, err)
				_, err = part.Write(tt.file)
				assert.NoError(t, err)
			}
			err = writer.Close()
			assert.NoError(t, err)

			// Set request
			c.Request = httptest.NewRequest("POST", "/upload", body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())

			// Test
			vc.UploadVideo(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// If successful upload, verify database
			if tt.expectedStatus == http.StatusOK {
				var uploadID int
				err := testDB.QueryRow("SELECT id FROM video_uploads WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1", user.ID).Scan(&uploadID)
				assert.NoError(t, err)
				assert.Equal(t, float64(uploadID), response["upload_id"])
			}
		})
	}
}

func TestVideoController_GetVideos(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Set up test environment variables
	os.Setenv("DROPBOX_ACCESS_TOKEN", "test_token")
	defer os.Unsetenv("DROPBOX_ACCESS_TOKEN")

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

	// Create test video uploads
	_, err = testDB.Exec(`
		INSERT INTO video_uploads (user_id, dropbox_url, file_name, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, user.ID, "https://test.com/video1.mp4", "test1.mp4", models.VideoUploadStatusUploaded)
	assert.NoError(t, err)

	_, err = testDB.Exec(`
		INSERT INTO video_uploads (user_id, dropbox_url, file_name, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, user.ID, "https://test.com/video2.mp4", "test2.mp4", models.VideoUploadStatusUploaded)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userEmail      string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "successful retrieval",
			userEmail:      user.Email,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "unauthorized",
			userEmail:      "",
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
		},
		{
			name:           "user not found",
			userEmail:      "nonexistent@example.com",
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create controller
			vc := &VideoController{
				video:        &MockVideo{},
				db:           testDB,
				userService:  userService,
				emailService: emailService,
			}

			// Setup Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set user email in context
			if tt.userEmail != "" {
				c.Set("user_email", tt.userEmail)
			}

			// Test
			vc.GetVideos(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				videos, ok := response["videos"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCount, len(videos))
			}
		})
	}
}
