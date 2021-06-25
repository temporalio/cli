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

import "github.com/urfave/cli/v2"

// Flags used to specify cli command line arguments
const (
	FlagUsername                         = "username"
	FlagPassword                         = "password"
	FlagKeyspace                         = "keyspace"
	FlagAddress                          = "address"
	FlagAddressWithAlias                 = FlagAddress + ", ad"
	FlagHistoryAddress                   = "history-address"
	FlagDBEngine                         = "db-engine"
	FlagDBAddress                        = "db-address"
	FlagDBPort                           = "db-port"
	FlagHistoryAddressWithAlias          = FlagHistoryAddress + ", had"
	FlagNamespaceID                      = "namespace-id"
	FlagNamespace                        = "namespace"
	FlagNamespaceWithAlias               = FlagNamespace + ", ns"
	FlagShardID                          = "shard-id"
	FlagShardIDWithAlias                 = FlagShardID + ", sid"
	FlagWorkflowID                       = "workflow-id"
	FlagWorkflowIDWithAlias              = FlagWorkflowID + ", wid, w"
	FlagRunID                            = "run-id"
	FlagTreeID                           = "tree-id"
	FlagBranchID                         = "branch-id"
	FlagNumberOfShards                   = "number-of-shards"
	FlagRunIDWithAlias                   = FlagRunID + ", rid, r"
	FlagTargetCluster                    = "target-cluster"
	FlagMinEventID                       = "min-event-id"
	FlagMaxEventID                       = "max-event-id"
	FlagStartEventVersion                = "start-event-version"
	FlagTaskQueue                        = "taskqueue"
	FlagTaskQueueWithAlias               = FlagTaskQueue + ", tq"
	FlagTaskQueueType                    = "taskqueuetype"
	FlagTaskQueueTypeWithAlias           = FlagTaskQueueType + ", tqt"
	FlagWorkflowIDReusePolicy            = "workflowidreusepolicy"
	FlagWorkflowIDReusePolicyAlias       = FlagWorkflowIDReusePolicy + ", wrp"
	FlagCronSchedule                     = "cron"
	FlagWorkflowType                     = "workflow-type"
	FlagWorkflowTypeWithAlias            = FlagWorkflowType + ", wt"
	FlagWorkflowStatus                   = "status"
	FlagWorkflowStatusWithAlias          = FlagWorkflowStatus + ", s"
	FlagExecutionTimeout                 = "execution-timeout"
	FlagExecutionTimeoutWithAlias        = FlagExecutionTimeout + ", et"
	FlagWorkflowTaskTimeout              = "workflow-task-timeout"
	FlagWorkflowTaskTimeoutWithAlias     = FlagWorkflowTaskTimeout + ", wtt"
	FlagContextTimeout                   = "context-timeout"
	FlagContextTimeoutWithAlias          = FlagContextTimeout + ", ct"
	FlagInput                            = "input"
	FlagInputWithAlias                   = FlagInput + ", i"
	FlagInputFile                        = "input-file"
	FlagInputFileWithAlias               = FlagInputFile + ", if"
	FlagExcludeFile                      = "exclude-file"
	FlagInputSeparator                   = "input-separator"
	FlagParallism                        = "input-parallism"
	FlagSkipCurrentOpen                  = "skip-current-open"
	FlagSkipBaseIsNotCurrent             = "skip-base-is-not-current"
	FlagDryRun                           = "dry-run"
	FlagNonDeterministicOnly             = "only-non-deterministic"
	FlagInputTopic                       = "input-topic"
	FlagInputTopicWithAlias              = FlagInputTopic + ", it"
	FlagHostFile                         = "host-file"
	FlagCluster                          = "cluster"
	FlagInputCluster                     = "input-cluster"
	FlagStartOffset                      = "start-offset"
	FlagTopic                            = "topic"
	FlagGroup                            = "group"
	FlagResult                           = "result"
	FlagIdentity                         = "identity"
	FlagDetail                           = "detail"
	FlagReason                           = "reason"
	FlagReasonWithAlias                  = FlagReason + ", re"
	FlagOpen                             = "open"
	FlagOpenWithAlias                    = FlagOpen + ", op"
	FlagMore                             = "more"
	FlagMoreWithAlias                    = FlagMore + ", m"
	FlagAll                              = "all"
	FlagAllWithAlias                     = FlagAll + ", a"
	FlagPageSize                         = "pagesize"
	FlagPageSizeWithAlias                = FlagPageSize + ", ps"
	FlagEarliestTime                     = "earliest-time"
	FlagEarliestTimeWithAlias            = FlagEarliestTime + ", et"
	FlagLatestTime                       = "latest-time"
	FlagLatestTimeWithAlias              = FlagLatestTime + ", lt"
	FlagPrintEventVersion                = "print-event-version"
	FlagPrintEventVersionWithAlias       = FlagPrintEventVersion + ", pev"
	FlagPrintFullyDetail                 = "print-full"
	FlagPrintFullyDetailWithAlias        = FlagPrintFullyDetail + ", pf"
	FlagPrintRawTime                     = "print-raw-time"
	FlagPrintRawTimeWithAlias            = FlagPrintRawTime + ", prt"
	FlagPrintRaw                         = "print-raw"
	FlagPrintRawWithAlias                = FlagPrintRaw + ", praw"
	FlagPrintDateTime                    = "print-datetime"
	FlagPrintDateTimeWithAlias           = FlagPrintDateTime + ", pdt"
	FlagPrintMemo                        = "print-memo"
	FlagPrintMemoWithAlias               = FlagPrintMemo + ", pme"
	FlagPrintSearchAttr                  = "print-search-attr"
	FlagPrintSearchAttrWithAlias         = FlagPrintSearchAttr + ", psa"
	FlagPrintJSON                        = "print-json"
	FlagPrintJSONWithAlias               = FlagPrintJSON + ", pjson"
	FlagDescription                      = "description"
	FlagDescriptionWithAlias             = FlagDescription + ", desc"
	FlagOwnerEmail                       = "owner-email"
	FlagOwnerEmailWithAlias              = FlagOwnerEmail + ", oe"
	FlagRetentionDays                    = "retention"
	FlagRetentionDaysWithAlias           = FlagRetentionDays + ", rd"
	FlagHistoryArchivalState             = "history-archival-state"
	FlagHistoryArchivalStateWithAlias    = FlagHistoryArchivalState + ", has"
	FlagHistoryArchivalURI               = "history-uri"
	FlagHistoryArchivalURIWithAlias      = FlagHistoryArchivalURI + ", huri"
	FlagHeartbeatedWithin                = "heartbeated-within"
	FlagVisibilityArchivalState          = "visibility-archival-state"
	FlagVisibilityArchivalStateWithAlias = FlagVisibilityArchivalState + ", vas"
	FlagVisibilityArchivalURI            = "visibility-uri"
	FlagVisibilityArchivalURIWithAlias   = FlagVisibilityArchivalURI + ", vuri"
	FlagName                             = "name"
	FlagNameWithAlias                    = FlagName + ", n"
	FlagOutputFilename                   = "output-filename"
	FlagOutputFilenameWithAlias          = FlagOutputFilename + ", of"
	FlagOutputFormat                     = "output"
	FlagQueryType                        = "query-type"
	FlagQueryTypeWithAlias               = FlagQueryType + ", qt"
	FlagQueryRejectCondition             = "query-reject-condition"
	FlagQueryRejectConditionWithAlias    = FlagQueryRejectCondition + ", qrc"
	FlagShowDetail                       = "show-detail"
	FlagShowDetailWithAlias              = FlagShowDetail + ", sd"
	FlagActiveClusterName                = "active-cluster"
	FlagActiveClusterNameWithAlias       = FlagActiveClusterName + ", ac"
	FlagClusters                         = "clusters"
	FlagClustersWithAlias                = FlagClusters + ", cl"
	FlagClusterMembershipRole            = "role"
	FlagIsGlobalNamespace                = "global-namespace"
	FlagIsGlobalNamespaceWithAlias       = FlagIsGlobalNamespace + ", gd"
	FlagNamespaceData                    = "namespace-data"
	FlagNamespaceDataWithAlias           = FlagNamespaceData + ", dmd"
	FlagEventID                          = "event-id"
	FlagEventIDWithAlias                 = FlagEventID + ", eid"
	FlagActivityID                       = "activity-id"
	FlagActivityIDWithAlias              = FlagActivityID + ", aid"
	FlagMaxFieldLength                   = "max-field-length"
	FlagMaxFieldLengthWithAlias          = FlagMaxFieldLength + ", maxl"
	FlagSecurityToken                    = "security-token"
	FlagSecurityTokenWithAlias           = FlagSecurityToken + ", st"
	FlagSkipErrorMode                    = "skip-errors"
	FlagSkipErrorModeWithAlias           = FlagSkipErrorMode + ", serr"
	FlagHeadersMode                      = "headers"
	FlagHeadersModeWithAlias             = FlagHeadersMode + ", he"
	FlagMessageType                      = "message-type"
	FlagMessageTypeWithAlias             = FlagMessageType + ", mt"
	FlagURL                              = "url"
	FlagIndex                            = "index"
	FlagBatchSize                        = "batch-size"
	FlagBatchSizeWithAlias               = FlagBatchSize + ", bs"
	FlagMemoKey                          = "memo-key"
	FlagMemo                             = "memo"
	FlagMemoFile                         = "memo-file"
	FlagSearchAttributesKey              = "search-attr-key"
	FlagSearchAttributesVal              = "search-attr-value"
	FlagSearchAttributesType             = "search-attr-type"
	FlagAddBadBinary                     = "add-bad-binary"
	FlagRemoveBadBinary                  = "remove-bad-binary"
	FlagResetType                        = "reset-type"
	FlagResetPointsOnly                  = "reset-points-only"
	FlagResetBadBinaryChecksum           = "reset-bad-binary-checksum"
	FlagListQuery                        = "query"
	FlagListQueryWithAlias               = FlagListQuery + ", q"
	FlagBatchType                        = "batch-type"
	FlagBatchTypeWithAlias               = FlagBatchType + ", bt"
	FlagSignalName                       = "signal-name"
	FlagSignalNameWithAlias              = FlagSignalName + ", sig"
	FlagTaskID                           = "task-id"
	FlagTaskType                         = "task-type"
	FlagMinReadLevel                     = "min-read-level"
	FlagMaxReadLevel                     = "max-read-level"
	FlagTaskVisibilityTimestamp          = "task-timestamp"
	FlagMinVisibilityTimestamp           = "min-visibility-ts"
	FlagMaxVisibilityTimestamp           = "max-visibility-ts"
	FlagStartingRPS                      = "starting-rps"
	FlagRPS                              = "rps"
	FlagJobID                            = "job-id"
	FlagJobIDWithAlias                   = FlagJobID + ", jid"
	FlagYes                              = "yes"
	FlagServiceConfigDir                 = "service-config-dir"
	FlagServiceConfigDirWithAlias        = FlagServiceConfigDir + ", scd"
	FlagServiceEnv                       = "service-env"
	FlagServiceEnvWithAlias              = FlagServiceEnv + ", se"
	FlagServiceZone                      = "service-zone"
	FlagServiceZoneWithAlias             = FlagServiceZone + ", sz"
	FlagEnableTLS                        = "tls"
	FlagTLSCertPath                      = "tls-cert-path"
	FlagTLSKeyPath                       = "tls-key-path"
	FlagTLSCaPath                        = "tls-ca-path"
	FlagTLSDisableHostVerification       = "tls-disable-host-verification"
	FlagTLSServerName                    = "tls-server-name"
	FlagDLQType                          = "dlq-type"
	FlagDLQTypeWithAlias                 = FlagDLQType + ", dt"
	FlagMaxMessageCount                  = "max-message-count"
	FlagMaxMessageCountWithAlias         = FlagMaxMessageCount + ", mmc"
	FlagLastMessageID                    = "last-message-id"
	FlagConcurrency                      = "concurrency"
	FlagReportRate                       = "report-rate"
	FlagLowerShardBound                  = "lower-shard-bound"
	FlagUpperShardBound                  = "upper-shard-bound"
	FlagInputDirectory                   = "input-directory"
	FlagAutoConfirm                      = "auto-confirm"
	FlagDataConverterPlugin              = "data-converter-plugin"
	FlagDataConverterPluginWithAlias     = FlagDataConverterPlugin + ", dcp"
	FlagWebURL                           = "web-ui-url"
	FlagVersion                          = "version"

	FlagProtoType  = "type"
	FlagHexData    = "hex-data"
	FlagHexFile    = "hex-file"
	FlagBinaryFile = "binary-file"
	FlagBase64Data = "base64-data"
	FlagBase64File = "base64-file"
)

