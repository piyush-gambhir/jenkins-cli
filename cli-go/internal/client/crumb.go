package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Crumb struct {
	RequestField string `json:"crumbRequestField"`
	Value        string `json:"crumb"`
	fetchedAt    time.Time
}

const crumbTTL = 5 * time.Minute

type crumbCache struct {
	mu    sync.Mutex
	crumb *Crumb
}

func (cc *crumbCache) get() *Crumb {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.crumb == nil {
		return nil
	}
	if time.Since(cc.crumb.fetchedAt) > crumbTTL {
		cc.crumb = nil
		return nil
	}
	return cc.crumb
}

func (cc *crumbCache) set(c *Crumb) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.crumb = c
}

func (c *Client) fetchCrumb() (*Crumb, error) {
	u := c.baseURL + "/crumbIssuer/api/json"

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating crumb request: %w", err)
	}
	req.SetBasicAuth(c.username, c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching crumb: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, err := readAllLimited(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading crumb error response: %w", err)
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Message:    parseErrorBody(string(body)),
			URL:        u,
		}
	}

	var crumb Crumb
	if err := json.NewDecoder(resp.Body).Decode(&crumb); err != nil {
		return nil, fmt.Errorf("decoding crumb response: %w", err)
	}
	crumb.fetchedAt = time.Now()

	return &crumb, nil
}

func (c *Client) ensureCrumb() (*Crumb, error) {
	if crumb := c.cache.get(); crumb != nil {
		return crumb, nil
	}

	crumb, err := c.fetchCrumb()
	if err != nil {
		return nil, err
	}
	if crumb != nil {
		c.cache.set(crumb)
	}
	return crumb, nil
}
