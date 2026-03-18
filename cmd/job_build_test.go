package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/piyush-gambhir/jenkins-cli/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

// newJobBuildTestServer creates a test server with a crumb endpoint returning 404
// (CSRF disabled) and a custom handler mux. Returns the server, a client, and the mux.
func newJobBuildTestServer(t *testing.T) (*httptest.Server, *client.Client, *http.ServeMux) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/crumbIssuer/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	ts := httptest.NewServer(mux)

	c := client.NewClient(config.Profile{
		URL:      ts.URL,
		Username: "admin",
		Token:    "secret-token",
	})

	return ts, c, mux
}

// TestTriggerBuild_Basic verifies that POST /job/{name}/build returns 201 with
// a Location header and that the queue URL is parsed correctly.
func TestTriggerBuild_Basic(t *testing.T) {
	ts, c, mux := newJobBuildTestServer(t)
	defer ts.Close()

	mux.HandleFunc("/job/my-pipeline/build", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Location", ts.URL+"/queue/item/42/")
		w.WriteHeader(http.StatusCreated)
	})

	ql, err := c.TriggerBuild("my-pipeline", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ql.QueueURL == "" {
		t.Fatal("expected non-empty queue URL")
	}
	if !strings.Contains(ql.QueueURL, "/queue/item/42") {
		t.Errorf("expected queue URL to contain /queue/item/42, got %q", ql.QueueURL)
	}
}

// TestTriggerBuild_WithParams verifies that POST /job/{name}/buildWithParameters
// is called when parameters are provided, and that params are encoded in the query.
func TestTriggerBuild_WithParams(t *testing.T) {
	ts, c, mux := newJobBuildTestServer(t)
	defer ts.Close()

	var gotPath string
	var gotBranch string
	var gotEnv string

	mux.HandleFunc("/job/my-pipeline/buildWithParameters", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBranch = r.URL.Query().Get("BRANCH")
		gotEnv = r.URL.Query().Get("ENV")
		w.Header().Set("Location", ts.URL+"/queue/item/43/")
		w.WriteHeader(http.StatusCreated)
	})

	params := map[string]string{"BRANCH": "main", "ENV": "staging"}
	ql, err := c.TriggerBuild("my-pipeline", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/job/my-pipeline/buildWithParameters" {
		t.Errorf("expected path /job/my-pipeline/buildWithParameters, got %q", gotPath)
	}
	if gotBranch != "main" {
		t.Errorf("expected BRANCH=main, got %q", gotBranch)
	}
	if gotEnv != "staging" {
		t.Errorf("expected ENV=staging, got %q", gotEnv)
	}
	if ql.QueueURL == "" {
		t.Fatal("expected non-empty queue URL")
	}
}

// TestTriggerBuild_Wait mocks queue polling (return "queued" then "executable")
// and verifies the build number is extracted correctly.
func TestTriggerBuild_Wait(t *testing.T) {
	ts, c, mux := newJobBuildTestServer(t)
	defer ts.Close()

	queueCallCount := 0

	mux.HandleFunc("/queue/item/44/api/json", func(w http.ResponseWriter, r *http.Request) {
		queueCallCount++
		w.Header().Set("Content-Type", "application/json")
		if queueCallCount == 1 {
			// First call: still queued
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         44,
				"why":        "Waiting for executor",
				"blocked":    true,
				"buildable":  true,
				"executable": map[string]interface{}{},
			})
		} else {
			// Second call: build started
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 44,
				"executable": map[string]interface{}{
					"number": 99,
					"url":    ts.URL + "/job/my-pipeline/99/",
				},
			})
		}
	})

	buildRef, err := c.WaitForQueuedBuild(ts.URL+"/queue/item/44/", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buildRef.Number != 99 {
		t.Errorf("expected build number 99, got %d", buildRef.Number)
	}
	if queueCallCount < 2 {
		t.Errorf("expected at least 2 queue poll calls, got %d", queueCallCount)
	}
}

// TestTriggerBuild_WaitTimeout mocks a slow queue and verifies timeout error.
func TestTriggerBuild_WaitTimeout(t *testing.T) {
	ts, c, mux := newJobBuildTestServer(t)
	defer ts.Close()

	mux.HandleFunc("/queue/item/45/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Always return "still queued" to trigger timeout
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         45,
			"why":        "Waiting for executor",
			"blocked":    true,
			"buildable":  true,
			"executable": map[string]interface{}{},
		})
	})

	_, err := c.WaitForQueuedBuild(ts.URL+"/queue/item/45/", 3*time.Second)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

// TestTriggerBuild_Follow mocks the progressive text API with two chunks
// and verifies output is captured correctly.
func TestTriggerBuild_Follow(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/logText/progressiveText") {
			start := r.URL.Query().Get("start")
			callCount++

			switch {
			case start == "0":
				w.Header().Set("X-Text-Size", "50")
				w.Header().Set("X-More-Data", "true")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "Building step 1...\n")
			case start == "50":
				w.Header().Set("X-Text-Size", "100")
				// No X-More-Data -> done
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "Finished: SUCCESS\n")
			default:
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := client.NewClient(config.Profile{URL: ts.URL, Username: "admin", Token: "tok"})

	var buf bytes.Buffer
	err := c.StreamBuildLog("my-pipeline", 1, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Building step 1") {
		t.Errorf("expected output to contain 'Building step 1', got %q", output)
	}
	if !strings.Contains(output, "Finished: SUCCESS") {
		t.Errorf("expected output to contain 'Finished: SUCCESS', got %q", output)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 progressive text requests, got %d", callCount)
	}
}

