package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Setup routes
	r.GET("/api/schedule", getScheduleEntries)
	r.POST("/api/schedule", createScheduleEntry)
	r.PUT("/api/schedule/:id", updateScheduleEntry)
	r.DELETE("/api/schedule/:id", deleteScheduleEntry)

	return r
}

func TestScheduleEndpoints(t *testing.T) {
	router := setupTestRouter()

	// Test data
	now := time.Now()
	testEntry := ScheduleEntry{
		Title:       "Test Meeting",
		Description: "Test Description",
		StartTime:   now.Add(1 * time.Hour),
		EndTime:     now.Add(2 * time.Hour),
	}

	// Test Create Schedule Entry
	t.Run("create schedule entry", func(t *testing.T) {
		body, _ := json.Marshal(testEntry)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/schedule", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response ScheduleEntry
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotZero(t, response.ID)
		assert.Equal(t, testEntry.Title, response.Title)
		assert.Equal(t, testEntry.Description, response.Description)
		assert.Equal(t, testEntry.StartTime.Format(time.RFC3339), response.StartTime.Format(time.RFC3339))
		assert.Equal(t, testEntry.EndTime.Format(time.RFC3339), response.EndTime.Format(time.RFC3339))

		// Store the ID for later tests
		testEntry.ID = response.ID
	})

	// Test Get Schedule Entries
	t.Run("get schedule entries", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/schedule", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var entries []ScheduleEntry
		err := json.Unmarshal(w.Body.Bytes(), &entries)
		assert.NoError(t, err)
		assert.NotEmpty(t, entries)
	})

	// Test Update Schedule Entry
	t.Run("update schedule entry", func(t *testing.T) {
		updatedEntry := testEntry
		updatedEntry.Title = "Updated Meeting"
		updatedEntry.Description = "Updated Description"

		body, _ := json.Marshal(updatedEntry)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/schedule/"+string(testEntry.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ScheduleEntry
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, updatedEntry.Title, response.Title)
		assert.Equal(t, updatedEntry.Description, response.Description)
	})

	// Test Delete Schedule Entry
	t.Run("delete schedule entry", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/schedule/"+string(testEntry.ID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Schedule entry deleted successfully", response["message"])
	})

	// Test Get Non-existent Entry
	t.Run("get non-existent entry", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/schedule/999", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestScheduleValidation(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name       string
		entry      ScheduleEntry
		wantStatus int
	}{
		{
			name: "missing required fields",
			entry: ScheduleEntry{
				Description: "Test Description",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "end time before start time",
			entry: ScheduleEntry{
				Title:       "Test Meeting",
				Description: "Test Description",
				StartTime:   time.Now().Add(2 * time.Hour),
				EndTime:     time.Now().Add(1 * time.Hour),
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.entry)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/schedule", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
