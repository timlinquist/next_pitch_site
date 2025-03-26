package models

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/test/testutils"
)

func TestUploadVideo(t *testing.T) {
	// Setup test database
	testDB := testutils.SetupTestDB(t)
	defer testDB.Close()
	defer testutils.CleanupTestDB(t, testDB)

	tests := []struct {
		name           string
		content        string
		filename       string
		envVars        map[string]string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "missing AWS credentials",
			content:        "test video content",
			filename:       "test.mp4",
			envVars:        map[string]string{},
			expectedStatus: 1,
			expectedFields: []string{"AWS credentials or bucket name environment variables are not set"},
		},
		{
			name:     "missing bucket name",
			content:  "test video content",
			filename: "test.mp4",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
			},
			expectedStatus: 1,
			expectedFields: []string{"AWS credentials or bucket name environment variables are not set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			os.Unsetenv("AWS_S3_BUCKET")

			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Create test reader
			reader := strings.NewReader(tt.content)

			// Test upload
			path, url, err := UploadVideo(reader, tt.filename)

			// Check results
			if tt.expectedStatus == 0 {
				assert.NoError(t, err)
				assert.Contains(t, path, tt.expectedFields[0])
				assert.Contains(t, url, tt.expectedFields[1])
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedFields[0])
				assert.Empty(t, path)
				assert.Empty(t, url)
			}
		})
	}
}

func TestDeleteVideo(t *testing.T) {
	// Setup test database
	testDB := testutils.SetupTestDB(t)
	defer testDB.Close()
	defer testutils.CleanupTestDB(t, testDB)

	tests := []struct {
		name           string
		key            string
		envVars        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing AWS credentials",
			key:            "videos/test.mp4",
			envVars:        map[string]string{},
			expectedStatus: 1,
			expectedError:  "AWS credentials or bucket name environment variables are not set",
		},
		{
			name: "missing bucket name",
			key:  "videos/test.mp4",
			envVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "test-key",
				"AWS_SECRET_ACCESS_KEY": "test-secret",
			},
			expectedStatus: 1,
			expectedError:  "AWS credentials or bucket name environment variables are not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			os.Unsetenv("AWS_S3_BUCKET")

			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Test deletion
			err := DeleteVideo(tt.key)

			// Check results
			if tt.expectedStatus == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
