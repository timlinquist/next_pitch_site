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
	"nextpitch.com/backend/handlers"
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

	// Set timeouts for large file uploads
	r.Use(func(c *gin.Context) {
		// Set a 30-minute timeout for the entire request
		c.Request.Header.Set("Connection", "keep-alive")
		c.Set("timeout", "30m")
		c.Next()
	})

	// CORS middleware
	r.Use(cors.Default())

	// Initialize services
	userService := services.NewUserService(db.DB)
	scheduleService := services.NewScheduleService(db.DB)
	emailService := services.NewEmailService()

	// Initialize handlers and controllers
	scheduleHandler := handlers.NewScheduleHandler(scheduleService, userService, emailService)
	contactController := controllers.NewContactController(emailService)
	userController := controllers.NewUserController(userService)

	videoController, err := controllers.NewVideoController(db.DB, userService, emailService)
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
		protected.GET("/schedule", scheduleHandler.GetScheduleEntries)
		protected.GET("/appointments/upcoming", scheduleHandler.GetUpcomingAppointmentsByEmail)
		protected.POST("/schedule", scheduleHandler.CreateScheduleEntry)
		protected.PUT("/schedule/:id", scheduleHandler.UpdateScheduleEntry)
		protected.DELETE("/schedule/:id", scheduleHandler.DeleteScheduleEntry)

		// Video upload route
		protected.POST("/video/upload", videoController.UploadVideo)
	}

	// Public routes
	r.POST("/api/contact", contactController.SendEmail)

	// Serve static files from frontend build directory, with SPA fallback
	distPath := filepath.Join("..", "frontend", "dist")
	r.Use(func(c *gin.Context) {
		// Skip API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.Next()
			return
		}
		// Try to serve a static file from dist
		filePath := filepath.Join(distPath, c.Request.URL.Path)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			c.File(filePath)
			c.Abort()
			return
		}
		c.Next()
	})

	// SPA fallback: serve index.html for all unmatched routes
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(distPath, "index.html"))
	})

	// Use PORT env var (Render sets this automatically), fallback to 8080 for local dev
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