var flagsForExecution = []cli.Flag{
	&cli.StringFlag{
		Name:  FlagWorkflowIDWithAlias,
		Usage: "WorkflowId",
	},
	&cli.StringFlag{
		Name:  FlagRunIDWithAlias,
		Usage: "RunId",
	},
}

var flagsForShowWorkflow = []cli.Flag{
	&cli.BoolFlag{
		Name:  FlagPrintDateTimeWithAlias,
		Usage: "Print timestamp",
	},
	&cli.BoolFlag{
		Name:  FlagPrintRawTimeWithAlias,
		Usage: "Print raw timestamp",
	},
	&cli.StringFlag{
		Name:  FlagOutputFilenameWithAlias,
		Usage: "Serialize history event to a file",
	},
	&cli.BoolFlag{
		Name:  FlagPrintFullyDetailWithAlias,
		Usage: "Print fully event detail",
	},
	&cli.BoolFlag{
		Name:  FlagPrintEventVersionWithAlias,
		Usage: "Print event version",
	},
	&cli.IntFlag{
		Name:  FlagEventIDWithAlias,
		Usage: "Print specific event details",
	},
	&cli.IntFlag{
		Name:  FlagMaxFieldLengthWithAlias,
		Usage: "Maximum length for each attribute field",
		Value: defaultMaxFieldLength,
	},
	&cli.BoolFlag{
		Name:  FlagResetPointsOnly,
		Usage: "Only show events that are eligible for reset",
	},
}

