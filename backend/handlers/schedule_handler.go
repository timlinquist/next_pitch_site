package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
)

type ScheduleHandler struct {
	scheduleService *services.ScheduleService
	userService     *services.UserService
}

func NewScheduleHandler(scheduleService *services.ScheduleService, userService *services.UserService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
		userService:     userService,
	}
}

func (h *ScheduleHandler) GetScheduleEntries(c *gin.Context) {
	entries, err := h.scheduleService.GetScheduleEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *ScheduleHandler) CreateScheduleEntry(c *gin.Context) {
	var entry models.ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user email from context
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Check if user is admin
	isAdmin, err := h.userService.IsAdmin(userEmail.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.scheduleService.CreateScheduleEntry(&entry, userEmail.(string), isAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (h *ScheduleHandler) UpdateScheduleEntry(c *gin.Context) {
	var entry models.ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user email from context
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Check if user is admin
	isAdmin, err := h.userService.IsAdmin(userEmail.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.scheduleService.UpdateScheduleEntry(&entry, userEmail.(string), isAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func (h *ScheduleHandler) DeleteScheduleEntry(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	entryID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	// Get user email from context
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	if err := h.scheduleService.DeleteScheduleEntry(entryID, userEmail.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
