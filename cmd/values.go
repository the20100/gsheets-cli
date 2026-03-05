package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/gsheets-cli/internal/output"
)

var valuesCmd = &cobra.Command{
	Use:   "values",
	Short: "Read and write cell values in a spreadsheet",
}

// ---- values get ----

var (
	valuesGetRender string
	valuesGetDim    string
	valuesGetHeader bool
)

var valuesGetCmd = &cobra.Command{
	Use:   "get <spreadsheet-id> <range>",
	Short: "Read values from a range",
	Long: `Read cell values from a Google Spreadsheet range using A1 notation.

Range examples:
  Sheet1!A1:C10    — specific range
  Sheet1!A:C       — entire columns A-C
  Sheet1!1:5       — rows 1-5
  Sheet1           — entire sheet
  A1:B2            — range on the first sheet (no sheet name)

Examples:
  gsheets values get SPREADSHEET_ID "Sheet1!A1:D10"
  gsheets values get SPREADSHEET_ID "Sheet1!A1:D10" --header
  gsheets values get SPREADSHEET_ID "Sheet1!A1:D10" --render FORMULA
  gsheets values get SPREADSHEET_ID "Sheet1!A1:D10" --json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := client.GetValues(args[0], args[1], valuesGetRender, valuesGetDim)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(result, output.IsPretty(cmd))
		}
		if len(result.Values) == 0 {
			fmt.Printf("Range %s is empty.\n", result.Range)
			return nil
		}
		fmt.Printf("Range: %s  (%d rows)\n\n", result.Range, len(result.Values))
		printValuesTable(result.Values, valuesGetHeader)
		return nil
	},
}

// ---- values update ----

var (
	valuesUpdateRows        []string
	valuesUpdateInputOption string
)

var valuesUpdateCmd = &cobra.Command{
	Use:   "update <spreadsheet-id> <range>",
	Short: "Write values to a range",
	Long: `Write values to a Google Spreadsheet range.

Use --row to specify rows of data. Each --row flag is one row;
values within a row are comma-separated. Use multiple --row flags for multiple rows.

Input options (--input):
  USER_ENTERED  — Google interprets the values (numbers, dates, formulas). Default.
  RAW           — Values are stored as-is without interpretation.

Examples:
  # Single cell
  gsheets values update SPREADSHEET_ID "Sheet1!A1" --row "Hello"

  # Single row across columns
  gsheets values update SPREADSHEET_ID "Sheet1!A1:C1" --row "Name,Age,City"

  # Multiple rows
  gsheets values update SPREADSHEET_ID "Sheet1!A2" \
    --row "Alice,30,New York" \
    --row "Bob,25,San Francisco"

  # Formula
  gsheets values update SPREADSHEET_ID "Sheet1!D1" --row "=SUM(A1:C1)"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(valuesUpdateRows) == 0 {
			return fmt.Errorf("at least one --row is required")
		}
		values := parseRows(valuesUpdateRows)
		result, err := client.UpdateValues(args[0], args[1], values, valuesUpdateInputOption)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(result, output.IsPretty(cmd))
		}
		fmt.Printf("Updated range: %s\n", result.UpdatedRange)
		fmt.Printf("Cells updated: %d (%d rows × %d cols)\n",
			result.UpdatedCells, result.UpdatedRows, result.UpdatedColumns)
		return nil
	},
}

// ---- values append ----

var (
	valuesAppendRows         []string
	valuesAppendInputOption  string
	valuesAppendInsertOption string
)

var valuesAppendCmd = &cobra.Command{
	Use:   "append <spreadsheet-id> <range>",
	Short: "Append rows after the last row in a range",
	Long: `Append new rows to the end of existing data in a Google Spreadsheet.

The range specifies the table to append to — data is added after the last
populated row in that range.

Insert options (--insert):
  INSERT_ROWS  — inserts new rows to accommodate the data. Default.
  OVERWRITE    — overwrites data starting after the last populated row.

Examples:
  gsheets values append SPREADSHEET_ID "Sheet1!A:C" --row "Alice,30,Engineer"
  gsheets values append SPREADSHEET_ID "Sheet1!A:C" \
    --row "Alice,30,Engineer" \
    --row "Bob,25,Designer"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(valuesAppendRows) == 0 {
			return fmt.Errorf("at least one --row is required")
		}
		values := parseRows(valuesAppendRows)
		result, err := client.AppendValues(args[0], args[1], values, valuesAppendInputOption, valuesAppendInsertOption)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(result, output.IsPretty(cmd))
		}
		fmt.Printf("Appended to range: %s\n", result.UpdatedRange)
		fmt.Printf("Cells written: %d (%d rows × %d cols)\n",
			result.UpdatedCells, result.UpdatedRows, result.UpdatedColumns)
		return nil
	},
}

// ---- values clear ----

var valuesClearCmd = &cobra.Command{
	Use:   "clear <spreadsheet-id> <range>",
	Short: "Clear values in a range (preserves formatting)",
	Long: `Clear all values in a Google Spreadsheet range.

Cell formatting is preserved — only the values are removed.

Examples:
  gsheets values clear SPREADSHEET_ID "Sheet1!A1:D10"
  gsheets values clear SPREADSHEET_ID "Sheet1!A:A"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		clearedRange, err := client.ClearValues(args[0], args[1])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(map[string]string{"cleared_range": clearedRange}, output.IsPretty(cmd))
		}
		fmt.Printf("Cleared range: %s\n", clearedRange)
		return nil
	},
}