var flagsForRunWorkflow = []cli.Flag{
	&cli.StringFlag{
		Name:  FlagTaskQueueWithAlias,
		Usage: "TaskQueue",
	},
	&cli.StringFlag{
		Name:  FlagWorkflowIDWithAlias,
		Usage: "WorkflowId",
	},
	&cli.StringFlag{
		Name:  FlagWorkflowTypeWithAlias,
		Usage: "WorkflowTypeName",
	},
	&cli.IntFlag{
		Name:  FlagExecutionTimeoutWithAlias,
		Usage: "Execution start to close timeout in seconds",
	},
	&cli.IntFlag{
		Name:  FlagWorkflowTaskTimeoutWithAlias,
		Value: defaultWorkflowTaskTimeoutInSeconds,
		Usage: "Workflow task start to close timeout in seconds",
	},
	&cli.StringFlag{
		Name: FlagCronSchedule,
		Usage: "Optional cron schedule for the workflow. Cron spec is as following: \n" +
			"\t┌───────────── minute (0 - 59) \n" +
			"\t│ ┌───────────── hour (0 - 23) \n" +
			"\t│ │ ┌───────────── day of the month (1 - 31) \n" +
			"\t│ │ │ ┌───────────── month (1 - 12) \n" +
			"\t│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) \n" +
			"\t│ │ │ │ │ \n" +
			"\t* * * * *",
	},
	&cli.StringFlag{
		Name: FlagWorkflowIDReusePolicyAlias,
		Usage: "Configure if the same workflow Id is allowed for use in new workflow execution. " +
			"Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate",
	},
	&cli.StringSliceFlag{
		Name: FlagInputWithAlias,
		Usage: "Optional input for the workflow in JSON format. If there are multiple parameters, pass each as a separate input flag. " +
			"Pass \"null\" for null values",
	},
	&cli.StringFlag{
		Name: FlagInputFileWithAlias,
		Usage: "Optional input for the workflow from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. " +
			"Input from file will be overwrite by input from command line",
	},
	&cli.StringFlag{
		Name:  FlagMemoKey,
		Usage: "Optional key of memo. If there are multiple keys, concatenate them and separate by space",
	},
	&cli.StringFlag{
		Name: FlagMemo,
		Usage: "Optional info that can be showed when list workflow, in JSON format. If there are multiple JSON, concatenate them and separate by space. " +
			"The order must be same as memo-key",
	},
	&cli.BoolFlag{
		Name:  FlagShowDetailWithAlias,
		Usage: "Show event details",
	},
	&cli.IntFlag{
		Name:  FlagMaxFieldLengthWithAlias,
		Usage: "Maximum length for each attribute field",
	},
	&cli.StringFlag{
		Name: FlagMemoFile,
		Usage: "Optional info that can be listed in list workflow, from JSON format file. If there are multiple JSON, concatenate them and separate by space or newline. " +
			"The order must be same as memo-key",
	},
	&cli.StringFlag{
		Name: FlagSearchAttributesKey,
		Usage: "Optional search attributes keys that can be be used in list query. If there are multiple keys, concatenate them and separate by |. " +
			"Use 'cluster get-search-attr' cmd to list legal keys.",
	},
	&cli.StringFlag{
		Name: FlagSearchAttributesVal,
		Usage: "Optional search attributes value that can be be used in list query. If there are multiple keys, concatenate them and separate by |. " +
			"If value is array, use json array like [\"a\",\"b\"], [1,2], [\"true\",\"false\"], [\"2019-06-07T17:16:34-08:00\",\"2019-06-07T18:16:34-08:00\"]. " +
			"Use 'cluster get-search-attr' cmd to list legal keys and value types",
	},
}

