// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"strings"

	"github.com/temporalio/tctl-kit/pkg/flags"
	"github.com/urfave/cli/v2"
)

func newWorkflowCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "start",
			Usage: "Start a new Workflow Execution",
			Flags: append(flagsForStartWorkflow, flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, false)
			},
		},
		{
			Name:  "execute",
			Usage: "Start a new Workflow Execution and print progress",
			Flags: append(flagsForStartWorkflow, flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, true)
			},
		},
		{
			Name:  "describe",
			Usage: "Show information about a Workflow Execution",
			Flags: append(flagsForExecution, []cli.Flag{
				&cli.BoolFlag{
					Name:  FlagResetPointsOnly,
					Usage: "Only show auto-reset points",
				},
				&cli.BoolFlag{
					Name:  FlagPrintRaw,
					Usage: "Print properties as they are stored",
				},
			}...),
			Action: func(c *cli.Context) error {
				return DescribeWorkflow(c)
			},
		},
		{
			Name:  "list",
			Usage: "List Workflow Executions based on a Query",
			Flags: append(flagsForWorkflowFiltering, flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ListWorkflow(c)
			},
		},
		{
			Name:  "show",
			Usage: "Show Event History for a Workflow Execution",
			Flags: append(append(flagsForExecution, flagsForShowWorkflow...), flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ShowHistory(c)
			},
		},
		{
			Name:  "query",
			Usage: "Query a Workflow Execution",
			Flags: append(flagsForStackTraceQuery,
				&cli.StringFlag{
					Name:     FlagType,
					Usage:    "The query type you want to run",
					Required: true,
				}),
			Action: func(c *cli.Context) error {
				return QueryWorkflow(c)

			},
		},
		{
			Name:  "stack",
			Usage: "Query a Workflow Execution with __stack_trace as the query type",
			Flags: flagsForStackTraceQuery,
			Action: func(c *cli.Context) error {
				return QueryWorkflowUsingStackTrace(c)
			},
		},
		{
			Name:  "signal",
			Usage: "Signal Workflow Execution by Id or List Filter",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    FlagWorkflowID,
					Aliases: FlagWorkflowIDAlias,
					Usage:   "Signal Workflow Execution by Id",
				},
				&cli.StringFlag{
					Name:    FlagRunID,
					Aliases: FlagRunIDAlias,
					Usage:   "Run Id",
				},
				&cli.StringFlag{
					Name:    FlagQuery,
					Aliases: FlagQueryAlias,
					Usage:   "Signal Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/",
				},
				&cli.StringFlag{
					Name:     FlagName,
					Usage:    "Signal Name",
					Required: true,
				},
				&cli.StringFlag{
					Name:    FlagInput,
					Aliases: FlagInputAlias,
					Usage:   "Input for the signal (JSON)",
				},
				&cli.StringFlag{
					Name:  FlagInputFile,
					Usage: "Input for the signal from file (JSON)",
				},
				&cli.StringFlag{
					Name:  FlagReason,
					Usage: "Reason for signaling with List Filter",
				},
				&cli.BoolFlag{
					Name:    FlagYes,
					Aliases: FlagYesAlias,
					Usage:   "Confirm all prompts",
				},
			},
			Action: func(c *cli.Context) error {
				return SignalWorkflow(c)
			},
		},
		{
			Name:  "count",
			Usage: "Count Workflow Executions (requires ElasticSearch to be enabled)",
			Flags: getFlagsForCount(),
			Action: func(c *cli.Context) error {
				return CountWorkflow(c)
			},
		},
		{
			Name:  "cancel",
			Usage: "Cancel a Workflow Execution",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    FlagWorkflowID,
					Aliases: FlagWorkflowIDAlias,
					Usage:   "Cancel Workflow Execution by Id",
				},
				&cli.StringFlag{
					Name:    FlagRunID,
					Aliases: FlagRunIDAlias,
					Usage:   "Run Id",
				},
				&cli.StringFlag{
					Name:    FlagQuery,
					Aliases: FlagQueryAlias,
					Usage:   "Cancel Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/",
				},
				&cli.StringFlag{
					Name:  FlagReason,
					Usage: "Reason for canceling with List Filter",
				},
				&cli.BoolFlag{
					Name:    FlagYes,
					Aliases: FlagYesAlias,
					Usage:   "Confirm all prompts",
				},
			},
			Action: func(c *cli.Context) error {
				return CancelWorkflow(c)
			},
		},
		{
			Name:  "terminate",
			Usage: "Terminate Workflow Execution by Id or List Filter",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    FlagWorkflowID,
					Aliases: FlagWorkflowIDAlias,
					Usage:   "Terminate Workflow Execution by Id",
				},
				&cli.StringFlag{
					Name:    FlagRunID,
					Aliases: FlagRunIDAlias,
					Usage:   "Run Id",
				},
				&cli.StringFlag{
					Name:    FlagQuery,
					Aliases: FlagQueryAlias,
					Usage:   "Terminate Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/",
				},
				&cli.StringFlag{
					Name:  FlagReason,
					Usage: "Reason for termination",
				},
				&cli.BoolFlag{
					Name:    FlagYes,
					Aliases: FlagYesAlias,
					Usage:   "Confirm all prompts",
				},
			},
			Action: func(c *cli.Context) error {
				return TerminateWorkflow(c)
			},
		},
		{
			Name:  "delete",
			Usage: "Delete a Workflow Execution",
			Flags: flagsForExecution,
			Action: func(c *cli.Context) error {
				return DeleteWorkflow(c)
			},
		},
		{
			Name:  "reset",
			Usage: "Reset a Workflow Execution by event Id or reset type",
			Flags: append(flagsForExecution, []cli.Flag{
				&cli.StringFlag{
					Name:  FlagEventID,
					Usage: "The eventId of any event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others",
				},
				&cli.StringFlag{
					Name:     FlagReason,
					Usage:    "Reason to reset",
					Required: true,
				},
				&cli.StringFlag{
					Name:  FlagType,
					Usage: "Event type to which you want to reset: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
				},
				&cli.StringFlag{
					Name: FlagResetReapplyType,
					Usage: "Event types to reapply after the reset point: " +
						strings.Join(mapKeysToArray(resetReapplyTypesMap), ", ") + ". (default: All)",
				},
			}...),
			Action: func(c *cli.Context) error {
				return ResetWorkflow(c)
			},
		},
		{
			Name:  "reset-batch",
			Usage: "Reset a batch of Workflow Executions by reset type: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    FlagQuery,
					Aliases: FlagQueryAlias,
					Usage:   "Visibility query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query",
				}, &cli.StringFlag{
					Name:  FlagInputFile,
					Usage: "Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id",
				},
				&cli.StringFlag{
					Name:  FlagExcludeFile,
					Value: "",
					Usage: "Input file that specifies Workflow Executions to exclude from resetting",
				},
				&cli.StringFlag{
					Name:  FlagInputSeparator,
					Value: "\t",
					Usage: "Separator for the input file. The default is a tab (\t)",
				},
				&cli.StringFlag{
					Name:     FlagReason,
					Usage:    "Reason for resetting the Workflow Executions",
					Required: true,
				},
				&cli.IntFlag{
					Name:  FlagParallelism,
					Value: 1,
					Usage: "Number of goroutines to run in parallel. Each goroutine processes one line for every second",
				},
				&cli.BoolFlag{
					Name:  FlagSkipCurrentOpen,
					Usage: "Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run",
				},
				&cli.BoolFlag{
					Name: FlagSkipBaseIsNotCurrent,
					// TODO https://github.com/uber/cadence/issues/2930
					// The right way to prevent needs server side implementation .
					// This client side is only best effort
					Usage: "Skip a Workflow Execution if the base Run is not the current Run",
				},
				&cli.BoolFlag{
					Name:  FlagNonDeterministic,
					Usage: "Reset Workflow Execution only if its last Event is WorkflowTaskFailed with a nondeterministic error",
				},
				&cli.StringFlag{
					Name:     FlagType,
					Usage:    "Event type to which you want to reset: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
					Required: true,
				},
				&cli.BoolFlag{
					Name:  FlagDryRun,
					Usage: "Simulate reset without resetting any Workflow Executions",
				},
			},
			Action: func(c *cli.Context) error {
				return ResetInBatch(c)
			},
		},
		{
			Name:   "trace",
			Usage:  "Trace progress of a Workflow Execution and its children",
			Flags:  append(flagsForExecution, flagsForTraceWorkflow...),
			Action: TraceWorkflow,
		},
	}
}
