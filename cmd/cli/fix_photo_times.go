package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ios-photo-backup/photo-backup-server/internal/config"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
)

var fixPhotoTimesCmd = &cobra.Command{
	Use:   "fix-photo-times",
	Short: "Fix file timestamps based on photo creation time",
	Long:  "Fix file timestamps based on photo creation time for a specific user",
	Run:   runFixPhotoTimes,
}

var (
	dryRun     bool
	storageDir string
)

func init() {
	fixPhotoTimesCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	fixPhotoTimesCmd.Flags().StringVar(&storageDir, "storage-dir", "", "Storage directory path (required)")
	fixPhotoTimesCmd.MarkFlagRequired("storage-dir")
	rootCmd.AddCommand(fixPhotoTimesCmd)
}

func runFixPhotoTimes(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get user ID from args
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: user ID is required\n")
		cmd.Usage()
		os.Exit(1)
	}

	var userID uint
	if _, err := fmt.Sscanf(args[0], "%d", &userID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid user ID: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := repository.InitDB(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	// Create photo repository for the user
	photoRepo := repository.NewPhotoRepository(db, userID)

	// Get all photos for the user
	photos, err := photoRepo.GetAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting photos: %v\n", err)
		os.Exit(1)
	}

	if len(photos) == 0 {
		fmt.Println("No photos found for user")
		return
	}

	// Counters for statistics
	totalPhotos := len(photos)
	totalFiles := 0
	fixedFiles := 0
	skippedFiles := 0
	failedFiles := 0

	fmt.Printf("Processing %d photos for user %d...\n\n", totalPhotos, userID)

	for _, photo := range photos {
		// Parse uploaded extensions
		var extensions []string
		if photo.UploadedExtensions != "" && photo.UploadedExtensions != "[]" {
			if err := json.Unmarshal([]byte(photo.UploadedExtensions), &extensions); err != nil {
				fmt.Printf("Warning: failed to parse extensions for photo %s: %v\n", photo.LocalID, err)
				continue
			}
		}

		// Add the main file_type as an extension
		if photo.FileType != "" {
			extensions = append(extensions, photo.FileType)
		}

		if len(extensions) == 0 {
			fmt.Printf("Warning: no extensions found for photo %s\n", photo.LocalID)
			continue
		}

		// Process each extension
		for _, ext := range extensions {
			totalFiles++
			fileName := photo.FileName + "." + ext
			filePath := filepath.Join(storageDir, photo.FilePath, fileName)

			// Check if file exists
			info, err := os.Stat(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("  Skipped: %s (file not found)\n", filePath)
					skippedFiles++
				} else {
					fmt.Printf("  Error: %s: %v\n", filePath, err)
					failedFiles++
				}
				continue
			}

			// Check if timestamp already matches
			modTime := info.ModTime()
			if modTime.Equal(photo.CreationTime) || modTime.Sub(photo.CreationTime) == 0 {
				fmt.Printf("  OK: %s (timestamp already correct)\n", fileName)
				continue
			}

			// Fix timestamp
			if dryRun {
				fmt.Printf("  [DRY-RUN] Would fix: %s -> %s\n", fileName, photo.CreationTime.Format("2006-01-02 15:04:05"))
				fixedFiles++
			} else {
				if err := os.Chtimes(filePath, photo.CreationTime, photo.CreationTime); err != nil {
					fmt.Printf("  Error fixing: %s: %v\n", filePath, err)
					failedFiles++
				} else {
					fmt.Printf("  Fixed: %s -> %s\n", fileName, photo.CreationTime.Format("2006-01-02 15:04:05"))
					fixedFiles++
				}
			}
		}
	}

	// Print summary
	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Total photos: %d\n", totalPhotos)
	fmt.Printf("Total files:  %d\n", totalFiles)
	fmt.Printf("Fixed:        %d\n", fixedFiles)
	fmt.Printf("Skipped:      %d\n", skippedFiles)
	fmt.Printf("Failed:       %d\n", failedFiles)

	if dryRun {
		fmt.Printf("\nNote: This was a dry-run. Run without --dry-run to apply changes.\n")
	}
}
