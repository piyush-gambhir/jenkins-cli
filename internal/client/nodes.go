package client

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Node represents a Jenkins node/agent.
type Node struct {
	DisplayName     string       `json:"displayName"`
	Description     string       `json:"description"`
	Idle            bool         `json:"idle"`
	JNLPAgent       bool         `json:"jnlpAgent"`
	LaunchSupported bool         `json:"launchSupported"`
	ManualLaunchAllowed bool     `json:"manualLaunchAllowed"`
	NumExecutors    int          `json:"numExecutors"`
	Offline         bool         `json:"offline"`
	OfflineCause    *OfflineCause `json:"offlineCause"`
	OfflineCauseReason string    `json:"offlineCauseReason"`
	TemporarilyOffline bool      `json:"temporarilyOffline"`
	MonitorData     json.RawMessage `json:"monitorData"`
	Executors       []Executor   `json:"executors"`
}

// OfflineCause represents why a node is offline.
type OfflineCause struct {
	Class       string `json:"_class"`
	Description string `json:"description"`
}

// Executor represents an executor on a node.
type Executor struct {
	Idle     bool          `json:"idle"`
	Number   int           `json:"number"`
	Progress int           `json:"progress"`
	CurrentExecutable *BuildRef `json:"currentExecutable"`
}

// ComputerResponse wraps the computer (node) list response.
type ComputerResponse struct {
	Computers     []Node `json:"computer"`
	TotalExecutors int   `json:"totalExecutors"`
	BusyExecutors  int   `json:"busyExecutors"`
}

// ListNodes lists all nodes.
func (c *Client) ListNodes() ([]Node, error) {
	data, err := c.Get("/computer", nil)
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	var resp ComputerResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing nodes: %w", err)
	}

	return resp.Computers, nil
}

// GetNode gets details about a specific node.
func (c *Client) GetNode(name string) (*Node, error) {
	path := fmt.Sprintf("/computer/%s", url.PathEscape(name))

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting node: %w", err)
	}

	var node Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("parsing node: %w", err)
	}

	return &node, nil
}

// CreateNode creates a new permanent agent node.
func (c *Client) CreateNode(name string, numExecutors int, remoteFSRoot, labels string) error {
	nodeJSON := fmt.Sprintf(`{
		"name": "%s",
		"nodeDescription": "",
		"numExecutors": %d,
		"remoteFS": "%s",
		"labelString": "%s",
		"mode": "NORMAL",
		"": ["hudson.slaves.JNLPLauncher", "hudson.slaves.RetentionStrategy$Always"],
		"launcher": {"stapler-class": "hudson.slaves.JNLPLauncher", "$class": "hudson.slaves.JNLPLauncher"},
		"retentionStrategy": {"stapler-class": "hudson.slaves.RetentionStrategy$Always", "$class": "hudson.slaves.RetentionStrategy$Always"},
		"nodeProperties": {"stapler-class-bag": "true"},
		"type": "hudson.slaves.DumbSlave"
	}`, name, numExecutors, remoteFSRoot, labels)

	query := url.Values{
		"name": {name},
		"type": {"hudson.slaves.DumbSlave"},
		"json": {nodeJSON},
	}

	_, err := c.PostForm("/computer/doCreateItem", query, nil)
	if err != nil {
		return fmt.Errorf("creating node: %w", err)
	}

	return nil
}

// DeleteNode deletes a node.
func (c *Client) DeleteNode(name string) error {
	path := fmt.Sprintf("/computer/%s/doDelete", url.PathEscape(name))

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("deleting node: %w", err)
	}

	return nil
}

// ToggleOffline takes a node offline or brings it online.
func (c *Client) ToggleOffline(name string, offline bool, message string) error {
	var path string
	if offline {
		path = fmt.Sprintf("/computer/%s/toggleOffline", url.PathEscape(name))
	} else {
		path = fmt.Sprintf("/computer/%s/toggleOffline", url.PathEscape(name))
	}

	query := url.Values{}
	if message != "" {
		query.Set("offlineMessage", message)
	}

	_, err := c.Post(path, query)
	if err != nil {
		return fmt.Errorf("toggling node offline status: %w", err)
	}

	return nil
}

// GetNodeLog gets the agent log for a node.
func (c *Client) GetNodeLog(name string) (string, error) {
	path := fmt.Sprintf("/computer/%s/logText/progressiveText", url.PathEscape(name))

	query := url.Values{"start": {"0"}}
	data, err := c.GetRaw(path, query)
	if err != nil {
		return "", fmt.Errorf("getting node log: %w", err)
	}

	return string(data), nil
}
