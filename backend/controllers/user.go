package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/services"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) GetCurrentUser(ctx *gin.Context) {
	// Get user email from Auth0 token
	userEmail, exists := ctx.Get("user_email")
	if !exists {
		log.Printf("[User] No user_email in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	log.Printf("[User] Getting user for email: %v", userEmail)
	// Get user from service
	user, err := c.userService.GetUserByEmail(userEmail.(string))
	if err != nil {
		log.Printf("[User] Error fetching user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
