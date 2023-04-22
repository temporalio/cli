package common

import (
	"fmt"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/config"
	"github.com/temporalio/tctl-kit/pkg/format"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
)

// Categories used to structure --help output
var (
	CategoryGlobal  = "Shared Options:"
	CategoryDisplay = "Display Options:"
	CategoryMain    = "Command Options:"
)

// Flags used to specify cli command line arguments
var (
	FlagActiveCluster              = "active-cluster"
	FlagActivityID                 = "activity-id"
	FlagAddress                    = "address"
	FlagArchive                    = "archived"
	FlagCalendar                   = "calendar"
	FlagCatchupWindow              = "catchup-window"
	FlagCluster                    = "cluster"
	FlagClusterAddress             = "frontend-address"
	FlagClusterEnableConnection    = "enable-connection"
	FlagCodecAuth                  = "codec-auth"
	FlagCodecEndpoint              = "codec-endpoint"
	FlagConcurrency                = "concurrency"
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
	FlagMetadata                   = "grpc-meta"
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
	FlagQueryUsage                 = "Filter results using an SQL-like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for more information."
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
	FlagTLS                        = "tls"
	FlagTLSCaPath                  = "tls-ca-path"
	FlagTLSCertPath                = "tls-cert-path"
	FlagTLSDisableHostVerification = "tls-disable-host-verification"
	FlagTLSKeyPath                 = "tls-key-path"
	FlagTLSServerName              = "tls-server-name"
	FlagType                       = "type"
	FlagUIIP                       = "ui-ip"
	FlagUIAssetPath                = "ui-asset-path"
	FlagUICodecEndpoint            = "ui-codec-endpoint"
	FlagUIPort                     = "ui-port"
	FlagUnpause                    = "unpause"
	FlagVisibilityArchivalState    = "visibility-archival-state"
	FlagVisibilityArchivalURI      = "visibility-uri"
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
		Name:     FlagEnv,
		Value:    config.DefaultEnv,
		Usage:    FlagEnvDefinition,
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagAddress,
		Value:    "",
		Usage:    FlagAddrDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_ADDRESS"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagNamespace,
		Aliases:  FlagNamespaceAlias,
		Value:    "default",
		Usage:    FlagNSAliasDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_NAMESPACE"},
		Category: CategoryGlobal,
	},
	&cli.StringSliceFlag{
		Name:     FlagMetadata,
		Usage:    FlagMetadataDefinition,
		Category: CategoryGlobal,
	},
	&cli.BoolFlag{
		Name:     FlagTLS,
		Usage:    FlagTLSDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_TLS"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagTLSCertPath,
		Value:    "",
		Usage:    FlagTLSCertPathDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_TLS_CERT"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagTLSKeyPath,
		Value:    "",
		Usage:    FlagTLSKeyPathDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_TLS_KEY"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagTLSCaPath,
		Value:    "",
		Usage:    FlagTLSCaPathDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_TLS_CA"},
		Category: CategoryGlobal,
	},
	&cli.BoolFlag{
		Name:     FlagTLSDisableHostVerification,
		Usage:    FlagTLSDisableHVDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_TLS_DISABLE_HOST_VERIFICATION"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagTLSServerName,
		Value:    "",
		Usage:    FlagTLSServerNameDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_TLS_SERVER_NAME"},
		Category: CategoryGlobal,
	},
	&cli.IntFlag{
		Name:     FlagContextTimeout,
		Value:    defaultContextTimeoutInSeconds,
		Usage:    FlagContextTimeoutDefinition,
		EnvVars:  []string{"TEMPORAL_CONTEXT_TIMEOUT"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagCodecEndpoint,
		Value:    "",
		Usage:    FlagCodecEndpointDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_CODEC_ENDPOINT"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     FlagCodecAuth,
		Value:    "",
		Usage:    FlagCodecAuthDefinition,
		EnvVars:  []string{"TEMPORAL_CLI_CODEC_AUTH"},
		Category: CategoryGlobal,
	},
	&cli.StringFlag{
		Name:     color.FlagColor,
		Usage:    fmt.Sprintf("when to use color: %v, %v, %v.", color.Auto, color.Always, color.Never),
		Value:    string(color.Auto),
		Category: CategoryDisplay,
	},
}

var FlagsForExecution = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagWorkflowID,
		Aliases:  FlagWorkflowIDAlias,
		Usage:    FlagWorkflowId,
		Required: true,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagRunID,
		Aliases:  FlagRunIDAlias,
		Usage:    FlagRunIdDefinition,
		Category: CategoryMain,
	},
}

var FlagsForShowWorkflow = []cli.Flag{
	&cli.IntFlag{
		Name:     FlagMaxFieldLength,
		Usage:    FlagMaxFieldLengthDefinition,
		Value:    defaultMaxFieldLength,
		Category: CategoryMain,
	},
	&cli.BoolFlag{
		Name:     FlagResetPointsOnly,
		Usage:    FlagResetPointsOnlyDefinition,
		Category: CategoryMain,
	},
	&cli.BoolFlag{
		Name:     output.FlagFollow,
		Aliases:  FlagFollowAlias,
		Usage:    FlagFollowAliasDefinition,
		Value:    false,
		Category: CategoryMain,
	},
}

