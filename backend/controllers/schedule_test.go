package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use test database configuration
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "postgres"
	dbname := "nextpitch_test"
	sslmode := "disable"

	connStr := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up the test database before each test
	_, err = testDB.Exec("TRUNCATE TABLE schedule_entries CASCADE")
	if err != nil {
		t.Fatalf("Failed to clean up test database: %v", err)
	}

	return testDB
}

func TestGetScheduleEntries(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Create test data
	_, err := testDB.Exec(`
		INSERT INTO schedule_entries (title, description, start_time, end_time)
		VALUES ($1, $2, $3, $4)
	`, "Test Event", "Test Description", time.Now(), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

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
	err = json.Unmarshal(w.Body.Bytes(), &entries)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Test Event", entries[0].Title)
}

func TestCreateScheduleEntry(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	sc := &ScheduleController{db: testDB}

	// Create test data
	entry := models.ScheduleEntry{
		Title:       "New Event",
		Description: "New Description",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
	}

	// Convert entry to JSON
	jsonData, err := json.Marshal(entry)
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
	assert.Equal(t, entry.Title, response.Title)
	assert.Equal(t, entry.Description, response.Description)
	assert.NotZero(t, response.ID)
}

func TestUpdateScheduleEntry(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insert test data
	var id int
	err := testDB.QueryRow(`
		INSERT INTO schedule_entries (title, description, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "Original Event", "Original Description", time.Now(), time.Now().Add(time.Hour)).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	sc := &ScheduleController{db: testDB}

	// Create update data
	entry := models.ScheduleEntry{
		ID:          id,
		Title:       "Updated Event",
		Description: "Updated Description",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
	}

	// Convert entry to JSON
	jsonData, err := json.Marshal(entry)
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
	assert.Equal(t, entry.Title, response.Title)
	assert.Equal(t, entry.Description, response.Description)
}

func TestDeleteScheduleEntry(t *testing.T) {
	// Setup
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insert test data
	var id int
	err := testDB.QueryRow(`
		INSERT INTO schedule_entries (title, description, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "Event to Delete", "Description", time.Now(), time.Now().Add(time.Hour)).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

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
	err = testDB.QueryRow("SELECT COUNT(*) FROM schedule_entries").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
