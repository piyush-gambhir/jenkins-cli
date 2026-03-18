package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

func newNodeTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
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

func TestListNodes(t *testing.T) {
	ts, c := newNodeTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/computer/api/json" {
			t.Errorf("expected path /computer/api/json, got %q", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ComputerResponse{
			Computers: []Node{
				{DisplayName: "built-in", NumExecutors: 2, Offline: false},
				{DisplayName: "agent-1", NumExecutors: 4, Offline: true},
			},
			TotalExecutors: 6,
			BusyExecutors:  1,
		})
	})
	defer ts.Close()

	nodes, err := c.ListNodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].DisplayName != "built-in" {
		t.Errorf("expected first node 'built-in', got %q", nodes[0].DisplayName)
	}
	if nodes[1].Offline != true {
		t.Error("expected second node to be offline")
	}
}

func TestGetNode(t *testing.T) {
	ts, c := newNodeTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/computer/agent-1/api/json" {
			t.Errorf("expected path /computer/agent-1/api/json, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Node{
			DisplayName:  "agent-1",
			NumExecutors: 4,
			Offline:      false,
			Idle:         true,
		})
	})
	defer ts.Close()

	node, err := c.GetNode("agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node.DisplayName != "agent-1" {
		t.Errorf("expected displayName 'agent-1', got %q", node.DisplayName)
	}
	if node.NumExecutors != 4 {
		t.Errorf("expected 4 executors, got %d", node.NumExecutors)
	}
	if !node.Idle {
		t.Error("expected node to be idle")
	}
}

func TestToggleOffline(t *testing.T) {
	var gotPath, gotMethod, gotMessage string
	ts, c := newNodeTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotMessage = r.URL.Query().Get("offlineMessage")
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.ToggleOffline("agent-1", true, "Maintenance window")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/computer/agent-1/toggleOffline" {
		t.Errorf("expected path /computer/agent-1/toggleOffline, got %q", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotMessage != "Maintenance window" {
		t.Errorf("expected offlineMessage 'Maintenance window', got %q", gotMessage)
	}
}

func TestDeleteNode(t *testing.T) {
	var gotPath, gotMethod string
	ts, c := newNodeTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.DeleteNode("agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/computer/agent-1/doDelete" {
		t.Errorf("expected path /computer/agent-1/doDelete, got %q", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
}
