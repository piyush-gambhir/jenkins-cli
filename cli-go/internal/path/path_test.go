package path

import (
	"testing"
)

func TestToJenkinsPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple job",
			input:    "my-job",
			expected: "/job/my-job",
		},
		{
			name:     "folder and job",
			input:    "folder/job",
			expected: "/job/folder/job/job",
		},
		{
			name:     "nested folders and job",
			input:    "folder/subfolder/job",
			expected: "/job/folder/job/subfolder/job/job",
		},
		{
			name:     "spaces in segments are URL-encoded",
			input:    "my folder/my job",
			expected: "/job/my%20folder/job/my%20job",
		},
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "single segment",
			input:    "single",
			expected: "/job/single",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToJenkinsPath(tt.input)
			if got != tt.expected {
				t.Errorf("ToJenkinsPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFromJenkinsPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple job path",
			input:    "/job/my-job",
			expected: "my-job",
		},
		{
			name:     "folder and job path",
			input:    "/job/folder/job/job",
			expected: "folder/job",
		},
		{
			name:     "deeply nested path",
			input:    "/job/a/job/b/job/c",
			expected: "a/b/c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromJenkinsPath(tt.input)
			if got != tt.expected {
				t.Errorf("FromJenkinsPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"my-job",
		"folder/job",
		"a/b/c",
		"single",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			jenkinsPath := ToJenkinsPath(input)
			got := FromJenkinsPath(jenkinsPath)
			if got != input {
				t.Errorf("round-trip failed: ToJenkinsPath(%q)=%q, FromJenkinsPath(...)=%q, want %q",
					input, jenkinsPath, got, input)
			}
		})
	}
}

func TestBuildJobPath(t *testing.T) {
	tests := []struct {
		name     string
		userPath string
		suffix   string
		expected string
	}{
		{
			name:     "job with build suffix",
			userPath: "folder/job",
			suffix:   "/build",
			expected: "/job/folder/job/job/build",
		},
		{
			name:     "job with no suffix",
			userPath: "folder/job",
			suffix:   "",
			expected: "/job/folder/job/job",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildJobPath(tt.userPath, tt.suffix)
			if got != tt.expected {
				t.Errorf("BuildJobPath(%q, %q) = %q, want %q", tt.userPath, tt.suffix, got, tt.expected)
			}
		})
	}
}
