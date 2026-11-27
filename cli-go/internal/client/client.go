package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
)

var sensitiveHeaders = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
}

func redactAuthHeaders(headers http.Header) http.Header {
	redacted := make(http.Header)
	for key, values := range headers {
		if sensitiveHeaders[http.CanonicalHeaderKey(key)] {
			redacted[key] = []string{"[REDACTED]"}
		} else {
			redacted[key] = values
		}
	}
	return redacted
}

type verboseTransport struct {
	transport http.RoundTripper
}

func (t *verboseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "--> %s %s\n", req.Method, req.URL)
	for key, values := range redactAuthHeaders(req.Header) {
		for _, v := range values {
			fmt.Fprintf(os.Stderr, "    %s: %s\n", key, v)
		}
	}
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "<-- ERROR: %v (%v)\n", err, time.Since(start))
		return resp, err
	}
	fmt.Fprintf(os.Stderr, "<-- %d %s (%v)\n", resp.StatusCode, resp.Status, time.Since(start))
	return resp, err
}

type Client struct {
	ctx        context.Context
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
	cache      crumbCache
	insecure   bool
}

func NewClient(profile config.Profile, verbose ...bool) *Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: profile.Insecure,
		},
	}

	var rt http.RoundTripper = transport
	if len(verbose) > 0 && verbose[0] {
		rt = &verboseTransport{transport: transport}
	}

	jar, _ := cookiejar.New(nil)

	return &Client{
		ctx:      context.Background(),
		baseURL:  strings.TrimRight(profile.URL, "/"),
		username: profile.Username,
		token:    profile.Token,
		insecure: profile.Insecure,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: rt,
			Jar:       jar,
		},
	}
}

func (c *Client) SetContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
}

func (c *Client) sleep(d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}

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

	return readAllLimited(resp.Body)
}

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

	return readAllLimited(resp.Body)
}

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

func (c *Client) GetJSON(path string, query url.Values, v interface{}) error {
	data, err := c.Get(path, query)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

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

	return readAllLimited(resp.Body)
}

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

	body, err := readAllLimited(resp.Body)
	return body, resp.Header, err
}

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

	return readAllLimited(resp.Body)
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func TreeParam(tree string) url.Values {
	if tree == "" {
		return nil
	}
	return url.Values{"tree": {tree}}
}

func DepthParam(depth int) url.Values {
	return url.Values{"depth": {fmt.Sprintf("%d", depth)}}
}

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
