package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ios-photo-backup/photo-backup-server/internal/config"
	"github.com/ios-photo-backup/photo-backup-server/internal/models"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
)

var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a new user",
	Long:  "Create a new user account for the Photo Backup Server",
	Run:   runCreateUser,
}

var (
	username string
	password string
)

func init() {
	createUserCmd.Flags().StringVarP(&username, "username", "u", "", "Username for the new user (required)")
	createUserCmd.Flags().StringVarP(&password, "password", "p", "", "Password for the new user (required)")
	createUserCmd.MarkFlagRequired("username")
	createUserCmd.MarkFlagRequired("password")
	rootCmd.AddCommand(createUserCmd)
}

func runCreateUser(cmd *cobra.Command, args []string) {
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

	// Check if user already exists
	existingUser, err := userRepo.FindByUsername(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking user: %v\n", err)
		os.Exit(1)
	}
	if existingUser != nil {
		fmt.Fprintf(os.Stderr, "Error: User '%s' already exists\n", username)
		os.Exit(1)
	}

	// Hash password
	passwordHash, err := config.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error hashing password: %v\n", err)
		os.Exit(1)
	}

	// Create user
	user := &models.User{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := userRepo.Create(user); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User '%s' created successfully (ID: %d)\n", user.Username, user.ID)
}
