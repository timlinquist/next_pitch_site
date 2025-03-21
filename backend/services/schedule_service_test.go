package services

import (
	"testing"
	"time"

	_ "github.com/lib/pq" // postgres driver
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/test/helpers"
)

func TestCreateScheduleEntry(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewScheduleService(testDB)

	tests := []struct {
		name          string
		entry         *models.ScheduleEntry
		userEmail     string
		isAdmin       bool
		expectedError string
	}{
		{
			name: "successful creation by admin",
			entry: &models.ScheduleEntry{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   time.Now().Add(24 * time.Hour),
				EndTime:     time.Now().Add(25 * time.Hour),
				UserEmail:   "admin@example.com",
				Recurrence:  models.RecurrenceNone,
			},
			userEmail:     "admin@example.com",
			isAdmin:       true,
			expectedError: "",
		},
		{
			name: "non-admin cannot create long event",
			entry: &models.ScheduleEntry{
				Title:       "Long Event",
				Description: "Test Description",
				StartTime:   time.Now().Add(72 * time.Hour),
				EndTime:     time.Now().Add(96 * time.Hour),
				UserEmail:   "user@example.com",
				Recurrence:  models.RecurrenceNone,
			},
			userEmail:     "user@example.com",
			isAdmin:       false,
			expectedError: "event duration exceeds maximum allowed duration for non-admin users",
		},
		{
			name: "non-admin can create short event",
			entry: &models.ScheduleEntry{
				Title:       "Short Event",
				Description: "Test Description",
				StartTime:   time.Now().Add(120 * time.Hour),
				EndTime:     time.Now().Add(121 * time.Hour),
				UserEmail:   "user@example.com",
				Recurrence:  models.RecurrenceWeekly,
			},
			userEmail:     "user@example.com",
			isAdmin:       false,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user
			user := &models.User{
				Email:   tt.userEmail,
				Name:    "Test User",
				IsAdmin: tt.isAdmin,
			}
			_, err := testDB.Exec(`
				INSERT INTO users (email, name, is_admin)
				VALUES ($1, $2, $3)
				ON CONFLICT (email) DO NOTHING
			`, user.Email, user.Name, user.IsAdmin)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			err = service.CreateScheduleEntry(tt.entry, tt.userEmail, user.IsAdmin)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %s but got none", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error %s but got %s", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify entry was created
			var count int
			err = testDB.QueryRow(`
				SELECT COUNT(*) FROM schedule_entries WHERE title = $1
			`, tt.entry.Title).Scan(&count)
			if err != nil {
				t.Fatalf("Failed to verify entry creation: %v", err)
			}
			if count != 1 {
				t.Errorf("Expected 1 entry but found %d", count)
			}
		})
	}
}

