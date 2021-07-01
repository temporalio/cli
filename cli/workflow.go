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

	"github.com/temporalio/tctl/pkg/flags"
	"github.com/urfave/cli/v2"
)

func newWorkflowCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "run",
			Usage: "start a new workflow execution and show progress",
			Flags: append(flagsForRunWorkflow, flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				RunWorkflow(c)
				return nil
			},
		},
		{
			Name:    "describe",
			Aliases: []string{"desc"},
			Usage:   "show information of workflow execution",
			Flags:   flagsForDescribeWorkflow,
			Action: func(c *cli.Context) error {
				DescribeWorkflow(c)
				return nil
			},
		},
		{
			Name:        "list",
			Aliases:     []string{"l"},
			Usage:       "list open or closed workflow executions",
			Description: "list one page (default size 10 items) by default, use flag --pagesize to change page size",
			Flags:       append(append(flagsForWorkflowFiltering, flagsForWorkflowRendering...), flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				ListWorkflow(c)
				return nil
			},
		},
		{
			Name:  "listarchived",
			Usage: "list archived workflow executions",
			Flags: append(append(flagsForListArchived, flagsForWorkflowRendering...), flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				ListArchivedWorkflow(c)
				return nil
			},
		},
		{
			Name:  "history",
			Usage: "show workflow history",
			Flags: append(append(flagsForExecution, flagsForShowWorkflow...), flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				ShowHistory(c)
				return nil
			},
		},
		{
			Name:    "observe",
			Aliases: []string{"ob"},
			Usage:   "show the progress of workflow history",
			Flags:   flagsForObserveHistory,
			Action: func(c *cli.Context) error {
				ObserveHistory(c)
				return nil
			},
		},
		{
			Name:  "query",
			Usage: "query workflow execution",
			Flags: flagsForQuery,
			Action: func(c *cli.Context) error {
				QueryWorkflow(c)
				return nil
			},
		},
		{
			Name:  "stack",
			Usage: "query workflow execution with __stack_trace as query type",
			Flags: flagsForStackTraceQuery,
			Action: func(c *cli.Context) error {
				QueryWorkflowUsingStackTrace(c)
				return nil
			},
		},
		{
			Name:    "signal",
			Aliases: []string{"s"},
			Usage:   "signal a workflow execution",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowId",
				},
				&cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunId",
				},
				&cli.StringFlag{
					Name:  FlagNameWithAlias,
					Usage: "SignalName",
				},
				&cli.StringFlag{
					Name:  FlagInputWithAlias,
					Usage: "Input for the signal, in JSON format.",
				},
				&cli.StringFlag{
					Name:  FlagInputFileWithAlias,
					Usage: "Input for the signal from JSON file.",
				},
			},
			Action: func(c *cli.Context) error {
				SignalWorkflow(c)
				return nil
			},
		},
		{
			Name: "scan",
			Usage: "scan workflow executions (need to enable Temporal server on ElasticSearch). " +
				"It will be faster than listall, but result are not sorted.",
			Flags: append(append(flagsForScan, flagsForWorkflowRendering...), flags.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				ScanAllWorkflow(c)
				return nil
			},
		},
		{
			Name:    "count",
			Aliases: []string{"cnt"},
			Usage:   "count number of workflow executions (need to enable Temporal server on ElasticSearch)",
			Flags:   getFlagsForCount(),
			Action: func(c *cli.Context) error {
				CountWorkflow(c)
				return nil
			},
		},
		{
			Name:    "cancel",
			Aliases: []string{"c"},
			Usage:   "cancel a workflow execution",
			Flags:   flagsForExecution,
			Action: func(c *cli.Context) error {
				CancelWorkflow(c)
				return nil
			},
		},
		{
			Name:    "terminate",
			Aliases: []string{"term"},
			Usage:   "terminate a new workflow execution",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowId",
				},
				&cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunId",
				},
				&cli.StringFlag{
					Name:  FlagReasonWithAlias,
					Usage: "The reason you want to terminate the workflow",
				},
			},
			Action: func(c *cli.Context) error {
				TerminateWorkflow(c)
				return nil
			},
		},
		{
			Name:    "reset",
			Aliases: []string{"rs"},
			Usage:   "reset the workflow, by either eventId or resetType.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  FlagWorkflowIDWithAlias,
					Usage: "WorkflowId",
				},
				&cli.StringFlag{
					Name:  FlagRunIDWithAlias,
					Usage: "RunId",
				},
				&cli.StringFlag{
					Name:  FlagEventID,
					Usage: "The eventId of any event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others",
				},
				&cli.StringFlag{
					Name:  FlagReason,
					Usage: "reason to do the reset",
				},
				&cli.StringFlag{
					Name:  FlagResetType,
					Usage: "where to reset. Support one of these: " + strings.Join(mapKeysToArray(resetTypesMap), ","),
				},
				&cli.StringFlag{
					Name:  FlagResetBadBinaryChecksum,
					Usage: "Binary checksum for resetType of BadBinary",
				},
			},
			Action: func(c *cli.Context) error {
				ResetWorkflow(c)
				return nil
			},
		},
		{
			Name: "reset-batch",
			Usage: "reset workflow in batch by resetType: " + strings.Join(mapKeysToArray(resetTypesMap), ",") +
				"To get base workflowIds/runIds to reset, source is from input file or visibility query.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  FlagInputFileWithAlias,
					Usage: "Input file to use for resetting, one workflow per line of WorkflowId and RunId. RunId is optional, default to current runId if not specified. ",
				},
				&cli.StringFlag{
					Name:  FlagListQueryWithAlias,
					Usage: "visibility query to get workflows to reset",
				},
				&cli.StringFlag{
					Name:  FlagExcludeFile,
					Value: "",
					Usage: "Another input file to use for excluding from resetting, only workflowId is needed.",
				},
				&cli.StringFlag{
					Name:  FlagInputSeparator,
					Value: "\t",
					Usage: "Separator for input file(default to tab)",
				},
				&cli.StringFlag{
					Name:  FlagReason,
					Usage: "Reason for reset",
				},
				&cli.IntFlag{
					Name:  FlagParallism,
					Value: 1,
					Usage: "Number of goroutines to run in parallel. Each goroutine would process one line for every second.",
				},
				&cli.BoolFlag{
					Name:  FlagSkipCurrentOpen,
					Usage: "Skip the workflow if the current run is open for the same workflowId as base.",
				},
				&cli.BoolFlag{
					Name: FlagSkipBaseIsNotCurrent,
					// TODO https://github.com/uber/cadence/issues/2930
					// The right way to prevent needs server side implementation .
					// This client side is only best effort
					Usage: "Skip if base run is not current run.",
				},
				&cli.BoolFlag{
					Name:  FlagNonDeterministicOnly,
					Usage: "Only apply onto workflows whose last event is workflowTaskFailed with non deterministic error.",
				},
				&cli.BoolFlag{
					Name:  FlagDryRun,
					Usage: "Not do real action of reset(just logging in STDOUT)",
				},
				&cli.StringFlag{
					Name:  FlagResetType,
					Usage: "where to reset. Support one of these: " + strings.Join(mapKeysToArray(resetTypesMap), ","),
				},
				&cli.StringFlag{
					Name:  FlagResetBadBinaryChecksum,
					Usage: "Binary checksum for resetType of BadBinary",
				},
			},
			Action: func(c *cli.Context) error {
				ResetInBatch(c)
				return nil
			},
		},
	}
}
