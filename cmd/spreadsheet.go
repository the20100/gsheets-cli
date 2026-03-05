package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/gsheets-cli/internal/output"
)

var spreadsheetCmd = &cobra.Command{
	Use:   "spreadsheet",
	Short: "Manage Google Spreadsheets",
}

// ---- spreadsheet get ----

var spreadsheetGetCmd = &cobra.Command{
	Use:   "get <spreadsheet-id>",
	Short: "Get spreadsheet metadata (title, sheets, URL)",
	Long: `Get metadata for a Google Spreadsheet.

The spreadsheet ID can be found in the URL:
  https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/edit

Examples:
  gsheets spreadsheet get 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
  gsheets spreadsheet get 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := client.GetSpreadsheet(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(info, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", info.ID},
			{"Title", info.Title},
			{"Sheets", fmt.Sprintf("%d", len(info.Sheets))},
			{"URL", info.URL},
		})
		if len(info.Sheets) > 0 {
			fmt.Println()
			headers := []string{"SHEET ID", "TITLE", "INDEX", "ROWS", "COLS"}
			rows := make([][]string, len(info.Sheets))
			for i, sh := range info.Sheets {
				rows[i] = []string{
					fmt.Sprintf("%d", sh.ID),
					output.Truncate(sh.Title, 40),
					fmt.Sprintf("%d", sh.Index),
					fmt.Sprintf("%d", sh.Rows),
					fmt.Sprintf("%d", sh.Cols),
				}
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

// ---- spreadsheet create ----

var spreadsheetCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new Google Spreadsheet",
	Long: `Create a new Google Spreadsheet with the given title.

The new spreadsheet will be created in the authenticated user's Drive
(or in the service account's accessible Drive).

Examples:
  gsheets spreadsheet create "My Budget 2026"
  gsheets spreadsheet create "Sales Data" --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := client.CreateSpreadsheet(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(info, output.IsPretty(cmd))
		}
		fmt.Printf("Spreadsheet created: %s\n", info.Title)
		fmt.Printf("ID:  %s\n", info.ID)
		fmt.Printf("URL: %s\n", info.URL)
		return nil
	},
}

func init() {
	spreadsheetCmd.AddCommand(spreadsheetGetCmd, spreadsheetCreateCmd)
	rootCmd.AddCommand(spreadsheetCmd)

	RegisterSchema("spreadsheet.get", SchemaEntry{
		Command:     "gsheets spreadsheet get <spreadsheet-id>",
		Description: "Get spreadsheet metadata (title, sheets, URL)",
		Args:        []SchemaArg{{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID from the URL"}},
		Flags: []SchemaFlag{
			{Name: "--json", Type: "bool", Desc: "Force JSON output"},
			{Name: "--pretty", Type: "bool", Desc: "Force pretty-printed JSON output"},
		},
		Examples: []string{
			"gsheets spreadsheet get 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
			"gsheets spreadsheet get 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms --json",
		},
		Mutating: false,
	})
	RegisterSchema("spreadsheet.create", SchemaEntry{
		Command:     "gsheets spreadsheet create <title>",
		Description: "Create a new Google Spreadsheet",
		Args:        []SchemaArg{{Name: "title", Required: true, Desc: "Title of the new spreadsheet"}},
		Flags: []SchemaFlag{
			{Name: "--json", Type: "bool", Desc: "Force JSON output"},
		},
		Examples: []string{
			`gsheets spreadsheet create "My Budget 2026"`,
			`gsheets spreadsheet create "Sales Data" --json`,
		},
		Mutating: true,
	})
}
