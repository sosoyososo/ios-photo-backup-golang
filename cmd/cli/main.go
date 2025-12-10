package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "photo-backup-cli",
	Short: "Photo Backup Server CLI",
	Long:  "Command-line tool for managing Photo Backup Server",
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().String("db-path", "./data/app.db", "Database file path")
	viper.BindPFlag("db_path", rootCmd.PersistentFlags().Lookup("db-path"))
}

func initConfig() {
	// Configuration is handled through viper
	viper.SetConfigName(".photo-backup-cli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
