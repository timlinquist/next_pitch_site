package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(cors.Default())

	// Setup routes
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	return r
}

func TestPingEndpoint(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		endpoint   string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "ping endpoint returns pong",
			endpoint:   "/api/ping",
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"pong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.endpoint, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}