var flagsForWorkflowFiltering = []cli.Flag{
	&cli.BoolFlag{
		Name:  FlagOpenWithAlias,
		Usage: "List for open workflow executions, default is to list for closed ones",
	},
	&cli.StringFlag{
		Name: FlagEarliestTimeWithAlias,
		Usage: "EarliestTime of start time, supported formats are '2006-01-02T15:04:05+07:00', raw UnixNano and " +
			"time range (N<duration>), where 0 < N < 1000000 and duration (full-notation/short-notation) can be second/s, " +
			"minute/m, hour/h, day/d, week/w, month/M or year/y. For example, '15minute' or '15m' implies last 15 minutes.",
	},
	&cli.StringFlag{
		Name: FlagLatestTimeWithAlias,
		Usage: "LatestTime of start time, supported formats are '2006-01-02T15:04:05+07:00', raw UnixNano and " +
			"time range (N<duration>), where 0 < N < 1000000 and duration (in full-notation/short-notation) can be second/s, " +
			"minute/m, hour/h, day/d, week/w, month/M or year/y. For example, '15minute' or '15m' implies last 15 minutes",
	},
	&cli.StringFlag{
		Name:  FlagWorkflowIDWithAlias,
		Usage: "WorkflowId",
	},
	&cli.StringFlag{
		Name:  FlagWorkflowTypeWithAlias,
		Usage: "WorkflowTypeName",
	},
	&cli.StringFlag{
		Name:  FlagWorkflowStatusWithAlias,
		Usage: "Workflow status [completed, failed, canceled, terminated, continuedasnew, timedout]",
	},
	&cli.StringFlag{
		Name: FlagListQueryWithAlias,
		Usage: "Optional SQL like query for use of search attributes. NOTE: using query will ignore all other filter flags including: " +
			"[open, earliest_time, latest_time, workflow_id, workflow_type]",
	},
}