// TestTriggerBuild_WaitAndFollow tests the full flow: trigger -> queue -> start -> stream -> complete.
func TestTriggerBuild_WaitAndFollow(t *testing.T) {
	queueCallCount := 0
	buildCallCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/crumbIssuer/api/json":
			w.WriteHeader(http.StatusNotFound)

		case r.URL.Path == "/job/my-pipeline/build" && r.Method == "POST":
			w.Header().Set("Location", "http://"+r.Host+"/queue/item/50")
			w.WriteHeader(http.StatusCreated)

		case r.URL.Path == "/queue/item/50/api/json":
			queueCallCount++
			w.Header().Set("Content-Type", "application/json")
			if queueCallCount == 1 {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":         50,
					"why":        "waiting",
					"executable": map[string]interface{}{},
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id": 50,
					"executable": map[string]interface{}{
						"number": 7,
						"url":    "http://" + r.Host + "/job/my-pipeline/7/",
					},
				})
			}

		case r.URL.Path == "/job/my-pipeline/7/api/json":
			buildCallCount++
			w.Header().Set("Content-Type", "application/json")
			if buildCallCount == 1 {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"number":   7,
					"result":   nil,
					"building": true,
					"duration": 0,
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"number":   7,
					"result":   "SUCCESS",
					"building": false,
					"duration": 5000,
				})
			}

		case strings.HasSuffix(r.URL.Path, "/logText/progressiveText"):
			w.Header().Set("X-Text-Size", "100")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Full build log output\n")

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	c := client.NewClient(config.Profile{URL: ts.URL, Username: "admin", Token: "tok"})

	build, err := c.TriggerBuildAndWait("my-pipeline", nil, 30*time.Second)
	if err != nil {
		t.Fatalf("TriggerBuildAndWait failed: %v", err)
	}
	if build.Number != 7 {
		t.Errorf("expected build number 7, got %d", build.Number)
	}
	if build.Result != "SUCCESS" {
		t.Errorf("expected result SUCCESS, got %q", build.Result)
	}

	// Now stream the log
	var buf bytes.Buffer
	err = c.StreamBuildLog("my-pipeline", 7, &buf)
	if err != nil {
		t.Fatalf("StreamBuildLog failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Full build log output") {
		t.Errorf("expected log to contain 'Full build log output', got %q", buf.String())
	}
}

// TestTriggerBuild_FailedBuild verifies that a build completing with FAILURE
// result is reported correctly.
func TestTriggerBuild_FailedBuild(t *testing.T) {
	queueCallCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/crumbIssuer/api/json":
			w.WriteHeader(http.StatusNotFound)

		case r.URL.Path == "/job/failing-job/build" && r.Method == "POST":
			w.Header().Set("Location", "http://"+r.Host+"/queue/item/60")
			w.WriteHeader(http.StatusCreated)

		case r.URL.Path == "/queue/item/60/api/json":
			queueCallCount++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 60,
				"executable": map[string]interface{}{
					"number": 3,
					"url":    "http://" + r.Host + "/job/failing-job/3/",
				},
			})

		case r.URL.Path == "/job/failing-job/3/api/json":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"number":   3,
				"result":   "FAILURE",
				"building": false,
				"duration": 2000,
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	c := client.NewClient(config.Profile{URL: ts.URL, Username: "admin", Token: "tok"})

	build, err := c.TriggerBuildAndWait("failing-job", nil, 30*time.Second)
	if err != nil {
		t.Fatalf("TriggerBuildAndWait failed: %v", err)
	}

	// The build completed -- the client should return it even with FAILURE result
	if build.Result != "FAILURE" {
		t.Errorf("expected result FAILURE, got %q", build.Result)
	}
	if build.Number != 3 {
		t.Errorf("expected build number 3, got %d", build.Number)
	}
}

// TestStreamBuildLog_ContextTimeout verifies that StreamBuildLog respects context
// cancellation and returns cleanly with partial output.
func TestStreamBuildLog_ContextTimeout(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		callCount++
		w.Header().Set("X-Text-Size", fmt.Sprintf("%d", callCount*100))
		w.Header().Set("X-More-Data", "true") // Always more data
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Chunk %d\n", callCount)
	}))
	defer ts.Close()

	c := client.NewClient(config.Profile{URL: ts.URL, Username: "admin", Token: "tok"})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var buf bytes.Buffer
	err := c.StreamBuildLog("my-pipeline", 1, &buf, ctx)
	if err != nil {
		t.Fatalf("expected nil error on context timeout, got: %v", err)
	}

	// We should have gotten at least some partial output
	output := buf.String()
	if !strings.Contains(output, "Chunk 1") {
		t.Errorf("expected at least 'Chunk 1' in partial output, got %q", output)
	}
}
