package photo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// IndexHandlerWithDeps handles photo indexing requests with dependency injection
func IndexHandlerWithDeps(db *gorm.DB, naming *service.PhotoNaming, fileStorage *service.FileStorage, storageDir string, tokenService *service.TokenService, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from context
		token, exists := c.Get("token")
		if !exists {
			appLogger.Info("Photo index request without token", logger.String("path", c.Request.URL.Path))
			errors.Unauthorized(c, "Token not found")
			return
		}

		tokenString, ok := token.(string)
		if !ok {
			appLogger.Info("Photo index request with invalid token format", logger.String("path", c.Request.URL.Path))
			errors.Unauthorized(c, "Invalid token format")
			return
		}

		// Validate token to get claims
		claims, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			appLogger.Warn("Token validation failed", logger.String("error", err.Error()))
			errors.Unauthorized(c, err.Error())
			return
		}

		// Extract user ID from claims
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			appLogger.Warn("Invalid token claims", logger.String("token", tokenString))
			errors.Unauthorized(c, "Invalid token claims")
			return
		}
		userID := uint(userIDFloat)

		// Create PhotoRepository for this user
		photoRepo := repository.NewPhotoRepository(db, userID)

		// Create PhotoService for this user
		photoService := service.NewPhotoService(photoRepo, naming, fileStorage, storageDir)

		var req struct {
			Date   string                         `json:"date" binding:"required"`
			Photos []service.PhotoIndexRequest    `json:"photos" binding:"required"`
		}

		// Bind request
		if err := c.ShouldBindJSON(&req); err != nil {
			appLogger.Warn("Invalid photo index request",
				logger.Uint("user_id", userID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "Invalid request format", err.Error())
			return
		}

		// Validate date format (YYYY-MM-DD)
		if len(req.Date) != 10 {
			appLogger.Warn("Invalid date format in photo index request",
				logger.Uint("user_id", userID),
				logger.String("date", req.Date))
			errors.BadRequest(c, "Invalid date format, expected YYYY-MM-DD", nil)
			return
		}

		appLogger.Info("Indexing photos",
			logger.Uint("user_id", userID),
			logger.String("date", req.Date),
			logger.Int("photo_count", len(req.Photos)))

		// Index photos
		responses, err := photoService.IndexPhotos(userID, req.Date, req.Photos)
		if err != nil {
			appLogger.Error("Photo indexing failed",
				logger.Uint("user_id", userID),
				logger.String("date", req.Date),
				logger.String("error", err.Error()))
			errors.InternalError(c, err.Error(), nil)
			return
		}

		appLogger.Info("Photo indexing successful",
			logger.Uint("user_id", userID),
			logger.String("date", req.Date),
			logger.Int("indexed_count", len(responses)))

		// Return success
		c.JSON(http.StatusOK, gin.H{
			"status":        "success",
			"date":          req.Date,
			"assigned_files": responses,
		})
	}
}
