package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"nextpitch.com/backend/models"
)

const (
	AdminEmail = "coachtim@thenextpitch.org"
)

// EmailServiceInterface defines the interface for email service
type EmailServiceInterface interface {
	SendVideoUploadNotification(user *models.User, fileName string) error
	SendContactEmail(form models.ContactForm) error
	SendAppointmentCancellationEmail(entry *models.ScheduleEntry) error
	SendAppointmentConfirmationEmail(entry *models.ScheduleEntry) error
	SendCampRegistrationConfirmation(reg *models.CampRegistration, athlete *models.Athlete, camp *models.Camp)
	SendAdminCampRegistrationNotification(reg *models.CampRegistration, athlete *models.Athlete, camp *models.Camp)
	QueueEmail(data EmailData)
}

type EmailType int

func (e EmailType) String() string {
	return [...]string{
		"EmailTypeCancellation",
		"EmailTypeConfirmation",
		"EmailTypeAdminCancellation",
		"EmailTypeAdminConfirmation",
		"EmailTypeContact",
		"EmailTypeVideoUpload",
		"EmailTypeCampRegistration",
		"EmailTypeAdminCampRegistration",
	}[e]
}

const (
	EmailTypeCancellation EmailType = iota
	EmailTypeConfirmation
	EmailTypeAdminCancellation
	EmailTypeAdminConfirmation
	EmailTypeContact
	EmailTypeVideoUpload
	EmailTypeCampRegistration
	EmailTypeAdminCampRegistration
)

type EmailData struct {
	Type     EmailType
	Data     interface{}
	To       string
	Subject  string
	Template string
}

type VideoUploadData struct {
	User       *models.User
	FileName   string
	UploadTime time.Time
}

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	templates    map[string]*template.Template
	emailChan    chan EmailData
}

func NewEmailService() *EmailService {
	s := &EmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    AdminEmail,
		templates:    make(map[string]*template.Template),
		emailChan:    make(chan EmailData, 100), // Buffer size of 100
	}

	// Log SMTP configuration (without sensitive data)
	log.Printf("[Email] Initializing email service with host: %s, port: %s, from: %s",
		s.smtpHost, s.smtpPort, s.fromEmail)

	// Load templates
	templateFiles := map[string]string{
		"cancellation":              "cancellation.html",
		"confirmation":              "confirmation.html",
		"admin_cancellation":        "admin_cancellation.html",
		"admin_confirmation":        "admin_confirmation.html",
		"contact":                   "contact.html",
		"video_upload":              "video_upload.html",
		"camp_registration":         "camp_registration_confirmation.html",
		"admin_camp_registration":   "admin_camp_registration.html",
	}

	// Set template directory path
	templateDir := filepath.Join("..", "backend", "templates", "email")
	log.Printf("[Email] Loading templates from: %s", templateDir)

	// List contents of template directory
	files, err := os.ReadDir(templateDir)
	if err != nil {
		log.Printf("[Email] ERROR: Failed to read template directory: %v", err)
		return s
	}

	log.Printf("[Email] Found files in template directory:")
	for _, file := range files {
		log.Printf("[Email] - %s", file.Name())
	}

	for name, file := range templateFiles {
		templatePath := filepath.Join(templateDir, file)
		log.Printf("[Email] Loading template %s from %s", name, templatePath)

		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			log.Printf("[Email] ERROR: Failed to parse template %s: %v", name, err)
			continue
		}
		s.templates[name] = tmpl
		log.Printf("[Email] Successfully loaded template %s", name)
	}

	// Log loaded templates summary
	log.Printf("[Email] Loaded %d templates: %v", len(s.templates), getMapKeys(s.templates))

	// Start the email processing worker
	log.Printf("[Email] Starting email processing worker")
	go s.processEmails()

	return s
}

