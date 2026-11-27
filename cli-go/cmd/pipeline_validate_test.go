package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/output"
)

func TestPipelineValidate_ExitStatusFollowsResult(t *testing.T) {
	jenkinsfile := filepath.Join(t.TempDir(), "Jenkinsfile")
	if err := os.WriteFile(jenkinsfile, []byte("pipeline {}"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		body    string
		wantErr bool
	}{
		{body: "Jenkinsfile successfully validated.", wantErr: false},
		{body: "Errors encountered validating Jenkinsfile:\nWorkflowScript: 1: Missing required section \"stages\"", wantErr: true},
	} {
		useTestJenkins(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(tc.body))
		}, output.FormatTable)

		cmd := newPipelineValidateCmd()
		if err := cmd.ParseFlags([]string{"--from-file", jenkinsfile}); err != nil {
			t.Fatal(err)
		}
		_, err := captureStdout(t, func() error { return cmd.RunE(cmd, nil) })
		if (err != nil) != tc.wantErr {
			t.Errorf("body %q: err = %v, wantErr %v", tc.body, err, tc.wantErr)
		}
	}
}
