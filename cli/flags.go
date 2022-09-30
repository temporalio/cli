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
	"fmt"

	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

// Flags used to specify cli command line arguments
var (
	FlagAddress                    = "address"
	FlagAuth                       = "auth"
	FlagNamespaceID                = "namespace-id"
	FlagNamespace                  = "namespace"
	FlagNamespaceAlias             = []string{"n"}
	FlagWorkflowID                 = "workflow-id"
	FlagWorkflowIDAlias            = []string{"w"}
	FlagRunID                      = "run-id"
	FlagRunIDAlias                 = []string{"r"}
	FlagTaskQueue                  = "task-queue"
	FlagTaskQueueAlias             = []string{"t"}
	FlagTaskQueueType              = "task-queue-type"
	FlagWorkflowIDReusePolicy      = "id-reuse-policy"
	FlagCronSchedule               = "cron"
	FlagWorkflowExecutionTimeout   = "execution-timeout"
	FlagWorkflowRunTimeout         = "run-timeout"
	FlagWorkflowTaskTimeout        = "task-timeout"
	FlagContextTimeout             = "context-timeout"
	FlagInput                      = "input"
	FlagInputAlias                 = []string{"i"}
	FlagInputFile                  = "input-file"
	FlagExcludeFile                = "exclude-file"
	FlagInputSeparator             = "input-separator"
	FlagParallelism                = "input-parallelism"
	FlagSkipCurrentOpen            = "skip-current-open"
	FlagSkipBaseIsNotCurrent       = "skip-base-is-not-current"
	FlagDryRun                     = "dry-run"
	FlagNonDeterministic           = "non-deterministic"
	FlagResult                     = "result"
	FlagIdentity                   = "identity"
	FlagDetail                     = "detail"
	FlagReason                     = "reason"
	FlagPrintRaw                   = "raw"
	FlagDescription                = "description"
	FlagOwnerEmail                 = "email"
	FlagRetention                  = "retention"
	FlagHistoryArchivalState       = "history-archival-state"
	FlagHistoryArchivalURI         = "history-uri"
	FlagVisibilityArchivalState    = "visibility-archival-state"
	FlagVisibilityArchivalURI      = "visibility-uri"
	FlagName                       = "name"
	FlagOutputFilename             = "output-filename"
	FlagQueryRejectCondition       = "reject-condition"
	FlagActiveCluster              = "active-cluster"
	FlagCluster                    = "cluster"
	FlagNamespaceData              = "data"
	FlagIsGlobalNamespace          = "global"
	FlagPromoteNamespace           = "promote-global"
	FlagEventID                    = "event-id"
	FlagActivityID                 = "activity-id"
	FlagMaxFieldLength             = "max-field-length"
	FlagMemo                       = "memo"
	FlagMemoFile                   = "memo-file"
	FlagSearchAttribute            = "search-attribute"
	FlagResetReapplyType           = "reapply-type"
	FlagResetPointsOnly            = "reset-points"
	FlagQuery                      = "query"
	FlagQueryAlias                 = []string{"q"}
	FlagQueryUsage                 = "Filter results using SQL like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for details"
	FlagArchive                    = "archived"
	FlagRPS                        = "rps"
	FlagJobID                      = "job-id"
	FlagYes                        = "yes"
	FlagYesAlias                   = []string{"y"}
	FlagTLSCertPath                = "tls-cert-path"
	FlagTLSKeyPath                 = "tls-key-path"
	FlagTLSCaPath                  = "tls-ca-path"
	FlagTLSDisableHostVerification = "tls-disable-host-verification"
	FlagTLSServerName              = "tls-server-name"
	FlagConcurrency                = "concurrency"
	FlagDataConverterPlugin        = "data-converter-plugin"
	FlagCodecAuth                  = "codec-auth"
	FlagCodecEndpoint              = "codec-endpoint"
	FlagWebURL                     = "url"
	FlagHeadersProviderPlugin      = "headers-provider-plugin"
	FlagPort                       = "port"
	FlagFollowAlias                = []string{"f"}
	FlagType                       = "type"
	FlagWorkflowType               = "workflow-type"
	FlagScheduleID                 = "schedule-id"
	FlagScheduleIDAlias            = []string{"s"}
	FlagOverlapPolicy              = "overlap-policy"
	FlagCalendar                   = "calendar"
	FlagInterval                   = "interval"
	FlagStartTime                  = "start-time"
	FlagEndTime                    = "end-time"
	FlagJitter                     = "jitter"
	FlagTimeZone                   = "time-zone"
	FlagNotes                      = "notes"
	FlagRemainingActions           = "remaining-actions"
	FlagCatchupWindow              = "catchup-window"
	FlagPauseOnFailure             = "pause-on-failure"
	FlagPause                      = "pause"
	FlagUnpause                    = "unpause"
	FlagFold                       = "fold"
	FlagNoFold                     = "no-fold"
	FlagDepth                      = "depth"
	FlagOutputAlias                = []string{"o"}
	FlagClusterAddress             = "frontend-address"
	FlagClusterEnableConnection    = "enable-connection"
)

var flagsForExecution = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagWorkflowID,
		Aliases:  FlagWorkflowIDAlias,
		Usage:    "Workflow Id",
		Required: true,
	},
	&cli.StringFlag{
		Name:    FlagRunID,
		Aliases: FlagRunIDAlias,
		Usage:   "Run Id",
	},
}

