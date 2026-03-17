package path

import (
	"net/url"
	"strings"
)

// ToJenkinsPath converts a user-friendly path like "folder1/folder2/job"
// to a Jenkins API path like "/job/folder1/job/folder2/job/job".
func ToJenkinsPath(userPath string) string {
	if userPath == "" {
		return ""
	}
	userPath = strings.Trim(userPath, "/")
	parts := strings.Split(userPath, "/")
	var segments []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		segments = append(segments, "job", url.PathEscape(p))
	}
	return "/" + strings.Join(segments, "/")
}

// FromJenkinsPath converts a Jenkins API path like "/job/folder1/job/folder2/job/job"
// back to a user-friendly path like "folder1/folder2/job".
func FromJenkinsPath(jenkinsPath string) string {
	jenkinsPath = strings.Trim(jenkinsPath, "/")
	parts := strings.Split(jenkinsPath, "/")
	var segments []string
	for i := 0; i < len(parts); i++ {
		if parts[i] == "job" && i+1 < len(parts) {
			decoded, err := url.PathUnescape(parts[i+1])
			if err != nil {
				decoded = parts[i+1]
			}
			segments = append(segments, decoded)
			i++ // skip next
		}
	}
	return strings.Join(segments, "/")
}

// BuildJobPath creates a full Jenkins API path for a job endpoint.
// Example: BuildJobPath("folder/job", "/build") -> "/job/folder/job/job/build"
func BuildJobPath(userPath, suffix string) string {
	jp := ToJenkinsPath(userPath)
	if suffix != "" {
		return jp + suffix
	}
	return jp
}
