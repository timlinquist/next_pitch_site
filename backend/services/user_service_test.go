package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/test/helpers"
)

func TestGetUserByEmail(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewUserService(testDB)

	tests := []struct {
		name          string
		email         string
		expectedUser  *models.User
		expectedError string
	}{
		{
			name:  "existing user",
			email: "test@example.com",
			expectedUser: &models.User{
				Email:   "test@example.com",
				Name:    "Test User",
				IsAdmin: false,
			},
			expectedError: "",
		},
		{
			name:          "non-existent user",
			email:         "nonexistent@example.com",
			expectedUser:  nil,
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedUser != nil {
				// Insert test user
				err := service.CreateUser(tt.expectedUser)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
			}

			user, err := service.GetUserByEmail(tt.email)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %s but got none", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error %s but got %s", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if user.Email != tt.expectedUser.Email {
				t.Errorf("Expected email %s but got %s", tt.expectedUser.Email, user.Email)
			}
			if user.Name != tt.expectedUser.Name {
				t.Errorf("Expected name %s but got %s", tt.expectedUser.Name, user.Name)
			}
			if user.IsAdmin != tt.expectedUser.IsAdmin {
				t.Errorf("Expected isAdmin %v but got %v", tt.expectedUser.IsAdmin, user.IsAdmin)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewUserService(testDB)

	tests := []struct {
		name          string
		user          *models.User
		expectedError string
	}{
		{
			name: "valid user",
			user: &models.User{
				Email:   "test@example.com",
				Name:    "Test User",
				IsAdmin: false,
			},
			expectedError: "",
		},
		{
			name: "duplicate email",
			user: &models.User{
				Email:   "test@example.com",
				Name:    "Another Test User",
				IsAdmin: false,
			},
			expectedError: "pq: duplicate key value violates unique constraint \"users_email_key\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateUser(tt.user)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %s but got none", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error %s but got %s", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify user was created correctly
			user, err := service.GetUserByEmail(tt.user.Email)
			if err != nil {
				t.Fatalf("Failed to get created user: %v", err)
			}

			if user.Email != tt.user.Email {
				t.Errorf("Expected email %s but got %s", tt.user.Email, user.Email)
			}
			if user.Name != tt.user.Name {
				t.Errorf("Expected name %s but got %s", tt.user.Name, user.Name)
			}
			if user.IsAdmin != tt.user.IsAdmin {
				t.Errorf("Expected isAdmin %v but got %v", tt.user.IsAdmin, user.IsAdmin)
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	// Setup test database and fixtures
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewUserService(testDB)

	tests := []struct {
		name          string
		email         string
		isAdmin       bool
		expectedError string
	}{
		{
			name:    "admin user",
			email:   "admin@example.com",
			isAdmin: true,
		},
		{
			name:    "non-admin user",
			email:   "user@example.com",
			isAdmin: false,
		},
		{
			name:          "user not found",
			email:         "nonexistent@example.com",
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			_, err := testDB.Exec(`
				TRUNCATE TABLE users CASCADE
			`)
			assert.NoError(t, err)

			// Insert test user if needed
			if tt.expectedError == "" {
				_, err := testDB.Exec(`
					INSERT INTO users (email, name, is_admin, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5)
				`, tt.email, "Test User", tt.isAdmin, time.Now(), time.Now())
				assert.NoError(t, err)
			}

			isAdmin, err := service.IsAdmin(tt.email)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.isAdmin, isAdmin)
		})
	}
}
