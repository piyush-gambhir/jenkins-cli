package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ServerInfo represents Jenkins server information.
type ServerInfo struct {
	Mode            string `json:"mode"`
	NodeDescription string `json:"nodeDescription"`
	NodeName        string `json:"nodeName"`
	NumExecutors    int    `json:"numExecutors"`
	Description     string `json:"description"`
	UseSecurity     bool   `json:"useSecurity"`
	UseCrumbs       bool   `json:"useCrumbs"`
	QuietingDown    bool   `json:"quietingDown"`
	URL             string `json:"url"`
	Views           []View `json:"views"`
	PrimaryView     *View  `json:"primaryView"`
	Slaved          bool   `json:"slaveAgentPort,omitempty"`
}

// GetServerInfo gets Jenkins server information.
func (c *Client) GetServerInfo() (*ServerInfo, error) {
	data, err := c.Get("", nil)
	if err != nil {
		return nil, fmt.Errorf("getting server info: %w", err)
	}

	var info ServerInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parsing server info: %w", err)
	}

	return &info, nil
}

// GetServerVersion returns the Jenkins version from the X-Jenkins header.
func (c *Client) GetServerVersion() (string, error) {
	resp, err := c.doRequest(requestOptions{
		method: "GET",
		path:   "/api/json",
	})
	if err != nil {
		return "", fmt.Errorf("getting server version: %w", err)
	}
	defer resp.Body.Close()

	version := resp.Header.Get("X-Jenkins")
	if version == "" {
		version = "unknown"
	}

	return version, nil
}

// Restart performs an immediate restart of Jenkins.
func (c *Client) Restart() error {
	_, err := c.Post("/restart", nil)
	if err != nil {
		return fmt.Errorf("restarting Jenkins: %w", err)
	}
	return nil
}

// SafeRestart performs a safe restart (waits for running builds).
func (c *Client) SafeRestart() error {
	_, err := c.Post("/safeRestart", nil)
	if err != nil {
		return fmt.Errorf("safe restarting Jenkins: %w", err)
	}
	return nil
}

// QuietDown puts Jenkins into quiet-down mode.
func (c *Client) QuietDown() error {
	_, err := c.Post("/quietDown", nil)
	if err != nil {
		return fmt.Errorf("quieting down: %w", err)
	}
	return nil
}

// CancelQuietDown cancels quiet-down mode.
func (c *Client) CancelQuietDown() error {
	_, err := c.Post("/cancelQuietDown", nil)
	if err != nil {
		return fmt.Errorf("cancelling quiet down: %w", err)
	}
	return nil
}

// RunScript executes a Groovy script on the Jenkins controller.
func (c *Client) RunScript(script string) (string, error) {
	formData := url.Values{
		"script": {script},
	}

	data, err := c.PostForm("/scriptText", nil, formData)
	if err != nil {
		return "", fmt.Errorf("running script: %w", err)
	}

	return string(data), nil
}
