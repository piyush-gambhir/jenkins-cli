package client

import (
	"net/http"
	"testing"
)

func TestCheckPluginUpdates_FiltersInstalledPluginsWithUpdates(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pluginManager/checkUpdatesServer":
			w.WriteHeader(http.StatusOK)
		case "/pluginManager/api/json":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"plugins": [
				{"shortName": "git", "version": "5.6.0", "hasUpdate": true},
				{"shortName": "credentials", "version": "1400.v1", "hasUpdate": false},
				{"shortName": "workflow-aggregator", "version": "600.v1", "hasUpdate": true}
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
	if updates[0].ShortName != "git" || updates[0].Version != "5.6.0" {
		t.Errorf("expected git 5.6.0, got %s %s", updates[0].ShortName, updates[0].Version)
	}
	if updates[1].ShortName != "workflow-aggregator" {
		t.Errorf("expected workflow-aggregator, got %q", updates[1].ShortName)
	}
}
