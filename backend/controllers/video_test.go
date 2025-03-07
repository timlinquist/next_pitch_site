package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockVideo implements a mock Video model for testing
type MockVideo struct {
	uploadFunc func(file io.Reader, filename string) (string, string, error)
}

func (m *MockVideo) Upload(file io.Reader, filename string) (string, string, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(file, filename)
	}
	return "", "", nil
}

func TestNewVideoController(t *testing.T) {
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
			errContains: "DROPBOX_ACCESS_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("DROPBOX_ACCESS_TOKEN", tt.envToken)
			defer os.Unsetenv("DROPBOX_ACCESS_TOKEN")

			// Test initialization
			got, err := NewVideoController()
			if tt.wantErr {
				if err == nil {
					t.Error("NewVideoController() error = nil, wantErr true")
				}
				if tt.errContains != "" && !errors.Is(err, errors.New(tt.errContains)) {
					t.Errorf("NewVideoController() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("NewVideoController() error = %v, wantErr false", err)
				return
			}
			if got == nil {
				t.Error("NewVideoController() returned nil without error")
			}
		})
	}
}

func TestVideoController_UploadVideo(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		filename       string
		fileSize       int64
		uploadFunc     func(file io.Reader, filename string) (string, string, error)
		wantStatus     int
		wantResponse   map[string]interface{}
		wantErrMessage string
	}{
		{
			name:        "successful upload",
			fileContent: "test video content",
			filename:    "test.mp4",
			fileSize:    1024, // 1KB
			uploadFunc: func(file io.Reader, filename string) (string, string, error) {
				return "/videos/test.mp4", "https://test.com/video.mp4", nil
			},
			wantStatus: http.StatusOK,
			wantResponse: map[string]interface{}{
				"message": "Video uploaded successfully",
				"path":    "/videos/test.mp4",
				"link":    "https://test.com/video.mp4",
			},
		},
		{
			name:           "missing file",
			fileContent:    "",
			filename:       "",
			fileSize:       0,
			uploadFunc:     nil,
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "No video file provided",
		},
		{
			name:           "file too large",
			fileContent:    "test video content",
			filename:       "test.mp4",
			fileSize:       maxFileSize + 1,
			uploadFunc:     nil,
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "File too large",
		},
		{
			name:           "invalid file type",
			fileContent:    "test video content",
			filename:       "test.txt",
			fileSize:       1024,
			uploadFunc:     nil,
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Invalid file type",
		},
		{
			name:        "upload failure",
			fileContent: "test video content",
			filename:    "test.mp4",
			fileSize:    1024,
			uploadFunc: func(file io.Reader, filename string) (string, string, error) {
				return "", "", errors.New("upload failed")
			},
			wantStatus:     http.StatusInternalServerError,
			wantErrMessage: "upload failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up Gin router
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			if tt.fileContent != "" {
				part, err := writer.CreateFormFile("video", tt.filename)
				if err != nil {
					t.Fatalf("Failed to create form file: %v", err)
				}
				part.Write([]byte(tt.fileContent))
			}
			writer.Close()

			// Set up request
			c.Request, _ = http.NewRequest("POST", "/api/video/upload", body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())

			// Create mock video model
			mockVideo := &MockVideo{
				uploadFunc: tt.uploadFunc,
			}

			// Create controller with mock
			controller := &VideoController{
				video: mockVideo,
			}

			// Test upload
			controller.UploadVideo(c)

			// Check response
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantResponse != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				assert.Equal(t, tt.wantResponse, response)
			}

			if tt.wantErrMessage != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				assert.Contains(t, response["error"], tt.wantErrMessage)
			}
		})
	}
}
