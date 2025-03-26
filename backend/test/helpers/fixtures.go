package helpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"nextpitch.com/backend/models"
)

type ScheduleFixtures struct {
	TestEvent     models.ScheduleEntry `json:"test_event"`
	NewEvent      models.ScheduleEntry `json:"new_event"`
	OriginalEvent models.ScheduleEntry `json:"original_event"`
	EventToDelete models.ScheduleEntry `json:"event_to_delete"`
}

type VideoFixtures struct {
	TestVideo     models.VideoUpload `json:"test_video"`
	NewVideo      models.VideoUpload `json:"new_video"`
	VideoToDelete models.VideoUpload `json:"video_to_delete"`
}

type TestDB struct {
	DB               *sql.DB
	ScheduleFixtures *ScheduleFixtures
	VideoFixtures    *VideoFixtures
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

func runMigrations(t interface{ Fatal(args ...interface{}) }, dbname string) {
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

	// Run migrations
	cmd := exec.Command("migrate",
		"-database", fmt.Sprintf("postgres://postgres:postgres@localhost:5432/%s?sslmode=disable", dbname),
		"-path", filepath.Join(backendDir, "db", "migrations"),
		"up")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to run migrations: %v\nOutput: %s", err, output))
	}
}

func SetupTestDB(t interface{ Fatal(args ...interface{}) }) *sql.DB {
	// Use test database configuration
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "postgres"
	dbname := "nextpitch_test"
	sslmode := "disable"

	// Connect to the test database
	connStr := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Run migrations
	runMigrations(t, dbname)

	// Clean up the test database before each test
	_, err = testDB.Exec(`
		TRUNCATE TABLE video_uploads CASCADE;
		TRUNCATE TABLE schedule_entries CASCADE;
		TRUNCATE TABLE users CASCADE;
	`)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to clean up test database: %v", err))
	}

	return testDB
}

func InsertTestData(t interface{ Fatal(args ...interface{}) }, db *sql.DB, entry models.ScheduleEntry) int {
	// First, create or get the user
	var userID int
	err := db.QueryRow(`
		WITH new_user AS (
			INSERT INTO users (email, name, is_admin)
			VALUES ($1, '', false)
			ON CONFLICT (email) DO NOTHING
			RETURNING id
		)
		SELECT id FROM new_user
		UNION ALL
		SELECT id FROM users WHERE email = $1
		LIMIT 1
	`, entry.UserEmail).Scan(&userID)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to create/get user: %v", err))
	}

	// Then insert the schedule entry
	var id int
	err = db.QueryRow(`
		INSERT INTO schedule_entries (title, description, start_time, end_time, user_id, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime, userID, entry.Recurrence).Scan(&id)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to insert test data: %v", err))
	}
	return id
}

// CleanupTestDB cleans up the test database
func (tdb *TestDB) CleanupTestDB(t *testing.T) {
	// Clean up the test database in the correct order
	_, err := tdb.DB.Exec(`
		TRUNCATE TABLE video_uploads CASCADE;
		TRUNCATE TABLE schedule_entries CASCADE;
		TRUNCATE TABLE users CASCADE;
	`)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to clean up test database: %v", err))
	}
}

// InsertTestData inserts a schedule entry into the test database
func (tdb *TestDB) InsertTestData(t *testing.T, entry models.ScheduleEntry) int {
	// First, create or get the user
	var userID int
	err := tdb.DB.QueryRow(`
		WITH new_user AS (
			INSERT INTO users (email, name, is_admin)
			VALUES ($1, '', false)
			ON CONFLICT (email) DO NOTHING
			RETURNING id
		)
		SELECT id FROM new_user
		UNION ALL
		SELECT id FROM users WHERE email = $1
		LIMIT 1
	`, entry.UserEmail).Scan(&userID)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to create/get user: %v", err))
	}

	// Then insert the schedule entry
	var id int
	err = tdb.DB.QueryRow(`
		INSERT INTO schedule_entries (title, description, start_time, end_time, user_id, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime, userID, entry.Recurrence).Scan(&id)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to insert test data: %v", err))
	}
	return id
}
