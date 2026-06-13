package client

import (
	"net/http"
	"testing"
)

// Regression test: update center entries identify plugins via "name", not
// "shortName" as in the pluginManager API. Mapping them straight into Plugin
// left ShortName empty and HasUpdate false for every row.
func TestCheckPluginUpdates_MapsUpdateCenterFields(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pluginManager/checkUpdatesServer":
			w.WriteHeader(http.StatusOK)
		case "/updateCenter/api/json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"updates": [
				{"name": "git", "version": "5.7.0"},
				{"name": "workflow-aggregator", "version": "608.v67378e9d3db_1"}
			]}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer ts.Close()

	updates, err := c.CheckPluginUpdates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(updates))
	}
	if updates[0].ShortName != "git" {
		t.Errorf("expected ShortName git, got %q", updates[0].ShortName)
	}
	if updates[0].Version != "5.7.0" {
		t.Errorf("expected Version 5.7.0, got %q", updates[0].Version)
	}
	if !updates[0].HasUpdate {
		t.Error("expected HasUpdate to be true for update center entries")
	}
	if updates[1].ShortName != "workflow-aggregator" {
		t.Errorf("expected ShortName workflow-aggregator, got %q", updates[1].ShortName)
	}
}
