package update

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintUpdateNoticeSuggestsSelfUpdateOnlyWhereSupported(t *testing.T) {
	info := &UpdateInfo{
		Available:      true,
		CurrentVersion: "0.2.7",
		LatestVersion:  "0.2.8",
		ReleaseURL:     "https://github.com/piyush-gambhir/jenkins-cli/releases/tag/v0.2.8",
	}
	oldGOOS := goos
	t.Cleanup(func() { goos = oldGOOS })

	for _, tc := range []struct {
		goos       string
		suggestCmd bool
	}{{"linux", true}, {"darwin", true}, {"windows", false}} {
		goos = tc.goos
		var buf bytes.Buffer
		PrintUpdateNotice(&buf, info)
		out := buf.String()
		if got := strings.Contains(out, "jenkins update"); got != tc.suggestCmd {
			t.Errorf("%s: suggests `jenkins update` = %t, want %t; notice:\n%s", tc.goos, got, tc.suggestCmd, out)
		}
		if !strings.Contains(out, info.ReleaseURL) {
			t.Errorf("%s: notice does not link the release; notice:\n%s", tc.goos, out)
		}
	}
}