var flagsForWorkflowRendering = []cli.Flag{
	&cli.BoolFlag{
		Name:  FlagPrintRawTimeWithAlias,
		Usage: "Print raw timestamp",
	},
	&cli.BoolFlag{
		Name:  FlagPrintDateTimeWithAlias,
		Usage: "Print full date time in '2006-01-02T15:04:05Z07:00' format",
	},
	&cli.BoolFlag{
		Name:  FlagPrintMemoWithAlias,
		Usage: "Print memo",
	},
	&cli.BoolFlag{
		Name:  FlagPrintSearchAttrWithAlias,
		Usage: "Print search attributes",
	},
	&cli.BoolFlag{
		Name:  FlagPrintFullyDetailWithAlias,
		Usage: "Print full message without table format",
	},
}

var flagsForScan = []cli.Flag{
	&cli.StringFlag{
		Name:  FlagListQueryWithAlias,
		Usage: "Optional SQL like query",
	},
}

var flagsForListArchived = []cli.Flag{
	&cli.StringFlag{
		Name:  FlagListQueryWithAlias,
		Usage: "SQL like query. Please check the documentation of the visibility archiver used by your namespace for detailed instructions",
	},
}

func getFlagsForCount() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  FlagListQueryWithAlias,
			Usage: "Optional SQL like query. e.g count all open workflows 'CloseTime = missing'; 'WorkflowType=\"wtype\" and CloseTime > 0'",
		},
	}
}

