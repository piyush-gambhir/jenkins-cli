package cmd

import (
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
)

// build artifacts lists in read-only mode but must not download files there.
func TestArtifactDownloadIsALocalWriteInReadOnlyMode(t *testing.T) {
	cmd := newBuildArtifactsCmd()
	profile := &config.Profile{ReadOnly: true}
	if err := checkPermissions(cmd, profile); err != nil {
		t.Fatalf("listing artifacts blocked in read-only mode: %v", err)
	}
	if err := cmd.Flags().Set("download", "true"); err != nil {
		t.Fatal(err)
	}
	if err := checkPermissions(cmd, profile); err == nil {
		t.Fatal("artifact download allowed in read-only mode")
	}
}
