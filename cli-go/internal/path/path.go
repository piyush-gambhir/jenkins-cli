package path

import (
	"net/url"
	"strings"
)

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
			i++
		}
	}
	return strings.Join(segments, "/")
}

func BuildJobPath(userPath, suffix string) string {
	jp := ToJenkinsPath(userPath)
	if suffix != "" {
		return jp + suffix
	}
	return jp
}
