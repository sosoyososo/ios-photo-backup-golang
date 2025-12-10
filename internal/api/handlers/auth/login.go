package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// LoginHandler handles login requests
func LoginHandler(authService *service.AuthService, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.LoginRequest

		// Bind request
		if err := c.ShouldBindJSON(&req); err != nil {
			appLogger.Warn("Invalid login request", logger.String("error", err.Error()))
			errors.BadRequest(c, "Invalid request format", err.Error())
			return
		}

		appLogger.Info("Login attempt", logger.String("username", req.Username))

		// Authenticate user
		resp, err := authService.Login(&req)
		if err != nil {
			appLogger.Auth(req.Username, "login", false)
			errors.Unauthorized(c, err.Error())
			return
		}

		appLogger.Auth(req.Username, "login", true)

		// Return success
		c.JSON(http.StatusOK, gin.H{
			"token":      resp.Token,
			"expires_at": resp.ExpiresAt,
		})
	}
}