func getMapKeys(m map[string]*template.Template) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (s *EmailService) processEmails() {
	log.Printf("[Email] Email processing worker started")
	for data := range s.emailChan {
		log.Printf("[Email] Processing email - Type: %s, Template: %s, To: %s, Subject: %s",
			data.Type.String(), data.Template, data.To, data.Subject)

		// Check if required SMTP settings are configured
		if s.smtpHost == "" || s.smtpPort == "" || s.smtpUsername == "" || s.smtpPassword == "" {
			log.Printf("[Email] ERROR: SMTP settings not properly configured - Host: %v, Port: %v, Username: %v, Password: %v",
				s.smtpHost != "", s.smtpPort != "", s.smtpUsername != "", s.smtpPassword != "")
			continue
		}

		// Verify template exists
		if _, ok := s.templates[data.Template]; !ok {
			log.Printf("[Email] ERROR: Template '%s' not found in available templates: %v",
				data.Template, getMapKeys(s.templates))
			continue
		} else {
			log.Printf("[Email] Found template '%s' for processing", data.Template)
		}

		if err := s.sendEmail(data); err != nil {
			log.Printf("[Email] ERROR: Failed to send email - Type: %s, Template: %s, To: %s, Error: %v",
				data.Type.String(), data.Template, data.To, err)
		} else {
			log.Printf("[Email] Successfully processed and sent email - Type: %s, Template: %s, To: %s",
				data.Type.String(), data.Template, data.To)
		}
	}
}

func (s *EmailService) QueueEmail(data EmailData) {
	log.Printf("[Email] Queueing email - Type: %s, Template: %s, To: %s, Subject: %s",
		data.Type.String(), data.Template, data.To, data.Subject)
	s.emailChan <- data
	log.Printf("[Email] Successfully queued email - Type: %s", data.Type.String())
}

func (s *EmailService) SendContactEmail(form models.ContactForm) error {
	log.Printf("[Email] Preparing to send contact form email from %s", form.Email)

	s.QueueEmail(EmailData{
		Type:     EmailTypeContact,
		Data:     form,
		To:       AdminEmail,
		Subject:  fmt.Sprintf("Contact Form: %s", form.Subject),
		Template: "contact",
	})

	log.Printf("[Email] Successfully queued contact form email from %s", form.Email)
	return nil
}

func (s *EmailService) SendAppointmentCancellationEmail(entry *models.ScheduleEntry) error {
	log.Printf("[Email] Preparing to send cancellation email for appointment %d", entry.ID)

	// Queue user email
	s.QueueEmail(EmailData{
		Type:     EmailTypeCancellation,
		Data:     entry,
		To:       entry.UserEmail,
		Subject:  "Appointment Cancellation Confirmation",
		Template: "cancellation",
	})

	// Queue admin email
	s.QueueEmail(EmailData{
		Type:     EmailTypeAdminCancellation,
		Data:     entry,
		To:       AdminEmail,
		Subject:  "Appointment Cancellation Notification",
		Template: "admin_cancellation",
	})

	log.Printf("[Email] Successfully queued cancellation emails for appointment %d", entry.ID)
	return nil
}

func (s *EmailService) SendAppointmentConfirmationEmail(entry *models.ScheduleEntry) error {
	log.Printf("[Email] Preparing to send confirmation email for appointment %d", entry.ID)

	// Queue user email
	s.QueueEmail(EmailData{
		Type:     EmailTypeConfirmation,
		Data:     entry,
		To:       entry.UserEmail,
		Subject:  "Appointment Confirmation",
		Template: "confirmation",
	})

	// Queue admin email
	s.QueueEmail(EmailData{
		Type:     EmailTypeAdminConfirmation,
		Data:     entry,
		To:       AdminEmail,
		Subject:  "New Appointment Scheduled",
		Template: "admin_confirmation",
	})

	log.Printf("[Email] Successfully queued confirmation emails for appointment %d", entry.ID)
	return nil
}

type CampRegistrationEmailData struct {
	Athlete  *models.Athlete
	Camp     *models.Camp
	Amount   string
	RegTime  time.Time
}

