package services

import (
	"database/sql"
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

func TestCreateRecurringScheduleEntry(t *testing.T) {
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

	// Create base time for all tests
	baseTime := time.Now().Add(24 * time.Hour)
	weeklyEndTime := baseTime.AddDate(0, 0, 14)
	biweeklyEndTime := baseTime.AddDate(0, 0, 28)
	monthlyEndTime := baseTime.AddDate(0, 2, 0)
	invalidEndTime := baseTime.AddDate(0, 0, -1)

	tests := []struct {
		name           string
		entry          *models.ScheduleEntry
		expectedCount  int
		expectedError  string
		checkInstances func(*testing.T, *sql.DB, int)
	}{
		{
			name: "weekly recurring event",
			entry: &models.ScheduleEntry{
				Title:             "Weekly Meeting",
				Description:       "Team sync",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(time.Hour),
				UserEmail:         user.Email,
				Recurrence:        models.RecurrenceWeekly,
				RecurrenceEndDate: &weeklyEndTime,
			},
			expectedCount: 3, // Original + 2 instances
			expectedError: "",
			checkInstances: func(t *testing.T, db *sql.DB, parentID int) {
				var count int
				err := db.QueryRow(`
					SELECT COUNT(*) FROM schedule_entries 
					WHERE parent_event_id = $1
				`, parentID).Scan(&count)
				if err != nil {
					t.Fatalf("Failed to check instances: %v", err)
				}
				if count != 2 { // Excluding parent
					t.Errorf("Expected 2 instances, got %d", count)
				}
			},
		},
		{
			name: "biweekly recurring event",
			entry: &models.ScheduleEntry{
				Title:             "Biweekly Meeting",
				Description:       "Team sync",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(time.Hour),
				UserEmail:         user.Email,
				Recurrence:        models.RecurrenceBiweekly,
				RecurrenceEndDate: &biweeklyEndTime,
			},
			expectedCount: 3, // Original + 2 instances
			expectedError: "",
			checkInstances: func(t *testing.T, db *sql.DB, parentID int) {
				var count int
				err := db.QueryRow(`
					SELECT COUNT(*) FROM schedule_entries 
					WHERE parent_event_id = $1
				`, parentID).Scan(&count)
				if err != nil {
					t.Fatalf("Failed to check instances: %v", err)
				}
				if count != 2 { // Excluding parent
					t.Errorf("Expected 2 instances, got %d", count)
				}
			},
		},
		{
			name: "monthly recurring event",
			entry: &models.ScheduleEntry{
				Title:             "Monthly Meeting",
				Description:       "Team sync",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(time.Hour),
				UserEmail:         user.Email,
				Recurrence:        models.RecurrenceMonthly,
				RecurrenceEndDate: &monthlyEndTime,
			},
			expectedCount: 3, // Original + 2 instances
			expectedError: "",
			checkInstances: func(t *testing.T, db *sql.DB, parentID int) {
				var count int
				err := db.QueryRow(`
					SELECT COUNT(*) FROM schedule_entries 
					WHERE parent_event_id = $1
				`, parentID).Scan(&count)
				if err != nil {
					t.Fatalf("Failed to check instances: %v", err)
				}
				if count != 2 { // Excluding parent
					t.Errorf("Expected 2 instances, got %d", count)
				}
			},
		},
		{
			name: "recurring event without end date",
			entry: &models.ScheduleEntry{
				Title:      "Invalid Recurring Event",
				StartTime:  baseTime,
				EndTime:    baseTime.Add(time.Hour),
				UserEmail:  user.Email,
				Recurrence: models.RecurrenceWeekly,
			},
			expectedCount: 1, // Only the parent event
			expectedError: "",
			checkInstances: func(t *testing.T, db *sql.DB, parentID int) {
				var count int
				err := db.QueryRow(`
					SELECT COUNT(*) FROM schedule_entries 
					WHERE parent_event_id = $1
				`, parentID).Scan(&count)
				if err != nil {
					t.Fatalf("Failed to check instances: %v", err)
				}
				if count != 0 {
					t.Errorf("Expected 0 instances, got %d", count)
				}
			},
		},
		{
			name: "recurring event with end date before start",
			entry: &models.ScheduleEntry{
				Title:             "Invalid End Date",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(time.Hour),
				UserEmail:         user.Email,
				Recurrence:        models.RecurrenceWeekly,
				RecurrenceEndDate: &invalidEndTime,
			},
			expectedCount: 1, // Only the parent event
			expectedError: "",
			checkInstances: func(t *testing.T, db *sql.DB, parentID int) {
				var count int
				err := db.QueryRow(`
					SELECT COUNT(*) FROM schedule_entries 
					WHERE parent_event_id = $1
				`, parentID).Scan(&count)
				if err != nil {
					t.Fatalf("Failed to check instances: %v", err)
				}
				if count != 0 {
					t.Errorf("Expected 0 instances, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing schedule entries
			_, err := testDB.Exec(`DELETE FROM schedule_entries`)
			if err != nil {
				t.Fatalf("Failed to clean up schedule entries: %v", err)
			}

			err = service.CreateScheduleEntry(tt.entry, tt.entry.UserEmail, user.IsAdmin)
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

			// Verify total number of entries
			var count int
			err = testDB.QueryRow(`
				SELECT COUNT(*) FROM schedule_entries 
				WHERE title = $1
			`, tt.entry.Title).Scan(&count)
			if err != nil {
				t.Fatalf("Failed to verify entry count: %v", err)
			}
			if count != tt.expectedCount {
				t.Errorf("Expected %d entries, got %d", tt.expectedCount, count)
			}

			// Check instances if provided
			if tt.checkInstances != nil {
				tt.checkInstances(t, testDB, tt.entry.ID)
			}
		})
	}
}

func TestGenerateRecurringInstances(t *testing.T) {
	// Create base time for all tests
	baseTime := time.Now().Add(24 * time.Hour)
	weeklyEndTime := baseTime.AddDate(0, 0, 14)
	biweeklyEndTime := baseTime.AddDate(0, 0, 28)
	monthlyEndTime := baseTime.AddDate(0, 2, 0)

	tests := []struct {
		name           string
		parentEvent    *models.ScheduleEntry
		expectedCount  int
		expectedError  string
		checkInstances func(*testing.T, []models.ScheduleEntry)
	}{
		{
			name: "weekly recurring event with correct duration",
			parentEvent: &models.ScheduleEntry{
				ID:                1,
				Title:             "Weekly Meeting",
				Description:       "Team sync",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(2 * time.Hour),
				UserEmail:         "test@example.com",
				Recurrence:        models.RecurrenceWeekly,
				RecurrenceEndDate: &weeklyEndTime,
			},
			expectedCount: 2,
			expectedError: "",
			checkInstances: func(t *testing.T, instances []models.ScheduleEntry) {
				// Check number of instances
				if len(instances) != 2 {
					t.Errorf("Expected 2 instances, got %d", len(instances))
				}

				// Check first instance
				firstInstance := instances[0]
				if firstInstance.Title != "Weekly Meeting" {
					t.Errorf("Expected title 'Weekly Meeting', got '%s'", firstInstance.Title)
				}
				if firstInstance.Description != "Team sync" {
					t.Errorf("Expected description 'Team sync', got '%s'", firstInstance.Description)
				}
				if firstInstance.StartTime != baseTime.AddDate(0, 0, 7) {
					t.Errorf("Expected start time %v, got %v", baseTime.AddDate(0, 0, 7), firstInstance.StartTime)
				}
				if firstInstance.EndTime != baseTime.AddDate(0, 0, 7).Add(2*time.Hour) {
					t.Errorf("Expected end time %v, got %v", baseTime.AddDate(0, 0, 7).Add(2*time.Hour), firstInstance.EndTime)
				}
				if firstInstance.Recurrence != models.RecurrenceNone {
					t.Errorf("Expected recurrence 'none', got '%s'", firstInstance.Recurrence)
				}
				if firstInstance.RecurrenceEndDate != nil {
					t.Error("Expected no recurrence end date")
				}
				if *firstInstance.ParentEventID != 1 {
					t.Errorf("Expected parent event ID 1, got %d", *firstInstance.ParentEventID)
				}

				// Check second instance
				secondInstance := instances[1]
				if secondInstance.StartTime != baseTime.AddDate(0, 0, 14) {
					t.Errorf("Expected start time %v, got %v", baseTime.AddDate(0, 0, 14), secondInstance.StartTime)
				}
				if secondInstance.EndTime != baseTime.AddDate(0, 0, 14).Add(2*time.Hour) {
					t.Errorf("Expected end time %v, got %v", baseTime.AddDate(0, 0, 14).Add(2*time.Hour), secondInstance.EndTime)
				}
			},
		},
		{
			name: "biweekly recurring event with correct duration",
			parentEvent: &models.ScheduleEntry{
				ID:                1,
				Title:             "Biweekly Meeting",
				Description:       "Team sync",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(2 * time.Hour),
				UserEmail:         "test@example.com",
				Recurrence:        models.RecurrenceBiweekly,
				RecurrenceEndDate: &biweeklyEndTime,
			},
			expectedCount: 2,
			expectedError: "",
			checkInstances: func(t *testing.T, instances []models.ScheduleEntry) {
				if len(instances) != 2 {
					t.Errorf("Expected 2 instances, got %d", len(instances))
				}

				// Check first instance
				firstInstance := instances[0]
				if firstInstance.StartTime != baseTime.AddDate(0, 0, 14) {
					t.Errorf("Expected start time %v, got %v", baseTime.AddDate(0, 0, 14), firstInstance.StartTime)
				}
				if firstInstance.EndTime != baseTime.AddDate(0, 0, 14).Add(2*time.Hour) {
					t.Errorf("Expected end time %v, got %v", baseTime.AddDate(0, 0, 14).Add(2*time.Hour), firstInstance.EndTime)
				}

				// Check second instance
				secondInstance := instances[1]
				if secondInstance.StartTime != baseTime.AddDate(0, 0, 28) {
					t.Errorf("Expected start time %v, got %v", baseTime.AddDate(0, 0, 28), secondInstance.StartTime)
				}
				if secondInstance.EndTime != baseTime.AddDate(0, 0, 28).Add(2*time.Hour) {
					t.Errorf("Expected end time %v, got %v", baseTime.AddDate(0, 0, 28).Add(2*time.Hour), secondInstance.EndTime)
				}
			},
		},
		{
			name: "monthly recurring event with correct duration",
			parentEvent: &models.ScheduleEntry{
				ID:                1,
				Title:             "Monthly Meeting",
				Description:       "Team sync",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(2 * time.Hour),
				UserEmail:         "test@example.com",
				Recurrence:        models.RecurrenceMonthly,
				RecurrenceEndDate: &monthlyEndTime,
			},
			expectedCount: 2,
			expectedError: "",
			checkInstances: func(t *testing.T, instances []models.ScheduleEntry) {
				if len(instances) != 2 {
					t.Errorf("Expected 2 instances, got %d", len(instances))
				}

				// Check first instance
				firstInstance := instances[0]
				if firstInstance.StartTime != baseTime.AddDate(0, 1, 0) {
					t.Errorf("Expected start time %v, got %v", baseTime.AddDate(0, 1, 0), firstInstance.StartTime)
				}
				if firstInstance.EndTime != baseTime.AddDate(0, 1, 0).Add(2*time.Hour) {
					t.Errorf("Expected end time %v, got %v", baseTime.AddDate(0, 1, 0).Add(2*time.Hour), firstInstance.EndTime)
				}

				// Check second instance
				secondInstance := instances[1]
				if secondInstance.StartTime != baseTime.AddDate(0, 2, 0) {
					t.Errorf("Expected start time %v, got %v", baseTime.AddDate(0, 2, 0), secondInstance.StartTime)
				}
				if secondInstance.EndTime != baseTime.AddDate(0, 2, 0).Add(2*time.Hour) {
					t.Errorf("Expected end time %v, got %v", baseTime.AddDate(0, 2, 0).Add(2*time.Hour), secondInstance.EndTime)
				}
			},
		},
		{
			name: "non-recurring event",
			parentEvent: &models.ScheduleEntry{
				ID:                1,
				Title:             "One-time Meeting",
				Description:       "Team sync",
				StartTime:         baseTime,
				EndTime:           baseTime.Add(2 * time.Hour),
				UserEmail:         "test@example.com",
				Recurrence:        models.RecurrenceNone,
				RecurrenceEndDate: &weeklyEndTime,
			},
			expectedCount: 0,
			expectedError: "",
			checkInstances: func(t *testing.T, instances []models.ScheduleEntry) {
				if len(instances) != 0 {
					t.Errorf("Expected 0 instances, got %d", len(instances))
				}
			},
		},
		{
			name: "recurring event without end date",
			parentEvent: &models.ScheduleEntry{
				ID:         1,
				Title:      "Invalid Meeting",
				StartTime:  baseTime,
				EndTime:    baseTime.Add(2 * time.Hour),
				UserEmail:  "test@example.com",
				Recurrence: models.RecurrenceWeekly,
			},
			expectedCount: 0,
			expectedError: "",
			checkInstances: func(t *testing.T, instances []models.ScheduleEntry) {
				if len(instances) != 0 {
					t.Errorf("Expected 0 instances, got %d", len(instances))
				}
			},
		},
		{
			name: "recurring event with end date before start",
			parentEvent: &models.ScheduleEntry{
				ID:         1,
				Title:      "Invalid Meeting",
				StartTime:  baseTime,
				EndTime:    baseTime.Add(2 * time.Hour),
				UserEmail:  "test@example.com",
				Recurrence: models.RecurrenceWeekly,
				RecurrenceEndDate: func() *time.Time {
					t := time.Time{}.AddDate(0, 0, -1)
					return &t
				}(),
			},
			expectedCount: 0,
			expectedError: "",
			checkInstances: func(t *testing.T, instances []models.ScheduleEntry) {
				if len(instances) != 0 {
					t.Errorf("Expected 0 instances, got %d", len(instances))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewScheduleService(nil) // We don't need a DB for this test
			instances, err := service.generateRecurringInstances(tt.parentEvent)

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

			if len(instances) != tt.expectedCount {
				t.Errorf("Expected %d instances, got %d", tt.expectedCount, len(instances))
			}

			if tt.checkInstances != nil {
				tt.checkInstances(t, instances)
			}
		})
	}
}

func TestBulkInsertEvents(t *testing.T) {
	// Setup test database
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/nextpitch_test?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Create service instance
	service := NewScheduleService(db)

	// Create test user
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING", userEmail, "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Get user ID
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", userEmail).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	// Create base time for tests
	baseTime := time.Now().Add(24 * time.Hour)
	baseTime = time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		events         []models.ScheduleEntry
		existingEvents []models.ScheduleEntry
		expectedResult *BulkInsertResult
		checkResult    func(*testing.T, *BulkInsertResult)
	}{
		{
			name: "successful bulk insert",
			events: []models.ScheduleEntry{
				{
					Title:     "Event 1",
					StartTime: baseTime,
					EndTime:   baseTime.Add(time.Hour),
					UserEmail: userEmail,
				},
				{
					Title:     "Event 2",
					StartTime: baseTime.Add(2 * time.Hour),
					EndTime:   baseTime.Add(3 * time.Hour),
					UserEmail: userEmail,
				},
			},
			expectedResult: &BulkInsertResult{
				SuccessfullyInserted: 2,
				FailedInstances:      []FailedInstance{},
			},
			checkResult: func(t *testing.T, result *BulkInsertResult) {
				if result.SuccessfullyInserted != 2 {
					t.Errorf("Expected 2 successful inserts, got %d", result.SuccessfullyInserted)
				}
				if len(result.FailedInstances) != 0 {
					t.Errorf("Expected no failed instances, got %d", len(result.FailedInstances))
				}
			},
		},
		{
			name: "partial success with overlaps",
			existingEvents: []models.ScheduleEntry{
				{
					Title:     "Existing Event",
					StartTime: baseTime.Add(time.Hour),
					EndTime:   baseTime.Add(2 * time.Hour),
					UserEmail: userEmail,
				},
			},
			events: []models.ScheduleEntry{
				{
					Title:     "Event 1",
					StartTime: baseTime,
					EndTime:   baseTime.Add(time.Hour),
					UserEmail: userEmail,
				},
				{
					Title:     "Event 2 (Overlapping)",
					StartTime: baseTime.Add(30 * time.Minute),
					EndTime:   baseTime.Add(90 * time.Minute),
					UserEmail: userEmail,
				},
				{
					Title:     "Event 3",
					StartTime: baseTime.Add(2 * time.Hour),
					EndTime:   baseTime.Add(3 * time.Hour),
					UserEmail: userEmail,
				},
			},
			expectedResult: &BulkInsertResult{
				SuccessfullyInserted: 2,
				FailedInstances: []FailedInstance{
					{
						Event: models.ScheduleEntry{
							Title:     "Event 2 (Overlapping)",
							StartTime: baseTime.Add(30 * time.Minute),
							EndTime:   baseTime.Add(90 * time.Minute),
						},
						Reason: "overlap",
					},
				},
			},
			checkResult: func(t *testing.T, result *BulkInsertResult) {
				if result.SuccessfullyInserted != 2 {
					t.Errorf("Expected 2 successful inserts, got %d", result.SuccessfullyInserted)
				}
				if len(result.FailedInstances) != 1 {
					t.Errorf("Expected 1 failed instance, got %d", len(result.FailedInstances))
				}
				if result.FailedInstances[0].Reason != "overlap" {
					t.Errorf("Expected overlap reason, got %s", result.FailedInstances[0].Reason)
				}
			},
		},
		{
			name: "all events fail due to overlaps",
			existingEvents: []models.ScheduleEntry{
				{
					Title:     "Existing Event 1",
					StartTime: baseTime,
					EndTime:   baseTime.Add(2 * time.Hour),
					UserEmail: userEmail,
				},
				{
					Title:     "Existing Event 2",
					StartTime: baseTime.Add(2 * time.Hour),
					EndTime:   baseTime.Add(4 * time.Hour),
					UserEmail: userEmail,
				},
			},
			events: []models.ScheduleEntry{
				{
					Title:     "Event 1 (Overlapping)",
					StartTime: baseTime.Add(30 * time.Minute),
					EndTime:   baseTime.Add(90 * time.Minute),
					UserEmail: userEmail,
				},
				{
					Title:     "Event 2 (Overlapping)",
					StartTime: baseTime.Add(2*time.Hour + 30*time.Minute),
					EndTime:   baseTime.Add(3*time.Hour + 30*time.Minute),
					UserEmail: userEmail,
				},
			},
			expectedResult: &BulkInsertResult{
				SuccessfullyInserted: 0,
				FailedInstances: []FailedInstance{
					{
						Event: models.ScheduleEntry{
							Title:     "Event 1 (Overlapping)",
							StartTime: baseTime.Add(30 * time.Minute),
							EndTime:   baseTime.Add(90 * time.Minute),
						},
						Reason: "overlap",
					},
					{
						Event: models.ScheduleEntry{
							Title:     "Event 2 (Overlapping)",
							StartTime: baseTime.Add(2*time.Hour + 30*time.Minute),
							EndTime:   baseTime.Add(3*time.Hour + 30*time.Minute),
						},
						Reason: "overlap",
					},
				},
			},
			checkResult: func(t *testing.T, result *BulkInsertResult) {
				if result.SuccessfullyInserted != 0 {
					t.Errorf("Expected 0 successful inserts, got %d", result.SuccessfullyInserted)
				}
				if len(result.FailedInstances) != 2 {
					t.Errorf("Expected 2 failed instances, got %d", len(result.FailedInstances))
				}
				for _, failed := range result.FailedInstances {
					if failed.Reason != "overlap" {
						t.Errorf("Expected overlap reason, got %s", failed.Reason)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing events
			_, err = db.Exec("DELETE FROM schedule_entries")
			if err != nil {
				t.Fatalf("Failed to clear schedule entries: %v", err)
			}

			// Insert existing events if any
			for _, event := range tt.existingEvents {
				_, err = db.Exec(`
					INSERT INTO schedule_entries (
						title, start_time, end_time, user_id, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6)
				`, event.Title, event.StartTime, event.EndTime, userID, time.Now(), time.Now())
				if err != nil {
					t.Fatalf("Failed to insert existing event: %v", err)
				}
			}

			// Run the test
			result, err := service.bulkInsertEvents(tt.events)
			if err != nil {
				t.Fatalf("bulkInsertEvents failed: %v", err)
			}

			// Check the result
			tt.checkResult(t, result)
		})
	}
}
