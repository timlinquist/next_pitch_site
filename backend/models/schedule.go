package models

import "time"

type RecurrenceType string

const (
	RecurrenceNone     RecurrenceType = "none"
	RecurrenceWeekly   RecurrenceType = "weekly"
	RecurrenceBiweekly RecurrenceType = "biweekly"
	RecurrenceMonthly  RecurrenceType = "monthly"
)

type ScheduleEntry struct {
	ID                int            `json:"id"`
	Title             string         `json:"title" binding:"required"`
	Description       string         `json:"description"`
	StartTime         time.Time      `json:"start_time" binding:"required"`
	EndTime           time.Time      `json:"end_time" binding:"required"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	UserEmail         string         `json:"user_email" binding:"required"`
	Recurrence        RecurrenceType `json:"recurrence" binding:"required,oneof=none weekly biweekly monthly"`
	RecurrenceEndDate *time.Time     `json:"recurrence_end_date"`
	ParentEventID     *int           `json:"parent_event_id"`
}
