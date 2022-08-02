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
	FlagUsername                      = "username"
	FlagPassword                      = "password"
	FlagKeyspace                      = "keyspace"
	FlagAddress                       = "address"
	FlagAuth                          = "auth"
	FlagHistoryAddress                = "history-address"
	FlagDBEngine                      = "db-engine"
	FlagDBAddress                     = "db-address"
	FlagDBPort                        = "db-port"
	FlagNamespaceID                   = "namespace-id"
	FlagNamespace                     = "namespace"
	FlagNamespaceAlias                = []string{"n"}
	FlagShardID                       = "shard-id"
	FlagWorkflowID                    = "workflow-id"
	FlagWorkflowIDAlias               = []string{"wid"}
	FlagRunID                         = "run-id"
	FlagRunIDAlias                    = []string{"rid"}
	FlagTreeID                        = "tree-id"
	FlagBranchID                      = "branch-id"
	FlagNumberOfShards                = "number-of-shards"
	FlagTargetCluster                 = "target-cluster"
	FlagMinEventID                    = "min-event-id"
	FlagMaxEventID                    = "max-event-id"
	FlagStartEventVersion             = "start-event-version"
	FlagTaskQueue                     = "task-queue"
	FlagTaskQueueAlias                = []string{"tq"}
	FlagTaskQueueType                 = "task-queue-type"
	FlagTaskQueueTypeAlias            = []string{"tqt"}
	FlagWorkflowIDReusePolicy         = "workflow-id-reuse-policy"
	FlagCronSchedule                  = "cron"
	FlagWorkflowType                  = "type"
	FlagWorkflowTypeAlias             = []string{"t"}
	FlagWorkflowStatus                = "status"
	FlagWorkflowExecutionTimeout      = "execution-timeout"
	FlagWorkflowExecutionTimeoutAlias = []string{"et"}
	FlagWorkflowRunTimeout            = "run-timeout"
	FlagWorkflowRunTimeoutAlias       = []string{"rt"}
	FlagWorkflowTaskTimeout           = "task-timeout"
	FlagWorkflowTaskTimeoutAlias      = []string{"tt"}
	FlagContextTimeout                = "context-timeout"
	FlagContextTimeoutAlias           = []string{"ct"}
	FlagInput                         = "input"
	FlagInputAlias                    = []string{"i"}
	FlagInputFile                     = "input-file"
	FlagInputFileAlias                = []string{"if"}
	FlagExcludeFile                   = "exclude-file"
	FlagInputSeparator                = "input-separator"
	FlagParallelism                   = "input-parallelism"
	FlagSkipCurrentOpen               = "skip-current-open"
	FlagSkipBaseIsNotCurrent          = "skip-base-is-not-current"
	FlagDryRun                        = "dry-run"
	FlagNonDeterministic              = "non-deterministic"
	FlagHostFile                      = "host-file"
	FlagCluster                       = "cluster"
	FlagInputCluster                  = "input-cluster"
	FlagTopic                         = "topic"
	FlagGroup                         = "group"
	FlagResult                        = "result"
	FlagIdentity                      = "identity"
	FlagDetail                        = "detail"
	FlagReason                        = "reason"
	FlagReasonAlias                   = []string{"r"}
	FlagOpen                          = "open"
	FlagPageSize                      = "pagesize"
	FlagPageSizeAlias                 = []string{"ps"}
	FlagFrom                          = "from"
	FlagTo                            = "to"
	FlagPrintEventVersion             = "print-event-version"
	FlagPrintEventVersionAlias        = []string{"pev"}
	FlagPrintFullyDetail              = "print-full"
	FlagPrintFullyDetailAlias         = []string{"pf"}
	FlagPrintRawTime                  = "print-raw-time"
	FlagPrintRawTimeAlias             = []string{"prt"}
	FlagPrintRaw                      = "raw"
	FlagPrintDateTime                 = "print-datetime"
	FlagPrintDateTimeAlias            = []string{"pdt"}
	FlagPrintMemo                     = "print-memo"
	FlagPrintMemoAlias                = []string{"pme"}
	FlagPrintSearchAttr               = "print-search-attr"
	FlagPrintSearchAttrAlias          = []string{"psa"}
	FlagPrintJSON                     = "print-json"
	FlagPrintJSONAlias                = []string{"pjson"}
	FlagDescription                   = "description"
	FlagDescriptionAlias              = []string{"desc"}
	FlagOwnerEmail                    = "owner-email"
	FlagOwnerEmailAlias               = []string{"oe"}
	FlagRetention                     = "retention"
	FlagRetentionAlias                = []string{"rd"}
	FlagHistoryArchivalState          = "history-archival-state"
	FlagHistoryArchivalStateAlias     = []string{"has"}
	FlagHistoryArchivalURI            = "history-uri"
	FlagHistoryArchivalURIAlias       = []string{"huri"}
	FlagHeartbeatedWithin             = "heartbeated-within"
	FlagVisibilityArchivalState       = "visibility-archival-state"
	FlagVisibilityArchivalStateAlias  = []string{"vas"}
	FlagVisibilityArchivalURI         = "visibility-uri"
	FlagVisibilityArchivalURIAlias    = []string{"vuri"}
	FlagName                          = "name"
	FlagOutputFilename                = "output-filename"
	FlagOutputFilenameAlias           = []string{"of"}
	FlagOutputFormat                  = "output"
	FlagQueryType                     = "query-type"
	FlagQueryTypeAlias                = []string{"qt"}
	FlagQueryRejectCondition          = "query-reject-condition"
	FlagQueryRejectConditionAlias     = []string{"qrc"}
	FlagShowDetail                    = "show-detail"
	FlagShowDetailAlias               = []string{"sd"}
	FlagActiveClusterName             = "active-cluster"
	FlagActiveClusterNameAlias        = []string{"ac"}
	FlagClusters                      = "clusters"
	FlagClustersAlias                 = []string{"cl"}
	FlagClusterMembershipRole         = "role"
	FlagIsGlobalNamespace             = "global-namespace"
	FlagIsGlobalNamespaceAlias        = []string{"gn"}
	FlagNamespaceData                 = "namespace-data"
	FlagNamespaceDataAlias            = []string{"nd"}
	FlagPromoteNamespace              = "promote-namespace"
	FlagEventID                       = "event-id"
	FlagEventIDAlias                  = []string{"eid"}
	FlagActivityID                    = "activity-id"
	FlagActivityIDAlias               = []string{"aid"}
	FlagMaxFieldLength                = "max-field-length"
	FlagMaxFieldLengthAlias           = []string{"maxl"}
	FlagSecurityToken                 = "security-token"
	FlagSecurityTokenAlias            = []string{"st"}
	FlagSkipErrorMode                 = "skip-errors"
	FlagSkipErrorModeAlias            = []string{"serr"}
	FlagHeadersMode                   = "headers"
	FlagHeadersModeAlias              = []string{"he"}
	FlagMessageType                   = "message-type"
	FlagMessageTypeAlias              = []string{"mt"}
	FlagURL                           = "url"
	FlagIndex                         = "index"
	FlagBatchSize                     = "batch-size"
	FlagBatchSizeAlias                = []string{"bs"}
	FlagMemoKey                       = "memo-key"
	FlagMemo                          = "memo"
	FlagMemoFile                      = "memo-file"
	FlagSearchAttributeKey            = "search-attribute-key"
	FlagSearchAttributeValue          = "search-attribute-value"
	FlagSearchAttributeType           = "search-attribute-type"
	FlagAddBadBinary                  = "add-bad-binary"
	FlagRemoveBadBinary               = "remove-bad-binary"
	FlagResetType                     = "reset-type"
	FlagResetReapplyType              = "reset-reapply-type"
	FlagResetPointsOnly               = "reset-points-only"
	FlagResetBadBinaryChecksum        = "reset-bad-binary-checksum"
	FlagListQuery                     = "query"
	FlagListQueryAlias                = []string{"q"}
	FlagListQueryUsage                = "Filter results using SQL like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for details"
	FlagArchive                       = "archived"
	FlagArchiveAlias                  = []string{"a"}
	FlagBatchType                     = "batch-type"
	FlagBatchTypeAlias                = []string{"bt"}
	FlagSignalName                    = "signal-name"
	FlagSignalNameAlias               = []string{"sn"}
	FlagTaskID                        = "task-id"
	FlagTaskType                      = "task-type"
	FlagMinReadLevel                  = "min-read-level"
	FlagMaxReadLevel                  = "max-read-level"
	FlagTaskVisibilityTimestamp       = "task-timestamp"
	FlagMinVisibilityTimestamp        = "min-visibility-ts"
	FlagMaxVisibilityTimestamp        = "max-visibility-ts"
	FlagStartingRPS                   = "starting-rps"
	FlagRPS                           = "rps"
	FlagJobID                         = "job-id"
	FlagJobIDAlias                    = []string{"jid"}
	FlagYes                           = "yes"
	FlagYesAlias                      = []string{"y"}
	FlagServiceConfigDir              = "service-config-dir"
	FlagServiceConfigDirAlias         = []string{"scd"}
	FlagServiceEnv                    = "service-env"
	FlagServiceEnvAlias               = []string{"se"}
	FlagServiceZone                   = "service-zone"
	FlagServiceZoneAlias              = []string{"sz"}
	FlagEnableTLS                     = "tls"
	FlagTLSCertPath                   = "tls-cert-path"
	FlagTLSKeyPath                    = "tls-key-path"
	FlagTLSCaPath                     = "tls-ca-path"
	FlagTLSDisableHostVerification    = "tls-disable-host-verification"
	FlagTLSServerName                 = "tls-server-name"
	FlagLastMessageID                 = "last-message-id"
	FlagConcurrency                   = "concurrency"
	FlagConcurrencyAlias              = []string{"c"}
	FlagReportRate                    = "report-rate"
	FlagLowerShardBound               = "lower-shard-bound"
	FlagUpperShardBound               = "upper-shard-bound"
	FlagInputDirectory                = "input-directory"
	FlagDataConverterPlugin           = "data-converter-plugin"
	FlagCodecAuth                     = "codec-auth"
	FlagCodecEndpoint                 = "codec-endpoint"
	FlagWebURL                        = "web-ui-url"
	FlagHeadersProviderPlugin         = "headers-provider-plugin"
	FlagHeadersProviderPluginOptions  = "headers-provider-plugin-options"
	FlagVersion                       = "version"
	FlagPort                          = "port"
	FlagEnableConnection              = "enable-connection"
	FlagFollowAlias                   = []string{"f"}
	FlagType                          = "type"
	FlagScheduleID                    = "schedule-id"
	FlagScheduleIDAlias               = []string{"sid"}
	FlagOverlapPolicy                 = "overlap-policy"
	FlagCalendar                      = "calendar"
	FlagCalendarAlias                 = []string{"cal"}
	FlagInterval                      = "interval"
	FlagStartTime                     = "start-time"
	FlagEndTime                       = "end-time"
	FlagJitter                        = "jitter"
	FlagTimeZone                      = "time-zone"
	FlagTimeZoneAlias                 = []string{"tz"}
	FlagInitialNotes                  = "initial-notes"
	FlagInitialPaused                 = "initial-paused"
	FlagRemainingActions              = "remaining-actions"
	FlagCatchupWindow                 = "catchup-window"
	FlagPauseOnFailure                = "pause-on-failure"
	FlagPause                         = "pause"
	FlagUnpause                       = "unpause"
	FlagFold                          = "fold"
	FlagNoFold                        = "no-fold"
	FlagDepth                         = "depth"
	FlagDepthAlias                    = []string{"d"}

	FlagProtoType  = "type"
	FlagHexData    = "hex-data"
	FlagHexFile    = "hex-file"
	FlagBinaryFile = "binary-file"
	FlagBase64Data = "base64-data"
	FlagBase64File = "base64-file"
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
		Name:    FlagOutputFilename,
		Aliases: FlagOutputFilenameAlias,
		Usage:   "Serialize history event to a file",
	},
	&cli.IntFlag{
		Name:    FlagMaxFieldLength,
		Aliases: FlagMaxFieldLengthAlias,
		Usage:   "Maximum length for each attribute field",
		Value:   defaultMaxFieldLength,
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

var flagsForStartWorkflow = []cli.Flag{
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
	&cli.StringFlag{
		Name:     FlagWorkflowType,
		Aliases:  FlagWorkflowTypeAlias,
		Usage:    "Workflow type name",
		Required: true,
	},
	&cli.IntFlag{
		Name:    FlagWorkflowRunTimeout,
		Aliases: FlagWorkflowRunTimeoutAlias,
		Usage:   "Single workflow run timeout (seconds)",
	},
	&cli.IntFlag{
		Name:    FlagWorkflowExecutionTimeout,
		Aliases: FlagWorkflowExecutionTimeoutAlias,
		Usage:   "Workflow Execution timeout, including retries and continue-as-new (seconds)",
	},
	&cli.IntFlag{
		Name:    FlagWorkflowTaskTimeout,
		Aliases: FlagWorkflowTaskTimeoutAlias,
		Value:   defaultWorkflowTaskTimeoutInSeconds,
		Usage:   "Workflow task start to close timeout (seconds)",
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
		Name:    FlagInputFile,
		Aliases: FlagInputFileAlias,
		Usage: "Pass an optional input for the Workflow from a JSON file." +
			" If there are multiple JSON files, concatenate them and separate by space or newline." +
			" Input from the command line overwrites input from the file",
	},
	&cli.IntFlag{
		Name:    FlagMaxFieldLength,
		Aliases: FlagMaxFieldLengthAlias,
		Usage:   "Maximum length for each attribute field",
	},
	&cli.StringSliceFlag{
		Name:  FlagMemoKey,
		Usage: "Pass a key for an optional memo",
	},
	&cli.StringSliceFlag{
		Name:  FlagMemo,
		Usage: "Pass a memo value. A memo is information in JSON format that can be shown when the Workflow is listed",
	},
	&cli.StringFlag{
		Name:  FlagMemoFile,
		Usage: "Pass information for a memo from a JSON file. If there are multiple values, separate them by newline.",
	},
	&cli.StringSliceFlag{
		Name:  FlagSearchAttributeKey,
		Usage: "Specify a Search Attribute key. See https://docs.temporal.io/docs/concepts/what-is-a-search-attribute/",
	},
	&cli.StringSliceFlag{
		Name:  FlagSearchAttributeValue,
		Usage: "Specify a Search Attribute value. If value is an array, use JSON format, such as [\"a\",\"b\"] or [1,2], [\"true\",\"false\"]",
	},
}

var flagsForWorkflowFiltering = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagListQuery,
		Aliases: FlagListQueryAlias,
		Usage:   FlagListQueryUsage,
	},
	&cli.BoolFlag{
		Name:    FlagArchive,
		Aliases: FlagArchiveAlias,
		Usage:   "List archived Workflow Executions (EXPERIMENTAL)",
	},
}

