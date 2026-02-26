package api

// SpreadsheetInfo is a simplified view of a Spreadsheet for display and JSON output.
type SpreadsheetInfo struct {
	ID     string      `json:"id"`
	Title  string      `json:"title"`
	URL    string      `json:"url"`
	Sheets []SheetInfo `json:"sheets"`
}

// SheetInfo represents a sheet/tab within a spreadsheet.
type SheetInfo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Index int64  `json:"index"`
	Type  string `json:"type"`
	Rows  int64  `json:"rows"`
	Cols  int64  `json:"cols"`
}

// ValuesResult represents values read from a range.
type ValuesResult struct {
	Range     string          `json:"range"`
	Dimension string          `json:"dimension"`
	Values    [][]interface{} `json:"values"`
}

// UpdateResult represents the result of a write operation.
type UpdateResult struct {
	SpreadsheetID  string `json:"spreadsheet_id"`
	UpdatedRange   string `json:"updated_range"`
	UpdatedRows    int64  `json:"updated_rows"`
	UpdatedColumns int64  `json:"updated_columns"`
	UpdatedCells   int64  `json:"updated_cells"`
}

// SheetsError is returned when the API responds with an error.
type SheetsError struct {
	StatusCode int
	Message    string
}

func (e *SheetsError) Error() string {
	return e.Message
}
