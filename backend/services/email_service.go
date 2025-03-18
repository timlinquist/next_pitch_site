package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"gopkg.in/gomail.v2"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/types"
)

type EmailService struct {
	smtpUsername string
	smtpPassword string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
	}
}

func (s *EmailService) SendCustomEmail(config types.EmailConfig) error {
	if s.smtpUsername == "" {
		return fmt.Errorf("SMTP_USERNAME environment variable is not set")
	}
	if s.smtpPassword == "" {
		return fmt.Errorf("SMTP_PASSWORD environment variable is not set")
	}

	log.Printf("[Email] Attempting to send email from %s to %s", config.From, config.To)

	m := gomail.NewMessage()
	m.SetHeader("From", config.From)
	m.SetHeader("To", config.To)
	m.SetHeader("Subject", config.Subject)
	m.SetBody("text/plain", config.Body)

	d := gomail.NewDialer("smtp.zoho.com", 465, s.smtpUsername, s.smtpPassword)
	d.TLSConfig = &tls.Config{ServerName: "smtp.zoho.com"}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("[Email] Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("[Email] Successfully sent email from %s to %s", config.From, config.To)
	return nil
}

func (s *EmailService) SendContactEmail(form models.ContactForm) error {
	config := types.EmailConfig{
		From:    form.Email,
		To:      "coachtim@thenextpitch.org",
		Subject: fmt.Sprintf("Contact Form: %s", form.Subject),
		Body: fmt.Sprintf(`
New Contact Form Submission

Name: %s
Email: %s
Subject: %s

Message:
%s
		`, form.Name, form.Email, form.Subject, form.Message),
	}

	return s.SendCustomEmail(config)
}

func (s *EmailService) SendAppointmentCancellationEmail(entry *models.ScheduleEntry) error {
	log.Printf("[Email] Preparing cancellation email for appointment %d", entry.ID)

	config := types.EmailConfig{
		From:    "coachtim@thenextpitch.org",
		To:      entry.UserEmail,
		Subject: "Appointment Cancellation - The Next Pitch",
		Body: fmt.Sprintf(`
Dear %s,

Your appointment scheduled for %s has been cancelled.

Appointment Details:
Title: %s
Date: %s
Time: %s - %s

If you would like to schedule a new appointment, please visit our website.

Best regards,
Coach Tim
		`, entry.UserEmail, entry.StartTime.Format("January 2, 2006"), entry.Title,
			entry.StartTime.Format("January 2, 2006"),
			entry.StartTime.Format("3:04 PM"),
			entry.EndTime.Format("3:04 PM")),
	}

	return s.SendCustomEmail(config)
}
