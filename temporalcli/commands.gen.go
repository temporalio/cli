// Code generated. DO NOT EDIT.

package temporalcli

import (
	"github.com/mattn/go-isatty"

	"github.com/spf13/cobra"

	"github.com/spf13/pflag"

	"os"

	"time"
)

var hasHighlighting = isatty.IsTerminal(os.Stdout.Fd())

type ClientOptions struct {
	Address                    string
	Namespace                  string
	ApiKey                     string
	GrpcMeta                   []string
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
}

func (v *ClientOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Address, "address", "127.0.0.1:7233", "Temporal Service gRPC endpoint.")
	cctx.BindFlagEnvVar(f.Lookup("address"), "TEMPORAL_ADDRESS")
	f.StringVarP(&v.Namespace, "namespace", "n", "default", "Temporal Service Namespace.")
	cctx.BindFlagEnvVar(f.Lookup("namespace"), "TEMPORAL_NAMESPACE")
	f.StringVar(&v.ApiKey, "api-key", "", "API key for request.")
	cctx.BindFlagEnvVar(f.Lookup("api-key"), "TEMPORAL_API_KEY")
	f.StringArrayVar(&v.GrpcMeta, "grpc-meta", nil, "HTTP headers for requests.\nformat as a `KEY=VALUE` pair\nMay be passed multiple times to set multiple headers.\n")
	f.BoolVar(&v.Tls, "tls", false, "Enable base TLS encryption.\nDoes not have additional options like mTLS or client certs.\n")
	cctx.BindFlagEnvVar(f.Lookup("tls"), "TEMPORAL_TLS")
	f.StringVar(&v.TlsCertPath, "tls-cert-path", "", "Path to x509 certificate.\nCan't be used with --tls-cert-data.\n")
	cctx.BindFlagEnvVar(f.Lookup("tls-cert-path"), "TEMPORAL_TLS_CERT")
	f.StringVar(&v.TlsCertData, "tls-cert-data", "", "Data for x509 certificate.\nCan't be used with --tls-cert-path.\n")
	cctx.BindFlagEnvVar(f.Lookup("tls-cert-data"), "TEMPORAL_TLS_CERT_DATA")
	f.StringVar(&v.TlsKeyPath, "tls-key-path", "", "Path to x509 private key.\nCan't be used with --tls-key-data.\n")
	cctx.BindFlagEnvVar(f.Lookup("tls-key-path"), "TEMPORAL_TLS_KEY")
	f.StringVar(&v.TlsKeyData, "tls-key-data", "", "Private certificate key data.\nCan't be used with --tls-key-path.\n")
	cctx.BindFlagEnvVar(f.Lookup("tls-key-data"), "TEMPORAL_TLS_KEY_DATA")
	f.StringVar(&v.TlsCaPath, "tls-ca-path", "", "Path to server CA certificate.\nCan't be used with --tls-ca-data.\n")
	cctx.BindFlagEnvVar(f.Lookup("tls-ca-path"), "TEMPORAL_TLS_CA")
	f.StringVar(&v.TlsCaData, "tls-ca-data", "", "Data for server CA certificate.\nCan't be used with --tls-ca-path.\n")
	cctx.BindFlagEnvVar(f.Lookup("tls-ca-data"), "TEMPORAL_TLS_CA_DATA")
	f.BoolVar(&v.TlsDisableHostVerification, "tls-disable-host-verification", false, "Disable TLS host-name verification.")
	cctx.BindFlagEnvVar(f.Lookup("tls-disable-host-verification"), "TEMPORAL_TLS_DISABLE_HOST_VERIFICATION")
	f.StringVar(&v.TlsServerName, "tls-server-name", "", "Override target TLS server name.")
	cctx.BindFlagEnvVar(f.Lookup("tls-server-name"), "TEMPORAL_TLS_SERVER_NAME")
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "Remote Codec Server endpoint.")
	cctx.BindFlagEnvVar(f.Lookup("codec-endpoint"), "TEMPORAL_CODEC_ENDPOINT")
	f.StringVar(&v.CodecAuth, "codec-auth", "", "Authorization header for Codec Server requests.")
	cctx.BindFlagEnvVar(f.Lookup("codec-auth"), "TEMPORAL_CODEC_AUTH")
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
	f.StringArrayVar(&v.Calendar, "calendar", nil, "Calendar specification in JSON.\nFor example: `{\"dayOfWeek\":\"Fri\",\"hour\":\"17\",\"minute\":\"5\"}`.\n")
	v.CatchupWindow = 0
	f.Var(&v.CatchupWindow, "catchup-window", "Maximum catch-up time for when the Service is unavailable.")
	f.StringArrayVar(&v.Cron, "cron", nil, "Calendar specification in cron string format.\nFor example: `\"30 12 * * Fri\"`.\n")
	f.Var(&v.EndTime, "end-time", "Schedule end time.")
	f.StringArrayVar(&v.Interval, "interval", nil, "Interval duration.\nFor example, 90m, or 60m/15m to include phase offset.\n")
	v.Jitter = 0
	f.Var(&v.Jitter, "jitter", "Max difference in time from the specification.\nVary the start time randomly within this amount.\n")
	f.StringVar(&v.Notes, "notes", "", "Initial notes field value.")
	f.BoolVar(&v.Paused, "paused", false, "Pause the Schedule immediately on creation.")
	f.BoolVar(&v.PauseOnFailure, "pause-on-failure", false, "Pause schedule after Workflow failures.")
	f.IntVar(&v.RemainingActions, "remaining-actions", 0, "Total allowed actions.\nDefault is zero (unlimited).\n")
	f.Var(&v.StartTime, "start-time", "Schedule start time.")
	f.StringVar(&v.TimeZone, "time-zone", "", "Interpret calendar specs with the `TZ` time zone.\nFor a list of time zones, see:\nhttps://en.wikipedia.org/wiki/List_of_tz_database_time_zones.\n")
	f.StringArrayVar(&v.ScheduleSearchAttribute, "schedule-search-attribute", nil, "Set schedule Search Attributes using `KEY=\"VALUE` pairs.\nKeys must be identifiers, and values must be JSON values.\nFor example: 'YourKey={\"your\": \"value\"}'.\nCan be passed multiple times.\n")
	f.StringArrayVar(&v.ScheduleMemo, "schedule-memo", nil, "Set schedule memo using `KEY=\"VALUE` pairs.\nKeys must be identifiers, and values must be JSON values.\nFor example: 'YourKey={\"your\": \"value\"}'.\nCan be passed multiple times.\n")
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

type SingleWorkflowOrBatchOptions struct {
	WorkflowId string
	Query      string
	RunId      string
	Reason     string
	Yes        bool
}

func (v *SingleWorkflowOrBatchOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID.\nYou must set either --workflow-id or --query.\n")
	f.StringVarP(&v.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.\nYou must set either --workflow-id or --query.\n")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID.\nOnly use with --workflow-id.\nCannot use with --query.\n")
	f.StringVar(&v.Reason, "reason", "", "Reason for batch operation.\nOnly use with --query.\nDefaults to user name.\n")
	f.BoolVarP(&v.Yes, "yes", "y", false, "Don't prompt to confirm signaling.\nOnly allowed when --query is present.\n")
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
}

func (v *SharedWorkflowStartOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID.\nIf not supplied, the Service generates a unique ID.\n")
	f.StringVar(&v.Type, "type", "", "Workflow Type name. Required. Aliased as \"--name\".")
	_ = cobra.MarkFlagRequired(f, "type")
	f.StringVarP(&v.TaskQueue, "task-queue", "t", "", "Workflow Task queue. Required.")
	_ = cobra.MarkFlagRequired(f, "task-queue")
	v.RunTimeout = 0
	f.Var(&v.RunTimeout, "run-timeout", "Fail a Workflow Run if it lasts longer than `DURATION`.\n")
	v.ExecutionTimeout = 0
	f.Var(&v.ExecutionTimeout, "execution-timeout", "Fail a WorkflowExecution if it lasts longer than `DURATION`.\nThis time-out includes retries and ContinueAsNew tasks.\n")
	v.TaskTimeout = Duration(10000 * time.Millisecond)
	f.Var(&v.TaskTimeout, "task-timeout", "Fail a Workflow Task if it lasts longer than `DURATION`.\nThis is the Start-to-close timeout for a Workflow Task.\n")
	f.StringArrayVar(&v.SearchAttribute, "search-attribute", nil, "Search Attribute in `KEY=VALUE` format.\nKeys must be identifiers, and values must be JSON values.\nFor example: 'YourKey={\"your\": \"value\"}'.\nCan be passed multiple times.\n")
	f.StringArrayVar(&v.Memo, "memo", nil, "Memo using 'KEY=\"VALUE\"' pairs.\nUse JSON values.\n")
}

type WorkflowStartOptions struct {
	Cron          string
	FailExisting  bool
	StartDelay    Duration
	IdReusePolicy StringEnum
}

func (v *WorkflowStartOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Cron, "cron", "", "Cron schedule for the Workflow.\nDeprecated.\nUse Schedules instead.\n")
	f.BoolVar(&v.FailExisting, "fail-existing", false, "Fail if the Workflow already exists.")
	v.StartDelay = 0
	f.Var(&v.StartDelay, "start-delay", "Delay before starting the Workflow Execution.\nCan't be used with cron schedules.\nIf the Workflow receives a signal or update prior to this time, the Workflow\nExecution starts immediately.\n")
	v.IdReusePolicy = NewStringEnum([]string{"AllowDuplicate", "AllowDuplicateFailedOnly", "RejectDuplicate", "TerminateIfRunning"}, "")
	f.Var(&v.IdReusePolicy, "id-reuse-policy", "Re-use policy for the Workflow ID in new Workflow Executions. Accepted values: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.")
}

type PayloadInputOptions struct {
	Input       []string
	InputFile   []string
	InputMeta   []string
	InputBase64 bool
}

func (v *PayloadInputOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringArrayVarP(&v.Input, "input", "i", nil, "Input value.\nUse JSON content or set --input-meta to override.\nCan't be combined with --input-file.\nCan be passed multiple times to pass multiple arguments.\n")
	f.StringArrayVar(&v.InputFile, "input-file", nil, "A path or paths for input file(s).\nUse JSON content or set --input-meta to override.\nCan't be combined with --input.\nCan be passed multiple times to pass multiple arguments.\n")
	f.StringArrayVar(&v.InputMeta, "input-meta", nil, "Input payload metadata as a `KEY=VALUE` pair.\nWhen the KEY is \"encoding\", this overrides the default (\"json/plain\").\nCan be passed multiple times.\n")
	f.BoolVar(&v.InputBase64, "input-base64", false, "Assume inputs are base64-encoded and attempt to decode them.\n")
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
	f.StringVar(&v.FirstExecutionRunId, "first-execution-run-id", "", "Parent Run ID.\nThe update is sent to the last Workflow Execution in the chain started\nwith this Run ID.\n")
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. Required.")
	_ = cobra.MarkFlagRequired(f, "workflow-id")
	f.StringVar(&v.UpdateId, "update-id", "", "Update ID.\nIf unset, defaults to a UUID.\n")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID.\nIf unset, looks for an Update against the currently-running Workflow Execution.\n")
}

type UpdateTargetingOptions struct {
	WorkflowId string
	UpdateId   string
	RunId      string
}

func (v *UpdateTargetingOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. Required.")
	_ = cobra.MarkFlagRequired(f, "workflow-id")
	f.StringVar(&v.UpdateId, "update-id", "", "Update ID.\nMust be unique per Workflow Execution.\n Required.")
	_ = cobra.MarkFlagRequired(f, "update-id")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID.\nIf unset, updates the currently-running Workflow Execution.\n")
}

type TemporalCommand struct {
	Command                 cobra.Command
	Env                     string
	EnvFile                 string
	LogLevel                StringEnum
	LogFormat               StringEnum
	Output                  StringEnum
	TimeFormat              StringEnum
	Color                   StringEnum
	NoJsonShorthandPayloads bool
}

