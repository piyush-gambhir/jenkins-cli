package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/piyush-gambhir/jenkins-cli/cli-go/internal/config"
	"github.com/spf13/cobra"
)

func TestCommandsResolve(t *testing.T) {
	oldOutput, oldExported, oldFormat := outputFormat, OutputFormat, outFormat
	t.Cleanup(func() { outputFormat, OutputFormat, outFormat = oldOutput, oldExported, oldFormat })
	// Invalid config and a release version must not cause offline resolution to read config or check updates.
	checks := stubNotifier(t, "0.2.9")
	if err := os.MkdirAll(config.ConfigDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config.ConfigDir(), "config.yaml"), []byte("invalid: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, command, effect string
		args                  []string
		flags                 map[string]string
		positional            []string
		wantErr               bool
	}{
		{name: "read", command: "jenkins job list", effect: "read", args: []string{"job", "list"}, flags: map[string]string{}, positional: []string{}},
		{name: "write with leading flag", command: "jenkins job build", effect: "remote_write", args: []string{"-o", "json", "job", "build", "my-pipeline"}, flags: map[string]string{"output": "json"}, positional: []string{"my-pipeline"}},
		{name: "alias", command: "jenkins node list", effect: "read", args: []string{"agents", "list"}, flags: map[string]string{}, positional: []string{}},
		{name: "unknown", args: []string{"nonsense"}, wantErr: true},
		{name: "root", wantErr: true},
		{name: "group", args: []string{"job"}, wantErr: true},
		{name: "invalid flag", args: []string{"job", "list", "--nonsense"}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			root := newRootCmd(false)
			root.SetOut(&out)
			root.SetErr(&bytes.Buffer{})
			root.SetArgs(append([]string{"commands", "resolve", "--"}, tc.args...))
			err := root.Execute()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var got resolution
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			want := resolution{Command: tc.command, Effect: tc.effect, Flags: tc.flags, Args: tc.positional}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %+v, want %+v", got, want)
			}
		})
	}
	if checks.Load() != 0 {
		t.Fatalf("offline resolver checked for updates %d times", checks.Load())
	}
}

func TestResolutionAnnotations(t *testing.T) {
	for _, tc := range []struct {
		name        string
		ann         map[string]string
		set         bool
		effect      string
		conditional bool
	}{
		{"remote", map[string]string{"mutates": "true", "writes-local": "true"}, false, "remote_write", false},
		{"conditional remote", map[string]string{"mutates": "conditional"}, false, "remote_write", true},
		{"flag remote unset", map[string]string{"mutates-when-flag": "apply"}, false, "read", false},
		{"flag remote set false", map[string]string{"mutates-when-flag": "apply"}, true, "remote_write", false},
		{"local", map[string]string{"writes-local": "true"}, false, "local_write", false},
		{"conditional local", map[string]string{"writes-local": "conditional"}, false, "local_write", true},
		{"self update", map[string]string{"self-update": "true"}, false, "local_write", false},
		{"flag local unset", map[string]string{"writes-local-when-flag": "apply"}, false, "read", false},
		{"flag local set", map[string]string{"writes-local-when-flag": "apply"}, true, "local_write", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &cobra.Command{Use: "probe", Annotations: tc.ann}
			target.Annotations["interactive"] = "true"
			target.Annotations["secret-output"] = "true"
			target.Flags().Bool("apply", false, "")
			args := []string{}
			if tc.set {
				args = []string{"--apply=false"}
			}
			if err := target.ParseFlags(args); err != nil {
				t.Fatal(err)
			}
			got := describeResolution(target)
			if got.Effect != tc.effect || got.Conditional != tc.conditional || !got.Interactive || !got.Secret {
				t.Fatalf("unexpected resolution: %+v", got)
			}
		})
	}
}
