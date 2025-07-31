package temporalcli

import (
	"testing"
)

func TestTemporalTaskQueueUpdateConfigCommand_Flags(t *testing.T) {
	// This test verifies that the command flags are properly defined
	// and the command can be created without errors

	cctx := &CommandContext{}
	parent := &TemporalTaskQueueCommand{}

	cmd := NewTemporalTaskQueueUpdateConfigCommand(cctx, parent)

	// Verify the command has the expected flags
	flags := cmd.Command.Flags()

	// Check that required flags exist
	if flags.Lookup("task-queue") == nil {
		t.Error("task-queue flag not found")
	}

	if flags.Lookup("task-queue-type") == nil {
		t.Error("task-queue-type flag not found")
	}

	if flags.Lookup("queue-rate-limit") == nil {
		t.Error("queue-rate-limit flag not found")
	}

	if flags.Lookup("fairness-key-rate-limit-default") == nil {
		t.Error("fairness-key-rate-limit-default flag not found")
	}

	if flags.Lookup("identity") == nil {
		t.Error("identity flag not found")
	}

	// Verify command metadata
	if cmd.Command.Use != "update-config [flags]" {
		t.Errorf("expected command use to be 'update-config [flags]', got '%s'", cmd.Command.Use)
	}

	if cmd.Command.Short != "Update Task Queue configuration" {
		t.Errorf("expected command short description to be 'Update Task Queue configuration', got '%s'", cmd.Command.Short)
	}
}
