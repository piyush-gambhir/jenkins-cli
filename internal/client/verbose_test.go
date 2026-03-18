package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

func TestVerboseTransport_LogsRequestAndResponse(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	profile := config.Profile{
		URL:      ts.URL,
		Username: "test",
		Token:    "test-token",
	}

	c := NewClient(profile, true)

	// Make a request
	req, err := http.NewRequest("GET", ts.URL+"/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Restore stderr and read captured output
	w.Close()
	os.Stderr = oldStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "--> GET") {
		t.Errorf("expected verbose output to contain '--> GET', got: %s", output)
	}
	if !strings.Contains(output, "<-- 200") {
		t.Errorf("expected verbose output to contain '<-- 200', got: %s", output)
	}
}

func TestNewClient_WithoutVerbose(t *testing.T) {
	profile := config.Profile{
		URL:      "http://localhost:8080",
		Username: "test",
		Token:    "test-token",
	}

	// Should not panic with no verbose arg
	c := NewClient(profile)
	if c == nil {
		t.Fatal("expected non-nil client")
	}

	// Should not panic with verbose=false
	c2 := NewClient(profile, false)
	if c2 == nil {
		t.Fatal("expected non-nil client with verbose=false")
	}
}

// TestVerboseTransport_RedactsAuthHeaders verifies that Authorization and Cookie
// headers are redacted in verbose output, while other headers like Jenkins-Crumb
// remain visible.
func TestVerboseTransport_RedactsAuthHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	profile := config.Profile{
		URL:      ts.URL,
		Username: "test",
		Token:    "super-secret-token",
	}
	c := NewClient(profile, true)

	req, err := http.NewRequest("GET", ts.URL+"/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.SetBasicAuth("test", "super-secret-token")
	req.Header.Set("Jenkins-Crumb", "crumb-value-123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session=abc123")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	w.Close()
	os.Stderr = oldStderr

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Authorization header should be redacted
	if strings.Contains(output, "super-secret-token") {
		t.Error("expected Authorization header value to be redacted, but token is visible in output")
	}
	if !strings.Contains(output, "Authorization: [REDACTED]") {
		t.Errorf("expected 'Authorization: [REDACTED]' in output, got:\n%s", output)
	}

	// Cookie header should be redacted
	if strings.Contains(output, "abc123") {
		t.Error("expected Cookie header value to be redacted, but session value is visible")
	}
	if !strings.Contains(output, "Cookie: [REDACTED]") {
		t.Errorf("expected 'Cookie: [REDACTED]' in output, got:\n%s", output)
	}

	// Jenkins-Crumb should remain visible
	if !strings.Contains(output, "crumb-value-123") {
		t.Errorf("expected Jenkins-Crumb to remain visible, got:\n%s", output)
	}

	// Content-Type should remain visible
	if !strings.Contains(output, "application/json") {
		t.Errorf("expected Content-Type to remain visible, got:\n%s", output)
	}
}

// TestRedactAuthHeaders_Unit tests the redactAuthHeaders function directly.
func TestRedactAuthHeaders_Unit(t *testing.T) {
	headers := http.Header{
		"Authorization": {"Basic dGVzdDp0b2tlbg=="},
		"Cookie":        {"session=abc123; JSESSIONID=xyz"},
		"Content-Type":  {"application/json"},
		"Jenkins-Crumb": {"crumb-123"},
		"X-Request-Id":  {"req-456"},
	}

	redacted := redactAuthHeaders(headers)

	// Auth headers should be redacted
	if v := redacted.Get("Authorization"); v != "[REDACTED]" {
		t.Errorf("expected Authorization to be [REDACTED], got %q", v)
	}
	if v := redacted.Get("Cookie"); v != "[REDACTED]" {
		t.Errorf("expected Cookie to be [REDACTED], got %q", v)
	}

	// Non-auth headers should remain
	if v := redacted.Get("Content-Type"); v != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", v)
	}
	if v := redacted.Get("Jenkins-Crumb"); v != "crumb-123" {
		t.Errorf("expected Jenkins-Crumb 'crumb-123', got %q", v)
	}
	if v := redacted.Get("X-Request-Id"); v != "req-456" {
		t.Errorf("expected X-Request-Id 'req-456', got %q", v)
	}
}
