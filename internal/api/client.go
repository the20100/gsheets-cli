package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiBase = "https://sheets.googleapis.com/v4/spreadsheets"

// RefreshFunc is called when the access token is expired.
// It returns the new access token and its Unix expiry timestamp.
type RefreshFunc func() (newToken string, expiresAt int64, err error)

// Client is an authenticated Google Sheets API v4 client.
type Client struct {
	token       string
	tokenExpiry int64
	refreshFn   RefreshFunc
	httpClient  *http.Client
}

// NewClient creates an authenticated Client.
// refreshFn may be nil if no token refresh is needed.
func NewClient(token string, tokenExpiry int64, refreshFn RefreshFunc) *Client {
	return &Client{
		token:       token,
		tokenExpiry: tokenExpiry,
		refreshFn:   refreshFn,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// ensureToken refreshes the access token if it expires within 60 seconds.
func (c *Client) ensureToken() error {
	if c.refreshFn == nil {
		return nil
	}
	if c.tokenExpiry > 0 && time.Now().Unix() < c.tokenExpiry-60 {
		return nil
	}
	newToken, expiresAt, err := c.refreshFn()
	if err != nil {
		return fmt.Errorf("refreshing token: %w", err)
	}
	c.token = newToken
	c.tokenExpiry = expiresAt
	return nil
}

func (c *Client) do(method, rawURL string, params url.Values, body any) ([]byte, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}
	if params != nil {
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		u.RawQuery = params.Encode()
		rawURL = u.String()
	}
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if jerr := json.Unmarshal(respBody, &errResp); jerr == nil && errResp.Error.Message != "" {
			return nil, &SheetsError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API error %d: %s", errResp.Error.Code, errResp.Error.Message),
			}
		}
		return nil, &SheetsError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		}
	}
	return respBody, nil
}

// ---- Spreadsheet methods ----

func (c *Client) GetSpreadsheet(id string) (*SpreadsheetInfo, error) {
	u := apiBase + "/" + url.PathEscape(id)
	body, err := c.do("GET", u, nil, nil)
	if err != nil {
		return nil, err
	}
	return parseSpreadsheet(body)
}

func (c *Client) CreateSpreadsheet(title string) (*SpreadsheetInfo, error) {
	payload := map[string]interface{}{
		"properties": map[string]string{"title": title},
	}
	body, err := c.do("POST", apiBase, nil, payload)
	if err != nil {
		return nil, err
	}
	return parseSpreadsheet(body)
}

// ---- Sheet methods ----

func (c *Client) ListSheets(spreadsheetID string) ([]SheetInfo, error) {
	info, err := c.GetSpreadsheet(spreadsheetID)
	if err != nil {
		return nil, err
	}
	return info.Sheets, nil
}

func (c *Client) AddSheet(spreadsheetID, title string, index int) (*SheetInfo, error) {
	props := map[string]interface{}{"title": title}
	if index >= 0 {
		props["index"] = index
	}
	payload := map[string]interface{}{
		"requests": []map[string]interface{}{
			{"addSheet": map[string]interface{}{"properties": props}},
		},
	}
	u := apiBase + "/" + url.PathEscape(spreadsheetID) + ":batchUpdate"
	body, err := c.do("POST", u, nil, payload)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Replies []struct {
			AddSheet *struct {
				Properties sheetProperties `json:"properties"`
			} `json:"addSheet"`
		} `json:"replies"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	for _, reply := range resp.Replies {
		if reply.AddSheet != nil {
			si := propsToSheetInfo(reply.AddSheet.Properties)
			return &si, nil
		}
	}
	return nil, fmt.Errorf("unexpected response: no addSheet reply")
}

func (c *Client) DeleteSheet(spreadsheetID string, sheetID int64) error {
	payload := map[string]interface{}{
		"requests": []map[string]interface{}{
			{"deleteSheet": map[string]interface{}{"sheetId": sheetID}},
		},
	}
	u := apiBase + "/" + url.PathEscape(spreadsheetID) + ":batchUpdate"
	_, err := c.do("POST", u, nil, payload)
	return err
}

func (c *Client) RenameSheet(spreadsheetID string, sheetID int64, newTitle string) error {
	payload := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"updateSheetProperties": map[string]interface{}{
					"properties": map[string]interface{}{
						"sheetId": sheetID,
						"title":   newTitle,
					},
					"fields": "title",
				},
			},
		},
	}
	u := apiBase + "/" + url.PathEscape(spreadsheetID) + ":batchUpdate"
	_, err := c.do("POST", u, nil, payload)
	return err
}

// ---- Values methods ----

func (c *Client) GetValues(spreadsheetID, rangeStr, renderOption, dimension string) (*ValuesResult, error) {
	u := apiBase + "/" + url.PathEscape(spreadsheetID) + "/values/" + url.PathEscape(rangeStr)
	params := url.Values{}
	if renderOption != "" {
		params.Set("valueRenderOption", renderOption)
	}
	if dimension != "" {
		params.Set("majorDimension", dimension)
	}
	body, err := c.do("GET", u, params, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Range          string          `json:"range"`
		MajorDimension string          `json:"majorDimension"`
		Values         [][]interface{} `json:"values"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &ValuesResult{
		Range:     resp.Range,
		Dimension: resp.MajorDimension,
		Values:    resp.Values,
	}, nil
}

