package update

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeReleases serves github.com/<repo>/releases/latest with status and the
// Location header built by location from the server URL (none when nil). It
// counts requests to releases/latest and to any other path, which would mean
// the redirect was followed.
func fakeReleases(t *testing.T, status int, location func(base string) string) (latest, other *atomic.Int32) {
	t.Helper()
	latest, other = &atomic.Int32{}, &atomic.Int32{}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+Repo+"/releases/latest" {
			other.Add(1)
			http.NotFound(w, r)
			return
		}
		latest.Add(1)
		if location != nil {
			w.Header().Set("Location", location(srv.URL))
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	old := releasesBaseURL
	releasesBaseURL = srv.URL
	t.Cleanup(func() { releasesBaseURL = old })
	return latest, other
}

func tagURL(tag string) func(string) string {
	return func(base string) string { return base + "/" + Repo + "/releases/tag/" + tag }
}

func TestFetchLatestFromRedirect(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		location func(string) string
		want     string // "" means an error
	}{
		{"302 to a semver tag", http.StatusFound, tagURL("v0.2.9"), "0.2.9"},
		{"pre-release tag", http.StatusFound, tagURL("v0.3.0-rc.1"), "0.3.0-rc.1"},
		{"relative Location", http.StatusFound, func(string) string { return "/" + Repo + "/releases/tag/v0.3.0" }, "0.3.0"},
		{"missing Location", http.StatusFound, nil, ""},
		{"foreign host", http.StatusFound, func(string) string { return "https://evil.example.com/" + Repo + "/releases/tag/v0.2.9" }, ""},
		{"other repo", http.StatusFound, func(base string) string { return base + "/someone/else/releases/tag/v0.2.9" }, ""},
		{"tag without v", http.StatusFound, tagURL("0.2.9"), ""},
		{"non-semver tag", http.StatusFound, tagURL("v1.2"), ""},
		{"named tag", http.StatusFound, tagURL("nightly"), ""},
		{"traversal in tag", http.StatusFound, tagURL("v0.2.9/../../evil"), ""},
		{"no release yet", http.StatusFound, func(base string) string { return base + "/" + Repo + "/releases" }, ""},
		{"not a redirect", http.StatusOK, tagURL("v0.2.9"), ""},
		{"server error", http.StatusInternalServerError, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			latest, other := fakeReleases(t, tc.status, tc.location)
			got, err := fetchLatest("0.2.8", time.Second)
			if tc.want == "" {
				if err == nil {
					t.Errorf("fetchLatest = %q, want an error", got)
				}
			} else if err != nil || got != tc.want {
				t.Errorf("fetchLatest = %q, %v; want %q", got, err, tc.want)
			}
			if latest.Load() != 1 || other.Load() != 0 {
				t.Errorf("requests: releases/latest %d, other %d; want 1 and 0 (the redirect must not be followed)", latest.Load(), other.Load())
			}
		})
	}
}

