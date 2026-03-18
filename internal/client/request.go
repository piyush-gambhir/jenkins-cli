package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// requestOptions configures an HTTP request.
type requestOptions struct {
	method      string
	path        string
	body        io.Reader
	contentType string
	query       url.Values
	headers     map[string]string
	rawOutput   bool // don't expect JSON
}

// buildURL constructs a full URL from the base URL, path, and query parameters.
func (c *Client) buildURL(path string, query url.Values) string {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

// newRequest creates an http.Request with auth and headers set.
func (c *Client) newRequest(opts requestOptions) (*http.Request, error) {
	fullURL := c.buildURL(opts.path, opts.query)

	req, err := http.NewRequest(opts.method, fullURL, opts.body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(c.username, c.token)

	if opts.contentType != "" {
		req.Header.Set("Content-Type", opts.contentType)
	}

	for k, v := range opts.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// doRequest executes an HTTP request and returns the response.
// For POST requests, it automatically injects the CSRF crumb.
func (c *Client) doRequest(opts requestOptions) (*http.Response, error) {
	if strings.ToUpper(opts.method) == "POST" {
		crumb, err := c.ensureCrumb()
		if err != nil {
			return nil, fmt.Errorf("getting crumb: %w", err)
		}
		if crumb != nil {
			if opts.headers == nil {
				opts.headers = make(map[string]string)
			}
			opts.headers[crumb.RequestField] = crumb.Value
		}
	}

	req, err := c.newRequest(opts)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

// checkResponse reads the response body and returns an APIError if status >= 400.
func checkResponse(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading error response body: %w", err)
	}
	return &APIError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Message:    parseErrorBody(string(body)),
		URL:        resp.Request.URL.String(),
	}
}
