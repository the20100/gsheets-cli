package validate

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// ResourceID validates a resource identifier against common agent hallucinations.
// Rejects path traversals, embedded query params, control characters, and double-encoding.
func ResourceID(id string) error {
	if id == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}

	// Reject control characters (bytes below ASCII 0x20)
	for _, r := range id {
		if r < 0x20 {
			return fmt.Errorf("resource ID contains control character at position %d", r)
		}
	}

	// Reject path traversal
	cleaned := filepath.Clean(id)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("resource ID contains path traversal: %q", id)
	}

	// Reject embedded query params
	if strings.ContainsAny(id, "?#") {
		return fmt.Errorf("resource ID contains query characters: %q", id)
	}

	// Reject percent-encoding (double-encoding defense — encode at HTTP layer only)
	if strings.Contains(id, "%") {
		return fmt.Errorf("resource ID contains percent-encoding: %q (do not pre-encode)", id)
	}

	return nil
}

// SafeString validates a string value for common injection patterns.
func SafeString(s string, maxLen int) error {
	if len([]rune(s)) > maxLen {
		return fmt.Errorf("string exceeds maximum length of %d characters", maxLen)
	}
	for i, r := range s {
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("string contains control character at position %d", i)
		}
	}
	return nil
}

// OutputPath validates and sandboxes an output file path to the current working directory.
func OutputPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid output path: %w", err)
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("cannot determine CWD: %w", err)
	}

	if !strings.HasPrefix(abs, cwd) {
		return "", fmt.Errorf("output path %q escapes current directory", path)
	}

	return abs, nil
}

// JSONPayload does basic sanity checks on a raw JSON payload string.
func JSONPayload(s string) error {
	for i, r := range s {
		if !unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
			return fmt.Errorf("JSON payload contains non-printable character at position %d", i)
		}
	}
	return nil
}
