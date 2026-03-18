package client

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

// newTestServer creates an httptest.Server and a Client pointing to it.
// The crumb endpoint returns a 404 by default (CSRF disabled).
func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	mux := http.NewServeMux()
	// Default: CSRF disabled
	mux.HandleFunc("/crumbIssuer/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/", handler)
	ts := httptest.NewServer(mux)

	c := NewClient(config.Profile{
		URL:      ts.URL,
		Username: "admin",
		Token:    "secret-token",
	})
	// Point the client's httpClient at the test server (TLS not needed)
	c.httpClient = ts.Client()

	return ts, c
}

// newTestServerWithCrumb creates a test server that returns a crumb and delegates to handler.
func newTestServerWithCrumb(handler http.HandlerFunc) (*httptest.Server, *Client) {
	mux := http.NewServeMux()
	mux.HandleFunc("/crumbIssuer/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"crumbRequestField": "Jenkins-Crumb",
			"crumb":             "test-crumb-value",
		})
	})
	mux.HandleFunc("/", handler)
	ts := httptest.NewServer(mux)

	c := NewClient(config.Profile{
		URL:      ts.URL,
		Username: "admin",
		Token:    "secret-token",
	})
	c.httpClient = ts.Client()

	return ts, c
}

func TestNewClient_Creation(t *testing.T) {
	c := NewClient(config.Profile{
		URL:      "http://jenkins.example.com/",
		Username: "admin",
		Token:    "my-token",
		Insecure: true,
	})

	if c.baseURL != "http://jenkins.example.com" {
		t.Errorf("expected trailing slash stripped, got %q", c.baseURL)
	}
	if c.username != "admin" {
		t.Errorf("expected username 'admin', got %q", c.username)
	}
	if c.token != "my-token" {
		t.Errorf("expected token 'my-token', got %q", c.token)
	}
	if !c.insecure {
		t.Error("expected insecure=true")
	}
	if c.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
}

func TestBasicAuth_Header(t *testing.T) {
	var gotAuth string
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})
	defer ts.Close()

	_, err := c.Get("/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret-token"))
	if gotAuth != expected {
		t.Errorf("Authorization header = %q, want %q", gotAuth, expected)
	}
}

func TestGet_AppendsApiJson(t *testing.T) {
	var gotPath string
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})
	defer ts.Close()

	_, err := c.Get("/job/my-job", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/job/my-job/api/json" {
		t.Errorf("expected path /job/my-job/api/json, got %q", gotPath)
	}
}

func TestGet_WithTreeParam(t *testing.T) {
	var gotTree string
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotTree = r.URL.Query().Get("tree")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	})
	defer ts.Close()

	query := url.Values{"tree": {"jobs[name,url]"}}
	_, err := c.Get("/", query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotTree != "jobs[name,url]" {
		t.Errorf("tree param = %q, want %q", gotTree, "jobs[name,url]")
	}
}

func TestPost_IncludesCrumb(t *testing.T) {
	var gotCrumb string
	ts, c := newTestServerWithCrumb(func(w http.ResponseWriter, r *http.Request) {
		gotCrumb = r.Header.Get("Jenkins-Crumb")
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	_, err := c.Post("/job/my-job/build", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotCrumb != "test-crumb-value" {
		t.Errorf("Jenkins-Crumb header = %q, want %q", gotCrumb, "test-crumb-value")
	}
}

func TestErrorParsing_404(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})
	defer ts.Close()

	_, err := c.Get("/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestErrorParsing_403(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	})
	defer ts.Close()

	_, err := c.Get("/admin", nil)
	if err == nil {
		t.Fatal("expected error for 403")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
}

func TestErrorParsing_HTMLBody(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<html><body><h1>Something went wrong</h1><p>details here</p></body></html>`))
	})
	defer ts.Close()

	_, err := c.Get("/broken", nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	// parseErrorBody should extract the h1 content
	if !strings.Contains(apiErr.Message, "Something went wrong") {
		t.Errorf("expected message to contain 'Something went wrong', got %q", apiErr.Message)
	}
}
