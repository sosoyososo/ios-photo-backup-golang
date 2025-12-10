package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// StatusHandler handles status check requests
func StatusHandler(tokenService *service.TokenService, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from context (set by AuthMiddleware)
		token, exists := c.Get("token")
		if !exists {
			appLogger.Info("Status check without token")
			errors.Unauthorized(c, "Token not found")
			return
		}

		tokenString, ok := token.(string)
		if !ok {
			appLogger.Info("Status check with invalid token format")
			errors.Unauthorized(c, "Invalid token format")
			return
		}

		// Check status
		resp, err := tokenService.Status(tokenString)
		if err != nil {
			appLogger.Warn("Status check failed", logger.String("error", err.Error()))
			errors.Unauthorized(c, err.Error())
			return
		}

		// Return success
		c.JSON(http.StatusOK, resp)
	}
}
