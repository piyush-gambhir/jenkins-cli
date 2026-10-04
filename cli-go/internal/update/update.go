// Package update checks GitHub Releases for new versions of the CLI, prints
// the update notice, and installs releases in place (see install.go).
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// Repo is the GitHub repository that publishes releases.
	Repo = "piyush-gambhir/jenkins-cli"
	// Binary is the executable name (without .exe).
	Binary = "jenkins"
	// SourceUpdateCommand updates a binary built from source with
	// `make install`, which installs into the Go bin directory. A plain
	// `go install` of the module would name the binary cli-go and build a dev
	// version, so it is not suggested.
	SourceUpdateCommand = "git pull && make install (in your jenkins-cli/cli-go checkout)"
	// EnvPrefix is the CLI's environment variable prefix.
	EnvPrefix = "JENKINS"

	// InstallSelf and InstallGo are the values of install_method.
	InstallSelf = "self"
	InstallGo   = "go"

	cacheTTL     = 24 * time.Hour
	noticeTTL    = 24 * time.Hour
	cacheFile    = "update-check.json"
	checkTimeout = 3 * time.Second
	maxAPIBytes  = 1 << 20
)

// apiBaseURL is the GitHub API root; tests point it at an httptest server.
var apiBaseURL = "https://api.github.com"

// releaseVersion accepts X.Y.Z with an optional pre-release suffix. Tags that
// do not match are rejected because they end up in URLs and terminal output.
var releaseVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$`)

// UpdateInfo holds the result of an update check.
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	PublishedAt    string
}

// cacheEntry is the on-disk JSON format of update-check.json.
type cacheEntry struct {
	LastChecked     string `json:"last_checked"`
	LatestVersion   string `json:"latest_version,omitempty"`
	ReleaseURL      string `json:"release_url,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
	CheckError      string `json:"check_error,omitempty"`
	NotifiedVersion string `json:"notified_version,omitempty"`
	NotifiedAt      string `json:"notified_at,omitempty"`
}

type githubRelease struct {
	TagName     string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
}

// NormalizeVersion trims spaces and a leading "v" ("v0.2.8" -> "0.2.8").
func NormalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// IsReleaseVersion reports whether v is a semver release build. Dev builds
// ("dev", empty, a bare commit hash) never check for or install updates.
func IsReleaseVersion(v string) bool {
	return parseSemver(NormalizeVersion(v)) != nil
}

// ReleaseURL returns the release notes page for version.
func ReleaseURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", Repo, NormalizeVersion(version))
}

// CheckForUpdate returns the latest release, using the 24h cache in configDir
// unless force is set. Failed checks are cached too, so a broken network does
// not trigger a request on every command. Dev builds never touch the network.
func CheckForUpdate(currentVersion, configDir string, force bool) (*UpdateInfo, error) {
	current := NormalizeVersion(currentVersion)
	if !IsReleaseVersion(current) {
		return &UpdateInfo{CurrentVersion: current}, nil
	}

	if !force {
		if info, ok := FreshCache(current, configDir); ok {
			return info, nil
		}
	}
	entry, err := readEntry(configDir)
	if err != nil {
		entry = &cacheEntry{}
	}

	latest, publishedAt, err := fetchLatest(current)
	entry.LastChecked = time.Now().UTC().Format(time.RFC3339)
	if err != nil {
		// Keep the last known release; only record the failure time.
		entry.CheckError = err.Error()
		writeEntry(configDir, entry)
		return nil, err
	}
	entry.CheckError = ""
	entry.LatestVersion = latest
	entry.ReleaseURL = ReleaseURL(latest)
	entry.PublishedAt = publishedAt
	writeEntry(configDir, entry)
	return infoFromEntry(current, entry), nil
}

// FreshCache returns the cached check result when it is less than 24h old,
// including a cached failure (which keeps the last known release). It never
// touches the network.
func FreshCache(currentVersion, configDir string) (*UpdateInfo, bool) {
	current := NormalizeVersion(currentVersion)
	entry, err := readEntry(configDir)
	if err != nil {
		return nil, false
	}
	checked, err := time.Parse(time.RFC3339, entry.LastChecked)
	if err != nil {
		return nil, false
	}
	if age := time.Since(checked); age < 0 || age >= cacheTTL {
		return nil, false
	}
	return infoFromEntry(current, entry), true
}

// CachedUpdate returns what the cache knows about the latest release, without
// any network access. It returns nil when nothing valid is cached, or when the
// cached release is older than the running one (the cache predates an update).
func CachedUpdate(currentVersion, configDir string) *UpdateInfo {
	current := NormalizeVersion(currentVersion)
	if !IsReleaseVersion(current) {
		return nil
	}
	entry, err := readEntry(configDir)
	if err != nil || !releaseVersion.MatchString(entry.LatestVersion) || isNewer(current, entry.LatestVersion) {
		return nil
	}
	return infoFromEntry(current, entry)
}

// ClearCache removes the update cache (after a successful update).
func ClearCache(configDir string) {
	_ = os.Remove(filepath.Join(configDir, cacheFile))
}

func infoFromEntry(current string, entry *cacheEntry) *UpdateInfo {
	info := &UpdateInfo{CurrentVersion: current}
	if !releaseVersion.MatchString(entry.LatestVersion) {
		return info
	}
	info.LatestVersion = entry.LatestVersion
	info.ReleaseURL = ReleaseURL(entry.LatestVersion)
	info.PublishedAt = entry.PublishedAt
	info.Available = isNewer(entry.LatestVersion, current)
	return info
}

