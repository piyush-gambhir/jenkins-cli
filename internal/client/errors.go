package client

import (
	"fmt"
	"regexp"
	"strings"
)

// APIError represents an error response from the Jenkins API.
type APIError struct {
	StatusCode int
	Status     string
	Message    string
	URL        string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("jenkins API error: %s (status %d) — %s", e.Status, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("jenkins API error: %s (status %d) url=%s", e.Status, e.StatusCode, e.URL)
}

// parseErrorBody attempts to extract a meaningful error message from Jenkins HTML error pages.
func parseErrorBody(body string) string {
	// Try to find error message in <h1> or <h2> tags
	for _, tag := range []string{"h1", "h2"} {
		re := regexp.MustCompile(fmt.Sprintf(`<%s>(.*?)</%s>`, tag, tag))
		matches := re.FindStringSubmatch(body)
		if len(matches) > 1 {
			msg := strings.TrimSpace(matches[1])
			if msg != "" && !strings.Contains(strings.ToLower(msg), "error") {
				return msg
			}
			if msg != "" {
				return msg
			}
		}
	}

	// Try to find <p> with error text
	re := regexp.MustCompile(`<p>(.*?)</p>`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		msg := strings.TrimSpace(matches[1])
		// strip HTML tags from inner content
		tagRe := regexp.MustCompile(`<[^>]*>`)
		msg = tagRe.ReplaceAllString(msg, "")
		if msg != "" {
			return msg
		}
	}

	// Try to find the <title> tag
	re = regexp.MustCompile(`<title>(.*?)</title>`)
	matches = re.FindStringSubmatch(body)
	if len(matches) > 1 {
		msg := strings.TrimSpace(matches[1])
		if msg != "" {
			return msg
		}
	}

	// Truncate raw body if nothing meaningful found
	body = strings.TrimSpace(body)
	if len(body) > 200 {
		return body[:200] + "..."
	}
	if body != "" {
		return body
	}
	return ""
}
