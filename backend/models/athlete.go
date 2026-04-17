package models

import (
	"time"
)

type Athlete struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Age         int       `json:"age" binding:"required"`
	YearsPlayed int       `json:"years_played"`
	Position    string    `json:"position"`
	UserID      *int      `json:"user_id"`
	ParentEmail string    `json:"parent_email"`
	ParentPhone string    `json:"parent_phone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
