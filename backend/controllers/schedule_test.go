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

	// Clean up before test
	testDB.CleanupTestDB(t)

	// Insert test data
	testDB.InsertTestData(t, testDB.Fixtures.TestEvent)

	// Create controller with test DB
	sc := &ScheduleController{db: testDB.DB}

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
	assert.Equal(t, testDB.Fixtures.TestEvent.Title, entries[0].Title)
}

func TestCreateScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Clean up before test
	testDB.CleanupTestDB(t)

	sc := &ScheduleController{db: testDB.DB}

	// Convert entry to JSON
	jsonData, err := json.Marshal(testDB.Fixtures.NewEvent)
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
	assert.Equal(t, testDB.Fixtures.NewEvent.Title, response.Title)
	assert.Equal(t, testDB.Fixtures.NewEvent.Description, response.Description)
	assert.NotZero(t, response.ID)
}

func TestUpdateScheduleEntry(t *testing.T) {
	// Setup
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()

	// Clean up before test
	testDB.CleanupTestDB(t)

	// Insert test data
	id := testDB.InsertTestData(t, testDB.Fixtures.OriginalEvent)

	sc := &ScheduleController{db: testDB.DB}

	// Create update data
	updatedEvent := testDB.Fixtures.OriginalEvent
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

	// Clean up before test
	testDB.CleanupTestDB(t)

	// Insert test data
	id := testDB.InsertTestData(t, testDB.Fixtures.EventToDelete)

	sc := &ScheduleController{db: testDB.DB}

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
	err := testDB.DB.QueryRow("SELECT COUNT(*) FROM schedule_entries").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
