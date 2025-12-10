package photo

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// UploadHandlerWithDeps handles photo upload requests with dependency injection
func UploadHandlerWithDeps(db *gorm.DB, naming *service.PhotoNaming, fileStorage *service.FileStorage, storageDir string, tokenService *service.TokenService, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from context
		token, exists := c.Get("token")
		if !exists {
			appLogger.Info("Photo upload request without token", logger.String("path", c.Request.URL.Path))
			errors.Unauthorized(c, "Token not found")
			return
		}

		tokenString, ok := token.(string)
		if !ok {
			appLogger.Info("Photo upload request with invalid token format", logger.String("path", c.Request.URL.Path))
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

		// Parse multipart form
		if err := c.Request.ParseMultipartForm(50 << 20); err != nil { // 50MB max
			appLogger.Warn("Failed to parse multipart form",
				logger.Uint("user_id", userID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "Failed to parse multipart form", err.Error())
			return
		}

		// Get form fields
		localID := c.Request.FormValue("local_id")
		fileType := c.Request.FormValue("file_type")

		if localID == "" {
			appLogger.Warn("Photo upload missing local_id", logger.Uint("user_id", userID))
			errors.BadRequest(c, "local_id is required", nil)
			return
		}

		if fileType == "" {
			appLogger.Warn("Photo upload missing file_type", logger.Uint("user_id", userID), logger.String("local_id", localID))
			errors.BadRequest(c, "file_type is required", nil)
			return
		}

		// Get file
		fileHeader, err := c.FormFile("file")
		if err != nil {
			appLogger.Warn("Photo upload missing file",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "file is required", err.Error())
			return
		}

		// Extract file extension from uploaded filename
		uploadedFilename := fileHeader.Filename
		ext := strings.ToLower(filepath.Ext(uploadedFilename))
		if ext == "" {
			appLogger.Warn("Photo upload missing file extension",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("filename", uploadedFilename))
			errors.BadRequest(c, "File must have an extension", nil)
			return
		}
		// Remove the dot from extension (e.g., ".jpg" -> "jpg")
		ext = ext[1:]

		appLogger.Info("Uploading photo",
			logger.Uint("user_id", userID),
			logger.String("local_id", localID),
			logger.String("file_type", fileType),
			logger.String("file_extension", ext),
			logger.String("uploaded_filename", uploadedFilename),
			logger.Int("file_size", int(fileHeader.Size)))

		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			appLogger.Error("Failed to open uploaded file",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "Failed to open file", err.Error())
			return
		}
		defer file.Close()

		// Read file data
		fileData := make([]byte, fileHeader.Size)
		if _, err := file.Read(fileData); err != nil {
			appLogger.Error("Failed to read uploaded file",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "Failed to read file", err.Error())
			return
		}

		// Upload photo with extension
		if err := photoService.UploadPhoto(userID, localID, ext, fileType, fileData); err != nil {
			appLogger.Error("Photo upload failed",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.InternalError(c, err.Error(), nil)
			return
		}

		appLogger.PhotoOperation("upload", localID, localID, userID, true)

		// Return success with filename
		// Build the expected filename from the indexed photo
		photo, err := photoRepo.FindByLocalID(localID)
		if err == nil && photo != nil {
			// Return the actual filename with extension
			c.JSON(http.StatusOK, gin.H{
				"status":    "success",
				"message":   "File uploaded",
				"local_id":  localID,
				"filename":  photo.FileName + "." + ext,
				"file_path": photo.FilePath + photo.FileName + "." + ext,
			})
		} else {
			// Fallback if we can't get photo info
			c.JSON(http.StatusOK, gin.H{
				"status":   "success",
				"message":  "File uploaded",
				"local_id": localID,
				"filename": localID + "." + ext,
			})
		}
	}
}
