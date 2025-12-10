package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// RefreshHandler handles token refresh requests
func RefreshHandler(tokenService *service.TokenService, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from context (set by AuthMiddleware)
		token, exists := c.Get("token")
		if !exists {
			appLogger.Info("Token refresh without token")
			errors.Unauthorized(c, "Token not found")
			return
		}

		tokenString, ok := token.(string)
		if !ok {
			appLogger.Info("Token refresh with invalid token format")
			errors.Unauthorized(c, "Invalid token format")
			return
		}

		appLogger.Info("Token refresh request")

		// Refresh token
		resp, err := tokenService.Refresh(tokenString)
		if err != nil {
			appLogger.Warn("Token refresh failed", logger.String("error", err.Error()))
			errors.Unauthorized(c, err.Error())
			return
		}

		appLogger.Info("Token refresh successful")

		// Return success
		c.JSON(http.StatusOK, gin.H{
			"token":      resp.Token,
			"expires_at": resp.ExpiresAt,
		})
	}
}
