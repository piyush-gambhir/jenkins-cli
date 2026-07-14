package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

// checkReadOnly simulates the read-only enforcement logic from PersistentPreRunE.
func checkReadOnly(readOnly bool, cmd *cobra.Command) error {
	if readOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode.\nTo disable, use --read-only=false or remove read_only from your config profile.", cmd.CommandPath())
	}
	return nil
}

func TestCheckReadOnly_WriteCmdBlocked(t *testing.T) {
	cmd := &cobra.Command{
		Use:         "delete",
		Annotations: map[string]string{"mutates": "true"},
	}

	err := checkReadOnly(true, cmd)
	if err == nil {
		t.Fatal("expected error when write command is run in read-only mode, got nil")
	}
}

func TestCheckReadOnly_WriteCmdAllowed(t *testing.T) {
	cmd := &cobra.Command{
		Use:         "delete",
		Annotations: map[string]string{"mutates": "true"},
	}

	err := checkReadOnly(false, cmd)
	if err != nil {
		t.Fatalf("expected no error when read-only is false, got: %v", err)
	}
}

func TestCheckReadOnly_ReadCmdAllowed(t *testing.T) {
	cmd := &cobra.Command{
		Use: "list",
	}

	err := checkReadOnly(true, cmd)
	if err != nil {
		t.Fatalf("expected no error for read command in read-only mode, got: %v", err)
	}
}
