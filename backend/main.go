package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type ContactForm struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
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

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	r := gin.Default()

	// CORS middleware
	r.Use(cors.Default())

	// API routes
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

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
