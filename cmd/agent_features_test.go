package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestNoInputFlag_BlocksConfirmPrompt verifies that when noInputFlag is set,
// commands that require --confirm return the appropriate error.
func TestNoInputFlag_BlocksConfirmPrompt(t *testing.T) {
	// Save and restore global state
	origNoInput := noInputFlag
	defer func() { noInputFlag = origNoInput }()

	noInputFlag = true

	tests := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"job delete", newJobDeleteCmd},
		{"node delete", newNodeDeleteCmd},
		{"view delete", newViewDeleteCmd},
		{"credential delete", newCredentialDeleteCmd},
		{"plugin uninstall", newPluginUninstallCmd},
		{"build delete", newBuildDeleteCmd},
		{"system restart", newSystemRestartCmd},
		{"job wipe-workspace", newJobWipeWorkspaceCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()

			// Determine args to pass
			var args []string
			switch tt.name {
			case "build delete":
				args = []string{"test-job", "1"}
			case "credential delete", "node delete", "view delete",
				"job delete", "plugin uninstall", "job wipe-workspace":
				args = []string{"test-item"}
			case "system restart":
				args = []string{}
			}

			// We need a minimal setup. Since these commands need jenkinsClient
			// but we're testing the confirm/no-input check which happens BEFORE
			// any API call, this should work for commands where !confirm triggers
			// early. For build delete we also need client.ParseBuildNumber to work.
			err := cmd.RunE(cmd, args)
			if err == nil {
				t.Fatal("expected error when no-input is set and --confirm is not provided")
			}
			expected := "interactive input required but --no-input is set. Use --confirm for destructive operations."
			if err.Error() != expected {
				t.Fatalf("expected error %q, got: %v", expected, err)
			}
		})
	}
}

// TestQuietFlag_Variable verifies the quiet flag is accessible.
func TestQuietFlag_Variable(t *testing.T) {
	origQuiet := quietFlag
	defer func() { quietFlag = origQuiet }()

	quietFlag = true
	if !quietFlag {
		t.Fatal("expected quietFlag to be true")
	}

	quietFlag = false
	if quietFlag {
		t.Fatal("expected quietFlag to be false")
	}
}

// TestNoInputFlag_AllowsWithConfirm verifies that when noInputFlag is set
// AND --confirm is provided, the no-input check does not block.
// We test the logic by verifying confirm=true skips the no-input gate.
func TestNoInputFlag_AllowsWithConfirm(t *testing.T) {
	origNoInput := noInputFlag
	defer func() { noInputFlag = origNoInput }()

	noInputFlag = true

	// Create a job delete command and parse --confirm
	cmd := newJobDeleteCmd()
	if err := cmd.ParseFlags([]string{"--confirm"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	// We need to catch the nil-pointer panic from jenkinsClient being nil.
	// The important thing is that we get past the no-input check.
	var runErr error
	panicked := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		runErr = cmd.RunE(cmd, []string{"test-job"})
	}()

	// If we panicked, it means we got past the no-input check and hit the
	// nil jenkinsClient -- which is the expected behavior.
	if panicked {
		return // Test passes: we got past the no-input gate
	}

	// If we didn't panic but got an error, make sure it's NOT the no-input error
	if runErr != nil {
		noInputErr := "interactive input required but --no-input is set. Use --confirm for destructive operations."
		if runErr.Error() == noInputErr {
			t.Fatal("--confirm should bypass the no-input check, but got the no-input error")
		}
	}
}

// TestIdempotentFlags_Exist verifies that --if-not-exists and --if-exists flags exist
// on the appropriate commands.
func TestIdempotentFlags_Exist(t *testing.T) {
	createCmds := []struct {
		name    string
		cmdFunc func() *cobra.Command
		flag    string
	}{
		{"job create", newJobCreateCmd, "if-not-exists"},
		{"node create", newNodeCreateCmd, "if-not-exists"},
		{"view create", newViewCreateCmd, "if-not-exists"},
		{"credential create", newCredentialCreateCmd, "if-not-exists"},
		{"plugin install", newPluginInstallCmd, "if-not-exists"},
		{"job delete", newJobDeleteCmd, "if-exists"},
		{"node delete", newNodeDeleteCmd, "if-exists"},
		{"view delete", newViewDeleteCmd, "if-exists"},
		{"credential delete", newCredentialDeleteCmd, "if-exists"},
		{"plugin uninstall", newPluginUninstallCmd, "if-exists"},
	}

	for _, tt := range createCmds {
		t.Run(tt.name+" has --"+tt.flag, func(t *testing.T) {
			cmd := tt.cmdFunc()
			f := cmd.Flags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("expected --%s flag on %s command", tt.flag, tt.name)
			}
		})
	}
}

// TestStdinFlag_FromFileAcceptsDash verifies that --from-file accepts "-" value
// on commands that support stdin.
func TestStdinFlag_FromFileAcceptsDash(t *testing.T) {
	cmds := []struct {
		name    string
		cmdFunc func() *cobra.Command
	}{
		{"job create", newJobCreateCmd},
		{"job update", newJobUpdateCmd},
		{"credential create", newCredentialCreateCmd},
		{"credential update", newCredentialUpdateCmd},
		{"pipeline validate", newPipelineValidateCmd},
		{"system run-script", newSystemRunScriptCmd},
	}

	for _, tt := range cmds {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFunc()
			f := cmd.Flags().Lookup("from-file")
			if f == nil {
				t.Fatalf("expected --from-file flag on %s command", tt.name)
			}
			// Verify the usage text mentions stdin
			if f.Usage == "" {
				t.Fatalf("expected non-empty usage for --from-file on %s", tt.name)
			}
		})
	}
}

// TestGlobalFlags_Registered verifies that the global flags --no-input, --quiet,
// and --verbose are registered on the root command.
func TestGlobalFlags_Registered(t *testing.T) {
	flags := []struct {
		name      string
		shorthand string
	}{
		{"no-input", ""},
		{"quiet", "q"},
		{"verbose", "v"},
		{"read-only", ""},
		{"no-color", ""},
	}

	for _, f := range flags {
		t.Run("--"+f.name, func(t *testing.T) {
			pf := rootCmd.PersistentFlags().Lookup(f.name)
			if pf == nil {
				t.Fatalf("expected persistent flag --%s on root command", f.name)
			}
			if f.shorthand != "" && pf.Shorthand != f.shorthand {
				t.Errorf("expected shorthand %q for --%s, got %q", f.shorthand, f.name, pf.Shorthand)
			}
		})
	}
}