func init() {
	// values get flags
	valuesGetCmd.Flags().StringVar(&valuesGetRender, "render", "",
		"Value render option: FORMATTED_VALUE (default), UNFORMATTED_VALUE, FORMULA")
	valuesGetCmd.Flags().StringVar(&valuesGetDim, "dimension", "",
		"Major dimension: ROWS (default), COLUMNS")
	valuesGetCmd.Flags().BoolVar(&valuesGetHeader, "header", false,
		"Use the first row as column headers in table output")

	// values update flags
	valuesUpdateCmd.Flags().StringArrayVar(&valuesUpdateRows, "row", nil,
		"A row of comma-separated values (repeatable for multiple rows)")
	valuesUpdateCmd.Flags().StringVar(&valuesUpdateInputOption, "input", "USER_ENTERED",
		"Input option: USER_ENTERED (interprets values) or RAW (literal)")

	// values append flags
	valuesAppendCmd.Flags().StringArrayVar(&valuesAppendRows, "row", nil,
		"A row of comma-separated values (repeatable for multiple rows)")
	valuesAppendCmd.Flags().StringVar(&valuesAppendInputOption, "input", "USER_ENTERED",
		"Input option: USER_ENTERED or RAW")
	valuesAppendCmd.Flags().StringVar(&valuesAppendInsertOption, "insert", "INSERT_ROWS",
		"Insert option: INSERT_ROWS (default) or OVERWRITE")

	valuesCmd.AddCommand(valuesGetCmd, valuesUpdateCmd, valuesAppendCmd, valuesClearCmd)
	rootCmd.AddCommand(valuesCmd)

	RegisterSchema("values.get", SchemaEntry{
		Command:     "gsheets values get <spreadsheet-id> <range>",
		Description: "Read values from a range (A1 notation)",
		Args: []SchemaArg{
			{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"},
			{Name: "range", Required: true, Desc: "A1 notation range, e.g. Sheet1!A1:C10"},
		},
		Flags: []SchemaFlag{
			{Name: "--render", Type: "string", Default: "FORMATTED_VALUE", Desc: "FORMATTED_VALUE, UNFORMATTED_VALUE, or FORMULA"},
			{Name: "--dimension", Type: "string", Default: "ROWS", Desc: "Major dimension: ROWS or COLUMNS"},
			{Name: "--header", Type: "bool", Desc: "Use first row as column headers in table output"},
			{Name: "--json", Type: "bool", Desc: "Force JSON output"},
		},
		Examples: []string{
			`gsheets values get SPREADSHEET_ID "Sheet1!A1:D10"`,
			`gsheets values get SPREADSHEET_ID "Sheet1!A1:D10" --header`,
			`gsheets values get SPREADSHEET_ID "Sheet1!A1:D10" --json`,
		},
		Mutating: false,
	})
	RegisterSchema("values.update", SchemaEntry{
		Command:     "gsheets values update <spreadsheet-id> <range>",
		Description: "Write values to a range",
		Args: []SchemaArg{
			{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"},
			{Name: "range", Required: true, Desc: "A1 notation range starting cell, e.g. Sheet1!A1"},
		},
		Flags: []SchemaFlag{
			{Name: "--row", Type: "string[]", Required: true, Desc: "Comma-separated row values (repeatable for multiple rows)"},
			{Name: "--input", Type: "string", Default: "USER_ENTERED", Desc: "USER_ENTERED (interprets values) or RAW (literal)"},
			{Name: "--json", Type: "bool", Desc: "Force JSON output"},
		},
		Examples: []string{
			`gsheets values update SPREADSHEET_ID "Sheet1!A1" --row "Hello"`,
			`gsheets values update SPREADSHEET_ID "Sheet1!A1:C1" --row "Name,Age,City"`,
			`gsheets values update SPREADSHEET_ID "Sheet1!A2" --row "Alice,30,NY" --row "Bob,25,SF"`,
		},
		Mutating: true,
	})
	RegisterSchema("values.append", SchemaEntry{
		Command:     "gsheets values append <spreadsheet-id> <range>",
		Description: "Append rows after the last row in a range",
		Args: []SchemaArg{
			{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"},
			{Name: "range", Required: true, Desc: "A1 notation range to append to, e.g. Sheet1!A:C"},
		},
		Flags: []SchemaFlag{
			{Name: "--row", Type: "string[]", Required: true, Desc: "Comma-separated row values (repeatable)"},
			{Name: "--input", Type: "string", Default: "USER_ENTERED", Desc: "USER_ENTERED or RAW"},
			{Name: "--insert", Type: "string", Default: "INSERT_ROWS", Desc: "INSERT_ROWS or OVERWRITE"},
		},
		Examples: []string{
			`gsheets values append SPREADSHEET_ID "Sheet1!A:C" --row "Alice,30,Engineer"`,
		},
		Mutating: true,
	})
	RegisterSchema("values.clear", SchemaEntry{
		Command:     "gsheets values clear <spreadsheet-id> <range>",
		Description: "Clear values in a range (preserves formatting)",
		Args: []SchemaArg{
			{Name: "spreadsheet-id", Required: true, Desc: "The spreadsheet ID"},
			{Name: "range", Required: true, Desc: "A1 notation range to clear"},
		},
		Examples: []string{
			`gsheets values clear SPREADSHEET_ID "Sheet1!A1:D10"`,
		},
		Mutating: true,
	})
}
