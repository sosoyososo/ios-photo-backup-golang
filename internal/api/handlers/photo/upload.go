package photo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/errors"
	"github.com/ios-photo-backup/photo-backup-server/internal/api/middleware"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// UploadHandlerWithDeps handles photo upload requests with dependency injection
func UploadHandlerWithDeps(db *gorm.DB, naming *service.PhotoNaming, fileStorage *service.FileStorage, storageDir string, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT middleware context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			appLogger.Warn("Photo upload request without valid user_id in context", logger.String("path", c.Request.URL.Path))
			errors.Unauthorized(c, "Invalid token claims")
			return
		}

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

		// Use file_type as the file extension
		// file_type should be the extension (e.g., "jpg", "heic", "png")
		ext := strings.ToLower(fileType)
		if ext == "" {
			appLogger.Warn("Photo upload missing file extension",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("file_type", fileType))
			errors.BadRequest(c, "file_type is required", nil)
			return
		}

		appLogger.Info("Uploading photo",
			logger.Uint("user_id", userID),
			logger.String("local_id", localID),
			logger.String("file_type", fileType),
			logger.String("file_extension", ext),
			logger.String("uploaded_filename", fileHeader.Filename),
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

// UploadStreamHandlerWithDeps handles photo upload requests using pure streaming (no multipart parsing)
func UploadStreamHandlerWithDeps(db *gorm.DB, naming *service.PhotoNaming, fileStorage *service.FileStorage, storageDir string, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT middleware context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			appLogger.Warn("Photo upload stream request without valid user_id in context", logger.String("path", c.Request.URL.Path))
			errors.Unauthorized(c, "Invalid token claims")
			return
		}

		// Create PhotoRepository for this user
		photoRepo := repository.NewPhotoRepository(db, userID)

		// Create PhotoService for this user
		photoService := service.NewPhotoService(photoRepo, naming, fileStorage, storageDir)

		// Get parameters from query string
		localID := c.Query("local_id")
		fileType := c.Query("file_type")

		if localID == "" {
			appLogger.Warn("Photo upload stream missing local_id", logger.Uint("user_id", userID))
			errors.BadRequest(c, "local_id is required", nil)
			return
		}

		if fileType == "" {
			appLogger.Warn("Photo upload stream missing file_type", logger.Uint("user_id", userID), logger.String("local_id", localID))
			errors.BadRequest(c, "file_type is required", nil)
			return
		}

		// Use file_type as the file extension
		ext := strings.ToLower(fileType)

		appLogger.Info("Streaming photo upload (raw body)",
			logger.Uint("user_id", userID),
			logger.String("local_id", localID),
			logger.String("file_type", fileType),
			logger.String("file_extension", ext))

		// Upload photo directly from request body (pure streaming)
		if err := photoService.UploadPhotoStream(userID, localID, ext, fileType, c.Request.Body); err != nil {
			appLogger.Error("Photo stream upload failed",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.InternalError(c, err.Error(), nil)
			return
		}

		appLogger.PhotoOperation("upload_stream", localID, localID, userID, true)

		// Return success with filename
		photo, err := photoRepo.FindByLocalID(localID)
		if err == nil && photo != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":    "success",
				"message":   "File uploaded (streamed)",
				"local_id":  localID,
				"filename":  photo.FileName + "." + ext,
				"file_path": photo.FilePath + photo.FileName + "." + ext,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":   "success",
				"message":  "File uploaded (streamed)",
				"local_id": localID,
				"filename": localID + "." + ext,
			})
		}
	}
}

