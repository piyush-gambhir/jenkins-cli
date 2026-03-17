package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/piyush-gambhir/jenkins-cli/internal/config"
)

// Client is a Jenkins API client.
type Client struct {
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
	cache      crumbCache
	insecure   bool
}

// NewClient creates a new Jenkins API client from a resolved profile.
func NewClient(profile config.Profile) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: profile.Insecure,
		},
	}

	return &Client{
		baseURL:  strings.TrimRight(profile.URL, "/"),
		username: profile.Username,
		token:    profile.Token,
		insecure: profile.Insecure,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// Get performs a GET request. If the path doesn't end with /api/json or contain /api/,
// it appends /api/json automatically unless rawOutput is true.
func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	apiPath := path
	if !strings.Contains(path, "/api/") && !strings.HasSuffix(path, "/config.xml") {
		apiPath = strings.TrimRight(path, "/") + "/api/json"
	}

	resp, err := c.doRequest(requestOptions{
		method: "GET",
		path:   apiPath,
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

// GetRaw performs a raw GET request without appending /api/json.
func (c *Client) GetRaw(path string, query url.Values) ([]byte, error) {
	resp, err := c.doRequest(requestOptions{
		method: "GET",
		path:   path,
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

// GetRawResponse performs a raw GET and returns the response without reading the body.
func (c *Client) GetRawResponse(path string, query url.Values) (*http.Response, error) {
	resp, err := c.doRequest(requestOptions{
		method: "GET",
		path:   path,
		query:  query,
	})
	if err != nil {
		return nil, err
	}

	if err := checkResponse(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

// GetJSON performs a GET request and unmarshals the JSON response.
func (c *Client) GetJSON(path string, query url.Values, v interface{}) error {
	data, err := c.Get(path, query)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Post performs a POST request with no body.
func (c *Client) Post(path string, query url.Values) ([]byte, error) {
	resp, err := c.doRequest(requestOptions{
		method: "POST",
		path:   path,
		query:  query,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

// PostXML performs a POST request with an XML body.
func (c *Client) PostXML(path string, query url.Values, xmlBody string) ([]byte, http.Header, error) {
	resp, err := c.doRequest(requestOptions{
		method:      "POST",
		path:        path,
		query:       query,
		body:        strings.NewReader(xmlBody),
		contentType: "application/xml",
	})
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, nil, err
	}

	body, err := io.ReadAll(resp.Body)
	return body, resp.Header, err
}

// PostForm performs a POST request with form-encoded body.
func (c *Client) PostForm(path string, query url.Values, formData url.Values) ([]byte, error) {
	var body io.Reader
	var contentType string

	if len(formData) > 0 {
		body = strings.NewReader(formData.Encode())
		contentType = "application/x-www-form-urlencoded"
	}

	resp, err := c.doRequest(requestOptions{
		method:      "POST",
		path:        path,
		query:       query,
		body:        body,
		contentType: contentType,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

// PostRaw performs a POST and returns the full response (headers included via http.Response).
func (c *Client) PostRaw(path string, query url.Values) (*http.Response, error) {
	resp, err := c.doRequest(requestOptions{
		method: "POST",
		path:   path,
		query:  query,
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Message:    parseErrorBody(string(body)),
			URL:        resp.Request.URL.String(),
		}
	}

	return resp, nil
}

// BaseURL returns the client's base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// TreeParam is a helper to build ?tree= query values.
func TreeParam(tree string) url.Values {
	if tree == "" {
		return nil
	}
	return url.Values{"tree": {tree}}
}

// DepthParam is a helper to build ?depth= query values.
func DepthParam(depth int) url.Values {
	return url.Values{"depth": {fmt.Sprintf("%d", depth)}}
}

// MergeQuery merges multiple url.Values into one.
func MergeQuery(queries ...url.Values) url.Values {
	result := url.Values{}
	for _, q := range queries {
		for k, vs := range q {
			for _, v := range vs {
				result.Add(k, v)
			}
		}
	}
	return result
}
