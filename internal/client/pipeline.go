package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	jpath "github.com/piyush-gambhir/jenkins-cli/internal/path"
)

// PipelineInput represents a pending pipeline input.
type PipelineInput struct {
	ID          string             `json:"id"`
	Message     string             `json:"message"`
	ProceedText string             `json:"proceedText"`
	AbortText   string             `json:"abortText"`
	Inputs      []PipelineInputParam `json:"inputs"`
	ProceedURL  string             `json:"proceedUrl"`
	AbortURL    string             `json:"abortUrl"`
	RedirectURL string             `json:"redirectApprovalUrl"`
}

// PipelineInputParam represents an input parameter.
type PipelineInputParam struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultValue string `json:"defaultParameterValue"`
	Description  string `json:"description"`
}

// PipelineInputResponse wraps a list of pending inputs.
type PipelineInputResponse struct {
	Inputs []PipelineInput `json:"inputActions,omitempty"`
}

// ValidateJenkinsfile validates a Jenkinsfile.
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

// ListPipelineInputs lists pending input actions for a build.
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

// SubmitPipelineInput submits (proceeds) an input action.
func (c *Client) SubmitPipelineInput(jobPath string, buildNumber int, inputID string, params map[string]string) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/input/%s/proceed", buildNumber, url.PathEscape(inputID))

	formData := url.Values{}
	if len(params) > 0 {
		jsonBytes, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshaling input params: %w", err)
		}
		formData.Set("json", string(jsonBytes))
	}

	_, err := c.PostForm(path, nil, formData)
	if err != nil {
		return fmt.Errorf("submitting pipeline input: %w", err)
	}

	return nil
}

// AbortPipelineInput aborts an input action.
func (c *Client) AbortPipelineInput(jobPath string, buildNumber int, inputID string) error {
	path := jpath.ToJenkinsPath(jobPath) + fmt.Sprintf("/%d/input/%s/abort", buildNumber, url.PathEscape(inputID))

	_, err := c.Post(path, nil)
	if err != nil {
		return fmt.Errorf("aborting pipeline input: %w", err)
	}

	return nil
}
