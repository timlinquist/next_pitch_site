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

type AgeGroupWithSpots struct {
	models.CampAgeGroup
	RegisteredCount int `json:"registered_count"`
	SpotsRemaining  int `json:"spots_remaining"`
}

type CampWithSpots struct {
	models.Camp
	RegisteredCount int                `json:"registered_count"`
	SpotsRemaining  *int              `json:"spots_remaining"`
	AgeGroups       []AgeGroupWithSpots `json:"age_groups,omitempty"`
}

func (ctrl *CampController) buildCampWithSpots(camp models.Camp) CampWithSpots {
	count, _ := ctrl.campService.GetCampRegistrationCount(camp.ID)
	cws := CampWithSpots{Camp: camp, RegisteredCount: count}

	ageGroups, _ := ctrl.campService.GetAgeGroupsByCampID(camp.ID)
	if len(ageGroups) > 0 {
		for _, g := range ageGroups {
			agCount, _ := ctrl.campService.GetAgeGroupRegistrationCount(camp.ID, g.MinAge, g.MaxAge)
			cws.AgeGroups = append(cws.AgeGroups, AgeGroupWithSpots{
				CampAgeGroup:    g,
				RegisteredCount: agCount,
				SpotsRemaining:  g.MaxCapacity - agCount,
			})
		}
	} else if camp.MaxCapacity != nil {
		remaining := *camp.MaxCapacity - count
		cws.SpotsRemaining = &remaining
	}

	return cws
}

func (ctrl *CampController) GetActiveCamps(c *gin.Context) {
	camps, err := ctrl.campService.GetActiveCamps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch camps"})
		return
	}

	var result []CampWithSpots
	for _, camp := range camps {
		result = append(result, ctrl.buildCampWithSpots(camp))
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

	c.JSON(http.StatusOK, ctrl.buildCampWithSpots(*camp))
}

func (ctrl *CampController) GetCampBySlug(c *gin.Context) {
	slug := c.Param("slug")

	camp, err := ctrl.campService.GetCampBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "camp not found"})
		return
	}

	c.JSON(http.StatusOK, ctrl.buildCampWithSpots(*camp))
}

type createCampRequest struct {
	models.Camp
	AgeGroups []models.CampAgeGroup `json:"age_groups"`
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

	var req createCampRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	if len(req.AgeGroups) > 0 && req.Camp.MaxCapacity != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot set both max_capacity and age_groups"})
		return
	}

	if len(req.AgeGroups) > 0 {
		if err := ctrl.campService.ValidateAgeGroups(req.AgeGroups); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.Camp.MaxCapacity = nil
	}

	if err := ctrl.campService.CreateCamp(&req.Camp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create camp: " + err.Error()})
		return
	}

	if len(req.AgeGroups) > 0 {
		if err := ctrl.campService.SetAgeGroups(req.Camp.ID, req.AgeGroups); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set age groups: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, ctrl.buildCampWithSpots(req.Camp))
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

	var req createCampRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	req.Camp.ID = id

	if len(req.AgeGroups) > 0 && req.Camp.MaxCapacity != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot set both max_capacity and age_groups"})
		return
	}

	if len(req.AgeGroups) > 0 {
		if err := ctrl.campService.ValidateAgeGroups(req.AgeGroups); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.Camp.MaxCapacity = nil
	}

	if err := ctrl.campService.UpdateCamp(&req.Camp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update camp: " + err.Error()})
		return
	}

	if err := ctrl.campService.SetAgeGroups(req.Camp.ID, req.AgeGroups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set age groups: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ctrl.buildCampWithSpots(req.Camp))
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
