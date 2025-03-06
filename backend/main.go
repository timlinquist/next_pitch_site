package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gopkg.in/gomail.v2"
)

var db *sql.DB

type ContactForm struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type ScheduleEntry struct {
	ID          int       `json:"id"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func initDB() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	fmt.Println("Successfully connected to database")
	return nil
}

func sendEmail(form ContactForm) error {
	m := gomail.NewMessage()
	m.SetHeader("From", form.Email)
	m.SetHeader("To", "info@nextpitch.com")
	m.SetHeader("Subject", fmt.Sprintf("Contact Form: %s", form.Subject))

	body := fmt.Sprintf(`
New Contact Form Submission

Name: %s
Email: %s
Subject: %s

Message:
%s
	`, form.Name, form.Email, form.Subject, form.Message)

	m.SetBody("text/plain", body)

	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		return fmt.Errorf("SMTP_PASSWORD environment variable is not set")
	}

	d := gomail.NewDialer("smtp.gmail.com", 587, "your-email@gmail.com", smtpPassword)
	// Note: For Gmail, you'll need to use an App Password instead of your regular password
	// You can generate one in your Google Account settings under Security > 2-Step Verification > App passwords

	return d.DialAndSend(m)
}

func getScheduleEntries(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, title, description, start_time, end_time, created_at, updated_at 
		FROM schedule_entries 
		ORDER BY start_time ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedule entries"})
		return
	}
	defer rows.Close()

	var entries []ScheduleEntry
	for rows.Next() {
		var entry ScheduleEntry
		err := rows.Scan(&entry.ID, &entry.Title, &entry.Description, &entry.StartTime, &entry.EndTime, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan schedule entry"})
			return
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}

func createScheduleEntry(c *gin.Context) {
	var entry ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.QueryRow(`
		INSERT INTO schedule_entries (title, description, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime).
		Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule entry"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func updateScheduleEntry(c *gin.Context) {
	id := c.Param("id")
	var entry ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.QueryRow(`
		UPDATE schedule_entries 
		SET title = $1, description = $2, start_time = $3, end_time = $4
		WHERE id = $5
		RETURNING id, created_at, updated_at
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime, id).
		Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Schedule entry not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schedule entry"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func deleteScheduleEntry(c *gin.Context) {
	id := c.Param("id")
	result, err := db.Exec("DELETE FROM schedule_entries WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule entry"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule entry deleted successfully"})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Initialize database connection
	if err := initDB(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	r := gin.Default()

	// CORS middleware
	r.Use(cors.Default())

	// API routes
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Schedule entries routes
	r.GET("/api/schedule", getScheduleEntries)
	r.POST("/api/schedule", createScheduleEntry)
	r.PUT("/api/schedule/:id", updateScheduleEntry)
	r.DELETE("/api/schedule/:id", deleteScheduleEntry)

	// Contact form submission
	r.POST("/api/contact", func(c *gin.Context) {
		var form ContactForm
		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := sendEmail(form); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
	})

	// Serve static files from frontend directory
	r.Static("/static", "../frontend")

	// Serve index.html for all routes (React Router will handle the routing)
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join("..", "frontend", "index.html"))
	})

	r.Run(":8080") // Run server on port 8080
}
