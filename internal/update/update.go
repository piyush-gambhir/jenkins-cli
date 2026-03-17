package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	cacheTTL  = 24 * time.Hour
	cacheFile = "update-check.json"
)

// UpdateInfo holds the result of an update check.
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	PublishedAt    string
}

// cacheEntry is the on-disk JSON format for the update check cache.
type cacheEntry struct {
	LastChecked   string `json:"last_checked"`
	LatestVersion string `json:"latest_version"`
	ReleaseURL    string `json:"release_url"`
	PublishedAt   string `json:"published_at,omitempty"`
}

// githubRelease represents the relevant fields from the GitHub Releases API.
type githubRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// CheckForUpdate checks GitHub for a newer release. Uses a 24h cache stored
// in configDir. Set forceCheck to true to bypass the cache.
func CheckForUpdate(currentVersion, repo, configDir string, forceCheck bool) (*UpdateInfo, error) {
	// Skip check for dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return &UpdateInfo{CurrentVersion: currentVersion}, nil
	}

	// Try cache first (unless forced)
	if !forceCheck {
		if info, err := readCache(currentVersion, configDir); err == nil && info != nil {
			return info, nil
		}
	}

	// Fetch latest release from GitHub
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "jenkins-cli/"+currentVersion)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("parsing release info: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Write cache
	writeCache(configDir, cacheEntry{
		LastChecked:   time.Now().UTC().Format(time.RFC3339),
		LatestVersion: latestVersion,
		ReleaseURL:    release.HTMLURL,
		PublishedAt:   release.PublishedAt,
	})

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseURL:     release.HTMLURL,
		PublishedAt:    release.PublishedAt,
		Available:      isNewer(latestVersion, currentVersion),
	}

	return info, nil
}

// PrintUpdateNotice prints a colored notice to the given writer if an update
// is available.
func PrintUpdateNotice(w io.Writer, info *UpdateInfo) {
	if info == nil || !info.Available {
		return
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "A new version of jenkins is available: v%s \u2192 v%s\n",
		info.CurrentVersion, info.LatestVersion)
	fmt.Fprintf(w, "Run `jenkins update` to update, or download from:\n")
	fmt.Fprintf(w, "%s\n", info.ReleaseURL)
}

// readCache reads the cached update check result. Returns nil if the cache is
// expired or doesn't exist.
func readCache(currentVersion, configDir string) (*UpdateInfo, error) {
	path := filepath.Join(configDir, cacheFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	lastChecked, err := time.Parse(time.RFC3339, entry.LastChecked)
	if err != nil {
		return nil, err
	}

	if time.Since(lastChecked) > cacheTTL {
		return nil, fmt.Errorf("cache expired")
	}

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  entry.LatestVersion,
		ReleaseURL:     entry.ReleaseURL,
		PublishedAt:    entry.PublishedAt,
		Available:      isNewer(entry.LatestVersion, currentVersion),
	}

	return info, nil
}

// writeCache persists the update check result to disk.
func writeCache(configDir string, entry cacheEntry) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}

	path := filepath.Join(configDir, cacheFile)
	_ = os.WriteFile(path, data, 0o600)
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

// parseSemver parses a "X.Y.Z" or "vX.Y.Z" string into [major, minor, patch].
// Returns nil if parsing fails.
func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")

	// Strip any pre-release suffix (e.g., "-rc1")
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
		if err != nil {
			return nil
		}
		result[i] = n
	}
	return result
}
