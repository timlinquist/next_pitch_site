package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/test/helpers"
)

func TestGetUserByEmail(t *testing.T) {
	// Setup test database and fixtures
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewUserService(testDB.DB)

	tests := []struct {
		name          string
		email         string
		expectedUser  *models.User
		expectedError string
	}{
		{
			name:  "successful user retrieval",
			email: "test@example.com",
			expectedUser: &models.User{
				Email:     "test@example.com",
				FirstName: sql.NullString{String: "Test", Valid: true},
				LastName:  sql.NullString{String: "User", Valid: true},
			},
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
			testDB.CleanupTestDB(t)

			// Insert test user if needed
			if tt.expectedUser != nil {
				_, err := testDB.DB.Exec(`
					INSERT INTO users (email, first_name, last_name, admin, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6)
				`, tt.expectedUser.Email, tt.expectedUser.FirstName.String, tt.expectedUser.LastName.String,
					false, time.Now(), time.Now())
				assert.NoError(t, err)
			}

			user, err := service.GetUserByEmail(tt.email)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, user.ID)
			assert.Equal(t, tt.expectedUser.Email, user.Email)
			assert.Equal(t, tt.expectedUser.FirstName, user.FirstName)
			assert.Equal(t, tt.expectedUser.LastName, user.LastName)
		})
	}
}

func TestCreateUser(t *testing.T) {
	// Setup test database and fixtures
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewUserService(testDB.DB)

	tests := []struct {
		name          string
		user          *models.User
		expectedError string
	}{
		{
			name: "successful user creation",
			user: &models.User{
				Email:     "test@example.com",
				FirstName: sql.NullString{String: "Test", Valid: true},
				LastName:  sql.NullString{String: "User", Valid: true},
				IsAdmin:   false,
			},
		},
		{
			name: "duplicate email",
			user: &models.User{
				Email:     "test@example.com",
				FirstName: sql.NullString{String: "Test", Valid: true},
				LastName:  sql.NullString{String: "User", Valid: true},
				IsAdmin:   false,
			},
			expectedError: "duplicate key value violates unique constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			testDB.CleanupTestDB(t)

			// Insert existing user for duplicate email test
			if tt.expectedError != "" {
				_, err := testDB.DB.Exec(`
					INSERT INTO users (email, first_name, last_name, admin, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6)
				`, tt.user.Email, tt.user.FirstName.String, tt.user.LastName.String,
					tt.user.IsAdmin, time.Now(), time.Now())
				assert.NoError(t, err)
			}

			err := service.CreateUser(tt.user)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.user.ID)

			// Verify the user was created
			var count int
			err = testDB.DB.QueryRow(`
				SELECT COUNT(*) FROM users WHERE email = $1
			`, tt.user.Email).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

func TestIsAdmin(t *testing.T) {
	// Setup test database and fixtures
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewUserService(testDB.DB)

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
			testDB.CleanupTestDB(t)

			// Insert test user if needed
			if tt.expectedError == "" {
				_, err := testDB.DB.Exec(`
					INSERT INTO users (email, first_name, last_name, admin, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6)
				`, tt.email, "Test", "User", tt.isAdmin, time.Now(), time.Now())
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
