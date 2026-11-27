package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/output"
)

func useTestJenkins(t *testing.T, handler http.HandlerFunc, format output.Format) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler(w, r)
	}))
	origClient, origFormat, origQuiet := jenkinsClient, outFormat, quietFlag
	t.Cleanup(func() {
		ts.Close()
		jenkinsClient, outFormat, quietFlag = origClient, origFormat, origQuiet
	})
	jenkinsClient = client.NewClient(config.Profile{URL: ts.URL, Username: "admin", Token: "tok"})
	outFormat = format
	quietFlag = true
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan []byte)
	go func() {
		b, _ := io.ReadAll(r)
		done <- b
	}()
	runErr := fn()
	os.Stdout = orig
	w.Close()
	return string(<-done), runErr
}

func TestQueueList_EmptyJSONIsValidJSON(t *testing.T) {
	useTestJenkins(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"items": []}`))
	}, output.FormatJSON)

	cmd := newQueueListCmd()
	out, err := captureStdout(t, func() error { return cmd.RunE(cmd, nil) })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var items []interface{}
	if err := json.Unmarshal([]byte(out), &items); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, out)
	}
}

func TestPluginCheckUpdates_JSONIsValidJSON(t *testing.T) {
	for _, plugins := range []string{
		`{"plugins": [{"shortName": "git", "version": "5.6.0", "hasUpdate": true}]}`,
		`{"plugins": [{"shortName": "git", "version": "5.6.0", "hasUpdate": false}]}`,
	} {
		useTestJenkins(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/pluginManager/api/json" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(plugins))
			}
		}, output.FormatJSON)

		cmd := newPluginCheckUpdatesCmd()
		out, err := captureStdout(t, func() error { return cmd.RunE(cmd, nil) })
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var updates []interface{}
		if err := json.Unmarshal([]byte(out), &updates); err != nil {
			t.Fatalf("output is not valid JSON: %v\nraw: %s", err, out)
		}
	}
}
