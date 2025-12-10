package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ios-photo-backup/photo-backup-server/internal/config"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
)

var listUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "List all users",
	Long:  "List all user accounts in the Photo Backup Server",
	Run:   runListUsers,
}

func init() {
	rootCmd.AddCommand(listUsersCmd)
}

func runListUsers(cmd *cobra.Command, args []string) {
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

	// Get all users
	users, err := userRepo.ListAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing users: %v\n", err)
		os.Exit(1)
	}

	// Print users
	if len(users) == 0 {
		fmt.Println("No users found")
		return
	}

	fmt.Printf("Total users: %d\n\n", len(users))
	fmt.Printf("%-5s %-20s %-20s\n", "ID", "Username", "Created At")
	fmt.Println(strings.Repeat("-", 50))

	for _, user := range users {
		fmt.Printf("%-5d %-20s %-20s\n",
			user.ID,
			user.Username,
			user.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
}
