package models

import (
	"time"
)

type CampAgeGroup struct {
	ID          int       `json:"id"`
	CampID      int       `json:"camp_id"`
	MinAge      int       `json:"min_age" binding:"required"`
	MaxAge      int       `json:"max_age" binding:"required"`
	MaxCapacity int       `json:"max_capacity" binding:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
