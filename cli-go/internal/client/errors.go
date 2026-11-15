package client

import (
	"fmt"
	"regexp"
	"strings"
)

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

func parseErrorBody(body string) string {
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

	re := regexp.MustCompile(`<p>(.*?)</p>`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		msg := strings.TrimSpace(matches[1])
		tagRe := regexp.MustCompile(`<[^>]*>`)
		msg = tagRe.ReplaceAllString(msg, "")
		if msg != "" {
			return msg
		}
	}

	re = regexp.MustCompile(`<title>(.*?)</title>`)
	matches = re.FindStringSubmatch(body)
	if len(matches) > 1 {
		msg := strings.TrimSpace(matches[1])
		if msg != "" {
			return msg
		}
	}

	body = strings.TrimSpace(body)
	if len(body) > 200 {
		return body[:200] + "..."
	}
	if body != "" {
		return body
	}
	return ""
}
