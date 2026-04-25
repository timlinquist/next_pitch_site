package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/services"
)

type LocationController struct {
	locationService *services.LocationService
	userService     *services.UserService
}

func NewLocationController(locationService *services.LocationService, userService *services.UserService) *LocationController {
	return &LocationController{
		locationService: locationService,
		userService:     userService,
	}
}

func (ctrl *LocationController) GetAll(c *gin.Context) {
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

	locs, err := ctrl.locationService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}
	c.JSON(http.StatusOK, locs)
}
