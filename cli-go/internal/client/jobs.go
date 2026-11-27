package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	jpath "github.com/piyush-gambhir/jenkins-cli/cli-go/internal/path"
)

type Job struct {
	Name         string         `json:"name"`
	URL          string         `json:"url"`
	Color        string         `json:"color"`
	FullName     string         `json:"fullName"`
	DisplayName  string         `json:"displayName"`
	Description  string         `json:"description"`
	Buildable    bool           `json:"buildable"`
	InQueue      bool           `json:"inQueue"`
	LastBuild    *BuildRef      `json:"lastBuild"`
	Jobs         []Job          `json:"jobs"`
	Class        string         `json:"_class"`
	Actions      []Action       `json:"actions"`
	Property     []Property     `json:"property"`
	HealthReport []HealthReport `json:"healthReport"`
}

type BuildRef struct {
	Number    int    `json:"number"`
	URL       string `json:"url"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
}

type Action struct {
	Class                string            `json:"_class"`
	ParameterDefinitions []ParamDefinition `json:"parameterDefinitions"`
}

type ParamDefinition struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"defaultParameterValue"`
}

type Property struct {
	Class                string            `json:"_class"`
	ParameterDefinitions []ParamDefinition `json:"parameterDefinitions"`
}

type HealthReport struct {
	Description string `json:"description"`
	Score       int    `json:"score"`
}

type JobListResponse struct {
	Jobs []Job `json:"jobs"`
}

func (c *Client) ListJobs(folder string) ([]Job, error) {
	path := ""
	if folder != "" {
		path = jpath.ToJenkinsPath(folder)
	}

	tree := "jobs[name,url,color,fullName,_class,lastBuild[number,result,timestamp]]"
	query := TreeParam(tree)

	data, err := c.Get(path, query)
	if err != nil {
		return nil, fmt.Errorf("listing jobs: %w", err)
	}

	var resp JobListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing jobs: %w", err)
	}

	return resp.Jobs, nil
}

func (c *Client) ListJobsRecursive(folder string) ([]Job, error) {
	jobs, err := c.ListJobs(folder)
	if err != nil {
		return nil, err
	}

	var result []Job
	for _, j := range jobs {
		if isFolder(j) {
			prefix := j.Name
			if folder != "" {
				prefix = folder + "/" + j.Name
			}
			children, err := c.ListJobsRecursive(prefix)
			if err != nil {
				return nil, err
			}
			result = append(result, children...)
		} else {
			result = append(result, j)
		}
	}

	return result, nil
}

func isFolder(j Job) bool {
	return j.Class == "com.cloudbees.hudson.plugins.folder.Folder" ||
		j.Class == "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" ||
		j.Class == "jenkins.branch.OrganizationFolder" ||
		len(j.Jobs) > 0
}

func (c *Client) GetJob(jobPath string) (*Job, error) {
	path := jpath.ToJenkinsPath(jobPath)

	data, err := c.Get(path, nil)
	if err != nil {
		return nil, fmt.Errorf("getting job: %w", err)
	}

	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, fmt.Errorf("parsing job: %w", err)
	}

	return &job, nil
}

func (c *Client) GetJobConfig(jobPath string) (string, error) {
	path := jpath.ToJenkinsPath(jobPath) + "/config.xml"

	data, err := c.GetRaw(path, nil)
	if err != nil {
		return "", fmt.Errorf("getting job config: %w", err)
	}

	return string(data), nil
}

func (c *Client) CreateJob(folder, name, configXML string) error {
	path := "/createItem"
	if folder != "" {
		path = jpath.ToJenkinsPath(folder) + "/createItem"
	}

	query := url.Values{"name": {name}}

	_, _, err := c.PostXML(path, query, configXML)
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}

	return nil
}

func (c *Client) UpdateJobConfig(jobPath, configXML string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/config.xml"

	_, _, err := c.PostXML(path, nil, configXML)
	if err != nil {
		return fmt.Errorf("updating job config: %w", err)
	}

	return nil
}

