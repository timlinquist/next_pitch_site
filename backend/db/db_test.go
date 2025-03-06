package db

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Save original environment variables
	originalHost := os.Getenv("DB_HOST")
	originalPort := os.Getenv("DB_PORT")
	originalUser := os.Getenv("DB_USER")
	originalPassword := os.Getenv("DB_PASSWORD")
	originalDBName := os.Getenv("DB_NAME")
	originalSSLMode := os.Getenv("DB_SSL_MODE")

	// Set test environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "nextpitch_test")
	os.Setenv("DB_SSL_MODE", "disable")

	// Restore original environment variables after test
	defer func() {
		os.Setenv("DB_HOST", originalHost)
		os.Setenv("DB_PORT", originalPort)
		os.Setenv("DB_USER", originalUser)
		os.Setenv("DB_PASSWORD", originalPassword)
		os.Setenv("DB_NAME", originalDBName)
		os.Setenv("DB_SSL_MODE", originalSSLMode)
	}()

	// Test database initialization
	err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test database connection
	if DB == nil {
		t.Fatal("Database connection is nil")
	}

	// Test ping
	err = DB.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}
