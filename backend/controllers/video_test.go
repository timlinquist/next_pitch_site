package controllers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
	"nextpitch.com/backend/test/helpers"
)

const testUserEmail = "test@example.com"

// mockS3Client implements the models.S3Client interface
type mockS3Client struct {
	putObjectFunc    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	deleteObjectFunc func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(ctx, params, optFns...)
	}
	return &s3.DeleteObjectOutput{}, nil
}

// mockEmailService mocks the email service
type mockEmailService struct {
	sendVideoUploadNotificationFunc func(user *models.User, fileName string) error
}

func (m *mockEmailService) SendVideoUploadNotification(user *models.User, fileName string) error {
	if m.sendVideoUploadNotificationFunc != nil {
		return m.sendVideoUploadNotificationFunc(user, fileName)
	}
	return nil
}

func (m *mockEmailService) SendContactEmail(form models.ContactForm) error {
	return nil
}

func (m *mockEmailService) SendAppointmentCancellationEmail(entry *models.ScheduleEntry) error {
	return nil
}

func (m *mockEmailService) SendAppointmentConfirmationEmail(entry *models.ScheduleEntry) error {
	return nil
}

func (m *mockEmailService) SendCampRegistrationConfirmation(reg *models.CampRegistration, athlete *models.Athlete, camp *models.Camp) {
}

func (m *mockEmailService) SendAdminCampRegistrationNotification(reg *models.CampRegistration, athlete *models.Athlete, camp *models.Camp) {
}

func (m *mockEmailService) QueueEmail(data services.EmailData) {
	// Do nothing
}

func (m *mockEmailService) processEmails() {
	// Do nothing
}

// Test helper functions
func setupTestController(t *testing.T) (*VideoController, *services.UserService, *mockEmailService) {
	// Setup test database
	testDB := helpers.SetupTestDB(t)

	// Create video_uploads table if it doesn't exist
	_, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS video_uploads (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			s3_url TEXT NOT NULL,
			file_name TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	assert.NoError(t, err)

	// Set AWS environment variables for testing
	os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	os.Setenv("AWS_S3_BUCKET", "test-bucket")

	// Setup mock S3 client
	mockS3 := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}
	originalFactory := models.CurrentS3ClientFactory
	models.CurrentS3ClientFactory = func() (models.S3Client, error) {
		return mockS3, nil
	}
	t.Cleanup(func() {
		models.CurrentS3ClientFactory = originalFactory
	})

	// Create user service with test DB connection
	userService := services.NewUserService(testDB)

	// Create mock email service
	mockEmail := &mockEmailService{}

	controller, err := NewVideoController(testDB, userService, mockEmail)
	assert.NoError(t, err)
	assert.NotNil(t, controller)

	return controller, userService, mockEmail
}

func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	return ctx, w
}

func setupTestUser(t *testing.T, userService *services.UserService) *models.User {
	user := &models.User{
		Email: testUserEmail,
		Name:  "Test User",
	}
	err := userService.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func TestUploadVideo(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful upload",
			setupContext: func(ctx *gin.Context) {
				// Simulate what the auth middleware would do
				ctx.Set("user_email", testUserEmail)

				// Create a test file
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("video", "test.mp4")
				part.Write([]byte("test video content"))
				writer.Close()

				ctx.Request = httptest.NewRequest("POST", "/upload", body)
				ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthenticated user",
			setupContext: func(ctx *gin.Context) {
				// Don't set user_email in context
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, userService, mockEmail := setupTestController(t)

			// Setup test user in database
			setupTestUser(t, userService)

			// Setup mock email service
			mockEmail.sendVideoUploadNotificationFunc = func(user *models.User, fileName string) error {
				return nil
			}

			ctx, w := createTestContext()
			tt.setupContext(ctx)

			controller.UploadVideo(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestGetVideos(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful retrieval",
			setupContext: func(ctx *gin.Context) {
				ctx.Set("user_email", testUserEmail)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthenticated user",
			setupContext: func(ctx *gin.Context) {
				// Don't set user_email in context
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, userService, _ := setupTestController(t)

			// Setup test user in database
			setupTestUser(t, userService)

			ctx, w := createTestContext()
			tt.setupContext(ctx)

			controller.GetVideos(ctx)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}