var flagsForShowWorkflow = []cli.Flag{
	&cli.StringFlag{
		Name:  FlagOutputFilename,
		Usage: "Serialize history event to a file",
	},
	&cli.IntFlag{
		Name:  FlagMaxFieldLength,
		Usage: "Maximum length for each attribute field",
		Value: defaultMaxFieldLength,
	},
	&cli.BoolFlag{
		Name:  FlagResetPointsOnly,
		Usage: "Only show events that are eligible for reset",
	},
	&cli.BoolFlag{
		Name:    output.FlagFollow,
		Aliases: FlagFollowAlias,
		Usage:   "Follow the progress of Workflow Execution",
		Value:   false,
	},
}

var flagsForStartWorkflow = append(flagsForStartWorkflowT,
	&cli.StringFlag{
		Name:     FlagType,
		Usage:    "Workflow type name",
		Required: true,
	})

var flagsForStartWorkflowLong = append(flagsForStartWorkflowT,
	&cli.StringFlag{
		Name:     FlagWorkflowType,
		Usage:    "Workflow type name",
		Required: true,
	})

var flagsForStartWorkflowT = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagWorkflowID,
		Aliases: FlagWorkflowIDAlias,
		Usage:   "Workflow Id",
	},
	&cli.StringFlag{
		Name:     FlagTaskQueue,
		Aliases:  FlagTaskQueueAlias,
		Usage:    "Task queue",
		Required: true,
	},
	&cli.IntFlag{
		Name:  FlagWorkflowRunTimeout,
		Usage: "Single workflow run timeout (seconds)",
	},
	&cli.IntFlag{
		Name:  FlagWorkflowExecutionTimeout,
		Usage: "Workflow Execution timeout, including retries and continue-as-new (seconds)",
	},
	&cli.IntFlag{
		Name:  FlagWorkflowTaskTimeout,
		Value: defaultWorkflowTaskTimeoutInSeconds,
		Usage: "Workflow task start to close timeout (seconds)",
	},
	&cli.StringFlag{
		Name: FlagCronSchedule,
		Usage: "Optional cron schedule for the Workflow. Cron spec is as following: \n" +
			"\t┌───────────── minute (0 - 59) \n" +
			"\t│ ┌───────────── hour (0 - 23) \n" +
			"\t│ │ ┌───────────── day of the month (1 - 31) \n" +
			"\t│ │ │ ┌───────────── month (1 - 12) \n" +
			"\t│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) \n" +
			"\t│ │ │ │ │ \n" +
			"\t* * * * *",
	},
	&cli.StringFlag{
		Name: FlagWorkflowIDReusePolicy,
		Usage: "Configure if the same Workflow Id is allowed for use in new Workflow Execution. " +
			"Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning",
	},
	&cli.StringSliceFlag{
		Name:    FlagInput,
		Aliases: FlagInputAlias,
		Usage:   "Optional input for the Workflow in JSON format. Pass \"null\" for null values",
	},
	&cli.StringFlag{
		Name: FlagInputFile,
		Usage: "Pass an optional input for the Workflow from a JSON file." +
			" If there are multiple JSON files, concatenate them and separate by space or newline." +
			" Input from the command line overwrites input from the file",
	},
	&cli.IntFlag{
		Name:  FlagMaxFieldLength,
		Usage: "Maximum length for each attribute field",
	},
	&cli.StringSliceFlag{
		Name:  FlagSearchAttribute,
		Usage: "Pass Search Attribute in a format key=value. Use valid JSON formats for value",
	},
	&cli.StringSliceFlag{
		Name:  FlagMemo,
		Usage: "Pass a memo in a format key=value. Use valid JSON formats for value",
	},
	&cli.StringFlag{
		Name:  FlagMemoFile,
		Usage: "Pass a memo from a file, where each line follows the format key=value. Use valid JSON formats for value",
	},
}

var flagsForWorkflowFiltering = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagQuery,
		Aliases: FlagQueryAlias,
		Usage:   FlagQueryUsage,
	},
	&cli.BoolFlag{
		Name:  FlagArchive,
		Usage: "List archived Workflow Executions (EXPERIMENTAL)",
	},
}

func getFlagsForCount() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlagQuery,
			Aliases: FlagQueryAlias,
			Usage:   FlagQueryUsage,
		},
	}
}

var flagsForStackTraceQuery = append(flagsForExecution, []cli.Flag{
	&cli.StringFlag{
		Name:    FlagInput,
		Aliases: FlagInputAlias,
		Usage:   "Optional input for the query, in JSON format. If there are multiple parameters, concatenate them and separate by space",
	},
	&cli.StringFlag{
		Name: FlagInputFile,
		Usage: "Optional input for the query from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. " +
			"Input from file will be overwrite by input from command line",
	},
	&cli.StringFlag{
		Name:  FlagQueryRejectCondition,
		Usage: "Optional flag to reject queries based on Workflow state. Valid values are \"not_open\" and \"not_completed_cleanly\"",
	},
}...)

var flagsForTraceWorkflow = []cli.Flag{
	&cli.IntFlag{
		Name:  FlagDepth,
		Value: -1,
		Usage: "Number of child workflows to expand, -1 to expand all child workflows",
	},
	&cli.IntFlag{
		Name:  FlagConcurrency,
		Value: 10,
		Usage: "Request concurrency",
	},
	&cli.StringFlag{
		Name:  FlagFold,
		Usage: fmt.Sprintf("Statuses for which child workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if --%s supplied", FlagNoFold),
		Value: "completed,canceled,terminated",
	},
	&cli.BoolFlag{
		Name:  FlagNoFold,
		Usage: "Disable folding. All child workflows within the set depth will be fetched and displayed",
	},
}
