package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	jpath "github.com/piyush-gambhir/jenkins-cli/internal/path"
)

// Job represents a Jenkins job.
type Job struct {
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	Color       string     `json:"color"`
	FullName    string     `json:"fullName"`
	DisplayName string     `json:"displayName"`
	Description string     `json:"description"`
	Buildable   bool       `json:"buildable"`
	InQueue     bool       `json:"inQueue"`
	LastBuild   *BuildRef  `json:"lastBuild"`
	Jobs        []Job      `json:"jobs"`
	Class       string     `json:"_class"`
	Actions     []Action   `json:"actions"`
	Property    []Property `json:"property"`
	HealthReport []HealthReport `json:"healthReport"`
}

// BuildRef is a minimal build reference.
type BuildRef struct {
	Number    int    `json:"number"`
	URL       string `json:"url"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
}

// Action represents a Jenkins action.
type Action struct {
	Class              string           `json:"_class"`
	ParameterDefinitions []ParamDefinition `json:"parameterDefinitions"`
}

// ParamDefinition defines a build parameter.
type ParamDefinition struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"defaultParameterValue"`
}

// Property represents a job property.
type Property struct {
	Class              string           `json:"_class"`
	ParameterDefinitions []ParamDefinition `json:"parameterDefinitions"`
}

// HealthReport represents a job health report.
type HealthReport struct {
	Description string `json:"description"`
	Score       int    `json:"score"`
}

// JobListResponse wraps a list of jobs.
type JobListResponse struct {
	Jobs []Job `json:"jobs"`
}

// ListJobs lists jobs in a folder (or root if folder is empty).
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

// ListJobsRecursive lists all jobs recursively.
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

// GetJob gets detailed info about a job.
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

// GetJobConfig gets the config.xml of a job.
func (c *Client) GetJobConfig(jobPath string) (string, error) {
	path := jpath.ToJenkinsPath(jobPath) + "/config.xml"

	data, err := c.GetRaw(path, nil)
	if err != nil {
		return "", fmt.Errorf("getting job config: %w", err)
	}

	return string(data), nil
}

// CreateJob creates a new job in the specified folder.
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

// UpdateJobConfig updates a job's config.xml.
func (c *Client) UpdateJobConfig(jobPath, configXML string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/config.xml"

	_, _, err := c.PostXML(path, nil, configXML)
	if err != nil {
		return fmt.Errorf("updating job config: %w", err)
	}

	return nil
}

// CopyJob copies a job.
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

// RenameJob renames a job.
func (c *Client) RenameJob(jobPath, newName string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/doRename"
	query := url.Values{"newName": {newName}}

	_, err := c.Post(path, query)
	if err != nil {
		return fmt.Errorf("renaming job: %w", err)
	}

	return nil
}

// DeleteJob deletes a job.
func (c *Client) DeleteJob(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/doDelete"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}

	return nil
}

// EnableJob enables a disabled job.
func (c *Client) EnableJob(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/enable"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("enabling job: %w", err)
	}

	return nil
}

// DisableJob disables a job.
func (c *Client) DisableJob(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/disable"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("disabling job: %w", err)
	}

	return nil
}

// WipeWorkspace wipes the workspace of a job.
func (c *Client) WipeWorkspace(jobPath string) error {
	path := jpath.ToJenkinsPath(jobPath) + "/doWipeOutWorkspace"

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("wiping workspace: %w", err)
	}

	return nil
}

// QueueLocation is the response from triggering a build.
type QueueLocation struct {
	QueueURL string
	QueueID  int
}

// TriggerBuild triggers a build for a job.
func (c *Client) TriggerBuild(jobPath string, params map[string]string) (*QueueLocation, error) {
	jp := jpath.ToJenkinsPath(jobPath)

	var path string
	var query url.Values

	if len(params) > 0 {
		path = jp + "/buildWithParameters"
		query = url.Values{}
		for k, v := range params {
			query.Set(k, v)
		}
	} else {
		path = jp + "/build"
	}

	resp, err := c.PostRaw(path, query)
	if err != nil {
		return nil, fmt.Errorf("triggering build: %w", err)
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	if loc == "" {
		return &QueueLocation{}, nil
	}

	return &QueueLocation{QueueURL: loc}, nil
}

// TriggerBuildAndWait triggers a build and waits for it to complete.
func (c *Client) TriggerBuildAndWait(jobPath string, params map[string]string, timeout time.Duration) (*Build, error) {
	ql, err := c.TriggerBuild(jobPath, params)
	if err != nil {
		return nil, err
	}

	if ql.QueueURL == "" {
		return nil, fmt.Errorf("no queue location returned")
	}

	// Wait for the build to start
	build, err := c.WaitForQueuedBuild(ql.QueueURL, timeout)
	if err != nil {
		return nil, err
	}

	// Wait for the build to complete
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		b, err := c.GetBuild(jobPath, build.Number)
		if err != nil {
			return nil, err
		}
		if !b.Building {
			return b, nil
		}
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("build timed out after %s", timeout)
}

// WaitForQueuedBuild waits for a queued build to get an executor and start.
func (c *Client) WaitForQueuedBuild(queueURL string, timeout time.Duration) (*BuildRef, error) {
	// Extract the queue item path from the URL
	// queueURL is like http://jenkins/queue/item/123/
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		data, err := c.GetRaw(extractQueuePath(queueURL)+"/api/json", nil)
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

		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("timed out waiting for build to start")
}

// extractQueuePath extracts the path portion from a queue URL.
func extractQueuePath(queueURL string) string {
	u, err := url.Parse(queueURL)
	if err != nil {
		return queueURL
	}
	return u.Path
}
