package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/models"
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

	email := userEmail.(string)
	log.Printf("[User] Getting user for email: %v", email)

	user, err := c.userService.GetUserByEmail(email)
	if err != nil && err.Error() == "user not found" {
		user = &models.User{Email: email, Name: "", IsAdmin: false}
		if createErr := c.userService.CreateUser(user); createErr != nil {
			log.Printf("[User] Error creating user: %v", createErr)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		log.Printf("[User] Created new user for email: %s", email)
	} else if err != nil {
		log.Printf("[User] Error fetching user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
