package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// TokenValidationMiddleware validates JWT tokens
func TokenValidationMiddleware(tokenService *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from context (set by AuthMiddleware)
		token, exists := c.Get("token")
		if !exists {
			errors.Unauthorized(c, "Token not found")
			c.Abort()
			return
		}

		tokenString, ok := token.(string)
		if !ok {
			errors.Unauthorized(c, "Invalid token format")
			c.Abort()
			return
		}

		// Validate token
		claims, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			errors.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Store validated claims in context
		c.Set("claims", claims)
		c.Next()
	}
}
