package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/the20100/gsheets-cli/internal/api"
	"github.com/the20100/gsheets-cli/internal/config"
)

var (
	jsonFlag   bool
	prettyFlag bool
	client     *api.Client
	cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "gsheets",
	Short: "Google Sheets CLI — manage spreadsheets via the API",
	Long: `gsheets is a CLI tool for the Google Sheets API.

It outputs JSON when piped (for agent use) and human-readable tables in a terminal.

Credential resolution order:
  1. GOOGLE_APPLICATION_CREDENTIALS env var (path to service account JSON)
  2. GSHEETS_CREDENTIALS env var (path to service account JSON)
  3. Config file  (~/.config/gsheets/config.json  via: gsheets auth set-credentials)
  4. Application Default Credentials (ADC) — gcloud auth application-default login

Examples:
  gsheets auth set-credentials /path/to/sa.json
  gsheets spreadsheet create "My Sheet"
  gsheets spreadsheet get <spreadsheet-id>
  gsheets sheet list <spreadsheet-id>
  gsheets values get <spreadsheet-id> "Sheet1!A1:C10"
  gsheets values update <spreadsheet-id> "Sheet1!A1" --row "Hello,World"`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "Force pretty-printed JSON output (implies --json)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if isAuthCommand(cmd) || cmd.Name() == "info" || cmd.Name() == "update" || cmd.Name() == "schema" {
			return nil
		}
		credsFile, err := resolveCredentials()
		if err != nil {
			return err
		}
		client, err = api.NewClient(credsFile)
		if err != nil {
			return fmt.Errorf("initializing Google Sheets client: %w", err)
		}
		return nil
	}

	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show tool info: config path, auth status, and environment",
	Run: func(cmd *cobra.Command, args []string) {
		printInfo()
	},
}

func printInfo() {
	fmt.Printf("gsheets — Google Sheets CLI\n\n")
	exe, _ := os.Executable()
	fmt.Printf("  binary:  %s\n", exe)
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Println("  config paths by OS:")
	fmt.Printf("    macOS:    ~/Library/Application Support/gsheets/config.json\n")
	fmt.Printf("    Linux:    ~/.config/gsheets/config.json\n")
	fmt.Printf("    Windows:  %%AppData%%\\gsheets\\config.json\n")
	fmt.Printf("  config:   %s\n", config.Path())
	fmt.Println()
	fmt.Printf("  GOOGLE_APPLICATION_CREDENTIALS = %s\n", maskOrEmpty(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	fmt.Printf("  GSHEETS_CREDENTIALS            = %s\n", maskOrEmpty(os.Getenv("GSHEETS_CREDENTIALS")))
}

func maskOrEmpty(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

// resolveEnv returns the value of the first non-empty environment variable from the given names.
func resolveEnv(names ...string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

// resolveCredentials returns the path to the credentials file, or "" for ADC.
func resolveCredentials() (string, error) {
	// 1. Try all credentials file env var aliases.
	if v := resolveEnv(
		"GOOGLE_APPLICATION_CREDENTIALS",
		"GSHEETS_CREDENTIALS",
		"GOOGLE_CREDENTIALS",
		"GCP_APPLICATION_CREDENTIALS",
		"GCP_CREDENTIALS",
		"GOOGLE_SERVICE_ACCOUNT_FILE",
		"GSHEETS_SA_FILE",
		"GCLOUD_CREDENTIALS",
	); v != "" {
		return v, nil
	}
	// 2. Config file
	var err error
	cfg, err = config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.CredentialsFile != "" {
		return cfg.CredentialsFile, nil
	}
	// 3. Fall through to ADC (Application Default Credentials)
	// This works when GOOGLE_APPLICATION_CREDENTIALS is set at OS level,
	// or after running: gcloud auth application-default login
	return "", nil
}

func isAuthCommand(cmd *cobra.Command) bool {
	if cmd.Name() == "auth" {
		return true
	}
	p := cmd.Parent()
	for p != nil {
		if p.Name() == "auth" {
			return true
		}
		p = p.Parent()
	}
	return false
}
