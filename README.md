# gsheets

A CLI tool for the Google Sheets API. Reads and writes spreadsheet data, manages sheets/tabs, and supports JSON output for agent/pipeline use.

## Install

```bash
git clone https://github.com/the20100/gsheets-cli
cd gsheets-cli
go build -o gsheets .
mv gsheets /usr/local/bin/
```

## Authentication

`gsheets` uses Google's standard credential resolution:

1. `GOOGLE_APPLICATION_CREDENTIALS` env var (path to service account JSON) — checked first
2. `GSHEETS_CREDENTIALS` env var (path to service account JSON)
3. Config file — set with `gsheets auth set-credentials`
4. Application Default Credentials (ADC) — `gcloud auth application-default login`

### Service Account (recommended for automation)

1. Go to [Google Cloud Console → Service Accounts](https://console.cloud.google.com/iam-admin/serviceaccounts)
2. Create a service account and grant it access to your spreadsheets (share the sheet with the service account email)
3. Create a JSON key and download it
4. Configure:

```bash
# Option A: store path in config
gsheets auth set-credentials /path/to/sa.json

# Option B: env var
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json
```

### User Account (interactive)

```bash
gcloud auth application-default login
```

## Usage

```
gsheets [command] [flags]

Global flags:
  --json     Force JSON output
  --pretty   Force pretty-printed JSON (implies --json)
```

Output is **auto-detected**: JSON when stdout is piped, human-readable tables in a terminal.

---

## Commands

### auth

```bash
gsheets auth set-credentials /path/to/sa.json  # Store credentials path
gsheets auth status                             # Show auth status
gsheets auth logout                             # Remove stored path
```

### info

```bash
gsheets info   # Show binary path, config location, env vars
```

### spreadsheet

```bash
# Get spreadsheet metadata
gsheets spreadsheet get SPREADSHEET_ID
gsheets spreadsheet get SPREADSHEET_ID --json

# Create a new spreadsheet
gsheets spreadsheet create "My Budget 2026"
```

### sheet

```bash
# List all sheets/tabs
gsheets sheet list SPREADSHEET_ID

# Add a new sheet
gsheets sheet add SPREADSHEET_ID "Q1 Data"
gsheets sheet add SPREADSHEET_ID "Summary" --index 0   # insert at position 0

# Rename a sheet (use sheet ID from 'sheet list')
gsheets sheet rename SPREADSHEET_ID SHEET_ID "New Name"

# Delete a sheet
gsheets sheet delete SPREADSHEET_ID SHEET_ID
```

### values

```bash
# Read a range
gsheets values get SPREADSHEET_ID "Sheet1!A1:D10"
gsheets values get SPREADSHEET_ID "Sheet1!A1:D10" --header      # first row as headers
gsheets values get SPREADSHEET_ID "Sheet1!A1:D10" --render FORMULA

# Write values
gsheets values update SPREADSHEET_ID "Sheet1!A1" --row "Hello"
gsheets values update SPREADSHEET_ID "Sheet1!A1:C1" --row "Name,Age,City"
gsheets values update SPREADSHEET_ID "Sheet1!A2" \
  --row "Alice,30,New York" \
  --row "Bob,25,San Francisco"

# Append rows (after last populated row)
gsheets values append SPREADSHEET_ID "Sheet1!A:C" \
  --row "Alice,30,Engineer" \
  --row "Bob,25,Designer"

# Clear values (preserves formatting)
gsheets values clear SPREADSHEET_ID "Sheet1!A1:D10"
```

### schema

Dump machine-readable command schemas for agent introspection:

```bash
gsheets schema                      # all commands as JSON
gsheets schema values.update        # single command schema
gsheets schema spreadsheet.get
```

### update

Self-update the binary from GitHub:

```bash
gsheets update
```

---

## Tips

- **Finding the Spreadsheet ID**: extract from the URL — `https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/edit`
- **Finding Sheet IDs**: run `gsheets sheet list SPREADSHEET_ID` — the `SHEET ID` column shows numeric IDs
- **JSON output**: use `--json` or pipe to jq: `gsheets sheet list ID --json | jq '.[].title'`
- **Formulas**: use `--input USER_ENTERED` (default) so `=SUM(A1:C1)` is evaluated
- **Raw values**: use `--input RAW` to store values without Google interpreting them
- **Ranges**: follow A1 notation — `Sheet1!A1:C10`, `Sheet1!A:A`, `Sheet1!1:5`, or just `A1:B2` for the first sheet
