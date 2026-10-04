package cmd

import (
	"strings"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/version"
)

// stubUpdateCheck pretends to run release 0.2.7 on osName with 0.2.8 available.
func stubUpdateCheck(t *testing.T, osName string) {
	t.Helper()
	oldGOOS, oldCheck, oldVersion, oldNoInput := goos, checkForUpdate, version.Version, noInputFlag
	t.Cleanup(func() {
		goos, checkForUpdate, version.Version, noInputFlag = oldGOOS, oldCheck, oldVersion, oldNoInput
	})
	goos = osName
	version.Version = "0.2.7"
	// --no-input keeps the test off stdin: the command must refuse before prompting.
	noInputFlag = true
	checkForUpdate = func(current, _, _ string, _ bool) (*update.UpdateInfo, error) {
		return &update.UpdateInfo{
			Available:      true,
			CurrentVersion: current,
			LatestVersion:  "0.2.8",
			ReleaseURL:     "https://github.com/piyush-gambhir/jenkins-cli/releases/tag/v0.2.8",
		}, nil
	}
}

func TestUpdateRefusesInstallOnWindows(t *testing.T) {
	stubUpdateCheck(t, "windows")
	cmd := newUpdateCmd()
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected update to refuse installing on Windows")
	}
	for _, want := range []string{"jenkins.exe", "https://github.com/piyush-gambhir/jenkins-cli/releases/tag/v0.2.8"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestUpdateCheckWorksOnWindows(t *testing.T) {
	stubUpdateCheck(t, "windows")
	cmd := newUpdateCmd()
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	cmd.SetArgs([]string{"--check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update --check on Windows: %v", err)
	}
}
