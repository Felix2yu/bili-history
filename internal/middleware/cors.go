package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a CORS middleware that allows all origins.
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		MaxAge:           12 * time.Hour,
	})
}
