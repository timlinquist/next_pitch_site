package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID          int            `json:"id"`
	FirstName   sql.NullString `json:"first_name"`
	LastName    sql.NullString `json:"last_name"`
	Email       string         `json:"email"`
	PhoneNumber sql.NullString `json:"phone_number"`
	IsAdmin     bool           `json:"is_admin"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
