package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/update"
	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/version"
)

type fakeInstaller struct{ versions *[]string }

func (f fakeInstaller) Install(_ context.Context, v string) error {
	*f.versions = append(*f.versions, v)
	return nil
}

type updateStub struct {
	checks   atomic.Int32
	installs []string
}

// stubUpdate runs release 0.2.8 against a fake release check reporting
// latest, with the given install method, an isolated config dir, and no
// real terminal or executable.
func stubUpdate(t *testing.T, latest, method string) *updateStub {
	t.Helper()
	oldCheck, oldMethod, oldExec, oldInstaller, oldStdin := checkForUpdate, installMethod, executablePath, newInstaller, stdinIsTerminal
	oldVersion, oldNoInput, oldReadOnly, oldOutput, oldQuiet, oldProfile := version.Version, noInputFlag, readOnlyFlag, outputFormat, quietFlag, profileFlag
	t.Cleanup(func() {
		checkForUpdate, installMethod, executablePath, newInstaller, stdinIsTerminal = oldCheck, oldMethod, oldExec, oldInstaller, oldStdin
		version.Version, noInputFlag, readOnlyFlag, outputFormat, quietFlag, profileFlag = oldVersion, oldNoInput, oldReadOnly, oldOutput, oldQuiet, oldProfile
	})
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("JENKINS_READ_ONLY", "")
	version.Version = "0.2.8"
	noInputFlag, readOnlyFlag, outputFormat, quietFlag, profileFlag = false, false, "", true, ""

	s := &updateStub{}
	checkForUpdate = func(current, _ string, force bool) (*update.UpdateInfo, error) {
		s.checks.Add(1)
		if !force {
			t.Error("update must bypass the cache")
		}
		return &update.UpdateInfo{
			Available:      latest != current,
			CurrentVersion: current,
			LatestVersion:  latest,
			ReleaseURL:     update.ReleaseURL(latest),
		}, nil
	}
	installMethod = func() string { return method }
	executablePath = func() (string, error) { return filepath.Join(t.TempDir(), "jenkins"), nil }
	newInstaller = func(string) releaseInstaller { return fakeInstaller{&s.installs} }
	stdinIsTerminal = func() bool { return false }
	return s
}

func runUpdate(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	cmd := newUpdateCmd()
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestUpdateCheckJSON(t *testing.T) {
	for _, tc := range []struct {
		latest, method string
		available      bool
	}{{"0.2.9", update.InstallSelf, true}, {"0.2.8", update.InstallGo, false}} {
		s := stubUpdate(t, tc.latest, tc.method)
		outputFormat = "json"
		out, err := runUpdate(t, "", "--check")
		if err != nil {
			t.Fatalf("update --check -o json: %v", err)
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("output is not JSON: %v\n%s", err, out)
		}
		want := map[string]any{
			"current_version":  "0.2.8",
			"latest_version":   tc.latest,
			"update_available": tc.available,
			"release_url":      "https://github.com/piyush-gambhir/jenkins-cli/releases/tag/v" + tc.latest,
			"install_method":   tc.method,
		}
		if len(got) != len(want) {
			t.Errorf("fields %v, want exactly %v", got, want)
		}
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%s = %v, want %v", k, got[k], v)
			}
		}
		if len(s.installs) != 0 {
			t.Errorf("--check installed %v", s.installs)
		}
	}
}

func TestUpdateCheckText(t *testing.T) {
	stubUpdate(t, "0.2.9", update.InstallSelf)
	out, err := runUpdate(t, "", "--check")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"v0.2.8", "v0.2.9", "Update available: yes", "releases/tag/v0.2.9", "jenkins update"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}

func TestUpdateAlreadyLatest(t *testing.T) {
	s := stubUpdate(t, "0.2.8", update.InstallSelf)
	out, err := runUpdate(t, "", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "already the latest version") || len(s.installs) != 0 {
		t.Errorf("output %q, installs %v; want an up-to-date message and no install", out, s.installs)
	}
}

