package services

import (
	"database/sql"
	"errors"
	"time"

	"nextpitch.com/backend/models"
)

type UserService struct {
	db DB
}

func NewUserService(db DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, email, first_name, last_name, phone_number, admin, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) CreateUser(user *models.User) error {
	now := time.Now()
	err := s.db.QueryRow(`
		INSERT INTO users (email, first_name, last_name, phone_number, admin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		user.Email,
		user.FirstName.String,
		user.LastName.String,
		user.PhoneNumber.String,
		user.IsAdmin,
		now,
		now,
	).Scan(&user.ID)

	if err != nil {
		return err
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (s *UserService) IsAdmin(email string) (bool, error) {
	var isAdmin bool
	err := s.db.QueryRow(`
		SELECT is_admin
		FROM users
		WHERE email = $1
	`, email).Scan(&isAdmin)

	if err == sql.ErrNoRows {
		return false, errors.New("user not found")
	}
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}
