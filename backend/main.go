package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"nextpitch.com/backend/controllers"
	"nextpitch.com/backend/db"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Initialize database connection
	if err := db.InitDB(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.DB.Close()

	r := gin.Default()

	// CORS middleware
	r.Use(cors.Default())

	// Initialize controllers
	scheduleController := controllers.NewScheduleController()
	contactController := controllers.NewContactController()

	videoController, err := controllers.NewVideoController()
	if err != nil {
		fmt.Printf("Error initializing video controller: %v\n", err)
		os.Exit(1)
	}

	// API routes
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Schedule entries routes
	r.GET("/api/schedule", scheduleController.GetScheduleEntries)
	r.GET("/api/appointments/upcoming", scheduleController.GetUpcomingAppointmentsByEmail)
	r.POST("/api/schedule", scheduleController.CreateScheduleEntry)

	r.PUT("/api/schedule/:id", scheduleController.UpdateScheduleEntry)
	r.DELETE("/api/schedule/:id", scheduleController.DeleteScheduleEntry)

	// Contact form submission
	r.POST("/api/contact", contactController.SendEmail)

	// Video upload route
	r.POST("/api/video/upload", videoController.UploadVideo)

	// Serve static files from frontend directory
	r.Static("/static", "../frontend")

	// Serve index.html for all routes (React Router will handle the routing)
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join("..", "frontend", "index.html"))
	})

	r.Run(":8080") // Run server on port 8080
}
