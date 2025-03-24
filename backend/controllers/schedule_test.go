package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
	"nextpitch.com/backend/test/helpers"
)

func TestGetScheduleEntries(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Create services
	scheduleService := services.NewScheduleService(testDB)
	userService := services.NewUserService(testDB)

	// Create controller with services
	sc := NewScheduleController(scheduleService, userService)

	// Create test user
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		IsAdmin: false,
	}
	err := userService.CreateUser(user)
	assert.NoError(t, err)

	// Create test entry
	entry := &models.ScheduleEntry{
		Title:       "Test Event",
		Description: "Test Description",
		StartTime:   time.Now().Add(24 * time.Hour),
		EndTime:     time.Now().Add(25 * time.Hour),
		UserEmail:   user.Email,
		Recurrence:  models.RecurrenceNone,
	}
	err = scheduleService.CreateScheduleEntry(entry, user.Email, user.IsAdmin)
	assert.NoError(t, err)

	// Setup Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_email", user.Email)

	// Test
	sc.GetScheduleEntries(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var entries []models.ScheduleEntry
	err = json.Unmarshal(w.Body.Bytes(), &entries)
	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 1)
	assert.Equal(t, entry.Title, entries[0].Title)
	assert.Equal(t, entry.Description, entries[0].Description)
}

func TestCreateScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Create services
	scheduleService := services.NewScheduleService(testDB)
	userService := services.NewUserService(testDB)

	// Create controller with services
	sc := NewScheduleController(scheduleService, userService)

	tests := []struct {
		name          string
		entry         *models.ScheduleEntry
		user          *models.User
		expectedCode  int
		expectedError string
	}{
		{
			name: "successful creation by admin",
			entry: &models.ScheduleEntry{
				Title:       "Test Event",
				Description: "Test Description",
				StartTime:   time.Now().Add(24 * time.Hour),
				EndTime:     time.Now().Add(48 * time.Hour),
				UserEmail:   "admin@example.com",
				Recurrence:  models.RecurrenceNone,
			},
			user: &models.User{
				Email:   "admin@example.com",
				Name:    "Test User",
				IsAdmin: true,
			},
			expectedCode:  http.StatusCreated,
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
			user: &models.User{
				Email:   "user@example.com",
				Name:    "Test User",
				IsAdmin: false,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "event duration exceeds maximum allowed duration for non-admin users",
		},
		{
			name: "non-admin can create short event",
			entry: &models.ScheduleEntry{
				Title:       "Short Event",
				Description: "Test Description",
				StartTime:   time.Now().Add(120 * time.Hour),
				EndTime:     time.Now().Add(121 * time.Hour),
				UserEmail:   "user2@example.com",
				Recurrence:  models.RecurrenceWeekly,
			},
			user: &models.User{
				Email:   "user2@example.com",
				Name:    "Test User",
				IsAdmin: false,
			},
			expectedCode:  http.StatusCreated,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user
			err := userService.CreateUser(tt.user)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Setup Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("user_email", tt.user.Email)

			// Set request body
			body, err := json.Marshal(tt.entry)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			c.Request = httptest.NewRequest("POST", "/schedule", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Test
			sc.CreateScheduleEntry(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
				return
			}

			var response models.ScheduleEntry
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.entry.Title, response.Title)
			assert.Equal(t, tt.entry.Description, response.Description)
			assert.NotZero(t, response.ID)
		})
	}
}

func TestUpdateScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Create services
	scheduleService := services.NewScheduleService(testDB)
	userService := services.NewUserService(testDB)

	// Create controller with services
	sc := NewScheduleController(scheduleService, userService)

	tests := []struct {
		name          string
		initialEntry  *models.ScheduleEntry
		updatedEntry  *models.ScheduleEntry
		user          *models.User
		expectedCode  int
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
			user: &models.User{
				Email:   "admin@example.com",
				Name:    "Test User",
				IsAdmin: true,
			},
			expectedCode:  http.StatusOK,
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
			user: &models.User{
				Email:   "user@example.com",
				Name:    "Test User",
				IsAdmin: false,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "event duration exceeds maximum allowed duration for non-admin users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user
			err := userService.CreateUser(tt.user)
			if err != nil {
				t.Fatalf("Failed to create test user: %v", err)
			}

			// Create initial entry
			tt.initialEntry.UserEmail = tt.user.Email
			err = scheduleService.CreateScheduleEntry(tt.initialEntry, tt.user.Email, tt.user.IsAdmin)
			if err != nil {
				t.Fatalf("Failed to create initial entry: %v", err)
			}

			// Setup update request
			tt.updatedEntry.ID = tt.initialEntry.ID
			tt.updatedEntry.UserEmail = tt.user.Email

			// Setup Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("user_email", tt.user.Email)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.initialEntry.ID)}}

			// Set request body
			body, err := json.Marshal(tt.updatedEntry)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			c.Request = httptest.NewRequest("PUT", "/schedule/"+strconv.Itoa(tt.initialEntry.ID), bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Test
			sc.UpdateScheduleEntry(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
				return
			}

			var response models.ScheduleEntry
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.updatedEntry.Title, response.Title)
			assert.Equal(t, tt.updatedEntry.Description, response.Description)
		})
	}
}

func TestDeleteScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Create services
	scheduleService := services.NewScheduleService(testDB)
	userService := services.NewUserService(testDB)

	// Create controller with services
	sc := NewScheduleController(scheduleService, userService)

	// Create test user
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		IsAdmin: false,
	}
	err := userService.CreateUser(user)
	assert.NoError(t, err)

	// Create test entry
	entry := &models.ScheduleEntry{
		Title:       "Test Event",
		Description: "Test Description",
		StartTime:   time.Now().Add(24 * time.Hour),
		EndTime:     time.Now().Add(25 * time.Hour),
		UserEmail:   user.Email,
		Recurrence:  models.RecurrenceNone,
	}
	err = scheduleService.CreateScheduleEntry(entry, user.Email, user.IsAdmin)
	assert.NoError(t, err)

	// Setup Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(entry.ID)}}
	c.Set("user_email", user.Email)
	c.Request = httptest.NewRequest("DELETE", "/schedule/"+strconv.Itoa(entry.ID), nil)
	c.Request.URL.RawQuery = fmt.Sprintf("delete_following=%v", false)

	// Test
	sc.DeleteScheduleEntry(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM schedule_entries").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteRecurringScheduleEntry(t *testing.T) {
	// Setup test database and services
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	scheduleService := services.NewScheduleService(testDB)
	userService := services.NewUserService(testDB)

	// Create test user
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		IsAdmin: true,
	}
	err := userService.CreateUser(user)
	assert.NoError(t, err)

	// Define base time for tests
	baseTime := time.Now().Add(24 * time.Hour)
	weeklyEndTime := baseTime.Add(14 * 24 * time.Hour) // 2 weeks

	tests := []struct {
		name            string
		setup           func() int64
		deleteFollowing bool
		expectedCount   int
		expectedError   string
	}{
		{
			name: "delete single event",
			setup: func() int64 {
				entry := &models.ScheduleEntry{
					Title:       "Single Event",
					Description: "Test Description",
					StartTime:   baseTime,
					EndTime:     baseTime.Add(time.Hour),
					UserEmail:   user.Email,
					Recurrence:  models.RecurrenceNone,
				}
				err := scheduleService.CreateScheduleEntry(entry, user.Email, user.IsAdmin)
				assert.NoError(t, err)
				return int64(entry.ID)
			},
			deleteFollowing: false,
			expectedCount:   0,
			expectedError:   "",
		},
		{
			name: "delete recurring event and instances after",
			setup: func() int64 {
				// Create a recurring event with instances
				entry := &models.ScheduleEntry{
					Title:             "Recurring Event",
					Description:       "Test Description",
					StartTime:         baseTime,
					EndTime:           baseTime.Add(time.Hour),
					UserEmail:         user.Email,
					Recurrence:        models.RecurrenceWeekly,
					RecurrenceEndDate: &weeklyEndTime,
				}
				err := scheduleService.CreateScheduleEntry(entry, user.Email, user.IsAdmin)
				assert.NoError(t, err)

				// Get the first instance ID (the one we'll delete)
				var instanceID int64
				err = testDB.QueryRow(`
					SELECT id FROM schedule_entries 
					WHERE parent_event_id = $1 
					ORDER BY start_time ASC 
					LIMIT 1
				`, entry.ID).Scan(&instanceID)
				assert.NoError(t, err)
				return instanceID
			},
			deleteFollowing: true,
			expectedCount:   1, // Only the parent event remains
			expectedError:   "",
		},
		{
			name: "delete recurring event that is the parent event and instances after",
			setup: func() int64 {
				// Create a recurring event with instances
				entry := &models.ScheduleEntry{
					Title:             "Recurring Event",
					Description:       "Test Description",
					StartTime:         baseTime,
					EndTime:           baseTime.Add(time.Hour),
					UserEmail:         user.Email,
					Recurrence:        models.RecurrenceWeekly,
					RecurrenceEndDate: &weeklyEndTime,
				}
				err := scheduleService.CreateScheduleEntry(entry, user.Email, user.IsAdmin)
				assert.NoError(t, err)
				return int64(entry.ID)
			},
			deleteFollowing: true,
			expectedCount:   0, // All events should be deleted
			expectedError:   "",
		},
		{
			name: "delete recurring event and not instances after",
			setup: func() int64 {
				// Create a recurring event with instances
				entry := &models.ScheduleEntry{
					Title:             "Recurring Event",
					Description:       "Test Description",
					StartTime:         baseTime,
					EndTime:           baseTime.Add(time.Hour),
					UserEmail:         user.Email,
					Recurrence:        models.RecurrenceWeekly,
					RecurrenceEndDate: &weeklyEndTime,
				}
				err := scheduleService.CreateScheduleEntry(entry, user.Email, user.IsAdmin)
				assert.NoError(t, err)

				// Get the first instance ID (the one we'll delete)
				var instanceID int64
				err = testDB.QueryRow(`
					SELECT id FROM schedule_entries 
					WHERE parent_event_id = $1 
					ORDER BY start_time ASC 
					LIMIT 1
				`, entry.ID).Scan(&instanceID)
				assert.NoError(t, err)
				return instanceID
			},
			deleteFollowing: false,
			expectedCount:   2, // Parent event + remaining instances
			expectedError:   "",
		},
		{
			name: "delete parent event without deleting instances",
			setup: func() int64 {
				// Create a recurring event with instances
				entry := &models.ScheduleEntry{
					Title:             "Parent Event",
					Description:       "Test Description",
					StartTime:         baseTime,
					EndTime:           baseTime.Add(time.Hour),
					UserEmail:         user.Email,
					Recurrence:        models.RecurrenceWeekly,
					RecurrenceEndDate: &weeklyEndTime,
				}
				err := scheduleService.CreateScheduleEntry(entry, user.Email, user.IsAdmin)
				assert.NoError(t, err)
				return int64(entry.ID)
			},
			deleteFollowing: false,
			expectedCount:   2, // Only the instances should remain
			expectedError:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing schedule entries
			_, err := testDB.Exec(`DELETE FROM schedule_entries`)
			if err != nil {
				t.Fatalf("Failed to clean up schedule entries: %v", err)
			}

			// Setup the test case and get the ID to delete
			eventID := tt.setup()

			// Setup Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "id", Value: strconv.FormatInt(eventID, 10)}}
			c.Set("user_email", user.Email)
			c.Request = httptest.NewRequest("DELETE", "/schedule/"+strconv.FormatInt(eventID, 10), nil)
			c.Request.URL.RawQuery = fmt.Sprintf("delete_following=%v", tt.deleteFollowing)

			// Test
			sc := NewScheduleController(scheduleService, userService)
			sc.DeleteScheduleEntry(c)

			// Assert
			if tt.expectedError != "" {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
				return
			}

			assert.Equal(t, http.StatusOK, w.Code)

			// Verify the number of remaining entries
			var count int
			err = testDB.QueryRow("SELECT COUNT(*) FROM schedule_entries").Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)

			// Verify remaining events
			var countRemaining int
			err = testDB.QueryRow(`
				SELECT COUNT(*) FROM schedule_entries se
				JOIN users u ON se.user_id = u.id
				WHERE u.email = $1 AND se.id != $2
			`, user.Email, eventID).Scan(&countRemaining)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, countRemaining)

			// Verify parent event ID for remaining events
			if tt.expectedCount > 0 {
				var parentID sql.NullInt64
				err := testDB.QueryRow(`
					SELECT se.parent_event_id FROM schedule_entries se
					JOIN users u ON se.user_id = u.id
					WHERE u.email = $1 AND se.id != $2
					ORDER BY se.start_time ASC LIMIT 1
				`, user.Email, eventID).Scan(&parentID)
				assert.NoError(t, err)

				if parentID.Valid {
					assert.Equal(t, eventID, parentID.Int64)
				} else {
					// If parent_event_id is NULL, this is a parent event itself
					assert.Equal(t, tt.expectedCount, countRemaining)
				}
			}
		})
	}
}
