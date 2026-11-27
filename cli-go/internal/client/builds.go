package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	jpath "github.com/piyush-gambhir/jenkins-cli/cli-go/internal/path"
)

type Build struct {
	Number            int               `json:"number"`
	URL               string            `json:"url"`
	Result            string            `json:"result"`
	Building          bool              `json:"building"`
	Timestamp         int64             `json:"timestamp"`
	Duration          int64             `json:"duration"`
	EstimatedDuration int64             `json:"estimatedDuration"`
	DisplayName       string            `json:"displayName"`
	Description       string            `json:"description"`
	FullDisplayName   string            `json:"fullDisplayName"`
	ID                string            `json:"id"`
	QueueID           int               `json:"queueId"`
	Artifacts         []Artifact        `json:"artifacts"`
	Actions           []json.RawMessage `json:"actions"`
	ChangeSet         *ChangeSet        `json:"changeSet,omitempty"`
	ChangeSets        []ChangeSet       `json:"changeSets,omitempty"`
}

type Artifact struct {
	DisplayPath  string `json:"displayPath"`
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
}

type ChangeSet struct {
	Items []ChangeItem `json:"items"`
	Kind  string       `json:"kind"`
}

type ChangeItem struct {
	CommitID  string `json:"commitId"`
	Timestamp int64  `json:"timestamp"`
	Author    Author `json:"author"`
	Message   string `json:"msg"`
}

type Author struct {
	FullName string `json:"fullName"`
}

type BuildListResponse struct {
	Builds    []Build `json:"builds"`
	AllBuilds []Build `json:"allBuilds"`
}

type TestReport struct {
	FailCount  int         `json:"failCount"`
	PassCount  int         `json:"passCount"`
	SkipCount  int         `json:"skipCount"`
	TotalCount int         `json:"totalCount"`
	Duration   float64     `json:"duration"`
	Suites     []TestSuite `json:"suites"`
}

type TestSuite struct {
	Name     string     `json:"name"`
	Duration float64    `json:"duration"`
	Cases    []TestCase `json:"cases"`
}

