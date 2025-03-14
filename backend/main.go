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
	"nextpitch.com/backend/middleware"
	"nextpitch.com/backend/services"
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

	// Initialize services
	userService := services.NewUserService(db.DB)

	// Initialize controllers
	scheduleController := controllers.NewScheduleController()
	contactController := controllers.NewContactController()
	userController := controllers.NewUserController(userService)

	videoController, err := controllers.NewVideoController()
	if err != nil {
		fmt.Printf("Error initializing video controller: %v\n", err)
		os.Exit(1)
	}

	// API routes
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Protected routes with Auth0 middleware
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// User routes
		protected.GET("/users/me", userController.GetCurrentUser)

		// Schedule entries routes
		protected.GET("/schedule", scheduleController.GetScheduleEntries)
		protected.GET("/appointments/upcoming", scheduleController.GetUpcomingAppointmentsByEmail)
		protected.POST("/schedule", scheduleController.CreateScheduleEntry)
		protected.PUT("/schedule/:id", scheduleController.UpdateScheduleEntry)
		protected.DELETE("/schedule/:id", scheduleController.DeleteScheduleEntry)

		// Video upload route
		protected.POST("/video/upload", videoController.UploadVideo)
	}

	// Public routes
	r.POST("/api/contact", contactController.SendEmail)

	// Serve static files from frontend directory
	r.Static("/static", "../frontend")

	// Serve index.html for all routes (React Router will handle the routing)
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join("..", "frontend", "index.html"))
	})

	r.Run(":8080") // Run server on port 8080
}
