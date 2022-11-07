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

package workflow

import (
	"fmt"
	"strings"

	"github.com/temporalio/tctl-kit/pkg/flags"
	"github.com/temporalio/temporal-cli/common"
	"github.com/urfave/cli/v2"
)

func NewWorkflowCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "start",
			Usage: "Start a new Workflow Execution",
			Flags: append(common.FlagsForStartWorkflow, flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, false)
			},
		},
		{
			Name:  "execute",
			Usage: "Start a new Workflow Execution and print progress",
			Flags: append(common.FlagsForStartWorkflow, flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, true)
			},
		},
		{
			Name:  "describe",
			Usage: "Show information about a Workflow Execution",
			Flags: append(common.FlagsForExecution, []cli.Flag{
				&cli.BoolFlag{
					Name:  common.FlagResetPointsOnly,
					Usage: "Only show auto-reset points",
				},
				&cli.BoolFlag{
					Name:  common.FlagPrintRaw,
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
			Flags: append(common.FlagsForWorkflowFiltering, flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ListWorkflow(c)
			},
		},
		{
			Name:  "show",
			Usage: "Show Event History for a Workflow Execution",
			Flags: append(append(common.FlagsForExecution, common.FlagsForShowWorkflow...), flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ShowHistory(c)
			},
		},
		{
			Name:  "query",
			Usage: "Query a Workflow Execution",
			Flags: append(common.FlagsForStackTraceQuery,
				&cli.StringFlag{
					Name:     common.FlagType,
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
			Flags: common.FlagsForStackTraceQuery,
			Action: func(c *cli.Context) error {
				return QueryWorkflowUsingStackTrace(c)
			},
		},
		{
			Name:  "signal",
			Usage: "Signal Workflow Execution by Id or List Filter",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    common.FlagWorkflowID,
					Aliases: common.FlagWorkflowIDAlias,
					Usage:   "Signal Workflow Execution by Id",
				},
				&cli.StringFlag{
					Name:    common.FlagRunID,
					Aliases: common.FlagRunIDAlias,
					Usage:   "Run Id",
				},
				&cli.StringFlag{
					Name:    common.FlagQuery,
					Aliases: common.FlagQueryAlias,
					Usage:   "Signal Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/",
				},
				&cli.StringFlag{
					Name:     common.FlagName,
					Usage:    "Signal Name",
					Required: true,
				},
				&cli.StringFlag{
					Name:    common.FlagInput,
					Aliases: common.FlagInputAlias,
					Usage:   "Input for the signal (JSON)",
				},
				&cli.StringFlag{
					Name:  common.FlagInputFile,
					Usage: "Input for the signal from file (JSON)",
				},
				&cli.StringFlag{
					Name:  common.FlagReason,
					Usage: "Reason for signaling with List Filter",
				},
				&cli.BoolFlag{
					Name:    common.FlagYes,
					Aliases: common.FlagYesAlias,
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
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    common.FlagQuery,
					Aliases: common.FlagQueryAlias,
					Usage:   common.FlagQueryUsage,
				},
			},
			Action: func(c *cli.Context) error {
				return CountWorkflow(c)
			},
		},
		{
			Name:  "cancel",
			Usage: "Cancel a Workflow Execution",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    common.FlagWorkflowID,
					Aliases: common.FlagWorkflowIDAlias,
					Usage:   "Cancel Workflow Execution by Id",
				},
				&cli.StringFlag{
					Name:    common.FlagRunID,
					Aliases: common.FlagRunIDAlias,
					Usage:   "Run Id",
				},
				&cli.StringFlag{
					Name:    common.FlagQuery,
					Aliases: common.FlagQueryAlias,
					Usage:   "Cancel Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/",
				},
				&cli.StringFlag{
					Name:  common.FlagReason,
					Usage: "Reason for canceling with List Filter",
				},
				&cli.BoolFlag{
					Name:    common.FlagYes,
					Aliases: common.FlagYesAlias,
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
					Name:    common.FlagWorkflowID,
					Aliases: common.FlagWorkflowIDAlias,
					Usage:   "Terminate Workflow Execution by Id",
				},
				&cli.StringFlag{
					Name:    common.FlagRunID,
					Aliases: common.FlagRunIDAlias,
					Usage:   "Run Id",
				},
				&cli.StringFlag{
					Name:    common.FlagQuery,
					Aliases: common.FlagQueryAlias,
					Usage:   "Terminate Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/",
				},
				&cli.StringFlag{
					Name:  common.FlagReason,
					Usage: "Reason for termination",
				},
				&cli.BoolFlag{
					Name:    common.FlagYes,
					Aliases: common.FlagYesAlias,
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
			Flags: common.FlagsForExecution,
			Action: func(c *cli.Context) error {
				return DeleteWorkflow(c)
			},
		},
		{
			Name:  "reset",
			Usage: "Reset a Workflow Execution by event Id or reset type",
			Flags: append(common.FlagsForExecution, []cli.Flag{
				&cli.StringFlag{
					Name:  common.FlagEventID,
					Usage: "The eventId of any event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others",
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason to reset",
					Required: true,
				},
				&cli.StringFlag{
					Name:  common.FlagType,
					Usage: "Event type to which you want to reset: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
				},
				&cli.StringFlag{
					Name: common.FlagResetReapplyType,
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
					Name:    common.FlagQuery,
					Aliases: common.FlagQueryAlias,
					Usage:   "Visibility query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query",
				}, &cli.StringFlag{
					Name:  common.FlagInputFile,
					Usage: "Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id",
				},
				&cli.StringFlag{
					Name:  common.FlagExcludeFile,
					Value: "",
					Usage: "Input file that specifies Workflow Executions to exclude from resetting",
				},
				&cli.StringFlag{
					Name:  common.FlagInputSeparator,
					Value: "\t",
					Usage: "Separator for the input file. The default is a tab (\t)",
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason for resetting the Workflow Executions",
					Required: true,
				},
				&cli.IntFlag{
					Name:  common.FlagParallelism,
					Value: 1,
					Usage: "Number of goroutines to run in parallel. Each goroutine processes one line for every second",
				},
				&cli.BoolFlag{
					Name:  common.FlagSkipCurrentOpen,
					Usage: "Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run",
				},
				&cli.BoolFlag{
					Name: common.FlagSkipBaseIsNotCurrent,
					// TODO https://github.com/uber/cadence/issues/2930
					// The right way to prevent needs server side implementation .
					// This client side is only best effort
					Usage: "Skip a Workflow Execution if the base Run is not the current Run",
				},
				&cli.BoolFlag{
					Name:  common.FlagNonDeterministic,
					Usage: "Reset Workflow Execution only if its last Event is WorkflowTaskFailed with a nondeterministic error",
				},
				&cli.StringFlag{
					Name:     common.FlagType,
					Usage:    "Event type to which you want to reset: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
					Required: true,
				},
				&cli.BoolFlag{
					Name:  common.FlagDryRun,
					Usage: "Simulate reset without resetting any Workflow Executions",
				},
			},
			Action: func(c *cli.Context) error {
				return ResetInBatch(c)
			},
		},
		{
			Name:  "trace",
			Usage: "Trace progress of a Workflow Execution and its children",
			Flags: append(common.FlagsForExecution,
				&cli.IntFlag{
					Name:  common.FlagDepth,
					Value: -1,
					Usage: "Number of child workflows to expand, -1 to expand all child workflows",
				},
				&cli.IntFlag{
					Name:  common.FlagConcurrency,
					Value: 10,
					Usage: "Request concurrency",
				},
				&cli.StringFlag{
					Name:  common.FlagFold,
					Usage: fmt.Sprintf("Statuses for which child workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if --%s supplied", common.FlagNoFold),
					Value: "completed,canceled,terminated",
				},
				&cli.BoolFlag{
					Name:  common.FlagNoFold,
					Usage: "Disable folding. All child workflows within the set depth will be fetched and displayed",
				}),
			Action: TraceWorkflow,
		},
	}
}
