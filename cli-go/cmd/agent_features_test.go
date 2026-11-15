package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/output"
)

func TestNoInputFlag_BlocksConfirmPrompt(t *testing.T) {
	origNoInput := noInputFlag
	defer func() { noInputFlag = origNoInput }()

	noInputFlag = true

	tests := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"job delete", newJobDeleteCmd},
		{"node delete", newNodeDeleteCmd},
		{"view delete", newViewDeleteCmd},
		{"credential delete", newCredentialDeleteCmd},
		{"plugin uninstall", newPluginUninstallCmd},
		{"build delete", newBuildDeleteCmd},
		{"system restart", newSystemRestartCmd},
		{"job wipe-workspace", newJobWipeWorkspaceCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()

			var args []string
			switch tt.name {
			case "build delete":
				args = []string{"test-job", "1"}
			case "credential delete", "node delete", "view delete",
				"job delete", "plugin uninstall", "job wipe-workspace":
				args = []string{"test-item"}
			case "system restart":
				args = []string{}
			}

			err := cmd.RunE(cmd, args)
			if err == nil {
				t.Fatal("expected error when no-input is set and --confirm is not provided")
			}
			expected := "interactive input required but --no-input is set. Use --confirm for destructive operations."
			if err.Error() != expected {
				t.Fatalf("expected error %q, got: %v", expected, err)
			}
		})
	}
}

func TestQuietFlag_Variable(t *testing.T) {
	origQuiet := quietFlag
	defer func() { quietFlag = origQuiet }()

	quietFlag = true
	if !quietFlag {
		t.Fatal("expected quietFlag to be true")
	}

	quietFlag = false
	if quietFlag {
		t.Fatal("expected quietFlag to be false")
	}
}

func TestNoInputFlag_AllowsWithConfirm(t *testing.T) {
	origNoInput := noInputFlag
	defer func() { noInputFlag = origNoInput }()

	noInputFlag = true

	cmd := newJobDeleteCmd()
	if err := cmd.ParseFlags([]string{"--confirm"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	var runErr error
	panicked := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		runErr = cmd.RunE(cmd, []string{"test-job"})
	}()

	if panicked {
		return
	}

	if runErr != nil {
		noInputErr := "interactive input required but --no-input is set. Use --confirm for destructive operations."
		if runErr.Error() == noInputErr {
			t.Fatal("--confirm should bypass the no-input check, but got the no-input error")
		}
	}
}

func TestIdempotentFlags_Exist(t *testing.T) {
	createCmds := []struct {
		name    string
		cmdFunc func() *cobra.Command
		flag    string
	}{
		{"job create", newJobCreateCmd, "if-not-exists"},
		{"node create", newNodeCreateCmd, "if-not-exists"},
		{"view create", newViewCreateCmd, "if-not-exists"},
		{"credential create", newCredentialCreateCmd, "if-not-exists"},
		{"plugin install", newPluginInstallCmd, "if-not-exists"},
		{"job delete", newJobDeleteCmd, "if-exists"},
		{"node delete", newNodeDeleteCmd, "if-exists"},
		{"view delete", newViewDeleteCmd, "if-exists"},
		{"credential delete", newCredentialDeleteCmd, "if-exists"},
		{"plugin uninstall", newPluginUninstallCmd, "if-exists"},
	}

	for _, tt := range createCmds {
		t.Run(tt.name+" has --"+tt.flag, func(t *testing.T) {
			cmd := tt.cmdFunc()
			f := cmd.Flags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("expected --%s flag on %s command", tt.flag, tt.name)
			}
		})
	}
}

func TestStdinFlag_FromFileAcceptsDash(t *testing.T) {
	cmds := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"job create", newJobCreateCmd},
		{"job update", newJobUpdateCmd},
		{"credential create", newCredentialCreateCmd},
		{"credential update", newCredentialUpdateCmd},
		{"pipeline validate", newPipelineValidateCmd},
		{"system run-script", newSystemRunScriptCmd},
	}

	for _, tt := range cmds {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()
			f := cmd.Flags().Lookup("from-file")
			if f == nil {
				t.Fatalf("expected --from-file flag on %s command", tt.name)
			}
			if f.Usage == "" {
				t.Fatalf("expected non-empty usage for --from-file on %s", tt.name)
			}
		})
	}
}

func TestStructuredJSONError(t *testing.T) {
	var buf bytes.Buffer
	testErr := errors.New("job not found: my-pipeline")

	output.WriteError(&buf, output.FormatJSON, testErr, 404)

	var parsed output.ErrorResponse
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	if parsed.Error != "job not found: my-pipeline" {
		t.Errorf("expected error 'job not found: my-pipeline', got %q", parsed.Error)
	}
	if parsed.StatusCode != 404 {
		t.Errorf("expected status_code 404, got %d", parsed.StatusCode)
	}
}

func TestStructuredJSONError_APIError(t *testing.T) {
	var buf bytes.Buffer
	apiErr := &client.APIError{
		StatusCode: 403,
		Status:     "403 Forbidden",
		Message:    "Permission denied",
		URL:        "http://jenkins/job/my-pipeline/build",
	}

	output.WriteError(&buf, output.FormatJSON, apiErr, apiErr.StatusCode)

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	if parsed["status_code"] != float64(403) {
		t.Errorf("expected status_code 403, got %v", parsed["status_code"])
	}
}

func TestIdempotentDelete_IfExists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/crumbIssuer/api/json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer ts.Close()

	origClient := jenkinsClient
	origNoInput := noInputFlag
	origQuiet := quietFlag
	defer func() {
		jenkinsClient = origClient
		noInputFlag = origNoInput
		quietFlag = origQuiet
	}()

	jenkinsClient = client.NewClient(config.Profile{
		URL:      ts.URL,
		Username: "admin",
		Token:    "tok",
	})
	noInputFlag = false
	quietFlag = true

	cmd := newJobDeleteCmd()
	if err := cmd.ParseFlags([]string{"--confirm", "--if-exists"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	err := cmd.RunE(cmd, []string{"nonexistent-job"})
	if err != nil {
		t.Fatalf("expected no error with --if-exists on 404, got: %v", err)
	}
}

func TestGlobalFlags_Registered(t *testing.T) {
	flags := []struct {
		name      string
		shorthand string
	}{
		{"no-input", ""},
		{"quiet", "q"},
		{"verbose", "v"},
		{"read-only", ""},
		{"no-color", ""},
	}

	for _, f := range flags {
		t.Run("--"+f.name, func(t *testing.T) {
			pf := rootCmd.PersistentFlags().Lookup(f.name)
			if pf == nil {
				t.Fatalf("expected persistent flag --%s on root command", f.name)
			}
			if f.shorthand != "" && pf.Shorthand != f.shorthand {
				t.Errorf("expected shorthand %q for --%s, got %q", f.shorthand, f.name, pf.Shorthand)
			}
		})
	}
}