func NewTemporalCommand(cctx *CommandContext) *TemporalCommand {
	var s TemporalCommand
	s.Command.Use = "temporal"
	s.Command.Short = "Temporal command-line interface and development server"
	if hasHighlighting {
		s.Command.Long = "The Temporal CLI manages, monitors, and debugs Temporal apps. It lets you run\na local Temporal Service, start Workflow Executions, pass messages to running\nWorkflows, inspect state, and more.\n\n* Start a local development service:\n      \x1b[1mtemporal server start-dev\x1b[0m\n* View help: pass \x1b[1m--help\x1b[0m to any command:\n      \x1b[1mtemporal activity complete --help\x1b[0m\n"
	} else {
		s.Command.Long = "The Temporal CLI manages, monitors, and debugs Temporal apps. It lets you run\na local Temporal Service, start Workflow Executions, pass messages to running\nWorkflows, inspect state, and more.\n\n* Start a local development service:\n      `temporal server start-dev`\n* View help: pass `--help` to any command:\n      `temporal activity complete --help`\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalActivityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalBatchCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalServerCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVar(&s.Env, "env", "default", "Active environment name (`ENV`).")
	cctx.BindFlagEnvVar(s.Command.PersistentFlags().Lookup("env"), "TEMPORAL_ENV")
	s.Command.PersistentFlags().StringVar(&s.EnvFile, "env-file", "", "Path to environment settings file.\n(defaults to `$HOME/.config/temporalio/temporal.yaml`).\n")
	s.LogLevel = NewStringEnum([]string{"debug", "info", "warn", "error", "never"}, "info")
	s.Command.PersistentFlags().Var(&s.LogLevel, "log-level", "Log level.\nDefault is \"info\" for most commands and \"warn\" for `server start-dev`.\n Accepted values: debug, info, warn, error, never.")
	s.LogFormat = NewStringEnum([]string{"text", "json", "pretty"}, "text")
	s.Command.PersistentFlags().Var(&s.LogFormat, "log-format", "Log format.\n\"pretty\" is an alias for \"text\". This is for compatibility reasons, please use \"text\".\n Accepted values: text, json, pretty.")
	s.Output = NewStringEnum([]string{"text", "json", "jsonl", "none"}, "text")
	s.Command.PersistentFlags().VarP(&s.Output, "output", "o", "Non-logging data output format. Accepted values: text, json, jsonl, none.")
	s.TimeFormat = NewStringEnum([]string{"relative", "iso", "raw"}, "relative")
	s.Command.PersistentFlags().Var(&s.TimeFormat, "time-format", "Time format. Accepted values: relative, iso, raw.")
	s.Color = NewStringEnum([]string{"always", "never", "auto"}, "auto")
	s.Command.PersistentFlags().Var(&s.Color, "color", "Output coloring. Accepted values: always, never, auto.")
	s.Command.PersistentFlags().BoolVar(&s.NoJsonShorthandPayloads, "no-json-shorthand-payloads", false, "Raw payload output, even if they are JSON.")
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
	s.Command.Short = "Complete or fail an Activity"
	if hasHighlighting {
		s.Command.Long = "Update an Activity's state to completed or failed. This marks an Activity\nas successfully finished or as having encountered an error:\n\n\x1b[1mtemporal activity complete \\\n    --activity-id=YourActivityId \\\n    --workflow-id=YourWorkflowId \\\n    --result='{\"YourResultKey\": \"YourResultValue\"}'\x1b[0m\n"
	} else {
		s.Command.Long = "Update an Activity's state to completed or failed. This marks an Activity\nas successfully finished or as having encountered an error:\n\n```\ntemporal activity complete \\\n    --activity-id=YourActivityId \\\n    --workflow-id=YourWorkflowId \\\n    --result='{\"YourResultKey\": \"YourResultValue\"}'\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalActivityCompleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalActivityFailCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
	return &s
}

type TemporalActivityCompleteCommand struct {
	Parent  *TemporalActivityCommand
	Command cobra.Command
	WorkflowReferenceOptions
	ActivityId string
	Result     string
	Identity   string
}

func NewTemporalActivityCompleteCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityCompleteCommand {
	var s TemporalActivityCompleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "complete [flags]"
	s.Command.Short = "Complete an Activity"
	if hasHighlighting {
		s.Command.Long = "Complete an Activity, marking it as successfully finished. Specify the\nActivity ID and include a JSON result for the returned value:\n\n\x1b[1mtemporal activity complete \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --result '{\"YourResultKey\": \"YourResultVal\"}'\x1b[0m\n"
	} else {
		s.Command.Long = "Complete an Activity, marking it as successfully finished. Specify the\nActivity ID and include a JSON result for the returned value:\n\n```\ntemporal activity complete \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --result '{\"YourResultKey\": \"YourResultVal\"}'\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "Activity ID to complete. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Result, "result", "", "Result `JSON` to return. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "result")
	s.Command.Flags().StringVar(&s.Identity, "identity", "", "Identity of the user submitting this request.")
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
	Identity   string
	Reason     string
}

func NewTemporalActivityFailCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityFailCommand {
	var s TemporalActivityFailCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "fail [flags]"
	s.Command.Short = "Fail an Activity"
	if hasHighlighting {
		s.Command.Long = "Fail an Activity, marking it as having encountered an error. Specify the\nActivity and Workflow IDs:\n\n\x1b[1mtemporal activity fail \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\x1b[0m\n"
	} else {
		s.Command.Long = "Fail an Activity, marking it as having encountered an error. Specify the\nActivity and Workflow IDs:\n\n```\ntemporal activity fail \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "Activity ID to fail. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Detail, "detail", "", "Reason for failing the Activity (JSON).")
	s.Command.Flags().StringVar(&s.Identity, "identity", "", "Identity of the user submitting this request.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for failing the Activity.")
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
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
		s.Command.Long = "List or terminate running batch jobs.\n\nA batch job executes a command on multiple Workflow Executions at once. Create\nbatch jobs by passing \x1b[1m--query\x1b[0m to commands that support it. For example, to\ncreate a batch job to cancel a set of Workflow Executions:\n\n\x1b[1mtemporal workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\x1b[0m\n\nQuery Quick Reference:\n\n\x1b[1m+----------------------------------------------------------------------------+\n| Composition:                                                               |\n| - Data types: String literals with single or double quotes,                |\n|   Numbers (integer and floating point), Booleans                           |\n| - Comparison: '=', '!=', '>', '>=', '<', '<='                              |\n| - Expressions/Operators:  'IN array', 'BETWEEN value AND value',           |\n|   'STARTS_WITH string', 'IS NULL', 'IS NOT NULL', 'expr AND expr',         |\n|   'expr OR expr', '( expr )'                                               |\n| - Array: '( comma-separated-values )'                                      |\n|                                                                            |\n| Please note:                                                               |\n| - Wrap attributes with backticks if it contains characters not in          |\n|   \x1b[1m[a-zA-Z0-9]\x1b[0m.                                                           |\n| - \x1b[1mSTARTS_WITH\x1b[0m is only available for Keyword search attributes.           |\n+----------------------------------------------------------------------------+\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation.\n"
	} else {
		s.Command.Long = "List or terminate running batch jobs.\n\nA batch job executes a command on multiple Workflow Executions at once. Create\nbatch jobs by passing `--query` to commands that support it. For example, to\ncreate a batch job to cancel a set of Workflow Executions:\n\n```\ntemporal workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\n```\n\nQuery Quick Reference:\n\n```\n+----------------------------------------------------------------------------+\n| Composition:                                                               |\n| - Data types: String literals with single or double quotes,                |\n|   Numbers (integer and floating point), Booleans                           |\n| - Comparison: '=', '!=', '>', '>=', '<', '<='                              |\n| - Expressions/Operators:  'IN array', 'BETWEEN value AND value',           |\n|   'STARTS_WITH string', 'IS NULL', 'IS NOT NULL', 'expr AND expr',         |\n|   'expr OR expr', '( expr )'                                               |\n| - Array: '( comma-separated-values )'                                      |\n|                                                                            |\n| Please note:                                                               |\n| - Wrap attributes with backticks if it contains characters not in          |\n|   `[a-zA-Z0-9]`.                                                           |\n| - `STARTS_WITH` is only available for Keyword search attributes.           |\n+----------------------------------------------------------------------------+\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation.\n"
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
		s.Command.Long = "Show the progress of an ongoing batch job. Pass a valid job ID to display its\ninformation:\n\n\x1b[1mtemporal batch describe \\\n    --job-id YourJobId\x1b[0m\n"
	} else {
		s.Command.Long = "Show the progress of an ongoing batch job. Pass a valid job ID to display its\ninformation:\n\n```\ntemporal batch describe \\\n    --job-id YourJobId\n```\n"
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
		s.Command.Long = "Return a list of batch jobs on the Service or within a single Namespace. For\nexample, list the batch jobs for \"YourNamespace\":\n\n\x1b[1mtemporal batch list \\\n    --namespace YourNamespace\x1b[0m\n"
	} else {
		s.Command.Long = "Return a list of batch jobs on the Service or within a single Namespace. For\nexample, list the batch jobs for \"YourNamespace\":\n\n```\ntemporal batch list \\\n    --namespace YourNamespace\n```\n"
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
		s.Command.Long = "Terminate a batch job with the provided job ID. You must provide a reason for\nthe termination. The Service stores this explanation as metadata for the\ntermination event for later reference:\n\n\x1b[1mtemporal batch terminate \\\n    --job-id YourJobId \\\n    --reason YourTerminationReason\x1b[0m\n"
	} else {
		s.Command.Long = "Terminate a batch job with the provided job ID. You must provide a reason for\nthe termination. The Service stores this explanation as metadata for the\ntermination event for later reference:\n\n```\ntemporal batch terminate \\\n    --job-id YourJobId \\\n    --reason YourTerminationReason\n```\n"
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
		s.Command.Long = "Environments manage key-value presets, auto-configuring Temporal CLI options\nfor you. You can set up distinct environments like \"dev\" and \"prod\" for\nconvenience:\n\n\x1b[1mtemporal env set \\\n    --env prod \\\n    --key address \\\n    --value production.f45a2.tmprl.cloud:7233\x1b[0m\n\nEach environment is isolated. Changes to \"prod\" presets won't affect \"dev\".\n\nFor easiest use, set a \x1b[1mTEMPORAL_ENV\x1b[0m environment variable in your shell. The\nTemporal CLI checks for an \x1b[1m--env\x1b[0m option first, then checks for the\n\x1b[1mTEMPORAL_ENV\x1b[0m environment variable. If neither is set, the CLI uses the\n\"default\" environment.\n"
	} else {
		s.Command.Long = "Environments manage key-value presets, auto-configuring Temporal CLI options\nfor you. You can set up distinct environments like \"dev\" and \"prod\" for\nconvenience:\n\n```\ntemporal env set \\\n    --env prod \\\n    --key address \\\n    --value production.f45a2.tmprl.cloud:7233\n```\n\nEach environment is isolated. Changes to \"prod\" presets won't affect \"dev\".\n\nFor easiest use, set a `TEMPORAL_ENV` environment variable in your shell. The\nTemporal CLI checks for an `--env` option first, then checks for the\n`TEMPORAL_ENV` environment variable. If neither is set, the CLI uses the\n\"default\" environment.\n"
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
		s.Command.Long = "Remove a presets environment entirely _or_ remove a key-value pair within an\nenvironment. If you don't specify an environment (with \x1b[1m--env\x1b[0m or by setting\nthe \x1b[1mTEMPORAL_ENV\x1b[0m variable), this command updates the \"default\" environment:\n\n\x1b[1mtemporal env delete \\\n    --env YourEnvironment\x1b[0m\n\nor\n\n\x1b[1mtemporal env delete \\\n    --env prod \\\n    --key tls-key-path\x1b[0m\n"
	} else {
		s.Command.Long = "Remove a presets environment entirely _or_ remove a key-value pair within an\nenvironment. If you don't specify an environment (with `--env` or by setting\nthe `TEMPORAL_ENV` variable), this command updates the \"default\" environment:\n\n```\ntemporal env delete \\\n    --env YourEnvironment\n```\n\nor\n\n```\ntemporal env delete \\\n    --env prod \\\n    --key tls-key-path\n```\n"
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
		s.Command.Long = "List the properties for a given environment:\n\n\x1b[1mtemporal env get \\\n    --env YourEnvironment\x1b[0m\n\nPrint a single property:\n\n\x1b[1mtemporal env get \\\n    --env YourEnvironment \\\n    --key YourPropertyKey\x1b[0m\n\nIf you don't specify an environment (with \x1b[1m--env\x1b[0m or by setting the\n\x1b[1mTEMPORAL_ENV\x1b[0m variable), this command lists properties of the \"default\"\nenvironment.\n"
	} else {
		s.Command.Long = "List the properties for a given environment:\n\n```\ntemporal env get \\\n    --env YourEnvironment\n```\n\nPrint a single property:\n\n```\ntemporal env get \\\n    --env YourEnvironment \\\n    --key YourPropertyKey\n```\n\nIf you don't specify an environment (with `--env` or by setting the\n`TEMPORAL_ENV` variable), this command lists properties of the \"default\"\nenvironment.\n"
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
	s.Command.Long = "List the environments you have set up on your local computer. Environments are\nstored in \"$HOME/.config/temporalio/temporal.yaml\".\n"
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
		s.Command.Long = "Assign a value to a property key and store it to an environment:\n\n\x1b[1mtemporal env set \\\n    --env environment \\\n    --key property \\\n    --value value\x1b[0m\n\nIf you don't specify an environment (with \x1b[1m--env\x1b[0m or by setting the\n\x1b[1mTEMPORAL_ENV\x1b[0m variable), this command sets properties in the \"default\"\nenvironment.\n\nStoring keys with CLI option names lets the CLI automatically set those\noptions for you. This reduces effort and helps avoid typos when issuing\ncommands.\n"
	} else {
		s.Command.Long = "Assign a value to a property key and store it to an environment:\n\n```\ntemporal env set \\\n    --env environment \\\n    --key property \\\n    --value value\n```\n\nIf you don't specify an environment (with `--env` or by setting the\n`TEMPORAL_ENV` variable), this command sets properties in the \"default\"\nenvironment.\n\nStoring keys with CLI option names lets the CLI automatically set those\noptions for you. This reduces effort and helps avoid typos when issuing\ncommands.\n"
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
		s.Command.Long = "Operator commands manage and fetch information about Namespaces, Search\nAttributes, Nexus Endpoints, and Temporal Services:\n\n\x1b[1mtemporal operator [command] [subcommand] [options]\x1b[0m\n\nFor example, to show information about the Temporal Service at the default\naddress (localhost):\n\n\x1b[1mtemporal operator cluster describe\x1b[0m\n"
	} else {
		s.Command.Long = "Operator commands manage and fetch information about Namespaces, Search\nAttributes, Nexus Endpoints, and Temporal Services:\n\n```\ntemporal operator [command] [subcommand] [options]\n```\n\nFor example, to show information about the Temporal Service at the default\naddress (localhost):\n\n```\ntemporal operator cluster describe\n```\n"
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
		s.Command.Long = "Perform operator actions on Temporal Services (also known as Clusters).\n\n\x1b[1mtemporal operator cluster [subcommand] [options]\x1b[0m\n\nFor example to check Service/Cluster health:\n\n\x1b[1mtemporal operator cluster health\x1b[0m\n"
	} else {
		s.Command.Long = "Perform operator actions on Temporal Services (also known as Clusters).\n\n```\ntemporal operator cluster [subcommand] [options]\n```\n\nFor example to check Service/Cluster health:\n\n```\ntemporal operator cluster health\n```\n"
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
		s.Command.Long = "View information about a Temporal Cluster (Service), including Cluster Name,\npersistence store, and visibility store. Add \x1b[1m--detail\x1b[0m for additional info:\n\n\x1b[1mtemporal operator cluster describe [--detail]\x1b[0m\n"
	} else {
		s.Command.Long = "View information about a Temporal Cluster (Service), including Cluster Name,\npersistence store, and visibility store. Add `--detail` for additional info:\n\n```\ntemporal operator cluster describe [--detail]\n```\n"
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
		s.Command.Long = "View information about the health of a Temporal Service:\n\n\x1b[1mtemporal operator cluster health\x1b[0m\n"
	} else {
		s.Command.Long = "View information about the health of a Temporal Service:\n\n```\ntemporal operator cluster health\n```\n"
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
		s.Command.Long = "Print a list of remote Temporal Clusters (Services) registered to the local\nService. Report details include the Cluster's name, ID, address, History Shard\ncount, Failover version, and availability:\n\n\x1b[1mtemporal operator cluster list [--limit max-count]\x1b[0m\n"
	} else {
		s.Command.Long = "Print a list of remote Temporal Clusters (Services) registered to the local\nService. Report details include the Cluster's name, ID, address, History Shard\ncount, Failover version, and availability:\n\n```\ntemporal operator cluster list [--limit max-count]\n```\n"
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
		s.Command.Long = "Remove a registered remote Temporal Cluster (Service) from the local Service.\n\n\x1b[1mtemporal operator cluster remove \\\n    --name YourClusterName\x1b[0m\n"
	} else {
		s.Command.Long = "Remove a registered remote Temporal Cluster (Service) from the local Service.\n\n```\ntemporal operator cluster remove \\\n    --name YourClusterName\n```\n"
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
		s.Command.Long = "Show Temporal Server information for Temporal Clusters (Service): Server\nversion, scheduling support, and more. This information helps diagnose\nproblems with the Temporal Server.\n\nThe command defaults to the local Service. Otherwise, use the\n\x1b[1m--frontend-address\x1b[0m option to specify a Cluster (Service) endpoint:\n\n\x1b[1mtemporal operator cluster system \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\x1b[0m\n"
	} else {
		s.Command.Long = "Show Temporal Server information for Temporal Clusters (Service): Server\nversion, scheduling support, and more. This information helps diagnose\nproblems with the Temporal Server.\n\nThe command defaults to the local Service. Otherwise, use the\n`--frontend-address` option to specify a Cluster (Service) endpoint:\n\n```\ntemporal operator cluster system \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\n```\n"
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
		s.Command.Long = "Add, remove, or update a registered (\"remote\") Temporal Cluster (Service).\n\n\x1b[1mtemporal operator cluster upsert [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator cluster upsert \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\n    --enable-connection false\x1b[0m\n"
	} else {
		s.Command.Long = "Add, remove, or update a registered (\"remote\") Temporal Cluster (Service).\n\n```\ntemporal operator cluster upsert [options]\n```\n\nFor example:\n\n```\ntemporal operator cluster upsert \\\n    --frontend-address \"YourRemoteEndpoint:YourRemotePort\"\n    --enable-connection false\n```\n"
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
		s.Command.Long = "Manage Temporal Cluster (Service) Namespaces:\n\n\x1b[1mtemporal operator namespace [command] [command options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator namespace create \\\n    --namespace YourNewNamespaceName\x1b[0m\n"
	} else {
		s.Command.Long = "Manage Temporal Cluster (Service) Namespaces:\n\n```\ntemporal operator namespace [command] [command options]\n```\n\nFor example:\n\n```\ntemporal operator namespace create \\\n    --namespace YourNewNamespaceName\n```\n"
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
		s.Command.Long = "Create a new Namespace on the Temporal Service:\n\n\x1b[1mtemporal operator namespace create \\\n    --namespace YourNewNamespaceName \\\n    [options]\x1b[0m`\n\nCreate a Namespace with multi-region data replication:\n\n\x1b[1mtemporal operator namespace create \\\n    --global \\\n    --namespace YourNewNamespaceName\x1b[0m\n\nConfigure settings like retention and Visibility Archival State as needed.\nFor example, the Visibility Archive can be set on a separate URI:\n\n\x1b[1mtemporal operator namespace create \\\n    --retention 5d \\\n    --visibility-archival-state enabled \\\n    --visibility-uri YourURI \\\n    --namespace YourNewNamespaceName\x1b[0m\n\nNote: URI values for archival states can't be changed once enabled.\n"
	} else {
		s.Command.Long = "Create a new Namespace on the Temporal Service:\n\n```\ntemporal operator namespace create \\\n    --namespace YourNewNamespaceName \\\n    [options]\n````\n\nCreate a Namespace with multi-region data replication:\n\n```\ntemporal operator namespace create \\\n    --global \\\n    --namespace YourNewNamespaceName\n```\n\nConfigure settings like retention and Visibility Archival State as needed.\nFor example, the Visibility Archive can be set on a separate URI:\n\n```\ntemporal operator namespace create \\\n    --retention 5d \\\n    --visibility-archival-state enabled \\\n    --visibility-uri YourURI \\\n    --namespace YourNewNamespaceName\n```\n\nNote: URI values for archival states can't be changed once enabled.\n"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active Cluster (Service) name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster (Service) names for Namespace creation.\nCan be passed multiple times.\n")
	s.Command.Flags().StringArrayVar(&s.Data, "data", nil, "Namespace data as `KEY=VALUE` pairs.\nKeys must be identifiers, and values must be JSON values.\nFor example: 'YourKey={\"your\": \"value\"}'.\nCan be passed multiple times.\n")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.Global, "global", false, "Enable multi-region data replication.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disabled")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Archive history to this `URI`.\nOnce enabled, can't be changed.\n")
	s.Retention = Duration(259200000 * time.Millisecond)
	s.Command.Flags().Var(&s.Retention, "retention", "Time to preserve closed Workflows before deletion.")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disabled")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Archive visibility data to this `URI`.\nOnce enabled, can't be changed.\n")
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
		s.Command.Long = "Removes a Namespace from the Service.\n\n\x1b[1mtemporal operator namespace delete [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator namespace delete \\\n    --namespace YourNamespaceName\x1b[0m\n"
	} else {
		s.Command.Long = "Removes a Namespace from the Service.\n\n```\ntemporal operator namespace delete [options]\n```\n\nFor example:\n\n```\ntemporal operator namespace delete \\\n    --namespace YourNamespaceName\n```\n"
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
		s.Command.Long = "Provide long-form information about a Namespace identified by its ID or name:\n\n\x1b[1mtemporal operator namespace describe \\\n    --namespace-id YourNamespaceId\x1b[0m\n\nor\n\n\x1b[1mtemporal operator namespace describe \\\n    --namespace YourNamespaceName\x1b[0m\n"
	} else {
		s.Command.Long = "Provide long-form information about a Namespace identified by its ID or name:\n\n```\ntemporal operator namespace describe \\\n    --namespace-id YourNamespaceId\n```\n\nor\n\n```\ntemporal operator namespace describe \\\n    --namespace YourNamespaceName\n```\n"
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
		s.Command.Long = "Display a detailed listing for all Namespaces on the Service:\n\n\x1b[1mtemporal operator namespace list\x1b[0m\n"
	} else {
		s.Command.Long = "Display a detailed listing for all Namespaces on the Service:\n\n```\ntemporal operator namespace list\n```\n"
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
		s.Command.Long = "Update a Namespace using properties you specify.\n\n\x1b[1mtemporal operator namespace update [options]\x1b[0m\n\nAssign a Namespace's active Cluster (Service):\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --active-cluster NewActiveCluster\x1b[0m\n\nPromote a Namespace for multi-region data replication:\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --promote-global\x1b[0m\n\nYou may update archives that were previously enabled or disabled. Note: URI\nvalues for archival states can't be changed once enabled.\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --history-archival-state enabled \\\n    --visibility-archival-state disabled\x1b[0m\n"
	} else {
		s.Command.Long = "Update a Namespace using properties you specify.\n\n```\ntemporal operator namespace update [options]\n```\n\nAssign a Namespace's active Cluster (Service):\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --active-cluster NewActiveCluster\n```\n\nPromote a Namespace for multi-region data replication:\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --promote-global\n```\n\nYou may update archives that were previously enabled or disabled. Note: URI\nvalues for archival states can't be changed once enabled.\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --history-archival-state enabled \\\n    --visibility-archival-state disabled\n```\n"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active Cluster (Service) name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster (Service) names.")
	s.Command.Flags().StringArrayVar(&s.Data, "data", nil, "Namespace data as `KEY=VALUE` pairs.\nKeys must be identifiers, and values must be JSON values.\nFor example: 'YourKey={\"your\": \"value\"}'.\nCan be passed multiple times.\n")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.PromoteGlobal, "promote-global", false, "Enable multi-region data replication.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Archive history to this `URI`.\nOnce enabled, can't be changed.\n")
	s.Retention = 0
	s.Command.Flags().Var(&s.Retention, "retention", "Length of time a closed Workflow is preserved before deletion.")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Archive visibility data to this `URI`.\nOnce enabled, can't be changed.\n")
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
	s.Command.Short = "Commands for managing Nexus resources (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Nexus commands enable managing Nexus resources.\n\nNexus commands follow this syntax:\n\n\x1b[1mtemporal operator nexus [command] [command] [command options]\x1b[0m\n"
	} else {
		s.Command.Long = "Nexus commands enable managing Nexus resources.\n\nNexus commands follow this syntax:\n\n```\ntemporal operator nexus [command] [command] [command options]\n```\n"
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
	s.Command.Short = "Commands for managing Nexus Endpoints (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Endpoint commands enable managing Nexus Endpoints.\n\nEndpoint commands follow this syntax:\n\n\x1b[1mtemporal operator nexus endpoint [command] [command options]\x1b[0m\n"
	} else {
		s.Command.Long = "Endpoint commands enable managing Nexus Endpoints.\n\nEndpoint commands follow this syntax:\n\n```\ntemporal operator nexus endpoint [command] [command options]\n```\n"
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
	Parent          *TemporalOperatorNexusEndpointCommand
	Command         cobra.Command
	Name            string
	Description     string
	DescriptionFile string
	TargetNamespace string
	TargetTaskQueue string
	TargetUrl       string
}

