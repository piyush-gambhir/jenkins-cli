package client

import (
	"net/http"
	"testing"
)

func TestCreateView_RejectsUnsupportedType(t *testing.T) {
	var posted bool
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		posted = true
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	if err := c.CreateView("Mine", "hudson.model.MyView"); err == nil {
		t.Fatal("expected an error for an unsupported view type")
	}
	if posted {
		t.Error("expected no request for an unsupported view type")
	}

	if err := c.CreateView("Team", "hudson.model.ListView"); err != nil {
		t.Fatalf("unexpected error for ListView: %v", err)
	}
	if !posted {
		t.Error("expected the ListView to be created")
	}
}
