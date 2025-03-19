package helpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"nextpitch.com/backend/models"
)

type ScheduleFixtures struct {
	TestEvent     models.ScheduleEntry `json:"test_event"`
	NewEvent      models.ScheduleEntry `json:"new_event"`
	OriginalEvent models.ScheduleEntry `json:"original_event"`
	EventToDelete models.ScheduleEntry `json:"event_to_delete"`
}

type TestDB struct {
	DB       *sql.DB
	Fixtures *ScheduleFixtures
}

// LoadFixtures loads test data from JSON files
func LoadFixtures(t *testing.T) *ScheduleFixtures {
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

// runMigrations runs database migrations
func runMigrations(t *testing.T, dbname string) {
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

// SetupTestDB creates a test database and loads fixtures
func SetupTestDB(t *testing.T) *TestDB {
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

	// Run migrations
	runMigrations(t, dbname)

	// Load fixtures
	fixtures := LoadFixtures(t)

	return &TestDB{
		DB:       testDB,
		Fixtures: fixtures,
	}
}

// CleanupTestDB cleans up the test database
func (tdb *TestDB) CleanupTestDB(t *testing.T) {
	// Clean up the test database
	_, err := tdb.DB.Exec("TRUNCATE TABLE schedule_entries, users CASCADE")
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to clean up test database: %v", err))
	}
}

// Close closes the database connection
func (tdb *TestDB) Close() error {
	return tdb.DB.Close()
}

// InsertTestData inserts a schedule entry into the test database
func (tdb *TestDB) InsertTestData(t *testing.T, entry models.ScheduleEntry) int {
	// First, create or get the user
	var userID int
	err := tdb.DB.QueryRow(`
		WITH new_user AS (
			INSERT INTO users (email, first_name, last_name)
			VALUES ($1, '', '')
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
		INSERT INTO schedule_entries (title, description, start_time, end_time, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime, userID).Scan(&id)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to insert test data: %v", err))
	}
	return id
}
