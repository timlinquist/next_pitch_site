package controllers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
)

type RegistrationController struct {
	registrationService *services.RegistrationService
	campService         *services.CampService
	emailService        *services.EmailService
	userService         *services.UserService
}

func NewRegistrationController(
	registrationService *services.RegistrationService,
	campService *services.CampService,
	emailService *services.EmailService,
	userService *services.UserService,
) *RegistrationController {
	return &RegistrationController{
		registrationService: registrationService,
		campService:         campService,
		emailService:        emailService,
		userService:         userService,
	}
}

type RegisterRequest struct {
	Athlete       models.Athlete `json:"athlete" binding:"required"`
	CampID        int            `json:"camp_id" binding:"required"`
	ParentEmail   string         `json:"parent_email" binding:"required"`
	ParentPhone   string         `json:"parent_phone"`
	PaymentMethod string         `json:"payment_method" binding:"required"`
}

func (ctrl *RegistrationController) RegisterForCamp(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	if req.PaymentMethod != "stripe" && req.PaymentMethod != "paypal" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_method must be 'stripe' or 'paypal'"})
		return
	}

	req.Athlete.ParentEmail = req.ParentEmail
	req.Athlete.ParentPhone = req.ParentPhone

	reg, err := ctrl.registrationService.CreateAthleteAndRegistration(
		&req.Athlete,
		req.CampID,
		models.PaymentMethod(req.PaymentMethod),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PaymentMethod == "stripe" {
		clientSecret, err := ctrl.registrationService.InitiateStripePayment(reg.ID)
		if err != nil {
			log.Printf("[Registration] Failed to create stripe payment intent: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate payment"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"registration_id": reg.ID,
			"client_secret":   clientSecret,
		})
		return
	}

	// PayPal flow
	orderID, err := ctrl.registrationService.CreatePayPalOrder(reg.ID)
	if err != nil {
		log.Printf("[Registration] Failed to create paypal order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate payment"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"registration_id": reg.ID,
		"paypal_order_id": orderID,
	})
}

type StripeConfirmRequest struct {
	RegistrationID int `json:"registration_id" binding:"required"`
}

func (ctrl *RegistrationController) ConfirmStripePayment(c *gin.Context) {
	var req StripeConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	if err := ctrl.registrationService.ConfirmStripePayment(req.RegistrationID); err != nil {
		log.Printf("[Registration] Failed to confirm stripe payment: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctrl.sendConfirmationEmails(req.RegistrationID)
	c.JSON(http.StatusOK, gin.H{"status": "paid"})
}

func (ctrl *RegistrationController) HandleStripeWebhook(c *gin.Context) {
	// Read raw body for signature verification - do NOT use ShouldBindJSON
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	reg, err := ctrl.registrationService.HandleStripeWebhook(payload, signature)
	if err != nil {
		log.Printf("[Registration] Webhook error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if reg != nil && reg.PaymentStatus == models.PaymentStatusPaid {
		ctrl.sendConfirmationEmails(reg.ID)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

type PayPalCaptureRequest struct {
	RegistrationID int    `json:"registration_id" binding:"required"`
	PaypalOrderID  string `json:"paypal_order_id" binding:"required"`
}

func (ctrl *RegistrationController) CapturePayPalPayment(c *gin.Context) {
	var req PayPalCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	if err := ctrl.registrationService.CapturePayPalOrder(req.RegistrationID, req.PaypalOrderID); err != nil {
		log.Printf("[Registration] Failed to capture paypal payment: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctrl.sendConfirmationEmails(req.RegistrationID)
	c.JSON(http.StatusOK, gin.H{"status": "paid"})
}

func (ctrl *RegistrationController) GetCampRegistrations(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	isAdmin, err := ctrl.userService.IsAdmin(userEmail.(string))
	if err != nil || !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	campID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid camp ID"})
		return
	}

	registrations, err := ctrl.registrationService.GetRegistrationsByCampID(campID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registrations"})
		return
	}

	c.JSON(http.StatusOK, registrations)
}

func (ctrl *RegistrationController) sendConfirmationEmails(registrationID int) {
	reg, err := ctrl.registrationService.GetRegistrationByID(registrationID)
	if err != nil {
		log.Printf("[Registration] Failed to get registration for email: %v", err)
		return
	}

	athlete, err := ctrl.registrationService.GetAthleteByID(reg.AthleteID)
	if err != nil {
		log.Printf("[Registration] Failed to get athlete for email: %v", err)
		return
	}

	camp, err := ctrl.campService.GetCampByID(reg.CampID)
	if err != nil {
		log.Printf("[Registration] Failed to get camp for email: %v", err)
		return
	}

	ctrl.emailService.SendCampRegistrationConfirmation(reg, athlete, camp)
	ctrl.emailService.SendAdminCampRegistrationNotification(reg, athlete, camp)
}