// UploadChunkHandlerWithDeps handles chunked photo upload requests
func UploadChunkHandlerWithDeps(db *gorm.DB, naming *service.PhotoNaming, fileStorage *service.FileStorage, storageDir string, appLogger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT middleware context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			appLogger.Warn("Photo chunk upload request without valid user_id in context", logger.String("path", c.Request.URL.Path))
			errors.Unauthorized(c, "Invalid token claims")
			return
		}

		// Create PhotoRepository for this user
		photoRepo := repository.NewPhotoRepository(db, userID)

		// Create PhotoService for this user
		photoService := service.NewPhotoService(photoRepo, naming, fileStorage, storageDir)

		// Parse multipart form
		if err := c.Request.ParseMultipartForm(50 << 20); err != nil { // 50MB max per chunk
			appLogger.Warn("Failed to parse multipart form for chunk upload",
				logger.Uint("user_id", userID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "Failed to parse multipart form", err.Error())
			return
		}

		// Get form fields
		localID := c.Request.FormValue("local_id")
		fileType := c.Request.FormValue("file_type")
		chunkNumberStr := c.Request.FormValue("chunk_number")
		totalChunksStr := c.Request.FormValue("total_chunks")

		if localID == "" {
			appLogger.Warn("Chunk upload missing local_id", logger.Uint("user_id", userID))
			errors.BadRequest(c, "local_id is required", nil)
			return
		}

		if fileType == "" {
			appLogger.Warn("Chunk upload missing file_type", logger.Uint("user_id", userID), logger.String("local_id", localID))
			errors.BadRequest(c, "file_type is required", nil)
			return
		}

		if chunkNumberStr == "" {
			appLogger.Warn("Chunk upload missing chunk_number", logger.Uint("user_id", userID), logger.String("local_id", localID))
			errors.BadRequest(c, "chunk_number is required", nil)
			return
		}

		if totalChunksStr == "" {
			appLogger.Warn("Chunk upload missing total_chunks", logger.Uint("user_id", userID), logger.String("local_id", localID))
			errors.BadRequest(c, "total_chunks is required", nil)
			return
		}

		// Parse chunk numbers
		var chunkNumber, totalChunks int
		if _, err := fmt.Sscanf(chunkNumberStr, "%d", &chunkNumber); err != nil {
			errors.BadRequest(c, "invalid chunk_number format", nil)
			return
		}
		if _, err := fmt.Sscanf(totalChunksStr, "%d", &totalChunks); err != nil {
			errors.BadRequest(c, "invalid total_chunks format", nil)
			return
		}

		// Validate chunk numbers
		if chunkNumber < 0 || chunkNumber >= totalChunks {
			appLogger.Warn("Invalid chunk numbers",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.Int("chunk_number", chunkNumber),
				logger.Int("total_chunks", totalChunks))
			errors.BadRequest(c, "invalid chunk numbers", nil)
			return
		}

		// Use file_type as the file extension
		ext := strings.ToLower(fileType)

		// Get chunk file
		fileHeader, err := c.FormFile("chunk_data")
		if err != nil {
			appLogger.Warn("Chunk upload missing chunk_data",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "chunk_data is required", err.Error())
			return
		}

		appLogger.Info("Uploading photo chunk",
			logger.Uint("user_id", userID),
			logger.String("local_id", localID),
			logger.String("file_type", fileType),
			logger.String("file_extension", ext),
			logger.Int("chunk_number", chunkNumber),
			logger.Int("total_chunks", totalChunks),
			logger.Int("chunk_size", int(fileHeader.Size)))

		// Open and read chunk file
		file, err := fileHeader.Open()
		if err != nil {
			appLogger.Error("Failed to open chunk file",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "Failed to open chunk file", err.Error())
			return
		}
		defer file.Close()

		chunkData := make([]byte, fileHeader.Size)
		if _, err := file.Read(chunkData); err != nil {
			appLogger.Error("Failed to read chunk file",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.String("error", err.Error()))
			errors.BadRequest(c, "Failed to read chunk file", err.Error())
			return
		}

		// Upload chunk
		isComplete, err := photoService.UploadPhotoChunk(userID, localID, ext, chunkNumber, totalChunks, chunkData)
		if err != nil {
			appLogger.Error("Chunk upload failed",
				logger.Uint("user_id", userID),
				logger.String("local_id", localID),
				logger.Int("chunk_number", chunkNumber),
				logger.String("error", err.Error()))
			errors.InternalError(c, err.Error(), nil)
			return
		}

		if isComplete {
			appLogger.PhotoOperation("upload_chunk_complete", localID, localID, userID, true)

			// Return success with filename
			photo, err := photoRepo.FindByLocalID(localID)
			if err == nil && photo != nil {
				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"message":     "File uploaded (chunked)",
					"local_id":    localID,
					"filename":    photo.FileName + "." + ext,
					"file_path":   photo.FilePath + photo.FileName + "." + ext,
					"is_complete": true,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"message":     "File uploaded (chunked)",
					"local_id":    localID,
					"filename":    localID + "." + ext,
					"is_complete": true,
				})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":      "success",
				"message":     "Chunk uploaded",
				"local_id":    localID,
				"chunk_number": chunkNumber,
				"total_chunks": totalChunks,
				"is_complete": false,
			})
		}
	}
}
