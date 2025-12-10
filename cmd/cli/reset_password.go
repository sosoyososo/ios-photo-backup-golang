package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ios-photo-backup/photo-backup-server/internal/config"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
)

var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Reset user password",
	Long:  "Reset the password for an existing user",
	Run:   runResetPassword,
}

var (
	resetUsername string
	resetPassword string
)

func init() {
	resetPasswordCmd.Flags().StringVarP(&resetUsername, "username", "u", "", "Username for password reset (required)")
	resetPasswordCmd.Flags().StringVarP(&resetPassword, "password", "p", "", "New password for the user (required)")
	resetPasswordCmd.MarkFlagRequired("username")
	resetPasswordCmd.MarkFlagRequired("password")
	rootCmd.AddCommand(resetPasswordCmd)
}

func runResetPassword(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := repository.InitDB(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	// Create user repository
	userRepo := repository.NewUserRepository(db)

	// Find user
	user, err := userRepo.FindByUsername(resetUsername)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding user: %v\n", err)
		os.Exit(1)
	}
	if user == nil {
		fmt.Fprintf(os.Stderr, "Error: User '%s' not found\n", resetUsername)
		os.Exit(1)
	}

	// Hash new password
	passwordHash, err := config.HashPassword(resetPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error hashing password: %v\n", err)
		os.Exit(1)
	}

	// Update password
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	if err := userRepo.Update(user); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating password: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Password reset successfully for user '%s'\n", user.Username)
}
