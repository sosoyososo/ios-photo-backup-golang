package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/middleware"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
)

// StatusResponse represents status check response
type StatusResponse struct {
	Status   string `json:"status"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// StatusHandler handles status check requests
func StatusHandler(appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from JWT middleware context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			appLogger.Warn("Status check without valid user_id in context")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token claims",
			})
			return
		}

		username, _ := middleware.GetUsername(c)

		// Return success
		c.JSON(http.StatusOK, StatusResponse{
			Status:   "online",
			UserID:   userID,
			Username: username,
		})
	}
}
