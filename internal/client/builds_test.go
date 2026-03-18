package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

func newBuildTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
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

func TestListBuilds(t *testing.T) {
	ts, c := newBuildTestServer(func(w http.ResponseWriter, r *http.Request) {
		// "my-job" -> /job/my-job/api/json
		if r.URL.Path != "/job/my-job/api/json" {
			t.Errorf("expected path /job/my-job/api/json, got %q", r.URL.Path)
		}
		tree := r.URL.Query().Get("tree")
		if tree == "" {
			t.Error("expected tree param to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BuildListResponse{
			Builds: []Build{
				{Number: 10, Result: "SUCCESS"},
				{Number: 9, Result: "FAILURE"},
			},
		})
	})
	defer ts.Close()

	builds, err := c.ListBuilds("my-job", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(builds) != 2 {
		t.Fatalf("expected 2 builds, got %d", len(builds))
	}
	if builds[0].Number != 10 {
		t.Errorf("expected first build #10, got #%d", builds[0].Number)
	}
	if builds[1].Result != "FAILURE" {
		t.Errorf("expected second build result FAILURE, got %q", builds[1].Result)
	}
}

func TestGetBuild(t *testing.T) {
	ts, c := newBuildTestServer(func(w http.ResponseWriter, r *http.Request) {
		// "my-job", number 42 -> /job/my-job/42/api/json
		if r.URL.Path != "/job/my-job/42/api/json" {
			t.Errorf("expected path /job/my-job/42/api/json, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Build{
			Number:   42,
			Result:   "SUCCESS",
			Building: false,
			Duration: 12345,
		})
	})
	defer ts.Close()

	build, err := c.GetBuild("my-job", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if build.Number != 42 {
		t.Errorf("expected build #42, got #%d", build.Number)
	}
	if build.Result != "SUCCESS" {
		t.Errorf("expected result SUCCESS, got %q", build.Result)
	}
}

func TestGetBuildLog(t *testing.T) {
	expectedLog := "Started by user admin\nBuilding...\nFinished: SUCCESS\n"
	ts, c := newBuildTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/my-job/5/consoleText" {
			t.Errorf("expected path /job/my-job/5/consoleText, got %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedLog))
	})
	defer ts.Close()

	log, err := c.GetBuildLog("my-job", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log != expectedLog {
		t.Errorf("expected log %q, got %q", expectedLog, log)
	}
}

func TestStreamBuildLog(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Should hit /job/my-job/1/logText/progressiveText
		if !strings.HasSuffix(r.URL.Path, "/logText/progressiveText") {
			t.Errorf("unexpected path: %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		start := r.URL.Query().Get("start")
		callCount++

		switch {
		case start == "0":
			// First chunk
			w.Header().Set("X-Text-Size", "100")
			w.Header().Set("X-More-Data", "true")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Line 1\nLine 2\n")
		case start == "100":
			// Second chunk, no more data
			w.Header().Set("X-Text-Size", "200")
			// No X-More-Data header -> stream ends
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Line 3\nDone\n")
		default:
			t.Errorf("unexpected start value: %q", start)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	c := NewClient(config.Profile{URL: ts.URL, Username: "admin", Token: "tok"})
	c.httpClient = ts.Client()

	var buf bytes.Buffer
	err := c.StreamBuildLog("my-job", 1, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Line 1") {
		t.Errorf("expected output to contain 'Line 1', got %q", output)
	}
	if !strings.Contains(output, "Line 3") {
		t.Errorf("expected output to contain 'Line 3', got %q", output)
	}
	if !strings.Contains(output, "Done") {
		t.Errorf("expected output to contain 'Done', got %q", output)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 requests, got %d", callCount)
	}
}

func TestStopBuild(t *testing.T) {
	var gotPath, gotMethod string
	ts, c := newBuildTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.StopBuild("my-job", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/job/my-job/7/stop" {
		t.Errorf("expected path /job/my-job/7/stop, got %q", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
}

func TestDeleteBuild(t *testing.T) {
	var gotPath, gotMethod string
	ts, c := newBuildTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	err := c.DeleteBuild("my-job", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/job/my-job/3/doDelete" {
		t.Errorf("expected path /job/my-job/3/doDelete, got %q", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("expected POST, got %s", gotMethod)
	}
}

func TestGetBuildArtifacts(t *testing.T) {
	ts, c := newBuildTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/my-job/10/api/json" {
			t.Errorf("expected path /job/my-job/10/api/json, got %q", r.URL.Path)
		}
		tree := r.URL.Query().Get("tree")
		if !strings.Contains(tree, "artifacts") {
			t.Error("expected tree param to contain 'artifacts'")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Build{
			Number: 10,
			Artifacts: []Artifact{
				{FileName: "app.jar", RelativePath: "target/app.jar"},
				{FileName: "report.html", RelativePath: "reports/report.html"},
			},
		})
	})
	defer ts.Close()

	artifacts, err := c.GetBuildArtifacts("my-job", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(artifacts))
	}
	if artifacts[0].FileName != "app.jar" {
		t.Errorf("expected first artifact 'app.jar', got %q", artifacts[0].FileName)
	}
}

func TestGetBuildTestReport(t *testing.T) {
	ts, c := newBuildTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/my-job/5/testReport/api/json" {
			t.Errorf("expected path /job/my-job/5/testReport/api/json, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TestReport{
			FailCount:  1,
			PassCount:  9,
			SkipCount:  2,
			TotalCount: 12,
			Suites: []TestSuite{
				{
					Name: "com.example.Tests",
					Cases: []TestCase{
						{Name: "testOne", Status: "PASSED"},
					},
				},
			},
		})
	})
	defer ts.Close()

	report, err := c.GetBuildTestReport("my-job", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.PassCount != 9 {
		t.Errorf("expected 9 passes, got %d", report.PassCount)
	}
	if report.FailCount != 1 {
		t.Errorf("expected 1 failure, got %d", report.FailCount)
	}
	if len(report.Suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(report.Suites))
	}
	if report.Suites[0].Name != "com.example.Tests" {
		t.Errorf("expected suite name 'com.example.Tests', got %q", report.Suites[0].Name)
	}
}
