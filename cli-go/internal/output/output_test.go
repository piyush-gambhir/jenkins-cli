package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{Writer: &buf}

	data := map[string]interface{}{
		"name":   "test-job",
		"status": "SUCCESS",
		"number": 42,
	}

	err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}

	if parsed["name"] != "test-job" {
		t.Errorf("expected name 'test-job', got %v", parsed["name"])
	}
	if parsed["status"] != "SUCCESS" {
		t.Errorf("expected status 'SUCCESS', got %v", parsed["status"])
	}
	// JSON numbers are float64 by default
	if parsed["number"] != float64(42) {
		t.Errorf("expected number 42, got %v", parsed["number"])
	}
}

func TestYAMLFormat(t *testing.T) {
	var buf bytes.Buffer
	f := &YAMLFormatter{Writer: &buf}

	data := map[string]interface{}{
		"name":   "test-job",
		"status": "SUCCESS",
	}

	err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify it's valid YAML
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput: %s", err, output)
	}

	if parsed["name"] != "test-job" {
		t.Errorf("expected name 'test-job', got %v", parsed["name"])
	}
	if parsed["status"] != "SUCCESS" {
		t.Errorf("expected status 'SUCCESS', got %v", parsed["status"])
	}
}

func TestTableFormat(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{Writer: &buf}

	type item struct {
		Name   string
		Status string
	}
	data := []item{
		{Name: "job1", Status: "SUCCESS"},
		{Name: "job2", Status: "FAILURE"},
	}

	def := &TableDef{
		Headers: []string{"NAME", "STATUS"},
		RowFunc: func(i interface{}) []string {
			it := i.(item)
			return []string{it.Name, it.Status}
		},
	}

	err := f.FormatTable(data, def)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that headers are present
	if !strings.Contains(output, "NAME") {
		t.Errorf("expected output to contain 'NAME', got:\n%s", output)
	}
	if !strings.Contains(output, "STATUS") {
		t.Errorf("expected output to contain 'STATUS', got:\n%s", output)
	}
	// Check that row data is present
	if !strings.Contains(output, "job1") {
		t.Errorf("expected output to contain 'job1', got:\n%s", output)
	}
	if !strings.Contains(output, "FAILURE") {
		t.Errorf("expected output to contain 'FAILURE', got:\n%s", output)
	}
}

func TestNewFormatter_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatJSON, &buf)

	_, ok := f.(*JSONFormatter)
	if !ok {
		t.Errorf("expected *JSONFormatter, got %T", f)
	}
}

func TestNewFormatter_Default_Table(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, &buf)

	_, ok := f.(*TableFormatter)
	if !ok {
		t.Errorf("expected *TableFormatter, got %T", f)
	}

	// Also test with empty format (should default to table)
	f2 := NewFormatter("", &buf)
	_, ok2 := f2.(*TableFormatter)
	if !ok2 {
		t.Errorf("expected *TableFormatter for empty format, got %T", f2)
	}
}
