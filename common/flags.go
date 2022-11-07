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

package common

import (
	"fmt"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

// Flags used to specify cli command line arguments
var (
	FlagActiveCluster              = "active-cluster"
	FlagActivityID                 = "activity-id"
	FlagAddress                    = "address"
	FlagArchive                    = "archived"
	FlagAuth                       = "auth"
	FlagCalendar                   = "calendar"
	FlagCatchupWindow              = "catchup-window"
	FlagCluster                    = "cluster"
	FlagClusterAddress             = "frontend-address"
	FlagClusterEnableConnection    = "enable-connection"
	FlagCodecAuth                  = "codec-auth"
	FlagCodecEndpoint              = "codec-endpoint"
	FlagConcurrency                = "concurrency"
	FlagConfig                     = "config"
	FlagContextTimeout             = "context-timeout"
	FlagCronSchedule               = "cron"
	FlagDBPath                     = "db-filename"
	FlagDepth                      = "depth"
	FlagDescription                = "description"
	FlagDetail                     = "detail"
	FlagDryRun                     = "dry-run"
	FlagDynamicConfigValue         = "dynamic-config-value"
	FlagEndTime                    = "end-time"
	FlagEnv                        = "env"
	FlagEventID                    = "event-id"
	FlagExcludeFile                = "exclude-file"
	FlagFold                       = "fold"
	FlagFollowAlias                = []string{"f"}
	FlagHeadersProviderPlugin      = "headers-provider-plugin"
	FlagHeadless                   = "headless"
	FlagHistoryArchivalState       = "history-archival-state"
	FlagHistoryArchivalURI         = "history-uri"
	FlagIdentity                   = "identity"
	FlagInput                      = "input"
	FlagInputAlias                 = []string{"i"}
	FlagInputFile                  = "input-file"
	FlagInputSeparator             = "input-separator"
	FlagInterval                   = "interval"
	FlagIP                         = "ip"
	FlagIsGlobalNamespace          = "global"
	FlagJitter                     = "jitter"
	FlagJobID                      = "job-id"
	FlagLogFormat                  = "log-format"
	FlagLogLevel                   = "log-level"
	FlagMaxFieldLength             = "max-field-length"
	FlagMemo                       = "memo"
	FlagMemoFile                   = "memo-file"
	FlagMetricsPort                = "metrics-port"
	FlagName                       = "name"
	FlagNamespace                  = "namespace"
	FlagNamespaceAlias             = []string{"n"}
	FlagNamespaceData              = "data"
	FlagNamespaceID                = "namespace-id"
	FlagNoFold                     = "no-fold"
	FlagNonDeterministic           = "non-deterministic"
	FlagNotes                      = "notes"
	FlagOutputAlias                = []string{"o"}
	FlagOutputFilename             = "output-filename"
	FlagOverlapPolicy              = "overlap-policy"
	FlagOwnerEmail                 = "email"
	FlagParallelism                = "input-parallelism"
	FlagPause                      = "pause"
	FlagPauseOnFailure             = "pause-on-failure"
	FlagPort                       = "port"
	FlagPragma                     = "sqlite-pragma"
	FlagPrintRaw                   = "raw"
	FlagPromoteNamespace           = "promote-global"
	FlagQuery                      = "query"
	FlagQueryAlias                 = []string{"q"}
	FlagQueryRejectCondition       = "reject-condition"
	FlagQueryUsage                 = "Filter results using SQL like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for details"
	FlagReason                     = "reason"
	FlagRemainingActions           = "remaining-actions"
	FlagResetPointsOnly            = "reset-points"
	FlagResetReapplyType           = "reapply-type"
	FlagResult                     = "result"
	FlagRetention                  = "retention"
	FlagRPS                        = "rps"
	FlagRunID                      = "run-id"
	FlagRunIDAlias                 = []string{"r"}
	FlagScheduleID                 = "schedule-id"
	FlagScheduleIDAlias            = []string{"s"}
	FlagSearchAttribute            = "search-attribute"
	FlagSkipBaseIsNotCurrent       = "skip-base-is-not-current"
	FlagSkipCurrentOpen            = "skip-current-open"
	FlagStartTime                  = "start-time"
	FlagTaskQueue                  = "task-queue"
	FlagTaskQueueAlias             = []string{"t"}
	FlagTaskQueueType              = "task-queue-type"
	FlagTimeZone                   = "time-zone"
	FlagTLSCaPath                  = "tls-ca-path"
	FlagTLSCertPath                = "tls-cert-path"
	FlagTLSDisableHostVerification = "tls-disable-host-verification"
	FlagTLSKeyPath                 = "tls-key-path"
	FlagTLSServerName              = "tls-server-name"
	FlagType                       = "type"
	FlagUIIP                       = "ui-ip"
	FlagUIPort                     = "ui-port"
	FlagUnpause                    = "unpause"
	FlagVisibilityArchivalState    = "visibility-archival-state"
	FlagVisibilityArchivalURI      = "visibility-uri"
	FlagWebURL                     = "url"
	FlagWorkflowExecutionTimeout   = "execution-timeout"
	FlagWorkflowID                 = "workflow-id"
	FlagWorkflowIDAlias            = []string{"w"}
	FlagWorkflowIDReusePolicy      = "id-reuse-policy"
	FlagWorkflowRunTimeout         = "run-timeout"
	FlagWorkflowTaskTimeout        = "task-timeout"
	FlagWorkflowType               = "workflow-type"
	FlagYes                        = "yes"
	FlagYesAlias                   = []string{"y"}
)

var SharedFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagEnv,
		Value:   "",
		Usage:   "Env name to read the client environment variables from",
		EnvVars: []string{"TEMPORAL_CLI_ADDRESS"},
	},
	&cli.StringFlag{
		Name:    FlagAddress,
		Value:   "",
		Usage:   "host:port for Temporal frontend service",
		EnvVars: []string{"TEMPORAL_CLI_ADDRESS"},
	},
	&cli.StringFlag{
		Name:    FlagNamespace,
		Aliases: FlagNamespaceAlias,
		Value:   "default",
		Usage:   "Temporal workflow namespace",
		EnvVars: []string{"TEMPORAL_CLI_NAMESPACE"},
	},
	&cli.StringFlag{
		Name:    FlagAuth,
		Value:   "",
		Usage:   "Authorization header to set for gRPC requests",
		EnvVars: []string{"TEMPORAL_CLI_AUTH"},
	},
	&cli.StringFlag{
		Name:    FlagTLSCertPath,
		Value:   "",
		Usage:   "Path to x509 certificate",
		EnvVars: []string{"TEMPORAL_CLI_TLS_CERT"},
	},
	&cli.StringFlag{
		Name:    FlagTLSKeyPath,
		Value:   "",
		Usage:   "Path to private key",
		EnvVars: []string{"TEMPORAL_CLI_TLS_KEY"},
	},
	&cli.StringFlag{
		Name:    FlagTLSCaPath,
		Value:   "",
		Usage:   "Path to server CA certificate",
		EnvVars: []string{"TEMPORAL_CLI_TLS_CA"},
	},
	&cli.BoolFlag{
		Name:    FlagTLSDisableHostVerification,
		Usage:   "Disable tls host name verification (tls must be enabled)",
		EnvVars: []string{"TEMPORAL_CLI_TLS_DISABLE_HOST_VERIFICATION"},
	},
	&cli.StringFlag{
		Name:    FlagTLSServerName,
		Value:   "",
		Usage:   "Override for target server name",
		EnvVars: []string{"TEMPORAL_CLI_TLS_SERVER_NAME"},
	},
	&cli.IntFlag{
		Name:    FlagContextTimeout,
		Value:   defaultContextTimeoutInSeconds,
		Usage:   "Optional timeout for context of RPC call in seconds",
		EnvVars: []string{"TEMPORAL_CONTEXT_TIMEOUT"},
	},
	&cli.StringFlag{
		Name:    FlagHeadersProviderPlugin,
		Value:   "",
		Usage:   "Headers provider plugin executable name",
		EnvVars: []string{"TEMPORAL_CLI_PLUGIN_HEADERS_PROVIDER"},
	},
	&cli.StringFlag{
		Name:    FlagCodecEndpoint,
		Value:   "",
		Usage:   "Remote Codec Server Endpoint",
		EnvVars: []string{"TEMPORAL_CLI_CODEC_ENDPOINT"},
	},
	&cli.StringFlag{
		Name:    FlagCodecAuth,
		Value:   "",
		Usage:   "Authorization header to set for requests to Codec Server",
		EnvVars: []string{"TEMPORAL_CLI_CODEC_AUTH"},
	},
	&cli.StringFlag{
		Name:  color.FlagColor,
		Usage: fmt.Sprintf("when to use color: %v, %v, %v.", color.Auto, color.Always, color.Never),
		Value: string(color.Auto),
	},
}