func fetchLatest(current string) (version, publishedAt string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()
	url := fmt.Sprintf("%s/repos/%s/releases/latest", apiBaseURL, Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "jenkins-cli/"+current)

	resp, err := (&http.Client{Timeout: checkTimeout}).Do(req)
	if err != nil {
		return "", "", fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return "", "", errors.New("GitHub API rate limit exceeded; try again later")
		}
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIBytes))
	if err != nil {
		return "", "", fmt.Errorf("reading response: %w", err)
	}
	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", fmt.Errorf("parsing release info: %w", err)
	}
	version = NormalizeVersion(release.TagName)
	if !releaseVersion.MatchString(version) {
		return "", "", fmt.Errorf("unexpected release tag %q", release.TagName)
	}
	return version, release.PublishedAt, nil
}

func readEntry(configDir string) (*cacheEntry, error) {
	data, err := os.ReadFile(filepath.Join(configDir, cacheFile))
	if err != nil {
		return nil, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// writeEntry saves the cache atomically (temp file + rename) so a process
// that exits mid-write never leaves a torn file. Errors are ignored: the
// cache is an optimization.
func writeEntry(configDir string, entry *cacheEntry) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(configDir, ".update-check-*")
	if err != nil {
		return
	}
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil || cerr != nil || os.Rename(tmp.Name(), filepath.Join(configDir, cacheFile)) != nil {
		_ = os.Remove(tmp.Name())
	}
}

// NotifierEnabled reports whether the background check should run at all.
// When it returns false there is no network access and no notice. topCommand
// is the name of the command directly under the root ("job" for
// `jenkins job list`).
func NotifierEnabled(getenv func(string) string, stderrIsTerminal, quiet bool, currentVersion, topCommand string) bool {
	if !stderrIsTerminal || quiet || !IsReleaseVersion(currentVersion) {
		return false
	}
	for _, name := range []string{"CI", EnvPrefix + "_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		if getenv(name) != "" {
			return false
		}
	}
	switch topCommand {
	case "update", "version", "completion", "help":
		return false
	}
	return !strings.HasPrefix(topCommand, "__complete")
}

// Notice returns the update notice text, starting with a blank line.
func Notice(info *UpdateInfo, installMethod string) string {
	updateCmd := Binary + " update"
	if installMethod == InstallGo {
		updateCmd = SourceUpdateCommand
	}
	return fmt.Sprintf("\nA new version of %s is available: v%s -> v%s\nUpdate with: %s\nRelease notes: %s\n",
		Binary, info.CurrentVersion, info.LatestVersion, updateCmd, ReleaseURL(info.LatestVersion))
}

// MaybeNotify prints the notice for info unless it was already shown for the
// same latest version within the last 24h, and records that it was shown.
func MaybeNotify(w io.Writer, info *UpdateInfo, configDir, installMethod string, now time.Time) bool {
	if info == nil || !info.Available || !releaseVersion.MatchString(info.LatestVersion) {
		return false
	}
	entry, err := readEntry(configDir)
	if err != nil {
		entry = &cacheEntry{}
	}
	if entry.NotifiedVersion == info.LatestVersion {
		if at, err := time.Parse(time.RFC3339, entry.NotifiedAt); err == nil {
			if age := now.Sub(at); age >= 0 && age < noticeTTL {
				return false
			}
		}
	}
	fmt.Fprint(w, Notice(info, installMethod))
	entry.NotifiedVersion = info.LatestVersion
	entry.NotifiedAt = now.UTC().Format(time.RFC3339)
	writeEntry(configDir, entry)
	return true
}

// ExecutablePath returns the running executable with symlinks resolved.
func ExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding current binary: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolving binary path: %w", err)
	}
	return resolved, nil
}

// CurrentInstallMethod reports how the running binary was installed.
func CurrentInstallMethod() string {
	path, err := ExecutablePath()
	if err != nil {
		return InstallSelf
	}
	home, _ := os.UserHomeDir()
	return DetectInstallMethod(path, os.Getenv, home, runtime.GOOS)
}

// DetectInstallMethod returns InstallGo when execPath is in a Go bin
// directory ($GOBIN, $GOPATH/bin for each GOPATH entry, or ~/go/bin), where
// the binary was built from source, and InstallSelf otherwise.
func DetectInstallMethod(execPath string, getenv func(string) string, home, goos string) string {
	var dirs []string
	if gobin := getenv("GOBIN"); gobin != "" {
		dirs = append(dirs, gobin)
	}
	for _, p := range filepath.SplitList(getenv("GOPATH")) {
		if p != "" {
			dirs = append(dirs, filepath.Join(p, "bin"))
		}
	}
	if home != "" {
		dirs = append(dirs, filepath.Join(home, "go", "bin"))
	}
	execDir := canonicalDir(filepath.Dir(execPath))
	for _, dir := range dirs {
		d := canonicalDir(dir)
		if d == execDir || (goos == "windows" && strings.EqualFold(d, execDir)) {
			return InstallGo
		}
	}
	return InstallSelf
}

func canonicalDir(dir string) string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return filepath.Clean(dir)
}

// RemoveOldBinary deletes the <bin>.exe.old left behind by a Windows update.
// The previous executable cannot be deleted while it is running, so the next
// start cleans it up. Best effort.
func RemoveOldBinary(goos, execPath string) {
	if goos != "windows" || execPath == "" {
		return
	}
	if err := os.Remove(execPath + ".old"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}
}

// isNewer returns true if latest is a higher semver than current.
func isNewer(latest, current string) bool {
	latestParts := parseSemver(latest)
	currentParts := parseSemver(current)
	if latestParts == nil || currentParts == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// parseSemver parses "X.Y.Z" or "vX.Y.Z" (ignoring a "-suffix") into
// [major, minor, patch]. Returns nil if parsing fails.
func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil
	}
	result := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil
		}
		result[i] = n
	}
	return result
}
