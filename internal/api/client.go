package api

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Client wraps the official Google Sheets API service.
type Client struct {
	srv *sheets.Service
}

// NewClient creates a new Sheets client.
// If credentialsFile is empty, it uses Application Default Credentials
// (GOOGLE_APPLICATION_CREDENTIALS env var or gcloud ADC).
func NewClient(credentialsFile string) (*Client, error) {
	ctx := context.Background()
	opts := []option.ClientOption{}
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}
	return &Client{srv: srv}, nil
}

// ---- Spreadsheet methods ----

// GetSpreadsheet retrieves spreadsheet metadata.
func (c *Client) GetSpreadsheet(id string) (*SpreadsheetInfo, error) {
	resp, err := c.srv.Spreadsheets.Get(id).Do()
	if err != nil {
		return nil, err
	}
	return spreadsheetToInfo(resp), nil
}

// CreateSpreadsheet creates a new spreadsheet with the given title.
func (c *Client) CreateSpreadsheet(title string) (*SpreadsheetInfo, error) {
	resp, err := c.srv.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}).Do()
	if err != nil {
		return nil, err
	}
	return spreadsheetToInfo(resp), nil
}

// ---- Sheet methods ----

// ListSheets returns all sheets/tabs in a spreadsheet.
func (c *Client) ListSheets(spreadsheetID string) ([]SheetInfo, error) {
	resp, err := c.srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, err
	}
	return spreadsheetToInfo(resp).Sheets, nil
}

// AddSheet adds a new sheet/tab to a spreadsheet.
// index < 0 means append at the end.
func (c *Client) AddSheet(spreadsheetID, title string, index int) (*SheetInfo, error) {
	addReq := &sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title: title,
		},
	}
	if index >= 0 {
		addReq.Properties.Index = int64(index)
	}
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{AddSheet: addReq},
		},
	}
	resp, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	if err != nil {
		return nil, err
	}
	for _, reply := range resp.Replies {
		if reply.AddSheet != nil {
			info := sheetToInfo(reply.AddSheet.Properties)
			return &info, nil
		}
	}
	return nil, fmt.Errorf("unexpected response: no AddSheet reply")
}

// DeleteSheet deletes a sheet by its numeric sheet ID.
func (c *Client) DeleteSheet(spreadsheetID string, sheetID int64) error {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{DeleteSheet: &sheets.DeleteSheetRequest{SheetId: sheetID}},
		},
	}
	_, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	return err
}

// RenameSheet renames a sheet by its numeric sheet ID.
func (c *Client) RenameSheet(spreadsheetID string, sheetID int64, newTitle string) error {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: &sheets.SheetProperties{
						SheetId: sheetID,
						Title:   newTitle,
					},
					Fields: "title",
				},
			},
		},
	}
	_, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, req).Do()
	return err
}

// ---- Values methods ----

// GetValues reads values from a range.
// renderOption: FORMATTED_VALUE (default), UNFORMATTED_VALUE, FORMULA
// dimension: ROWS (default), COLUMNS
func (c *Client) GetValues(spreadsheetID, rangeStr, renderOption, dimension string) (*ValuesResult, error) {
	call := c.srv.Spreadsheets.Values.Get(spreadsheetID, rangeStr)
	if renderOption != "" {
		call = call.ValueRenderOption(renderOption)
	}
	if dimension != "" {
		call = call.MajorDimension(dimension)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, err
	}
	return &ValuesResult{
		Range:     resp.Range,
		Dimension: resp.MajorDimension,
		Values:    resp.Values,
	}, nil
}

// UpdateValues writes values to a range.
// inputOption: USER_ENTERED (default), RAW
func (c *Client) UpdateValues(spreadsheetID, rangeStr string, values [][]interface{}, inputOption string) (*UpdateResult, error) {
	if inputOption == "" {
		inputOption = "USER_ENTERED"
	}
	vr := &sheets.ValueRange{
		Values: values,
	}
	resp, err := c.srv.Spreadsheets.Values.Update(spreadsheetID, rangeStr, vr).
		ValueInputOption(inputOption).Do()
	if err != nil {
		return nil, err
	}
	return &UpdateResult{
		SpreadsheetID:  resp.SpreadsheetId,
		UpdatedRange:   resp.UpdatedRange,
		UpdatedRows:    resp.UpdatedRows,
		UpdatedColumns: resp.UpdatedColumns,
		UpdatedCells:   resp.UpdatedCells,
	}, nil
}

// AppendValues appends values after the last row in a range.
// inputOption: USER_ENTERED (default), RAW
// insertOption: INSERT_ROWS (default), OVERWRITE
func (c *Client) AppendValues(spreadsheetID, rangeStr string, values [][]interface{}, inputOption, insertOption string) (*UpdateResult, error) {
	if inputOption == "" {
		inputOption = "USER_ENTERED"
	}
	if insertOption == "" {
		insertOption = "INSERT_ROWS"
	}
	vr := &sheets.ValueRange{
		Values: values,
	}
	resp, err := c.srv.Spreadsheets.Values.Append(spreadsheetID, rangeStr, vr).
		ValueInputOption(inputOption).
		InsertDataOption(insertOption).Do()
	if err != nil {
		return nil, err
	}
	result := &UpdateResult{SpreadsheetID: resp.SpreadsheetId}
	if resp.Updates != nil {
		result.UpdatedRange = resp.Updates.UpdatedRange
		result.UpdatedRows = resp.Updates.UpdatedRows
		result.UpdatedColumns = resp.Updates.UpdatedColumns
		result.UpdatedCells = resp.Updates.UpdatedCells
	}
	return result, nil
}

// ClearValues clears values in a range (preserves formatting).
func (c *Client) ClearValues(spreadsheetID, rangeStr string) (string, error) {
	resp, err := c.srv.Spreadsheets.Values.Clear(spreadsheetID, rangeStr, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return "", err
	}
	return resp.ClearedRange, nil
}

// ---- helpers ----

func spreadsheetToInfo(s *sheets.Spreadsheet) *SpreadsheetInfo {
	info := &SpreadsheetInfo{
		ID:    s.SpreadsheetId,
		Title: s.Properties.Title,
		URL:   s.SpreadsheetUrl,
	}
	for _, sh := range s.Sheets {
		info.Sheets = append(info.Sheets, sheetToInfo(sh.Properties))
	}
	return info
}

func sheetToInfo(p *sheets.SheetProperties) SheetInfo {
	info := SheetInfo{
		ID:    p.SheetId,
		Title: p.Title,
		Index: p.Index,
		Type:  p.SheetType,
	}
	if p.GridProperties != nil {
		info.Rows = p.GridProperties.RowCount
		info.Cols = p.GridProperties.ColumnCount
	}
	return info
}
