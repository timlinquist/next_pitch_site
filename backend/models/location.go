package models

import "time"

type Location struct {
	ID        int       `json:"id"`
	VenueName *string   `json:"venue_name"`
	Street    string    `json:"street" binding:"required"`
	City      string    `json:"city" binding:"required"`
	State     string    `json:"state" binding:"required"`
	Zip       string    `json:"zip" binding:"required"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