func NewTemporalOperatorNexusEndpointCreateCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointCreateCommand {
	var s TemporalOperatorNexusEndpointCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Create a new Nexus Endpoint (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Create a new Nexus Endpoint on the Server.\n\nAn endpoint name is used in workflow code to invoke Nexus operations.  The\nendpoint target may either be a worker, in which case \x1b[1m--target-namespace\x1b[0m and\n\x1b[1m--target-task-queue\x1b[0m must both be provided, or an external URL, in which case\n\x1b[1m--target-url\x1b[0m must be provided.\n\nThis command will fail if an endpoint with the same name is already registered.\n\n\x1b[1mtemporal operator nexus endpoint create \\\n  --name your-endpoint \\\n  --target-namespace your-namespace \\\n  --target-task-queue your-task-queue \\\n  --description-file DESCRIPTION.md\x1b[0m\n"
	} else {
		s.Command.Long = "Create a new Nexus Endpoint on the Server.\n\nAn endpoint name is used in workflow code to invoke Nexus operations.  The\nendpoint target may either be a worker, in which case `--target-namespace` and\n`--target-task-queue` must both be provided, or an external URL, in which case\n`--target-url` must be provided.\n\nThis command will fail if an endpoint with the same name is already registered.\n\n```\ntemporal operator nexus endpoint create \\\n  --name your-endpoint \\\n  --target-namespace your-namespace \\\n  --target-task-queue your-task-queue \\\n  --description-file DESCRIPTION.md\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Endpoint name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Endpoint description in markdown format (encoded using the configured codec server).\n")
	s.Command.Flags().StringVar(&s.DescriptionFile, "description-file", "", "Endpoint description file in markdown format (encoded using the configured codec server).\n")
	s.Command.Flags().StringVar(&s.TargetNamespace, "target-namespace", "", "Namespace in which a handler worker will be polling for Nexus tasks on.")
	s.Command.Flags().StringVar(&s.TargetTaskQueue, "target-task-queue", "", "Task Queue in which a handler worker will be polling for Nexus tasks on.")
	s.Command.Flags().StringVar(&s.TargetUrl, "target-url", "", "URL to direct Nexus requests to.")
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
	Name    string
}

func NewTemporalOperatorNexusEndpointDeleteCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointDeleteCommand {
	var s TemporalOperatorNexusEndpointDeleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete [flags]"
	s.Command.Short = "Delete a Nexus Endpoint (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Delete a Nexus Endpoint configuration from the Server.\n\n\x1b[1mtemporal operator nexus endpoint delete --name your-endpoint\x1b[0m\n"
	} else {
		s.Command.Long = "Delete a Nexus Endpoint configuration from the Server.\n\n```\ntemporal operator nexus endpoint delete --name your-endpoint\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Endpoint name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
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
	Name    string
}

func NewTemporalOperatorNexusEndpointGetCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointGetCommand {
	var s TemporalOperatorNexusEndpointGetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "get [flags]"
	s.Command.Short = "Get a Nexus Endpoint by name (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Get a Nexus Endpoint configuration by name from the Server.\n\n\x1b[1mtemporal operator nexus endpoint get --name your-endpoint\x1b[0m\n"
	} else {
		s.Command.Long = "Get a Nexus Endpoint configuration by name from the Server.\n\n```\ntemporal operator nexus endpoint get --name your-endpoint\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Endpoint name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
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
	s.Command.Short = "List Nexus Endpoints (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "List all Nexus Endpoint configurations on the Server.\n\n\x1b[1mtemporal operator nexus endpoint list\x1b[0m\n"
	} else {
		s.Command.Long = "List all Nexus Endpoint configurations on the Server.\n\n```\ntemporal operator nexus endpoint list\n```\n"
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
	Parent           *TemporalOperatorNexusEndpointCommand
	Command          cobra.Command
	Name             string
	Description      string
	DescriptionFile  string
	UnsetDescription bool
	TargetNamespace  string
	TargetTaskQueue  string
	TargetUrl        string
}

func NewTemporalOperatorNexusEndpointUpdateCommand(cctx *CommandContext, parent *TemporalOperatorNexusEndpointCommand) *TemporalOperatorNexusEndpointUpdateCommand {
	var s TemporalOperatorNexusEndpointUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update an existing Nexus Endpoint (EXPERIMENTAL)"
	if hasHighlighting {
		s.Command.Long = "Update an existing Nexus Endpoint on the Server.\n\nAn endpoint name is used in workflow code to invoke Nexus operations.  The\nendpoint target may either be a worker, in which case \x1b[1m--target-namespace\x1b[0m and\n\x1b[1m--target-task-queue\x1b[0m must both be provided, or an external URL, in which case\n\x1b[1m--target-url\x1b[0m must be provided.\n\nThe endpoint is patched; existing fields for which flags are not provided are\nleft as they were.\n\nUpdate only the target task queue:\n\n\x1b[1mtemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --target-task-queue your-other-queue\x1b[0m\n\nUpdate only the description:\n\n\x1b[1mtemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --description-file DESCRIPTION.md\x1b[0m\n"
	} else {
		s.Command.Long = "Update an existing Nexus Endpoint on the Server.\n\nAn endpoint name is used in workflow code to invoke Nexus operations.  The\nendpoint target may either be a worker, in which case `--target-namespace` and\n`--target-task-queue` must both be provided, or an external URL, in which case\n`--target-url` must be provided.\n\nThe endpoint is patched; existing fields for which flags are not provided are\nleft as they were.\n\nUpdate only the target task queue:\n\n```\ntemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --target-task-queue your-other-queue\n```\n\nUpdate only the description:\n\n```\ntemporal operator nexus endpoint update \\\n  --name your-endpoint \\\n  --description-file DESCRIPTION.md\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Endpoint name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Endpoint description in markdown format (encoded using the configured codec server).\n")
	s.Command.Flags().StringVar(&s.DescriptionFile, "description-file", "", "Endpoint description file in markdown format (encoded using the configured codec server).\n")
	s.Command.Flags().BoolVar(&s.UnsetDescription, "unset-description", false, "Unset the description.")
	s.Command.Flags().StringVar(&s.TargetNamespace, "target-namespace", "", "Namespace in which a handler worker will be polling for Nexus tasks on.")
	s.Command.Flags().StringVar(&s.TargetTaskQueue, "target-task-queue", "", "Task Queue in which a handler worker will be polling for Nexus tasks on.")
	s.Command.Flags().StringVar(&s.TargetUrl, "target-url", "", "URL to direct Nexus requests to.")
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
		s.Command.Long = "Create, list, or remove Search Attributes fields stored in a Workflow\nExecution's metadata:\n\n\x1b[1mtemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\x1b[0m\n\nSupported types include: Text, Keyword, Int, Double, Bool, Datetime, and\nKeywordList.\n"
	} else {
		s.Command.Long = "Create, list, or remove Search Attributes fields stored in a Workflow\nExecution's metadata:\n\n```\ntemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\n```\n\nSupported types include: Text, Keyword, Int, Double, Bool, Datetime, and\nKeywordList.\n"
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
		s.Command.Long = "Add one or more custom Search Attributes:\n\n\x1b[1mtemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\x1b[0m\n"
	} else {
		s.Command.Long = "Add one or more custom Search Attributes:\n\n```\ntemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Name, "name", nil, "Search Attribute name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Type = NewStringEnumArray([]string{"Text", "Keyword", "Int", "Double", "Bool", "Datetime", "KeywordList"}, []string{})
	s.Command.Flags().Var(&s.Type, "type", "Search Attribute type.\n Accepted values: Text, Keyword, Int, Double, Bool, Datetime, KeywordList. Required.")
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
		s.Command.Long = "Display a list of active Search Attributes that can be assigned or used with\nWorkflow Queries. You can manage this list and add attributes as needed:\n\n\x1b[1mtemporal operator search-attribute list\x1b[0m\n"
	} else {
		s.Command.Long = "Display a list of active Search Attributes that can be assigned or used with\nWorkflow Queries. You can manage this list and add attributes as needed:\n\n```\ntemporal operator search-attribute list\n```\n"
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
		s.Command.Long = "Remove custom Search Attributes from the options that can be assigned or used\nwith Workflow Queries.\n\n\x1b[1mtemporal operator search-attribute remove \\\n    --name YourAttributeName\x1b[0m\n\nRemove attributes without confirmation:\n\n\x1b[1mtemporal operator search-attribute remove \\\n    --name YourAttributeName \\\n    --yes\x1b[0m\n"
	} else {
		s.Command.Long = "Remove custom Search Attributes from the options that can be assigned or used\nwith Workflow Queries.\n\n```\ntemporal operator search-attribute remove \\\n    --name YourAttributeName\n```\n\nRemove attributes without confirmation:\n\n```\ntemporal operator search-attribute remove \\\n    --name YourAttributeName \\\n    --yes\n```\n"
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
		s.Command.Long = "Create, use, and update Schedules that allow Workflow Executions to be created\nat specified times:\n\n\x1b[1mtemporal schedule [commands] [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal schedule describe \\\n    --schedule-id \"YourScheduleId\"\x1b[0m\n"
	} else {
		s.Command.Long = "Create, use, and update Schedules that allow Workflow Executions to be created\nat specified times:\n\n```\ntemporal schedule [commands] [options]\n```\n\nFor example:\n\n```\ntemporal schedule describe \\\n    --schedule-id \"YourScheduleId\"\n```\n"
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
		s.Command.Long = "Batch-execute actions that would have run during a specified time interval.\nUse this command to fill in Workflow runs from when a Schedule was paused,\nbefore a Schedule was created, from the future, or to re-process a previously\nexecuted interval.\n\nBackfills require a Schedule ID and the time period covered by the request.\nIt's best to use the \x1b[1mBufferAll\x1b[0m or \x1b[1mAllowAll\x1b[0m policies to avoid conflicts\nand ensure no Workflow Executions are skipped.\n\nFor example:\n\n\x1b[1m  temporal schedule backfill \\\n    --schedule-id \"YourScheduleId\" \\\n    --start-time \"2022-05-01T00:00:00Z\" \\\n    --end-time \"2022-05-31T23:59:59Z\" \\\n    --overlap-policy BufferAll\x1b[0m\n\nThe policies include:\n\n* **AllowAll**: Allow unlimited concurrent Workflow Executions. This\n  significantly speeds up the backfilling process on systems that support\n  concurrency. You must ensure running Workflow Executions do not interfere\n  with each other.\n* **BufferAll**: Buffer all incoming Workflow Executions while waiting for\n  the running Workflow Execution to complete.\n* **Skip**: If a previous Workflow Execution is still running, discard new\n  Workflow Executions.\n* **BufferOne**: Same as 'Skip' but buffer a single Workflow Execution to be\n  run after the previous Execution completes. Discard other Workflow\n  Executions.\n* **CancelOther**: Cancel the running Workflow Execution and replace it with\n  the incoming new Workflow Execution.\n* **TerminateOther**: Terminate the running Workflow Execution and replace\n  it with the incoming new Workflow Execution.\n"
	} else {
		s.Command.Long = "Batch-execute actions that would have run during a specified time interval.\nUse this command to fill in Workflow runs from when a Schedule was paused,\nbefore a Schedule was created, from the future, or to re-process a previously\nexecuted interval.\n\nBackfills require a Schedule ID and the time period covered by the request.\nIt's best to use the `BufferAll` or `AllowAll` policies to avoid conflicts\nand ensure no Workflow Executions are skipped.\n\nFor example:\n\n```\n  temporal schedule backfill \\\n    --schedule-id \"YourScheduleId\" \\\n    --start-time \"2022-05-01T00:00:00Z\" \\\n    --end-time \"2022-05-31T23:59:59Z\" \\\n    --overlap-policy BufferAll\n```\n\nThe policies include:\n\n* **AllowAll**: Allow unlimited concurrent Workflow Executions. This\n  significantly speeds up the backfilling process on systems that support\n  concurrency. You must ensure running Workflow Executions do not interfere\n  with each other.\n* **BufferAll**: Buffer all incoming Workflow Executions while waiting for\n  the running Workflow Execution to complete.\n* **Skip**: If a previous Workflow Execution is still running, discard new\n  Workflow Executions.\n* **BufferOne**: Same as 'Skip' but buffer a single Workflow Execution to be\n  run after the previous Execution completes. Discard other Workflow\n  Executions.\n* **CancelOther**: Cancel the running Workflow Execution and replace it with\n  the incoming new Workflow Execution.\n* **TerminateOther**: Terminate the running Workflow Execution and replace\n  it with the incoming new Workflow Execution.\n"
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
		s.Command.Long = "Create a new Schedule on the Temporal Service. A Schedule automatically starts\nnew Workflow Executions at the times you specify.\n\nFor example:\n\n\x1b[1m  temporal schedule create \\\n    --schedule-id \"YourScheduleId\" \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"30\"}' \\\n    --workflow-id YourBaseWorkflowIdName \\\n    --task-queue YourTaskQueue \\\n    --type YourWorkflowType\x1b[0m\n\nSchedules support any combination of \x1b[1m--calendar\x1b[0m, \x1b[1m--interval\x1b[0m, and \x1b[1m--cron\x1b[0m:\n\n* Shorthand \x1b[1m--interval\x1b[0m strings.\n  For example: 45m (every 45 minutes) or 6h/5h (every 6 hours, at the top of\n  the 5th hour).\n* JSON \x1b[1m--calendar\x1b[0m, as in the preceding example.\n* Unix-style \x1b[1m--cron\x1b[0m strings and robfig declarations\n  (@daily/@weekly/@every X/etc).\n  For example, every Friday at 12:30 PM: \x1b[1m30 12 * * Fri\x1b[0m.\n"
	} else {
		s.Command.Long = "Create a new Schedule on the Temporal Service. A Schedule automatically starts\nnew Workflow Executions at the times you specify.\n\nFor example:\n\n```\n  temporal schedule create \\\n    --schedule-id \"YourScheduleId\" \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"30\"}' \\\n    --workflow-id YourBaseWorkflowIdName \\\n    --task-queue YourTaskQueue \\\n    --type YourWorkflowType\n```\n\nSchedules support any combination of `--calendar`, `--interval`, and `--cron`:\n\n* Shorthand `--interval` strings.\n  For example: 45m (every 45 minutes) or 6h/5h (every 6 hours, at the top of\n  the 5th hour).\n* JSON `--calendar`, as in the preceding example.\n* Unix-style `--cron` strings and robfig declarations\n  (@daily/@weekly/@every X/etc).\n  For example, every Friday at 12:30 PM: `30 12 * * Fri`.\n"
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
		s.Command.Long = "Deletes a Schedule on the front end Service:\n\n\x1b[1mtemporal schedule delete \\\n    --schedule-id YourScheduleId\x1b[0m\n\nRemoving a Schedule won't affect the Workflow Executions it started that are\nstill running. To cancel or terminate these Workflow Executions, use \x1b[1mtemporal\nworkflow delete\x1b[0m with the \x1b[1mTemporalScheduledById\x1b[0m Search Attribute instead.\n"
	} else {
		s.Command.Long = "Deletes a Schedule on the front end Service:\n\n```\ntemporal schedule delete \\\n    --schedule-id YourScheduleId\n```\n\nRemoving a Schedule won't affect the Workflow Executions it started that are\nstill running. To cancel or terminate these Workflow Executions, use `temporal\nworkflow delete` with the `TemporalScheduledById` Search Attribute instead.\n"
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
		s.Command.Long = "Show a Schedule configuration, including information about past, current, and\nfuture Workflow runs:\n\n\x1b[1mtemporal schedule describe \\\n    --schedule-id YourScheduleId\x1b[0m\n"
	} else {
		s.Command.Long = "Show a Schedule configuration, including information about past, current, and\nfuture Workflow runs:\n\n```\ntemporal schedule describe \\\n    --schedule-id YourScheduleId\n```\n"
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
		s.Command.Long = "Lists the Schedules hosted by a Namespace:\n\n\x1b[1mtemporal schedule list \\\n    --namespace YourNamespace\x1b[0m\n"
	} else {
		s.Command.Long = "Lists the Schedules hosted by a Namespace:\n\n```\ntemporal schedule list \\\n    --namespace YourNamespace\n```\n"
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
		s.Command.Long = "Pause or unpause a Schedule by passing a flag with your desired state:\n\n\x1b[1mtemporal schedule toggle \\\n    --schedule-id \"YourScheduleId\" \\\n    --pause \\\n    --reason \"YourReason\"\x1b[0m\n\nand\n\n\x1b[1mtemporal schedule toggle\n    --schedule-id \"YourScheduleId\" \\\n    --unpause \\\n    --reason \"YourReason\"\x1b[0m\n\nThe \x1b[1m--reason\x1b[0m text updates the Schedule's \x1b[1mnotes\x1b[0m field for operations\ncommunication. It defaults to \"(no reason provided)\" if omitted. This field is\nalso visible on the Service Web UI.\n"
	} else {
		s.Command.Long = "Pause or unpause a Schedule by passing a flag with your desired state:\n\n```\ntemporal schedule toggle \\\n    --schedule-id \"YourScheduleId\" \\\n    --pause \\\n    --reason \"YourReason\"\n```\n\nand\n\n```\ntemporal schedule toggle\n    --schedule-id \"YourScheduleId\" \\\n    --unpause \\\n    --reason \"YourReason\"\n```\n\nThe `--reason` text updates the Schedule's `notes` field for operations\ncommunication. It defaults to \"(no reason provided)\" if omitted. This field is\nalso visible on the Service Web UI.\n"
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
		s.Command.Long = "Trigger a Schedule to run immediately:\n\n\x1b[1mtemporal schedule trigger \\\n    --schedule-id \"YourScheduleId\"\x1b[0m\n"
	} else {
		s.Command.Long = "Trigger a Schedule to run immediately:\n\n```\ntemporal schedule trigger \\\n    --schedule-id \"YourScheduleId\"\n```\n"
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
		s.Command.Long = "Update an existing Schedule with new configuration details, including time\nspecifications, action, and policies:\n\n\x1b[1mtemporal schedule update \\\n--schedule-id \"YourScheduleId\" \\\n--workflow-type \"NewWorkflowType\"\x1b[0m\n"
	} else {
		s.Command.Long = "Update an existing Schedule with new configuration details, including time\nspecifications, action, and policies:\n\n```\ntemporal schedule update \\\n--schedule-id \"YourScheduleId\" \\\n--workflow-type \"NewWorkflowType\"\n```\n"
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
		s.Command.Long = "Run a development Temporal Server on your local system. View the Web UI for\nthe default configuration at http://localhost:8233:\n\n\x1b[1mtemporal server start-dev\x1b[0m\n\nAdd persistence for Workflow Executions across runs:\n\n\x1b[1mtemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\x1b[0m\n\nSet the port from the front-end gRPC Service (7233 default):\n\n\x1b[1mtemporal server start-dev \\\n    --port 7234 \\\n    --ui-port 8234 \\\n    --metrics-port 57271\x1b[0m\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n\x1b[1mtemporal server start-dev \\\n    --ui-port 3000\x1b[0m\n"
	} else {
		s.Command.Long = "Run a development Temporal Server on your local system. View the Web UI for\nthe default configuration at http://localhost:8233:\n\n```\ntemporal server start-dev\n```\n\nAdd persistence for Workflow Executions across runs:\n\n```\ntemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\n```\n\nSet the port from the front-end gRPC Service (7233 default):\n\n```\ntemporal server start-dev \\\n    --port 7234 \\\n    --ui-port 8234 \\\n    --metrics-port 57271\n```\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n```\ntemporal server start-dev \\\n    --ui-port 3000\n```\n"
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
		s.Command.Long = "Run a development Temporal Server on your local system. View the Web UI for\nthe default configuration at http://localhost:8233:\n\n\x1b[1mtemporal server start-dev\x1b[0m\n\nAdd persistence for Workflow Executions across runs:\n\n\x1b[1mtemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\x1b[0m\n\nSet the port from the front-end gRPC Service (7233 default):\n\n\x1b[1mtemporal server start-dev \\\n    --port 7000\x1b[0m\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n\x1b[1mtemporal server start-dev \\\n    --ui-port 3000\x1b[0m\n"
	} else {
		s.Command.Long = "Run a development Temporal Server on your local system. View the Web UI for\nthe default configuration at http://localhost:8233:\n\n```\ntemporal server start-dev\n```\n\nAdd persistence for Workflow Executions across runs:\n\n```\ntemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\n```\n\nSet the port from the front-end gRPC Service (7233 default):\n\n```\ntemporal server start-dev \\\n    --port 7000\n```\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n```\ntemporal server start-dev \\\n    --ui-port 3000\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.DbFilename, "db-filename", "f", "", "Path to file for persistent Temporal state store.\nBy default, Workflow Executions are lost when the server process dies.\n")
	s.Command.Flags().StringArrayVarP(&s.Namespace, "namespace", "n", nil, "Namespaces to be created at launch.\nThe \"default\" Namespace is always created automatically.\n")
	s.Command.Flags().IntVarP(&s.Port, "port", "p", 7233, "Port for the front-end gRPC Service.")
	s.Command.Flags().IntVar(&s.HttpPort, "http-port", 0, "Port for the HTTP API service.\nDefault is off.\n")
	s.Command.Flags().IntVar(&s.MetricsPort, "metrics-port", 0, "Port for '/metrics'.\nDefault is off.\n")
	s.Command.Flags().IntVar(&s.UiPort, "ui-port", 0, "Port for the Web UI.\nDefault is '--port' value + 1000.\n")
	s.Command.Flags().BoolVar(&s.Headless, "headless", false, "Disable the Web UI.")
	s.Command.Flags().StringVar(&s.Ip, "ip", "localhost", "IP address bound to the frontend Service.")
	s.Command.Flags().StringVar(&s.UiIp, "ui-ip", "", "IP address bound to the Web UI.\nDefault is same as '--ip' value.\n")
	s.Command.Flags().StringVar(&s.UiPublicPath, "ui-public-path", "", "The public base path for the Web UI.\nDefault is `/`.\n")
	s.Command.Flags().StringVar(&s.UiAssetPath, "ui-asset-path", "", "UI custom assets path.")
	s.Command.Flags().StringVar(&s.UiCodecEndpoint, "ui-codec-endpoint", "", "UI remote codec HTTP end.")
	s.Command.Flags().StringArrayVar(&s.SqlitePragma, "sqlite-pragma", nil, "SQLite pragma statements in \"PRAGMA=VALUE\" format.")
	s.Command.Flags().StringArrayVar(&s.DynamicConfigValue, "dynamic-config-value", nil, "Dynamic configuration value using `KEY=VALUE` pairs.\nKeys must be identifiers, and values must be JSON values.\nFor example: 'YourKey=\"YourString\"'.\nCan be passed multiple times.\n")
	s.Command.Flags().BoolVar(&s.LogConfig, "log-config", false, "Log the server config to stderr.")
	s.Command.Flags().StringArrayVar(&s.SearchAttribute, "search-attribute", nil, "Search attributes to register using `KEY=VALUE` pairs.\nKeys must be identifiers, and values must be the search\nattribute type, which is one of the following:\nText, Keyword, Int, Double, Bool, Datetime, KeywordList.\n")
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
		s.Command.Long = "Inspect and update Task Queues, the queues that Workers poll for Workflow and\nActivity tasks:\n\n\x1b[1mtemporal task-queue [command] [command options] \\\n    --task-queue YourTaskQueue\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal task-queue describe \\\n    --task-queue YourTaskQueue\x1b[0m\n"
	} else {
		s.Command.Long = "Inspect and update Task Queues, the queues that Workers poll for Workflow and\nActivity tasks:\n\n```\ntemporal task-queue [command] [command options] \\\n    --task-queue YourTaskQueue\n```\n\nFor example:\n\n```\ntemporal task-queue describe \\\n    --task-queue YourTaskQueue\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalTaskQueueDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueGetBuildIdReachabilityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueGetBuildIdsCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueListPartitionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueUpdateBuildIdsCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(cctx, s.Command.PersistentFlags())
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
}

func NewTemporalTaskQueueDescribeCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueDescribeCommand {
	var s TemporalTaskQueueDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Show active Workers"
	if hasHighlighting {
		s.Command.Long = "Display a list of active Workers that have recently polled a Task Queue. The\nTemporal Server records each poll request time. A \x1b[1mLastAccessTime\x1b[0m over one\nminute may indicate the Worker is at capacity or has shut down. Temporal\nWorkers are removed if 5 minutes have passed since the last poll request.\n\n\x1b[1mtemporal task-queue describe \\\n  --task-queue YourTaskQueue\x1b[0m\n\nThis command provides poller information for a given Task Queue.\nWorkflow and Activity polling use separate Task Queues:\n\n\x1b[1mtemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type \"activity\"\x1b[0m\n\nThis command provides the following task queue statistics:\n- \x1b[1mApproximateBacklogCount\x1b[0m: The approximate number of tasks backlogged in this\n  task queue. May count expired tasks but eventually converges to the right\n  value.\n- \x1b[1mApproximateBacklogAge\x1b[0m: Approximate age of the oldest task in the backlog,\n  based on its creation time, measured in seconds.\n- \x1b[1mTasksAddRate\x1b[0m: Approximate rate at which tasks are being added to the task\n  queue, measured in tasks per second, averaged over the last 30 seconds.\n  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- \x1b[1mTasksDispatchRate\x1b[0m: Approximate rate at which tasks are being dispatched from\n  the task queue, measured in tasks per second, averaged over the last 30\n  seconds.  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- \x1b[1mBacklog Increase Rate\x1b[0m: Approximate rate at which the backlog size is\n  increasing (if positive) or decreasing (if negative), measured in tasks per\n  second, averaged over the last 30 seconds.  This is equivalent to:\n  \x1b[1mTasksAddRate\x1b[0m - \x1b[1mTasksDispatchRate\x1b[0m.\n\nNOTE: The \x1b[1mTasksAddRate\x1b[0m and \x1b[1mTasksDispatchRate\x1b[0m metrics may differ from the\nactual rate of add/dispatch, because tasks may be dispatched eagerly to an\navailable worker, or may apply only to specific workers (they are \"sticky\").\nSuch tasks are not counted by these metrics.\n\nSafely retire Workers assigned a Build ID by checking reachability across\nall task types. Use the flag \x1b[1m--report-reachability\x1b[0m:\n\n\x1b[1mtemporal task-queue describe \\\n  --task-queue YourTaskQueue \\\n  --build-id \"YourBuildId\" \\\n  --report-reachability\x1b[0m\n\nTask reachability information is returned for the requested versions and all\ntask types, which can be used to safely retire Workers with old code versions,\nprovided that they were assigned a Build ID.\n\nNote that task reachability status is experimental and may significantly change\nor be removed in a future release. Also, determining task reachability incurs a\nnon-trivial computing cost.\n\nTask reachability states are reported per build ID. The state may be one of the\nfollowing:\n\n- \x1b[1mReachable\x1b[0m: using the current versioning rules, the Build ID may be used\n  by new Workflow Executions or Activities OR there are currently open\n  Workflow or backlogged Activity tasks assigned to the queue.\n- \x1b[1mClosedWorkflowsOnly\x1b[0m: the Build ID does not have open Workflow Executions\n  and can't be reached by new Workflow Executions. It MAY have closed\n  Workflow Executions within the Namespace retention period.\n- \x1b[1mUnreachable\x1b[0m: this Build ID is not used for new Workflow Executions and\n  isn't used by any existing Workflow Execution within the retention period.\n\nTask reachability is eventually consistent. You may experience a delay until\nreachability converges to the most accurate value. This is designed to act\nin the most conservative way until convergence. For example, \x1b[1mReachable\x1b[0m is\nmore conservative than \x1b[1mClosedWorkflowsOnly\x1b[0m.\n"
	} else {
		s.Command.Long = "Display a list of active Workers that have recently polled a Task Queue. The\nTemporal Server records each poll request time. A `LastAccessTime` over one\nminute may indicate the Worker is at capacity or has shut down. Temporal\nWorkers are removed if 5 minutes have passed since the last poll request.\n\n```\ntemporal task-queue describe \\\n  --task-queue YourTaskQueue\n```\n\nThis command provides poller information for a given Task Queue.\nWorkflow and Activity polling use separate Task Queues:\n\n```\ntemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type \"activity\"\n```\n\nThis command provides the following task queue statistics:\n- `ApproximateBacklogCount`: The approximate number of tasks backlogged in this\n  task queue. May count expired tasks but eventually converges to the right\n  value.\n- `ApproximateBacklogAge`: Approximate age of the oldest task in the backlog,\n  based on its creation time, measured in seconds.\n- `TasksAddRate`: Approximate rate at which tasks are being added to the task\n  queue, measured in tasks per second, averaged over the last 30 seconds.\n  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- `TasksDispatchRate`: Approximate rate at which tasks are being dispatched from\n  the task queue, measured in tasks per second, averaged over the last 30\n  seconds.  Includes tasks dispatched immediately without going to the backlog\n  (sync-matched tasks), as well as tasks added to the backlog. (See note below.)\n- `Backlog Increase Rate`: Approximate rate at which the backlog size is\n  increasing (if positive) or decreasing (if negative), measured in tasks per\n  second, averaged over the last 30 seconds.  This is equivalent to:\n  `TasksAddRate` - `TasksDispatchRate`.\n\nNOTE: The `TasksAddRate` and `TasksDispatchRate` metrics may differ from the\nactual rate of add/dispatch, because tasks may be dispatched eagerly to an\navailable worker, or may apply only to specific workers (they are \"sticky\").\nSuch tasks are not counted by these metrics.\n\nSafely retire Workers assigned a Build ID by checking reachability across\nall task types. Use the flag `--report-reachability`:\n\n```\ntemporal task-queue describe \\\n  --task-queue YourTaskQueue \\\n  --build-id \"YourBuildId\" \\\n  --report-reachability\n```\n\nTask reachability information is returned for the requested versions and all\ntask types, which can be used to safely retire Workers with old code versions,\nprovided that they were assigned a Build ID.\n\nNote that task reachability status is experimental and may significantly change\nor be removed in a future release. Also, determining task reachability incurs a\nnon-trivial computing cost.\n\nTask reachability states are reported per build ID. The state may be one of the\nfollowing:\n\n- `Reachable`: using the current versioning rules, the Build ID may be used\n  by new Workflow Executions or Activities OR there are currently open\n  Workflow or backlogged Activity tasks assigned to the queue.\n- `ClosedWorkflowsOnly`: the Build ID does not have open Workflow Executions\n  and can't be reached by new Workflow Executions. It MAY have closed\n  Workflow Executions within the Namespace retention period.\n- `Unreachable`: this Build ID is not used for new Workflow Executions and\n  isn't used by any existing Workflow Execution within the retention period.\n\nTask reachability is eventually consistent. You may experience a delay until\nreachability converges to the most accurate value. This is designed to act\nin the most conservative way until convergence. For example, `Reachable` is\nmore conservative than `ClosedWorkflowsOnly`.\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.TaskQueueType = NewStringEnumArray([]string{"workflow", "activity", "nexus"}, []string{})
	s.Command.Flags().Var(&s.TaskQueueType, "task-queue-type", "Task Queue type. If not specified, all types are reported.\n Accepted values: workflow, activity, nexus.")
	s.Command.Flags().StringArrayVar(&s.SelectBuildId, "select-build-id", nil, "Filter the Task Queue based on Build ID.")
	s.Command.Flags().BoolVar(&s.SelectUnversioned, "select-unversioned", false, "Include the unversioned queue.")
	s.Command.Flags().BoolVar(&s.SelectAllActive, "select-all-active", false, "Include all active versions.\nA version is active if it had new tasks or polls recently.\n")
	s.Command.Flags().BoolVar(&s.ReportReachability, "report-reachability", false, "Display task reachability information.")
	s.Command.Flags().BoolVar(&s.LegacyMode, "legacy-mode", false, "Enable a legacy mode for servers that do not support rules-based \nworker versioning.\nThis mode only provides pollers info.\n")
	s.TaskQueueTypeLegacy = NewStringEnum([]string{"workflow", "activity"}, "workflow")
	s.Command.Flags().Var(&s.TaskQueueTypeLegacy, "task-queue-type-legacy", "Task Queue type (legacy mode only). Accepted values: workflow, activity.")
	s.Command.Flags().IntVar(&s.PartitionsLegacy, "partitions-legacy", 1, "Query partitions 1 through `N`.\nExperimental/Temporary feature.\nLegacy mode only.\n")
	s.Command.Flags().BoolVar(&s.DisableStats, "disable-stats", false, "Disable task queue statistics.")
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
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nShow if a given Build ID can be used for new, existing, or closed Workflows\nin Namespaces that support Worker versioning:\n\n\x1b[1mtemporal task-queue get-build-id-reachability \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\nYou can specify the \x1b[1m--build-id\x1b[0m and \x1b[1m--task-queue\x1b[0m flags multiple times. If\n\x1b[1m--task-queue\x1b[0m is omitted, the command checks Build ID reachability against\nall Task Queues.\n"
	} else {
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nShow if a given Build ID can be used for new, existing, or closed Workflows\nin Namespaces that support Worker versioning:\n\n```\ntemporal task-queue get-build-id-reachability \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\nYou can specify the `--build-id` and `--task-queue` flags multiple times. If\n`--task-queue` is omitted, the command checks Build ID reachability against\nall Task Queues.\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.BuildId, "build-id", nil, "One or more Build ID strings.\nCan be passed multiple times.\n")
	s.ReachabilityType = NewStringEnum([]string{"open", "closed", "existing"}, "existing")
	s.Command.Flags().Var(&s.ReachabilityType, "reachability-type", "Reachability filter.\n`open`: reachable by one or more open workflows.\n`closed`: reachable by one or more closed workflows.\n`existing`: reachable by either.\nNew Workflow Executions reachable by a Build ID are always reported.\n Accepted values: open, closed, existing.")
	s.Command.Flags().StringArrayVarP(&s.TaskQueue, "task-queue", "t", nil, "Search only the specified task queue(s).\nCan be passed multiple times.\n")
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
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nFetch sets of compatible Build IDs for specified Task Queues and display their\ninformation:\n\n\x1b[1mtemporal task-queue get-build-ids \\\n    --task-queue YourTaskQueue\x1b[0m\n\nThis command is limited to Namespaces that support Worker versioning.\n"
	} else {
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nFetch sets of compatible Build IDs for specified Task Queues and display their\ninformation:\n\n```\ntemporal task-queue get-build-ids \\\n    --task-queue YourTaskQueue\n```\n\nThis command is limited to Namespaces that support Worker versioning.\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name.\n Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Flags().IntVar(&s.MaxSets, "max-sets", 0, "Max return count.\nUse 1 for default major version.\nUse 0 for all sets.\n")
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
		s.Command.Long = "Display a Task Queue's partition list with assigned matching nodes:\n\n\x1b[1mtemporal task-queue list-partition \\\n    --task-queue YourTaskQueue\x1b[0m\n"
	} else {
		s.Command.Long = "Display a Task Queue's partition list with assigned matching nodes:\n\n```\ntemporal task-queue list-partition \\\n    --task-queue YourTaskQueue\n```\n"
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
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nAdd or change a Task Queue's compatible Build IDs for Namespaces using Worker\nversioning:\n\n\x1b[1mtemporal task-queue update-build-ids [subcommands] [options] \\\n    --task-queue YourTaskQueue\x1b[0m\n"
	} else {
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nAdd or change a Task Queue's compatible Build IDs for Namespaces using Worker\nversioning:\n\n```\ntemporal task-queue update-build-ids [subcommands] [options] \\\n    --task-queue YourTaskQueue\n```\n"
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
		s.Command.Long = "Add a compatible Build ID to a Task Queue's existing version set. Provide an\nexisting Build ID and a new Build ID:\n\n\x1b[1mtemporal task-queue update-build-ids add-new-compatible \\\n    --task-queue YourTaskQueue \\\n    --existing-compatible-build-id \"YourExistingBuildId\" \\\n    --build-id \"YourNewBuildId\"\x1b[0m\n\nThe new ID is stored in the set containing the existing ID and becomes the new\ndefault for that set.\n\nThis command is limited to Namespaces that support Worker versioning.\n"
	} else {
		s.Command.Long = "Add a compatible Build ID to a Task Queue's existing version set. Provide an\nexisting Build ID and a new Build ID:\n\n```\ntemporal task-queue update-build-ids add-new-compatible \\\n    --task-queue YourTaskQueue \\\n    --existing-compatible-build-id \"YourExistingBuildId\" \\\n    --build-id \"YourNewBuildId\"\n```\n\nThe new ID is stored in the set containing the existing ID and becomes the new\ndefault for that set.\n\nThis command is limited to Namespaces that support Worker versioning.\n"
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
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nCreate a new Task Queue Build ID set, add a Build ID to it, and make it the\noverall Task Queue default. The new set will be incompatible with previous\nsets and versions.\n\n\x1b[1mtemporal task-queue update-build-ids add-new-default \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourNewBuildId\"\x1b[0m\n\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nCreate a new Task Queue Build ID set, add a Build ID to it, and make it the\noverall Task Queue default. The new set will be incompatible with previous\nsets and versions.\n\n```\ntemporal task-queue update-build-ids add-new-default \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourNewBuildId\"\n```\n\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n"
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
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nEstablish an existing Build ID as the default in its Task Queue set. New tasks\ncompatible with this set will now be dispatched to this ID:\n\n\x1b[1mtemporal task-queue update-build-ids promote-id-in-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nEstablish an existing Build ID as the default in its Task Queue set. New tasks\ncompatible with this set will now be dispatched to this ID:\n\n```\ntemporal task-queue update-build-ids promote-id-in-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n"
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
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nPromote a Build ID set to be the default on a Task Queue. Identify the set by\nproviding a Build ID within it. If the set is already the default, this\ncommand has no effect:\n\n\x1b[1mtemporal task-queue update-build-ids promote-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "+-----------------------------------------------------------------------------+\n| CAUTION: This command is deprecated and will be removed in a later release. |\n+-----------------------------------------------------------------------------+\n\nPromote a Build ID set to be the default on a Task Queue. Identify the set by\nproviding a Build ID within it. If the set is already the default, this\ncommand has no effect:\n\n```\ntemporal task-queue update-build-ids promote-set \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\n+------------------------------------------------------------------------+\n| NOTICE: This command is limited to Namespaces that support Worker      |\n| versioning. Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                     |\n+------------------------------------------------------------------------+\n"
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
		s.Command.Long = "+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n\nProvides commands to add, list, remove, or replace Worker Build ID assignment\nand redirect rules associated with Task Queues:\n\n\x1b[1mtemporal task-queue versioning [subcommands] [options] \\\n    --task-queue YourTaskQueue\x1b[0m\n\nTask Queues support the following versioning rules and policies:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID.\n"
	} else {
		s.Command.Long = "+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n\nProvides commands to add, list, remove, or replace Worker Build ID assignment\nand redirect rules associated with Task Queues:\n\n```\ntemporal task-queue versioning [subcommands] [options] \\\n    --task-queue YourTaskQueue\n```\n\nTask Queues support the following versioning rules and policies:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID.\n"
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
		s.Command.Long = "Add a new redirect rule for a given Task Queue. You may add at most one\nredirect rule for each distinct source build ID:\n\n\x1b[1mtemporal task-queue versioning add-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourSourceBuildID\" \\\n    --target-build-id \"YourTargetBuildID\"\x1b[0m\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Add a new redirect rule for a given Task Queue. You may add at most one\nredirect rule for each distinct source build ID:\n\n```\ntemporal task-queue versioning add-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourSourceBuildID\" \\\n    --target-build-id \"YourTargetBuildID\"\n```\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
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
		s.Command.Long = "Complete a Build ID's rollout and clean up unnecessary rules that might have\nbeen created during a gradual rollout:\n\n\x1b[1mtemporal task-queue versioning commit-build-id \\\n    --task-queue YourTaskQueue\n    --build-id \"YourBuildId\"\x1b[0m\n\nThis command automatically applies the following atomic changes:\n\n- Adds an unconditional assignment rule for the target Build ID at the\n  end of the list.\n- Removes all previously added assignment rules to the given target\n  Build ID.\n- Removes any unconditional assignment rules for other Build IDs.\n\nRejects requests when there have been no recent pollers for this Build ID.\nThis prevents committing invalid Build IDs. Use the \x1b[1m--force\x1b[0m option to\noverride this validation.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Complete a Build ID's rollout and clean up unnecessary rules that might have\nbeen created during a gradual rollout:\n\n```\ntemporal task-queue versioning commit-build-id \\\n    --task-queue YourTaskQueue\n    --build-id \"YourBuildId\"\n```\n\nThis command automatically applies the following atomic changes:\n\n- Adds an unconditional assignment rule for the target Build ID at the\n  end of the list.\n- Removes all previously added assignment rules to the given target\n  Build ID.\n- Removes any unconditional assignment rules for other Build IDs.\n\nRejects requests when there have been no recent pollers for this Build ID.\nThis prevents committing invalid Build IDs. Use the `--force` option to\noverride this validation.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
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
		s.Command.Long = "Deletes a rule identified by its index in the Task Queue's list of assignment\nrules:\n\n\x1b[1mtemporal task-queue versioning delete-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index YourIntegerRuleIndex\x1b[0m\n\nBy default, the Task Queue must retain one unconditional rule, such as \"no\nhint filter\" or \"percentage\". Otherwise, the delete operation is rejected.\nUse the \x1b[1m--force\x1b[0m option to override this validation.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Deletes a rule identified by its index in the Task Queue's list of assignment\nrules:\n\n```\ntemporal task-queue versioning delete-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index YourIntegerRuleIndex\n```\n\nBy default, the Task Queue must retain one unconditional rule, such as \"no\nhint filter\" or \"percentage\". Otherwise, the delete operation is rejected.\nUse the `--force` option to override this validation.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Position of the assignment rule to be replaced.\nRequests for invalid indices will fail.\n Required.")
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
		s.Command.Long = "Deletes the routing rule for the given source Build ID.\n\n\x1b[1mtemporal task-queue versioning delete-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourBuildId\"\x1b[0m\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Deletes the routing rule for the given source Build ID.\n\n```\ntemporal task-queue versioning delete-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id \"YourBuildId\"\n```\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
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
		s.Command.Long = "Retrieve all the Worker Build ID assignments and redirect rules associated\nwith a Task Queue:\n\n\x1b[1mtemporal task-queue versioning get-rules \\\n    --task-queue YourTaskQueue\x1b[0m\n\nTask Queues support the following versioning rules:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Retrieve all the Worker Build ID assignments and redirect rules associated\nwith a Task Queue:\n\n```\ntemporal task-queue versioning get-rules \\\n    --task-queue YourTaskQueue\n```\n\nTask Queues support the following versioning rules:\n\n- Assignment Rules: manage how new executions are assigned to run on specific\n  Worker Build IDs. Each Task Queue stores a list of ordered Assignment Rules,\n  which are evaluated from first to last. Assignment Rules also allow for\n  gradual rollout of new Build IDs by setting ramp percentage.\n- Redirect Rules: automatically assign work for a source Build ID to a target\n  Build ID. You may add at most one redirect rule for each source Build ID.\n  Redirect rules require that a target Build ID is fully compatible with\n  the source Build ID.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
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
		s.Command.Long = "Inserts a new assignment rule for this Task Queue. Rules are evaluated in\norder, starting from index 0. The first applicable rule is applied, and the\nrest ignored:\n\n\x1b[1mtemporal task-queue versioning insert-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\x1b[0m\n\nIf you do not specify a \x1b[1m--rule-index\x1b[0m, this command inserts at index 0.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Inserts a new assignment rule for this Task Queue. Rules are evaluated in\norder, starting from index 0. The first applicable rule is applied, and the\nrest ignored:\n\n```\ntemporal task-queue versioning insert-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --build-id \"YourBuildId\"\n```\n\nIf you do not specify a `--rule-index`, this command inserts at index 0.\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Target Build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Insertion position.\nRanges from 0 (insert at start) to count (append).\nAny number greater than the count is treated as \"append\".\n")
	s.Command.Flags().IntVar(&s.Percentage, "percentage", 100, "Traffic percent to send to target Build ID.\n")
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
		s.Command.Long = "Change an assignment rule for this Task Queue. By default, this enforces one\nunconditional rule (no hint filter or percentage). Otherwise, the operation\nwill be rejected. Set \x1b[1mforce\x1b[0m to true to bypass this validation.\n\n\x1b[1mtemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\"\x1b[0m\n\nTo assign multiple assignment rules to a single Build ID, use\n'insert-assignment-rule'.\n\nTo update the percent:\n\n\x1b[1mtemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\" \\\n    --percentage AnIntegerPercent\x1b[0m\n\nPercent may vary between 0 and 100 (default).\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Change an assignment rule for this Task Queue. By default, this enforces one\nunconditional rule (no hint filter or percentage). Otherwise, the operation\nwill be rejected. Set `force` to true to bypass this validation.\n\n```\ntemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\"\n```\n\nTo assign multiple assignment rules to a single Build ID, use\n'insert-assignment-rule'.\n\nTo update the percent:\n\n```\ntemporal task-queue versioning replace-assignment-rule \\\n    --task-queue YourTaskQueue \\\n    --rule-index AnIntegerIndex \\\n    --build-id \"YourBuildId\" \\\n    --percentage AnIntegerPercent\n```\n\nPercent may vary between 0 and 100 (default).\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Target Build ID. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Position of the assignment rule to be replaced.\nRequests for invalid indices will fail.\n Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "rule-index")
	s.Command.Flags().IntVar(&s.Percentage, "percentage", 100, "Divert percent of traffic to target Build ID.\n")
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
		s.Command.Long = "Updates a Build ID's redirect rule on a Task Queue by replacing its target\nBuild ID:\n\n\x1b[1mtemporal task-queue versioning replace-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id YourSourceBuildId \\\n    --target-build-id YourNewTargetBuildId\x1b[0m\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
	} else {
		s.Command.Long = "Updates a Build ID's redirect rule on a Task Queue by replacing its target\nBuild ID:\n\n```\ntemporal task-queue versioning replace-redirect-rule \\\n    --task-queue YourTaskQueue \\\n    --source-build-id YourSourceBuildId \\\n    --target-build-id YourNewTargetBuildId\n```\n\n+---------------------------------------------------------------------+\n| CAUTION: Worker versioning is experimental. Versioning commands are |\n| subject to change.                                                  |\n+---------------------------------------------------------------------+\n"
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
		s.Command.Long = "Workflow commands perform operations on Workflow Executions:\n\n\x1b[1mtemporal workflow [command] [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal workflow list\x1b[0m\n"
	} else {
		s.Command.Long = "Workflow commands perform operations on Workflow Executions:\n\n```\ntemporal workflow [command] [options]\n```\n\nFor example:\n\n```\ntemporal workflow list\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkflowCancelCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowCountCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowExecuteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowFixHistoryJsonCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowQueryCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowResetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowResultCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowShowCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowSignalCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowStackCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowStartCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowTerminateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowTraceCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowUpdateCommand(cctx, &s).Command)
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
		s.Command.Long = "Canceling a running Workflow Execution records a\n\x1b[1mWorkflowExecutionCancelRequested\x1b[0m event in the Event History. The Service\nschedules a new Command Task, and the Workflow Execution performs any cleanup\nwork supported by its implementation.\n\nUse the Workflow ID to cancel an Execution:\n\n\x1b[1mtemporal workflow cancel \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nA visibility Query lets you send bulk cancellations to Workflow Executions\nmatching the results:\n\n\x1b[1mtemporal workflow cancel \\\n    --query YourQuery\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference.\n"
	} else {
		s.Command.Long = "Canceling a running Workflow Execution records a\n`WorkflowExecutionCancelRequested` event in the Event History. The Service\nschedules a new Command Task, and the Workflow Execution performs any cleanup\nwork supported by its implementation.\n\nUse the Workflow ID to cancel an Execution:\n\n```\ntemporal workflow cancel \\\n    --workflow-id YourWorkflowId\n```\n\nA visibility Query lets you send bulk cancellations to Workflow Executions\nmatching the results:\n\n```\ntemporal workflow cancel \\\n    --query YourQuery\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference.\n"
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
		s.Command.Long = "Show a count of Workflow Executions, regardless of execution state (running,\nterminated, etc). Use \x1b[1m--query\x1b[0m to select a subset of Workflow Executions:\n\n\x1b[1mtemporal workflow count \\\n    --query YourQuery\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference.\n"
	} else {
		s.Command.Long = "Show a count of Workflow Executions, regardless of execution state (running,\nterminated, etc). Use `--query` to select a subset of Workflow Executions:\n\n```\ntemporal workflow count \\\n    --query YourQuery\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference.\n"
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
		s.Command.Long = "Delete a Workflow Executions and its Event History:\n\n\x1b[1mtemporal workflow delete \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nThe removal executes asynchronously. If the Execution is Running, the Service\nterminates it before deletion.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference.\n"
	} else {
		s.Command.Long = "Delete a Workflow Executions and its Event History:\n\n```\ntemporal workflow delete \\\n    --workflow-id YourWorkflowId\n```\n\nThe removal executes asynchronously. If the Execution is Running, the Service\nterminates it before deletion.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference.\n"
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
		s.Command.Long = "Display information about a specific Workflow Execution:\n\n\x1b[1mtemporal workflow describe \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nShow the Workflow Execution's auto-reset points:\n\n\x1b[1mtemporal workflow describe \\\n    --workflow-id YourWorkflowId \\\n    --reset-points true\x1b[0m\n"
	} else {
		s.Command.Long = "Display information about a specific Workflow Execution:\n\n```\ntemporal workflow describe \\\n    --workflow-id YourWorkflowId\n```\n\nShow the Workflow Execution's auto-reset points:\n\n```\ntemporal workflow describe \\\n    --workflow-id YourWorkflowId \\\n    --reset-points true\n```\n"
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
		s.Command.Long = "Establish a new Workflow Execution and direct its progress to stdout. The\ncommand blocks and returns when the Workflow Execution completes. If your\nWorkflow requires input, pass valid JSON:\n\n\x1b[1mtemporal workflow execute\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\x1b[0m\n\nUse \x1b[1m--event-details\x1b[0m to relay updates to the command-line output in JSON\nformat. When using JSON output (\x1b[1m--output json\x1b[0m), this includes the entire\n\"history\" JSON key for the run.\n"
	} else {
		s.Command.Long = "Establish a new Workflow Execution and direct its progress to stdout. The\ncommand blocks and returns when the Workflow Execution completes. If your\nWorkflow requires input, pass valid JSON:\n\n```\ntemporal workflow execute\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\n```\n\nUse `--event-details` to relay updates to the command-line output in JSON\nformat. When using JSON output (`--output json`), this includes the entire\n\"history\" JSON key for the run.\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.Detailed, "detailed", false, "Display events as sections instead of table.\nDoes not apply to JSON output.\n")
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
		s.Command.Long = "Reserialize an Event History JSON file:\n\n\x1b[1mtemporal workflow fix-history-json \\\n    --source /path/to/original.json \\\n    --target /path/to/reserialized.json\x1b[0m\n"
	} else {
		s.Command.Long = "Reserialize an Event History JSON file:\n\n```\ntemporal workflow fix-history-json \\\n    --source /path/to/original.json \\\n    --target /path/to/reserialized.json\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Source, "source", "s", "", "Path to the original file. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source")
	s.Command.Flags().StringVarP(&s.Target, "target", "t", "", "Path to the results file.\nWhen omitted, output is sent to stdout.\n")
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
}

func NewTemporalWorkflowListCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowListCommand {
	var s TemporalWorkflowListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Show Workflow Executions"
	if hasHighlighting {
		s.Command.Long = "List Workflow Executions. By default, this command returns up to 10 closed\nWorkflow Executions. The optional \x1b[1m--query\x1b[0m limits the output to Workflows\nmatching a Query:\n\n\x1b[1mtemporal workflow list \\\n    --query YourQuery\x1b[1m\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[0mtemporal batch --help` for a quick reference.\n\nView a list of archived Workflow Executions:\n\n\x1b[1mtemporal workflow list \\\n    --archived\x1b[0m\n"
	} else {
		s.Command.Long = "List Workflow Executions. By default, this command returns up to 10 closed\nWorkflow Executions. The optional `--query` limits the output to Workflows\nmatching a Query:\n\n```\ntemporal workflow list \\\n    --query YourQuery`\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference.\n\nView a list of archived Workflow Executions:\n\n```\ntemporal workflow list \\\n    --archived\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.")
	s.Command.Flags().BoolVar(&s.Archived, "archived", false, "Limit output to archived Workflow Executions.")
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Maximum number of Workflow Executions to display.")
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
	Name            string
	RejectCondition StringEnum
}

func NewTemporalWorkflowQueryCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowQueryCommand {
	var s TemporalWorkflowQueryCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "query [flags]"
	s.Command.Short = "Retrieve Workflow Execution state"
	if hasHighlighting {
		s.Command.Long = "Send a Query to a Workflow Execution by Workflow ID to retrieve its state.\nThis synchronous operation exposes the internal state of a running Workflow\nExecution, which constantly changes. You can query both running and completed\nWorkflow Executions:\n\n\x1b[1mtemporal workflow query \\\n    --workflow-id YourWorkflowId\n    --type YourQueryType\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\x1b[0m\n"
	} else {
		s.Command.Long = "Send a Query to a Workflow Execution by Workflow ID to retrieve its state.\nThis synchronous operation exposes the internal state of a running Workflow\nExecution, which constantly changes. You can query both running and completed\nWorkflow Executions:\n\n```\ntemporal workflow query \\\n    --workflow-id YourWorkflowId\n    --type YourQueryType\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Query Type/Name. Required. Aliased as \"--type\".")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.RejectCondition = NewStringEnum([]string{"not_open", "not_completed_cleanly"}, "")
	s.Command.Flags().Var(&s.RejectCondition, "reject-condition", "Optional flag for rejecting Queries based on Workflow state.\n Accepted values: not_open, not_completed_cleanly.")
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
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
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "reset [flags]"
	s.Command.Short = "Move Workflow Execution history point"
	if hasHighlighting {
		s.Command.Long = "Reset a Workflow Execution so it can resume from a point in its Event History\nwithout losing its progress up to that point:\n\n\x1b[1mtemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --event-id YourLastEvent\x1b[0m\n\nStart from where the Workflow Execution last continued as new:\n\n\x1b[1mtemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --type LastContinuedAsNew\x1b[0m\n\nFor batch resets, limit your resets to FirstWorkflowTask, LastWorkflowTask, or\nBuildId. Do not use Workflow IDs, run IDs, or event IDs with this command.\n\nVisit https://docs.temporal.io/visibility to read more about Search\nAttributes and Query creation.\n"
	} else {
		s.Command.Long = "Reset a Workflow Execution so it can resume from a point in its Event History\nwithout losing its progress up to that point:\n\n```\ntemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --event-id YourLastEvent\n```\n\nStart from where the Workflow Execution last continued as new:\n\n```\ntemporal workflow reset \\\n    --workflow-id YourWorkflowId \\\n    --type LastContinuedAsNew\n```\n\nFor batch resets, limit your resets to FirstWorkflowTask, LastWorkflowTask, or\nBuildId. Do not use Workflow IDs, run IDs, or event IDs with this command.\n\nVisit https://docs.temporal.io/visibility to read more about Search\nAttributes and Query creation.\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow ID.\nRequired for non-batch reset operations.\n")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID.")
	s.Command.Flags().IntVarP(&s.EventId, "event-id", "e", 0, "Event ID to reset to.\nEvent must occur after `WorkflowTaskStarted`.\n`WorkflowTaskCompleted`, `WorkflowTaskFailed`, etc. are valid.\n")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for reset. Required.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "reason")
	s.ReapplyType = NewStringEnum([]string{"All", "Signal", "None"}, "All")
	s.Command.Flags().Var(&s.ReapplyType, "reapply-type", "Types of events to re-apply after reset point.\nDeprecated.\nUse --reapply-exclude instead.\n Accepted values: All, Signal, None.")
	s.ReapplyExclude = NewStringEnumArray([]string{"All", "Signal", "Update"}, []string{})
	s.Command.Flags().Var(&s.ReapplyExclude, "reapply-exclude", "Exclude these event types from re-application.\n Accepted values: All, Signal, Update.")
	s.Type = NewStringEnum([]string{"FirstWorkflowTask", "LastWorkflowTask", "LastContinuedAsNew", "BuildId"}, "")
	s.Command.Flags().VarP(&s.Type, "type", "t", "The event type for the reset. Accepted values: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew, BuildId.")
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "A Build ID.\nUse only with the BuildId `--type`.\nResets the first Workflow task processed by this ID.\nBy default, this reset may be in a prior run, earlier than a Continue\nas New point.\n")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm.\nOnly allowed when `--query` is present.\n")
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
		s.Command.Long = "Wait for and print the result of a Workflow Execution:\n\n\x1b[1mtemporal workflow result \\\n    --workflow-id YourWorkflowId\x1b[0m\n"
	} else {
		s.Command.Long = "Wait for and print the result of a Workflow Execution:\n\n```\ntemporal workflow result \\\n    --workflow-id YourWorkflowId\n```\n"
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
		s.Command.Long = "Show a Workflow Execution's Event History.\nWhen using JSON output (\x1b[1m--output json\x1b[0m), you may pass the results to an SDK\nto perform a replay:\n\n\x1b[1mtemporal workflow show \\\n    --workflow-id YourWorkflowId\n    --output json\x1b[0m\n"
	} else {
		s.Command.Long = "Show a Workflow Execution's Event History.\nWhen using JSON output (`--output json`), you may pass the results to an SDK\nto perform a replay:\n\n```\ntemporal workflow show \\\n    --workflow-id YourWorkflowId\n    --output json\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVarP(&s.Follow, "follow", "f", false, "Follow the Workflow Execution progress in real time.\nDoes not apply to JSON output.\n")
	s.Command.Flags().BoolVar(&s.Detailed, "detailed", false, "Display events as detailed sections instead of table.\nDoes not apply to JSON output.\n")
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
		s.Command.Long = "Send an asynchronous notification (Signal) to a running Workflow Execution by\nits Workflow ID. The Signal is written to the History. When you include\n\x1b[1m--input\x1b[0m, that data is available for the Workflow Execution to consume:\n\n\x1b[1mtemporal workflow signal \\\n    --workflow-id YourWorkflowId \\\n    --name YourSignal \\\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\x1b[0m\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference.\n"
	} else {
		s.Command.Long = "Send an asynchronous notification (Signal) to a running Workflow Execution by\nits Workflow ID. The Signal is written to the History. When you include\n`--input`, that data is available for the Workflow Execution to consume:\n\n```\ntemporal workflow signal \\\n    --workflow-id YourWorkflowId \\\n    --name YourSignal \\\n    --input '{\"YourInputKey\": \"YourInputValue\"}'\n```\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference.\n"
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
		s.Command.Long = "Perform a Query on a Workflow Execution using a \x1b[1m__stack_trace\x1b[0m-type Query.\nDisplay a stack trace of the threads and routines currently in use by the\nWorkflow for troubleshooting:\n\n\x1b[1mtemporal workflow stack \\\n    --workflow-id YourWorkflowId\x1b[0m\n"
	} else {
		s.Command.Long = "Perform a Query on a Workflow Execution using a `__stack_trace`-type Query.\nDisplay a stack trace of the threads and routines currently in use by the\nWorkflow for troubleshooting:\n\n```\ntemporal workflow stack \\\n    --workflow-id YourWorkflowId\n```\n"
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
		s.Command.Long = "Start a new Workflow Execution. Returns the Workflow- and Run-IDs:\n\n\x1b[1mtemporal workflow start \\\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\x1b[0m\n"
	} else {
		s.Command.Long = "Start a new Workflow Execution. Returns the Workflow- and Run-IDs:\n\n```\ntemporal workflow start \\\n    --workflow-id YourWorkflowId \\\n    --type YourWorkflow \\\n    --task-queue YourTaskQueue \\\n    --input '{\"some-key\": \"some-value\"}'\n```\n"
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

type TemporalWorkflowTerminateCommand struct {
	Parent     *TemporalWorkflowCommand
	Command    cobra.Command
	WorkflowId string
	Query      string
	RunId      string
	Reason     string
	Yes        bool
}

func NewTemporalWorkflowTerminateCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowTerminateCommand {
	var s TemporalWorkflowTerminateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "terminate [flags]"
	s.Command.Short = "Forcefully end a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "Terminate a Workflow Execution:\n\n\x1b[1mtemporal workflow terminate \\\n    --reason YourReasonForTermination \\\n    --workflow-id YourWorkflowId\x1b[0m\n\nThe reason is optional and defaults to the current user's name. The reason\nis stored in the Event History as part of the \x1b[1mWorkflowExecutionTerminated\x1b[0m\nevent. This becomes the closing Event in the Workflow Execution's history.\n\nExecutions may be terminated in bulk via a visibility Query list filter:\n\n\x1b[1mtemporal workflow terminate \\\n    --query YourQuery \\\n    --reason YourReasonForTermination\x1b[0m\n\nWorkflow code cannot see or respond to terminations. To perform clean-up work\nin your Workflow code, use \x1b[1mtemporal workflow cancel\x1b[0m instead.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See \x1b[1mtemporal batch --help\x1b[0m for a quick reference.\n"
	} else {
		s.Command.Long = "Terminate a Workflow Execution:\n\n```\ntemporal workflow terminate \\\n    --reason YourReasonForTermination \\\n    --workflow-id YourWorkflowId\n```\n\nThe reason is optional and defaults to the current user's name. The reason\nis stored in the Event History as part of the `WorkflowExecutionTerminated`\nevent. This becomes the closing Event in the Workflow Execution's history.\n\nExecutions may be terminated in bulk via a visibility Query list filter:\n\n```\ntemporal workflow terminate \\\n    --query YourQuery \\\n    --reason YourReasonForTermination\n```\n\nWorkflow code cannot see or respond to terminations. To perform clean-up work\nin your Workflow code, use `temporal workflow cancel` instead.\n\nVisit https://docs.temporal.io/visibility to read more about Search Attributes\nand Query creation. See `temporal batch --help` for a quick reference.\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow ID.\nYou must set either --workflow-id or --query.\n")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Content for an SQL-like `QUERY` List Filter.\nYou must set either --workflow-id or --query.\n")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID.\nCan only be set with --workflow-id.\nDo not use with --query.\n")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for termination.\nDefaults to message with the current user's name.\n")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Don't prompt to confirm termination.\nCan only be used with --query.\n")
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
		s.Command.Long = "Display the progress of a Workflow Execution and its child workflows with a\nreal-time trace. This view helps you understand how Workflows are proceeding:\n\n\x1b[1mtemporal workflow trace \\\n    --workflow-id YourWorkflowId\x1b[0m\n"
	} else {
		s.Command.Long = "Display the progress of a Workflow Execution and its child workflows with a\nreal-time trace. This view helps you understand how Workflows are proceeding:\n\n```\ntemporal workflow trace \\\n    --workflow-id YourWorkflowId\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Fold, "fold", nil, "Fold away Child Workflows with the specified statuses.\nCase-insensitive.\nIgnored if --no-fold supplied.\nAvailable values: running, completed, failed, canceled, terminated,\ntimedout, continueasnew.\nCan be passed multiple times.\n")
	s.Command.Flags().BoolVar(&s.NoFold, "no-fold", false, "Disable folding.\nFetch and display Child Workflows within the set depth.\n")
	s.Command.Flags().IntVar(&s.Depth, "depth", -1, "Set depth for your Child Workflow fetches.\nPass -1 to fetch child workflows at any depth.\n")
	s.Command.Flags().IntVar(&s.Concurrency, "concurrency", 10, "Number of Workflow Histories to fetch at a time.\n")
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
	s.Command.Short = "Updates (Experimental)"
	s.Command.Long = "An Update is a synchronous call to a Workflow Execution that can change its\nstate, control its flow, and return a result.\n\nExperimental.\n"
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
	s.Command.Short = "Obtain status info about a specific Update (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Given a Workflow Execution and an Update ID, return information about its current status, including\na result if it has finished.\n\nExperimental.\n\n\x1b[1mtemporal workflow update describe \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\x1b[0m\n"
	} else {
		s.Command.Long = "Given a Workflow Execution and an Update ID, return information about its current status, including\na result if it has finished.\n\nExperimental.\n\n```\ntemporal workflow update describe \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\n```\n"
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
	s.Command.Short = "Send an Update and wait for it to complete (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to complete or fail. You can also use this to wait for an existing\nupdate to complete, by submitting an existing update ID.\n\nExperimental.\n\n\x1b[1mtemporal workflow update execute \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\x1b[0m\n"
	} else {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to complete or fail. You can also use this to wait for an existing\nupdate to complete, by submitting an existing update ID.\n\nExperimental.\n\n```\ntemporal workflow update execute \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\n```\n"
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
	s.Command.Short = "Wait for a specific Update to complete (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Given a Workflow Execution and an Update ID, wait for the Update to complete or fail and\nprint the result.\n\nExperimental.\n\n\x1b[1mtemporal workflow update result \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\x1b[0m\n"
	} else {
		s.Command.Long = "Given a Workflow Execution and an Update ID, wait for the Update to complete or fail and\nprint the result.\n\nExperimental.\n\n```\ntemporal workflow update result \\\n    --workflow-id YourWorkflowId \\\n    --update-id YourUpdateId\n```\n"
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
	s.Command.Short = "Send an Update and wait for it to be accepted or rejected (Experimental)"
	if hasHighlighting {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to be accepted or rejected. You can subsequently wait for the update\nto complete by using \x1b[1mtemporal workflow update execute\x1b[0m.\n\nExperimental.\n\n\x1b[1mtemporal workflow update start \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\x1b[0m\n"
	} else {
		s.Command.Long = "Send a message to a Workflow Execution to invoke an Update handler, and wait for\nthe update to be accepted or rejected. You can subsequently wait for the update\nto complete by using `temporal workflow update execute`.\n\nExperimental.\n\n```\ntemporal workflow update start \\\n    --workflow-id YourWorkflowId \\\n    --name YourUpdate \\\n    --input '{\"some-key\": \"some-value\"}'\n```\n"
	}
	s.Command.Args = cobra.NoArgs
	s.WaitForStage = NewStringEnum([]string{"accepted"}, "")
	s.Command.Flags().Var(&s.WaitForStage, "wait-for-stage", "Update stage to wait for.\nThe only option is `accepted`, but this option is  required. This is to allow\na future version of the CLI to choose a default value.\n Accepted values: accepted. Required.")
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
