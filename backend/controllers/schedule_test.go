package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/test/helpers"
)

func TestGetScheduleEntries(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Load fixtures
	fixtures := helpers.LoadFixtures(t)

	// Insert test data
	helpers.InsertTestData(t, testDB, fixtures.TestEvent)

	// Create controller with test DB
	sc := &ScheduleController{db: testDB}

	// Setup Gin router
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test
	sc.GetScheduleEntries(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var entries []models.ScheduleEntry
	err := json.Unmarshal(w.Body.Bytes(), &entries)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, fixtures.TestEvent.Title, entries[0].Title)
}

func TestCreateScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Load fixtures
	fixtures := helpers.LoadFixtures(t)

	sc := &ScheduleController{db: testDB}

	// Convert entry to JSON
	jsonData, err := json.Marshal(fixtures.NewEvent)
	assert.NoError(t, err)

	// Setup Gin router
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/schedule", bytes.NewBuffer(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")

	// Test
	sc.CreateScheduleEntry(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.ScheduleEntry
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, fixtures.NewEvent.Title, response.Title)
	assert.Equal(t, fixtures.NewEvent.Description, response.Description)
	assert.NotZero(t, response.ID)
}

func TestUpdateScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Load fixtures
	fixtures := helpers.LoadFixtures(t)

	// Insert test data
	id := helpers.InsertTestData(t, testDB, fixtures.OriginalEvent)

	sc := &ScheduleController{db: testDB}

	// Create update data
	updatedEvent := fixtures.OriginalEvent
	updatedEvent.ID = id
	updatedEvent.Title = "Updated Event"
	updatedEvent.Description = "Updated Description"

	// Convert entry to JSON
	jsonData, err := json.Marshal(updatedEvent)
	assert.NoError(t, err)

	// Setup Gin router
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", id)}}
	c.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/schedule/%d", id), bytes.NewBuffer(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")

	// Test
	sc.UpdateScheduleEntry(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ScheduleEntry
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, updatedEvent.Title, response.Title)
	assert.Equal(t, updatedEvent.Description, response.Description)
}

func TestDeleteScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Load fixtures
	fixtures := helpers.LoadFixtures(t)

	// Insert test data
	id := helpers.InsertTestData(t, testDB, fixtures.EventToDelete)

	sc := &ScheduleController{db: testDB}

	// Setup Gin router
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: fmt.Sprintf("%d", id)}}

	// Test
	sc.DeleteScheduleEntry(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM schedule_entries").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
