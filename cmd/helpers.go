package cmd

import (
	"fmt"
	"strings"

	"github.com/the20100/gsheets-cli/internal/output"
)

// parseRows converts a slice of comma-separated row strings into a 2D interface slice
// suitable for the Google Sheets API values methods.
// Each string in rows represents one row; values within a row are comma-separated.
// Example: ["John,25,Engineer", "Jane,30,Manager"] → [["John","25","Engineer"],["Jane","30","Manager"]]
func parseRows(rows []string) [][]interface{} {
	if len(rows) == 0 {
		return nil
	}
	result := make([][]interface{}, len(rows))
	for i, row := range rows {
		parts := strings.Split(row, ",")
		cells := make([]interface{}, len(parts))
		for j, p := range parts {
			cells[j] = strings.TrimSpace(p)
		}
		result[i] = cells
	}
	return result
}

// colLetter converts a zero-based column index to a spreadsheet column letter (A, B, ..., Z, AA, AB, ...).
func colLetter(n int) string {
	result := ""
	for {
		result = string(rune('A'+n%26)) + result
		n = n/26 - 1
		if n < 0 {
			break
		}
	}
	return result
}

// printValuesTable renders a 2D values slice as a human-readable table.
// useFirstRowAsHeader: if true, the first row is used as column headers.
func printValuesTable(values [][]interface{}, useFirstRowAsHeader bool) {
	if len(values) == 0 {
		fmt.Println("(empty range)")
		return
	}

	// Find max columns
	maxCols := 0
	for _, row := range values {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	var headers []string
	var dataRows [][]interface{}

	if useFirstRowAsHeader && len(values) > 0 {
		for _, h := range values[0] {
			headers = append(headers, fmt.Sprintf("%v", h))
		}
		// Pad headers if data rows have more cols
		for len(headers) < maxCols {
			headers = append(headers, colLetter(len(headers)))
		}
		dataRows = values[1:]
	} else {
		headers = make([]string, maxCols)
		for i := 0; i < maxCols; i++ {
			headers[i] = colLetter(i)
		}
		dataRows = values
	}

	rows := make([][]string, len(dataRows))
	for i, row := range dataRows {
		cells := make([]string, maxCols)
		for j := 0; j < maxCols; j++ {
			if j < len(row) {
				cells[j] = output.Truncate(fmt.Sprintf("%v", row[j]), 50)
			}
		}
		rows[i] = cells
	}
	output.PrintTable(headers, rows)
}
