package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	jpath "github.com/piyush-gambhir/jenkins-cli/cli-go/internal/path"
)

type PipelineInput struct {
	ID          string               `json:"id"`
	Message     string               `json:"message"`
	ProceedText string               `json:"proceedText"`
	AbortText   string               `json:"abortText"`
	Inputs      []PipelineInputParam `json:"inputs"`
	ProceedURL  string               `json:"proceedUrl"`
	AbortURL    string               `json:"abortUrl"`
	RedirectURL string               `json:"redirectApprovalUrl"`
}

type PipelineInputParam struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"defaultParameterValue"`
	Description  string      `json:"description"`
}

type PipelineInputResponse struct {
	Inputs []PipelineInput `json:"inputActions,omitempty"`
}

func (c *Client) ValidateJenkinsfile(content string) (string, error) {
	formData := url.Values{
		"jenkinsfile": {content},
	}

	data, err := c.PostForm("/pipeline-model-converter/validate", nil, formData)
	if err != nil {
		return "", fmt.Errorf("validating Jenkinsfile: %w", err)
	}

	return string(data), nil
}

func (c *Client) ListPipelineInputs(jobPath string, buildNumber int) ([]PipelineInput, error) {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/wfapi/pendingInputActions", buildNumber)

	data, err := c.GetRaw(path, nil)
	if err != nil {
		return nil, fmt.Errorf("listing pipeline inputs: %w", err)
	}

	var inputs []PipelineInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, fmt.Errorf("parsing pipeline inputs: %w", err)
	}

	return inputs, nil
}

func (c *Client) SubmitPipelineInput(jobPath string, buildNumber int, inputID string, params map[string]string) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/input/%s/proceed", buildNumber, url.PathEscape(inputID))

	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}
	sort.Strings(names)
	submitted := make([]map[string]string, 0, len(names))
	for _, name := range names {
		submitted = append(submitted, map[string]string{"name": name, "value": params[name]})
	}
	jsonBytes, err := json.Marshal(map[string]interface{}{"parameter": submitted})
	if err != nil {
		return fmt.Errorf("marshaling input params: %w", err)
	}
	formData := url.Values{"json": {string(jsonBytes)}}

	_, err = c.PostForm(path, nil, formData)
	if err != nil {
		return fmt.Errorf("submitting pipeline input: %w", err)
	}

	return nil
}

func (c *Client) AbortPipelineInput(jobPath string, buildNumber int, inputID string) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/input/%s/abort", buildNumber, url.PathEscape(inputID))

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("aborting pipeline input: %w", err)
	}

	return nil
}
