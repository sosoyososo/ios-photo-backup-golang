package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/api/handlers/auth"
	"github.com/ios-photo-backup/photo-backup-server/internal/api/handlers/photo"
	"github.com/ios-photo-backup/photo-backup-server/internal/api/handlers/user"
	"github.com/ios-photo-backup/photo-backup-server/internal/api/middleware"
	"github.com/ios-photo-backup/photo-backup-server/internal/config"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
	"github.com/ios-photo-backup/photo-backup-server/internal/service"
)

// SetupRoutes sets up all API routes
func SetupRoutes(db *gorm.DB, cfg *config.Config, appLogger *logger.Logger) *gin.Engine {
	// Create Gin router
	router := gin.Default()

	// Add logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		appLogger.HTTPRequest(
			param.Method,
			param.Path,
			param.ClientIP,
			param.StatusCode,
			param.Latency,
		)
		return ""
	}))
	router.Use(gin.Recovery())

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Create services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecretPath)
	tokenService := service.NewTokenService(tokenRepo, cfg.JWTSecretPath)

	// Create photo services (photo repository will be created per-request with user ID)
	naming := service.NewPhotoNaming()
	fileStorage := service.NewFileStorage(cfg)

	// Public routes (no authentication required)
	public := router.Group("/")
	{
		public.POST("/login", auth.LoginHandler(authService, appLogger))
	}

	// Protected routes (authentication required)
	protected := router.Group("/")
	protected.Use(middleware.JWTMiddleware(tokenService, appLogger))
	{
		// Add refresh and status endpoints
		protected.POST("/refresh", auth.RefreshHandler(tokenService, appLogger))
		protected.GET("/status", user.StatusHandler(appLogger))

		// Add photo endpoints
		// PhotoRepository will be created per-request with user ID from JWT
		protected.POST("/photos/index", photo.IndexHandlerWithDeps(db, naming, fileStorage, cfg.StorageDir, appLogger))
		protected.POST("/photos/upload", photo.UploadHandlerWithDeps(db, naming, fileStorage, cfg.StorageDir, appLogger))
	}

	// Add a simple health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	return router
}
