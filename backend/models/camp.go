package models

import (
	"time"
)

type Camp struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	PriceCents  *int      `json:"price_cents"`
	MaxCapacity *int      `json:"max_capacity"`
	Slug        *string   `json:"slug"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
