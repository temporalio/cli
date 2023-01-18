package workflow

import (
	"fmt"
	"strings"

	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

func NewWorkflowCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "start",
			Usage: "Start a new Workflow Execution.",
			UsageText: "When invoked successfully, the Workflow and Run Ids of the recently started Workflow are returned.",
			Flags: append(common.FlagsForStartWorkflow, common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, false)
			},
		},
		{
			Name:  "execute",
			Usage: "Start a new Workflow Execution and prints its progress.",
			UsageText: "Single quotes('') are used to wrap input as JSON.",
			Flags: append(common.FlagsForStartWorkflow, common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, true)
			},
		},
		{
			Name:  "describe",
			Usage: "Show information about a Workflow Execution.",
			UsageText: "This information can be used to locate a Workflow Execution that failed.",
			Flags: append(common.FlagsForExecution, []cli.Flag{
				&cli.BoolFlag{
					Name:     common.FlagResetPointsOnly,
					Usage:    "Only show auto-reset points.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagPrintRaw,
					Usage:    "Print properties as they are stored.",
					Category: common.CategoryMain,
				},
			}...),
			Action: func(c *cli.Context) error {
				return DescribeWorkflow(c)
			},
		},
		{
			Name:  "list",
			Usage: "List Workflow Executions based on a Query.",
			UsageText: "By default, this command lists up to 10 closed Workflow Executions.",
			Flags: append(common.FlagsForWorkflowFiltering, common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ListWorkflow(c)
			},
		},
		{
			Name:  "show",
			Usage: "Show Event History for a Workflow Execution.",
			Flags: append(append(common.FlagsForExecution, common.FlagsForShowWorkflow...), common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ShowHistory(c)
			},
		},
		{
			Name:  "query",
			Usage: "Query a Workflow Execution.",
			UsageText: "Queries can retrieve all or part of the Workflow state within given parameters. Queries can also be used on completed Workflows.",
			Flags: append(common.FlagsForStackTraceQuery,
				&cli.StringFlag{
					Name:     common.FlagType,
					Usage:    "The query type you want to run.",
					Required: true,
					Category: common.CategoryMain,
				}),
			Action: func(c *cli.Context) error {
				return QueryWorkflow(c)

			},
		},
		{
			Name:  "stack",
			Usage: "Query a Workflow Execution with __stack_trace as the query type.",
			Flags: common.FlagsForStackTraceQuery,
			Action: func(c *cli.Context) error {
				return QueryWorkflowUsingStackTrace(c)
			},
		},
		{
			Name:  "signal",
			Usage: "Signal Workflow Execution by Id or List Filter.",
			UsageText: "",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    "Signal Workflow Execution by Id.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    "Run Id",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    "Signal Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagName,
					Usage:    "Signal Name",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagInput,
					Aliases:  common.FlagInputAlias,
					Usage:    "Input for the signal (JSON).",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagInputFile,
					Usage:    "Input for the signal from file (JSON).",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason for signaling with List Filter.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    "Confirm all prompts.",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return SignalWorkflow(c)
			},
		},
		{
			Name:  "count",
			Usage: "Count Workflow Executions (requires ElasticSearch to be enabled).",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    common.FlagQueryUsage,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return CountWorkflow(c)
			},
		},
		{
			Name:  "cancel",
			Usage: "Cancel a Workflow Execution.",
			UsageText: "Canceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History. A new command task will be scheduled. After cancellation, the Workflow Execution can perform cleanup work.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    "Cancel Workflow Execution by Id.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    "Run Id",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    "Cancel Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason for canceling with List Filter.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    "Confirm all prompts.",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return CancelWorkflow(c)
			},
		},
		{
			Name:  "terminate",
			Usage: "Terminate Workflow Execution by Id or List Filter.",
			UsageText: "Terminating a running Workflow records a `WorkflowExecutionTerminated` event as the closing event. Command tasks cannot be scheduled after this. ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    "Terminate Workflow Execution by Id.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    "Run Id",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    "Terminate Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason for termination.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    "Confirm all prompts.",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return TerminateWorkflow(c)
			},
		},
		{
			Name:  "delete",
			Usage: "Deletes a Workflow Execution.",
			Flags: common.FlagsForExecution,
			Action: func(c *cli.Context) error {
				return DeleteWorkflow(c)
			},
		},
		{
			Name:  "reset",
			Usage: "Resets a Workflow Execution by Event Id or reset type.",
			UsageText: "A reset allows the Workflow to be resumed from a certain point without losing your parameters or Event History.",
			Flags: append(common.FlagsForExecution, []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagEventID,
					Usage:    "The eventId of any event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason to reset.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagType,
					Usage:    "Event type to which you want to reset: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name: common.FlagResetReapplyType,
					Usage: "Event types to reapply after the reset point: " +
						strings.Join(mapKeysToArray(resetReapplyTypesMap), ", ") + ". (default: All)",
					Category: common.CategoryMain,
				},
			}...),
			Action: func(c *cli.Context) error {
				return ResetWorkflow(c)
			},
		},
		{
			Name:  "reset-batch",
			Usage: "Reset a batch of Workflow Executions by reset type: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
			UsageText: "Resetting a Workflow allows the process to resume from a certain point without losing your parameters or Event History.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    "Visibility query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query.",
					Category: common.CategoryMain,
				}, &cli.StringFlag{
					Name:     common.FlagInputFile,
					Usage:    "Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagExcludeFile,
					Value:    "",
					Usage:    "Input file that specifies Workflow Executions to exclude from resetting.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagInputSeparator,
					Value:    "\t",
					Usage:    "Separator for the input file. The default is a tab (\t).",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason for resetting the Workflow Executions.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.IntFlag{
					Name:     common.FlagParallelism,
					Value:    1,
					Usage:    "Number of goroutines to run in parallel. Each goroutine processes one line for every second.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagSkipCurrentOpen,
					Usage:    "Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name: common.FlagSkipBaseIsNotCurrent,
					// TODO https://github.com/uber/cadence/issues/2930
					// The right way to prevent needs server side implementation .
					// This client side is only best effort
					Usage:    "Skip a Workflow Execution if the base Run is not the current Run.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagNonDeterministic,
					Usage:    "Reset Workflow Execution only if its last Event is WorkflowTaskFailed with a nondeterministic error.",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagType,
					Usage:    "Event type to which you want to reset: " + strings.Join(mapKeysToArray(resetTypesMap), ", "),
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagDryRun,
					Usage:    "Simulate reset without resetting any Workflow Executions.",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return ResetInBatch(c)
			},
		},
		{
			Name:  "trace",
			Usage: "Trace progress of a Workflow Execution and its children.",
			Flags: append(common.FlagsForExecution,
				&cli.IntFlag{
					Name:     common.FlagDepth,
					Value:    -1,
					Usage:    "Number of Child Workflows to expand, -1 to expand all Child Workflows.",
					Category: common.CategoryMain,
				},
				&cli.IntFlag{
					Name:     common.FlagConcurrency,
					Value:    10,
					Usage:    "Request concurrency",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagFold,
					Usage:    fmt.Sprintf("Statuses for which Child Workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if --%s supplied.", common.FlagNoFold),
					Value:    "completed,canceled,terminated",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagNoFold,
					Usage:    "Disable folding. All Child Workflows within the set depth will be fetched and displayed.",
					Category: common.CategoryMain,
				}),
			Action: TraceWorkflow,
		},
	}
}
