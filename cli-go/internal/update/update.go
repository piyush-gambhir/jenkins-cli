package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	Repo                = "piyush-gambhir/jenkins-cli"
	Binary              = "jenkins"
	SourceUpdateCommand = "git pull && make install (in your jenkins-cli/cli-go checkout)"
	EnvPrefix           = "JENKINS"

	InstallSelf = "self"
	InstallGo   = "go"

	cacheTTL  = 24 * time.Hour
	noticeTTL = 24 * time.Hour
	cacheFile = "update-check.json"

	backgroundTimeout = 3 * time.Second
	forcedTimeout     = 15 * time.Second
)

var releasesBaseURL = "https://github.com"

var releaseVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$`)

type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
}

type cacheEntry struct {
	LastChecked     string `json:"last_checked"`
	LatestVersion   string `json:"latest_version,omitempty"`
	ReleaseURL      string `json:"release_url,omitempty"`
	CheckError      string `json:"check_error,omitempty"`
	NotifiedVersion string `json:"notified_version,omitempty"`
	NotifiedAt      string `json:"notified_at,omitempty"`
}

func NormalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func IsReleaseVersion(v string) bool {
	return parseSemver(NormalizeVersion(v)) != nil
}

func ReleaseURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", Repo, NormalizeVersion(version))
}

func CheckForUpdate(currentVersion, configDir string, force bool) (*UpdateInfo, error) {
	current := NormalizeVersion(currentVersion)
	if !IsReleaseVersion(current) {
		return &UpdateInfo{CurrentVersion: current}, nil
	}

	if !force {
		if info, ok := FreshCache(current, configDir); ok {
			return info, nil
		}
		recordAttempt(configDir)
	}
	timeout := backgroundTimeout
	if force {
		timeout = forcedTimeout
	}
	latest, err := fetchLatest(current, timeout)

	entry, rerr := readEntry(configDir)
	if rerr != nil {
		entry = &cacheEntry{}
	}
	entry.LastChecked = time.Now().UTC().Format(time.RFC3339)
	if err != nil {
		entry.CheckError = err.Error()
		writeEntry(configDir, entry)
		return nil, err
	}
	entry.CheckError = ""
	entry.LatestVersion = latest
	entry.ReleaseURL = ReleaseURL(latest)
	writeEntry(configDir, entry)
	return infoFromEntry(current, entry), nil
}

func recordAttempt(configDir string) {
	entry, err := readEntry(configDir)
	if err != nil {
		entry = &cacheEntry{}
	}
	entry.LastChecked = time.Now().UTC().Format(time.RFC3339)
	writeEntry(configDir, entry)
}

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

func infoFromEntry(current string, entry *cacheEntry) *UpdateInfo {
	info := &UpdateInfo{CurrentVersion: current}
	if !releaseVersion.MatchString(entry.LatestVersion) {
		return info
	}
	info.LatestVersion = entry.LatestVersion
	info.ReleaseURL = ReleaseURL(entry.LatestVersion)
	info.Available = isNewer(entry.LatestVersion, current)
	return info
}

func fetchLatest(current string, timeout time.Duration) (string, error) {
	base, err := url.Parse(releasesBaseURL)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesBaseURL+"/"+Repo+"/releases/latest", nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "jenkins-cli/"+current)
	client := &http.Client{
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("checking for updates: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
	default:
		return "", fmt.Errorf("checking for updates: %s returned status %d, not a redirect to the latest release", req.URL.Redacted(), resp.StatusCode)
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return "", errors.New("checking for updates: the latest-release redirect has no Location header")
	}
	loc, err := req.URL.Parse(location)
	if err != nil {
		return "", fmt.Errorf("checking for updates: invalid Location %q: %w", location, err)
	}
	if loc.Scheme != base.Scheme || !strings.EqualFold(loc.Host, base.Host) {
		return "", fmt.Errorf("checking for updates: the latest-release redirect points to another host (%s)", loc.Redacted())
	}
	prefix := "/" + Repo + "/releases/tag/"
	if len(loc.Path) <= len(prefix) || !strings.EqualFold(loc.Path[:len(prefix)], prefix) {
		return "", fmt.Errorf("checking for updates: no release tag in redirect to %s", loc.Redacted())
	}
	tag := loc.Path[len(prefix):]
	if !strings.HasPrefix(tag, "v") || !releaseVersion.MatchString(tag[1:]) {
		return "", fmt.Errorf("checking for updates: latest release tag %q is not a semver v tag", tag)
	}
	return tag[1:], nil
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

func Notice(info *UpdateInfo, installMethod string) string {
	updateCmd := Binary + " update"
	if installMethod == InstallGo {
		updateCmd = SourceUpdateCommand
	}
	return fmt.Sprintf("\nA new version of %s is available: v%s -> v%s\nUpdate with: %s\nRelease notes: %s\n",
		Binary, info.CurrentVersion, info.LatestVersion, updateCmd, ReleaseURL(info.LatestVersion))
}

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

func CurrentInstallMethod() string {
	path, err := ExecutablePath()
	if err != nil {
		return InstallSelf
	}
	home, _ := os.UserHomeDir()
	return DetectInstallMethod(path, os.Getenv, home, runtime.GOOS)
}

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

func RemoveOldBinary(goos, execPath string) {
	if goos != "windows" || execPath == "" {
		return
	}
	if err := os.Remove(execPath + ".old"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}
}

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