func TestNotifierEnabled(t *testing.T) {
	env := func(kv ...string) func(string) string {
		m := map[string]string{}
		for i := 0; i+1 < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) string { return m[k] }
	}
	for _, tc := range []struct {
		name    string
		getenv  func(string) string
		tty     bool
		quiet   bool
		version string
		command string
		want    bool
	}{
		{"interactive release build", env(), true, false, "0.2.8", "job", true},
		{"v-prefixed source build", env(), true, false, "v0.2.8-3-gabc1234", "job", true},
		{"stderr not a terminal", env(), false, false, "0.2.8", "job", false},
		{"CI set", env("CI", "true"), true, false, "0.2.8", "job", false},
		{"CI set to anything", env("CI", "0"), true, false, "0.2.8", "job", false},
		{"prefixed opt-out", env("JENKINS_NO_UPDATE_NOTIFIER", "1"), true, false, "0.2.8", "job", false},
		{"generic opt-out", env("NO_UPDATE_NOTIFIER", "yes"), true, false, "0.2.8", "job", false},
		{"quiet", env(), true, true, "0.2.8", "job", false},
		{"dev build", env(), true, false, "dev", "job", false},
		{"empty version", env(), true, false, "", "job", false},
		{"commit hash version", env(), true, false, "2b9e25a", "job", false},
		{"update command", env(), true, false, "0.2.8", "update", false},
		{"version command", env(), true, false, "0.2.8", "version", false},
		{"completion command", env(), true, false, "0.2.8", "completion", false},
		{"help command", env(), true, false, "0.2.8", "help", false},
		{"__complete", env(), true, false, "0.2.8", "__complete", false},
		{"__completeNoDesc", env(), true, false, "0.2.8", "__completeNoDesc", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NotifierEnabled(tc.getenv, tc.tty, tc.quiet, tc.version, tc.command); got != tc.want {
				t.Errorf("NotifierEnabled = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestNoticeFormat(t *testing.T) {
	info := &UpdateInfo{Available: true, CurrentVersion: "0.2.8", LatestVersion: "0.2.9"}
	want := "\nA new version of jenkins is available: v0.2.8 -> v0.2.9\n" +
		"Update with: jenkins update\n" +
		"Release notes: https://github.com/piyush-gambhir/jenkins-cli/releases/tag/v0.2.9\n"
	if got := Notice(info, InstallSelf); got != want {
		t.Errorf("self notice:\n%q\nwant:\n%q", got, want)
	}
	goNotice := Notice(info, InstallGo)
	if !strings.Contains(goNotice, "\nUpdate with: git pull && make install (in your jenkins-cli/cli-go checkout)\n") {
		t.Errorf("go-bin notice does not suggest rebuilding from source:\n%s", goNotice)
	}
	if strings.Contains(goNotice, "jenkins update") {
		t.Errorf("go-bin notice still suggests jenkins update:\n%s", goNotice)
	}
}

func TestMaybeNotifyOncePerVersionPer24h(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	info := &UpdateInfo{Available: true, CurrentVersion: "0.2.8", LatestVersion: "0.2.9"}
	notify := func(info *UpdateInfo, at time.Time) string {
		var buf bytes.Buffer
		MaybeNotify(&buf, info, dir, InstallSelf, at)
		return buf.String()
	}

	if out := notify(info, now); !strings.Contains(out, "v0.2.8 -> v0.2.9") {
		t.Fatalf("first run printed %q, want the notice", out)
	}
	if out := notify(info, now.Add(time.Minute)); out != "" {
		t.Errorf("second run within 24h printed %q, want nothing", out)
	}
	if out := notify(info, now.Add(23*time.Hour)); out != "" {
		t.Errorf("run after 23h printed %q, want nothing", out)
	}
	newer := &UpdateInfo{Available: true, CurrentVersion: "0.2.8", LatestVersion: "0.2.10"}
	if out := notify(newer, now.Add(23*time.Hour+time.Minute)); !strings.Contains(out, "v0.2.10") {
		t.Errorf("a newer release printed %q, want the notice", out)
	}
	if out := notify(newer, now.Add(48*time.Hour)); !strings.Contains(out, "v0.2.10") {
		t.Errorf("same release after 24h printed %q, want the notice again", out)
	}
	if out := notify(&UpdateInfo{CurrentVersion: "0.2.10", LatestVersion: "0.2.10"}, now.Add(96*time.Hour)); out != "" {
		t.Errorf("no update printed %q", out)
	}
}

func TestDetectInstallMethod(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	gobin := filepath.Join(root, "gobin")
	gopath := filepath.Join(root, "gopath")
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

	for _, tc := range []struct {
		name string
		exec string
		env  map[string]string
		want string
	}{
		{"GOBIN", filepath.Join(gobin, "jenkins"), map[string]string{"GOBIN": gobin}, InstallGo},
		{"GOPATH bin", filepath.Join(gopath, "bin", "jenkins"), map[string]string{"GOPATH": gopath}, InstallGo},
		{"second GOPATH entry", filepath.Join(gopath, "bin", "jenkins"), map[string]string{"GOPATH": root + string(os.PathListSeparator) + gopath}, InstallGo},
		{"default ~/go/bin", filepath.Join(home, "go", "bin", "jenkins"), nil, InstallGo},
		{"install script dir", filepath.Join(root, "usr", "local", "bin", "jenkins"), map[string]string{"GOBIN": gobin, "GOPATH": gopath}, InstallSelf},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetectInstallMethod(tc.exec, env(tc.env), home, "linux"); got != tc.want {
				t.Errorf("DetectInstallMethod(%s) = %s, want %s", tc.exec, got, tc.want)
			}
		})
	}
}

func TestCheckForUpdateCachesResultForADay(t *testing.T) {
	hits, _ := fakeReleases(t, http.StatusFound, tagURL("v0.2.9"))
	dir := t.TempDir()

	info, err := CheckForUpdate("0.2.8", dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available || info.LatestVersion != "0.2.9" || info.ReleaseURL != ReleaseURL("0.2.9") {
		t.Fatalf("unexpected info %+v", info)
	}
	if _, err := CheckForUpdate("0.2.8", dir, false); err != nil {
		t.Fatal(err)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("GitHub queried %d times, want 1 (second check should use the cache)", n)
	}
	if _, err := CheckForUpdate("0.2.8", dir, true); err != nil {
		t.Fatal(err)
	}
	if n := hits.Load(); n != 2 {
		t.Errorf("forced check did not bypass the cache: %d requests", n)
	}
}

func TestCheckForUpdateCachesFailures(t *testing.T) {
	hits, _ := fakeReleases(t, http.StatusInternalServerError, nil)
	dir := t.TempDir()
	if _, err := CheckForUpdate("0.2.8", dir, false); err == nil {
		t.Fatal("expected an error from a failing GitHub API")
	}
	info, err := CheckForUpdate("0.2.8", dir, false)
	if err != nil || info.Available {
		t.Fatalf("cached failure returned %+v, %v; want no update and no error", info, err)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("GitHub queried %d times after a failure, want 1", n)
	}
}

func TestCheckForUpdateRejectsUnexpectedTag(t *testing.T) {
	fakeReleases(t, http.StatusFound, tagURL("v0.2.9/../../evil"))
	if _, err := CheckForUpdate("0.2.8", t.TempDir(), true); err == nil {
		t.Fatal("expected an unsafe tag to be rejected")
	}
}

func TestDevBuildsNeverQueryGitHub(t *testing.T) {
	hits, _ := fakeReleases(t, http.StatusFound, tagURL("v0.2.9"))
	for _, v := range []string{"dev", "", "2b9e25a"} {
		info, err := CheckForUpdate(v, t.TempDir(), true)
		if err != nil || info.Available {
			t.Errorf("CheckForUpdate(%q) = %+v, %v", v, info, err)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("dev builds queried GitHub %d times", n)
	}
}

func TestCachedUpdateReadsOnlyTheCache(t *testing.T) {
	hits, _ := fakeReleases(t, http.StatusFound, tagURL("v9.9.9"))
	dir := t.TempDir()
	if info := CachedUpdate("0.2.8", dir); info != nil {
		t.Fatalf("empty cache returned %+v", info)
	}
	stale := cacheEntry{LastChecked: time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339), LatestVersion: "0.2.9"}
	data, _ := json.Marshal(stale)
	if err := os.WriteFile(filepath.Join(dir, cacheFile), data, 0o600); err != nil {
		t.Fatal(err)
	}
	info := CachedUpdate("0.2.8", dir)
	if info == nil || info.LatestVersion != "0.2.9" || !info.Available {
		t.Fatalf("CachedUpdate = %+v, want latest 0.2.9 with an update available", info)
	}
	if info := CachedUpdate("0.3.0", dir); info != nil {
		t.Errorf("a cached release older than the running one was reported: %+v", info)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("CachedUpdate queried GitHub %d times", n)
	}
}

func TestRemoveOldBinary(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "jenkins.exe")
	if err := os.WriteFile(exe+".old", []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	RemoveOldBinary("linux", exe)
	if _, err := os.Stat(exe + ".old"); err != nil {
		t.Fatalf("non-Windows run touched %s.old: %v", exe, err)
	}
	RemoveOldBinary("windows", exe)
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Errorf("leftover %s.old was not removed: %v", exe, err)
	}
}
