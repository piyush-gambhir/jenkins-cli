package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

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
	SlaveAgentPort  int    `json:"slaveAgentPort,omitempty"`
}

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

func (c *Client) Restart() error {
	_, err := c.Post("/restart", nil)
	if err != nil {
		return fmt.Errorf("restarting Jenkins: %w", err)
	}
	return nil
}

func (c *Client) SafeRestart() error {
	_, err := c.Post("/safeRestart", nil)
	if err != nil {
		return fmt.Errorf("safe restarting Jenkins: %w", err)
	}
	return nil
}

func (c *Client) QuietDown() error {
	_, err := c.Post("/quietDown", nil)
	if err != nil {
		return fmt.Errorf("quieting down: %w", err)
	}
	return nil
}

func (c *Client) CancelQuietDown() error {
	_, err := c.Post("/cancelQuietDown", nil)
	if err != nil {
		return fmt.Errorf("cancelling quiet down: %w", err)
	}
	return nil
}

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
