package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/output"
)

func TestJobCreate_IfNotExistsOnlySkipsExistingJob(t *testing.T) {
	configXML := filepath.Join(t.TempDir(), "config.xml")
	if err := os.WriteFile(configXML, []byte("<project/>"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		exists  bool
		wantErr bool
	}{
		{exists: true, wantErr: false},
		{exists: false, wantErr: true},
	} {
		useTestJenkins(t, func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == "POST" && r.URL.Path == "/job/team/createItem":
				w.WriteHeader(http.StatusBadRequest)
			case r.Method == "GET" && r.URL.Path == "/job/team/job/app/api/json" && tc.exists:
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"name": "app"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}, output.FormatTable)

		cmd := newJobCreateCmd()
		if err := cmd.ParseFlags([]string{"--from-file", configXML, "--folder", "team", "--if-not-exists"}); err != nil {
			t.Fatal(err)
		}
		err := cmd.RunE(cmd, []string{"app"})
		if (err != nil) != tc.wantErr {
			t.Errorf("exists=%v: err = %v, wantErr %v", tc.exists, err, tc.wantErr)
		}
	}
}
