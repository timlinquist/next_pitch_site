package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
)

type ScheduleController struct {
	scheduleService *services.ScheduleService
	userService     *services.UserService
}

func NewScheduleController(scheduleService *services.ScheduleService, userService *services.UserService) *ScheduleController {
	return &ScheduleController{
		scheduleService: scheduleService,
		userService:     userService,
	}
}

func (sc *ScheduleController) GetScheduleEntries(c *gin.Context) {
	entries, err := sc.scheduleService.GetScheduleEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedule entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (sc *ScheduleController) CreateScheduleEntry(c *gin.Context) {
	var entry models.ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user email from context
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User email not found in context"})
		return
	}

	// Check if user is admin
	isAdmin, err := sc.userService.IsAdmin(userEmail.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user permissions"})
		return
	}

	err = sc.scheduleService.CreateScheduleEntry(&entry, userEmail.(string), isAdmin)
	if err != nil {
		switch err.Error() {
		case "event duration exceeds maximum allowed duration for non-admin users",
			"event overlaps with existing events":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (sc *ScheduleController) UpdateScheduleEntry(c *gin.Context) {
	id := c.Param("id")
	entryID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var entry models.ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.ID = int(entryID)

	// Get user email from context
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User email not found in context"})
		return
	}

	// Check if user is admin
	isAdmin, err := sc.userService.IsAdmin(userEmail.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user permissions"})
		return
	}

	err = sc.scheduleService.UpdateScheduleEntry(&entry, userEmail.(string), isAdmin)
	if err != nil {
		switch err.Error() {
		case "schedule entry not found or unauthorized":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "event duration exceeds maximum allowed duration for non-admin users",
			"event overlaps with existing events":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, entry)
}

func (sc *ScheduleController) DeleteScheduleEntry(c *gin.Context) {
	id := c.Param("id")
	entryID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Get user email from context
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User email not found in context"})
		return
	}

	err = sc.scheduleService.DeleteScheduleEntry(entryID, userEmail.(string))
	if err != nil {
		switch err.Error() {
		case "schedule entry not found or unauthorized":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule entry deleted successfully"})
}

func (sc *ScheduleController) GetUpcomingAppointmentsByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	entries, err := sc.scheduleService.GetUpcomingAppointmentsByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}

	c.JSON(http.StatusOK, entries)
}
