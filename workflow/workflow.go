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
			Usage: common.StartWorkflowDefinition,
			UsageText: common.StartWorkflowUsageText,
			Flags: append(common.FlagsForStartWorkflow, common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, false)
			},
		},
		{
			Name:  "execute",
			Usage: common.ExecuteWorkflowDefinition,
			UsageText: common.ExecuteWorkflowUsageText,
			Flags: append(common.FlagsForStartWorkflow, common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return StartWorkflow(c, true)
			},
		},
		{
			Name:  "describe",
			Usage: common.DescribeWorkflowDefinition,
			UsageText: common.DescribeWorkflowUsageText,
			Flags: append(common.FlagsForExecution, []cli.Flag{
				&cli.BoolFlag{
					Name:     common.FlagResetPointsOnly,
					Usage:    common.FlagResetPointsUsage,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagPrintRaw,
					Usage:    common.FlagPrintRawUsage,
					Category: common.CategoryMain,
				},
			}...),
			Action: func(c *cli.Context) error {
				return DescribeWorkflow(c)
			},
		},
		{
			Name:  "list",
			Usage: common.ListWorkflowDefinition,
			UsageText: common.ListWorkflowUsageText,
			Flags: append(common.FlagsForWorkflowFiltering, common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ListWorkflow(c)
			},
		},
		{
			Name:  "show",
			Usage: common.ShowWorkflowDefinition,
			UsageText: common.WorkflowShowUsageText,
			Flags: append(append(common.FlagsForExecution, common.FlagsForShowWorkflow...), common.FlagsForPaginationAndRendering...),
			Action: func(c *cli.Context) error {
				return ShowHistory(c)
			},
		},
		{
			Name:  "query",
			Usage: common.QueryWorkflowDefinition,
			UsageText: common.QueryWorkflowUsageText,
			Flags: append(common.FlagsForStackTraceQuery,
				&cli.StringFlag{
					Name:     common.FlagType,
					Usage:    common.QueryFlagTypeUsage,
					Required: true,
					Category: common.CategoryMain,
				}),
			Action: func(c *cli.Context) error {
				return QueryWorkflow(c)

			},
		},
		{
			Name:  "stack",
			Usage: common.StackWorkflowDefinition,
			UsageText: common.WorkflowStackUsageText,
			Flags: common.FlagsForStackTraceQuery,
			Action: func(c *cli.Context) error {
				return QueryWorkflowUsingStackTrace(c)
			},
		},
		{
			Name:  "signal",
			Usage: common.SignalWorkflowDefinition,
			UsageText: common.WorkflowSignalUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    common.FlagWorkflowSignalUsage,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    common.FlagRunIdDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    common.FlagQueryDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagName,
					Usage:    common.FlagSignalName,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagInput,
					Aliases:  common.FlagInputAlias,
					Usage:    common.FlagInputSignal,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagInputFile,
					Usage:    common.FlagInputFileSignal,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    common.FlagYesDefinition,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return SignalWorkflow(c)
			},
		},
		{
			Name:  "count",
			Usage: common.CountWorkflowDefinition,
			UsageText: common.WorkflowCountUsageText,
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
			Usage: common.CancelWorkflowDefinition,
			UsageText: common.CancelWorkflowUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    common.FlagCancelWorkflow,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    common.FlagRunIdDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    common.FlagQueryDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    common.FlagYesDefinition,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return CancelWorkflow(c)
			},
		},
		{
			Name:  "terminate",
			Usage: common.TerminateWorkflowDefinition,
			UsageText: common.TerminateWorkflowUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    common.FlagWorkflowIDTerminate,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    common.FlagRunIdDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    common.FlagQueryTerminate,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    common.FlagYesDefinition,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return TerminateWorkflow(c)
			},
		},
		{
			Name:  "delete",
			Usage: common.DeleteWorkflowDefinition,
			UsageText: common.WorkflowDeleteUsageText,
			Flags: common.FlagsForExecution,
			Action: func(c *cli.Context) error {
				return DeleteWorkflow(c)
			},
		},
		{
			Name:  "reset",
			Usage: common.ResetWorkflowDefinition,
			UsageText: common.ResetWorkflowUsageText,
			Flags: append(common.FlagsForExecution, []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagEventID,
					Usage:    common.FlagEventIDDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
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
			Usage: "Reset a batch of Workflow Executions by reset type (" + strings.Join(mapKeysToArray(resetTypesMap), "), "),
			UsageText: common.ResetBatchUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagQuery,
					Aliases:  common.FlagQueryAlias,
					Usage:    common.FlagQueryResetBatch,
					Category: common.CategoryMain,
				}, &cli.StringFlag{
					Name:     common.FlagInputFile,
					Usage:    common.FlagInputFileReset,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagExcludeFile,
					Value:    "",
					Usage:    common.FlagExcludeFileDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagInputSeparator,
					Value:    "\t",
					Usage:    common.FlagInputSeparatorDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.IntFlag{
					Name:     common.FlagParallelism,
					Value:    1,
					Usage:    common.FlagParallelismDefinition,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagSkipCurrentOpen,
					Usage:    common.FlagSkipCurrentOpenDefinition,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name: common.FlagSkipBaseIsNotCurrent,
					// TODO https://github.com/uber/cadence/issues/2930
					// The right way to prevent needs server side implementation .
					// This client side is only best effort
					Usage:   common.FlagSkipBaseDefinition,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagNonDeterministic,
					Usage:    common.FlagNonDeterministicDefinition,
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
					Usage:    common.FlagDryRunDefinition,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return ResetInBatch(c)
			},
		},
		{
			Name:  "trace",
			Usage: common.TraceWorkflowDefinition,
			UsageText: common.WorkflowTraceUsageText,
			Flags: append(common.FlagsForExecution,
				&cli.IntFlag{
					Name:     common.FlagDepth,
					Value:    -1,
					Usage:    common.FlagDepthDefinition,
					Category: common.CategoryMain,
				},
				&cli.IntFlag{
					Name:     common.FlagConcurrency,
					Value:    10,
					Usage:    common.FlagConcurrencyDefinition,
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
					Usage:    common.FlagNoFoldDefinition,
					Category: common.CategoryMain,
				}),
			Action: TraceWorkflow,
		},
	}
}
