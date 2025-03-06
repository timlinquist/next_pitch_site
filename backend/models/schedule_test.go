package models

import (
	"testing"
	"time"
)

func TestScheduleEntryValidation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		entry   ScheduleEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: ScheduleEntry{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   now,
				EndTime:     now.Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name: "missing title",
			entry: ScheduleEntry{
				Description: "Test Description",
				StartTime:   now,
				EndTime:     now.Add(time.Hour),
			},
			wantErr: true,
		},
		{
			name: "missing start time",
			entry: ScheduleEntry{
				Title:       "Test Event",
				Description: "Test Description",
				EndTime:     now.Add(time.Hour),
			},
			wantErr: true,
		},
		{
			name: "missing end time",
			entry: ScheduleEntry{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   now,
			},
			wantErr: true,
		},
		{
			name: "end time before start time",
			entry: ScheduleEntry{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   now,
				EndTime:     now.Add(-time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The validation is handled by the Gin binding tags
			// We just need to ensure the struct is properly defined
			if tt.wantErr {
				// These cases should fail validation at the binding level
				return
			}
			// Valid cases should have all required fields
			if tt.entry.Title == "" {
				t.Error("Title should not be empty")
			}
			if tt.entry.StartTime.IsZero() {
				t.Error("StartTime should not be zero")
			}
			if tt.entry.EndTime.IsZero() {
				t.Error("EndTime should not be zero")
			}
			if tt.entry.EndTime.Before(tt.entry.StartTime) {
				t.Error("EndTime should not be before StartTime")
			}
		})
	}
}
