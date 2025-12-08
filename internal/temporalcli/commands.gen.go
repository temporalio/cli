// Code generated. DO NOT EDIT.

package temporalcli

import (
	"fmt"

	"github.com/mattn/go-isatty"

	"github.com/spf13/cobra"

	"github.com/spf13/pflag"

	"os"

	"regexp"

	"strconv"

	"strings"

	"time"
)

var hasHighlighting = isatty.IsTerminal(os.Stdout.Fd())

type ClientOptions struct {
	Address                    string
	ClientAuthority            string
	Namespace                  string
	ApiKey                     string
	GrpcMeta                   []string
	Headers                    []string
	Tls                        bool
	TlsCertPath                string
	TlsCertData                string
	TlsKeyPath                 string
	TlsKeyData                 string
	TlsCaPath                  string
	TlsCaData                  string
	TlsDisableHostVerification bool
	TlsServerName              string
	CodecEndpoint              string
	CodecAuth                  string
	CodecHeader                []string
	Identity                   string
}

func (v *ClientOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Address, "address", "localhost:7233", "Temporal Service gRPC endpoint.")
	f.StringVar(&v.ClientAuthority, "client-authority", "", "Temporal gRPC client :authority pseudoheader.")
	f.StringVarP(&v.Namespace, "namespace", "n", "default", "Temporal Service Namespace.")
	f.StringVar(&v.ApiKey, "api-key", "", "API key for request.")
	f.StringArrayVar(&v.GrpcMeta, "grpc-meta", nil, "HTTP headers for requests. Format as a `KEY=VALUE` pair. May be passed multiple times to set multiple headers. Can also be made available via environment variable as `TEMPORAL_GRPC_META_[name]`.")
	f.StringArrayVar(&v.Headers, "headers", nil, "Temporal workflow headers in 'KEY=VALUE' format. May be passed multiple times to set multiple Temporal headers. Note: These are workflow headers, not gRPC headers.")
	f.BoolVar(&v.Tls, "tls", false, "Enable base TLS encryption. Does not have additional options like mTLS or client certs. This is defaulted to true if api-key or any other TLS options are present. Use --tls=false to explicitly disable.")
	f.StringVar(&v.TlsCertPath, "tls-cert-path", "", "Path to x509 certificate. Can't be used with --tls-cert-data.")
	f.StringVar(&v.TlsCertData, "tls-cert-data", "", "Data for x509 certificate. Can't be used with --tls-cert-path.")
	f.StringVar(&v.TlsKeyPath, "tls-key-path", "", "Path to x509 private key. Can't be used with --tls-key-data.")
	f.StringVar(&v.TlsKeyData, "tls-key-data", "", "Private certificate key data. Can't be used with --tls-key-path.")
	f.StringVar(&v.TlsCaPath, "tls-ca-path", "", "Path to server CA certificate. Can't be used with --tls-ca-data.")
	f.StringVar(&v.TlsCaData, "tls-ca-data", "", "Data for server CA certificate. Can't be used with --tls-ca-path.")
	f.BoolVar(&v.TlsDisableHostVerification, "tls-disable-host-verification", false, "Disable TLS host-name verification.")
	f.StringVar(&v.TlsServerName, "tls-server-name", "", "Override target TLS server name.")
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "Remote Codec Server endpoint.")
	f.StringVar(&v.CodecAuth, "codec-auth", "", "Authorization header for Codec Server requests.")
	f.StringArrayVar(&v.CodecHeader, "codec-header", nil, "HTTP headers for requests to codec server. Format as a `KEY=VALUE` pair. May be passed multiple times to set multiple headers.")
	f.StringVar(&v.Identity, "identity", "", "The identity of the user or client submitting this request. Defaults to \"temporal-cli:$USER@$HOST\".")
}

type OverlapPolicyOptions struct {
	OverlapPolicy StringEnum
}

func (v *OverlapPolicyOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	v.OverlapPolicy = NewStringEnum([]string{"Skip", "BufferOne", "BufferAll", "CancelOther", "TerminateOther", "AllowAll"}, "Skip")
	f.Var(&v.OverlapPolicy, "overlap-policy", "Policy for handling overlapping Workflow Executions. Accepted values: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll.")
}

type ScheduleIdOptions struct {
	ScheduleId string
}

func (v *ScheduleIdOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.ScheduleId, "schedule-id", "s", "", "Schedule ID. Required.")
	_ = cobra.MarkFlagRequired(f, "schedule-id")
}

type ScheduleConfigurationOptions struct {
	Calendar                []string
	CatchupWindow           Duration
	Cron                    []string
	EndTime                 Timestamp
	Interval                []string
	Jitter                  Duration
	Notes                   string
	Paused                  bool
	PauseOnFailure          bool
	RemainingActions        int
	StartTime               Timestamp
	TimeZone                string
	ScheduleSearchAttribute []string
	ScheduleMemo            []string
}

func (v *ScheduleConfigurationOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringArrayVar(&v.Calendar, "calendar", nil, "Calendar specification in JSON. For example: `{\"dayOfWeek\":\"Fri\",\"hour\":\"17\",\"minute\":\"5\"}`.")
	v.CatchupWindow = 0
	f.Var(&v.CatchupWindow, "catchup-window", "Maximum catch-up time for when the Service is unavailable.")
	f.StringArrayVar(&v.Cron, "cron", nil, "Calendar specification in cron string format. For example: `\"30 12 * * Fri\"`.")
	f.Var(&v.EndTime, "end-time", "Schedule end time.")
	f.StringArrayVar(&v.Interval, "interval", nil, "Interval duration. For example, 90m, or 60m/15m to include phase offset.")
	v.Jitter = 0
	f.Var(&v.Jitter, "jitter", "Max difference in time from the specification. Vary the start time randomly within this amount.")
	f.StringVar(&v.Notes, "notes", "", "Initial notes field value.")
	f.BoolVar(&v.Paused, "paused", false, "Pause the Schedule immediately on creation.")
	f.BoolVar(&v.PauseOnFailure, "pause-on-failure", false, "Pause schedule after Workflow failures.")
	f.IntVar(&v.RemainingActions, "remaining-actions", 0, "Total allowed actions. Default is zero (unlimited).")
	f.Var(&v.StartTime, "start-time", "Schedule start time.")
	f.StringVar(&v.TimeZone, "time-zone", "", "Interpret calendar specs with the `TZ` time zone. For a list of time zones, see: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones.")
	f.StringArrayVar(&v.ScheduleSearchAttribute, "schedule-search-attribute", nil, "Set schedule Search Attributes using `KEY=\"VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={\"your\": \"value\"}'. Can be passed multiple times.")
	f.StringArrayVar(&v.ScheduleMemo, "schedule-memo", nil, "Set schedule memo using `KEY=\"VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={\"your\": \"value\"}'. Can be passed multiple times.")
}

type WorkflowReferenceOptions struct {
	WorkflowId string
	RunId      string
}

func (v *WorkflowReferenceOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. Required.")
	_ = cobra.MarkFlagRequired(f, "workflow-id")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID.")
}

type DeploymentNameOptions struct {
	Name string
}

func (v *DeploymentNameOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.Name, "name", "d", "", "Name for a Worker Deployment. Required.")
	_ = cobra.MarkFlagRequired(f, "name")
}

type DeploymentVersionOptions struct {
	DeploymentName string
	BuildId        string
}

func (v *DeploymentVersionOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.DeploymentName, "deployment-name", "", "Name of the Worker Deployment. Required.")
	_ = cobra.MarkFlagRequired(f, "deployment-name")
	f.StringVar(&v.BuildId, "build-id", "", "Build ID of the Worker Deployment Version. Required.")
	_ = cobra.MarkFlagRequired(f, "build-id")
}

type DeploymentVersionOrUnversionedOptions struct {
	DeploymentName string
	BuildId        string
	Unversioned    bool
}

func (v *DeploymentVersionOrUnversionedOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.DeploymentName, "deployment-name", "", "Name of the Worker Deployment. Required.")
	_ = cobra.MarkFlagRequired(f, "deployment-name")
	f.StringVar(&v.BuildId, "build-id", "", "Build ID of the Worker Deployment Version. Required unless --unversioned is specified.")
	f.BoolVar(&v.Unversioned, "unversioned", false, "Set unversioned workers as the target version. Cannot be used with --build-id.")
}

type DeploymentReferenceOptions struct {
	SeriesName string
	BuildId    string
}

func (v *DeploymentReferenceOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.SeriesName, "series-name", "", "Series Name for a Worker Deployment. Required.")
	_ = cobra.MarkFlagRequired(f, "series-name")
	f.StringVar(&v.BuildId, "build-id", "", "Build ID for a Worker Deployment. Required.")
	_ = cobra.MarkFlagRequired(f, "build-id")
}

type SingleWorkflowOrBatchOptions struct {
	WorkflowId string
	Query      string
	RunId      string
	Reason     string
	Yes        bool
	Rps        float32
}

func (v *SingleWorkflowOrBatchOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. You must set either --workflow-id or --query.")
	f.StringVarP(&v.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter. You must set either --workflow-id or --query.")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID. Only use with --workflow-id. Cannot use with --query.")
	f.StringVar(&v.Reason, "reason", "", "Reason for batch operation. Only use with --query. Defaults to user name.")
	f.BoolVarP(&v.Yes, "yes", "y", false, "Don't prompt to confirm signaling. Only allowed when --query is present.")
	f.Float32Var(&v.Rps, "rps", 0, "Limit batch's requests per second. Only allowed if query is present.")
}

type SharedWorkflowStartOptions struct {
	WorkflowId       string
	Type             string
	TaskQueue        string
	RunTimeout       Duration
	ExecutionTimeout Duration
	TaskTimeout      Duration
	SearchAttribute  []string
	Memo             []string
	StaticSummary    string
	StaticDetails    string
	PriorityKey      int
	FairnessKey      string
	FairnessWeight   float32
}

func (v *SharedWorkflowStartOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. If not supplied, the Service generates a unique ID.")
	f.StringVar(&v.Type, "type", "", "Workflow Type name. Required. Aliased as \"--name\".")
	_ = cobra.MarkFlagRequired(f, "type")
	f.StringVarP(&v.TaskQueue, "task-queue", "t", "", "Workflow Task queue. Required.")
	_ = cobra.MarkFlagRequired(f, "task-queue")
	v.RunTimeout = 0
	f.Var(&v.RunTimeout, "run-timeout", "Fail a Workflow Run if it lasts longer than `DURATION`.")
	v.ExecutionTimeout = 0
	f.Var(&v.ExecutionTimeout, "execution-timeout", "Fail a WorkflowExecution if it lasts longer than `DURATION`. This time-out includes retries and ContinueAsNew tasks.")
	v.TaskTimeout = Duration(10000 * time.Millisecond)
	f.Var(&v.TaskTimeout, "task-timeout", "Fail a Workflow Task if it lasts longer than `DURATION`. This is the Start-to-close timeout for a Workflow Task.")
	f.StringArrayVar(&v.SearchAttribute, "search-attribute", nil, "Search Attribute in `KEY=VALUE` format. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={\"your\": \"value\"}'. Can be passed multiple times.")
	f.StringArrayVar(&v.Memo, "memo", nil, "Memo using 'KEY=\"VALUE\"' pairs. Use JSON values.")
	f.StringVar(&v.StaticSummary, "static-summary", "", "Static Workflow summary for human consumption in UIs. Uses Temporal Markdown formatting, should be a single line. EXPERIMENTAL.")
	f.StringVar(&v.StaticDetails, "static-details", "", "Static Workflow details for human consumption in UIs. Uses Temporal Markdown formatting, may be multiple lines. EXPERIMENTAL.")
	f.IntVar(&v.PriorityKey, "priority-key", 0, "Priority key (1-5, lower numbers = higher priority). Tasks in a queue should be processed in close-to-priority-order. Default is 3 when not specified.")
	f.StringVar(&v.FairnessKey, "fairness-key", "", "Fairness key (max 64 bytes) for proportional task dispatch. Tasks with same key share capacity based on their weight.")
	f.Float32Var(&v.FairnessWeight, "fairness-weight", 0, "Weight [0.001-1000] for this fairness key. Keys are dispatched proportionally to their weights.")
}

type WorkflowStartOptions struct {
	Cron             string
	FailExisting     bool
	StartDelay       Duration
	IdReusePolicy    StringEnum
	IdConflictPolicy StringEnum
}

func (v *WorkflowStartOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Cron, "cron", "", "Cron schedule for the Workflow.")
	_ = f.MarkDeprecated("cron", "Use Schedules instead.")
	f.BoolVar(&v.FailExisting, "fail-existing", false, "Fail if the Workflow already exists.")
	v.StartDelay = 0
	f.Var(&v.StartDelay, "start-delay", "Delay before starting the Workflow Execution. Can't be used with cron schedules. If the Workflow receives a signal or update prior to this time, the Workflow Execution starts immediately.")
	v.IdReusePolicy = NewStringEnum([]string{"AllowDuplicate", "AllowDuplicateFailedOnly", "RejectDuplicate", "TerminateIfRunning"}, "")
	f.Var(&v.IdReusePolicy, "id-reuse-policy", "Re-use policy for the Workflow ID in new Workflow Executions. Accepted values: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.")
	v.IdConflictPolicy = NewStringEnum([]string{"Fail", "UseExisting", "TerminateExisting"}, "")
	f.Var(&v.IdConflictPolicy, "id-conflict-policy", "Determines how to resolve a conflict when spawning a new Workflow Execution with a particular Workflow Id used by an existing Open Workflow Execution. Accepted values: Fail, UseExisting, TerminateExisting.")
}

type PayloadInputOptions struct {
	Input       []string
	InputFile   []string
	InputMeta   []string
	InputBase64 bool
}

func (v *PayloadInputOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringArrayVarP(&v.Input, "input", "i", nil, "Input value. Use JSON content or set --input-meta to override. Can't be combined with --input-file. Can be passed multiple times to pass multiple arguments.")
	f.StringArrayVar(&v.InputFile, "input-file", nil, "A path or paths for input file(s). Use JSON content or set --input-meta to override. Can't be combined with --input. Can be passed multiple times to pass multiple arguments.")
	f.StringArrayVar(&v.InputMeta, "input-meta", nil, "Input payload metadata as a `KEY=VALUE` pair. When the KEY is \"encoding\", this overrides the default (\"json/plain\"). Can be passed multiple times. Repeated metadata keys are applied to the corresponding inputs in the provided order.")
	f.BoolVar(&v.InputBase64, "input-base64", false, "Assume inputs are base64-encoded and attempt to decode them.")
}

type UpdateStartingOptions struct {
	Name                string
	FirstExecutionRunId string
	WorkflowId          string
	UpdateId            string
	RunId               string
}

func (v *UpdateStartingOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Name, "name", "", "Handler method name. Required. Aliased as \"--type\".")
	_ = cobra.MarkFlagRequired(f, "name")
	f.StringVar(&v.FirstExecutionRunId, "first-execution-run-id", "", "Parent Run ID. The update is sent to the last Workflow Execution in the chain started with this Run ID.")
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. Required.")
	_ = cobra.MarkFlagRequired(f, "workflow-id")
	f.StringVar(&v.UpdateId, "update-id", "", "Update ID. If unset, defaults to a UUID.")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID. If unset, looks for an Update against the currently-running Workflow Execution.")
}

type UpdateTargetingOptions struct {
	WorkflowId string
	UpdateId   string
	RunId      string
}

func (v *UpdateTargetingOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. Required.")
	_ = cobra.MarkFlagRequired(f, "workflow-id")
	f.StringVar(&v.UpdateId, "update-id", "", "Update ID. Must be unique per Workflow Execution. Required.")
	_ = cobra.MarkFlagRequired(f, "update-id")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID. If unset, updates the currently-running Workflow Execution.")
}

type NexusEndpointIdentityOptions struct {
	Name string
}

func (v *NexusEndpointIdentityOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Name, "name", "", "Endpoint name. Required.")
	_ = cobra.MarkFlagRequired(f, "name")
}

type NexusEndpointConfigOptions struct {
	Description     string
	DescriptionFile string
	TargetNamespace string
	TargetTaskQueue string
	TargetUrl       string
}

func (v *NexusEndpointConfigOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Description, "description", "", "Nexus Endpoint description. You may use Markdown formatting in the Nexus Endpoint description.")
	f.StringVar(&v.DescriptionFile, "description-file", "", "Path to the Nexus Endpoint description file. The contents of the description file may use Markdown formatting.")
	f.StringVar(&v.TargetNamespace, "target-namespace", "", "Namespace where a handler Worker polls for Nexus tasks.")
	f.StringVar(&v.TargetTaskQueue, "target-task-queue", "", "Task Queue that a handler Worker polls for Nexus tasks.")
	f.StringVar(&v.TargetUrl, "target-url", "", "An external Nexus Endpoint that receives forwarded Nexus requests. May be used as an alternative to `--target-namespace` and `--target-task-queue`. EXPERIMENTAL.")
}

type QueryModifiersOptions struct {
	RejectCondition StringEnum
}

func (v *QueryModifiersOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	v.RejectCondition = NewStringEnum([]string{"not_open", "not_completed_cleanly"}, "")
	f.Var(&v.RejectCondition, "reject-condition", "Optional flag for rejecting Queries based on Workflow state. Accepted values: not_open, not_completed_cleanly.")
}

type WorkflowUpdateOptionsOptions struct {
	VersioningOverrideBehavior       StringEnum
	VersioningOverrideDeploymentName string
	VersioningOverrideBuildId        string
}

func (v *WorkflowUpdateOptionsOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	v.VersioningOverrideBehavior = NewStringEnum([]string{"pinned", "auto_upgrade"}, "")
	f.Var(&v.VersioningOverrideBehavior, "versioning-override-behavior", "Override the versioning behavior of a Workflow. Accepted values: pinned, auto_upgrade. Required.")
	_ = cobra.MarkFlagRequired(f, "versioning-override-behavior")
	f.StringVar(&v.VersioningOverrideDeploymentName, "versioning-override-deployment-name", "", "When overriding to a `pinned` behavior, specifies the Deployment Name of the version to target.")
	f.StringVar(&v.VersioningOverrideBuildId, "versioning-override-build-id", "", "When overriding to a `pinned` behavior, specifies the Build ID of the version to target.")
}

type TemporalCommand struct {
	Command                 cobra.Command
	Env                     string
	EnvFile                 string
	ConfigFile              string
	Profile                 string
	DisableConfigFile       bool
	DisableConfigEnv        bool
	LogLevel                StringEnum
	LogFormat               StringEnum
	Output                  StringEnum
	TimeFormat              StringEnum
	Color                   StringEnum
	NoJsonShorthandPayloads bool
	CommandTimeout          Duration
	ClientConnectTimeout    Duration
}

func NewTemporalCommand(cctx *CommandContext) *TemporalCommand {
	var s TemporalCommand
	s.Command.Use = "temporal"
	s.Command.Short = "Temporal command-line interface and development server"
	if hasHighlighting {
		s.Command.Long = "The Temporal CLI manages, monitors, and debugs Temporal apps. It lets you run\na local Temporal Service, start Workflow Executions, pass messages to running\nWorkflows, inspect state, and more.\n\n* Start a local development service:\n      \x1b[1mtemporal server start-dev\x1b[0m\n* View help: pass \x1b[1m--help\x1b[0m to any command:\n      \x1b[1mtemporal activity complete --help\x1b[0m"
	} else {
		s.Command.Long = "The Temporal CLI manages, monitors, and debugs Temporal apps. It lets you run\na local Temporal Service, start Workflow Executions, pass messages to running\nWorkflows, inspect state, and more.\n\n* Start a local development service:\n      `temporal server start-dev`\n* View help: pass `--help` to any command:\n      `temporal activity complete --help`"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalActivityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalBatchCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalConfigCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalServerCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVar(&s.Env, "env", "default", "Active environment name (`ENV`).")
	cctx.BindFlagEnvVar(s.Command.PersistentFlags().Lookup("env"), "TEMPORAL_ENV")
	s.Command.PersistentFlags().StringVar(&s.EnvFile, "env-file", "", "Path to environment settings file. Defaults to `$HOME/.config/temporalio/temporal.yaml`.")
	s.Command.PersistentFlags().StringVar(&s.ConfigFile, "config-file", "", "File path to read TOML config from, defaults to `$CONFIG_PATH/temporal/temporal.toml` where `$CONFIG_PATH` is defined as `$HOME/.config` on Unix, `$HOME/Library/Application Support` on macOS, and `%AppData%` on Windows. EXPERIMENTAL.")
	s.Command.PersistentFlags().StringVar(&s.Profile, "profile", "", "Profile to use for config file. EXPERIMENTAL.")
	s.Command.PersistentFlags().BoolVar(&s.DisableConfigFile, "disable-config-file", false, "If set, disables loading environment config from config file. EXPERIMENTAL.")
	s.Command.PersistentFlags().BoolVar(&s.DisableConfigEnv, "disable-config-env", false, "If set, disables loading environment config from environment variables. EXPERIMENTAL.")
	s.LogLevel = NewStringEnum([]string{"debug", "info", "warn", "error", "never"}, "info")
	s.Command.PersistentFlags().Var(&s.LogLevel, "log-level", "Log level. Default is \"info\" for most commands and \"warn\" for `server start-dev`. Accepted values: debug, info, warn, error, never.")
	s.LogFormat = NewStringEnum([]string{"text", "json", "pretty"}, "text")
	s.Command.PersistentFlags().Var(&s.LogFormat, "log-format", "Log format. Accepted values: text, json.")
	s.Output = NewStringEnum([]string{"text", "json", "jsonl", "none"}, "text")
	s.Command.PersistentFlags().VarP(&s.Output, "output", "o", "Non-logging data output format. Accepted values: text, json, jsonl, none.")
	s.TimeFormat = NewStringEnum([]string{"relative", "iso", "raw"}, "relative")
	s.Command.PersistentFlags().Var(&s.TimeFormat, "time-format", "Time format. Accepted values: relative, iso, raw.")
	s.Color = NewStringEnum([]string{"always", "never", "auto"}, "auto")
	s.Command.PersistentFlags().Var(&s.Color, "color", "Output coloring. Accepted values: always, never, auto.")
	s.Command.PersistentFlags().BoolVar(&s.NoJsonShorthandPayloads, "no-json-shorthand-payloads", false, "Raw payload output, even if the JSON option was used.")
	s.CommandTimeout = 0
	s.Command.PersistentFlags().Var(&s.CommandTimeout, "command-timeout", "The command execution timeout. 0s means no timeout.")
	s.ClientConnectTimeout = 0
	s.Command.PersistentFlags().Var(&s.ClientConnectTimeout, "client-connect-timeout", "The client connection timeout. 0s means no timeout.")
	s.initCommand(cctx)
	return &s
}

type TemporalActivityCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
	ClientOptions
}

func NewTemporalActivityCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalActivityCommand {
	var s TemporalActivityCommand
	s.Parent = parent
	s.Command.Use = "activity"
	s.Command.Short = "Complete, update, pause, unpause, reset or fail an Activity"
	if hasHighlighting {
		s.Command.Long = "Update an Activity's options, manage activity lifecycle or update\nan Activity's state to completed or failed.\n\nUpdating activity state marks an Activity as successfully finished or as\nhaving encountered an error.\n\n\x1b[1mtemporal activity complete \\\n    --activity-id=YourActivityId \\\n    --workflow-id=YourWorkflowId \\\n    --result='{\"YourResultKey\": \"YourResultValue\"}'\x1b[0m"
	} else {
		s.Command.Long = "Update an Activity's options, manage activity lifecycle or update\nan Activity's state to completed or failed.\n\nUpdating activity state marks an Activity as successfully finished or as\nhaving encountered an error.\n\n```\ntemporal activity complete \\\n    --activity-id=YourActivityId \\\n    --workflow-id=YourWorkflowId \\\n    --result='{\"YourResultKey\": \"YourResultValue\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalActivityCompleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalActivityFailCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalActivityPauseCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalActivityResetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalActivityUnpauseCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalActivityUpdateOptionsCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalActivityCompleteCommand struct {
	Parent  *TemporalActivityCommand
	Command cobra.Command
	WorkflowReferenceOptions
	ActivityId string
	Result     string
}

func NewTemporalActivityCompleteCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityCompleteCommand {
	var s TemporalActivityCompleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "complete [flags]"
	s.Command.Short = "Complete an Activity"
	if hasHighlighting {
		s.Command.Long = "Complete an Activity, marking it as successfully finished. Specify the\nActivity ID and include a JSON result for the returned value:\n\n\x1b[1mtemporal activity complete \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --result '{\"YourResultKey\": \"YourResultVal\"}'\x1b[0m"
	} else {
		s.Command.Long = "Complete an Activity, marking it as successfully finished. Specify the\nActivity ID and include a JSON result for the returned value:\n\n```\ntemporal activity complete \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --result '{\"YourResultKey\": \"YourResultVal\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "Activity ID to complete. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Result, "result", "", "Result `JSON` to return. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "result")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalActivityFailCommand struct {
	Parent  *TemporalActivityCommand
	Command cobra.Command
	WorkflowReferenceOptions
	ActivityId string
	Detail     string
	Reason     string
}

func NewTemporalActivityFailCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityFailCommand {
	var s TemporalActivityFailCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "fail [flags]"
	s.Command.Short = "Fail an Activity"
	if hasHighlighting {
		s.Command.Long = "Fail an Activity, marking it as having encountered an error. Specify the\nActivity and Workflow IDs:\n\n\x1b[1mtemporal activity fail \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\x1b[0m"
	} else {
		s.Command.Long = "Fail an Activity, marking it as having encountered an error. Specify the\nActivity and Workflow IDs:\n\n```\ntemporal activity fail \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "Activity ID to fail. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Detail, "detail", "", "Reason for failing the Activity (JSON).")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for failing the Activity.")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalActivityPauseCommand struct {
	Parent  *TemporalActivityCommand
	Command cobra.Command
	WorkflowReferenceOptions
	ActivityId   string
	ActivityType string
	Identity     string
}

func NewTemporalActivityPauseCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityPauseCommand {
	var s TemporalActivityPauseCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "pause [flags]"
	s.Command.Short = "Pause an Activity"
	if hasHighlighting {
		s.Command.Long = "Pause an Activity.\n\nIf the Activity is not currently running (e.g. because it previously\nfailed), it will not be run again until it is unpaused.\n\nHowever, if the Activity is currently running, it will run until the next\ntime it fails, completes, or times out, at which point the pause will kick in.\n\nIf the Activity is on its last retry attempt and fails, the failure will\nbe returned to the caller, just as if the Activity had not been paused.\n\nActivities should be specified either by their Activity ID or Activity Type.\n\nFor example, specify the Activity and Workflow IDs like this:\n\n\x1b[1mtemporal activity pause \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nTo later unpause the activity, see unpause. You may also want to\nreset the activity to unpause it while also starting it from the beginning."
	} else {
		s.Command.Long = "Pause an Activity.\n\nIf the Activity is not currently running (e.g. because it previously\nfailed), it will not be run again until it is unpaused.\n\nHowever, if the Activity is currently running, it will run until the next\ntime it fails, completes, or times out, at which point the pause will kick in.\n\nIf the Activity is on its last retry attempt and fails, the failure will\nbe returned to the caller, just as if the Activity had not been paused.\n\nActivities should be specified either by their Activity ID or Activity Type.\n\nFor example, specify the Activity and Workflow IDs like this:\n\n```\ntemporal activity pause \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n```\n\nTo later unpause the activity, see unpause. You may also want to\nreset the activity to unpause it while also starting it from the beginning."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.ActivityId, "activity-id", "a", "", "The Activity ID to pause. Either `activity-id` or `activity-type` must be provided, but not both.")
	s.Command.Flags().StringVar(&s.ActivityType, "activity-type", "", "All activities of the Activity Type will be paused. Either `activity-id` or `activity-type` must be provided, but not both. Note: Pausing Activity by Type is an experimental feature and may change in the future.")
	s.Command.Flags().StringVar(&s.Identity, "identity", "", "The identity of the user or client submitting this request.")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalActivityResetCommand struct {
	Parent  *TemporalActivityCommand
	Command cobra.Command
	SingleWorkflowOrBatchOptions
	ActivityId             string
	ActivityType           string
	KeepPaused             bool
	ResetAttempts          bool
	ResetHeartbeats        bool
	MatchAll               bool
	Jitter                 Duration
	RestoreOriginalOptions bool
}

func NewTemporalActivityResetCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityResetCommand {
	var s TemporalActivityResetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "reset [flags]"
	s.Command.Short = "Reset an Activity"
	if hasHighlighting {
		s.Command.Long = "Reset an activity. This restarts the activity as if it were first being\nscheduled. That is, it will reset both the number of attempts and the\nactivity timeout, as well as, optionally, the\nheartbeat details.\n\nIf the activity may be executing (i.e. it has not yet timed out), the\nreset will take effect the next time it fails, heartbeats, or times out.\nIf is waiting for a retry (i.e. has failed or timed out), the reset\nwill apply immediately.\n\nIf the activity is already paused, it will be unpaused by default.\nYou can specify \x1b[1mkeep_paused\x1b[0m to prevent this.\n\nIf the activity is paused and the \x1b[1mkeep_paused\x1b[0m flag is not provided,\nit will be unpaused. If the activity is paused and \x1b[1mkeep_paused\x1b[0m flag\nis provided - it will stay paused.\n\nActivities can be specified by their Activity ID or Activity Type.\n\n### Resetting activities that heartbeat {#reset-heartbeats}\n\nActivities that heartbeat will receive a Canceled failure\nthe next time they heartbeat after a reset.\n\nIf, in your Activity, you need to do any cleanup when an Activity is\nreset, handle this error and then re-throw it when you've cleaned up.\n\nIf the \x1b[1mreset_heartbeats\x1b[0m flag is set, the heartbeat details will also be cleared.\n\nSpecify the Activity Type of ID and Workflow IDs:\n\n\x1b[1mtemporal activity reset \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n    --keep-paused\n    --reset-heartbeats\x1b[0m\n\nEither \x1b[1mactivity-id\x1b[0m, \x1b[1mactivity-type\x1b[0m, or \x1b[1m--match-all\x1b[0m must be specified.\n\nActivities can be reset in bulk with a visibility query list filter. \nFor example, if you want to reset activities of type Foo:\n\n\x1b[1mtemporal activity reset \\\n    --query 'TemporalResetInfo=\"property:activityType=Foo\"'\x1b[0m"
	} else {
		s.Command.Long = "Reset an activity. This restarts the activity as if it were first being\nscheduled. That is, it will reset both the number of attempts and the\nactivity timeout, as well as, optionally, the\nheartbeat details.\n\nIf the activity may be executing (i.e. it has not yet timed out), the\nreset will take effect the next time it fails, heartbeats, or times out.\nIf is waiting for a retry (i.e. has failed or timed out), the reset\nwill apply immediately.\n\nIf the activity is already paused, it will be unpaused by default.\nYou can specify `keep_paused` to prevent this.\n\nIf the activity is paused and the `keep_paused` flag is not provided,\nit will be unpaused. If the activity is paused and `keep_paused` flag\nis provided - it will stay paused.\n\nActivities can be specified by their Activity ID or Activity Type.\n\n### Resetting activities that heartbeat {#reset-heartbeats}\n\nActivities that heartbeat will receive a Canceled failure\nthe next time they heartbeat after a reset.\n\nIf, in your Activity, you need to do any cleanup when an Activity is\nreset, handle this error and then re-throw it when you've cleaned up.\n\nIf the `reset_heartbeats` flag is set, the heartbeat details will also be cleared.\n\nSpecify the Activity Type of ID and Workflow IDs:\n\n```\ntemporal activity reset \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n    --keep-paused\n    --reset-heartbeats\n```\n\nEither `activity-id`, `activity-type`, or `--match-all` must be specified.\n\nActivities can be reset in bulk with a visibility query list filter. \nFor example, if you want to reset activities of type Foo:\n\n```\ntemporal activity reset \\\n    --query 'TemporalResetInfo=\"property:activityType=Foo\"'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.ActivityId, "activity-id", "a", "", "The Activity ID to reset.  Mutually exclusive with `--query`, `--match-all`, and `--activity-type`. Requires `--workflow-id` to be specified.")
	s.Command.Flags().StringVar(&s.ActivityType, "activity-type", "", "Activities of this Type will be reset. Mutually exclusive with `--match-all` and `activity-id`.")
	s.Command.Flags().BoolVar(&s.KeepPaused, "keep-paused", false, "If the activity was paused, it will stay paused.")
	s.Command.Flags().BoolVar(&s.ResetAttempts, "reset-attempts", false, "Reset the activity attempts.")
	s.Command.Flags().BoolVar(&s.ResetHeartbeats, "reset-heartbeats", false, "Reset the Activity's heartbeats. Only works with --reset-attempts.")
	s.Command.Flags().BoolVar(&s.MatchAll, "match-all", false, "Every activity should be reset. Every activity should be updated. Mutually exclusive with `--activity-id` and `--activity-type`.")
	s.Jitter = 0
	s.Command.Flags().Var(&s.Jitter, "jitter", "The activity will reset at random a time within the specified duration. Can only be used with --query.")
	s.Command.Flags().BoolVar(&s.RestoreOriginalOptions, "restore-original-options", false, "Restore the original options of the activity.")
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalActivityUnpauseCommand struct {
	Parent  *TemporalActivityCommand
	Command cobra.Command
	SingleWorkflowOrBatchOptions
	ActivityId      string
	ActivityType    string
	ResetAttempts   bool
	ResetHeartbeats bool
	MatchAll        bool
	Jitter          Duration
}

func NewTemporalActivityUnpauseCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityUnpauseCommand {
	var s TemporalActivityUnpauseCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "unpause [flags]"
	s.Command.Short = "Unpause an Activity"
	if hasHighlighting {
		s.Command.Long = "Re-schedule a previously-paused Activity for execution.\n\nIf the Activity is not running and is past its retry timeout, it will be\nscheduled immediately. Otherwise, it will be scheduled after its retry\ntimeout expires.\n\nUse \x1b[1m--reset-attempts\x1b[0m to reset the number of previous run attempts to\nzero. For example, if an Activity is near the maximum number of attempts\nN specified in its retry policy, \x1b[1m--reset-attempts\x1b[0m will allow the\nActivity to be retried another N times after unpausing.\n\nUse \x1b[1m--reset-heartbeat\x1b[0m to reset the Activity's heartbeats.\n\nActivities can be specified by their Activity ID or Activity Type.\nOne of those parameters must be provided.\n\nSpecify the Activity ID or Type and Workflow IDs:\n\n\x1b[1mtemporal activity unpause \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n    --reset-attempts\n    --reset-heartbeats\x1b[0m\n\nActivities can be unpaused in bulk via a visibility Query list filter.\nFor example, if you want to unpause activities of type Foo that you\npreviously paused, do:\n\n\x1b[1mtemporal activity unpause \\\n    --query 'TemporalPauseInfo=\"property:activityType=Foo\"'\x1b[0m"
	} else {
		s.Command.Long = "Re-schedule a previously-paused Activity for execution.\n\nIf the Activity is not running and is past its retry timeout, it will be\nscheduled immediately. Otherwise, it will be scheduled after its retry\ntimeout expires.\n\nUse `--reset-attempts` to reset the number of previous run attempts to\nzero. For example, if an Activity is near the maximum number of attempts\nN specified in its retry policy, `--reset-attempts` will allow the\nActivity to be retried another N times after unpausing.\n\nUse `--reset-heartbeat` to reset the Activity's heartbeats.\n\nActivities can be specified by their Activity ID or Activity Type.\nOne of those parameters must be provided.\n\nSpecify the Activity ID or Type and Workflow IDs:\n\n```\ntemporal activity unpause \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n    --reset-attempts\n    --reset-heartbeats\n```\n\nActivities can be unpaused in bulk via a visibility Query list filter.\nFor example, if you want to unpause activities of type Foo that you\npreviously paused, do:\n\n```\ntemporal activity unpause \\\n    --query 'TemporalPauseInfo=\"property:activityType=Foo\"'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.ActivityId, "activity-id", "a", "", "The Activity ID to unpause.  Mutually exclusive with `--query`, `--match-all`, and `--activity-type`. Requires `--workflow-id` to be specified.")
	s.Command.Flags().StringVar(&s.ActivityType, "activity-type", "", "Activities of this Type will unpause. Can only be used without --match-all. Either `activity-id` or `activity-type` must be provided, but not both.")
	s.Command.Flags().BoolVar(&s.ResetAttempts, "reset-attempts", false, "Reset the activity attempts.")
	s.Command.Flags().BoolVar(&s.ResetHeartbeats, "reset-heartbeats", false, "Reset the Activity's heartbeats. Only works with --reset-attempts.")
	s.Command.Flags().BoolVar(&s.MatchAll, "match-all", false, "Every paused activity should be unpaused. This flag is ignored if activity-type is provided.")
	s.Jitter = 0
	s.Command.Flags().Var(&s.Jitter, "jitter", "The activity will start at random a time within the specified duration. Can only be used with --query.")
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalActivityUpdateOptionsCommand struct {
	Parent  *TemporalActivityCommand
	Command cobra.Command
	SingleWorkflowOrBatchOptions
	ActivityId              string
	ActivityType            string
	MatchAll                bool
	TaskQueue               string
	ScheduleToCloseTimeout  Duration
	ScheduleToStartTimeout  Duration
	StartToCloseTimeout     Duration
	HeartbeatTimeout        Duration
	RetryInitialInterval    Duration
	RetryMaximumInterval    Duration
	RetryBackoffCoefficient float32
	RetryMaximumAttempts    int
	RestoreOriginalOptions  bool
}

func NewTemporalActivityUpdateOptionsCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityUpdateOptionsCommand {
	var s TemporalActivityUpdateOptionsCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update-options [flags]"
	s.Command.Short = "Update Activity options"
	if hasHighlighting {
		s.Command.Long = "Update the options of a running Activity that were passed into it from\na Workflow. Updates are incremental, only changing the specified options.\n\nFor example:\n\n\x1b[1mtemporal activity update-options \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --task-queue NewTaskQueueName \\\n    --schedule-to-close-timeout DURATION \\\n    --schedule-to-start-timeout DURATION \\\n    --start-to-close-timeout DURATION \\\n    --heartbeat-timeout DURATION \\\n    --retry-initial-interval DURATION \\\n    --retry-maximum-interval DURATION \\\n    --retry-backoff-coefficient NewBackoffCoefficient \\\n    --retry-maximum-attempts NewMaximumAttempts\x1b[0m\n\nYou may follow this command with \x1b[1mtemporal activity reset\x1b[0m, and the new values will apply after the reset.\n\nEither \x1b[1mactivity-id\x1b[0m, \x1b[1mactivity-type\x1b[0m, or \x1b[1m--match-all\x1b[0m must be specified.\n\nActivity options can be updated in bulk with a visibility query list filter. \nFor example, if you want to reset for activities of type Foo, do:\n\n\x1b[1mtemporal activity update-options \\\n    --query 'TemporalPauseInfo=\"property:activityType=Foo\"'\n    ...\x1b[0m"
	} else {
		s.Command.Long = "Update the options of a running Activity that were passed into it from\na Workflow. Updates are incremental, only changing the specified options.\n\nFor example:\n\n```\ntemporal activity update-options \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --task-queue NewTaskQueueName \\\n    --schedule-to-close-timeout DURATION \\\n    --schedule-to-start-timeout DURATION \\\n    --start-to-close-timeout DURATION \\\n    --heartbeat-timeout DURATION \\\n    --retry-initial-interval DURATION \\\n    --retry-maximum-interval DURATION \\\n    --retry-backoff-coefficient NewBackoffCoefficient \\\n    --retry-maximum-attempts NewMaximumAttempts\n```\n\nYou may follow this command with `temporal activity reset`, and the new values will apply after the reset.\n\nEither `activity-id`, `activity-type`, or `--match-all` must be specified.\n\nActivity options can be updated in bulk with a visibility query list filter. \nFor example, if you want to reset for activities of type Foo, do:\n\n```\ntemporal activity update-options \\\n    --query 'TemporalPauseInfo=\"property:activityType=Foo\"'\n    ...\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.ActivityId, "activity-id", "a", "", "The Activity ID to update options. Mutually exclusive with `--query`, `--match-all`, and `--activity-type`. Requires `--workflow-id` to be specified.")
	s.Command.Flags().StringVar(&s.ActivityType, "activity-type", "", "Activities of this Type will be updated. Mutually exclusive with `--match-all` and `activity-id`.")
	s.Command.Flags().BoolVar(&s.MatchAll, "match-all", false, "Every activity should be updated. Mutually exclusive with `--activity-id` and `--activity-type`.")
	s.Command.Flags().StringVar(&s.TaskQueue, "task-queue", "", "Name of the task queue for the Activity.")
	s.ScheduleToCloseTimeout = 0
	s.Command.Flags().Var(&s.ScheduleToCloseTimeout, "schedule-to-close-timeout", "Indicates how long the caller is willing to wait for an activity completion. Limits how long retries will be attempted.")
	s.ScheduleToStartTimeout = 0
	s.Command.Flags().Var(&s.ScheduleToStartTimeout, "schedule-to-start-timeout", "Limits time an activity task can stay in a task queue before a worker picks it up. This timeout is always non retryable, as all a retry would achieve is to put it back into the same queue. Defaults to the schedule-to-close timeout or workflow execution timeout if not specified.")
	s.StartToCloseTimeout = 0
	s.Command.Flags().Var(&s.StartToCloseTimeout, "start-to-close-timeout", "Maximum time an activity is allowed to execute after being picked up by a worker. This timeout is always retryable.")
	s.HeartbeatTimeout = 0
	s.Command.Flags().Var(&s.HeartbeatTimeout, "heartbeat-timeout", "Maximum permitted time between successful worker heartbeats.")
	s.RetryInitialInterval = 0
	s.Command.Flags().Var(&s.RetryInitialInterval, "retry-initial-interval", "Interval of the first retry. If retryBackoffCoefficient is 1.0 then it is used for all retries.")
	s.RetryMaximumInterval = 0
	s.Command.Flags().Var(&s.RetryMaximumInterval, "retry-maximum-interval", "Maximum interval between retries. Exponential backoff leads to interval increase. This value is the cap of the increase.")
	s.Command.Flags().Float32Var(&s.RetryBackoffCoefficient, "retry-backoff-coefficient", 0, "Coefficient used to calculate the next retry interval. The next retry interval is previous interval multiplied by the backoff coefficient. Must be 1 or larger.")
	s.Command.Flags().IntVar(&s.RetryMaximumAttempts, "retry-maximum-attempts", 0, "Maximum number of attempts. When exceeded the retries stop even if not expired yet. Setting this value to 1 disables retries. Setting this value to 0 means unlimited attempts(up to the timeouts).")
	s.Command.Flags().BoolVar(&s.RestoreOriginalOptions, "restore-original-options", false, "Restore the original options of the activity.")
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalBatchCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
	ClientOptions
}

func NewTemporalBatchCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalBatchCommand {
	var s TemporalBatchCommand
	s.Parent = parent
	s.Command.Use = "batch"
	s.Command.Short = "Manage running batch jobs"
	if hasHighlighting {
		s.Command.Long = "List or terminate running batch jobs.\n\nA batch job executes a command on multiple Workflow Executions at once. Create\nbatch jobs by passing \x1b[1m--query\x1b[0m to commands that support it. For example, to\ncreate a batch job to cancel a set of Workflow Executions:\n\n\x1b[1mtemporal workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\x1b[0m\n\nQuery Quick Reference:\n\n\x1b[1m+----------------------------------------------------------------------------+\n| Composition:                                                               |\n| - Data types: String literals with single or double quotes,                |\n|   Numbers (integer and floating point), Booleans                           |\n| - Comparison: '=', '!=', '>', '>=', '<', '<='                              |\n| - Expressions/Operators:  'IN array', 'BETWEEN value AND value',           |\n|   'STARTS_WITH string', 'IS NULL', 'IS NOT NULL', 'expr AND expr',         |\n|   'expr OR expr', '( expr )'                                               |\n| - Array: '( comma-separated-values )'                                      |\n|                                                                            |\n| Please note:                                                               |\n| - Wrap attributes with backticks if it contains characters not in          |\n|   [a-zA-Z0-9].                                                             |\n| - STARTS_WITH is only available for Keyword search attributes.             |\n+----------------------------------------------------------------------------+\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation."
	} else {
		s.Command.Long = "List or terminate running batch jobs.\n\nA batch job executes a command on multiple Workflow Executions at once. Create\nbatch jobs by passing `--query` to commands that support it. For example, to\ncreate a batch job to cancel a set of Workflow Executions:\n\n```\ntemporal workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\n```\n\nQuery Quick Reference:\n\n```\n+----------------------------------------------------------------------------+\n| Composition:                                                               |\n| - Data types: String literals with single or double quotes,                |\n|   Numbers (integer and floating point), Booleans                           |\n| - Comparison: '=', '!=', '>', '>=', '<', '<='                              |\n| - Expressions/Operators:  'IN array', 'BETWEEN value AND value',           |\n|   'STARTS_WITH string', 'IS NULL', 'IS NOT NULL', 'expr AND expr',         |\n|   'expr OR expr', '( expr )'                                               |\n| - Array: '( comma-separated-values )'                                      |\n|                                                                            |\n| Please note:                                                               |\n| - Wrap attributes with backticks if it contains characters not in          |\n|   [a-zA-Z0-9].                                                             |\n| - STARTS_WITH is only available for Keyword search attributes.             |\n+----------------------------------------------------------------------------+\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalBatchDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalBatchListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalBatchTerminateCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalBatchDescribeCommand struct {
	Parent  *TemporalBatchCommand
	Command cobra.Command
	JobId   string
}

func NewTemporalBatchDescribeCommand(cctx *CommandContext, parent *TemporalBatchCommand) *TemporalBatchDescribeCommand {
	var s TemporalBatchDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Show batch job progress"
	if hasHighlighting {
		s.Command.Long = "Show the progress of an ongoing batch job. Pass a valid job ID to display its\ninformation:\n\n\x1b[1mtemporal batch describe \\\n    --job-id YourJobId\x1b[0m"
	} else {
		s.Command.Long = "Show the progress of an ongoing batch job. Pass a valid job ID to display its\ninformation:\n\n```\ntemporal batch describe \\\n    --job-id YourJobId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.JobId, "job-id", "", "Batch job ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "job-id")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalBatchListCommand struct {
	Parent  *TemporalBatchCommand
	Command cobra.Command
	Limit   int
}

func NewTemporalBatchListCommand(cctx *CommandContext, parent *TemporalBatchCommand) *TemporalBatchListCommand {
	var s TemporalBatchListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List all batch jobs"
	if hasHighlighting {
		s.Command.Long = "Return a list of batch jobs on the Service or within a single Namespace. For\nexample, list the batch jobs for \"YourNamespace\":\n\n\x1b[1mtemporal batch list \\\n    --namespace YourNamespace\x1b[0m"
	} else {
		s.Command.Long = "Return a list of batch jobs on the Service or within a single Namespace. For\nexample, list the batch jobs for \"YourNamespace\":\n\n```\ntemporal batch list \\\n    --namespace YourNamespace\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Maximum number of batch jobs to display.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalBatchTerminateCommand struct {
	Parent  *TemporalBatchCommand
	Command cobra.Command
	JobId   string
	Reason  string
}

func NewTemporalBatchTerminateCommand(cctx *CommandContext, parent *TemporalBatchCommand) *TemporalBatchTerminateCommand {
	var s TemporalBatchTerminateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "terminate [flags]"
	s.Command.Short = "Forcefully end a batch job"
	if hasHighlighting {
		s.Command.Long = "Terminate a batch job with the provided job ID. You must provide a reason for\nthe termination. The Service stores this explanation as metadata for the\ntermination event for later reference:\n\n\x1b[1mtemporal batch terminate \\\n    --job-id YourJobId \\\n    --reason YourTerminationReason\x1b[0m"
	} else {
		s.Command.Long = "Terminate a batch job with the provided job ID. You must provide a reason for\nthe termination. The Service stores this explanation as metadata for the\ntermination event for later reference:\n\n```\ntemporal batch terminate \\\n    --job-id YourJobId \\\n    --reason YourTerminationReason\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.JobId, "job-id", "", "Job ID to terminate. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "job-id")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for terminating the batch job. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "reason")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalConfigCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
}

func NewTemporalConfigCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalConfigCommand {
	var s TemporalConfigCommand
	s.Parent = parent
	s.Command.Use = "config"
	s.Command.Short = "Manage config files (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Config files are TOML files that contain profiles, with each profile\ncontaining configuration for connecting to Temporal.\n\n\x1b[1mtemporal config set \\\n    --prop address \\\n    --value us-west-2.aws.api.temporal.io:7233\x1b[0m\n\nThe default config file path is \x1b[1m$CONFIG_PATH/temporalio/temporal.toml\x1b[0m where\n\x1b[1m$CONFIG_PATH\x1b[0m is defined as \x1b[1m$HOME/.config\x1b[0m on Unix,\n\x1b[1m$HOME/Library/Application Support\x1b[0m on macOS, and \x1b[1m%AppData%\x1b[0m on Windows.\nThis can be overridden with the \x1b[1mTEMPORAL_CONFIG_FILE\x1b[0m environment\nvariable or \x1b[1m--config-file\x1b[0m.\n\nThe default profile is \x1b[1mdefault\x1b[0m. This can be overridden with the\n\x1b[1mTEMPORAL_PROFILE\x1b[0m environment variable or \x1b[1m--profile\x1b[0m."
	} else {
		s.Command.Long = "Config files are TOML files that contain profiles, with each profile\ncontaining configuration for connecting to Temporal.\n\n```\ntemporal config set \\\n    --prop address \\\n    --value us-west-2.aws.api.temporal.io:7233\n```\n\nThe default config file path is `$CONFIG_PATH/temporalio/temporal.toml` where\n`$CONFIG_PATH` is defined as `$HOME/.config` on Unix,\n`$HOME/Library/Application Support` on macOS, and `%AppData%` on Windows.\nThis can be overridden with the `TEMPORAL_CONFIG_FILE` environment\nvariable or `--config-file`.\n\nThe default profile is `default`. This can be overridden with the\n`TEMPORAL_PROFILE` environment variable or `--profile`."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalConfigDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalConfigDeleteProfileCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalConfigGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalConfigListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalConfigSetCommand(cctx, &s).Command)
	return &s
}

type TemporalConfigDeleteCommand struct {
	Parent  *TemporalConfigCommand
	Command cobra.Command
	Prop    string
}

func NewTemporalConfigDeleteCommand(cctx *CommandContext, parent *TemporalConfigCommand) *TemporalConfigDeleteCommand {
	var s TemporalConfigDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a config file property (EXPERIMENTAL)\n"
	if hasHighlighting {
		s.Command.Long = "Remove a property within a profile.\n\n\x1b[1mtemporal env delete \\\n    --prop tls.client_cert_path\x1b[0m"
	} else {
		s.Command.Long = "Remove a property within a profile.\n\n```\ntemporal env delete \\\n    --prop tls.client_cert_path\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Prop, "prop", "p", "", "Specific property to delete. If unset, deletes entire profile. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "prop")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalConfigDeleteProfileCommand struct {
	Parent  *TemporalConfigCommand
	Command cobra.Command
}

func NewTemporalConfigDeleteProfileCommand(cctx *CommandContext, parent *TemporalConfigCommand) *TemporalConfigDeleteProfileCommand {
	var s TemporalConfigDeleteProfileCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete-profile [flags]"
	s.Command.Short = "Delete an entire config profile (EXPERIMENTAL)\n"
	if hasHighlighting {
		s.Command.Long = "Remove a full profile entirely. The \x1b[1m--profile\x1b[0m must be set explicitly.\n\n\x1b[1mtemporal env delete-profile \\\n    --profile my-profile\x1b[0m"
	} else {
		s.Command.Long = "Remove a full profile entirely. The `--profile` must be set explicitly.\n\n```\ntemporal env delete-profile \\\n    --profile my-profile\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalConfigGetCommand struct {
	Parent  *TemporalConfigCommand
	Command cobra.Command
	Prop    string
}

func NewTemporalConfigGetCommand(cctx *CommandContext, parent *TemporalConfigCommand) *TemporalConfigGetCommand {
	var s TemporalConfigGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Show config file properties (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Display specific properties or the entire profile.\n\n\x1b[1mtemporal config get \\\n    --prop address\x1b[0m\n\nor\n\n\x1b[1mtemporal config get\x1b[0m"
	} else {
		s.Command.Long = "Display specific properties or the entire profile.\n\n```\ntemporal config get \\\n    --prop address\n```\n\nor\n\n```\ntemporal config get\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Prop, "prop", "p", "", "Specific property to get.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalConfigListCommand struct {
	Parent  *TemporalConfigCommand
	Command cobra.Command
}

func NewTemporalConfigListCommand(cctx *CommandContext, parent *TemporalConfigCommand) *TemporalConfigListCommand {
	var s TemporalConfigListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Show config file profiles (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "List profile names in the config file.\n\n\x1b[1mtemporal config list\x1b[0m"
	} else {
		s.Command.Long = "List profile names in the config file.\n\n```\ntemporal config list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalConfigSetCommand struct {
	Parent  *TemporalConfigCommand
	Command cobra.Command
	Prop    string
	Value   string
}

func NewTemporalConfigSetCommand(cctx *CommandContext, parent *TemporalConfigCommand) *TemporalConfigSetCommand {
	var s TemporalConfigSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set [flags]"
	s.Command.Short = "Set config file properties (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Assign a value to a property and store it in the config file:\n\n\x1b[1mtemporal config set \\\n    --prop address \\\n    --value us-west-2.aws.api.temporal.io:7233\x1b[0m"
	} else {
		s.Command.Long = "Assign a value to a property and store it in the config file:\n\n```\ntemporal config set \\\n    --prop address \\\n    --value us-west-2.aws.api.temporal.io:7233\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Prop, "prop", "p", "", "Property name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "prop")
	s.Command.Flags().StringVarP(&s.Value, "value", "v", "", "Property value. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "value")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalEnvCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
}

func NewTemporalEnvCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalEnvCommand {
	var s TemporalEnvCommand
	s.Parent = parent
	s.Command.Use = "env"
	s.Command.Short = "Manage environments"
	if hasHighlighting {
		s.Command.Long = "Environments manage key-value presets, auto-configuring Temporal CLI options\nfor you. You can set up distinct environments like \"dev\" and \"prod\" for\nconvenience:\n\n\x1b[1mtemporal env set \\\n    --env prod \\\n    --key address \\\n    --value production.f45a2.tmprl.cloud:7233\x1b[0m\n\nEach environment is isolated. Changes to \"prod\" presets won't affect \"dev\".\n\nFor easiest use, set a \x1b[1mTEMPORAL_ENV\x1b[0m environment variable in your shell. The\nTemporal CLI checks for an \x1b[1m--env\x1b[0m option first, then checks for the\n\x1b[1mTEMPORAL_ENV\x1b[0m environment variable. If neither is set, the CLI uses the\n\"default\" environment."
	} else {
		s.Command.Long = "Environments manage key-value presets, auto-configuring Temporal CLI options\nfor you. You can set up distinct environments like \"dev\" and \"prod\" for\nconvenience:\n\n```\ntemporal env set \\\n    --env prod \\\n    --key address \\\n    --value production.f45a2.tmprl.cloud:7233\n```\n\nEach environment is isolated. Changes to \"prod\" presets won't affect \"dev\".\n\nFor easiest use, set a `TEMPORAL_ENV` environment variable in your shell. The\nTemporal CLI checks for an `--env` option first, then checks for the\n`TEMPORAL_ENV` environment variable. If neither is set, the CLI uses the\n\"default\" environment."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalEnvDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvSetCommand(cctx, &s).Command)
	return &s
}

type TemporalEnvDeleteCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
	Key     string
}

func NewTemporalEnvDeleteCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvDeleteCommand {
	var s TemporalEnvDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete an environment or environment property"
	if hasHighlighting {
		s.Command.Long = "Remove a presets environment entirely _or_ remove a key-value pair within an\nenvironment. If you don't specify an environment (with \x1b[1m--env\x1b[0m or by setting\nthe \x1b[1mTEMPORAL_ENV\x1b[0m variable), this command updates the \"default\" environment:\n\n\x1b[1mtemporal env delete \\\n    --env YourEnvironment\x1b[0m\n\nor\n\n\x1b[1mtemporal env delete \\\n    --env prod \\\n    --key tls-key-path\x1b[0m"
	} else {
		s.Command.Long = "Remove a presets environment entirely _or_ remove a key-value pair within an\nenvironment. If you don't specify an environment (with `--env` or by setting\nthe `TEMPORAL_ENV` variable), this command updates the \"default\" environment:\n\n```\ntemporal env delete \\\n    --env YourEnvironment\n```\n\nor\n\n```\ntemporal env delete \\\n    --env prod \\\n    --key tls-key-path\n```"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVarP(&s.Key, "key", "k", "", "Property name.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalEnvGetCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
	Key     string
}

func NewTemporalEnvGetCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvGetCommand {
	var s TemporalEnvGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Show environment properties"
	if hasHighlighting {
		s.Command.Long = "List the properties for a given environment:\n\n\x1b[1mtemporal env get \\\n    --env YourEnvironment\x1b[0m\n\nPrint a single property:\n\n\x1b[1mtemporal env get \\\n    --env YourEnvironment \\\n    --key YourPropertyKey\x1b[0m\n\nIf you don't specify an environment (with \x1b[1m--env\x1b[0m or by setting the\n\x1b[1mTEMPORAL_ENV\x1b[0m variable), this command lists properties of the \"default\"\nenvironment."
	} else {
		s.Command.Long = "List the properties for a given environment:\n\n```\ntemporal env get \\\n    --env YourEnvironment\n```\n\nPrint a single property:\n\n```\ntemporal env get \\\n    --env YourEnvironment \\\n    --key YourPropertyKey\n```\n\nIf you don't specify an environment (with `--env` or by setting the\n`TEMPORAL_ENV` variable), this command lists properties of the \"default\"\nenvironment."
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVarP(&s.Key, "key", "k", "", "Property name.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalEnvListCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
}

func NewTemporalEnvListCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvListCommand {
	var s TemporalEnvListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Show environment names"
	s.Command.Long = "List the environments you have set up on your local computer. Environments are\nstored in \"$HOME/.config/temporalio/temporal.yaml\"."
	s.Command.Args = cobra.NoArgs
	s.Command.Annotations = make(map[string]string)
	s.Command.Annotations["ignoresMissingEnv"] = "true"
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalEnvSetCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
	Key     string
	Value   string
}

func NewTemporalEnvSetCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvSetCommand {
	var s TemporalEnvSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set [flags]"
	s.Command.Short = "Set environment properties"
	if hasHighlighting {
		s.Command.Long = "Assign a value to a property key and store it to an environment:\n\n\x1b[1mtemporal env set \\\n    --env environment \\\n    --key property \\\n    --value value\x1b[0m\n\nIf you don't specify an environment (with \x1b[1m--env\x1b[0m or by setting the\n\x1b[1mTEMPORAL_ENV\x1b[0m variable), this command sets properties in the \"default\"\nenvironment.\n\nStoring keys with CLI option names lets the CLI automatically set those\noptions for you. This reduces effort and helps avoid typos when issuing\ncommands."
	} else {
		s.Command.Long = "Assign a value to a property key and store it to an environment:\n\n```\ntemporal env set \\\n    --env environment \\\n    --key property \\\n    --value value\n```\n\nIf you don't specify an environment (with `--env` or by setting the\n`TEMPORAL_ENV` variable), this command sets properties in the \"default\"\nenvironment.\n\nStoring keys with CLI option names lets the CLI automatically set those\noptions for you. This reduces effort and helps avoid typos when issuing\ncommands."
	}
	s.Command.Args = cobra.MaximumNArgs(2)
	s.Command.Annotations = make(map[string]string)
	s.Command.Annotations["ignoresMissingEnv"] = "true"
	s.Command.Flags().StringVarP(&s.Key, "key", "k", "", "Property name (required).")
	s.Command.Flags().StringVarP(&s.Value, "value", "v", "", "Property value (required).")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
	ClientOptions
}

func NewTemporalOperatorCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalOperatorCommand {
	var s TemporalOperatorCommand
	s.Parent = parent
	s.Command.Use = "operator"
	s.Command.Short = "Manage Temporal deployments"
	if hasHighlighting {
		s.Command.Long = "Operator commands manage and fetch information about Namespaces, Search\nAttributes, Nexus Endpoints, and Temporal Services:\n\n\x1b[1mtemporal operator [command] [subcommand] [options]\x1b[0m\n\nFor example, to show information about the Temporal Service at the default\naddress (localhost):\n\n\x1b[1mtemporal operator cluster describe\x1b[0m"
	} else {
		s.Command.Long = "Operator commands manage and fetch information about Namespaces, Search\nAttributes, Nexus Endpoints, and Temporal Services:\n\n```\ntemporal operator [command] [subcommand] [options]\n```\n\nFor example, to show information about the Temporal Service at the default\naddress (localhost):\n\n```\ntemporal operator cluster describe\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalOperatorClusterCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNamespaceCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNexusCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorSearchAttributeCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalOperatorClusterCommand struct {
	Parent  *TemporalOperatorCommand
	Command cobra.Command
}

func NewTemporalOperatorClusterCommand(cctx *CommandContext, parent *TemporalOperatorCommand) *TemporalOperatorClusterCommand {
	var s TemporalOperatorClusterCommand
	s.Parent = parent
	s.Command.Use = "cluster"
	s.Command.Short = "Manage a Temporal Cluster"
	if hasHighlighting {
		s.Command.Long = "Perform operator actions on Temporal Services (also known as Clusters).\n\n\x1b[1mtemporal operator cluster [subcommand] [options]\x1b[0m\n\nFor example to check Service/Cluster health:\n\n\x1b[1mtemporal operator cluster health\x1b[0m"
	} else {
		s.Command.Long = "Perform operator actions on Temporal Services (also known as Clusters).\n\n```\ntemporal operator cluster [subcommand] [options]\n```\n\nFor example to check Service/Cluster health:\n\n```\ntemporal operator cluster health\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalOperatorClusterDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorClusterHealthCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorClusterListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorClusterRemoveCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorClusterSystemCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorClusterUpsertCommand(cctx, &s).Command)
	return &s
}

type TemporalOperatorClusterDescribeCommand struct {
	Parent  *TemporalOperatorClusterCommand
	Command cobra.Command
	Detail  bool
}

func NewTemporalOperatorClusterDescribeCommand(cctx *CommandContext, parent *TemporalOperatorClusterCommand) *TemporalOperatorClusterDescribeCommand {
	var s TemporalOperatorClusterDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Show Temporal Cluster information"
	if hasHighlighting {
		s.Command.Long = "View information about a Temporal Cluster (Service), including Cluster Name,\npersistence store, and visibility store. Add \x1b[1m--detail\x1b[0m for additional info:\n\n\x1b[1mtemporal operator cluster describe [--detail]\x1b[0m"
	} else {
		s.Command.Long = "View information about a Temporal Cluster (Service), including Cluster Name,\npersistence store, and visibility store. Add `--detail` for additional info:\n\n```\ntemporal operator cluster describe [--detail]\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.Detail, "detail", false, "Show history shard count and Cluster/Service version information.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorClusterHealthCommand struct {
	Parent  *TemporalOperatorClusterCommand
	Command cobra.Command
}

func NewTemporalOperatorClusterHealthCommand(cctx *CommandContext, parent *TemporalOperatorClusterCommand) *TemporalOperatorClusterHealthCommand {
	var s TemporalOperatorClusterHealthCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "health [flags]"
	s.Command.Short = "Check Temporal Service health"
	if hasHighlighting {
		s.Command.Long = "View information about the health of a Temporal Service:\n\n\x1b[1mtemporal operator cluster health\x1b[0m"
	} else {
		s.Command.Long = "View information about the health of a Temporal Service:\n\n```\ntemporal operator cluster health\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorClusterListCommand struct {
	Parent  *TemporalOperatorClusterCommand
	Command cobra.Command
	Limit   int
}

func NewTemporalOperatorClusterListCommand(cctx *CommandContext, parent *TemporalOperatorClusterCommand) *TemporalOperatorClusterListCommand {
	var s TemporalOperatorClusterListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Show Temporal Clusters"
	if hasHighlighting {
		s.Command.Long = "Print a list of remote Temporal Clusters (Services) registered to the local\nService. Report details include the Cluster's name, ID, address, History Shard\ncount, Failover version, and availability:\n\n\x1b[1mtemporal operator cluster list [--limit max-count]\x1b[0m"
	} else {
		s.Command.Long = "Print a list of remote Temporal Clusters (Services) registered to the local\nService. Report details include the Cluster's name, ID, address, History Shard\ncount, Failover version, and availability:\n\n```\ntemporal operator cluster list [--limit max-count]\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Maximum number of Clusters to display.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorClusterRemoveCommand struct {
	Parent  *TemporalOperatorClusterCommand
	Command cobra.Command
	Name    string
}

func NewTemporalOperatorClusterRemoveCommand(cctx *CommandContext, parent *TemporalOperatorClusterCommand) *TemporalOperatorClusterRemoveCommand {
	var s TemporalOperatorClusterRemoveCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "remove [flags]"
	s.Command.Short = "Remove a Temporal Cluster"
	if hasHighlighting {
		s.Command.Long = "Remove a registered remote Temporal Cluster (Service) from the local Service.\n\n\x1b[1mtemporal operator cluster remove \\\n    --name YourClusterName\x1b[0m"
	} else {
		s.Command.Long = "Remove a registered remote Temporal Cluster (Service) from the local Service.\n\n```\ntemporal operator cluster remove \\\n    --name YourClusterName\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Cluster/Service name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorClusterSystemCommand struct {
	Parent  *TemporalOperatorClusterCommand
	Command cobra.Command
}

func NewTemporalOperatorClusterSystemCommand(cctx *CommandContext, parent *TemporalOperatorClusterCommand) *TemporalOperatorClusterSystemCommand {
	var s TemporalOperatorClusterSystemCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "system [flags]"
	s.Command.Short = "Show Temporal Cluster info"
	if hasHighlighting {
		s.Command.Long = "Show Temporal Server information for Temporal Clusters (Service): Server\nversion, scheduling support, and more. This information helps diagnose\nproblems with the Temporal Server.\n\nThe command defaults to the local Service. Otherwise, use the\n\x1b[1m--frontend-address\x1b[0m option to specify a Cluster (Service) endpoint:\n\n\x1b[1mtemporal operator cluster system \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\x1b[0m"
	} else {
		s.Command.Long = "Show Temporal Server information for Temporal Clusters (Service): Server\nversion, scheduling support, and more. This information helps diagnose\nproblems with the Temporal Server.\n\nThe command defaults to the local Service. Otherwise, use the\n`--frontend-address` option to specify a Cluster (Service) endpoint:\n\n```\ntemporal operator cluster system \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorClusterUpsertCommand struct {
	Parent           *TemporalOperatorClusterCommand
	Command          cobra.Command
	FrontendAddress  string
	EnableConnection bool
}

func NewTemporalOperatorClusterUpsertCommand(cctx *CommandContext, parent *TemporalOperatorClusterCommand) *TemporalOperatorClusterUpsertCommand {
	var s TemporalOperatorClusterUpsertCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "upsert [flags]"
	s.Command.Short = "Add/update a Temporal Cluster"
	if hasHighlighting {
		s.Command.Long = "Add, remove, or update a registered (\"remote\") Temporal Cluster (Service).\n\n\x1b[1mtemporal operator cluster upsert [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator cluster upsert \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\n    --enable-connection false\x1b[0m"
	} else {
		s.Command.Long = "Add, remove, or update a registered (\"remote\") Temporal Cluster (Service).\n\n```\ntemporal operator cluster upsert [options]\n```\n\nFor example:\n\n```\ntemporal operator cluster upsert \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\n    --enable-connection false\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.FrontendAddress, "frontend-address", "", "Remote endpoint. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "frontend-address")
	s.Command.Flags().BoolVar(&s.EnableConnection, "enable-connection", false, "Set the connection to \"enabled\".")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNamespaceCommand struct {
	Parent  *TemporalOperatorCommand
	Command cobra.Command
}

func NewTemporalOperatorNamespaceCommand(cctx *CommandContext, parent *TemporalOperatorCommand) *TemporalOperatorNamespaceCommand {
	var s TemporalOperatorNamespaceCommand
	s.Parent = parent
	s.Command.Use = "namespace"
	s.Command.Short = "Namespace operations"
	if hasHighlighting {
		s.Command.Long = "Manage Temporal Cluster (Service) Namespaces:\n\n\x1b[1mtemporal operator namespace [command] [command options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator namespace create \\\n    --namespace YourNewNamespaceName\x1b[0m"
	} else {
		s.Command.Long = "Manage Temporal Cluster (Service) Namespaces:\n\n```\ntemporal operator namespace [command] [command options]\n```\n\nFor example:\n\n```\ntemporal operator namespace create \\\n    --namespace YourNewNamespaceName\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalOperatorNamespaceCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNamespaceDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNamespaceDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNamespaceListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNamespaceUpdateCommand(cctx, &s).Command)
	return &s
}

type TemporalOperatorNamespaceCreateCommand struct {
	Parent                  *TemporalOperatorNamespaceCommand
	Command                 cobra.Command
	ActiveCluster           string
	Cluster                 []string
	Data                    []string
	Description             string
	Email                   string
	Global                  bool
	HistoryArchivalState    StringEnum
	HistoryUri              string
	Retention               Duration
	VisibilityArchivalState StringEnum
	VisibilityUri           string
}

func NewTemporalOperatorNamespaceCreateCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceCreateCommand {
	var s TemporalOperatorNamespaceCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Register a new Namespace"
	if hasHighlighting {
		s.Command.Long = "Create a new Namespace on the Temporal Service:\n\n\x1b[1mtemporal operator namespace create \\\n    --namespace YourNewNamespaceName \\\n    [options]\x1b[0m\n\nCreate a Namespace with multi-region data replication:\n\n\x1b[1mtemporal operator namespace create \\\n    --global \\\n    --namespace YourNewNamespaceName\x1b[0m\n\nConfigure settings like retention and Visibility Archival State as needed.\nFor example, the Visibility Archive can be set on a separate URI:\n\n\x1b[1mtemporal operator namespace create \\\n    --retention 5d \\\n    --visibility-archival-state enabled \\\n    --visibility-uri YourURI \\\n    --namespace YourNewNamespaceName\x1b[0m\n\nNote: URI values for archival states can't be changed once enabled."
	} else {
		s.Command.Long = "Create a new Namespace on the Temporal Service:\n\n```\ntemporal operator namespace create \\\n    --namespace YourNewNamespaceName \\\n    [options]\n```\n\nCreate a Namespace with multi-region data replication:\n\n```\ntemporal operator namespace create \\\n    --global \\\n    --namespace YourNewNamespaceName\n```\n\nConfigure settings like retention and Visibility Archival State as needed.\nFor example, the Visibility Archive can be set on a separate URI:\n\n```\ntemporal operator namespace create \\\n    --retention 5d \\\n    --visibility-archival-state enabled \\\n    --visibility-uri YourURI \\\n    --namespace YourNewNamespaceName\n```\n\nNote: URI values for archival states can't be changed once enabled."
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active Cluster (Service) name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster (Service) names for Namespace creation. Can be passed multiple times.")
	s.Command.Flags().StringArrayVar(&s.Data, "data", nil, "Namespace data as `KEY=VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: `YourKey={\"your\": \"value\"}` Can be passed multiple times.")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.Global, "global", false, "Enable multi-region data replication.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disabled")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Archive history to this `URI`. Once enabled, can't be changed.")
	s.Retention = Duration(259200000 * time.Millisecond)
	s.Command.Flags().Var(&s.Retention, "retention", "Time to preserve closed Workflows before deletion.")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disabled")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Archive visibility data to this `URI`. Once enabled, can't be changed.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNamespaceDeleteCommand struct {
	Parent  *TemporalOperatorNamespaceCommand
	Command cobra.Command
	Yes     bool
}

func NewTemporalOperatorNamespaceDeleteCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceDeleteCommand {
	var s TemporalOperatorNamespaceDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a Namespace"
	if hasHighlighting {
		s.Command.Long = "Removes a Namespace from the Service.\n\n\x1b[1mtemporal operator namespace delete [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator namespace delete \\\n    --namespace YourNamespaceName\x1b[0m"
	} else {
		s.Command.Long = "Removes a Namespace from the Service.\n\n```\ntemporal operator namespace delete [options]\n```\n\nFor example:\n\n```\ntemporal operator namespace delete \\\n    --namespace YourNamespaceName\n```"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Request confirmation before deletion.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNamespaceDescribeCommand struct {
	Parent      *TemporalOperatorNamespaceCommand
	Command     cobra.Command
	NamespaceId string
}

func NewTemporalOperatorNamespaceDescribeCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceDescribeCommand {
	var s TemporalOperatorNamespaceDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Describe a Namespace"
	if hasHighlighting {
		s.Command.Long = "Provide long-form information about a Namespace identified by its ID or name:\n\n\x1b[1mtemporal operator namespace describe \\\n    --namespace-id YourNamespaceId\x1b[0m\n\nor\n\n\x1b[1mtemporal operator namespace describe \\\n    --namespace YourNamespaceName\x1b[0m"
	} else {
		s.Command.Long = "Provide long-form information about a Namespace identified by its ID or name:\n\n```\ntemporal operator namespace describe \\\n    --namespace-id YourNamespaceId\n```\n\nor\n\n```\ntemporal operator namespace describe \\\n    --namespace YourNamespaceName\n```"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.NamespaceId, "namespace-id", "", "Namespace ID.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNamespaceListCommand struct {
	Parent  *TemporalOperatorNamespaceCommand
	Command cobra.Command
}

func NewTemporalOperatorNamespaceListCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceListCommand {
	var s TemporalOperatorNamespaceListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List Namespaces"
	if hasHighlighting {
		s.Command.Long = "Display a detailed listing for all Namespaces on the Service:\n\n\x1b[1mtemporal operator namespace list\x1b[0m"
	} else {
		s.Command.Long = "Display a detailed listing for all Namespaces on the Service:\n\n```\ntemporal operator namespace list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNamespaceUpdateCommand struct {
	Parent                  *TemporalOperatorNamespaceCommand
	Command                 cobra.Command
	ActiveCluster           string
	Cluster                 []string
	Data                    []string
	Description             string
	Email                   string
	PromoteGlobal           bool
	HistoryArchivalState    StringEnum
	HistoryUri              string
	ReplicationState        StringEnum
	Retention               Duration
	VisibilityArchivalState StringEnum
	VisibilityUri           string
}

func NewTemporalOperatorNamespaceUpdateCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceUpdateCommand {
	var s TemporalOperatorNamespaceUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update a Namespace"
	if hasHighlighting {
		s.Command.Long = "Update a Namespace using properties you specify.\n\n\x1b[1mtemporal operator namespace update [options]\x1b[0m\n\nAssign a Namespace's active Cluster (Service):\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --active-cluster NewActiveCluster\x1b[0m\n\nPromote a Namespace for multi-region data replication:\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --promote-global\x1b[0m\n\nYou may update archives that were previously enabled or disabled. Note: URI\nvalues for archival states can't be changed once enabled.\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --history-archival-state enabled \\\n    --visibility-archival-state disabled\x1b[0m"
	} else {
		s.Command.Long = "Update a Namespace using properties you specify.\n\n```\ntemporal operator namespace update [options]\n```\n\nAssign a Namespace's active Cluster (Service):\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --active-cluster NewActiveCluster\n```\n\nPromote a Namespace for multi-region data replication:\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --promote-global\n```\n\nYou may update archives that were previously enabled or disabled. Note: URI\nvalues for archival states can't be changed once enabled.\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --history-archival-state enabled \\\n    --visibility-archival-state disabled\n```"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active Cluster (Service) name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster (Service) names.")
	s.Command.Flags().StringArrayVar(&s.Data, "data", nil, "Namespace data as `KEY=VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: `YourKey={\"your\": \"value\"}` Can be passed multiple times.")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.PromoteGlobal, "promote-global", false, "Enable multi-region data replication.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Archive history to this `URI`. Once enabled, can't be changed.")
	s.ReplicationState = NewStringEnum([]string{"normal", "handover"}, "")
	s.Command.Flags().Var(&s.ReplicationState, "replication-state", "Replication state. Accepted values: normal, handover.")
	s.Retention = 0
	s.Command.Flags().Var(&s.Retention, "retention", "Length of time a closed Workflow is preserved before deletion.")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Archive visibility data to this `URI`. Once enabled, can't be changed.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNexusCommand struct {
	Parent  *TemporalOperatorCommand
	Command cobra.Command
}

func NewTemporalOperatorNexusCommand(cctx *CommandContext, parent *TemporalOperatorCommand) *TemporalOperatorNexusCommand {
	var s TemporalOperatorNexusCommand
	s.Parent = parent
	s.Command.Use = "nexus"
	s.Command.Short = "Commands for managing Nexus resources"
	if hasHighlighting {
		s.Command.Long = "These commands manage Nexus resources.\n\nNexus commands follow this syntax:\n\n\x1b[1mtemporal operator nexus [command] [subcommand] [options]\x1b[0m"
	} else {
		s.Command.Long = "These commands manage Nexus resources.\n\nNexus commands follow this syntax:\n\n```\ntemporal operator nexus [command] [subcommand] [options]\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalOperatorNexusEndpointCommand(cctx, &s).Command)
	return &s
}

type TemporalOperatorNexusEndpointCommand struct {
	Parent  *TemporalOperatorNexusCommand
	Command cobra.Command
}

func NewTemporalOperatorNexusEndpointCommand(cctx *CommandContext, parent *TemporalOperatorNexusCommand) *TemporalOperatorNexusEndpointCommand {
	var s TemporalOperatorNexusEndpointCommand
	s.Parent = parent
	s.Command.Use = "endpoint"
	s.Command.Short = "Commands for managing Nexus Endpoints"
	if hasHighlighting {
		s.Command.Long = "These commands manage Nexus Endpoints.\n\nNexus Endpoint commands follow this syntax:\n\n\x1b[1mtemporal operator nexus endpoint [command] [options]\x1b[0m"
	} else {
		s.Command.Long = "These commands manage Nexus Endpoints.\n\nNexus Endpoint commands follow this syntax:\n\n```\ntemporal operator nexus endpoint [command] [options]\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalOperatorNexusEndpointCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNexusEndpointDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNexusEndpointGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNexusEndpointListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNexusEndpointUpdateCommand(cctx, &s).Command)
	return &s
}

type TemporalOperatorNexusEndpointCreateCommand struct {
	Parent  *TemporalOperatorNexusEndpointCommand
	Command cobra.Command
	NexusEndpointIdentityOptions
	NexusEndpointConfigOptions
}

func NewTemporalOperatorNexusEndpointCreateCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointCreateCommand {
	var s TemporalOperatorNexusEndpointCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a Nexus Endpoint"
	if hasHighlighting {
		s.Command.Long = "Create a Nexus Endpoint on the Server.\n\nA Nexus Endpoint name is used in Workflow code to invoke Nexus Operations.\nThe endpoint target may either be a Worker, in which case\n\x1b[1m--target-namespace\x1b[0m and \x1b[1m--target-task-queue\x1b[0m must both be provided, or\nan external URL, in which case \x1b[1m--target-url\x1b[0m must be provided.\n\nThis command will fail if an Endpoint with the same name is already\nregistered.\n\n\x1b[1mtemporal operator nexus endpoint create \\\n  --name your-endpoint \\\n  --target-namespace your-namespace \\\n  --target-task-queue your-task-queue \\\n  --description-file DESCRIPTION.md\x1b[0m"
	} else {
		s.Command.Long = "Create a Nexus Endpoint on the Server.\n\nA Nexus Endpoint name is used in Workflow code to invoke Nexus Operations.\nThe endpoint target may either be a Worker, in which case\n`--target-namespace` and `--target-task-queue` must both be provided, or\nan external URL, in which case `--target-url` must be provided.\n\nThis command will fail if an Endpoint with the same name is already\nregistered.\n\n```\ntemporal operator nexus endpoint create \\\n  --name your-endpoint \\\n  --target-namespace your-namespace \\\n  --target-task-queue your-task-queue \\\n  --description-file DESCRIPTION.md\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.NexusEndpointIdentityOptions.buildFlags(cctx, s.Command.Flags())
	s.NexusEndpointConfigOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNexusEndpointDeleteCommand struct {
	Parent  *TemporalOperatorNexusEndpointCommand
	Command cobra.Command
	NexusEndpointIdentityOptions
}

func NewTemporalOperatorNexusEndpointDeleteCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointDeleteCommand {
	var s TemporalOperatorNexusEndpointDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a Nexus Endpoint"
	if hasHighlighting {
		s.Command.Long = "Delete a Nexus Endpoint from the Server.\n\n\x1b[1mtemporal operator nexus endpoint delete --name your-endpoint\x1b[0m"
	} else {
		s.Command.Long = "Delete a Nexus Endpoint from the Server.\n\n```\ntemporal operator nexus endpoint delete --name your-endpoint\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.NexusEndpointIdentityOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNexusEndpointGetCommand struct {
	Parent  *TemporalOperatorNexusEndpointCommand
	Command cobra.Command
	NexusEndpointIdentityOptions
}

func NewTemporalOperatorNexusEndpointGetCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointGetCommand {
	var s TemporalOperatorNexusEndpointGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get a Nexus Endpoint by name (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Get a Nexus Endpoint by name from the Server.\n\n\x1b[1mtemporal operator nexus endpoint get --name your-endpoint\x1b[0m"
	} else {
		s.Command.Long = "Get a Nexus Endpoint by name from the Server.\n\n```\ntemporal operator nexus endpoint get --name your-endpoint\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.NexusEndpointIdentityOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNexusEndpointListCommand struct {
	Parent  *TemporalOperatorNexusEndpointCommand
	Command cobra.Command
}

func NewTemporalOperatorNexusEndpointListCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointListCommand {
	var s TemporalOperatorNexusEndpointListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List Nexus Endpoints"
	if hasHighlighting {
		s.Command.Long = "List all Nexus Endpoints on the Server.\n\n\x1b[1mtemporal operator nexus endpoint list\x1b[0m"
	} else {
		s.Command.Long = "List all Nexus Endpoints on the Server.\n\n```\ntemporal operator nexus endpoint list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorNexusEndpointUpdateCommand struct {
	Parent  *TemporalOperatorNexusEndpointCommand
	Command cobra.Command
	NexusEndpointIdentityOptions
	NexusEndpointConfigOptions
	UnsetDescription bool
}

func NewTemporalOperatorNexusEndpointUpdateCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointUpdateCommand {
	var s TemporalOperatorNexusEndpointUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update an existing Nexus Endpoint"
	if hasHighlighting {
		s.Command.Long = "Update an existing Nexus Endpoint on the Server.\n\nA Nexus Endpoint name is used in Workflow code to invoke Nexus Operations.\nThe Endpoint target may either be a Worker, in which case\n\x1b[1m--target-namespace\x1b[0m and \x1b[1m--target-task-queue\x1b[0m must both be provided, or\nan external URL, in which case \x1b[1m--target-url\x1b[0m must be provided.\n\nThe Endpoint is patched; existing fields for which flags are not provided\nare left as they were.\n\nUpdate only the target task queue:\n\n\x1b[1mtemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --target-task-queue your-other-queue\x1b[0m\n\nUpdate only the description:\n\n\x1b[1mtemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --description-file DESCRIPTION.md\x1b[0m"
	} else {
		s.Command.Long = "Update an existing Nexus Endpoint on the Server.\n\nA Nexus Endpoint name is used in Workflow code to invoke Nexus Operations.\nThe Endpoint target may either be a Worker, in which case\n`--target-namespace` and `--target-task-queue` must both be provided, or\nan external URL, in which case `--target-url` must be provided.\n\nThe Endpoint is patched; existing fields for which flags are not provided\nare left as they were.\n\nUpdate only the target task queue:\n\n```\ntemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --target-task-queue your-other-queue\n```\n\nUpdate only the description:\n\n```\ntemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --description-file DESCRIPTION.md\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.UnsetDescription, "unset-description", false, "Unset the description.")
	s.NexusEndpointIdentityOptions.buildFlags(cctx, s.Command.Flags())
	s.NexusEndpointConfigOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorSearchAttributeCommand struct {
	Parent  *TemporalOperatorCommand
	Command cobra.Command
}

func NewTemporalOperatorSearchAttributeCommand(cctx *CommandContext, parent *TemporalOperatorCommand) *TemporalOperatorSearchAttributeCommand {
	var s TemporalOperatorSearchAttributeCommand
	s.Parent = parent
	s.Command.Use = "search-attribute"
	s.Command.Short = "Search Attribute operations"
	if hasHighlighting {
		s.Command.Long = "Create, list, or remove Search Attributes fields stored in a Workflow\nExecution's metadata:\n\n\x1b[1mtemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\x1b[0m\n\nSupported types include: Text, Keyword, Int, Double, Bool, Datetime, and\nKeywordList.\n\nIf you wish to delete a Search Attribute, please contact support\nat https://support.temporal.io."
	} else {
		s.Command.Long = "Create, list, or remove Search Attributes fields stored in a Workflow\nExecution's metadata:\n\n```\ntemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\n```\n\nSupported types include: Text, Keyword, Int, Double, Bool, Datetime, and\nKeywordList.\n\nIf you wish to delete a Search Attribute, please contact support\nat https://support.temporal.io."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalOperatorSearchAttributeCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorSearchAttributeListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorSearchAttributeRemoveCommand(cctx, &s).Command)
	return &s
}

type TemporalOperatorSearchAttributeCreateCommand struct {
	Parent  *TemporalOperatorSearchAttributeCommand
	Command cobra.Command
	Name    []string
	Type    StringEnumArray
}

func NewTemporalOperatorSearchAttributeCreateCommand(cctx *CommandContext, parent *TemporalOperatorSearchAttributeCommand) *TemporalOperatorSearchAttributeCreateCommand {
	var s TemporalOperatorSearchAttributeCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Add custom Search Attributes"
	if hasHighlighting {
		s.Command.Long = "Add one or more custom Search Attributes:\n\n\x1b[1mtemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\x1b[0m"
	} else {
		s.Command.Long = "Add one or more custom Search Attributes:\n\n```\ntemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Name, "name", nil, "Search Attribute name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Type = NewStringEnumArray([]string{"Text", "Keyword", "Int", "Double", "Bool", "Datetime", "KeywordList"}, []string{})
	s.Command.Flags().Var(&s.Type, "type", "Search Attribute type. Accepted values: Text, Keyword, Int, Double, Bool, Datetime, KeywordList. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "type")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorSearchAttributeListCommand struct {
	Parent  *TemporalOperatorSearchAttributeCommand
	Command cobra.Command
}

func NewTemporalOperatorSearchAttributeListCommand(cctx *CommandContext, parent *TemporalOperatorSearchAttributeCommand) *TemporalOperatorSearchAttributeListCommand {
	var s TemporalOperatorSearchAttributeListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List Search Attributes"
	if hasHighlighting {
		s.Command.Long = "Display a list of active Search Attributes that can be assigned or used with\nWorkflow Queries. You can manage this list and add attributes as needed:\n\n\x1b[1mtemporal operator search-attribute list\x1b[0m"
	} else {
		s.Command.Long = "Display a list of active Search Attributes that can be assigned or used with\nWorkflow Queries. You can manage this list and add attributes as needed:\n\n```\ntemporal operator search-attribute list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalOperatorSearchAttributeRemoveCommand struct {
	Parent  *TemporalOperatorSearchAttributeCommand
	Command cobra.Command
	Name    []string
	Yes     bool
}

func NewTemporalOperatorSearchAttributeRemoveCommand(cctx *CommandContext, parent *TemporalOperatorSearchAttributeCommand) *TemporalOperatorSearchAttributeRemoveCommand {
	var s TemporalOperatorSearchAttributeRemoveCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "remove [flags]"
	s.Command.Short = "Remove custom Search Attributes"
	if hasHighlighting {
		s.Command.Long = "Remove custom Search Attributes from the options that can be assigned or used\nwith Workflow Queries.\n\n\x1b[1mtemporal operator search-attribute remove \\\n    --name YourAttributeName\x1b[0m\n\nRemove attributes without confirmation:\n\n\x1b[1mtemporal operator search-attribute remove \\\n    --name YourAttributeName \\\n    --yes\x1b[0m"
	} else {
		s.Command.Long = "Remove custom Search Attributes from the options that can be assigned or used\nwith Workflow Queries.\n\n```\ntemporal operator search-attribute remove \\\n    --name YourAttributeName\n```\n\nRemove attributes without confirmation:\n\n```\ntemporal operator search-attribute remove \\\n    --name YourAttributeName \\\n    --yes\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Name, "name", nil, "Search Attribute name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm removal.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
	ClientOptions
}

func NewTemporalScheduleCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalScheduleCommand {
	var s TemporalScheduleCommand
	s.Parent = parent
	s.Command.Use = "schedule"
	s.Command.Short = "Perform operations on Schedules"
	if hasHighlighting {
		s.Command.Long = "Create, use, and update Schedules that allow Workflow Executions to be created\nat specified times:\n\n\x1b[1mtemporal schedule [commands] [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal schedule describe \\\n    --schedule-id \"YourScheduleId\"\x1b[0m"
	} else {
		s.Command.Long = "Create, use, and update Schedules that allow Workflow Executions to be created\nat specified times:\n\n```\ntemporal schedule [commands] [options]\n```\n\nFor example:\n\n```\ntemporal schedule describe \\\n    --schedule-id \"YourScheduleId\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalScheduleBackfillCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleCreateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleToggleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleTriggerCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleUpdateCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalScheduleBackfillCommand struct {
	Parent  *TemporalScheduleCommand
	Command cobra.Command
	OverlapPolicyOptions
	ScheduleIdOptions
	EndTime   Timestamp
	StartTime Timestamp
}

func NewTemporalScheduleBackfillCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleBackfillCommand {
	var s TemporalScheduleBackfillCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "backfill [flags]"
	s.Command.Short = "Backfill past actions"
	if hasHighlighting {
		s.Command.Long = "Batch-execute actions that would have run during a specified time interval.\nUse this command to fill in Workflow runs from when a Schedule was paused,\nbefore a Schedule was created, from the future, or to re-process a previously\nexecuted interval.\n\nBackfills require a Schedule ID and the time period covered by the request.\nIt's best to use the \x1b[1mBufferAll\x1b[0m or \x1b[1mAllowAll\x1b[0m policies to avoid conflicts\nand ensure no Workflow Executions are skipped.\n\nFor example:\n\n\x1b[1mtemporal schedule backfill \\\n    --schedule-id \"YourScheduleId\" \\\n    --start-time \"2022-05-01T00:00:00Z\" \\\n    --end-time \"2022-05-31T23:59:59Z\" \\\n    --overlap-policy BufferAll\x1b[0m\n\nThe policies include:\n\n* **AllowAll**: Allow unlimited concurrent Workflow Executions. This\n  significantly speeds up the backfilling process on systems that support\n  concurrency. You must ensure running Workflow Executions do not interfere\n  with each other.\n* **BufferAll**: Buffer all incoming Workflow Executions while waiting for\n  the running Workflow Execution to complete.\n* **Skip**: If a previous Workflow Execution is still running, discard new\n  Workflow Executions.\n* **BufferOne**: Same as 'Skip' but buffer a single Workflow Execution to be\n  run after the previous Execution completes. Discard other Workflow\n  Executions.\n* **CancelOther**: Cancel the running Workflow Execution and replace it with\n  the incoming new Workflow Execution.\n* **TerminateOther**: Terminate the running Workflow Execution and replace\n  it with the incoming new Workflow Execution."
	} else {
		s.Command.Long = "Batch-execute actions that would have run during a specified time interval.\nUse this command to fill in Workflow runs from when a Schedule was paused,\nbefore a Schedule was created, from the future, or to re-process a previously\nexecuted interval.\n\nBackfills require a Schedule ID and the time period covered by the request.\nIt's best to use the `BufferAll` or `AllowAll` policies to avoid conflicts\nand ensure no Workflow Executions are skipped.\n\nFor example:\n\n```\ntemporal schedule backfill \\\n    --schedule-id \"YourScheduleId\" \\\n    --start-time \"2022-05-01T00:00:00Z\" \\\n    --end-time \"2022-05-31T23:59:59Z\" \\\n    --overlap-policy BufferAll\n```\n\nThe policies include:\n\n* **AllowAll**: Allow unlimited concurrent Workflow Executions. This\n  significantly speeds up the backfilling process on systems that support\n  concurrency. You must ensure running Workflow Executions do not interfere\n  with each other.\n* **BufferAll**: Buffer all incoming Workflow Executions while waiting for\n  the running Workflow Execution to complete.\n* **Skip**: If a previous Workflow Execution is still running, discard new\n  Workflow Executions.\n* **BufferOne**: Same as 'Skip' but buffer a single Workflow Execution to be\n  run after the previous Execution completes. Discard other Workflow\n  Executions.\n* **CancelOther**: Cancel the running Workflow Execution and replace it with\n  the incoming new Workflow Execution.\n* **TerminateOther**: Terminate the running Workflow Execution and replace\n  it with the incoming new Workflow Execution."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().Var(&s.EndTime, "end-time", "Backfill end time. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "end-time")
	s.Command.Flags().Var(&s.StartTime, "start-time", "Backfill start time. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "start-time")
	s.OverlapPolicyOptions.buildFlags(cctx, s.Command.Flags())
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleCreateCommand struct {
	Parent  *TemporalScheduleCommand
	Command cobra.Command
	ScheduleConfigurationOptions
	ScheduleIdOptions
	OverlapPolicyOptions
	SharedWorkflowStartOptions
	PayloadInputOptions
}

func NewTemporalScheduleCreateCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleCreateCommand {
	var s TemporalScheduleCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a new Schedule"
	if hasHighlighting {
		s.Command.Long = "Create a new Schedule on the Temporal Service. A Schedule automatically starts\nnew Workflow Executions at the times you specify.\n\nFor example:\n\n\x1b[1m  temporal schedule create \\\n    --schedule-id \"YourScheduleId\" \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"30\"}' \\\n    --workflow-id YourBaseWorkflowIdName \\\n    --task-queue YourTaskQueue \\\n    --type YourWorkflowType\x1b[0m\n\nSchedules support any combination of \x1b[1m--calendar\x1b[0m, \x1b[1m--interval\x1b[0m, and \x1b[1m--cron\x1b[0m:\n\n* Shorthand \x1b[1m--interval\x1b[0m strings.\n  For example: 45m (every 45 minutes) or 6h/5h (every 6 hours, at the top of\n  the 5th hour).\n* JSON \x1b[1m--calendar\x1b[0m, as in the preceding example.\n* Unix-style \x1b[1m--cron\x1b[0m strings and robfig declarations\n  (@daily/@weekly/@every X/etc).\n  For example, every Friday at 12:30 PM: \x1b[1m30 12 * * Fri\x1b[0m."
	} else {
		s.Command.Long = "Create a new Schedule on the Temporal Service. A Schedule automatically starts\nnew Workflow Executions at the times you specify.\n\nFor example:\n\n```\n  temporal schedule create \\\n    --schedule-id \"YourScheduleId\" \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"30\"}' \\\n    --workflow-id YourBaseWorkflowIdName \\\n    --task-queue YourTaskQueue \\\n    --type YourWorkflowType\n```\n\nSchedules support any combination of `--calendar`, `--interval`, and `--cron`:\n\n* Shorthand `--interval` strings.\n  For example: 45m (every 45 minutes) or 6h/5h (every 6 hours, at the top of\n  the 5th hour).\n* JSON `--calendar`, as in the preceding example.\n* Unix-style `--cron` strings and robfig declarations\n  (@daily/@weekly/@every X/etc).\n  For example, every Friday at 12:30 PM: `30 12 * * Fri`."
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleConfigurationOptions.buildFlags(cctx, s.Command.Flags())
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.OverlapPolicyOptions.buildFlags(cctx, s.Command.Flags())
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"name": "type",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleDeleteCommand struct {
	Parent  *TemporalScheduleCommand
	Command cobra.Command
	ScheduleIdOptions
}

func NewTemporalScheduleDeleteCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleDeleteCommand {
	var s TemporalScheduleDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Remove a Schedule"
	if hasHighlighting {
		s.Command.Long = "Deletes a Schedule on the front end Service:\n\n\x1b[1mtemporal schedule delete \\\n    --schedule-id YourScheduleId\x1b[0m\n\nRemoving a Schedule won't affect the Workflow Executions it started that are\nstill running. To cancel or terminate these Workflow Executions, use \x1b[1mtemporal\nworkflow delete\x1b[0m with the \x1b[1mTemporalScheduledById\x1b[0m Search Attribute instead."
	} else {
		s.Command.Long = "Deletes a Schedule on the front end Service:\n\n```\ntemporal schedule delete \\\n    --schedule-id YourScheduleId\n```\n\nRemoving a Schedule won't affect the Workflow Executions it started that are\nstill running. To cancel or terminate these Workflow Executions, use `temporal\nworkflow delete` with the `TemporalScheduledById` Search Attribute instead."
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleDescribeCommand struct {
	Parent  *TemporalScheduleCommand
	Command cobra.Command
	ScheduleIdOptions
}

func NewTemporalScheduleDescribeCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleDescribeCommand {
	var s TemporalScheduleDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Display Schedule state"
	if hasHighlighting {
		s.Command.Long = "Show a Schedule configuration, including information about past, current, and\nfuture Workflow runs:\n\n\x1b[1mtemporal schedule describe \\\n    --schedule-id YourScheduleId\x1b[0m"
	} else {
		s.Command.Long = "Show a Schedule configuration, including information about past, current, and\nfuture Workflow runs:\n\n```\ntemporal schedule describe \\\n    --schedule-id YourScheduleId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleListCommand struct {
	Parent     *TemporalScheduleCommand
	Command    cobra.Command
	Long       bool
	ReallyLong bool
	Query      string
}

func NewTemporalScheduleListCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleListCommand {
	var s TemporalScheduleListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Display hosted Schedules"
	if hasHighlighting {
		s.Command.Long = "Lists the Schedules hosted by a Namespace:\n\n\x1b[1mtemporal schedule list \\\n    --namespace YourNamespace\x1b[0m"
	} else {
		s.Command.Long = "Lists the Schedules hosted by a Namespace:\n\n```\ntemporal schedule list \\\n    --namespace YourNamespace\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVarP(&s.Long, "long", "l", false, "Show detailed information.")
	s.Command.Flags().BoolVar(&s.ReallyLong, "really-long", false, "Show extensive information in non-table form.")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Filter results using given List Filter.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleToggleCommand struct {
	Parent  *TemporalScheduleCommand
	Command cobra.Command
	ScheduleIdOptions
	Pause   bool
	Reason  string
	Unpause bool
}

func NewTemporalScheduleToggleCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleToggleCommand {
	var s TemporalScheduleToggleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "toggle [flags]"
	s.Command.Short = "Pause or unpause a Schedule"
	if hasHighlighting {
		s.Command.Long = "Pause or unpause a Schedule by passing a flag with your desired state:\n\n\x1b[1mtemporal schedule toggle \\\n    --schedule-id \"YourScheduleId\" \\\n    --pause \\\n    --reason \"YourReason\"\x1b[0m\n\nand\n\n\x1b[1mtemporal schedule toggle\n    --schedule-id \"YourScheduleId\" \\\n    --unpause \\\n    --reason \"YourReason\"\x1b[0m\n\nThe \x1b[1m--reason\x1b[0m text updates the Schedule's \x1b[1mnotes\x1b[0m field for operations\ncommunication. It defaults to \"(no reason provided)\" if omitted. This field is\nalso visible on the Service Web UI."
	} else {
		s.Command.Long = "Pause or unpause a Schedule by passing a flag with your desired state:\n\n```\ntemporal schedule toggle \\\n    --schedule-id \"YourScheduleId\" \\\n    --pause \\\n    --reason \"YourReason\"\n```\n\nand\n\n```\ntemporal schedule toggle\n    --schedule-id \"YourScheduleId\" \\\n    --unpause \\\n    --reason \"YourReason\"\n```\n\nThe `--reason` text updates the Schedule's `notes` field for operations\ncommunication. It defaults to \"(no reason provided)\" if omitted. This field is\nalso visible on the Service Web UI."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.Pause, "pause", false, "Pause the Schedule.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "(no reason provided)", "Reason for pausing or unpausing the Schedule.")
	s.Command.Flags().BoolVar(&s.Unpause, "unpause", false, "Unpause the Schedule.")
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleTriggerCommand struct {
	Parent  *TemporalScheduleCommand
	Command cobra.Command
	ScheduleIdOptions
	OverlapPolicyOptions
}

func NewTemporalScheduleTriggerCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleTriggerCommand {
	var s TemporalScheduleTriggerCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "trigger [flags]"
	s.Command.Short = "Immediately run a Schedule"
	if hasHighlighting {
		s.Command.Long = "Trigger a Schedule to run immediately:\n\n\x1b[1mtemporal schedule trigger \\\n    --schedule-id \"YourScheduleId\"\x1b[0m"
	} else {
		s.Command.Long = "Trigger a Schedule to run immediately:\n\n```\ntemporal schedule trigger \\\n    --schedule-id \"YourScheduleId\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.OverlapPolicyOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalScheduleUpdateCommand struct {
	Parent  *TemporalScheduleCommand
	Command cobra.Command
	ScheduleConfigurationOptions
	ScheduleIdOptions
	OverlapPolicyOptions
	SharedWorkflowStartOptions
	PayloadInputOptions
}

func NewTemporalScheduleUpdateCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleUpdateCommand {
	var s TemporalScheduleUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update Schedule details"
	if hasHighlighting {
		s.Command.Long = "Update an existing Schedule with new configuration details, including time\nspecifications, action, and policies:\n\n\x1b[1mtemporal schedule update \\\n    --schedule-id \"YourScheduleId\" \\\n    --workflow-type \"NewWorkflowType\"\x1b[0m"
	} else {
		s.Command.Long = "Update an existing Schedule with new configuration details, including time\nspecifications, action, and policies:\n\n```\ntemporal schedule update \\\n    --schedule-id \"YourScheduleId\" \\\n    --workflow-type \"NewWorkflowType\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleConfigurationOptions.buildFlags(cctx, s.Command.Flags())
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.OverlapPolicyOptions.buildFlags(cctx, s.Command.Flags())
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"name": "type",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalServerCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
}

func NewTemporalServerCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalServerCommand {
	var s TemporalServerCommand
	s.Parent = parent
	s.Command.Use = "server"
	s.Command.Short = "Run Temporal Server"
	if hasHighlighting {
		s.Command.Long = "Run a development Temporal Server on your local system.\n\n\x1b[1m+------------------------------------------------------------------------+\n| WARNING: The development server is not intended for production use.    |\n| It skips certain HTTP security checks to make local use simpler.       |\n|                                                                        |\n| For production use, see:                                               |\n| https://docs.temporal.io/production-deployment                         |\n+------------------------------------------------------------------------+\x1b[0m\n\nView the Web UI for the default configuration at: http://localhost:8233\n\n\x1b[1mtemporal server start-dev\x1b[0m\n\nAdd persistence for Workflow Executions across runs:\n\n\x1b[1mtemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\x1b[0m\n\nSet the port from the front-end gRPC Service (7233 default):\n\n\x1b[1mtemporal server start-dev \\\n    --port 7234 \\\n    --ui-port 8234 \\\n    --metrics-port 57271\x1b[0m\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n\x1b[1mtemporal server start-dev \\\n    --ui-port 3000\x1b[0m"
	} else {
		s.Command.Long = "Run a development Temporal Server on your local system.\n\n```\n+------------------------------------------------------------------------+\n| WARNING: The development server is not intended for production use.    |\n| It skips certain HTTP security checks to make local use simpler.       |\n|                                                                        |\n| For production use, see:                                               |\n| https://docs.temporal.io/production-deployment                         |\n+------------------------------------------------------------------------+\n```\n\nView the Web UI for the default configuration at: http://localhost:8233\n\n```\ntemporal server start-dev\n```\n\nAdd persistence for Workflow Executions across runs:\n\n```\ntemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\n```\n\nSet the port from the front-end gRPC Service (7233 default):\n\n```\ntemporal server start-dev \\\n    --port 7234 \\\n    --ui-port 8234 \\\n    --metrics-port 57271\n```\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n```\ntemporal server start-dev \\\n    --ui-port 3000\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalServerStartDevCommand(cctx, &s).Command)
	return &s
}

type TemporalServerStartDevCommand struct {
	Parent             *TemporalServerCommand
	Command            cobra.Command
	DbFilename         string
	Namespace          []string
	Port               int
	HttpPort           int
	MetricsPort        int
	UiPort             int
	Headless           bool
	Ip                 string
	UiIp               string
	UiPublicPath       string
	UiAssetPath        string
	UiCodecEndpoint    string
	SqlitePragma       []string
	DynamicConfigValue []string
	LogConfig          bool
	SearchAttribute    []string
}

func NewTemporalServerStartDevCommand(cctx *CommandContext, parent *TemporalServerCommand) *TemporalServerStartDevCommand {
	var s TemporalServerStartDevCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "start-dev [flags]"
	s.Command.Short = "Start Temporal development server"
	if hasHighlighting {
		s.Command.Long = "Run a development Temporal Server on your local system.\n\n\x1b[1m+------------------------------------------------------------------------+\n| WARNING: The development server is not intended for production use.    |\n| It skips certain HTTP security checks to make local use simpler.       |\n|                                                                        |\n| For production use, see:                                               |\n| https://docs.temporal.io/production-deployment                         |\n+------------------------------------------------------------------------+\x1b[0m\n\nView the Web UI for the default configuration at: http://localhost:8233\n\n\x1b[1mtemporal server start-dev\x1b[0m\n\nAdd persistence for Workflow Executions across runs:\n\n\x1b[1mtemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\x1b[0m\n\nSet the port from the front-end gRPC Service (7233 default):\n\n\x1b[1mtemporal server start-dev \\\n    --port 7000\x1b[0m\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n\x1b[1mtemporal server start-dev \\\n    --ui-port 3000\x1b[0m"
	} else {
		s.Command.Long = "Run a development Temporal Server on your local system.\n\n```\n+------------------------------------------------------------------------+\n| WARNING: The development server is not intended for production use.    |\n| It skips certain HTTP security checks to make local use simpler.       |\n|                                                                        |\n| For production use, see:                                               |\n| https://docs.temporal.io/production-deployment                         |\n+------------------------------------------------------------------------+\n```\n\nView the Web UI for the default configuration at: http://localhost:8233\n\n```\ntemporal server start-dev\n```\n\nAdd persistence for Workflow Executions across runs:\n\n```\ntemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\n```\n\nSet the port from the front-end gRPC Service (7233 default):\n\n```\ntemporal server start-dev \\\n    --port 7000\n```\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n```\ntemporal server start-dev \\\n    --ui-port 3000\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.DbFilename, "db-filename", "f", "", "Path to file for persistent Temporal state store. By default, Workflow Executions are lost when the server process dies.")
	s.Command.Flags().StringArrayVarP(&s.Namespace, "namespace", "n", nil, "Namespaces to be created at launch. The \"default\" Namespace is always created automatically.")
	s.Command.Flags().IntVarP(&s.Port, "port", "p", 7233, "Port for the front-end gRPC Service.")
	s.Command.Flags().IntVar(&s.HttpPort, "http-port", 0, "Port for the HTTP API service. Defaults to a random free port.")
	s.Command.Flags().IntVar(&s.MetricsPort, "metrics-port", 0, "Port for the '/metrics' HTTP endpoint. Defaults to a random free port.")
	s.Command.Flags().IntVar(&s.UiPort, "ui-port", 0, "Port for the Web UI. Defaults to '--port' value + 1000.")
	s.Command.Flags().BoolVar(&s.Headless, "headless", false, "Disable the Web UI.")
	s.Command.Flags().StringVar(&s.Ip, "ip", "localhost", "IP address bound to the front-end Service.")
	s.Command.Flags().StringVar(&s.UiIp, "ui-ip", "", "IP address bound to the Web UI. Defaults to same as '--ip' value.")
	s.Command.Flags().StringVar(&s.UiPublicPath, "ui-public-path", "", "The public base path for the Web UI. Defaults to `/`.")
	s.Command.Flags().StringVar(&s.UiAssetPath, "ui-asset-path", "", "UI custom assets path.")
	s.Command.Flags().StringVar(&s.UiCodecEndpoint, "ui-codec-endpoint", "", "UI remote codec HTTP endpoint.")
	s.Command.Flags().StringArrayVar(&s.SqlitePragma, "sqlite-pragma", nil, "SQLite pragma statements in \"PRAGMA=VALUE\" format.")
	s.Command.Flags().StringArrayVar(&s.DynamicConfigValue, "dynamic-config-value", nil, "Dynamic configuration value using `KEY=VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: `YourKey=\"YourString\"` Can be passed multiple times.")
	s.Command.Flags().BoolVar(&s.LogConfig, "log-config", false, "Log the server config to stderr.")
	s.Command.Flags().StringArrayVar(&s.SearchAttribute, "search-attribute", nil, "Search attributes to register using `KEY=VALUE` pairs. Keys must be identifiers, and values must be the search attribute type, which is one of the following: Text, Keyword, Int, Double, Bool, Datetime, KeywordList.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
	ClientOptions
}

func NewTemporalTaskQueueCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalTaskQueueCommand {
	var s TemporalTaskQueueCommand
	s.Parent = parent
	s.Command.Use = "task-queue"
	s.Command.Short = "Manage Task Queues"
	if hasHighlighting {
		s.Command.Long = "Inspect and update Task Queues, the queues that Workers poll for Workflow and\nActivity tasks:\n\n\x1b[1mtemporal task-queue [command] [command options] \\\n    --task-queue YourTaskQueue\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal task-queue describe \\\n    --task-queue YourTaskQueue\x1b[0m"
	} else {
		s.Command.Long = "Inspect and update Task Queues, the queues that Workers poll for Workflow and\nActivity tasks:\n\n```\ntemporal task-queue [command] [command options] \\\n    --task-queue YourTaskQueue\n```\n\nFor example:\n\n```\ntemporal task-queue describe \\\n    --task-queue YourTaskQueue\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalTaskQueueConfigCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueGetBuildIdReachabilityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueGetBuildIdsCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueListPartitionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueUpdateBuildIdsCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalTaskQueueConfigCommand struct {
	Parent  *TemporalTaskQueueCommand
	Command cobra.Command
}

func NewTemporalTaskQueueConfigCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueConfigCommand {
	var s TemporalTaskQueueConfigCommand
	s.Parent = parent
	s.Command.Use = "config"
	s.Command.Short = "Get and set Task Queue configuration"
	if hasHighlighting {
		s.Command.Long = "Manage Task Queue configuration:\n\n\x1b[1mtemporal task-queue config [command] [options]\x1b[0m\n\nAvailable commands:\n- \x1b[1mget\x1b[0m: Retrieve the current configuration for a task queue\n- \x1b[1mset\x1b[0m: Update the configuration for a task queue"
	} else {
		s.Command.Long = "Manage Task Queue configuration:\n\n```\ntemporal task-queue config [command] [options]\n```\n\nAvailable commands:\n- `get`: Retrieve the current configuration for a task queue\n- `set`: Update the configuration for a task queue"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalTaskQueueConfigGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueConfigSetCommand(cctx, &s).Command)
	return &s
}

type TemporalTaskQueueConfigGetCommand struct {
	Parent        *TemporalTaskQueueConfigCommand
	Command       cobra.Command
	TaskQueue     string
	TaskQueueType StringEnum
}

func NewTemporalTaskQueueConfigGetCommand(cctx *CommandContext, parent *TemporalTaskQueueConfigCommand) *TemporalTaskQueueConfigGetCommand {
	var s TemporalTaskQueueConfigGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get Task Queue configuration"
	if hasHighlighting {
		s.Command.Long = "Retrieve the current configuration for a Task Queue:\n\n\x1b[1mtemporal task-queue config get \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type activity\x1b[0m\n\nThis command returns the current configuration including:\n- Queue rate limit: The overall rate limit of the task queue.\n  This setting overrides the worker rate limit if set.\n  Unless modified, this is the system-defined rate limit.\n- Fairness key rate limit defaults: Default rate limits for fairness keys.\n  If set, each individual fairness key will be limited to this rate,\n  scaled by the weight of the fairness key."
	} else {
		s.Command.Long = "Retrieve the current configuration for a Task Queue:\n\n```\ntemporal task-queue config get \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type activity\n```\n\nThis command returns the current configuration including:\n- Queue rate limit: The overall rate limit of the task queue.\n  This setting overrides the worker rate limit if set.\n  Unless modified, this is the system-defined rate limit.\n- Fairness key rate limit defaults: Default rate limits for fairness keys.\n  If set, each individual fairness key will be limited to this rate,\n  scaled by the weight of the fairness key."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.TaskQueueType = NewStringEnum([]string{"workflow", "activity", "nexus"}, "")
	s.Command.Flags().Var(&s.TaskQueueType, "task-queue-type", "Task Queue type. Accepted values: workflow, activity, nexus. Accepted values: workflow, activity, nexus. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue-type")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueConfigSetCommand struct {
	Parent                     *TemporalTaskQueueConfigCommand
	Command                    cobra.Command
	TaskQueue                  string
	TaskQueueType              StringEnum
	QueueRpsLimit              string
	QueueRpsLimitReason        string
	FairnessKeyRpsLimitDefault string
	FairnessKeyRpsLimitReason  string
}

func NewTemporalTaskQueueConfigSetCommand(cctx *CommandContext, parent *TemporalTaskQueueConfigCommand) *TemporalTaskQueueConfigSetCommand {
	var s TemporalTaskQueueConfigSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set [flags]"
	s.Command.Short = "Set Task Queue configuration"
	if hasHighlighting {
		s.Command.Long = "Update configuration settings for a Task Queue.\n\n\x1b[1mtemporal task-queue config set \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type activity \\\n    --namespace YourNamespace \\\n    --queue-rps-limit <requests_per_second:float> \\\n    --queue-rps-limit-reason <reason_string> \\\n    --fairness-key-rps-limit-default <requests_per_second:float> \\\n    --fairness-key-rps-limit-reason <reason_string>\x1b[0m\n\nThis command supports updating:\n- Queue rate limits: Controls the overall rate limit of the task queue.\n  This setting overrides the worker rate limit if set.\n  Unless modified, this is the system-defined rate limit.\n- Fairness key rate limit defaults: Sets default rate limits for fairness keys.\n  If set, each individual fairness key will be limited to this rate,\n  scaled by the weight of the fairness key.\n\nTo unset a rate limit, pass in 'default', for example: --queue-rps-limit default"
	} else {
		s.Command.Long = "Update configuration settings for a Task Queue.\n\n```\ntemporal task-queue config set \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type activity \\\n    --namespace YourNamespace \\\n    --queue-rps-limit <requests_per_second:float> \\\n    --queue-rps-limit-reason <reason_string> \\\n    --fairness-key-rps-limit-default <requests_per_second:float> \\\n    --fairness-key-rps-limit-reason <reason_string>\n```\n\nThis command supports updating:\n- Queue rate limits: Controls the overall rate limit of the task queue.\n  This setting overrides the worker rate limit if set.\n  Unless modified, this is the system-defined rate limit.\n- Fairness key rate limit defaults: Sets default rate limits for fairness keys.\n  If set, each individual fairness key will be limited to this rate,\n  scaled by the weight of the fairness key.\n\nTo unset a rate limit, pass in 'default', for example: --queue-rps-limit default"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.TaskQueueType = NewStringEnum([]string{"workflow", "activity", "nexus"}, "")
	s.Command.Flags().Var(&s.TaskQueueType, "task-queue-type", "Task Queue type. Accepted values: workflow, activity, nexus. Accepted values: workflow, activity, nexus. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue-type")
	s.Command.Flags().StringVar(&s.QueueRpsLimit, "queue-rps-limit", "", "Queue rate limit in requests per second. Accepts a float; or 'default' to unset.")
	overrideFlagDisplayType(s.Command.Flags().Lookup("queue-rps-limit"), "float|default")
	s.Command.Flags().StringVar(&s.QueueRpsLimitReason, "queue-rps-limit-reason", "", "Reason for queue rate limit update.")
	s.Command.Flags().StringVar(&s.FairnessKeyRpsLimitDefault, "fairness-key-rps-limit-default", "", "Fairness key rate limit default in requests per second. Accepts a float; or 'default' to unset.")
	overrideFlagDisplayType(s.Command.Flags().Lookup("fairness-key-rps-limit-default"), "float|default")
	s.Command.Flags().StringVar(&s.FairnessKeyRpsLimitReason, "fairness-key-rps-limit-reason", "", "Reason for fairness key rate limit update.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueDescribeCommand struct {
	Parent              *TemporalTaskQueueCommand
	Command             cobra.Command
	TaskQueue           string
	TaskQueueType       StringEnumArray
	SelectBuildId       []string
	SelectUnversioned   bool
	SelectAllActive     bool
	ReportReachability  bool
	LegacyMode          bool
	TaskQueueTypeLegacy StringEnum
	PartitionsLegacy    int
	DisableStats        bool
	ReportConfig        bool
}

func NewTemporalTaskQueueDescribeCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueDescribeCommand {
	var s TemporalTaskQueueDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Show active Workers"
	if hasHighlighting {
		s.Command.Long = "Display a list of active Workers that have recently polled a Task Queue. The\nTemporal Server records each poll request time. A \x1b[1mLastAccessTime\x1b[0m over one\nminute may indicate the Worker is at capacity or has shut down. Temporal\nWorkers are removed if 5 minutes have passed since the last poll request.\n\n\x1b[1mtemporal task-queue describe \\\n  --task-queue YourTaskQueue\x1b[0m\n\nThis command provides poller information for a given Task Queue.\nWorkflow and Activity polling use separate Task Queues:\n\n\x1b[1mtemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type \"activity\"\x1b[0m\n\nThis command provides the following task queue statistics:\n- \x1b[1mApproximateBacklogCount\x1b[0m: The approximate number of tasks backlogged in this\n  task queue. May count expired tasks but eventually converges to the right\n  value.\n- \x1b[1mApproximateBacklogAge\x1b[0m: Approximate age of the oldest task in the backlog,\n  based on its creation time, measured in seconds.\n- \x1b[1mTasksAddRate\x1b[0m: Approximate rate at which tasks are being added to the task\n  queue, measured in tasks per second, averaged over the last 30 seconds.\n  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- \x1b[1mTasksDispatchRate\x1b[0m: Approximate rate at which tasks are being dispatched from\n  the task queue, measured in tasks per second, averaged over the last 30\n  seconds.  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- \x1b[1mBacklogIncreaseRate\x1b[0m: Approximate rate at which the backlog size is\n  increasing (if positive) or decreasing (if negative), measured in tasks per\n  second, averaged over the last 30 seconds.  This is roughly equivalent to:\n  \x1b[1mTasksAddRate\x1b[0m - \x1b[1mTasksDispatchRate\x1b[0m.\n\nNOTE: The \x1b[1mTasksAddRate\x1b[0m and \x1b[1mTasksDispatchRate\x1b[0m metrics may differ from the\nactual rate of add/dispatch, because tasks may be dispatched eagerly to an\navailable worker, or may apply only to specific workers (they are \"sticky\").\nSuch tasks are not counted by these metrics. Despite the inaccuracy of\nthese two metrics, the derived metric of \x1b[1mBacklogIncreaseRate\x1b[0m is accurate\nfor backlogs older than a few seconds.\n\nSafely retire Workers assigned a Build ID by checking reachability across\nall task types. Use the flag \x1b[1m--report-reachability\x1b[0m:\n\n\x1b[1mtemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --select-build-id \"YourBuildId\" \\\n    --report-reachability\x1b[0m\n\nTask reachability information is returned for the requested versions and all\ntask types, which can be used to safely retire Workers with old code versions,\nprovided that they were assigned a Build ID.\n\nNote that task reachability status is experimental and may significantly change\nor be removed in a future release. Also, determining task reachability incurs a\nnon-trivial computing cost.\n\nTask reachability states are reported per build ID. The state may be one of the\nfollowing:\n\n- \x1b[1mReachable\x1b[0m: using the current versioning rules, the Build ID may be used\n  by new Workflow Executions or Activities OR there are currently open\n  Workflow or backlogged Activity tasks assigned to the queue.\n- \x1b[1mClosedWorkflowsOnly\x1b[0m: the Build ID does not have open Workflow Executions\n  and can't be reached by new Workflow Executions. It MAY have closed\n  Workflow Executions within the Namespace retention period.\n- \x1b[1mUnreachable\x1b[0m: this Build ID is not used for new Workflow Executions and\n  isn't used by any existing Workflow Execution within the retention period.\n\nTask reachability is eventually consistent. You may experience a delay until\nreachability converges to the most accurate value. This is designed to act\nin the most conservative way until convergence. For example, \x1b[1mReachable\x1b[0m is\nmore conservative than \x1b[1mClosedWorkflowsOnly\x1b[0m."
	} else {
		s.Command.Long = "Display a list of active Workers that have recently polled a Task Queue. The\nTemporal Server records each poll request time. A `LastAccessTime` over one\nminute may indicate the Worker is at capacity or has shut down. Temporal\nWorkers are removed if 5 minutes have passed since the last poll request.\n\n```\ntemporal task-queue describe \\\n  --task-queue YourTaskQueue\n```\n\nThis command provides poller information for a given Task Queue.\nWorkflow and Activity polling use separate Task Queues:\n\n```\ntemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type \"activity\"\n```\n\nThis command provides the following task queue statistics:\n- `ApproximateBacklogCount`: The approximate number of tasks backlogged in this\n  task queue. May count expired tasks but eventually converges to the right\n  value.\n- `ApproximateBacklogAge`: Approximate age of the oldest task in the backlog,\n  based on its creation time, measured in seconds.\n- `TasksAddRate`: Approximate rate at which tasks are being added to the task\n  queue, measured in tasks per second, averaged over the last 30 seconds.\n  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- `TasksDispatchRate`: Approximate rate at which tasks are being dispatched from\n  the task queue, measured in tasks per second, averaged over the last 30\n  seconds.  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- `BacklogIncreaseRate`: Approximate rate at which the backlog size is\n  increasing (if positive) or decreasing (if negative), measured in tasks per\n  second, averaged over the last 30 seconds.  This is roughly equivalent to:\n  `TasksAddRate` - `TasksDispatchRate`.\n\nNOTE: The `TasksAddRate` and `TasksDispatchRate` metrics may differ from the\nactual rate of add/dispatch, because tasks may be dispatched eagerly to an\navailable worker, or may apply only to specific workers (they are \"sticky\").\nSuch tasks are not counted by these metrics. Despite the inaccuracy of\nthese two metrics, the derived metric of `BacklogIncreaseRate` is accurate\nfor backlogs older than a few seconds.\n\nSafely retire Workers assigned a Build ID by checking reachability across\nall task types. Use the flag `--report-reachability`:\n\n```\ntemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --select-build-id \"YourBuildId\" \\\n    --report-reachability\n```\n\nTask reachability information is returned for the requested versions and all\ntask types, which can be used to safely retire Workers with old code versions,\nprovided that they were assigned a Build ID.\n\nNote that task reachability status is experimental and may significantly change\nor be removed in a future release. Also, determining task reachability incurs a\nnon-trivial computing cost.\n\nTask reachability states are reported per build ID. The state may be one of the\nfollowing:\n\n- `Reachable`: using the current versioning rules, the Build ID may be used\n  by new Workflow Executions or Activities OR there are currently open\n  Workflow or backlogged Activity tasks assigned to the queue.\n- `ClosedWorkflowsOnly`: the Build ID does not have open Workflow Executions\n  and can't be reached by new Workflow Executions. It MAY have closed\n  Workflow Executions within the Namespace retention period.\n- `Unreachable`: this Build ID is not used for new Workflow Executions and\n  isn't used by any existing Workflow Execution within the retention period.\n\nTask reachability is eventually consistent. You may experience a delay until\nreachability converges to the most accurate value. This is designed to act\nin the most conservative way until convergence. For example, `Reachable` is\nmore conservative than `ClosedWorkflowsOnly`."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.TaskQueueType = NewStringEnumArray([]string{"workflow", "activity", "nexus"}, []string{})
	s.Command.Flags().Var(&s.TaskQueueType, "task-queue-type", "Task Queue type. If not specified, all types are reported. Accepted values: workflow, activity, nexus.")
	s.Command.Flags().StringArrayVar(&s.SelectBuildId, "select-build-id", nil, "Filter the Task Queue based on Build ID.")
	s.Command.Flags().BoolVar(&s.SelectUnversioned, "select-unversioned", false, "Include the unversioned queue.")
	s.Command.Flags().BoolVar(&s.SelectAllActive, "select-all-active", false, "Include all active versions. A version is active if it had new tasks or polls recently.")
	s.Command.Flags().BoolVar(&s.ReportReachability, "report-reachability", false, "Display task reachability information.")
	s.Command.Flags().BoolVar(&s.LegacyMode, "legacy-mode", false, "Enable a legacy mode for servers that do not support rules-based worker versioning. This mode only provides pollers info.")
	s.TaskQueueTypeLegacy = NewStringEnum([]string{"workflow", "activity"}, "workflow")
	s.Command.Flags().Var(&s.TaskQueueTypeLegacy, "task-queue-type-legacy", "Task Queue type (legacy mode only). Accepted values: workflow, activity.")
	s.Command.Flags().IntVar(&s.PartitionsLegacy, "partitions-legacy", 1, "Query partitions 1 through `N`. Experimental/Temporary feature. Legacy mode only.")
	s.Command.Flags().BoolVar(&s.DisableStats, "disable-stats", false, "Disable task queue statistics.")
	s.Command.Flags().BoolVar(&s.ReportConfig, "report-config", false, "Include task queue configuration in the response. When enabled, the command will return the current rate limit configuration for the task queue.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueGetBuildIdReachabilityCommand struct {
	Parent           *TemporalTaskQueueCommand
	Command          cobra.Command
	BuildId          []string
	ReachabilityType StringEnum
	TaskQueue        []string
}

func NewTemporalTaskQueueGetBuildIdReachabilityCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueGetBuildIdReachabilityCommand {
	var s TemporalTaskQueueGetBuildIdReachabilityCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get-build-id-reachability [flags]"
	s.Command.Short = "Show Build ID availability (Deprecated)"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\x1b[0m\n\nShow if a given Build ID can be used for new, existing, or closed Workflows\nin Namespaces that support Worker versioning:\n\n\x1b[1mtemporal task-queue get-build-id-reachability \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\nYou can specify the \x1b[1m--build-id\x1b[0m and \x1b[1m--task-queue\x1b[0m flags multiple times. If\n\x1b[1m--task-queue\x1b[0m is omitted, the command checks Build ID reachability against\nall Task Queues."
	} else {
		s.Command.Long = "```\n+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n```\n\nShow if a given Build ID can be used for new, existing, or closed Workflows\nin Namespaces that support Worker versioning:\n\n```\ntemporal task-queue get-build-id-reachability \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\nYou can specify the `--build-id` and `--task-queue` flags multiple times. If\n`--task-queue` is omitted, the command checks Build ID reachability against\nall Task Queues."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.BuildId, "build-id", nil, "One or more Build ID strings. Can be passed multiple times.")
	s.ReachabilityType = NewStringEnum([]string{"open", "closed", "existing"}, "existing")
	s.Command.Flags().Var(&s.ReachabilityType, "reachability-type", "Reachability filter. `open`: reachable by one or more open workflows. `closed`: reachable by one or more closed workflows. `existing`: reachable by either. New Workflow Executions reachable by a Build ID are always reported. Accepted values: open, closed, existing.")
	s.Command.Flags().StringArrayVarP(&s.TaskQueue, "task-queue", "t", nil, "Search only the specified task queue(s). Can be passed multiple times.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueGetBuildIdsCommand struct {
	Parent    *TemporalTaskQueueCommand
	Command   cobra.Command
	TaskQueue string
	MaxSets   int
}

func NewTemporalTaskQueueGetBuildIdsCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueGetBuildIdsCommand {
	var s TemporalTaskQueueGetBuildIdsCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get-build-ids [flags]"
	s.Command.Short = "Fetch Build ID versions (Deprecated)"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\x1b[0m\n\nFetch sets of compatible Build IDs for specified Task Queues and display their\ninformation:\n\n\x1b[1mtemporal task-queue get-build-ids \\\n    --task-queue YourTaskQueue\x1b[0m\n\nThis command is limited to Namespaces that support Worker versioning."
	} else {
		s.Command.Long = "```\n+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n```\n\nFetch sets of compatible Build IDs for specified Task Queues and display their\ninformation:\n\n```\ntemporal task-queue get-build-ids \\\n    --task-queue YourTaskQueue\n```\n\nThis command is limited to Namespaces that support Worker versioning."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Flags().IntVar(&s.MaxSets, "max-sets", 0, "Max return count. Use 1 for default major version. Use 0 for all sets.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueListPartitionCommand struct {
	Parent    *TemporalTaskQueueCommand
	Command   cobra.Command
	TaskQueue string
}

func NewTemporalTaskQueueListPartitionCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueListPartitionCommand {
	var s TemporalTaskQueueListPartitionCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list-partition [flags]"
	s.Command.Short = "List Task Queue partitions"
	if hasHighlighting {
		s.Command.Long = "Display a Task Queue's partition list with assigned matching nodes:\n\n\x1b[1mtemporal task-queue list-partition \\\n    --task-queue YourTaskQueue\x1b[0m"
	} else {
		s.Command.Long = "Display a Task Queue's partition list with assigned matching nodes:\n\n```\ntemporal task-queue list-partition \\\n    --task-queue YourTaskQueue\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueUpdateBuildIdsCommand struct {
	Parent  *TemporalTaskQueueCommand
	Command cobra.Command
}

func NewTemporalTaskQueueUpdateBuildIdsCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueUpdateBuildIdsCommand {
	var s TemporalTaskQueueUpdateBuildIdsCommand
	s.Parent = parent
	s.Command.Use = "update-build-ids"
	s.Command.Short = "Manage Build IDs (Deprecated)"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\x1b[0m\n\nAdd or change a Task Queue's compatible Build IDs for Namespaces using Worker\nversioning:\n\n\x1b[1mtemporal task-queue update-build-ids [subcommands] [options] \\\n    --task-queue YourTaskQueue\x1b[0m"
	} else {
		s.Command.Long = "```\n+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n```\n\nAdd or change a Task Queue's compatible Build IDs for Namespaces using Worker\nversioning:\n\n```\ntemporal task-queue update-build-ids [subcommands] [options] \\\n    --task-queue YourTaskQueue\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalTaskQueueUpdateBuildIdsAddNewCompatibleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueUpdateBuildIdsAddNewDefaultCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueUpdateBuildIdsPromoteIdInSetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueUpdateBuildIdsPromoteSetCommand(cctx, &s).Command)
	return &s
}

type TemporalTaskQueueUpdateBuildIdsAddNewCompatibleCommand struct {
	Parent                    *TemporalTaskQueueUpdateBuildIdsCommand
	Command                   cobra.Command
	BuildId                   string
	TaskQueue                 string
	ExistingCompatibleBuildId string
	SetAsDefault              bool
}

func NewTemporalTaskQueueUpdateBuildIdsAddNewCompatibleCommand(cctx *CommandContext, parent *TemporalTaskQueueUpdateBuildIdsCommand) *TemporalTaskQueueUpdateBuildIdsAddNewCompatibleCommand {
	var s TemporalTaskQueueUpdateBuildIdsAddNewCompatibleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "add-new-compatible [flags]"
	s.Command.Short = "Add compatible Build ID"
	if hasHighlighting {
		s.Command.Long = "Add a compatible Build ID to a Task Queue's existing version set. Provide an\nexisting Build ID and a new Build ID:\n\n\x1b[1mtemporal task-queue update-build-ids add-new-compatible \\\n    --task-queue YourTaskQueue \\\n    --existing-compatible-build-id \"YourExistingBuildId\" \\\n    --build-id \"YourNewBuildId\"\x1b[0m\n\nThe new ID is stored in the set containing the existing ID and becomes the new\ndefault for that set.\n\nThis command is limited to Namespaces that support Worker versioning."
	} else {
		s.Command.Long = "Add a compatible Build ID to a Task Queue's existing version set. Provide an\nexisting Build ID and a new Build ID:\n\n```\ntemporal task-queue update-build-ids add-new-compatible \\\n    --task-queue YourTaskQueue \\\n    --existing-compatible-build-id \"YourExistingBuildId\" \\\n    --build-id \"YourNewBuildId\"\n```\n\nThe new ID is stored in the set containing the existing ID and becomes the new\ndefault for that set.\n\nThis command is limited to Namespaces that support Worker versioning."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Build ID to be added. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Flags().StringVar(&s.ExistingCompatibleBuildId, "existing-compatible-build-id", "", "Pre-existing Build ID in this Task Queue. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "existing-compatible-build-id")
	s.Command.Flags().BoolVar(&s.SetAsDefault, "set-as-default", false, "Set the expanded Build ID set as the Task Queue default.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueUpdateBuildIdsAddNewDefaultCommand struct {
	Parent    *TemporalTaskQueueUpdateBuildIdsCommand
	Command   cobra.Command
	BuildId   string
	TaskQueue string
}

func NewTemporalTaskQueueUpdateBuildIdsAddNewDefaultCommand(cctx *CommandContext, parent *TemporalTaskQueueUpdateBuildIdsCommand) *TemporalTaskQueueUpdateBuildIdsAddNewDefaultCommand {
	var s TemporalTaskQueueUpdateBuildIdsAddNewDefaultCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "add-new-default [flags]"
	s.Command.Short = "Set new default Build ID set (Deprecated)"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\x1b[0m\n\nCreate a new Task Queue Build ID set, add a Build ID to it, and make it the\noverall Task Queue default. The new set will be incompatible with previous\nsets and versions.\n\n\x1b[1mtemporal task-queue update-build-ids add-new-default \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourNewBuildId\"\x1b[0m\n\n\x1b[1m+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "```\n+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n```\n\nCreate a new Task Queue Build ID set, add a Build ID to it, and make it the\noverall Task Queue default. The new set will be incompatible with previous\nsets and versions.\n\n```\ntemporal task-queue update-build-ids add-new-default \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourNewBuildId\"\n```\n\n```\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Build ID to be added. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueUpdateBuildIdsPromoteIdInSetCommand struct {
	Parent    *TemporalTaskQueueUpdateBuildIdsCommand
	Command   cobra.Command
	BuildId   string
	TaskQueue string
}

func NewTemporalTaskQueueUpdateBuildIdsPromoteIdInSetCommand(cctx *CommandContext, parent *TemporalTaskQueueUpdateBuildIdsCommand) *TemporalTaskQueueUpdateBuildIdsPromoteIdInSetCommand {
	var s TemporalTaskQueueUpdateBuildIdsPromoteIdInSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "promote-id-in-set [flags]"
	s.Command.Short = "Set Build ID as set default (Deprecated)"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\x1b[0m\n\nEstablish an existing Build ID as the default in its Task Queue set. New tasks\ncompatible with this set will now be dispatched to this ID:\n\n\x1b[1mtemporal task-queue update-build-ids promote-id-in-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\n\x1b[1m+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "```\n+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n```\n\nEstablish an existing Build ID as the default in its Task Queue set. New tasks\ncompatible with this set will now be dispatched to this ID:\n\n```\ntemporal task-queue update-build-ids promote-id-in-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\n```\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Build ID to set as default. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueUpdateBuildIdsPromoteSetCommand struct {
	Parent    *TemporalTaskQueueUpdateBuildIdsCommand
	Command   cobra.Command
	BuildId   string
	TaskQueue string
}

func NewTemporalTaskQueueUpdateBuildIdsPromoteSetCommand(cctx *CommandContext, parent *TemporalTaskQueueUpdateBuildIdsCommand) *TemporalTaskQueueUpdateBuildIdsPromoteSetCommand {
	var s TemporalTaskQueueUpdateBuildIdsPromoteSetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "promote-set [flags]"
	s.Command.Short = "Promote Build ID set (Deprecated)"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\x1b[0m\n\nPromote a Build ID set to be the default on a Task Queue. Identify the set by\nproviding a Build ID within it. If the set is already the default, this\ncommand has no effect:\n\n\x1b[1mtemporal task-queue update-build-ids promote-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\n\x1b[1m+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "```\n+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n```\n\nPromote a Build ID set to be the default on a Task Queue. Identify the set by\nproviding a Build ID within it. If the set is already the default, this\ncommand has no effect:\n\n```\ntemporal task-queue update-build-ids promote-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\n```\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Build ID within the promoted set. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningCommand struct {
	Parent    *TemporalTaskQueueCommand
	Command   cobra.Command
	TaskQueue string
}

func NewTemporalTaskQueueVersioningCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueVersioningCommand {
	var s TemporalTaskQueueVersioningCommand
	s.Parent = parent
	s.Command.Use = "versioning"
	s.Command.Short = "Manage Task Queue Build ID handling (Experimental)"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m\n\nProvides commands to add, list, remove, or replace Worker Build ID assignment\nand redirect rules associated with Task Queues:\n\n\x1b[1mtemporal task-queue versioning [subcommands] [options] \\\n    --task-queue YourTaskQueue\x1b[0m\n\nTask Queues support the following versioning rules and policies:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID."
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```\n\nProvides commands to add, list, remove, or replace Worker Build ID assignment\nand redirect rules associated with Task Queues:\n\n```\ntemporal task-queue versioning [subcommands] [options] \\\n    --task-queue YourTaskQueue\n```\n\nTask Queues support the following versioning rules and policies:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningAddRedirectRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningCommitBuildIdCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningDeleteAssignmentRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningDeleteRedirectRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningGetRulesCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningInsertAssignmentRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningReplaceAssignmentRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningReplaceRedirectRuleCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.PersistentFlags(), "task-queue")
	return &s
}

type TemporalTaskQueueVersioningAddRedirectRuleCommand struct {
	Parent        *TemporalTaskQueueVersioningCommand
	Command       cobra.Command
	SourceBuildId string
	TargetBuildId string
	Yes           bool
}

func NewTemporalTaskQueueVersioningAddRedirectRuleCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningAddRedirectRuleCommand {
	var s TemporalTaskQueueVersioningAddRedirectRuleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "add-redirect-rule [flags]"
	s.Command.Short = "Add Task Queue redirect rules (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Add a new redirect rule for a given Task Queue. You may add at most one\nredirect rule for each distinct source build ID:\n\n\x1b[1mtemporal task-queue versioning add-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourSourceBuildID\" \\\n    --target-build-id \"YourTargetBuildID\"\x1b[0m\n\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Add a new redirect rule for a given Task Queue. You may add at most one\nredirect rule for each distinct source build ID:\n\n```\ntemporal task-queue versioning add-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourSourceBuildID\" \\\n    --target-build-id \"YourTargetBuildID\"\n```\n\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.SourceBuildId, "source-build-id", "", "Source build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source-build-id")
	s.Command.Flags().StringVar(&s.TargetBuildId, "target-build-id", "", "Target build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "target-build-id")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningCommitBuildIdCommand struct {
	Parent  *TemporalTaskQueueVersioningCommand
	Command cobra.Command
	BuildId string
	Force   bool
	Yes     bool
}

func NewTemporalTaskQueueVersioningCommitBuildIdCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningCommitBuildIdCommand {
	var s TemporalTaskQueueVersioningCommitBuildIdCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "commit-build-id [flags]"
	s.Command.Short = "Complete Build ID rollout (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Complete a Build ID's rollout and clean up unnecessary rules that might have\nbeen created during a gradual rollout:\n\n\x1b[1mtemporal task-queue versioning commit-build-id \\\n    --task-queue YourTaskQueue\n    --build-id \"YourBuildId\"\x1b[0m\n\nThis command automatically applies the following atomic changes:\n\n- Adds an unconditional assignment rule for the target Build ID at the\n  end of the list.\n- Removes all previously added assignment rules to the given target\n  Build ID.\n- Removes any unconditional assignment rules for other Build IDs.\n\nRejects requests when there have been no recent pollers for this Build ID.\nThis prevents committing invalid Build IDs. Use the \x1b[1m--force\x1b[0m option to\noverride this validation.\n\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Complete a Build ID's rollout and clean up unnecessary rules that might have\nbeen created during a gradual rollout:\n\n```\ntemporal task-queue versioning commit-build-id \\\n    --task-queue YourTaskQueue\n    --build-id \"YourBuildId\"\n```\n\nThis command automatically applies the following atomic changes:\n\n- Adds an unconditional assignment rule for the target Build ID at the\n  end of the list.\n- Removes all previously added assignment rules to the given target\n  Build ID.\n- Removes any unconditional assignment rules for other Build IDs.\n\nRejects requests when there have been no recent pollers for this Build ID.\nThis prevents committing invalid Build IDs. Use the `--force` option to\noverride this validation.\n\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Target build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().BoolVar(&s.Force, "force", false, "Bypass recent-poller validation.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningDeleteAssignmentRuleCommand struct {
	Parent    *TemporalTaskQueueVersioningCommand
	Command   cobra.Command
	RuleIndex int
	Force     bool
	Yes       bool
}

func NewTemporalTaskQueueVersioningDeleteAssignmentRuleCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningDeleteAssignmentRuleCommand {
	var s TemporalTaskQueueVersioningDeleteAssignmentRuleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete-assignment-rule [flags]"
	s.Command.Short = "Removes a Task Queue assignment rule (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Deletes a rule identified by its index in the Task Queue's list of assignment\nrules.\n\n\x1b[1mtemporal task-queue versioning delete-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index YourIntegerRuleIndex\x1b[0m\n\nBy default, the Task Queue must retain one unconditional rule, such as \"no\nhint filter\" or \"percentage\". Otherwise, the delete operation is rejected.\nUse the \x1b[1m--force\x1b[0m option to override this validation.\n\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Deletes a rule identified by its index in the Task Queue's list of assignment\nrules.\n\n```\ntemporal task-queue versioning delete-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index YourIntegerRuleIndex\n```\n\nBy default, the Task Queue must retain one unconditional rule, such as \"no\nhint filter\" or \"percentage\". Otherwise, the delete operation is rejected.\nUse the `--force` option to override this validation.\n\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Position of the assignment rule to be replaced. Requests for invalid indices will fail. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "rule-index")
	s.Command.Flags().BoolVar(&s.Force, "force", false, "Bypass one-unconditional-rule validation.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningDeleteRedirectRuleCommand struct {
	Parent        *TemporalTaskQueueVersioningCommand
	Command       cobra.Command
	SourceBuildId string
	Yes           bool
}

func NewTemporalTaskQueueVersioningDeleteRedirectRuleCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningDeleteRedirectRuleCommand {
	var s TemporalTaskQueueVersioningDeleteRedirectRuleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete-redirect-rule [flags]"
	s.Command.Short = "Removes Build-ID routing rule (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Deletes the routing rule for the given source Build ID.\n\n\x1b[1mtemporal task-queue versioning delete-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourBuildId\"\x1b[0m\n\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Deletes the routing rule for the given source Build ID.\n\n```\ntemporal task-queue versioning delete-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourBuildId\"\n```\n\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.SourceBuildId, "source-build-id", "", "Source Build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source-build-id")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningGetRulesCommand struct {
	Parent  *TemporalTaskQueueVersioningCommand
	Command cobra.Command
}

func NewTemporalTaskQueueVersioningGetRulesCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningGetRulesCommand {
	var s TemporalTaskQueueVersioningGetRulesCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get-rules [flags]"
	s.Command.Short = "Fetch Worker Build ID assignments and redirect rules (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Retrieve all the Worker Build ID assignments and redirect rules associated\nwith a Task Queue:\n\n\x1b[1mtemporal task-queue versioning get-rules \\\n    --task-queue YourTaskQueue\x1b[0m\n\nTask Queues support the following versioning rules:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID.\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Retrieve all the Worker Build ID assignments and redirect rules associated\nwith a Task Queue:\n\n```\ntemporal task-queue versioning get-rules \\\n    --task-queue YourTaskQueue\n```\n\nTask Queues support the following versioning rules:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID.\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningInsertAssignmentRuleCommand struct {
	Parent     *TemporalTaskQueueVersioningCommand
	Command    cobra.Command
	BuildId    string
	RuleIndex  int
	Percentage int
	Yes        bool
}

func NewTemporalTaskQueueVersioningInsertAssignmentRuleCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningInsertAssignmentRuleCommand {
	var s TemporalTaskQueueVersioningInsertAssignmentRuleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "insert-assignment-rule [flags]"
	s.Command.Short = "Add an assignment rule at a index (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Inserts a new assignment rule for this Task Queue. Rules are evaluated in\norder, starting from index 0. The first applicable rule is applied, and the\nrest ignored:\n\n\x1b[1mtemporal task-queue versioning insert-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\nIf you do not specify a \x1b[1m--rule-index\x1b[0m, this command inserts at index 0.\n\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Inserts a new assignment rule for this Task Queue. Rules are evaluated in\norder, starting from index 0. The first applicable rule is applied, and the\nrest ignored:\n\n```\ntemporal task-queue versioning insert-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\nIf you do not specify a `--rule-index`, this command inserts at index 0.\n\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Target Build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Insertion position. Ranges from 0 (insert at start) to count (append). Any number greater than the count is treated as \"append\".")
	s.Command.Flags().IntVar(&s.Percentage, "percentage", 100, "Traffic percent to send to target Build ID.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningReplaceAssignmentRuleCommand struct {
	Parent     *TemporalTaskQueueVersioningCommand
	Command    cobra.Command
	BuildId    string
	RuleIndex  int
	Percentage int
	Yes        bool
	Force      bool
}

func NewTemporalTaskQueueVersioningReplaceAssignmentRuleCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningReplaceAssignmentRuleCommand {
	var s TemporalTaskQueueVersioningReplaceAssignmentRuleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "replace-assignment-rule [flags]"
	s.Command.Short = "Update assignment rule at index (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Change an assignment rule for this Task Queue. By default, this enforces one\nunconditional rule (no hint filter or percentage). Otherwise, the operation\nwill be rejected. Set \x1b[1mforce\x1b[0m to true to bypass this validation.\n\n\x1b[1mtemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\"\x1b[0m\n\nTo assign multiple assignment rules to a single Build ID, use\n'insert-assignment-rule'.\n\nTo update the percent:\n\n\x1b[1mtemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\" \\\n    --percentage AnIntegerPercent\x1b[0m\n\nPercent may vary between 0 and 100 (default).\n\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Change an assignment rule for this Task Queue. By default, this enforces one\nunconditional rule (no hint filter or percentage). Otherwise, the operation\nwill be rejected. Set `force` to true to bypass this validation.\n\n```\ntemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\"\n```\n\nTo assign multiple assignment rules to a single Build ID, use\n'insert-assignment-rule'.\n\nTo update the percent:\n\n```\ntemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\" \\\n    --percentage AnIntegerPercent\n```\n\nPercent may vary between 0 and 100 (default).\n\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Target Build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Position of the assignment rule to be replaced. Requests for invalid indices will fail. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "rule-index")
	s.Command.Flags().IntVar(&s.Percentage, "percentage", 100, "Divert percent of traffic to target Build ID.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.")
	s.Command.Flags().BoolVar(&s.Force, "force", false, "Bypass the validation that one unconditional rule remains.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalTaskQueueVersioningReplaceRedirectRuleCommand struct {
	Parent        *TemporalTaskQueueVersioningCommand
	Command       cobra.Command
	SourceBuildId string
	TargetBuildId string
	Yes           bool
}

func NewTemporalTaskQueueVersioningReplaceRedirectRuleCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningReplaceRedirectRuleCommand {
	var s TemporalTaskQueueVersioningReplaceRedirectRuleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "replace-redirect-rule [flags]"
	s.Command.Short = "Change the target for a Build ID's redirect (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Updates a Build ID's redirect rule on a Task Queue by replacing its target\nBuild ID:\n\n\x1b[1mtemporal task-queue versioning replace-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id YourSourceBuildId \\\n    --target-build-id YourNewTargetBuildId\x1b[0m\n\n\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\x1b[0m"
	} else {
		s.Command.Long = "Updates a Build ID's redirect rule on a Task Queue by replacing its target\nBuild ID:\n\n```\ntemporal task-queue versioning replace-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id YourSourceBuildId \\\n    --target-build-id YourNewTargetBuildId\n```\n\n```\n+---------------------------------------------------------------------+\n| CAUTION: This API has been deprecated by Worker Deployment.         |\n+---------------------------------------------------------------------+\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.SourceBuildId, "source-build-id", "", "Source Build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source-build-id")
	s.Command.Flags().StringVar(&s.TargetBuildId, "target-build-id", "", "Target Build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "target-build-id")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
	ClientOptions
}

func NewTemporalWorkerCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalWorkerCommand {
	var s TemporalWorkerCommand
	s.Parent = parent
	s.Command.Use = "worker"
	s.Command.Short = "Read or update Worker state"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker is experimental. Worker commands are subject to     |\n| change.                                                             |\n+---------------------------------------------------------------------+\x1b[0m\n\nModify or read state associated with a Worker, for example,\nusing Worker Deployments commands:\n\n\x1b[1mtemporal worker deployment\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker is experimental. Worker commands are subject to     |\n| change.                                                             |\n+---------------------------------------------------------------------+\n```\n\nModify or read state associated with a Worker, for example,\nusing Worker Deployments commands:\n\n```\ntemporal worker deployment\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkerDeploymentCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerListCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalWorkerDeploymentCommand struct {
	Parent  *TemporalWorkerCommand
	Command cobra.Command
}

func NewTemporalWorkerDeploymentCommand(cctx *CommandContext, parent *TemporalWorkerCommand) *TemporalWorkerDeploymentCommand {
	var s TemporalWorkerDeploymentCommand
	s.Parent = parent
	s.Command.Use = "deployment"
	s.Command.Short = "Describe, list, and operate on Worker Deployments and Versions"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nDeployment commands perform operations on Worker Deployments:\n\n\x1b[1mtemporal worker deployment [command] [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal worker deployment list\x1b[0m\n\nLists the Deployments in the client's namespace.\n\nArguments can be Worker Deployment Versions associated with\na Deployment, specified using the Deployment name and Build ID.\n\nFor example:\n\n\x1b[1mtemporal worker deployment set-current-version \\\n         --deployment-name YourDeploymentName --build-id YourBuildID\x1b[0m\n\nSets the current Deployment Version for a given Deployment."
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nDeployment commands perform operations on Worker Deployments:\n\n```\ntemporal worker deployment [command] [options]\n```\n\nFor example:\n\n```\ntemporal worker deployment list\n```\n\nLists the Deployments in the client's namespace.\n\nArguments can be Worker Deployment Versions associated with\na Deployment, specified using the Deployment name and Build ID.\n\nFor example:\n\n```\ntemporal worker deployment set-current-version \\\n         --deployment-name YourDeploymentName --build-id YourBuildID\n```\n\nSets the current Deployment Version for a given Deployment."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkerDeploymentDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentDeleteVersionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentDescribeVersionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentManagerIdentityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentSetCurrentVersionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentSetRampingVersionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentUpdateMetadataVersionCommand(cctx, &s).Command)
	return &s
}

type TemporalWorkerDeploymentDeleteCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
	DeploymentNameOptions
}

func NewTemporalWorkerDeploymentDeleteCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentDeleteCommand {
	var s TemporalWorkerDeploymentDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a Worker Deployment"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nRemove a Worker Deployment given its Deployment Name.\nA Deployment can only be deleted if it has no Version in it.\n\n\x1b[1mtemporal worker deployment delete [options]\x1b[0m\n\nFor example, setting the user identity that removed the deployment:\n\n\x1b[1mtemporal worker deployment delete \\\n    --name YourDeploymentName \\\n    --identity YourIdentity\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nRemove a Worker Deployment given its Deployment Name.\nA Deployment can only be deleted if it has no Version in it.\n\n```\ntemporal worker deployment delete [options]\n```\n\nFor example, setting the user identity that removed the deployment:\n\n```\ntemporal worker deployment delete \\\n    --name YourDeploymentName \\\n    --identity YourIdentity\n\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.DeploymentNameOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentDeleteVersionCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
	DeploymentVersionOptions
	SkipDrainage bool
}

func NewTemporalWorkerDeploymentDeleteVersionCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentDeleteVersionCommand {
	var s TemporalWorkerDeploymentDeleteVersionCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete-version [flags]"
	s.Command.Short = "Delete a Worker Deployment Version"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nRemove a Worker Deployment Version given its fully-qualified identifier.\nThis is rarely needed during normal operation\nsince unused Versions are eventually garbage collected.\nThe client can delete a Version only when all of the following conditions\nare met:\n  - It is not the Current or Ramping Version for this Deployment.\n  - It has no active pollers, i.e., none of the task queues in the\n  Version have pollers.\n  - It is not draining. This requirement can be ignored with the option\n\x1b[1m--skip-drainage\x1b[0m.\n\x1b[1mtemporal worker deployment delete-version [options]\x1b[0m\n\nFor example, skipping the drainage restriction:\n\n\x1b[1mtemporal worker deployment delete-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --skip-drainage\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nRemove a Worker Deployment Version given its fully-qualified identifier.\nThis is rarely needed during normal operation\nsince unused Versions are eventually garbage collected.\nThe client can delete a Version only when all of the following conditions\nare met:\n  - It is not the Current or Ramping Version for this Deployment.\n  - It has no active pollers, i.e., none of the task queues in the\n  Version have pollers.\n  - It is not draining. This requirement can be ignored with the option\n`--skip-drainage`.\n```\ntemporal worker deployment delete-version [options]\n```\n\nFor example, skipping the drainage restriction:\n\n```\ntemporal worker deployment delete-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --skip-drainage\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.SkipDrainage, "skip-drainage", false, "Ignore the deletion requirement of not draining.")
	s.DeploymentVersionOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentDescribeCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
	DeploymentNameOptions
}

func NewTemporalWorkerDeploymentDescribeCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentDescribeCommand {
	var s TemporalWorkerDeploymentDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Show properties of a Worker Deployment"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nDescribe properties of a Worker Deployment, such as the versions\nassociated with it, routing information of new or existing tasks\nexecuted by this deployment, or its creation time.\n\n\x1b[1mtemporal worker deployment describe [options]\x1b[0m\n\nFor example, to describe a deployment \x1b[1mYourDeploymentName\x1b[0m in the default\nnamespace:\n\n\x1b[1mtemporal worker deployment describe \\\n    --name YourDeploymentName\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nDescribe properties of a Worker Deployment, such as the versions\nassociated with it, routing information of new or existing tasks\nexecuted by this deployment, or its creation time.\n\n```\ntemporal worker deployment describe [options]\n```\n\nFor example, to describe a deployment `YourDeploymentName` in the default\nnamespace:\n\n```\ntemporal worker deployment describe \\\n    --name YourDeploymentName\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.DeploymentNameOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentDescribeVersionCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
	DeploymentVersionOptions
}

func NewTemporalWorkerDeploymentDescribeVersionCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentDescribeVersionCommand {
	var s TemporalWorkerDeploymentDescribeVersionCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe-version [flags]"
	s.Command.Short = "Show properties of a Worker Deployment Version"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nDescribe properties of a Worker Deployment Version, such as the task\nqueues polled by workers in this Deployment Version, or drainage\ninformation required to safely decommission workers, or user-provided\nmetadata, or its creation/modification time.\n\n\x1b[1mtemporal worker deployment describe-version [options]\x1b[0m\n\nFor example, to describe a deployment version  in a deployment\n\x1b[1mYourDeploymentName\x1b[0m, with Build ID \x1b[1mYourBuildID\x1b[0m, and in the default\nnamespace:\n\n\x1b[1mtemporal worker deployment describe-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nDescribe properties of a Worker Deployment Version, such as the task\nqueues polled by workers in this Deployment Version, or drainage\ninformation required to safely decommission workers, or user-provided\nmetadata, or its creation/modification time.\n\n```\ntemporal worker deployment describe-version [options]\n```\n\nFor example, to describe a deployment version  in a deployment\n`YourDeploymentName`, with Build ID `YourBuildID`, and in the default\nnamespace:\n\n```\ntemporal worker deployment describe-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.DeploymentVersionOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentListCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
}

func NewTemporalWorkerDeploymentListCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentListCommand {
	var s TemporalWorkerDeploymentListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Enumerate Worker Deployments in the client's namespace"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nList existing Worker Deployments in the client's namespace.\n\n\x1b[1mtemporal worker deployment list [options]\x1b[0m\n\nFor example, listing Deployments in YourDeploymentNamespace:\n\n\x1b[1mtemporal worker deployment list \\\n    --namespace YourDeploymentNamespace\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nList existing Worker Deployments in the client's namespace.\n\n```\ntemporal worker deployment list [options]\n```\n\nFor example, listing Deployments in YourDeploymentNamespace:\n\n```\ntemporal worker deployment list \\\n    --namespace YourDeploymentNamespace\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentManagerIdentityCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
}

func NewTemporalWorkerDeploymentManagerIdentityCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentManagerIdentityCommand {
	var s TemporalWorkerDeploymentManagerIdentityCommand
	s.Parent = parent
	s.Command.Use = "manager-identity"
	s.Command.Short = "Manager Identity commands change the `ManagerIdentity` of a Worker Deployment"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nManager Identity commands change the \x1b[1mManagerIdentity\x1b[0m of a Worker Deployment:\n\n\x1b[1mtemporal worker deployment manager-identity [command] [options]\x1b[0m\n\nWhen present, \x1b[1mManagerIdentity\x1b[0m is the identity of the user that has the \nexclusive right to make changes to this Worker Deployment. Empty by default.\nWhen set, users whose identity does not match the \x1b[1mManagerIdentity\x1b[0m will not\nbe able to change the Worker Deployment.\n\nThis is especially useful in environments where multiple users (such as CLI\nusers and automated controllers) may interact with the same Worker Deployment.\n\x1b[1mManagerIdentity\x1b[0m allows different users to communicate with one another about\nwho is expected to make changes to the Worker Deployment.\n\nThe current Manager Identity is returned with \x1b[1mdescribe\x1b[0m:\n\x1b[1m temporal worker deployment describe \\\n    --deployment-name YourDeploymentName\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nManager Identity commands change the `ManagerIdentity` of a Worker Deployment:\n\n```\ntemporal worker deployment manager-identity [command] [options]\n```\n\nWhen present, `ManagerIdentity` is the identity of the user that has the \nexclusive right to make changes to this Worker Deployment. Empty by default.\nWhen set, users whose identity does not match the `ManagerIdentity` will not\nbe able to change the Worker Deployment.\n\nThis is especially useful in environments where multiple users (such as CLI\nusers and automated controllers) may interact with the same Worker Deployment.\n`ManagerIdentity` allows different users to communicate with one another about\nwho is expected to make changes to the Worker Deployment.\n\nThe current Manager Identity is returned with `describe`:\n```\n temporal worker deployment describe \\\n    --deployment-name YourDeploymentName\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkerDeploymentManagerIdentitySetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkerDeploymentManagerIdentityUnsetCommand(cctx, &s).Command)
	return &s
}

type TemporalWorkerDeploymentManagerIdentitySetCommand struct {
	Parent          *TemporalWorkerDeploymentManagerIdentityCommand
	Command         cobra.Command
	ManagerIdentity string
	Self            bool
	DeploymentName  string
	Yes             bool
}

func NewTemporalWorkerDeploymentManagerIdentitySetCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentManagerIdentityCommand) *TemporalWorkerDeploymentManagerIdentitySetCommand {
	var s TemporalWorkerDeploymentManagerIdentitySetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set [flags]"
	s.Command.Short = "Set the Manager Identity of a Worker Deployment"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nSet the \x1b[1mManagerIdentity\x1b[0m of a Worker Deployment given its Deployment Name.\n\nWhen present, \x1b[1mManagerIdentity\x1b[0m is the identity of the user that has the \nexclusive right to make changes to this Worker Deployment. Empty by default.\nWhen set, users whose identity does not match the \x1b[1mManagerIdentity\x1b[0m will not\nbe able to change the Worker Deployment.\n\nThis is especially useful in environments where multiple users (such as CLI\nusers and automated controllers) may interact with the same Worker Deployment.\n\x1b[1mManagerIdentity\x1b[0m allows different users to communicate with one another about\nwho is expected to make changes to the Worker Deployment.\n\n\x1b[1mtemporal worker deployment manager-identity set [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal worker deployment manager-identity set \\\n   --deployment-name DeploymentName \\\n   --self \\\n   --identity YourUserIdentity # optional, populated by CLI if not provided\x1b[0m\n\nSets the Manager Identity of the Deployment to the identity of the user making \nthis request. If you don't specifically pass an identity field, the CLI will \ngenerate your identity for you.\n\nFor example:\n\x1b[1mtemporal worker deployment manager-identity set \\\n   --deployment-name DeploymentName \\\n   --manager-identity NewManagerIdentity\x1b[0m\n\nSets the Manager Identity of the Deployment to any string."
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nSet the `ManagerIdentity` of a Worker Deployment given its Deployment Name.\n\nWhen present, `ManagerIdentity` is the identity of the user that has the \nexclusive right to make changes to this Worker Deployment. Empty by default.\nWhen set, users whose identity does not match the `ManagerIdentity` will not\nbe able to change the Worker Deployment.\n\nThis is especially useful in environments where multiple users (such as CLI\nusers and automated controllers) may interact with the same Worker Deployment.\n`ManagerIdentity` allows different users to communicate with one another about\nwho is expected to make changes to the Worker Deployment.\n\n```\ntemporal worker deployment manager-identity set [options]\n```\n\nFor example:\n\n```\ntemporal worker deployment manager-identity set \\\n   --deployment-name DeploymentName \\\n   --self \\\n   --identity YourUserIdentity # optional, populated by CLI if not provided\n```\n\nSets the Manager Identity of the Deployment to the identity of the user making \nthis request. If you don't specifically pass an identity field, the CLI will \ngenerate your identity for you.\n\nFor example:\n```\ntemporal worker deployment manager-identity set \\\n   --deployment-name DeploymentName \\\n   --manager-identity NewManagerIdentity\n```\n\nSets the Manager Identity of the Deployment to any string."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ManagerIdentity, "manager-identity", "", "New Manager Identity. Required unless --self is specified.")
	s.Command.Flags().BoolVar(&s.Self, "self", false, "Set Manager Identity to the identity of the user submitting this request. Required unless --manager-identity is specified.")
	s.Command.Flags().StringVar(&s.DeploymentName, "deployment-name", "", "Name for a Worker Deployment. Required.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm set Manager Identity.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentManagerIdentityUnsetCommand struct {
	Parent         *TemporalWorkerDeploymentManagerIdentityCommand
	Command        cobra.Command
	DeploymentName string
	Yes            bool
}

func NewTemporalWorkerDeploymentManagerIdentityUnsetCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentManagerIdentityCommand) *TemporalWorkerDeploymentManagerIdentityUnsetCommand {
	var s TemporalWorkerDeploymentManagerIdentityUnsetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "unset [flags]"
	s.Command.Short = "Unset the Manager Identity of a Worker Deployment"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nUnset the \x1b[1mManagerIdentity\x1b[0m of a Worker Deployment given its Deployment Name.\n\nWhen present, \x1b[1mManagerIdentity\x1b[0m is the identity of the user that has the \nexclusive right to make changes to this Worker Deployment. Empty by default.\nWhen set, users whose identity does not match the \x1b[1mManagerIdentity\x1b[0m will not\nbe able to change the Worker Deployment.\n\nThis is especially useful in environments where multiple users (such as CLI\nusers and automated controllers) may interact with the same Worker Deployment.\n\x1b[1mManagerIdentity\x1b[0m allows different users to communicate with one another about\nwho is expected to make changes to the Worker Deployment.\n\n\x1b[1mtemporal worker deployment manager-identity unset [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal worker deployment manager-identity unset \\\n   --deployment-name YourDeploymentName\x1b[0m\n\nClears the Manager Identity field for a given Deployment."
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nUnset the `ManagerIdentity` of a Worker Deployment given its Deployment Name.\n\nWhen present, `ManagerIdentity` is the identity of the user that has the \nexclusive right to make changes to this Worker Deployment. Empty by default.\nWhen set, users whose identity does not match the `ManagerIdentity` will not\nbe able to change the Worker Deployment.\n\nThis is especially useful in environments where multiple users (such as CLI\nusers and automated controllers) may interact with the same Worker Deployment.\n`ManagerIdentity` allows different users to communicate with one another about\nwho is expected to make changes to the Worker Deployment.\n\n```\ntemporal worker deployment manager-identity unset [options]\n```\n\nFor example:\n\n```\ntemporal worker deployment manager-identity unset \\\n   --deployment-name YourDeploymentName\n```\n\nClears the Manager Identity field for a given Deployment."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.DeploymentName, "deployment-name", "", "Name for a Worker Deployment. Required.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm unset Manager Identity.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentSetCurrentVersionCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
	DeploymentVersionOrUnversionedOptions
	IgnoreMissingTaskQueues bool
	AllowNoPollers          bool
	Yes                     bool
}

func NewTemporalWorkerDeploymentSetCurrentVersionCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentSetCurrentVersionCommand {
	var s TemporalWorkerDeploymentSetCurrentVersionCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set-current-version [flags]"
	s.Command.Short = "Make a Worker Deployment Version Current for a Deployment"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nSet the Current Version for a Deployment.\nWhen a Version is current, Workers of that Deployment Version will receive\ntasks from new Workflows, and from existing AutoUpgrade Workflows that\nare running on this Deployment.\n\nIf not all the expected Task Queues are being polled by Workers in the\nnew Version the request will fail. To override this protection use\n\x1b[1m--ignore-missing-task-queues\x1b[0m. Note that this would ignore task queues\nin a deployment that are not yet discovered, leading to inconsistent task\nqueue configuration.\n\n\x1b[1mtemporal worker deployment set-current-version [options]\x1b[0m\n\nFor example, to set the Current Version of a deployment\n\x1b[1mYourDeploymentName\x1b[0m, with a version with Build ID \x1b[1mYourBuildID\x1b[0m, and\nin the default namespace:\n\n\x1b[1mtemporal worker deployment set-current-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID\x1b[0m\n\nThe target of set-current-version can also be unversioned workers:\n\n\x1b[1mtemporal worker deployment set-current-version \\\n    --deployment-name YourDeploymentName --unversioned\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nSet the Current Version for a Deployment.\nWhen a Version is current, Workers of that Deployment Version will receive\ntasks from new Workflows, and from existing AutoUpgrade Workflows that\nare running on this Deployment.\n\nIf not all the expected Task Queues are being polled by Workers in the\nnew Version the request will fail. To override this protection use\n`--ignore-missing-task-queues`. Note that this would ignore task queues\nin a deployment that are not yet discovered, leading to inconsistent task\nqueue configuration.\n\n```\ntemporal worker deployment set-current-version [options]\n```\n\nFor example, to set the Current Version of a deployment\n`YourDeploymentName`, with a version with Build ID `YourBuildID`, and\nin the default namespace:\n\n```\ntemporal worker deployment set-current-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID\n```\n\nThe target of set-current-version can also be unversioned workers:\n\n```\ntemporal worker deployment set-current-version \\\n    --deployment-name YourDeploymentName --unversioned\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.IgnoreMissingTaskQueues, "ignore-missing-task-queues", false, "Override protection to accidentally remove task queues.")
	s.Command.Flags().BoolVar(&s.AllowNoPollers, "allow-no-pollers", false, "Override protection and set version as current even if it has no pollers.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm set Current Version.")
	s.DeploymentVersionOrUnversionedOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentSetRampingVersionCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
	DeploymentVersionOrUnversionedOptions
	Percentage              float32
	Delete                  bool
	IgnoreMissingTaskQueues bool
	AllowNoPollers          bool
	Yes                     bool
}

func NewTemporalWorkerDeploymentSetRampingVersionCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentSetRampingVersionCommand {
	var s TemporalWorkerDeploymentSetRampingVersionCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "set-ramping-version [flags]"
	s.Command.Short = "Change Version Ramping settings for a Worker Deployment"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nSet the Ramping Version and Percentage for a Deployment.\n\nThe Ramping Version can be set using deployment name and build ID,\nor set to unversioned workers using the --unversioned flag.\n\nThe Ramping Percentage is a float with values in the range [0, 100].\nA value of 100 does not make the Ramping Version Current, use\n\x1b[1mset-current-version\x1b[0m instead.\n\nTo remove a Ramping Version use the flag \x1b[1m--delete\x1b[0m.\n\nIf not all the expected Task Queues are being polled by Workers in the\nnew Ramping Version the request will fail. To override this protection use\n\x1b[1m--ignore-missing-task-queues\x1b[0m. Note that this would ignore task queues\nin a deployment that are not yet discovered, leading to inconsistent task\nqueue configuration.\n\n\x1b[1mtemporal worker deployment set-ramping-version [options]\x1b[0m\n\nFor example, to set the Ramping Version of a deployment\n\x1b[1mYourDeploymentName\x1b[0m, with a version with Build ID \x1b[1mYourBuildID\x1b[0m, with\n10 percent of tasks redirected to this version, and\nusing the default namespace:\n\n\x1b[1mtemporal worker deployment set-ramping-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --percentage 10.0\x1b[0m\n\nAnd to remove that ramping:\n\x1b[1mtemporal worker deployment set-ramping-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --delete\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nSet the Ramping Version and Percentage for a Deployment.\n\nThe Ramping Version can be set using deployment name and build ID,\nor set to unversioned workers using the --unversioned flag.\n\nThe Ramping Percentage is a float with values in the range [0, 100].\nA value of 100 does not make the Ramping Version Current, use\n`set-current-version` instead.\n\nTo remove a Ramping Version use the flag `--delete`.\n\nIf not all the expected Task Queues are being polled by Workers in the\nnew Ramping Version the request will fail. To override this protection use\n`--ignore-missing-task-queues`. Note that this would ignore task queues\nin a deployment that are not yet discovered, leading to inconsistent task\nqueue configuration.\n\n```\ntemporal worker deployment set-ramping-version [options]\n```\n\nFor example, to set the Ramping Version of a deployment\n`YourDeploymentName`, with a version with Build ID `YourBuildID`, with\n10 percent of tasks redirected to this version, and\nusing the default namespace:\n\n```\ntemporal worker deployment set-ramping-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --percentage 10.0\n```\n\nAnd to remove that ramping:\n```\ntemporal worker deployment set-ramping-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --delete\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().Float32Var(&s.Percentage, "percentage", 0, "Percentage of tasks redirected to the Ramping Version. Valid range [0,100].")
	s.Command.Flags().BoolVar(&s.Delete, "delete", false, "Delete the Ramping Version.")
	s.Command.Flags().BoolVar(&s.IgnoreMissingTaskQueues, "ignore-missing-task-queues", false, "Override protection to accidentally remove task queues.")
	s.Command.Flags().BoolVar(&s.AllowNoPollers, "allow-no-pollers", false, "Override protection and set version as ramping even if it has no pollers.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm set Ramping Version.")
	s.DeploymentVersionOrUnversionedOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDeploymentUpdateMetadataVersionCommand struct {
	Parent  *TemporalWorkerDeploymentCommand
	Command cobra.Command
	DeploymentVersionOptions
	Metadata      []string
	RemoveEntries []string
}

func NewTemporalWorkerDeploymentUpdateMetadataVersionCommand(cctx *CommandContext, parent *TemporalWorkerDeploymentCommand) *TemporalWorkerDeploymentUpdateMetadataVersionCommand {
	var s TemporalWorkerDeploymentUpdateMetadataVersionCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update-metadata-version [flags]"
	s.Command.Short = "Change user-provided metadata for a Version"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\x1b[0m\n\nUpdate metadata associated with a Worker Deployment Version.\n\nFor example:\n\n\x1b[1m temporal worker deployment update-metadata-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --metadata bar=1 \\\n    --metadata foo=true\x1b[0m\n\nThe current metadata is also returned with \x1b[1mdescribe-version\x1b[0m:\n\x1b[1m temporal worker deployment describe-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worker Deployment is experimental. Deployment commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n```\n\nUpdate metadata associated with a Worker Deployment Version.\n\nFor example:\n\n```\n temporal worker deployment update-metadata-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n    --metadata bar=1 \\\n    --metadata foo=true\n```\n\nThe current metadata is also returned with `describe-version`:\n```\n temporal worker deployment describe-version \\\n    --deployment-name YourDeploymentName --build-id YourBuildID \\\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Metadata, "metadata", nil, "Set deployment metadata using `KEY=\"VALUE\"` pairs. Keys must be identifiers, and values must be JSON values. For example: `YourKey={\"your\": \"value\"}` Can be passed multiple times.")
	s.Command.Flags().StringArrayVar(&s.RemoveEntries, "remove-entries", nil, "Keys of entries to be deleted from metadata. Can be passed multiple times.")
	s.DeploymentVersionOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerDescribeCommand struct {
	Parent            *TemporalWorkerCommand
	Command           cobra.Command
	WorkerInstanceKey string
}

func NewTemporalWorkerDescribeCommand(cctx *CommandContext, parent *TemporalWorkerCommand) *TemporalWorkerDescribeCommand {
	var s TemporalWorkerDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Returns information about a specific worker (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Look up information of a specific worker.\n\n\x1b[1mtemporal worker describe --namespace YourNamespace --worker-instance-key YourKey\x1b[0m"
	} else {
		s.Command.Long = "Look up information of a specific worker.\n\n```\ntemporal worker describe --namespace YourNamespace --worker-instance-key YourKey\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.WorkerInstanceKey, "worker-instance-key", "", "Worker instance key to describe. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "worker-instance-key")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkerListCommand struct {
	Parent  *TemporalWorkerCommand
	Command cobra.Command
	Query   string
	Limit   int
}

func NewTemporalWorkerListCommand(cctx *CommandContext, parent *TemporalWorkerCommand) *TemporalWorkerListCommand {
	var s TemporalWorkerListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "List worker status information in a specific namespace (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Get a list of workers to the specified namespace.\n\n\x1b[1mtemporal worker list --namespace YourNamespace --query 'taskQueue=\"YourTaskQueue\"'\x1b[0m"
	} else {
		s.Command.Long = "Get a list of workers to the specified namespace.\n\n```\ntemporal worker list --namespace YourNamespace --query 'taskQueue=\"YourTaskQueue\"'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.")
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Maximum number of workers to display.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowCommand struct {
	Parent  *TemporalCommand
	Command cobra.Command
	ClientOptions
}

func NewTemporalWorkflowCommand(cctx *CommandContext, parent *TemporalCommand) *TemporalWorkflowCommand {
	var s TemporalWorkflowCommand
	s.Parent = parent
	s.Command.Use = "workflow"
	s.Command.Short = "Start, list, and operate on Workflows"
	if hasHighlighting {
		s.Command.Long = "Workflow commands perform operations on Workflow Executions:\n\n\x1b[1mtemporal workflow [command] [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal workflow list\x1b[0m"
	} else {
		s.Command.Long = "Workflow commands perform operations on Workflow Executions:\n\n```\ntemporal workflow [command] [options]\n```\n\nFor example:\n\n```\ntemporal workflow list\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkflowCancelCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowCountCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowExecuteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowExecuteUpdateWithStartCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowFixHistoryJsonCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowMetadataCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowQueryCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowResetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowResultCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowShowCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowSignalCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowSignalWithStartCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowStackCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowStartCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowStartUpdateWithStartCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowTerminateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowTraceCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowUpdateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowUpdateOptionsCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalWorkflowCancelCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SingleWorkflowOrBatchOptions
}

func NewTemporalWorkflowCancelCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowCancelCommand {
	var s TemporalWorkflowCancelCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "cancel [flags]"
	s.Command.Short = "Send cancellation to Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Canceling a running Workflow Execution records a\n\x1b[1mWorkflowExecutionCancelRequested\x1b[0m event in the Event History. The Service\nschedules a new Command Task, and the Workflow Execution performs any cleanup\nwork supported by its implementation.\n\nUse the Workflow ID to cancel an Execution:\n\n\x1b[1mtemporal workflow cancel \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nA visibility Query lets you send bulk cancellations to Workflow Executions\nmatching the results:\n\n\x1b[1mtemporal workflow cancel \\\n    --query YourQuery\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference."
	} else {
		s.Command.Long = "Canceling a running Workflow Execution records a\n`WorkflowExecutionCancelRequested` event in the Event History. The Service\nschedules a new Command Task, and the Workflow Execution performs any cleanup\nwork supported by its implementation.\n\nUse the Workflow ID to cancel an Execution:\n\n```\ntemporal workflow cancel \\\n    --workflow-id YourWorkflowId\n```\n\nA visibility Query lets you send bulk cancellations to Workflow Executions\nmatching the results:\n\n```\ntemporal workflow cancel \\\n    --query YourQuery\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference."
	}
	s.Command.Args = cobra.NoArgs
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowCountCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	Query   string
}

func NewTemporalWorkflowCountCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowCountCommand {
	var s TemporalWorkflowCountCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "count [flags]"
	s.Command.Short = "Number of Workflow Executions"
	if hasHighlighting {
		s.Command.Long = "Show a count of Workflow Executions, regardless of execution state (running,\nterminated, etc). Use \x1b[1m--query\x1b[0m to select a subset of Workflow Executions:\n\n\x1b[1mtemporal workflow count \\\n    --query YourQuery\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference."
	} else {
		s.Command.Long = "Show a count of Workflow Executions, regardless of execution state (running,\nterminated, etc). Use `--query` to select a subset of Workflow Executions:\n\n```\ntemporal workflow count \\\n    --query YourQuery\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowDeleteCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SingleWorkflowOrBatchOptions
}

func NewTemporalWorkflowDeleteCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowDeleteCommand {
	var s TemporalWorkflowDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Remove Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Delete a Workflow Executions and its Event History:\n\n\x1b[1mtemporal workflow delete \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nThe removal executes asynchronously. If the Execution is Running, the Service\nterminates it before deletion.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference."
	} else {
		s.Command.Long = "Delete a Workflow Executions and its Event History:\n\n```\ntemporal workflow delete \\\n    --workflow-id YourWorkflowId\n```\n\nThe removal executes asynchronously. If the Execution is Running, the Service\nterminates it before deletion.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference."
	}
	s.Command.Args = cobra.NoArgs
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowDescribeCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowReferenceOptions
	ResetPoints bool
	Raw         bool
}

func NewTemporalWorkflowDescribeCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowDescribeCommand {
	var s TemporalWorkflowDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Show Workflow Execution info"
	if hasHighlighting {
		s.Command.Long = "Display information about a specific Workflow Execution:\n\n\x1b[1mtemporal workflow describe \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nShow the Workflow Execution's auto-reset points:\n\n\x1b[1mtemporal workflow describe \\\n    --workflow-id YourWorkflowId \\\n    --reset-points true\x1b[0m"
	} else {
		s.Command.Long = "Display information about a specific Workflow Execution:\n\n```\ntemporal workflow describe \\\n    --workflow-id YourWorkflowId\n```\n\nShow the Workflow Execution's auto-reset points:\n\n```\ntemporal workflow describe \\\n    --workflow-id YourWorkflowId \\\n    --reset-points true\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.ResetPoints, "reset-points", false, "Show auto-reset points only.")
	s.Command.Flags().BoolVar(&s.Raw, "raw", false, "Print properties without changing their format.")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowExecuteCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SharedWorkflowStartOptions
	WorkflowStartOptions
	PayloadInputOptions
	Detailed bool
}

func NewTemporalWorkflowExecuteCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowExecuteCommand {
	var s TemporalWorkflowExecuteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "execute [flags]"
	s.Command.Short = "Start new Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Establish a new Workflow Execution and direct its progress to stdout. The\ncommand blocks and returns when the Workflow Execution completes. If your\nWorkflow requires input, pass valid JSON:\n\n\x1b[1mtemporal workflow execute\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\x1b[0m\n\nUse \x1b[1m--event-details\x1b[0m to relay updates to the command-line output in JSON\nformat. When using JSON output (\x1b[1m--output json\x1b[0m), this includes the entire\n\"history\" JSON key for the run."
	} else {
		s.Command.Long = "Establish a new Workflow Execution and direct its progress to stdout. The\ncommand blocks and returns when the Workflow Execution completes. If your\nWorkflow requires input, pass valid JSON:\n\n```\ntemporal workflow execute\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\n```\n\nUse `--event-details` to relay updates to the command-line output in JSON\nformat. When using JSON output (`--output json`), this includes the entire\n\"history\" JSON key for the run."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.Detailed, "detailed", false, "Display events as sections instead of table. Does not apply to JSON output.")
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"name": "type",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowExecuteUpdateWithStartCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SharedWorkflowStartOptions
	WorkflowStartOptions
	PayloadInputOptions
	UpdateName                string
	UpdateFirstExecutionRunId string
	UpdateId                  string
	RunId                     string
	UpdateInput               []string
	UpdateInputFile           []string
	UpdateInputMeta           []string
	UpdateInputBase64         bool
}

func NewTemporalWorkflowExecuteUpdateWithStartCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowExecuteUpdateWithStartCommand {
	var s TemporalWorkflowExecuteUpdateWithStartCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "execute-update-with-start [flags]"
	s.Command.Short = "Send an Update-With-Start and wait for it to complete (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to complete. If the Workflow Execution is not running, then a new workflow\nexecution is started and the update is sent.\n\nExperimental.\n\n\x1b[1mtemporal workflow execute-update-with-start \\\n  --update-name YourUpdate \\\n  --update-input '{\"update-key\": \"update-value\"}' \\\n  --workflow-id YourWorkflowId \\\n  --type YourWorkflowType \\\n  --task-queue YourTaskQueue \\\n  --id-conflict-policy Fail \\\n  --input '{\"wf-key\": \"wf-value\"}'\x1b[0m"
	} else {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to complete. If the Workflow Execution is not running, then a new workflow\nexecution is started and the update is sent.\n\nExperimental.\n\n```\ntemporal workflow execute-update-with-start \\\n  --update-name YourUpdate \\\n  --update-input '{\"update-key\": \"update-value\"}' \\\n  --workflow-id YourWorkflowId \\\n  --type YourWorkflowType \\\n  --task-queue YourTaskQueue \\\n  --id-conflict-policy Fail \\\n  --input '{\"wf-key\": \"wf-value\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.UpdateName, "update-name", "", "Update name. Required. Aliased as \"--update-type\".")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "update-name")
	s.Command.Flags().StringVar(&s.UpdateFirstExecutionRunId, "update-first-execution-run-id", "", "Parent Run ID. The update is sent to the last Workflow Execution in the chain started with this Run ID.")
	s.Command.Flags().StringVar(&s.UpdateId, "update-id", "", "Update ID. If unset, defaults to a UUID.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID. If unset, looks for an Update against the currently-running Workflow Execution.")
	s.Command.Flags().StringArrayVar(&s.UpdateInput, "update-input", nil, "Update input value. Use JSON content or set --update-input-meta to override. Can't be combined with --update-input-file. Can be passed multiple times to pass multiple arguments.")
	s.Command.Flags().StringArrayVar(&s.UpdateInputFile, "update-input-file", nil, "A path or paths for input file(s). Use JSON content or set --update-input-meta to override. Can't be combined with --update-input. Can be passed multiple times to pass multiple arguments.")
	s.Command.Flags().StringArrayVar(&s.UpdateInputMeta, "update-input-meta", nil, "Input update payload metadata as a `KEY=VALUE` pair. When the KEY is \"encoding\", this overrides the default (\"json/plain\"). Can be passed multiple times.")
	s.Command.Flags().BoolVar(&s.UpdateInputBase64, "update-input-base64", false, "Assume update inputs are base64-encoded and attempt to decode them.")
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"name":        "type",
		"update-type": "update-name",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowFixHistoryJsonCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	Source  string
	Target  string
}

func NewTemporalWorkflowFixHistoryJsonCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowFixHistoryJsonCommand {
	var s TemporalWorkflowFixHistoryJsonCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "fix-history-json [flags]"
	s.Command.Short = "Updates an event history JSON file"
	if hasHighlighting {
		s.Command.Long = "Reserialize an Event History JSON file:\n\n\x1b[1mtemporal workflow fix-history-json \\\n    --source /path/to/original.json \\\n    --target /path/to/reserialized.json\x1b[0m"
	} else {
		s.Command.Long = "Reserialize an Event History JSON file:\n\n```\ntemporal workflow fix-history-json \\\n    --source /path/to/original.json \\\n    --target /path/to/reserialized.json\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Source, "source", "s", "", "Path to the original file. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source")
	s.Command.Flags().StringVarP(&s.Target, "target", "t", "", "Path to the results file. When omitted, output is sent to stdout.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowListCommand struct {
	Parent   *TemporalWorkflowCommand
	Command  cobra.Command
	Query    string
	Archived bool
	Limit    int
	PageSize int
}

func NewTemporalWorkflowListCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowListCommand {
	var s TemporalWorkflowListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Show Workflow Executions"
	if hasHighlighting {
		s.Command.Long = "List Workflow Executions. The optional \x1b[1m--query\x1b[0m limits the output to\nWorkflows matching a Query:\n\n\x1b[1mtemporal workflow list \\\n    --query YourQuery\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference.\n\nView a list of archived Workflow Executions:\n\n\x1b[1mtemporal workflow list \\\n    --archived\x1b[0m"
	} else {
		s.Command.Long = "List Workflow Executions. The optional `--query` limits the output to\nWorkflows matching a Query:\n\n```\ntemporal workflow list \\\n    --query YourQuery\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference.\n\nView a list of archived Workflow Executions:\n\n```\ntemporal workflow list \\\n    --archived\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.")
	s.Command.Flags().BoolVar(&s.Archived, "archived", false, "Limit output to archived Workflow Executions. EXPERIMENTAL.")
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Maximum number of Workflow Executions to display.")
	s.Command.Flags().IntVar(&s.PageSize, "page-size", 0, "Maximum number of Workflow Executions to fetch at a time from the server.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowMetadataCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowReferenceOptions
	QueryModifiersOptions
}

func NewTemporalWorkflowMetadataCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowMetadataCommand {
	var s TemporalWorkflowMetadataCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "metadata [flags]"
	s.Command.Short = "Query the Workflow for user-specified metadata"
	if hasHighlighting {
		s.Command.Long = "Issue a Query for and display user-set metadata like summary and\ndetails for a specific Workflow Execution:\n\n\x1b[1mtemporal workflow metadata \\\n    --workflow-id YourWorkflowId\x1b[0m"
	} else {
		s.Command.Long = "Issue a Query for and display user-set metadata like summary and\ndetails for a specific Workflow Execution:\n\n```\ntemporal workflow metadata \\\n    --workflow-id YourWorkflowId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.QueryModifiersOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowQueryCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	PayloadInputOptions
	WorkflowReferenceOptions
	QueryModifiersOptions
	Name string
}

func NewTemporalWorkflowQueryCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowQueryCommand {
	var s TemporalWorkflowQueryCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "query [flags]"
	s.Command.Short = "Retrieve Workflow Execution state"
	if hasHighlighting {
		s.Command.Long = "Send a Query to a Workflow Execution by Workflow ID to retrieve its state.\nThis synchronous operation exposes the internal state of a running Workflow\nExecution, which constantly changes. You can query both running and completed\nWorkflow Executions:\n\n\x1b[1mtemporal workflow query \\\n    --workflow-id YourWorkflowId\n    --type YourQueryType\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\x1b[0m"
	} else {
		s.Command.Long = "Send a Query to a Workflow Execution by Workflow ID to retrieve its state.\nThis synchronous operation exposes the internal state of a running Workflow\nExecution, which constantly changes. You can query both running and completed\nWorkflow Executions:\n\n```\ntemporal workflow query \\\n    --workflow-id YourWorkflowId\n    --type YourQueryType\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Query Type/Name. Required. Aliased as \"--type\".")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.QueryModifiersOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"type": "name",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowResetCommand struct {
	Parent         *TemporalWorkflowCommand
	Command        cobra.Command
	WorkflowId     string
	RunId          string
	EventId        int
	Reason         string
	ReapplyType    StringEnum
	ReapplyExclude StringEnumArray
	Type           StringEnum
	BuildId        string
	Query          string
	Yes            bool
}

func NewTemporalWorkflowResetCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowResetCommand {
	var s TemporalWorkflowResetCommand
	s.Parent = parent
	s.Command.Use = "reset"
	s.Command.Short = "Move Workflow Execution history point"
	if hasHighlighting {
		s.Command.Long = "Reset a Workflow Execution so it can resume from a point in its Event History\nwithout losing its progress up to that point:\n\n\x1b[1mtemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --event-id YourLastEvent\x1b[0m\n\nStart from where the Workflow Execution last continued as new:\n\n\x1b[1mtemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --type LastContinuedAsNew\x1b[0m\n\nFor batch resets, limit your resets to FirstWorkflowTask, LastWorkflowTask, or\nBuildId. Do not use Workflow IDs, run IDs, or event IDs with this command.\n\nVisit https://docs.temporal.io/visibility to read more about Search\nAttributes and Query creation."
	} else {
		s.Command.Long = "Reset a Workflow Execution so it can resume from a point in its Event History\nwithout losing its progress up to that point:\n\n```\ntemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --event-id YourLastEvent\n```\n\nStart from where the Workflow Execution last continued as new:\n\n```\ntemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --type LastContinuedAsNew\n```\n\nFor batch resets, limit your resets to FirstWorkflowTask, LastWorkflowTask, or\nBuildId. Do not use Workflow IDs, run IDs, or event IDs with this command.\n\nVisit https://docs.temporal.io/visibility to read more about Search\nAttributes and Query creation."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkflowResetWithWorkflowUpdateOptionsCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow ID. Required for non-batch reset operations.")
	s.Command.PersistentFlags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID.")
	s.Command.PersistentFlags().IntVarP(&s.EventId, "event-id", "e", 0, "Event ID to reset to. Event must occur after `WorkflowTaskStarted`. `WorkflowTaskCompleted`, `WorkflowTaskFailed`, etc. are valid.")
	s.Command.PersistentFlags().StringVar(&s.Reason, "reason", "", "Reason for reset. Required.")
	_ = cobra.MarkFlagRequired(s.Command.PersistentFlags(), "reason")
	s.ReapplyType = NewStringEnum([]string{"All", "Signal", "None"}, "All")
	s.Command.PersistentFlags().Var(&s.ReapplyType, "reapply-type", "Types of events to re-apply after reset point. Accepted values: All, Signal, None.")
	_ = s.Command.PersistentFlags().MarkDeprecated("reapply-type", "Use --reapply-exclude instead.")
	s.ReapplyExclude = NewStringEnumArray([]string{"All", "Signal", "Update"}, []string{})
	s.Command.PersistentFlags().Var(&s.ReapplyExclude, "reapply-exclude", "Exclude these event types from re-application. Accepted values: All, Signal, Update.")
	s.Type = NewStringEnum([]string{"FirstWorkflowTask", "LastWorkflowTask", "LastContinuedAsNew", "BuildId"}, "")
	s.Command.PersistentFlags().VarP(&s.Type, "type", "t", "The event type for the reset. Accepted values: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew, BuildId.")
	s.Command.PersistentFlags().StringVar(&s.BuildId, "build-id", "", "A Build ID. Use only with the BuildId `--type`. Resets the first Workflow task processed by this ID. By default, this reset may be in a prior run, earlier than a Continue as New point.")
	s.Command.PersistentFlags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.")
	s.Command.PersistentFlags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm. Only allowed when `--query` is present.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowResetWithWorkflowUpdateOptionsCommand struct {
	Parent  *TemporalWorkflowResetCommand
	Command cobra.Command
	WorkflowUpdateOptionsOptions
}

func NewTemporalWorkflowResetWithWorkflowUpdateOptionsCommand(cctx *CommandContext, parent *TemporalWorkflowResetCommand) *TemporalWorkflowResetWithWorkflowUpdateOptionsCommand {
	var s TemporalWorkflowResetWithWorkflowUpdateOptionsCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "with-workflow-update-options [flags]"
	s.Command.Short = "Update options on reset workflow"
	s.Command.Long = "Run Workflow Update Options atomically after the Workflow is reset.\nWorkflows selected by the reset command are forwarded onto the subcommand."
	s.Command.Args = cobra.NoArgs
	s.WorkflowUpdateOptionsOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowResultCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowReferenceOptions
}

func NewTemporalWorkflowResultCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowResultCommand {
	var s TemporalWorkflowResultCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "result [flags]"
	s.Command.Short = "Wait for and show the result of a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Wait for and print the result of a Workflow Execution:\n\n\x1b[1mtemporal workflow result \\\n    --workflow-id YourWorkflowId\x1b[0m"
	} else {
		s.Command.Long = "Wait for and print the result of a Workflow Execution:\n\n```\ntemporal workflow result \\\n    --workflow-id YourWorkflowId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowShowCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowReferenceOptions
	Follow   bool
	Detailed bool
}

func NewTemporalWorkflowShowCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowShowCommand {
	var s TemporalWorkflowShowCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "show [flags]"
	s.Command.Short = "Display Event History"
	if hasHighlighting {
		s.Command.Long = "Show a Workflow Execution's Event History.\nWhen using JSON output (\x1b[1m--output json\x1b[0m), you may pass the results to an SDK\nto perform a replay:\n\n\x1b[1mtemporal workflow show \\\n    --workflow-id YourWorkflowId\n    --output json\x1b[0m"
	} else {
		s.Command.Long = "Show a Workflow Execution's Event History.\nWhen using JSON output (`--output json`), you may pass the results to an SDK\nto perform a replay:\n\n```\ntemporal workflow show \\\n    --workflow-id YourWorkflowId\n    --output json\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVarP(&s.Follow, "follow", "f", false, "Follow the Workflow Execution progress in real time. Does not apply to JSON output.")
	s.Command.Flags().BoolVar(&s.Detailed, "detailed", false, "Display events as detailed sections instead of table. Does not apply to JSON output.")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowSignalCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SingleWorkflowOrBatchOptions
	PayloadInputOptions
	Name string
}

func NewTemporalWorkflowSignalCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowSignalCommand {
	var s TemporalWorkflowSignalCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "signal [flags]"
	s.Command.Short = "Send a message to a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Send an asynchronous notification (Signal) to a running Workflow Execution by\nits Workflow ID. The Signal is written to the History. When you include\n\x1b[1m--input\x1b[0m, that data is available for the Workflow Execution to consume:\n\n\x1b[1mtemporal workflow signal \\\n    --workflow-id YourWorkflowId \\\n    --name YourSignal \\\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference."
	} else {
		s.Command.Long = "Send an asynchronous notification (Signal) to a running Workflow Execution by\nits Workflow ID. The Signal is written to the History. When you include\n`--input`, that data is available for the Workflow Execution to consume:\n\n```\ntemporal workflow signal \\\n    --workflow-id YourWorkflowId \\\n    --name YourSignal \\\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Signal name. Required. Aliased as \"--type\".")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"type": "name",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowSignalWithStartCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SharedWorkflowStartOptions
	WorkflowStartOptions
	PayloadInputOptions
	SignalName        string
	SignalInput       []string
	SignalInputFile   []string
	SignalInputMeta   []string
	SignalInputBase64 bool
}

func NewTemporalWorkflowSignalWithStartCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowSignalWithStartCommand {
	var s TemporalWorkflowSignalWithStartCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "signal-with-start [flags]"
	s.Command.Short = "Send a message to a Workflow Execution, start the execution if it isn't running"
	if hasHighlighting {
		s.Command.Long = "Send an asynchronous notification (Signal) to a Workflow Execution.\nIf the Workflow Execution is not running or is not found, it starts the\nworkflow then sends the signal.\n\n\x1b[1mtemporal workflow signal-with-start \\\n  --signal-name YourSignal \\\n  --signal-input '{\"some-key\": \"some-value\"}' \\\n  --workflow-id YourWorkflowId \\\n  --type YourWorkflowType \\\n  --task-queue YourTaskQueue \\\n  --input '{\"some-key\": \"some-value\"}'\x1b[0m"
	} else {
		s.Command.Long = "Send an asynchronous notification (Signal) to a Workflow Execution.\nIf the Workflow Execution is not running or is not found, it starts the\nworkflow then sends the signal.\n\n```\ntemporal workflow signal-with-start \\\n  --signal-name YourSignal \\\n  --signal-input '{\"some-key\": \"some-value\"}' \\\n  --workflow-id YourWorkflowId \\\n  --type YourWorkflowType \\\n  --task-queue YourTaskQueue \\\n  --input '{\"some-key\": \"some-value\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.SignalName, "signal-name", "", "Signal name. Required. Aliased as \"--signal-type\".")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "signal-name")
	s.Command.Flags().StringArrayVar(&s.SignalInput, "signal-input", nil, "Signal input value. Use JSON content or set --signal-input-meta to override. Can't be combined with --signal-input-file. Can be passed multiple times to pass multiple arguments.")
	s.Command.Flags().StringArrayVar(&s.SignalInputFile, "signal-input-file", nil, "A path or paths for input file(s). Use JSON content or set --signal-input-meta to override. Can't be combined with --signal-input. Can be passed multiple times to pass multiple arguments.")
	s.Command.Flags().StringArrayVar(&s.SignalInputMeta, "signal-input-meta", nil, "Input signal payload metadata as a `KEY=VALUE` pair. When the KEY is \"encoding\", this overrides the default (\"json/plain\"). Can be passed multiple times.")
	s.Command.Flags().BoolVar(&s.SignalInputBase64, "signal-input-base64", false, "Assume signal inputs are base64-encoded and attempt to decode them.")
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"name":        "type",
		"signal-type": "signal-name",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowStackCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowReferenceOptions
	RejectCondition StringEnum
}

func NewTemporalWorkflowStackCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowStackCommand {
	var s TemporalWorkflowStackCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "stack [flags]"
	s.Command.Short = "Trace a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Perform a Query on a Workflow Execution using a \x1b[1m__stack_trace\x1b[0m-type Query.\nDisplay a stack trace of the threads and routines currently in use by the\nWorkflow for troubleshooting:\n\n\x1b[1mtemporal workflow stack \\\n    --workflow-id YourWorkflowId\x1b[0m"
	} else {
		s.Command.Long = "Perform a Query on a Workflow Execution using a `__stack_trace`-type Query.\nDisplay a stack trace of the threads and routines currently in use by the\nWorkflow for troubleshooting:\n\n```\ntemporal workflow stack \\\n    --workflow-id YourWorkflowId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.RejectCondition = NewStringEnum([]string{"not_open", "not_completed_cleanly"}, "")
	s.Command.Flags().Var(&s.RejectCondition, "reject-condition", "Optional flag to reject Queries based on Workflow state. Accepted values: not_open, not_completed_cleanly.")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowStartCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SharedWorkflowStartOptions
	WorkflowStartOptions
	PayloadInputOptions
}

func NewTemporalWorkflowStartCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowStartCommand {
	var s TemporalWorkflowStartCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "start [flags]"
	s.Command.Short = "Initiate a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Start a new Workflow Execution. Returns the Workflow- and Run-IDs:\n\n\x1b[1mtemporal workflow start \\\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\x1b[0m"
	} else {
		s.Command.Long = "Start a new Workflow Execution. Returns the Workflow- and Run-IDs:\n\n```\ntemporal workflow start \\\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"name": "type",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowStartUpdateWithStartCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SharedWorkflowStartOptions
	WorkflowStartOptions
	PayloadInputOptions
	UpdateName                string
	UpdateFirstExecutionRunId string
	UpdateWaitForStage        StringEnum
	UpdateId                  string
	RunId                     string
	UpdateInput               []string
	UpdateInputFile           []string
	UpdateInputMeta           []string
	UpdateInputBase64         bool
}

func NewTemporalWorkflowStartUpdateWithStartCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowStartUpdateWithStartCommand {
	var s TemporalWorkflowStartUpdateWithStartCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "start-update-with-start [flags]"
	s.Command.Short = "Send an Update-With-Start and wait for it to be accepted or rejected (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to be accepted or rejected. If the Workflow Execution is not running,\nthen a new workflow execution is started and the update is sent.\n\nExperimental.\n\n\x1b[1mtemporal workflow start-update-with-start \\\n  --update-name YourUpdate \\\n  --update-input '{\"update-key\": \"update-value\"}' \\\n  --update-wait-for-stage accepted \\\n  --workflow-id YourWorkflowId \\\n  --type YourWorkflowType \\\n  --task-queue YourTaskQueue \\\n  --id-conflict-policy Fail \\\n  --input '{\"wf-key\": \"wf-value\"}'\x1b[0m"
	} else {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to be accepted or rejected. If the Workflow Execution is not running,\nthen a new workflow execution is started and the update is sent.\n\nExperimental.\n\n```\ntemporal workflow start-update-with-start \\\n  --update-name YourUpdate \\\n  --update-input '{\"update-key\": \"update-value\"}' \\\n  --update-wait-for-stage accepted \\\n  --workflow-id YourWorkflowId \\\n  --type YourWorkflowType \\\n  --task-queue YourTaskQueue \\\n  --id-conflict-policy Fail \\\n  --input '{\"wf-key\": \"wf-value\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.UpdateName, "update-name", "", "Update name. Required. Aliased as \"--update-type\".")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "update-name")
	s.Command.Flags().StringVar(&s.UpdateFirstExecutionRunId, "update-first-execution-run-id", "", "Parent Run ID. The update is sent to the last Workflow Execution in the chain started with this Run ID.")
	s.UpdateWaitForStage = NewStringEnum([]string{"accepted"}, "")
	s.Command.Flags().Var(&s.UpdateWaitForStage, "update-wait-for-stage", "Update stage to wait for. The only option is `accepted`, but this option is required. This is to allow a future version of the CLI to choose a default value. Accepted values: accepted. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "update-wait-for-stage")
	s.Command.Flags().StringVar(&s.UpdateId, "update-id", "", "Update ID. If unset, defaults to a UUID.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID. If unset, looks for an Update against the currently-running Workflow Execution.")
	s.Command.Flags().StringArrayVar(&s.UpdateInput, "update-input", nil, "Update input value. Use JSON content or set --update-input-meta to override. Can't be combined with --update-input-file. Can be passed multiple times to pass multiple arguments.")
	s.Command.Flags().StringArrayVar(&s.UpdateInputFile, "update-input-file", nil, "A path or paths for input file(s). Use JSON content or set --update-input-meta to override. Can't be combined with --update-input. Can be passed multiple times to pass multiple arguments.")
	s.Command.Flags().StringArrayVar(&s.UpdateInputMeta, "update-input-meta", nil, "Input update payload metadata as a `KEY=VALUE` pair. When the KEY is \"encoding\", this overrides the default (\"json/plain\"). Can be passed multiple times.")
	s.Command.Flags().BoolVar(&s.UpdateInputBase64, "update-input-base64", false, "Assume update inputs are base64-encoded and attempt to decode them.")
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"name":        "type",
		"update-type": "update-name",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowTerminateCommand struct {
	Parent     *TemporalWorkflowCommand
	Command    cobra.Command
	WorkflowId string
	Query      string
	RunId      string
	Reason     string
	Yes        bool
	Rps        float32
}

func NewTemporalWorkflowTerminateCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowTerminateCommand {
	var s TemporalWorkflowTerminateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "terminate [flags]"
	s.Command.Short = "Forcefully end a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Terminate a Workflow Execution:\n\n\x1b[1mtemporal workflow terminate \\\n    --reason YourReasonForTermination \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nThe reason is optional and defaults to the current user's name. The reason\nis stored in the Event History as part of the \x1b[1mWorkflowExecutionTerminated\x1b[0m\nevent. This becomes the closing Event in the Workflow Execution's history.\n\nExecutions may be terminated in bulk via a visibility Query list filter:\n\n\x1b[1mtemporal workflow terminate \\\n    --query YourQuery \\\n    --reason YourReasonForTermination\x1b[0m\n\nWorkflow code cannot see or respond to terminations. To perform clean-up work\nin your Workflow code, use \x1b[1mtemporal workflow cancel\x1b[0m instead.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference."
	} else {
		s.Command.Long = "Terminate a Workflow Execution:\n\n```\ntemporal workflow terminate \\\n    --reason YourReasonForTermination \\\n    --workflow-id YourWorkflowId\n```\n\nThe reason is optional and defaults to the current user's name. The reason\nis stored in the Event History as part of the `WorkflowExecutionTerminated`\nevent. This becomes the closing Event in the Workflow Execution's history.\n\nExecutions may be terminated in bulk via a visibility Query list filter:\n\n```\ntemporal workflow terminate \\\n    --query YourQuery \\\n    --reason YourReasonForTermination\n```\n\nWorkflow code cannot see or respond to terminations. To perform clean-up work\nin your Workflow code, use `temporal workflow cancel` instead.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow ID. You must set either --workflow-id or --query.")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter. You must set either --workflow-id or --query.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID. Can only be set with --workflow-id. Do not use with --query.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for termination. Defaults to message with the current user's name.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm termination. Can only be used with --query.")
	s.Command.Flags().Float32Var(&s.Rps, "rps", 0, "Limit batch's requests per second. Only allowed if query is present.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowTraceCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowReferenceOptions
	Fold        []string
	NoFold      bool
	Depth       int
	Concurrency int
}

func NewTemporalWorkflowTraceCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowTraceCommand {
	var s TemporalWorkflowTraceCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "trace [flags]"
	s.Command.Short = "Workflow Execution live progress"
	if hasHighlighting {
		s.Command.Long = "Display the progress of a Workflow Execution and its child workflows with a\nreal-time trace. This view helps you understand how Workflows are proceeding:\n\n\x1b[1mtemporal workflow trace \\\n    --workflow-id YourWorkflowId\x1b[0m"
	} else {
		s.Command.Long = "Display the progress of a Workflow Execution and its child workflows with a\nreal-time trace. This view helps you understand how Workflows are proceeding:\n\n```\ntemporal workflow trace \\\n    --workflow-id YourWorkflowId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Fold, "fold", nil, "Fold away Child Workflows with the specified statuses. Case-insensitive. Ignored if --no-fold supplied. Available values: running, completed, failed, canceled, terminated, timedout, continueasnew. Can be passed multiple times.")
	s.Command.Flags().BoolVar(&s.NoFold, "no-fold", false, "Disable folding. Fetch and display Child Workflows within the set depth.")
	s.Command.Flags().IntVar(&s.Depth, "depth", -1, "Set depth for your Child Workflow fetches. Pass -1 to fetch child workflows at any depth.")
	s.Command.Flags().IntVar(&s.Concurrency, "concurrency", 10, "Number of Workflow Histories to fetch at a time.")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowUpdateCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowUpdateCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowUpdateCommand {
	var s TemporalWorkflowUpdateCommand
	s.Parent = parent
	s.Command.Use = "update"
	s.Command.Short = "Send and interact with Updates"
	s.Command.Long = "An Update is a synchronous call to a Workflow Execution that can change its\nstate, control its flow, and return a result."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkflowUpdateDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowUpdateExecuteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowUpdateResultCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowUpdateStartCommand(cctx, &s).Command)
	return &s
}

type TemporalWorkflowUpdateDescribeCommand struct {
	Parent  *TemporalWorkflowUpdateCommand
	Command cobra.Command
	UpdateTargetingOptions
}

func NewTemporalWorkflowUpdateDescribeCommand(cctx *CommandContext, parent *TemporalWorkflowUpdateCommand) *TemporalWorkflowUpdateDescribeCommand {
	var s TemporalWorkflowUpdateDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Obtain status info about a specific Update"
	if hasHighlighting {
		s.Command.Long = "Given a Workflow Execution and an Update ID, return information about its current status, including\na result if it has finished.\n\n\x1b[1mtemporal workflow update describe \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\x1b[0m"
	} else {
		s.Command.Long = "Given a Workflow Execution and an Update ID, return information about its current status, including\na result if it has finished.\n\n```\ntemporal workflow update describe \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.UpdateTargetingOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowUpdateExecuteCommand struct {
	Parent  *TemporalWorkflowUpdateCommand
	Command cobra.Command
	UpdateStartingOptions
	PayloadInputOptions
}

func NewTemporalWorkflowUpdateExecuteCommand(cctx *CommandContext, parent *TemporalWorkflowUpdateCommand) *TemporalWorkflowUpdateExecuteCommand {
	var s TemporalWorkflowUpdateExecuteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "execute [flags]"
	s.Command.Short = "Send an Update and wait for it to complete"
	if hasHighlighting {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to complete or fail. You can also use this to wait for an existing\nupdate to complete, by submitting an existing update ID.\n\n\x1b[1mtemporal workflow update execute \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\x1b[0m"
	} else {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to complete or fail. You can also use this to wait for an existing\nupdate to complete, by submitting an existing update ID.\n\n```\ntemporal workflow update execute \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.UpdateStartingOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"type": "name",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowUpdateResultCommand struct {
	Parent  *TemporalWorkflowUpdateCommand
	Command cobra.Command
	UpdateTargetingOptions
}

func NewTemporalWorkflowUpdateResultCommand(cctx *CommandContext, parent *TemporalWorkflowUpdateCommand) *TemporalWorkflowUpdateResultCommand {
	var s TemporalWorkflowUpdateResultCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "result [flags]"
	s.Command.Short = "Wait for a specific Update to complete"
	if hasHighlighting {
		s.Command.Long = "Given a Workflow Execution and an Update ID, wait for the Update to complete or fail and\nprint the result.\n\n\x1b[1mtemporal workflow update result \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\x1b[0m"
	} else {
		s.Command.Long = "Given a Workflow Execution and an Update ID, wait for the Update to complete or fail and\nprint the result.\n\n```\ntemporal workflow update result \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.UpdateTargetingOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowUpdateStartCommand struct {
	Parent  *TemporalWorkflowUpdateCommand
	Command cobra.Command
	UpdateStartingOptions
	PayloadInputOptions
	WaitForStage StringEnum
}

func NewTemporalWorkflowUpdateStartCommand(cctx *CommandContext, parent *TemporalWorkflowUpdateCommand) *TemporalWorkflowUpdateStartCommand {
	var s TemporalWorkflowUpdateStartCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "start [flags]"
	s.Command.Short = "Send an Update and wait for it to be accepted or rejected"
	if hasHighlighting {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to be accepted or rejected. You can subsequently wait for the update\nto complete by using \x1b[1mtemporal workflow update execute\x1b[0m.\n\n\x1b[1mtemporal workflow update start \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\n    --wait-for-stage accepted\x1b[0m"
	} else {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to be accepted or rejected. You can subsequently wait for the update\nto complete by using `temporal workflow update execute`.\n\n```\ntemporal workflow update start \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\n    --wait-for-stage accepted\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.WaitForStage = NewStringEnum([]string{"accepted"}, "")
	s.Command.Flags().Var(&s.WaitForStage, "wait-for-stage", "Update stage to wait for. The only option is `accepted`, but this option is  required. This is to allow a future version of the CLI to choose a default value. Accepted values: accepted. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "wait-for-stage")
	s.UpdateStartingOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().SetNormalizeFunc(aliasNormalizer(map[string]string{
		"type": "name",
	}))
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowUpdateOptionsCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	SingleWorkflowOrBatchOptions
	VersioningOverrideBehavior       StringEnum
	VersioningOverrideDeploymentName string
	VersioningOverrideBuildId        string
}

func NewTemporalWorkflowUpdateOptionsCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowUpdateOptionsCommand {
	var s TemporalWorkflowUpdateOptionsCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update-options [flags]"
	s.Command.Short = "Change Workflow Execution Options"
	if hasHighlighting {
		s.Command.Long = "\x1b[1m+---------------------------------------------------------------------+\n| CAUTION: Worflow update-options is experimental. Workflow Execution |\n| properties are subject to change.                                   |\n+---------------------------------------------------------------------+\x1b[0m\n\nModify properties of Workflow Executions:\n\n\x1b[1mtemporal workflow update-options [options]\x1b[0m\n\nIt can override the Worker Deployment configuration of a\nWorkflow Execution, which controls Worker Versioning.\n\nFor example, to force Workers in the current Deployment execute the\nnext Workflow Task change behavior to \x1b[1mauto_upgrade\x1b[0m:\n\n\x1b[1mtemporal workflow update-options \\\n    --workflow-id YourWorkflowId \\\n    --versioning-override-behavior auto_upgrade\x1b[0m\n\nor to pin the workflow execution to a Worker Deployment, set behavior\nto \x1b[1mpinned\x1b[0m:\n\n\x1b[1mtemporal workflow update-options \\\n    --workflow-id YourWorkflowId \\\n    --versioning-override-behavior pinned \\\n    --versioning-override-deployment-name YourDeploymentName \\\n    --versioning-override-build-id YourDeploymentBuildId\x1b[0m\n\nTo remove any previous overrides, set the behavior to\n\x1b[1munspecified\x1b[0m:\n\n\x1b[1mtemporal workflow update-options \\\n    --workflow-id YourWorkflowId \\\n    --versioning-override-behavior unspecified\x1b[0m\n\nTo see the current override use \x1b[1mtemporal workflow describe\x1b[0m"
	} else {
		s.Command.Long = "```\n+---------------------------------------------------------------------+\n| CAUTION: Worflow update-options is experimental. Workflow Execution |\n| properties are subject to change.                                   |\n+---------------------------------------------------------------------+\n```\n\nModify properties of Workflow Executions:\n\n```\ntemporal workflow update-options [options]\n```\n\nIt can override the Worker Deployment configuration of a\nWorkflow Execution, which controls Worker Versioning.\n\nFor example, to force Workers in the current Deployment execute the\nnext Workflow Task change behavior to `auto_upgrade`:\n\n```\ntemporal workflow update-options \\\n    --workflow-id YourWorkflowId \\\n    --versioning-override-behavior auto_upgrade\n```\n\nor to pin the workflow execution to a Worker Deployment, set behavior\nto `pinned`:\n\n```\ntemporal workflow update-options \\\n    --workflow-id YourWorkflowId \\\n    --versioning-override-behavior pinned \\\n    --versioning-override-deployment-name YourDeploymentName \\\n    --versioning-override-build-id YourDeploymentBuildId\n```\n\nTo remove any previous overrides, set the behavior to\n`unspecified`:\n\n```\ntemporal workflow update-options \\\n    --workflow-id YourWorkflowId \\\n    --versioning-override-behavior unspecified\n```\n\nTo see the current override use `temporal workflow describe`"
	}
	s.Command.Args = cobra.NoArgs
	s.VersioningOverrideBehavior = NewStringEnum([]string{"unspecified", "pinned", "auto_upgrade"}, "")
	s.Command.Flags().Var(&s.VersioningOverrideBehavior, "versioning-override-behavior", "Override the versioning behavior of a Workflow. Accepted values: unspecified, pinned, auto_upgrade. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "versioning-override-behavior")
	s.Command.Flags().StringVar(&s.VersioningOverrideDeploymentName, "versioning-override-deployment-name", "", "When overriding to a `pinned` behavior, specifies the Deployment Name of the version to target.")
	s.Command.Flags().StringVar(&s.VersioningOverrideBuildId, "versioning-override-build-id", "", "When overriding to a `pinned` behavior, specifies the Build ID of the version to target.")
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

var reDays = regexp.MustCompile(`(\d+(\.\d*)?|(\.\d+))d`)

type Duration time.Duration

// ParseDuration is like time.ParseDuration, but supports unit "d" for days
// (always interpreted as exactly 24 hours).
func ParseDuration(s string) (time.Duration, error) {
	s = reDays.ReplaceAllStringFunc(s, func(v string) string {
		fv, err := strconv.ParseFloat(strings.TrimSuffix(v, "d"), 64)
		if err != nil {
			return v // will cause time.ParseDuration to return an error
		}
		return fmt.Sprintf("%fh", 24*fv)
	})
	return time.ParseDuration(s)
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d *Duration) String() string {
	return d.Duration().String()
}

func (d *Duration) Set(s string) error {
	p, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(p)
	return nil
}

func (d *Duration) Type() string {
	return "duration"
}

type StringEnum struct {
	Allowed            []string
	Value              string
	ChangedFromDefault bool
}

func NewStringEnum(allowed []string, value string) StringEnum {
	return StringEnum{Allowed: allowed, Value: value}
}

func (s *StringEnum) String() string { return s.Value }

func (s *StringEnum) Set(p string) error {
	for _, allowed := range s.Allowed {
		if p == allowed {
			s.Value = p
			s.ChangedFromDefault = true
			return nil
		}
	}
	return fmt.Errorf("%v is not one of required values of %v", p, strings.Join(s.Allowed, ", "))
}

func (*StringEnum) Type() string { return "string" }

type StringEnumArray struct {
	Allowed map[string]string
	Values  []string
}

func NewStringEnumArray(allowed []string, values []string) StringEnumArray {
	var allowedMap = make(map[string]string)
	for _, str := range allowed {
		allowedMap[strings.ToLower(str)] = str
	}
	return StringEnumArray{Allowed: allowedMap, Values: values}
}

func (s *StringEnumArray) String() string { return strings.Join(s.Values, ",") }

func (s *StringEnumArray) Set(p string) error {
	val, ok := s.Allowed[strings.ToLower(p)]
	if !ok {
		values := make([]string, 0, len(s.Allowed))
		for _, v := range s.Allowed {
			values = append(values, v)
		}
		return fmt.Errorf("invalid value: %s, allowed values are: %s", p, strings.Join(values, ", "))
	}
	s.Values = append(s.Values, val)
	return nil
}

func (*StringEnumArray) Type() string { return "string" }

type Timestamp time.Time

func (t Timestamp) Time() time.Time {
	return time.Time(t)
}

func (t *Timestamp) String() string {
	return t.Time().Format(time.RFC3339)
}

func (t *Timestamp) Set(s string) error {
	p, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = Timestamp(p)
	return nil
}

func (t *Timestamp) Type() string {
	return "timestamp"
}
