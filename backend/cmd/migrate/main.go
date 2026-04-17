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
	action := flag.String("action", "", "Migration action (up/down/force)")
	version := flag.Int("version", 0, "Version to force to (required for force action)")
	flag.Parse()

	if *action == "" {
		log.Fatal("Please specify an action (up/down/force)")
	}

	if *action == "force" && *version == 0 {
		log.Fatal("Please specify a version number when using force action")
	}

	// Prefer DATABASE_URL (provided by Render), fall back to individual env vars
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSL_MODE")

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
	}

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
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		fmt.Println("Migrations reverted successfully")
	case "force":
		if err := m.Force(*version); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Database version forced to %d\n", *version)
	default:
		log.Fatal("Invalid action. Use 'up', 'down', or 'force'")
	}
}
