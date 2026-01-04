package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// RefreshHandler handles token refresh requests
func RefreshHandler(tokenService *service.TokenService, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from context (set by JWTMiddleware)
		token, exists := c.Get("token")
		if !exists {
			appLogger.Info("Token refresh without token in context")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Token not found",
			})
			return
		}

		tokenString, ok := token.(string)
		if !ok {
			appLogger.Info("Token refresh with invalid token format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token format",
			})
			return
		}

		appLogger.Info("Token refresh request")

		// Refresh token (still needs db operations: delete old, save new)
		resp, err := tokenService.Refresh(tokenString)
		if err != nil {
			appLogger.Warn("Token refresh failed", logger.String("error", err.Error()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": err.Error(),
			})
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