// all flags of query except QueryType
var flagsForStackTraceQuery = []cli.Flag{
	&cli.StringFlag{
		Name:  FlagWorkflowIDWithAlias,
		Usage: "WorkflowId",
	},
	&cli.StringFlag{
		Name:  FlagRunIDWithAlias,
		Usage: "RunId",
	},
	&cli.StringFlag{
		Name:  FlagInputWithAlias,
		Usage: "Optional input for the query, in JSON format. If there are multiple parameters, concatenate them and separate by space.",
	},
	&cli.StringFlag{
		Name: FlagInputFileWithAlias,
		Usage: "Optional input for the query from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. " +
			"Input from file will be overwrite by input from command line",
	},
	&cli.StringFlag{
		Name:  FlagQueryRejectConditionWithAlias,
		Usage: "Optional flag to reject queries based on workflow state. Valid values are \"not_open\" and \"not_completed_cleanly\"",
	},
}

var flagsForQuery = append(flagsForStackTraceQuery,
	&cli.StringFlag{
		Name:  FlagQueryTypeWithAlias,
		Usage: "The query type you want to run",
	})

var flagsForDescribeWorkflow = append(flagsForExecution, []cli.Flag{
	&cli.BoolFlag{
		Name:  FlagPrintRawWithAlias,
		Usage: "Print properties as they are stored",
	},
	&cli.BoolFlag{
		Name:  FlagResetPointsOnly,
		Usage: "Only show auto-reset points",
	},
}...)

var flagsForObserveHistory = append(flagsForExecution, []cli.Flag{
	&cli.BoolFlag{
		Name:  FlagShowDetailWithAlias,
		Usage: "Optional show event details",
	},
	&cli.IntFlag{
		Name:  FlagMaxFieldLengthWithAlias,
		Usage: "Optional maximum length for each attribute field when show details",
	},
}...)

func getDBFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  FlagDBEngine,
			Value: "cassandra",
			Usage: "Type of the DB engine to use (cassandra, mysql, postgres..)",
		},
		&cli.StringFlag{
			Name:  FlagDBAddress,
			Value: "127.0.0.1",
			Usage: "persistence address",
		},
		&cli.IntFlag{
			Name:  FlagDBPort,
			Value: 9042,
			Usage: "persistence port",
		},
		&cli.StringFlag{
			Name:  FlagUsername,
			Usage: "DB username",
		},
		&cli.StringFlag{
			Name:  FlagPassword,
			Usage: "DB password",
		},
		&cli.StringFlag{
			Name:  FlagKeyspace,
			Value: "temporal",
			Usage: "DB keyspace",
		},
		&cli.BoolFlag{
			Name:  FlagEnableTLS,
			Usage: "enable TLS over the DB connection",
		},
		&cli.StringFlag{
			Name:  FlagTLSCertPath,
			Usage: "DB tls client cert path (tls must be enabled)",
		},
		&cli.StringFlag{
			Name:  FlagTLSKeyPath,
			Usage: "DB tls client key path (tls must be enabled)",
		},
		&cli.StringFlag{
			Name:  FlagTLSCaPath,
			Usage: "DB tls client ca path (tls must be enabled)",
		},
		&cli.BoolFlag{
			Name:  FlagTLSDisableHostVerification,
			Usage: "DB tls verify hostname and server cert (tls must be enabled)",
		},
	}
}
