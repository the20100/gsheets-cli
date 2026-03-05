package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/the20100/gsheets-cli/internal/output"
)

var sheetCmd = &cobra.Command{
	Use:   "sheet",
	Short: "Manage sheets (tabs) within a spreadsheet",
}

// ---- sheet list ----

var sheetListCmd = &cobra.Command{
	Use:   "list <spreadsheet-id>",
	Short: "List all sheets/tabs in a spreadsheet",
	Long: `List all sheets (tabs) in a Google Spreadsheet.

Examples:
  gsheets sheet list 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
  gsheets sheet list 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sheets, err := client.ListSheets(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(sheets, output.IsPretty(cmd))
		}
		if len(sheets) == 0 {
			fmt.Println("No sheets found.")
			return nil
		}
		headers := []string{"SHEET ID", "TITLE", "INDEX", "TYPE", "ROWS", "COLS"}
		rows := make([][]string, len(sheets))
		for i, sh := range sheets {
			rows[i] = []string{
				fmt.Sprintf("%d", sh.ID),
				output.Truncate(sh.Title, 40),
				fmt.Sprintf("%d", sh.Index),
				sh.Type,
				fmt.Sprintf("%d", sh.Rows),
				fmt.Sprintf("%d", sh.Cols),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

// ---- sheet add ----

var sheetAddIndex int

var sheetAddCmd = &cobra.Command{
	Use:   "add <spreadsheet-id> <title>",
	Short: "Add a new sheet/tab to a spreadsheet",
	Long: `Add a new sheet (tab) to a Google Spreadsheet.

Use --index to specify the position (0-based). Without --index, the sheet is appended.

Examples:
  gsheets sheet add 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms "Q1 Data"
  gsheets sheet add 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms "Summary" --index 0`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		spreadsheetID := args[0]
		title := args[1]
		info, err := client.AddSheet(spreadsheetID, title, sheetAddIndex)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(info, output.IsPretty(cmd))
		}
		fmt.Printf("Sheet added: %s\n", info.Title)
		fmt.Printf("Sheet ID: %d\n", info.ID)
		fmt.Printf("Index:    %d\n", info.Index)
		return nil
	},
}

// ---- sheet delete ----

var sheetDeleteCmd = &cobra.Command{
	Use:   "delete <spreadsheet-id> <sheet-id>",
	Short: "Delete a sheet/tab by its numeric sheet ID",
	Long: `Delete a sheet (tab) from a Google Spreadsheet.

Use 'gsheets sheet list <spreadsheet-id>' to find sheet IDs.
WARNING: This permanently deletes the sheet and all its data.

Examples:
  gsheets sheet delete 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms 0
  gsheets sheet delete 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms 1234567890`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		spreadsheetID := args[0]
		sheetID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid sheet ID %q (must be numeric): %w", args[1], err)
		}
		if err := client.DeleteSheet(spreadsheetID, sheetID); err != nil {
			return err
		}
		fmt.Printf("Sheet %d deleted.\n", sheetID)
		return nil
	},
}

// ---- sheet rename ----

var sheetRenameCmd = &cobra.Command{
	Use:   "rename <spreadsheet-id> <sheet-id> <new-title>",
	Short: "Rename a sheet/tab",
	Long: `Rename a sheet (tab) in a Google Spreadsheet.

Use 'gsheets sheet list <spreadsheet-id>' to find sheet IDs.

Examples:
  gsheets sheet rename 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms 0 "January"
  gsheets sheet rename 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms 1234567890 "New Name"`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		spreadsheetID := args[0]
		sheetID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid sheet ID %q (must be numeric): %w", args[1], err)
		}
		newTitle := args[2]
		if err := client.RenameSheet(spreadsheetID, sheetID, newTitle); err != nil {
			return err
		}
		fmt.Printf("Sheet %d renamed to: %s\n", sheetID, newTitle)
		return nil
	},
}

func init() {
	sheetAddCmd.Flags().IntVar(&sheetAddIndex, "index", -1, "Position of the new sheet (0-based, default: append)")

	sheetCmd.AddCommand(sheetListCmd, sheetAddCmd, sheetDeleteCmd, sheetRenameCmd)
	rootCmd.AddCommand(sheetCmd)

	RegisterSchema("sheet.list", SchemaEntry{
		Command:     "gsheets sheet list <spreadsheet-id>",
		Description: "List all sheets/tabs in a spreadsheet",
		Args:        []SchemaArg{{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"}},
		Flags: []SchemaFlag{
			{Name: "--json", Type: "bool", Desc: "Force JSON output"},
		},
		Examples: []string{"gsheets sheet list 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms"},
		Mutating: false,
	})
	RegisterSchema("sheet.add", SchemaEntry{
		Command:     "gsheets sheet add <spreadsheet-id> <title>",
		Description: "Add a new sheet/tab to a spreadsheet",
		Args: []SchemaArg{
			{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"},
			{Name: "title", Required: true, Desc: "Title for the new sheet"},
		},
		Flags: []SchemaFlag{
			{Name: "--index", Type: "int", Default: "-1", Desc: "Position of the new sheet (0-based, default: append)"},
			{Name: "--json", Type: "bool", Desc: "Force JSON output"},
		},
		Examples: []string{
			`gsheets sheet add SPREADSHEET_ID "Q1 Data"`,
			`gsheets sheet add SPREADSHEET_ID "Summary" --index 0`,
		},
		Mutating: true,
	})
	RegisterSchema("sheet.delete", SchemaEntry{
		Command:     "gsheets sheet delete <spreadsheet-id> <sheet-id>",
		Description: "Delete a sheet/tab by its numeric sheet ID (WARNING: permanent)",
		Args: []SchemaArg{
			{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"},
			{Name: "sheet-id", Required: true, Desc: "Numeric sheet ID (from sheet.list)"},
		},
		Examples: []string{"gsheets sheet delete SPREADSHEET_ID 1234567890"},
		Mutating: true,
	})
	RegisterSchema("sheet.rename", SchemaEntry{
		Command:     "gsheets sheet rename <spreadsheet-id> <sheet-id> <new-title>",
		Description: "Rename a sheet/tab",
		Args: []SchemaArg{
			{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"},
			{Name: "sheet-id", Required: true, Desc: "Numeric sheet ID (from sheet.list)"},
			{Name: "new-title", Required: true, Desc: "New title for the sheet"},
		},
		Examples: []string{`gsheets sheet rename SPREADSHEET_ID 1234567890 "January"`},
		Mutating: true,
	})
}