func (c *Client) CopyJob(srcName, destName, folder string) error {
	path := "/createItem"
	if folder != "" {
		path = jpath.ToJenkinsPath(folder) + "/createItem"
	}

	query := url.Values{
		"name": {destName},
		"mode": {"copy"},
		"from": {srcName},
	}

	_, err := c.Post(path, query)
	if err != nil {
		return fmt.Errorf("copying job: %w", err)
	}

	return nil
}

func (c *Client) RenameJob(jobPath, newName string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/doRename"
	query := url.Values{"newName": {newName}}

	_, err := c.Post(path, query)
	if err != nil {
		return fmt.Errorf("renaming job: %w", err)
	}

	return nil
}

func (c *Client) DeleteJob(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/doDelete"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}

	return nil
}

func (c *Client) EnableJob(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/enable"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("enabling job: %w", err)
	}

	return nil
}

func (c *Client) DisableJob(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/disable"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("disabling job: %w", err)
	}

	return nil
}

func (c *Client) WipeWorkspace(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/doWipeOutWorkspace"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("wiping workspace: %w", err)
	}

	return nil
}

type QueueLocation struct {
	QueueURL string
	QueueID  int
}

func (c *Client) TriggerBuild(jobPath string, params map[string]string) (*QueueLocation, error) {
	jp := jpath.ToJenkinsPath(jobPath)

	opts := requestOptions{method: "POST", path: jp + "/build", stopAt303: true}
	if len(params) > 0 {
		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}
		opts.path = jp + "/buildWithParameters"
		opts.body = strings.NewReader(form.Encode())
		opts.contentType = "application/x-www-form-urlencoded"
	}

	resp, err := c.doRequest(opts)
	if err != nil {
		return nil, fmt.Errorf("triggering build: %w", err)
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return nil, fmt.Errorf("triggering build: %w", err)
	}

	loc := resp.Header.Get("Location")
	if loc == "" {
		return &QueueLocation{}, nil
	}

	return &QueueLocation{QueueURL: loc}, nil
}

func (c *Client) TriggerBuildAndWait(jobPath string, params map[string]string, timeout time.Duration) (*Build, error) {
	ql, err := c.TriggerBuild(jobPath, params)
	if err != nil {
		return nil, err
	}

	if ql.QueueURL == "" {
		return nil, fmt.Errorf("no queue location returned")
	}

	build, err := c.WaitForQueuedBuild(ql.QueueURL, timeout)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		b, err := c.GetBuild(jobPath, build.Number)
		if err != nil {
			return nil, err
		}
		if !b.Building {
			return b, nil
		}
		if err := c.sleep(2 * time.Second); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("build timed out after %s", timeout)
}

func (c *Client) WaitForQueuedBuild(queueURL string, timeout time.Duration) (*BuildRef, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		data, err := c.GetRaw(c.queueItemPath(queueURL)+"/api/json", nil)
		if err != nil {
			return nil, fmt.Errorf("checking queue: %w", err)
		}

		var item QueueItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("parsing queue item: %w", err)
		}

		if item.Executable.Number > 0 {
			return &BuildRef{
				Number: item.Executable.Number,
				URL:    item.Executable.URL,
			}, nil
		}

		if item.Cancelled {
			return nil, fmt.Errorf("build was cancelled in queue")
		}

		if err := c.sleep(2 * time.Second); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("timed out waiting for build to start")
}

func (c *Client) queueItemPath(queueURL string) string {
	p := queueURL
	if u, err := url.Parse(queueURL); err == nil {
		p = u.Path
	}
	if base, err := url.Parse(c.baseURL); err == nil && base.Path != "" && strings.HasPrefix(p, base.Path+"/") {
		p = strings.TrimPrefix(p, base.Path)
	}
	return strings.TrimRight(p, "/")
}
