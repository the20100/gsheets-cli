package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/the20100/gsheets-cli/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Google Sheets authentication",
}

var authSetCredentialsCmd = &cobra.Command{
	Use:   "set-credentials <path-to-service-account-json>",
	Short: "Save a service account credentials file path to the config",
	Long: `Save the path to a Google service account JSON key file.

To create a service account:
  1. Go to https://console.cloud.google.com/iam-admin/serviceaccounts
  2. Create a service account and grant it access to your spreadsheets
  3. Create a JSON key and download it
  4. Run: gsheets auth set-credentials /path/to/key.json

Alternatively, set the GOOGLE_APPLICATION_CREDENTIALS env var,
or use: gcloud auth application-default login (for user credentials).

The path is stored at:
  macOS:   ~/Library/Application Support/gsheets/config.json
  Linux:   ~/.config/gsheets/config.json
  Windows: %AppData%\gsheets\config.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		// Verify the file exists
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("credentials file not found: %s", path)
		}
		if err := config.Save(&config.Config{CredentialsFile: path}); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Credentials path saved to %s\n", config.Path())
		fmt.Printf("File: %s\n", path)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Printf("Config: %s\n\n", config.Path())

		if v := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); v != "" {
			fmt.Println("Source: GOOGLE_APPLICATION_CREDENTIALS env var (takes priority)")
			fmt.Printf("File:   %s\n", v)
		} else if v := os.Getenv("GSHEETS_CREDENTIALS"); v != "" {
			fmt.Println("Source: GSHEETS_CREDENTIALS env var")
			fmt.Printf("File:   %s\n", v)
		} else if c.CredentialsFile != "" {
			fmt.Println("Source: config file")
			fmt.Printf("File:   %s\n", c.CredentialsFile)
		} else {
			fmt.Println("Source: Application Default Credentials (ADC)")
			fmt.Println("       (uses GOOGLE_APPLICATION_CREDENTIALS or gcloud ADC)")
			fmt.Println()
			fmt.Println("To configure explicitly, run one of:")
			fmt.Println("  gsheets auth set-credentials /path/to/sa.json")
			fmt.Println("  export GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json")
			fmt.Println("  gcloud auth application-default login")
		}
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the saved credentials path from the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return fmt.Errorf("removing config: %w", err)
		}
		fmt.Println("Credentials path removed from config.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authSetCredentialsCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
