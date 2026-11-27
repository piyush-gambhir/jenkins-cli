package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type requestOptions struct {
	ctx         context.Context
	method      string
	path        string
	body        io.Reader
	contentType string
	query       url.Values
	headers     map[string]string
	noRedirect  bool
	stopAt303   bool
}

func (c *Client) buildURL(path string, query url.Values) string {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

func (c *Client) newRequest(opts requestOptions) (*http.Request, error) {
	fullURL := c.buildURL(opts.path, opts.query)

	ctx := opts.ctx
	if ctx == nil {
		ctx = c.ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, opts.method, fullURL, opts.body)
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

	httpClient := c.httpClient
	if opts.noRedirect {
		noRedirectClient := *c.httpClient
		noRedirectClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		httpClient = &noRedirectClient
	} else if opts.stopAt303 {
		stopClient := *c.httpClient
		next := c.httpClient.CheckRedirect
		stopClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if req.Response != nil && req.Response.StatusCode == http.StatusSeeOther {
				return http.ErrUseLastResponse
			}
			if next != nil {
				return next(req, via)
			}
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		}
		httpClient = &stopClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}
	body, err := readAllLimited(resp.Body)
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
