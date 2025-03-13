package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Parse command line flags
	action := flag.String("action", "", "Migration action (up/down)")
	flag.Parse()

	if *action == "" {
		log.Fatal("Please specify an action (up/down)")
	}

	// Get database connection details from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	// Construct database URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)

	// Create migration instance
	m, err := migrate.New(
		"file://db/migrations",
		dbURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	// Run migration
	switch *action {
	case "up":
		if err := m.Up(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Migrations reverted successfully")
	default:
		log.Fatal("Invalid action. Use 'up' or 'down'")
	}
}
