package helpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"nextpitch.com/backend/models"
)

type ScheduleFixtures struct {
	TestEvent     models.ScheduleEntry `json:"test_event"`
	NewEvent      models.ScheduleEntry `json:"new_event"`
	OriginalEvent models.ScheduleEntry `json:"original_event"`
	EventToDelete models.ScheduleEntry `json:"event_to_delete"`
}

func LoadFixtures(t interface{ Fatal(args ...interface{}) }) *ScheduleFixtures {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to get current working directory: %v", err))
	}

	// Find the backend directory by looking for go.mod
	backendDir := cwd
	for {
		if _, err := os.Stat(filepath.Join(backendDir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(backendDir)
		if parent == backendDir {
			t.Fatal("Could not find go.mod file")
		}
		backendDir = parent
	}

	// Get the absolute path to the fixtures directory
	fixturesPath := filepath.Join(backendDir, "test", "fixtures", "schedule_entries.json")
	data, err := os.ReadFile(fixturesPath)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to read fixtures file: %v", err))
	}

	var fixtures ScheduleFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(fmt.Sprintf("Failed to parse fixtures: %v", err))
	}

	return &fixtures
}

func SetupTestDB(t interface{ Fatal(args ...interface{}) }) *sql.DB {
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
		t.Fatal(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Clean up the test database before each test
	_, err = testDB.Exec("TRUNCATE TABLE schedule_entries CASCADE")
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to clean up test database: %v", err))
	}

	return testDB
}

func InsertTestData(t interface{ Fatal(args ...interface{}) }, db *sql.DB, entry models.ScheduleEntry) int {
	var id int
	err := db.QueryRow(`
		INSERT INTO schedule_entries (title, description, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime).Scan(&id)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to insert test data: %v", err))
	}
	return id
}