var FlagsForStartWorkflow = append(FlagsForStartWorkflowT,
	&cli.StringFlag{
		Name:     FlagType,
		Usage:    FlagWFTypeDefinition,
		Required: true,
		Category: CategoryMain,
	})

var FlagsForStartWorkflowLong = append(FlagsForStartWorkflowT,
	&cli.StringFlag{
		Name:     FlagWorkflowType,
		Usage:    FlagWFTypeDefinition,
		Required: true,
		Category: CategoryMain,
	})

var FlagsForStartWorkflowT = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagWorkflowID,
		Aliases:  FlagWorkflowIDAlias,
		Usage:    FlagWorkflowId,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagTaskQueue,
		Aliases:  FlagTaskQueueAlias,
		Usage:    FlagTaskQueueDefinition,
		Required: true,
		Category: CategoryMain,
	},
	&cli.IntFlag{
		Name:     FlagWorkflowRunTimeout,
		Usage:    FlagWorkflowRunTimeoutDefinition,
		Category: CategoryMain,
	},
	&cli.IntFlag{
		Name:     FlagWorkflowExecutionTimeout,
		Usage:    FlagWorkflowExecutionTimeoutDefinition,
		Category: CategoryMain,
	},
	&cli.IntFlag{
		Name:     FlagWorkflowTaskTimeout,
		Value:    defaultWorkflowTaskTimeoutInSeconds,
		Usage:    FlagWorkflowTaskTimeoutDefinition,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagCronSchedule,
		Usage:    FlagCronScheduleDefinition,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagWorkflowIDReusePolicy,
		Usage:    FlagWorkflowIdReusePolicyDefinition,
		Category: CategoryMain,
	},
	&cli.StringSliceFlag{
		Name:     FlagInput,
		Aliases:  FlagInputAlias,
		Usage:    FlagInputDefinition,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagInputFile,
		Usage:    FlagInputFileDefinition,
		Category: CategoryMain,
	},
	&cli.IntFlag{
		Name:     FlagMaxFieldLength,
		Usage:    FlagMaxFieldLengthDefinition,
		Category: CategoryMain,
	},
	&cli.StringSliceFlag{
		Name:     FlagSearchAttribute,
		Usage:    FlagSearchAttributeDefinition,
		Category: CategoryMain,
	},
	&cli.StringSliceFlag{
		Name:     FlagMemo,
		Usage:    FlagMemoDefinition,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagMemoFile,
		Usage:    FlagMemoFileDefinition,
		Category: CategoryMain,
	},
}

var FlagsForWorkflowFiltering = []cli.Flag{
	&cli.StringFlag{
		Name:     FlagQuery,
		Aliases:  FlagQueryAlias,
		Usage:    FlagQueryUsage,
		Category: CategoryMain,
	},
	&cli.BoolFlag{
		Name:     FlagArchive,
		Usage:    FlagArchiveDefinition,
		Category: CategoryMain,
	},
}

var FlagsForStackTraceQuery = append(FlagsForExecution, []cli.Flag{
	&cli.StringFlag{
		Name:     FlagInput,
		Aliases:  FlagInputAlias,
		Usage:    FlagInputDefinition,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagInputFile,
		Usage:    FlagInputFileDefinition,
		Category: CategoryMain,
	},
	&cli.StringFlag{
		Name:     FlagQueryRejectCondition,
		Usage:    FlagQueryRejectConditionDefinition,
		Category: CategoryMain,
	},
}...)

var FlagsForPagination = []cli.Flag{
	&cli.IntFlag{
		Name:     output.FlagLimit,
		Usage:    FlagLimitDefinition,
		Category: CategoryDisplay,
	},
	&cli.StringFlag{
		Name:     pager.FlagPager,
		Usage:    FlagPagerDefinition,
		EnvVars:  []string{"PAGER"},
		Category: CategoryDisplay,
	},
	&cli.BoolFlag{
		Name:     pager.FlagNoPager,
		Aliases:  []string{"P"},
		Usage:    FlagNoPagerDefinition,
		Category: CategoryDisplay,
	},
}

var FlagsForFormatting = []cli.Flag{
	&cli.StringFlag{
		Name:     output.FlagOutput,
		Aliases:  []string{"o"},
		Usage:    output.UsageText,
		Value:    string(output.Table),
		Category: CategoryDisplay,
	},
	&cli.StringFlag{
		Name:     format.FlagTimeFormat,
		Usage:    fmt.Sprintf("Format time as: %v, %v, %v.", format.Relative, format.ISO, format.Raw),
		Value:    string(format.Relative),
		Category: CategoryDisplay,
	},
	&cli.StringFlag{
		Name:     output.FlagFields,
		Usage:    FlagFieldsDefinition,
		Category: CategoryDisplay,
	},
}

var FlagsForPaginationAndRendering = append(FlagsForPagination, FlagsForFormatting...)

func WithFlags(commands []*cli.Command, newFlags []cli.Flag) []*cli.Command {
	for _, cmd := range commands {
		if len(cmd.Subcommands) == 0 {
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

		WithFlags(cmd.Subcommands, newFlags)
	}

	return commands
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
