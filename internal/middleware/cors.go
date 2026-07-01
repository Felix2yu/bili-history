package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a CORS middleware that allows all origins.
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Allow all origins
			return true
		},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Set-Cookie"},
		MaxAge:           12 * time.Hour,
		AllowPrivateNetwork: true,
	})
}

// isAllowedOrigin checks if the origin is allowed.
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return true
	}
	// Allow localhost and common development origins
	allowed := []string{
		"http://localhost",
		"http://127.0.0.1",
		"https://localhost",
		"https://127.0.0.1",
		"https://b.yufei.im",
		"http://b.yufei.im",
	}
	for _, a := range allowed {
		if strings.HasPrefix(origin, a) {
			return true
		}
	}
	return true // Allow all for now
}
