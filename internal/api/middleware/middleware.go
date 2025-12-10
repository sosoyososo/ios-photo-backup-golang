package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract token (remove "Bearer " prefix)
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Token is required",
			})
			c.Abort()
			return
		}

		// Store token in context for handlers to use
		c.Set("token", token)
		c.Next()
	}
}
