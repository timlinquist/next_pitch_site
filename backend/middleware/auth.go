package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

type CustomClaims struct {
	Email string `json:"email"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func AuthMiddleware() gin.HandlerFunc {
	issuerURL := fmt.Sprintf("https://%s/", os.Getenv("AUTH0_DOMAIN"))
	log.Printf("[Auth] Starting middleware with issuer URL: %s", issuerURL)

	// Parse the Auth0 domain URL for the JWKS provider
	auth0Domain := "dev-cx0z71mw7lq3og41.us.auth0.com"
	issuerURLParsed, err := url.Parse(fmt.Sprintf("https://%s/", auth0Domain))
	if err != nil {
		panic(err)
	}

	provider := jwks.NewCachingProvider(issuerURLParsed, 5)

	// Create a validator using the Auth0 domain as issuer and API identifier as audience
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		fmt.Sprintf("https://%s/", auth0Domain),
		[]string{"https://thenextpitch.org"},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
	)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		log.Println("[Auth] Middleware executed")

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("[Auth] No authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization header"})
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Println("[Auth] Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		log.Printf("[Auth] Attempting to validate token: %s...", token)
		claims, err := jwtValidator.ValidateToken(c.Request.Context(), token)
		if err != nil {
			log.Printf("[Auth] Token validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		// Extract email from claims and set in context
		if customClaims, ok := claims.(*validator.ValidatedClaims); ok {
			log.Printf("[Auth] Full claims: %+v", customClaims)
			log.Printf("[Auth] Claims debug - Subject: %s", customClaims.RegisteredClaims.Subject)
			log.Printf("[Auth] Claims debug - Issuer: %s", customClaims.RegisteredClaims.Issuer)
			log.Printf("[Auth] Claims debug - Audience: %v", customClaims.RegisteredClaims.Audience)

			// Try getting email from different possible locations
			var email string

			// Try custom claims first
			if cc, ok := customClaims.CustomClaims.(*CustomClaims); ok && cc.Email != "" {
				email = cc.Email
				log.Printf("[Auth] Found email in custom claims: %s", email)
			}

			// If still empty, we need to configure Auth0 to include email in the token
			if email == "" {
				log.Printf("[Auth] No email found in token claims. Please configure Auth0 to include email in the token.")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "No email found in token"})
				c.Abort()
				return
			}

			log.Printf("[Auth] Successfully validated token for email: %s", email)
			c.Set("user_email", email)
		} else {
			log.Printf("[Auth] Failed to cast claims to ValidatedClaims")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process token claims"})
			c.Abort()
			return
		}
		c.Next()
	}
}
