package services

import (
	"testing"
	"time"

	_ "github.com/lib/pq" // postgres driver
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/test/helpers"
)

func TestCreateScheduleEntry(t *testing.T) {
	// Setup test database and fixtures
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewScheduleService(testDB.DB)

	tests := []struct {
		name          string
		entry         *models.ScheduleEntry
		isAdmin       bool
		expectedError string
	}{
		{
			name: "successful creation by admin",
			entry: &models.ScheduleEntry{
				Title:     "Test Event",
				StartTime: fixedTime(),
				EndTime:   fixedTime().Add(3 * time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin: true,
		},
		{
			name: "non-admin cannot create long event",
			entry: &models.ScheduleEntry{
				Title:     "Test Event",
				StartTime: fixedTime(),
				EndTime:   fixedTime().Add(3 * time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin:       false,
			expectedError: "non-admin users cannot create events longer than 2 hours",
		},
		{
			name: "non-admin can create short event",
			entry: &models.ScheduleEntry{
				Title:     "Test Event",
				StartTime: fixedTime(),
				EndTime:   fixedTime().Add(time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			testDB.CleanupTestDB(t)

			// Create the user first
			_, err := testDB.DB.Exec(`
				INSERT INTO users (email, first_name, last_name, admin, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, tt.entry.UserEmail, "Test", "User", tt.isAdmin, time.Now(), time.Now())
			assert.NoError(t, err)

			err = service.CreateScheduleEntry(tt.entry, tt.entry.UserEmail, tt.isAdmin)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)

			// Verify the entry was created
			var count int
			err = testDB.DB.QueryRow(`
				SELECT COUNT(*) 
				FROM schedule_entries se
				JOIN users u ON se.user_id = u.id
				WHERE u.email = $1
			`, tt.entry.UserEmail).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

func TestUpdateScheduleEntry(t *testing.T) {
	// Setup test database and fixtures
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewScheduleService(testDB.DB)

	// Insert a test event to update
	testEvent := testDB.Fixtures.TestEvent
	id := testDB.InsertTestData(t, testEvent)

	tests := []struct {
		name          string
		entry         *models.ScheduleEntry
		isAdmin       bool
		expectedError string
	}{
		{
			name: "successful update by admin",
			entry: &models.ScheduleEntry{
				ID:        id,
				Title:     "Updated Event",
				StartTime: fixedTime(),
				EndTime:   fixedTime().Add(3 * time.Hour),
				UserEmail: testEvent.UserEmail,
			},
			isAdmin: true,
		},
		{
			name: "non-admin cannot update to long event",
			entry: &models.ScheduleEntry{
				ID:        id,
				Title:     "Updated Event",
				StartTime: fixedTime(),
				EndTime:   fixedTime().Add(3 * time.Hour),
				UserEmail: testEvent.UserEmail,
			},
			isAdmin:       false,
			expectedError: "non-admin users cannot create events longer than 2 hours",
		},
		{
			name: "non-admin can update to short event",
			entry: &models.ScheduleEntry{
				ID:        id,
				Title:     "Updated Event",
				StartTime: fixedTime(),
				EndTime:   fixedTime().Add(time.Hour),
				UserEmail: testEvent.UserEmail,
			},
			isAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			testDB.CleanupTestDB(t)
			// Re-insert the test event
			id = testDB.InsertTestData(t, testEvent)
			tt.entry.ID = id

			err := service.UpdateScheduleEntry(tt.entry, tt.entry.UserEmail, tt.isAdmin)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)

			// Verify the entry was updated
			var title string
			err = testDB.DB.QueryRow(`
				SELECT se.title 
				FROM schedule_entries se
				JOIN users u ON se.user_id = u.id
				WHERE se.id = $1 AND u.email = $2
			`, id, tt.entry.UserEmail).Scan(&title)
			assert.NoError(t, err)
			assert.Equal(t, "Updated Event", title)
		})
	}
}

func TestCheckOverlappingEvents(t *testing.T) {
	// Setup test database and fixtures
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	service := NewScheduleService(testDB.DB)

	// Create a fixed time for testing
	baseTime := fixedTime()

	// Create test event data
	testEvent := models.ScheduleEntry{
		Title:     "Test Event",
		StartTime: baseTime,
		EndTime:   baseTime.Add(time.Hour),
		UserEmail: "test@example.com",
	}

	tests := []struct {
		name          string
		startTime     time.Time
		endTime       time.Time
		excludeID     int
		expectedCount int
	}{
		{
			name:          "no overlapping events",
			startTime:     baseTime.Add(24 * time.Hour),
			endTime:       baseTime.Add(25 * time.Hour),
			excludeID:     0,
			expectedCount: 0,
		},
		{
			name:          "overlapping event exists",
			startTime:     baseTime,
			endTime:       baseTime.Add(time.Hour),
			excludeID:     0,
			expectedCount: 1,
		},
		{
			name:          "overlapping event excluded",
			startTime:     baseTime,
			endTime:       baseTime.Add(time.Hour),
			excludeID:     0, // Will be set to the actual ID after insertion
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			testDB.CleanupTestDB(t)

			// Create the user first
			_, err := testDB.DB.Exec(`
				INSERT INTO users (email, first_name, last_name, admin, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, testEvent.UserEmail, "Test", "User", false, time.Now(), time.Now())
			assert.NoError(t, err)

			// Insert the test event
			id := testDB.InsertTestData(t, testEvent)

			// Set the excludeID for the "overlapping event excluded" test
			if tt.name == "overlapping event excluded" {
				tt.excludeID = id
			}

			var count int
			if tt.excludeID > 0 {
				overlapping, err := service.checkOverlappingEventsForUpdate(tt.excludeID, tt.startTime, tt.endTime)
				assert.NoError(t, err)
				if overlapping {
					count = 1
				}
			} else {
				overlapping, err := service.checkOverlappingEvents(tt.startTime, tt.endTime)
				assert.NoError(t, err)
				if overlapping {
					count = 1
				}
			}
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}
