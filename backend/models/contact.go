package models

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

type ContactForm struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func SendEmail(form ContactForm) error {
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
