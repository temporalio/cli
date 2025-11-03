// Code generated. DO NOT EDIT.

package temporalcloudcli

import (
	"github.com/mattn/go-isatty"

	"github.com/spf13/cobra"

	"github.com/temporalio/cli/temporalcli"

	"os"
)

var hasHighlighting = isatty.IsTerminal(os.Stdout.Fd())

type TemporalCloudCommand struct {
	Command cobra.Command
	Output  temporalcli.StringEnum
}

func NewTemporalCloudCommand(cctx *temporalcli.CommandContext) *TemporalCloudCommand {
	var s TemporalCloudCommand
	s.Command.Use = "temporal-cloud"
	s.Command.Short = "Temporal Cloud command-line interface"
	if hasHighlighting {
		s.Command.Long = "The Temporal Cloud CLI provides management and operations for Temporal Cloud.\n\nExample:\n\n\x1b[1mtemporal-cloud hello\x1b[0m"
	} else {
		s.Command.Long = "The Temporal Cloud CLI provides management and operations for Temporal Cloud.\n\nExample:\n\n```\ntemporal-cloud hello\n```"
	}
	s.Command.AddCommand(&NewTemporalCloudHelloCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalCloudUserCommand(cctx, &s).Command)
	s.Output = temporalcli.NewStringEnum([]string{"text", "json"}, "text")
	s.Command.PersistentFlags().VarP(&s.Output, "output", "o", "Output format. Accepted values: text, json.")
	return &s
}

type TemporalCloudHelloCommand struct {
	Parent  *TemporalCloudCommand
	Command cobra.Command
}

func NewTemporalCloudHelloCommand(cctx *temporalcli.CommandContext, parent *TemporalCloudCommand) *TemporalCloudHelloCommand {
	var s TemporalCloudHelloCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "hello [flags]"
	s.Command.Short = "Print hello world message"
	if hasHighlighting {
		s.Command.Long = "Print a hello world message to test the Temporal Cloud CLI extension.\n\nExample:\n\n\x1b[1mtemporal-cloud hello\x1b[0m"
	} else {
		s.Command.Long = "Print a hello world message to test the Temporal Cloud CLI extension.\n\nExample:\n\n```\ntemporal-cloud hello\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalCloudUserCommand struct {
	Parent  *TemporalCloudCommand
	Command cobra.Command
}

func NewTemporalCloudUserCommand(cctx *temporalcli.CommandContext, parent *TemporalCloudCommand) *TemporalCloudUserCommand {
	var s TemporalCloudUserCommand
	s.Parent = parent
	s.Command.Use = "user"
	s.Command.Short = "Manage users"
	s.Command.Long = "Commands for managing users."
	s.Command.AddCommand(&NewTemporalCloudUserGetCommand(cctx, &s).Command)
	return &s
}

type TemporalCloudUserGetCommand struct {
	Parent    *TemporalCloudUserCommand
	Command   cobra.Command
	ShowEmail bool
}

func NewTemporalCloudUserGetCommand(cctx *temporalcli.CommandContext, parent *TemporalCloudUserCommand) *TemporalCloudUserGetCommand {
	var s TemporalCloudUserGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get user details"
	s.Command.Long = "Shows details for a specific user."
	s.Command.Args = cobra.ExactArgs(1)
	s.Command.Flags().BoolVar(&s.ShowEmail, "show-email", false, "Show the user's email address.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}