func TestUpdateScheduleEntry(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewScheduleService(testDB)

	tests := []struct {
		name          string
		initialEntry  *models.ScheduleEntry
		updatedEntry  *models.ScheduleEntry
		userEmail     string
		isAdmin       bool
		expectedError string
	}{
		{
			name: "successful update by admin",
			initialEntry: &models.ScheduleEntry{
				Title:       "Initial Event",
				Description: "Initial Description",
				StartTime:   time.Now().Add(24 * time.Hour),
				EndTime:     time.Now().Add(25 * time.Hour),
				Recurrence:  models.RecurrenceNone,
			},
			updatedEntry: &models.ScheduleEntry{
				Title:       "Updated Event",
				Description: "Updated Description",
				StartTime:   time.Now().Add(48 * time.Hour),
				EndTime:     time.Now().Add(72 * time.Hour),
				Recurrence:  models.RecurrenceMonthly,
			},
			userEmail:     "admin@example.com",
			isAdmin:       true,
			expectedError: "",
		},
		{
			name: "non-admin cannot update to long event",
			initialEntry: &models.ScheduleEntry{
				Title:       "Initial Event",
				Description: "Initial Description",
				StartTime:   time.Now().Add(96 * time.Hour),
				EndTime:     time.Now().Add(97 * time.Hour),
				Recurrence:  models.RecurrenceNone,
			},
			updatedEntry: &models.ScheduleEntry{
				Title:       "Updated Event",
				Description: "Updated Description",
				StartTime:   time.Now().Add(120 * time.Hour),
				EndTime:     time.Now().Add(144 * time.Hour),
				Recurrence:  models.RecurrenceNone,
			},
			userEmail:     "user@example.com",
			isAdmin:       false,
			expectedError: "event duration exceeds maximum allowed duration for non-admin users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user
			user := &models.User{
				Email:   tt.userEmail,
				Name:    "Test User",
				IsAdmin: tt.isAdmin,
			}
			_, err := testDB.Exec(`
				INSERT INTO users (email, name, is_admin)
				VALUES ($1, $2, $3)
				ON CONFLICT (email) DO NOTHING
			`, user.Email, user.Name, user.IsAdmin)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Create initial entry
			tt.initialEntry.UserEmail = tt.userEmail
			err = service.CreateScheduleEntry(tt.initialEntry, tt.userEmail, user.IsAdmin)
			if err != nil {
				t.Fatalf("Failed to create initial entry: %v", err)
			}

			// Update entry
			tt.updatedEntry.ID = tt.initialEntry.ID
			tt.updatedEntry.UserEmail = tt.userEmail
			err = service.UpdateScheduleEntry(tt.updatedEntry, tt.userEmail, user.IsAdmin)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error %s but got none", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error %s but got %s", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify entry was updated
			var title, description string
			err = testDB.QueryRow(`
				SELECT title, description FROM schedule_entries WHERE id = $1
			`, tt.updatedEntry.ID).Scan(&title, &description)
			if err != nil {
				t.Fatalf("Failed to verify entry update: %v", err)
			}
			if title != tt.updatedEntry.Title {
				t.Errorf("Expected title %s but got %s", tt.updatedEntry.Title, title)
			}
			if description != tt.updatedEntry.Description {
				t.Errorf("Expected description %s but got %s", tt.updatedEntry.Description, description)
			}
		})
	}
}

func TestCheckOverlappingEvents(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewScheduleService(testDB)

	// Create test user
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		IsAdmin: false,
	}
	_, err := testDB.Exec(`
		INSERT INTO users (email, name, is_admin)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
	`, user.Email, user.Name, user.IsAdmin)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test event
	event := &models.ScheduleEntry{
		Title:       "Test Event",
		Description: "Test Description",
		StartTime:   time.Now().Add(24 * time.Hour),
		EndTime:     time.Now().Add(25 * time.Hour),
		UserEmail:   user.Email,
		Recurrence:  models.RecurrenceNone,
	}
	err = service.CreateScheduleEntry(event, user.Email, user.IsAdmin)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	// Test overlapping event
	overlappingEvent := &models.ScheduleEntry{
		Title:       "Overlapping Event",
		Description: "Test Description",
		StartTime:   time.Now().Add(24 * time.Hour).Add(30 * time.Minute),
		EndTime:     time.Now().Add(25 * time.Hour).Add(30 * time.Minute),
		UserEmail:   user.Email,
		Recurrence:  models.RecurrenceNone,
	}
	err = service.CreateScheduleEntry(overlappingEvent, user.Email, user.IsAdmin)
	if err == nil {
		t.Error("Expected error for overlapping event but got none")
	} else if err.Error() != "event overlaps with existing events" {
		t.Errorf("Expected error 'event overlaps with existing events' but got %s", err.Error())
	}

	// Test non-overlapping event
	nonOverlappingEvent := &models.ScheduleEntry{
		Title:       "Non-Overlapping Event",
		Description: "Test Description",
		StartTime:   time.Now().Add(26 * time.Hour),
		EndTime:     time.Now().Add(27 * time.Hour),
		UserEmail:   user.Email,
		Recurrence:  models.RecurrenceNone,
	}
	err = service.CreateScheduleEntry(nonOverlappingEvent, user.Email, user.IsAdmin)
	if err != nil {
		t.Errorf("Unexpected error for non-overlapping event: %v", err)
	}
}