func (s *EmailService) SendCampRegistrationConfirmation(reg *models.CampRegistration, athlete *models.Athlete, camp *models.Camp) {
	log.Printf("[Email] Preparing to send camp registration confirmation for registration %d", reg.ID)

	data := CampRegistrationEmailData{
		Athlete: athlete,
		Camp:    camp,
		Amount:  fmt.Sprintf("$%.2f", float64(reg.AmountCents)/100),
		RegTime: time.Now(),
	}

	s.QueueEmail(EmailData{
		Type:     EmailTypeCampRegistration,
		Data:     data,
		To:       reg.ParentEmail,
		Subject:  fmt.Sprintf("Camp Registration Confirmation - %s", camp.Name),
		Template: "camp_registration",
	})
}

func (s *EmailService) SendAdminCampRegistrationNotification(reg *models.CampRegistration, athlete *models.Athlete, camp *models.Camp) {
	log.Printf("[Email] Preparing to send admin camp registration notification for registration %d", reg.ID)

	data := CampRegistrationEmailData{
		Athlete: athlete,
		Camp:    camp,
		Amount:  fmt.Sprintf("$%.2f", float64(reg.AmountCents)/100),
		RegTime: time.Now(),
	}

	s.QueueEmail(EmailData{
		Type:     EmailTypeAdminCampRegistration,
		Data:     data,
		To:       AdminEmail,
		Subject:  fmt.Sprintf("New Camp Registration - %s: %s", camp.Name, athlete.Name),
		Template: "admin_camp_registration",
	})
}

func (s *EmailService) SendVideoUploadNotification(user *models.User, fileName string) error {
	log.Printf("[Email] Preparing to send video upload notification for user %s", user.Email)

	data := VideoUploadData{
		User:       user,
		FileName:   fileName,
		UploadTime: time.Now(),
	}

	s.QueueEmail(EmailData{
		Type:     EmailTypeVideoUpload,
		Data:     data,
		To:       AdminEmail,
		Subject:  fmt.Sprintf("New Video Upload from %s", user.Name),
		Template: "video_upload",
	})

	log.Printf("[Email] Successfully queued video upload notification for user %s", user.Email)
	return nil
}

func (s *EmailService) sendEmail(data EmailData) error {
	// Log template lookup
	tmpl, ok := s.templates[data.Template]
	if !ok {
		return fmt.Errorf("template %s not found in available templates: %v",
			data.Template, getMapKeys(s.templates))
	}
	log.Printf("[Email] Found template %s, executing with data", data.Template)

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data.Data); err != nil {
		return fmt.Errorf("error executing template %s: %v", data.Template, err)
	}
	log.Printf("[Email] Successfully executed template %s", data.Template)

	// Format From header with name and email
	fromHeader := fmt.Sprintf("The Next Pitch <%s>", s.fromEmail)

	// Create email message
	message := []byte("To: " + data.To + "\r\n" +
		"From: " + fromHeader + "\r\n" +
		"Subject: " + data.Subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n" +
		buf.String())

	// Connect to SMTP server
	log.Printf("[Email] Connecting to SMTP server %s:%s", s.smtpHost, s.smtpPort)

	// Create SMTP client
	smtpClient, err := smtp.Dial(fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer smtpClient.Close()

	// Send EHLO
	if err = smtpClient.Hello("localhost"); err != nil {
		return fmt.Errorf("EHLO failed: %v", err)
	}

	// Start TLS
	if ok, _ := smtpClient.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName: s.smtpHost,
			MinVersion: tls.VersionTLS12,
		}
		if err = smtpClient.StartTLS(config); err != nil {
			return fmt.Errorf("StartTLS failed: %v", err)
		}
		log.Printf("[Email] TLS connection established")
	}

	// Authenticate
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	if err = smtpClient.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %v", err)
	}
	log.Printf("[Email] Authentication successful")

	// Set sender and recipient
	if err = smtpClient.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err = smtpClient.Rcpt(data.To); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send the email body
	writer, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("failed to create message writer: %v", err)
	}
	defer writer.Close()

	_, err = writer.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	// Close the connection and handle any errors
	if err = smtpClient.Quit(); err != nil {
		// Log the error but don't return it since the message was already sent
		log.Printf("[Email] Warning: Error closing SMTP connection: %v", err)
	}

	log.Printf("[Email] Successfully sent email to %s", data.To)
	return nil
}
