package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestWriteError_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("something went wrong")

	WriteError(&buf, FormatJSON, err, 500)

	output := buf.String()

	// Verify it's valid JSON
	var parsed ErrorResponse
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	if parsed.Error != "something went wrong" {
		t.Errorf("expected error 'something went wrong', got %q", parsed.Error)
	}
	if parsed.StatusCode != 500 {
		t.Errorf("expected status_code 500, got %d", parsed.StatusCode)
	}
}

func TestWriteError_JSON_NoStatusCode(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("validation failed")

	WriteError(&buf, FormatJSON, err, 0)

	output := buf.String()

	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(output), &parsed); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", jsonErr, output)
	}

	if parsed["error"] != "validation failed" {
		t.Errorf("expected error 'validation failed', got %v", parsed["error"])
	}

	// status_code should be omitted when 0
	if _, exists := parsed["status_code"]; exists {
		t.Error("expected status_code to be omitted when 0")
	}
}

func TestWriteError_Table(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("something went wrong")

	WriteError(&buf, FormatTable, err, 500)

	output := buf.String()

	if !strings.Contains(output, "Error: something went wrong") {
		t.Errorf("expected plain text error, got: %s", output)
	}

	// Should NOT be JSON
	var parsed map[string]interface{}
	if json.Unmarshal([]byte(output), &parsed) == nil {
		t.Error("expected non-JSON output for table format")
	}
}

func TestWriteError_YAML(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("something went wrong")

	WriteError(&buf, FormatYAML, err, 404)

	output := buf.String()

	// YAML format should fall through to plain text
	if !strings.Contains(output, "Error: something went wrong") {
		t.Errorf("expected plain text error for YAML format, got: %s", output)
	}
}
