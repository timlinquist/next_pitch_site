package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
)

func TestGetUserByEmail(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	db := &testDB{db: mockDB}
	service := NewUserService(db)

	tests := []struct {
		name          string
		email         string
		setupMock     func()
		expectedUser  *models.User
		expectedError string
	}{
		{
			name:  "successful user retrieval",
			email: "test@example.com",
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT id, email, first_name, last_name, phone_number, admin, created_at, updated_at
					FROM users
					WHERE email = \$1
				`).WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "phone_number", "admin", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", "Test", "User", sql.NullString{}, false, time.Now(), time.Now()))
			},
			expectedUser: &models.User{
				ID:        1,
				Email:     "test@example.com",
				FirstName: sql.NullString{String: "Test", Valid: true},
				LastName:  sql.NullString{String: "User", Valid: true},
			},
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT id, email, first_name, last_name, phone_number, admin, created_at, updated_at
					FROM users
					WHERE email = \$1
				`).WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			user, err := service.GetUserByEmail(tt.email)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedUser.ID, user.ID)
			assert.Equal(t, tt.expectedUser.Email, user.Email)
			assert.Equal(t, tt.expectedUser.FirstName, user.FirstName)
			assert.Equal(t, tt.expectedUser.LastName, user.LastName)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	db := &testDB{db: mockDB}
	service := NewUserService(db)

	tests := []struct {
		name          string
		user          *models.User
		setupMock     func()
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
			setupMock: func() {
				mock.ExpectQuery(`
					INSERT INTO users \(email, first_name, last_name, phone_number, admin, created_at, updated_at\)
					VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\)
					RETURNING id
				`).WithArgs("test@example.com", "Test", "User", sql.NullString{}, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name: "duplicate email",
			user: &models.User{
				Email:     "existing@example.com",
				FirstName: sql.NullString{String: "Test", Valid: true},
				LastName:  sql.NullString{String: "User", Valid: true},
				IsAdmin:   false,
			},
			setupMock: func() {
				mock.ExpectQuery(`
					INSERT INTO users \(email, first_name, last_name, phone_number, admin, created_at, updated_at\)
					VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\)
					RETURNING id
				`).WithArgs("existing@example.com", "Test", "User", sql.NullString{}, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "sql: no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			err := service.CreateUser(tt.user)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.user.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestIsAdmin(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	db := &testDB{db: mockDB}
	service := NewUserService(db)

	tests := []struct {
		name          string
		email         string
		setupMock     func()
		expectedAdmin bool
		expectedError string
	}{
		{
			name:  "admin user",
			email: "admin@example.com",
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT admin
					FROM users
					WHERE email = \$1
				`).WithArgs("admin@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"admin"}).AddRow(true))
			},
			expectedAdmin: true,
		},
		{
			name:  "non-admin user",
			email: "user@example.com",
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT admin
					FROM users
					WHERE email = \$1
				`).WithArgs("user@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"admin"}).AddRow(false))
			},
			expectedAdmin: false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT admin
					FROM users
					WHERE email = \$1
				`).WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			isAdmin, err := service.IsAdmin(tt.email)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAdmin, isAdmin)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
