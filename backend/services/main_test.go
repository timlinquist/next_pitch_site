package services

import (
	"time"

	_ "github.com/lib/pq" // postgres driver
)

// This file serves as a central location for common test setup and imports.
// The postgres driver is imported here since it's needed by multiple test files.

// Helper function to create a fixed timestamp for testing
func fixedTime() time.Time {
	return time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
}
