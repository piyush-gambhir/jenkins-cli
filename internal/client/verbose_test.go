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
