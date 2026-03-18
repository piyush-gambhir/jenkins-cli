package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

// newJobTestServer creates a test server with crumb disabled and a custom handler.
func newJobTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	mux := http.NewServeMux()
	mux.HandleFunc("/crumbIssuer/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/", handler)
	ts := httptest.NewServer(mux)

	c := NewClient(config.Profile{URL: ts.URL, Username: "admin", Token: "tok"})
	c.httpClient = ts.Client()
	return ts, c
}

func TestListJobs(t *testing.T) {
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/json" {
			t.Errorf("expected path /api/json, got %q", r.URL.Path)
		}
		tree := r.URL.Query().Get("tree")
		if tree == "" {
			t.Error("expected tree query param to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(JobListResponse{
			Jobs: []Job{
				{Name: "job1", Color: "blue"},
				{Name: "job2", Color: "red"},
			},
		})
	})
	defer ts.Close()

	jobs, err := c.ListJobs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if jobs[0].Name != "job1" {
		t.Errorf("expected first job name 'job1', got %q", jobs[0].Name)
	}
}

func TestListJobsInFolder(t *testing.T) {
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		// folder "my-folder" should become /job/my-folder/api/json
		if r.URL.Path != "/job/my-folder/api/json" {
			t.Errorf("expected path /job/my-folder/api/json, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(JobListResponse{
			Jobs: []Job{{Name: "inner-job", Color: "blue"}},
		})
	})
	defer ts.Close()

	jobs, err := c.ListJobs("my-folder")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Name != "inner-job" {
		t.Errorf("expected job name 'inner-job', got %q", jobs[0].Name)
	}
}

func TestGetJob(t *testing.T) {
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		// "my-folder/my-job" -> /job/my-folder/job/my-job/api/json
		if r.URL.Path != "/job/my-folder/job/my-job/api/json" {
			t.Errorf("expected path /job/my-folder/job/my-job/api/json, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Job{
			Name:     "my-job",
			FullName: "my-folder/my-job",
			Color:    "blue",
		})
	})
	defer ts.Close()

	job, err := c.GetJob("my-folder/my-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Name != "my-job" {
		t.Errorf("expected name 'my-job', got %q", job.Name)
	}
	if job.FullName != "my-folder/my-job" {
		t.Errorf("expected fullName 'my-folder/my-job', got %q", job.FullName)
	}
}

func TestGetJobConfig(t *testing.T) {
	expectedXML := `<?xml version="1.0"?><project><description>test</description></project>`
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		// GetJobConfig uses GetRaw, path should be /job/my-job/config.xml
		if r.URL.Path != "/job/my-job/config.xml" {
			t.Errorf("expected path /job/my-job/config.xml, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(expectedXML))
	})
	defer ts.Close()

	xmlStr, err := c.GetJobConfig("my-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if xmlStr != expectedXML {
		t.Errorf("expected XML %q, got %q", expectedXML, xmlStr)
	}
}

func TestCreateJob(t *testing.T) {
	configXML := `<project><description>new job</description></project>`
	var gotPath, gotName, gotContentType string
	var gotBody string

	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotName = r.URL.Query().Get("name")
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.CreateJob("my-folder", "new-job", configXML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/job/my-folder/createItem" {
		t.Errorf("expected path /job/my-folder/createItem, got %q", gotPath)
	}
	if gotName != "new-job" {
		t.Errorf("expected name query param 'new-job', got %q", gotName)
	}
	if gotContentType != "application/xml" {
		t.Errorf("expected Content-Type application/xml, got %q", gotContentType)
	}
	if gotBody != configXML {
		t.Errorf("expected body %q, got %q", configXML, gotBody)
	}
}

func TestTriggerBuild(t *testing.T) {
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/my-job/build" {
			t.Errorf("expected path /job/my-job/build, got %q", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Location", "http://jenkins/queue/item/42/")
		w.WriteHeader(http.StatusCreated)
	})
	defer ts.Close()

	ql, err := c.TriggerBuild("my-job", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ql.QueueURL != "http://jenkins/queue/item/42/" {
		t.Errorf("expected queue URL 'http://jenkins/queue/item/42/', got %q", ql.QueueURL)
	}
}

func TestTriggerParameterizedBuild(t *testing.T) {
	var gotPath string
	var gotBranch string
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBranch = r.URL.Query().Get("BRANCH")
		w.Header().Set("Location", "http://jenkins/queue/item/99/")
		w.WriteHeader(http.StatusCreated)
	})
	defer ts.Close()

	params := map[string]string{"BRANCH": "main"}
	ql, err := c.TriggerBuild("my-job", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/job/my-job/buildWithParameters" {
		t.Errorf("expected path /job/my-job/buildWithParameters, got %q", gotPath)
	}
	if gotBranch != "main" {
		t.Errorf("expected BRANCH=main, got %q", gotBranch)
	}
	if !strings.Contains(ql.QueueURL, "queue/item/99") {
		t.Errorf("expected queue URL containing 'queue/item/99', got %q", ql.QueueURL)
	}
}

func TestDeleteJob(t *testing.T) {
	var gotPath, gotMethod string
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.DeleteJob("my-folder/my-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/job/my-folder/job/my-job/doDelete" {
		t.Errorf("expected path /job/my-folder/job/my-job/doDelete, got %q", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
}

func TestEnableJob(t *testing.T) {
	var gotPath string
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.EnableJob("my-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/job/my-job/enable" {
		t.Errorf("expected path /job/my-job/enable, got %q", gotPath)
	}
}

func TestDisableJob(t *testing.T) {
	var gotPath string
	ts, c := newJobTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.DisableJob("my-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/job/my-job/disable" {
		t.Errorf("expected path /job/my-job/disable, got %q", gotPath)
	}
}