type TestCase struct {
	ClassName string  `json:"className"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Duration  float64 `json:"duration"`
	ErrorMsg  string  `json:"errorDetails"`
}

type EnvVars struct {
	EnvMap map[string]string `json:"envMap"`
}

type EnvVarsWrapper struct {
	EnvMap map[string]string `json:"envMap"`
}

type PipelineStage struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	StartTimeMillis     int64  `json:"startTimeMillis"`
	DurationMillis      int64  `json:"durationMillis"`
	PauseDurationMillis int64  `json:"pauseDurationMillis"`
}

type PipelineRun struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	StartTimeMillis int64           `json:"startTimeMillis"`
	DurationMillis  int64           `json:"durationMillis"`
	Stages          []PipelineStage `json:"stages"`
}

func (c *Client) ListBuilds(jobPath string, limit int) ([]Build, error) {
	path := jpath.ToJenkinsPath(jobPath)

	field := "builds"
	if limit > 100 {
		field = "allBuilds"
	}
	tree := fmt.Sprintf("%s[number,url,result,building,timestamp,duration,displayName]{0,%d}", field, limit)
	query := TreeParam(tree)

	data, err := c.Get(path, query)
	if err != nil {
		return nil, fmt.Errorf("listing builds: %w", err)
	}

	var resp BuildListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing builds: %w", err)
	}

	if field == "allBuilds" {
		return resp.AllBuilds, nil
	}
	return resp.Builds, nil
}

func (c *Client) GetBuild(jobPath string, number int) (*Build, error) {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d", number)

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting build: %w", err)
	}

	var build Build
	if err := json.Unmarshal(data, &build); err != nil {
		return nil, fmt.Errorf("parsing build: %w", err)
	}

	return &build, nil
}

func (c *Client) GetBuildLog(jobPath string, number int) (string, error) {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/consoleText", number)

	data, err := c.GetRaw(path, nil)
	if err != nil {
		return "", fmt.Errorf("getting build log: %w", err)
	}

	return string(data), nil
}

func (c *Client) StreamBuildLog(jobPath string, number int, writer io.Writer, ctx ...context.Context) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/logText/progressiveText", number)
	start := "0"

	var streamCtx context.Context
	if len(ctx) > 0 && ctx[0] != nil {
		streamCtx = ctx[0]
	}

	for {
		if streamCtx != nil {
			select {
			case <-streamCtx.Done():
				return nil
			default:
			}
		}

		query := url.Values{"start": {start}}

		resp, err := c.doRequest(requestOptions{
			ctx:    streamCtx,
			method: "GET",
			path:   path,
			query:  query,
		})
		if err != nil {
			return fmt.Errorf("streaming log: %w", err)
		}

		if err := checkResponse(resp); err != nil {
			resp.Body.Close()
			return fmt.Errorf("streaming log: %w", err)
		}
		if _, err := io.Copy(writer, resp.Body); err != nil {
			resp.Body.Close()
			return fmt.Errorf("reading streamed log: %w", err)
		}
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("closing streamed log response: %w", err)
		}

		moreData := resp.Header.Get("X-More-Data")
		newStart := resp.Header.Get("X-Text-Size")

		if newStart != "" {
			start = newStart
		}

		if !strings.EqualFold(moreData, "true") {
			break
		}

		if streamCtx != nil {
			select {
			case <-streamCtx.Done():
				return nil
			case <-time.After(1 * time.Second):
			}
		} else {
			if err := c.sleep(1 * time.Second); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) StopBuild(jobPath string, number int) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/stop", number)

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("stopping build: %w", err)
	}

	return nil
}

func (c *Client) DeleteBuild(jobPath string, number int) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/doDelete", number)

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("deleting build: %w", err)
	}

	return nil
}

func (c *Client) GetBuildArtifacts(jobPath string, number int) ([]Artifact, error) {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d", number)

	tree := "artifacts[displayPath,fileName,relativePath]"
	query := TreeParam(tree)

	data, err := c.Get(path, query)
	if err != nil {
		return nil, fmt.Errorf("getting artifacts: %w", err)
	}

	var build Build
	if err := json.Unmarshal(data, &build); err != nil {
		return nil, fmt.Errorf("parsing artifacts: %w", err)
	}

	return build.Artifacts, nil
}

func (c *Client) DownloadArtifact(jobPath string, number int, relativePath string) ([]byte, error) {
	var buf bytes.Buffer
	if err := c.DownloadArtifactTo(c.ctx, jobPath, number, relativePath, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) DownloadArtifactTo(ctx context.Context, jobPath string, number int, relativePath string, w io.Writer) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/artifact/%s", number, relativePath)

	resp, err := c.doRequest(requestOptions{ctx: ctx, method: "GET", path: path})
	if err != nil {
		return fmt.Errorf("downloading artifact: %w", err)
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return fmt.Errorf("downloading artifact: %w", err)
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("writing artifact: %w", err)
	}

	return nil
}

func (c *Client) GetBuildTestReport(jobPath string, number int) (*TestReport, error) {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/testReport", number)

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting test report: %w", err)
	}

	var report TestReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing test report: %w", err)
	}

	return &report, nil
}

func (c *Client) GetBuildEnvVars(jobPath string, number int) (map[string]string, error) {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/injectedEnvVars", number)

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting env vars: %w", err)
	}

	var wrapper EnvVarsWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parsing env vars: %w", err)
	}

	return wrapper.EnvMap, nil
}

func (c *Client) GetBuildStages(jobPath string, number int) (*PipelineRun, error) {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/wfapi/describe", number)

	data, err := c.GetRaw(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting pipeline stages: %w", err)
	}

	var run PipelineRun
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parsing pipeline stages: %w", err)
	}

	return &run, nil
}

func (c *Client) ReplayBuild(jobPath string, number int) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/replay/run", number)

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("replaying build: %w", err)
	}

	return nil
}

func FormatDuration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	if d < time.Second {
		return fmt.Sprintf("%dms", ms)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func FormatTimestamp(ms int64) string {
	if ms == 0 {
		return "N/A"
	}
	t := time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond))
	return t.Local().Format("2006-01-02 15:04:05")
}

func ColorToStatus(color string) string {
	switch {
	case strings.HasSuffix(color, "_anime"):
		return "RUNNING"
	case strings.HasPrefix(color, "blue"):
		return "SUCCESS"
	case strings.HasPrefix(color, "red"):
		return "FAILURE"
	case strings.HasPrefix(color, "yellow"):
		return "UNSTABLE"
	case strings.HasPrefix(color, "grey"), strings.HasPrefix(color, "disabled"):
		return "DISABLED"
	case strings.HasPrefix(color, "aborted"):
		return "ABORTED"
	case strings.HasPrefix(color, "notbuilt"):
		return "NOT_BUILT"
	default:
		return strings.ToUpper(color)
	}
}

func ParseBuildNumber(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid build number: %s", s)
	}
	return n, nil
}