var FlagsForExecution = []cli.Flag{
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

var FlagsForShowWorkflow = []cli.Flag{
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

var FlagsForStartWorkflow = append(FlagsForStartWorkflowT,
	&cli.StringFlag{
		Name:     FlagType,
		Usage:    "Workflow type name",
		Required: true,
	})

var FlagsForStartWorkflowLong = append(FlagsForStartWorkflowT,
	&cli.StringFlag{
		Name:     FlagWorkflowType,
		Usage:    "Workflow type name",
		Required: true,
	})

var FlagsForStartWorkflowT = []cli.Flag{
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

var FlagsForWorkflowFiltering = []cli.Flag{
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

var FlagsForStackTraceQuery = append(FlagsForExecution, []cli.Flag{
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

// func WithFlags(commands []*cli.Command, newFlags []cli.Flag) []*cli.Command {
// 	for _, c := range commands {
// 		for _, subc := range c.Subcommands {
// 			for _, newF := range newFlags {
// 				flagExists := false
// 				for _, subf := range subc.Flags {
// 					if intersects(subf.Names(), newF.Names()) {
// 						flagExists = true
// 						continue
// 					}
// 				}

// 				if !flagExists {
// 					subc.Flags = append(subc.Flags, newF)
// 				}
// 			}
// 		}
// 	}

// 	return commands
// }

func WithFlags2(ctx *cli.Context, newFlags []cli.Flag) {
	cmd := ctx.Command

	for _, newF := range newFlags {
		flagExists := false
		for _, subf := range cmd.Flags {
			if intersects(subf.Names(), newF.Names()) {
				flagExists = true
				continue
			}
		}

		if !flagExists {
			cmd.Flags = append(cmd.Flags, newF)
		}
	}
}

func intersects(slice1 []string, slice2 []string) bool {
	for _, s1 := range slice1 {
		for _, s2 := range slice2 {
			if s1 == s2 {
				return true
			}
		}
	}
	return false
}
