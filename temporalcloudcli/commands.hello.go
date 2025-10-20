package temporalcloudcli

import (
	"encoding/json"
	"fmt"

	"github.com/temporalio/cli/temporalcli"
)

func (c *TemporalCloudHelloCommand) run(cctx *temporalcli.CommandContext, args []string) error {
	message := map[string]string{
		"message": "Hello from Temporal Cloud CLI!",
		"version": "0.1.0",
	}

	// Check output format
	if c.Parent.Output.Value == "json" {
		encoder := json.NewEncoder(cctx.Options.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(message)
	}

	// Text output
	fmt.Fprintln(cctx.Options.Stdout, message["message"])
	fmt.Fprintf(cctx.Options.Stdout, "Version: %s\n", message["version"])
	return nil
}
