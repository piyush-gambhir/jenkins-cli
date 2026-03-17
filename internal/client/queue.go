package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// QueueItem represents an item in the Jenkins build queue.
type QueueItem struct {
	ID         int            `json:"id"`
	Task       QueueTask      `json:"task"`
	Why        string         `json:"why"`
	Blocked    bool           `json:"blocked"`
	Buildable  bool           `json:"buildable"`
	Stuck      bool           `json:"stuck"`
	Cancelled  bool           `json:"cancelled"`
	InQueueSince int64        `json:"inQueueSince"`
	Executable QueueExecutable `json:"executable"`
	Actions    []json.RawMessage `json:"actions"`
}

// QueueTask represents the task associated with a queue item.
type QueueTask struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

// QueueExecutable is the build that was created from a queue item.
type QueueExecutable struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

// QueueResponse wraps the queue list response.
type QueueResponse struct {
	Items []QueueItem `json:"items"`
}

// ListQueue lists all items in the build queue.
func (c *Client) ListQueue() ([]QueueItem, error) {
	data, err := c.Get("/queue", nil)
	if err != nil {
		return nil, fmt.Errorf("listing queue: %w", err)
	}

	var resp QueueResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing queue: %w", err)
	}

	return resp.Items, nil
}

// GetQueueItem gets details about a queue item.
func (c *Client) GetQueueItem(id int) (*QueueItem, error) {
	path := fmt.Sprintf("/queue/item/%d", id)

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting queue item: %w", err)
	}

	var item QueueItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("parsing queue item: %w", err)
	}

	return &item, nil
}

// CancelQueueItem cancels a queued build.
func (c *Client) CancelQueueItem(id int) error {
	path := fmt.Sprintf("/queue/cancelItem")
	query := map[string][]string{"id": {fmt.Sprintf("%d", id)}}

	_, err := c.Post(path, query)
	if err != nil {
		return fmt.Errorf("cancelling queue item: %w", err)
	}

	return nil
}

// WaitForBuild waits for a queue item to get a build number.
func (c *Client) WaitForBuild(queueID int, timeout time.Duration) (*BuildRef, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		item, err := c.GetQueueItem(queueID)
		if err != nil {
			return nil, err
		}

		if item.Executable.Number > 0 {
			return &BuildRef{
				Number: item.Executable.Number,
				URL:    item.Executable.URL,
			}, nil
		}

		if item.Cancelled {
			return nil, fmt.Errorf("queue item %d was cancelled", queueID)
		}

		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("timed out waiting for build from queue item %d", queueID)
}
