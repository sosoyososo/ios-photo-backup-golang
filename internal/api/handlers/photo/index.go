package photo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/api/middleware"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// IndexHandlerWithDeps handles photo indexing requests with dependency injection
func IndexHandlerWithDeps(db *gorm.DB, naming *service.PhotoNaming, fileStorage *service.FileStorage, storageDir string, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT middleware context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			appLogger.Warn("Photo index request without valid user_id in context", logger.String("path", c.Request.URL.Path))
			errors.Unauthorized(c, "Invalid token claims")
			return
		}

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
