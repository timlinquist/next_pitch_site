package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/models"
	"nextpitch.com/backend/services"
)

type CampController struct {
	campService *services.CampService
	userService *services.UserService
}

func NewCampController(campService *services.CampService, userService *services.UserService) *CampController {
	return &CampController{
		campService: campService,
		userService: userService,
	}
}

func (ctrl *CampController) GetActiveCamps(c *gin.Context) {
	camps, err := ctrl.campService.GetActiveCamps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch camps"})
		return
	}

	// Include registration count for each camp
	type CampWithSpots struct {
		models.Camp
		RegisteredCount int  `json:"registered_count"`
		SpotsRemaining  *int `json:"spots_remaining"`
	}

	var result []CampWithSpots
	for _, camp := range camps {
		count, err := ctrl.campService.GetCampRegistrationCount(camp.ID)
		if err != nil {
			count = 0
		}
		cws := CampWithSpots{Camp: camp, RegisteredCount: count}
		if camp.MaxCapacity != nil {
			remaining := *camp.MaxCapacity - count
			cws.SpotsRemaining = &remaining
		}
		result = append(result, cws)
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CampController) GetCampByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid camp ID"})
		return
	}

	camp, err := ctrl.campService.GetCampByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "camp not found"})
		return
	}

	count, _ := ctrl.campService.GetCampRegistrationCount(camp.ID)

	type CampWithSpots struct {
		models.Camp
		RegisteredCount int  `json:"registered_count"`
		SpotsRemaining  *int `json:"spots_remaining"`
	}

	result := CampWithSpots{Camp: *camp, RegisteredCount: count}
	if camp.MaxCapacity != nil {
		remaining := *camp.MaxCapacity - count
		result.SpotsRemaining = &remaining
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *CampController) CreateCamp(c *gin.Context) {
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

	var camp models.Camp
	if err := c.ShouldBindJSON(&camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	if err := ctrl.campService.CreateCamp(&camp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create camp: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, camp)
}

func (ctrl *CampController) UpdateCamp(c *gin.Context) {
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

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid camp ID"})
		return
	}

	var camp models.Camp
	if err := c.ShouldBindJSON(&camp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	camp.ID = id

	if err := ctrl.campService.UpdateCamp(&camp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update camp: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, camp)
}

func (ctrl *CampController) DeactivateCamp(c *gin.Context) {
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

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid camp ID"})
		return
	}

	if err := ctrl.campService.DeactivateCamp(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate camp: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
