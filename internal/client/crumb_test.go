package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

func TestFetchCrumb_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"crumbRequestField": "Jenkins-Crumb",
				"crumb":             "abc123",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(config.Profile{URL: ts.URL, Username: "u", Token: "t"})
	c.httpClient = ts.Client()

	crumb, err := c.fetchCrumb()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if crumb == nil {
		t.Fatal("expected crumb, got nil")
	}
	if crumb.RequestField != "Jenkins-Crumb" {
		t.Errorf("RequestField = %q, want %q", crumb.RequestField, "Jenkins-Crumb")
	}
	if crumb.Value != "abc123" {
		t.Errorf("Value = %q, want %q", crumb.Value, "abc123")
	}
}

func TestFetchCrumb_Disabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(config.Profile{URL: ts.URL, Username: "u", Token: "t"})
	c.httpClient = ts.Client()

	crumb, err := c.fetchCrumb()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if crumb != nil {
		t.Errorf("expected nil crumb when CSRF disabled, got %+v", crumb)
	}
}

func TestCrumbCache_NotExpired(t *testing.T) {
	var crumbHits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			atomic.AddInt32(&crumbHits, 1)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"crumbRequestField": "Jenkins-Crumb",
				"crumb":             "cached-crumb",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(config.Profile{URL: ts.URL, Username: "u", Token: "t"})
	c.httpClient = ts.Client()

	// First call fetches the crumb
	crumb1, err := c.ensureCrumb()
	if err != nil {
		t.Fatalf("first ensureCrumb: %v", err)
	}

	// Second call should use cache
	crumb2, err := c.ensureCrumb()
	if err != nil {
		t.Fatalf("second ensureCrumb: %v", err)
	}

	if crumb1.Value != crumb2.Value {
		t.Errorf("cached crumb mismatch: %q vs %q", crumb1.Value, crumb2.Value)
	}

	hits := atomic.LoadInt32(&crumbHits)
	if hits != 1 {
		t.Errorf("expected crumb endpoint hit once, got %d", hits)
	}
}

func TestCrumbInjection_OnPost(t *testing.T) {
	var gotCrumbHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"crumbRequestField": "Jenkins-Crumb",
				"crumb":             "injected-crumb",
			})
			return
		}
		gotCrumbHeader = r.Header.Get("Jenkins-Crumb")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(config.Profile{URL: ts.URL, Username: "u", Token: "t"})
	c.httpClient = ts.Client()

	_, err := c.Post("/some/action", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotCrumbHeader != "injected-crumb" {
		t.Errorf("expected Jenkins-Crumb header 'injected-crumb' on POST, got %q", gotCrumbHeader)
	}
}

func TestCrumbInjection_NotOnGet(t *testing.T) {
	var gotCrumbHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"crumbRequestField": "Jenkins-Crumb",
				"crumb":             "should-not-appear",
			})
			return
		}
		gotCrumbHeader = r.Header.Get("Jenkins-Crumb")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer ts.Close()

	c := NewClient(config.Profile{URL: ts.URL, Username: "u", Token: "t"})
	c.httpClient = ts.Client()

	_, err := c.Get("/some/path", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotCrumbHeader != "" {
		t.Errorf("expected no Jenkins-Crumb header on GET, got %q", gotCrumbHeader)
	}
}
