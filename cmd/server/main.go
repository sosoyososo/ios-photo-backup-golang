package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ios-photo-backup/photo-backup-server/internal/config"
	"github.com/ios-photo-backup/photo-backup-server/internal/logger"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
	"github.com/ios-photo-backup/photo-backup-server/internal/api/routes"
)

func main() {
	// Initialize structured logger
	appLogger := logger.NewDefault()
	defer appLogger.Close()

	appLogger.Info("Starting Photo Backup Server", logger.String("version", "1.0.0"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		appLogger.Error("Failed to load configuration", logger.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("Configuration loaded", logger.String("db_path", cfg.DatabasePath))

	// Initialize application (create directories, JWT secret)
	if err := config.InitializeApp(cfg); err != nil {
		appLogger.Error("Failed to initialize application", logger.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("Application initialized", logger.String("storage_dir", cfg.StorageDir))

	// Setup custom temp directory for multipart uploads
	tmpDir := cfg.StorageDir + "/tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		appLogger.Error("Failed to create temp directory", logger.String("error", err.Error()))
		os.Exit(1)
	}
	if err := os.Setenv("TMPDIR", tmpDir); err != nil {
		appLogger.Error("Failed to set TMPDIR", logger.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("Custom temp directory configured", logger.String("tmp_dir", tmpDir))

	// Initialize database
	db, err := repository.InitDB(cfg.DatabasePath)
	if err != nil {
		appLogger.Error("Failed to initialize database", logger.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("Database initialized", logger.String("db_path", cfg.DatabasePath))

	// Setup routes
	router := routes.SetupRoutes(db, cfg, appLogger)
	appLogger.Info("Routes configured")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	serverErr := make(chan error, 1)

	go func() {
		appLogger.Info("Server starting", logger.String("address", addr))
		if err := router.Run(addr); err != nil {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		appLogger.Info("Received shutdown signal", logger.String("signal", sig.String()))
	case err := <-serverErr:
		appLogger.Error("Server error", logger.String("error", err.Error()))
	}

	// Graceful shutdown
	appLogger.Info("Starting graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// TODO: Close database connections and cleanup
	// This would typically involve calling db.Close() if available

	// Wait for shutdown to complete
	<-shutdownCtx.Done()
	appLogger.Info("Server shutdown complete")
}
