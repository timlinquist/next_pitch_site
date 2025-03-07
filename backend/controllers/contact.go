package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/models"
)

type ContactController struct{}

func NewContactController() *ContactController {
	return &ContactController{}
}

func (c *ContactController) SendEmail(ctx *gin.Context) {
	var form models.ContactForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.SendEmail(form); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}
