package temporalcloudcli

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/temporalio/cli/temporalcli"
)

//go:embed commands.yml
var CommandsYAML []byte

// Execute runs the cloud CLI
func Execute(ctx context.Context, options temporalcli.CommandOptions) {
	// Set defaults
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}
	if options.Fail == nil {
		options.Fail = func(err error) {
			fmt.Fprintln(options.Stderr, err)
			os.Exit(1)
		}
	}

	// Create command context - reuse temporalcli's CommandContext
	cctx := &temporalcli.CommandContext{
		Context: ctx,
		Options: options,
	}

	// Create and execute root command
	cmd := NewTemporalCloudCommand(cctx)
	cmd.Command.SetArgs(options.Args)
	cmd.Command.SetOut(options.Stdout)
	cmd.Command.SetErr(options.Stderr)
	cmd.Command.SetIn(options.Stdin)

	if err := cmd.Command.ExecuteContext(ctx); err != nil {
		options.Fail(err)
	}
}
