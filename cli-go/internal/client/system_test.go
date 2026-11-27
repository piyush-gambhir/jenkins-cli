package client

import (
	"net/http"
	"testing"
)

func TestRestart_TreatsRedirectToRestartingPageAsSuccess(t *testing.T) {
	for _, tc := range []struct {
		path    string
		restart func(*Client) error
	}{
		{path: "/restart", restart: (*Client).Restart},
		{path: "/safeRestart", restart: (*Client).SafeRestart},
	} {
		ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == tc.path {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		})

		if err := tc.restart(c); err != nil {
			t.Errorf("%s: expected success, got %v", tc.path, err)
		}
		ts.Close()
	}
}
