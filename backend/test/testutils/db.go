package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
)

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

func CleanupTestDB(t *testing.T, db *sql.DB) {
	// Clean up the test database in the correct order
	_, err := db.Exec(`
		TRUNCATE TABLE video_uploads CASCADE;
		TRUNCATE TABLE schedule_entries CASCADE;
		TRUNCATE TABLE users CASCADE;
	`)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to clean up test database: %v", err))
	}
}