func TestUpdateYesInstalls(t *testing.T) {
	s := stubUpdate(t, "0.2.9", update.InstallSelf)
	cache := filepath.Join(config.ConfigDir(), "update-check.json")
	if err := os.MkdirAll(filepath.Dir(cache), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, []byte(`{"latest_version":"0.2.9"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := runUpdate(t, "", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.installs) != 1 || s.installs[0] != "0.2.9" {
		t.Fatalf("installs = %v, want [0.2.9]", s.installs)
	}
	for _, want := range []string{"Updated jenkins v0.2.8 -> v0.2.9", "releases/tag/v0.2.9"} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Errorf("update cache was not cleared: %v", err)
	}
}

func TestUpdateWindowsInstalls(t *testing.T) {
	// The Windows refusal is gone: the platform no longer changes the flow.
	s := stubUpdate(t, "0.2.9", update.InstallSelf)
	if _, err := runUpdate(t, "", "-y"); err != nil {
		t.Fatal(err)
	}
	if len(s.installs) != 1 {
		t.Errorf("installs = %v", s.installs)
	}
}

func TestUpdateConfirmation(t *testing.T) {
	t.Run("no-input without --yes fails", func(t *testing.T) {
		s := stubUpdate(t, "0.2.9", update.InstallSelf)
		noInputFlag = true
		_, err := runUpdate(t, "")
		if err == nil || !strings.Contains(err.Error(), "--yes") || len(s.installs) != 0 {
			t.Errorf("err %v, installs %v; want a --yes hint and no install", err, s.installs)
		}
	})
	t.Run("non-terminal stdin without --yes fails", func(t *testing.T) {
		s := stubUpdate(t, "0.2.9", update.InstallSelf)
		_, err := runUpdate(t, "y\n")
		if err == nil || !strings.Contains(err.Error(), "--yes") || len(s.installs) != 0 {
			t.Errorf("err %v, installs %v; want a --yes hint and no install", err, s.installs)
		}
	})
	for _, tc := range []struct {
		answer  string
		install bool
	}{{"\n", true}, {"y\n", true}, {"YES\n", true}, {"n\n", false}, {"no\n", false}} {
		t.Run("prompt answer "+strings.TrimSpace(tc.answer), func(t *testing.T) {
			s := stubUpdate(t, "0.2.9", update.InstallSelf)
			stdinIsTerminal = func() bool { return true }
			out, err := runUpdate(t, tc.answer)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, "Update now? [Y/n]") {
				t.Errorf("no prompt in output:\n%s", out)
			}
			if got := len(s.installs) == 1; got != tc.install {
				t.Errorf("installed = %t, want %t", got, tc.install)
			}
		})
	}
}

func TestUpdateGoInstallDoesNotSelfReplace(t *testing.T) {
	s := stubUpdate(t, "0.2.9", update.InstallGo)
	out, err := runUpdate(t, "", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Update with: git pull && make install") || len(s.installs) != 0 {
		t.Errorf("output %q, installs %v; want the source update command and no install", out, s.installs)
	}
}

func TestUpdateReadOnly(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T)
	}{
		{"flag", func(*testing.T) { readOnlyFlag = true }},
		{"env", func(t *testing.T) { t.Setenv("JENKINS_READ_ONLY", "true") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := stubUpdate(t, "0.2.9", update.InstallSelf)
			tc.setup(t)
			_, err := runUpdate(t, "", "--yes")
			if err == nil || !strings.Contains(err.Error(), "read-only") {
				t.Fatalf("update in read-only mode: err = %v, want a read-only refusal", err)
			}
			if len(s.installs) != 0 || s.checks.Load() != 0 {
				t.Errorf("read-only update still checked (%d) or installed (%v)", s.checks.Load(), s.installs)
			}
			if _, err := runUpdate(t, "", "--check"); err != nil {
				t.Errorf("update --check in read-only mode: %v", err)
			}
		})
	}
}

// stubNotifier makes the notifier think it runs release 0.2.8 in an
// interactive terminal with no opt-out set, and counts release checks.
func stubNotifier(t *testing.T, latest string) *atomic.Int32 {
	t.Helper()
	oldCheck, oldTTY, oldVersion, oldQuiet, oldMethod, oldResult := checkForUpdate, stderrIsTerminal, version.Version, quietFlag, installMethod, updateResult
	t.Cleanup(func() {
		checkForUpdate, stderrIsTerminal, version.Version, quietFlag, installMethod, updateResult = oldCheck, oldTTY, oldVersion, oldQuiet, oldMethod, oldResult
	})
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, name := range []string{"CI", "JENKINS_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		t.Setenv(name, "")
	}
	version.Version, quietFlag = "0.2.8", false
	stderrIsTerminal = func() bool { return true }
	installMethod = func() string { return update.InstallSelf }
	var checks atomic.Int32
	checkForUpdate = func(current, _ string, _ bool) (*update.UpdateInfo, error) {
		checks.Add(1)
		return &update.UpdateInfo{Available: true, CurrentVersion: current, LatestVersion: latest}, nil
	}
	return &checks
}

// waitForUpdateResult waits for the background check so the notice test is
// deterministic; the real PostRun never waits.
func waitForUpdateResult(t *testing.T) {
	t.Helper()
	select {
	case info := <-updateResult:
		ch := make(chan *update.UpdateInfo, 1)
		ch <- info
		updateResult = ch
	case <-time.After(5 * time.Second):
		t.Fatal("background update check did not finish")
	}
}

func TestUpdateNotifierShowsNotice(t *testing.T) {
	checks := stubNotifier(t, "0.2.9")
	startUpdateCheck("job")
	waitForUpdateResult(t)
	var out bytes.Buffer
	printUpdateNotice(&out)
	if checks.Load() != 1 || !strings.Contains(out.String(), "A new version of jenkins is available: v0.2.8 -> v0.2.9") {
		t.Errorf("checks %d, notice %q", checks.Load(), out.String())
	}
}

func TestUpdateNotifierSuppressed(t *testing.T) {
	for _, tc := range []struct {
		name  string
		top   string
		setup func(t *testing.T)
	}{
		{"stderr not a terminal", "job", func(*testing.T) { stderrIsTerminal = func() bool { return false } }},
		{"CI", "job", func(t *testing.T) { t.Setenv("CI", "true") }},
		{"JENKINS_NO_UPDATE_NOTIFIER", "job", func(t *testing.T) { t.Setenv("JENKINS_NO_UPDATE_NOTIFIER", "1") }},
		{"NO_UPDATE_NOTIFIER", "job", func(t *testing.T) { t.Setenv("NO_UPDATE_NOTIFIER", "1") }},
		{"quiet", "job", func(*testing.T) { quietFlag = true }},
		{"dev build", "job", func(*testing.T) { version.Version = "dev" }},
		{"update", "update", func(*testing.T) {}},
		{"version", "version", func(*testing.T) {}},
		{"completion", "completion", func(*testing.T) {}},
		{"help", "help", func(*testing.T) {}},
		{"__complete", "__complete", func(*testing.T) {}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checks := stubNotifier(t, "0.2.9")
			tc.setup(t)
			startUpdateCheck(tc.top)
			var out bytes.Buffer
			printUpdateNotice(&out)
			if updateResult != nil || checks.Load() != 0 || out.Len() != 0 {
				t.Errorf("suppressed notifier ran: result channel %v, checks %d, output %q", updateResult != nil, checks.Load(), out.String())
			}
		})
	}
}

func TestUpdateNoticeNeverWaits(t *testing.T) {
	stubNotifier(t, "0.2.9")
	release := make(chan struct{})
	defer close(release)
	slow := checkForUpdate
	checkForUpdate = func(current, dir string, force bool) (*update.UpdateInfo, error) {
		<-release
		return slow(current, dir, force)
	}
	startUpdateCheck("job")
	var out bytes.Buffer
	start := time.Now()
	printUpdateNotice(&out)
	if elapsed := time.Since(start); elapsed > time.Second || out.Len() != 0 {
		t.Errorf("printUpdateNotice waited %v and printed %q for a check still in flight", elapsed, out.String())
	}
}

func TestTopLevelCommandDrivesSkips(t *testing.T) {
	for _, tc := range []struct {
		args      []string
		skipsAuth bool
	}{
		{[]string{"update"}, true},
		{[]string{"version"}, true},
		{[]string{"login"}, true},
		{[]string{"job", "update"}, false},
		{[]string{"credential", "update"}, false},
		{[]string{"job", "list"}, false},
	} {
		cmd, _, err := rootCmd.Find(tc.args)
		if err != nil {
			t.Fatalf("finding %v: %v", tc.args, err)
		}
		if got := skipsAuth(topLevelName(cmd)); got != tc.skipsAuth {
			t.Errorf("%v: skipsAuth = %t, want %t", tc.args, got, tc.skipsAuth)
		}
	}
	// Cobra adds these lazily during Execute, so check them by name.
	for _, name := range []string{"completion", "help", "__complete", "__completeNoDesc"} {
		if !skipsAuth(name) {
			t.Errorf("%s needs a Jenkins server, want it to skip auth", name)
		}
	}
}

func TestJobUpdateIsNotTreatedAsSelfUpdate(t *testing.T) {
	stubNotifier(t, "0.2.9")
	stderrIsTerminal = func() bool { return false }
	oldReadOnly := readOnlyFlag
	t.Cleanup(func() { readOnlyFlag = oldReadOnly })
	t.Setenv("JENKINS_URL", "http://127.0.0.1:1")
	// Before the fix, `job update` skipped auth and read-only checks because
	// its leaf name is "update", then dereferenced a nil client. Run the root
	// hook directly: rootCmd.Execute would add cobra's lazy commands to the
	// shared tree.
	cmd, _, err := rootCmd.Find([]string{"job", "update"})
	if err != nil {
		t.Fatal(err)
	}
	readOnlyFlag = true
	err = rootCmd.PersistentPreRunE(cmd, []string{"my-job"})
	if err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("job update --read-only: err = %v, want a read-only refusal", err)
	}
}

func TestVersionShowsCachedLatestOnly(t *testing.T) {
	checks := stubNotifier(t, "0.2.9")
	run := func() string {
		cmd := newVersionCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs(nil)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}
	if out := run(); strings.Contains(out, "latest") {
		t.Errorf("version without a cache printed latest:\n%s", out)
	}
	dir := config.ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "update-check.json"), []byte(`{"last_checked":"2026-10-01T00:00:00Z","latest_version":"0.2.9"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out := run()
	if !strings.HasPrefix(out, "jenkins version 0.2.8 (commit: ") || !strings.Contains(out, "\nlatest: 0.2.9\nupdate_available: true\n") {
		t.Errorf("version output:\n%s", out)
	}
	if checks.Load() != 0 {
		t.Errorf("version queried GitHub %d times", checks.Load())
	}
}
