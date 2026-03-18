package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

func newQueueTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
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

func TestListQueue(t *testing.T) {
	ts, c := newQueueTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/queue/api/json" {
			t.Errorf("expected path /queue/api/json, got %q", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(QueueResponse{
			Items: []QueueItem{
				{ID: 1, Why: "Waiting for executor", Task: QueueTask{Name: "job1"}},
				{ID: 2, Why: "In the quiet period", Task: QueueTask{Name: "job2"}},
			},
		})
	})
	defer ts.Close()

	items, err := c.ListQueue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 queue items, got %d", len(items))
	}
	if items[0].ID != 1 {
		t.Errorf("expected first item ID=1, got %d", items[0].ID)
	}
	if items[0].Task.Name != "job1" {
		t.Errorf("expected first item task name 'job1', got %q", items[0].Task.Name)
	}
	if items[1].Why != "In the quiet period" {
		t.Errorf("expected second item Why 'In the quiet period', got %q", items[1].Why)
	}
}

func TestCancelQueueItem(t *testing.T) {
	var gotPath, gotMethod, gotID string
	ts, c := newQueueTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotID = r.URL.Query().Get("id")
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.CancelQueueItem(123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/queue/cancelItem" {
		t.Errorf("expected path /queue/cancelItem, got %q", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotID != "123" {
		t.Errorf("expected id=123, got %q", gotID)
	}
}

func TestGetQueueItem(t *testing.T) {
	ts, c := newQueueTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/queue/item/456/api/json" {
			t.Errorf("expected path /queue/item/456/api/json, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(QueueItem{
			ID:      456,
			Why:     "Waiting for executor",
			Blocked: true,
			Task:    QueueTask{Name: "deploy-pipeline"},
			Executable: QueueExecutable{
				Number: 10,
				URL:    "http://jenkins/job/deploy-pipeline/10/",
			},
		})
	})
	defer ts.Close()

	item, err := c.GetQueueItem(456)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != 456 {
		t.Errorf("expected ID=456, got %d", item.ID)
	}
	if item.Task.Name != "deploy-pipeline" {
		t.Errorf("expected task name 'deploy-pipeline', got %q", item.Task.Name)
	}
	if item.Executable.Number != 10 {
		t.Errorf("expected executable number 10, got %d", item.Executable.Number)
	}
	if !item.Blocked {
		t.Error("expected item to be blocked")
	}
}
