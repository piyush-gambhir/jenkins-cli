package client

import (
	"net/http"
	"testing"
)

// Regression test: defaultParameterValue is an object (a ParameterValue with
// name/value), null, or absent in the Jenkins API — never a plain string.
// Typing it as string made `pipeline input-list` fail for any input step
// that declares parameters.
func TestListPipelineInputs_ObjectDefaultParameterValue(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/my-pipeline/7/wfapi/pendingInputActions" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"id": "deploy-approval",
				"message": "Deploy to production?",
				"proceedText": "Proceed",
				"inputs": [
					{
						"name": "ENV",
						"type": "StringParameterDefinition",
						"description": "Target environment",
						"defaultParameterValue": {
							"_class": "hudson.model.StringParameterValue",
							"name": "ENV",
							"value": "staging"
						}
					},
					{
						"name": "FORCE",
						"type": "BooleanParameterDefinition",
						"defaultParameterValue": {
							"_class": "hudson.model.BooleanParameterValue",
							"name": "FORCE",
							"value": false
						}
					},
					{
						"name": "NO_DEFAULT",
						"type": "PasswordParameterDefinition",
						"defaultParameterValue": null
					}
				]
			}
		]`))
	})
	defer ts.Close()

	inputs, err := c.ListPipelineInputs("my-pipeline", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(inputs))
	}
	if inputs[0].ID != "deploy-approval" {
		t.Errorf("expected id deploy-approval, got %q", inputs[0].ID)
	}
	if len(inputs[0].Inputs) != 3 {
		t.Fatalf("expected 3 input params, got %d", len(inputs[0].Inputs))
	}
	dv, ok := inputs[0].Inputs[0].DefaultValue.(map[string]interface{})
	if !ok {
		t.Fatalf("expected object default value, got %T", inputs[0].Inputs[0].DefaultValue)
	}
	if dv["value"] != "staging" {
		t.Errorf("expected default value staging, got %v", dv["value"])
	}
	if inputs[0].Inputs[2].DefaultValue != nil {
		t.Errorf("expected nil default value, got %v", inputs[0].Inputs[2].DefaultValue)
	}
}

// Inputs may also be empty or absent entirely (plain `input` step without
// parameters).
func TestListPipelineInputs_NoParams(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id": "ok", "message": "Continue?", "proceedText": "Proceed"}]`))
	})
	defer ts.Close()

	inputs, err := c.ListPipelineInputs("my-pipeline", 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 1 || inputs[0].ID != "ok" {
		t.Fatalf("unexpected inputs: %+v", inputs)
	}
}