var flagsForScan = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagListQuery,
		Aliases: FlagListQueryAlias,
		Usage:   FlagListQueryUsage,
	},
}

var flagsForListArchived = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagListQuery,
		Aliases:  FlagListQueryAlias,
		Usage:    "SQL like query. Please check the documentation of the visibility archiver used by your namespace for detailed instructions",
		Required: true,
	},
}

func getFlagsForCount() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    FlagListQuery,
			Aliases: FlagListQueryAlias,
			Usage:   FlagListQueryUsage,
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
		Name:    FlagInputFile,
		Aliases: FlagInputFileAlias,
		Usage: "Optional input for the query from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. " +
			"Input from file will be overwrite by input from command line",
	},
	&cli.StringFlag{
		Name:    FlagQueryRejectCondition,
		Aliases: FlagQueryRejectConditionAlias,
		Usage:   "Optional flag to reject queries based on Workflow state. Valid values are \"not_open\" and \"not_completed_cleanly\"",
	},
}...)

var flagsForTraceWorkflow = []cli.Flag{
	&cli.IntFlag{
		Name:    FlagDepth,
		Aliases: FlagDepthAlias,
		Value:   -1,
		Usage:   "Number of child workflows to expand, -1 to expand all child workflows",
	},
	&cli.IntFlag{
		Name:    FlagConcurrency,
		Aliases: FlagConcurrencyAlias,
		Value:   10,
		Usage:   "Request concurrency",
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
