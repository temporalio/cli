package temporalcloudcli

import (
	"fmt"

	"github.com/temporalio/cli/temporalcli"
)

func (c *TemporalCloudUserCommand) run(cctx *temporalcli.CommandContext, args []string) error {
	// User command is a parent command with subcommands
	// Should not be called directly
	return fmt.Errorf("please specify a subcommand")
}

func (c *TemporalCloudUserGetCommand) run(cctx *temporalcli.CommandContext, args []string) error {
	username := args[0]

	// Mock user data
	user := map[string]interface{}{
		"username": username,
		"role":     "developer",
	}

	if c.ShowEmail {
		user["email"] = fmt.Sprintf("%s@example.com", username)
	}

	fmt.Fprintf(cctx.Options.Stdout, "Details for user: %s\n", user["username"])
	fmt.Fprintf(cctx.Options.Stdout, "  Role: %s\n", user["role"])
	if c.ShowEmail {
		fmt.Fprintf(cctx.Options.Stdout, "  Email: %s\n", user["email"])
	}

	return nil
}