func (c *Client) UpdateValues(spreadsheetID, rangeStr string, values [][]interface{}, inputOption string) (*UpdateResult, error) {
	if inputOption == "" {
		inputOption = "USER_ENTERED"
	}
	u := apiBase + "/" + url.PathEscape(spreadsheetID) + "/values/" + url.PathEscape(rangeStr)
	params := url.Values{}
	params.Set("valueInputOption", inputOption)
	payload := map[string]interface{}{"values": values}
	body, err := c.do("PUT", u, params, payload)
	if err != nil {
		return nil, err
	}
	var resp struct {
		SpreadsheetID  string `json:"spreadsheetId"`
		UpdatedRange   string `json:"updatedRange"`
		UpdatedRows    int64  `json:"updatedRows"`
		UpdatedColumns int64  `json:"updatedColumns"`
		UpdatedCells   int64  `json:"updatedCells"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &UpdateResult{
		SpreadsheetID:  resp.SpreadsheetID,
		UpdatedRange:   resp.UpdatedRange,
		UpdatedRows:    resp.UpdatedRows,
		UpdatedColumns: resp.UpdatedColumns,
		UpdatedCells:   resp.UpdatedCells,
	}, nil
}

func (c *Client) AppendValues(spreadsheetID, rangeStr string, values [][]interface{}, inputOption, insertOption string) (*UpdateResult, error) {
	if inputOption == "" {
		inputOption = "USER_ENTERED"
	}
	if insertOption == "" {
		insertOption = "INSERT_ROWS"
	}
	u := apiBase + "/" + url.PathEscape(spreadsheetID) + "/values/" + url.PathEscape(rangeStr) + ":append"
	params := url.Values{}
	params.Set("valueInputOption", inputOption)
	params.Set("insertDataOption", insertOption)
	payload := map[string]interface{}{"values": values}
	body, err := c.do("POST", u, params, payload)
	if err != nil {
		return nil, err
	}
	var resp struct {
		SpreadsheetID string `json:"spreadsheetId"`
		Updates       *struct {
			UpdatedRange   string `json:"updatedRange"`
			UpdatedRows    int64  `json:"updatedRows"`
			UpdatedColumns int64  `json:"updatedColumns"`
			UpdatedCells   int64  `json:"updatedCells"`
		} `json:"updates"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	result := &UpdateResult{SpreadsheetID: resp.SpreadsheetID}
	if resp.Updates != nil {
		result.UpdatedRange = resp.Updates.UpdatedRange
		result.UpdatedRows = resp.Updates.UpdatedRows
		result.UpdatedColumns = resp.Updates.UpdatedColumns
		result.UpdatedCells = resp.Updates.UpdatedCells
	}
	return result, nil
}

func (c *Client) ClearValues(spreadsheetID, rangeStr string) (string, error) {
	u := apiBase + "/" + url.PathEscape(spreadsheetID) + "/values/" + url.PathEscape(rangeStr) + ":clear"
	body, err := c.do("POST", u, nil, map[string]interface{}{})
	if err != nil {
		return "", err
	}
	var resp struct {
		ClearedRange string `json:"clearedRange"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	return resp.ClearedRange, nil
}

// ---- helpers ----

type sheetProperties struct {
	SheetID        int64  `json:"sheetId"`
	Title          string `json:"title"`
	Index          int64  `json:"index"`
	SheetType      string `json:"sheetType"`
	GridProperties *struct {
		RowCount    int64 `json:"rowCount"`
		ColumnCount int64 `json:"columnCount"`
	} `json:"gridProperties"`
}

func propsToSheetInfo(p sheetProperties) SheetInfo {
	si := SheetInfo{
		ID:    p.SheetID,
		Title: p.Title,
		Index: p.Index,
		Type:  p.SheetType,
	}
	if p.GridProperties != nil {
		si.Rows = p.GridProperties.RowCount
		si.Cols = p.GridProperties.ColumnCount
	}
	return si
}

func parseSpreadsheet(data []byte) (*SpreadsheetInfo, error) {
	var resp struct {
		SpreadsheetID  string `json:"spreadsheetId"`
		SpreadsheetURL string `json:"spreadsheetUrl"`
		Properties     struct {
			Title string `json:"title"`
		} `json:"properties"`
		Sheets []struct {
			Properties sheetProperties `json:"properties"`
		} `json:"sheets"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	info := &SpreadsheetInfo{
		ID:    resp.SpreadsheetID,
		Title: resp.Properties.Title,
		URL:   resp.SpreadsheetURL,
	}
	for _, sh := range resp.Sheets {
		info.Sheets = append(info.Sheets, propsToSheetInfo(sh.Properties))
	}
	return info, nil
}
