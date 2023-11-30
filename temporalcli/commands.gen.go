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

type TemporalCommand struct {
	Command    cobra.Command
	Env        string
	EnvFile    string
	LogLevel   StringEnum
	LogFormat  StringEnum
	Output     StringEnum
	TimeFormat StringEnum
	Color      StringEnum
}

func NewTemporalCommand(cctx *CommandContext) *TemporalCommand {
	var s TemporalCommand
	s.Command.Use = "temporal"
	s.Command.Short = "Temporal command-line interface and development server."
	s.Command.Long = ""
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalEnvCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalServerCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVar(&s.Env, "env", "default", "Environment to read environmental variables from.")
	s.Command.PersistentFlags().StringVar(&s.EnvFile, "env-file", "", "File to read all environments (defaults to `$HOME/.config/temporalio/temporal.yaml`).")
	s.LogLevel = NewStringEnum([]string{"debug", "info", "warn", "error", "off"}, "info")
	s.Command.PersistentFlags().Var(&s.LogLevel, "log-level", "Log level.")
	s.LogFormat = NewStringEnum([]string{"text", "json"}, "text")
	s.Command.PersistentFlags().Var(&s.LogFormat, "log-format", "Log format.")
	s.Output = NewStringEnum([]string{"text", "json"}, "text")
	s.Command.PersistentFlags().VarP(&s.Output, "output", "o", "Data output format.")
	s.TimeFormat = NewStringEnum([]string{"relative", "iso", "raw"}, "relative")
	s.Command.PersistentFlags().Var(&s.TimeFormat, "time-format", "Time format.")
	s.Color = NewStringEnum([]string{"always", "never", "auto"}, "auto")
	s.Command.PersistentFlags().Var(&s.Color, "color", "Set coloring.")
	cctx.BindConfigFlags(s.Command.PersistentFlags())
	s.initCommand(cctx)
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
	s.Command.Short = "Manage environments."
	s.Command.Long = "Use the '--env <env name>' option with other commands to point the CLI at a different Temporal Server instance. If --env\nis not passed, the 'default' environment is used."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalEnvDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvGetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvSetCommand(cctx, &s).Command)
	cctx.BindConfigFlags(s.Command.PersistentFlags())
	return &s
}

type TemporalEnvDeleteCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
}

func NewTemporalEnvDeleteCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvDeleteCommand {
	var s TemporalEnvDeleteCommand
	s.Parent = parent
	s.Command.Use = "delete"
	s.Command.Short = "Delete an environment or environment property."
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal env delete [environment or property]\x1b[0m\n\nDelete an environment or just a single property:\n\n\x1b[1mtemporal env delete prod\x1b[0m\n\x1b[1mtemporal env delete prod.tls-cert-path\x1b[0m"
	} else {
		s.Command.Long = "`temporal env delete [environment or property]`\n\nDelete an environment or just a single property:\n\n`temporal env delete prod`\n`temporal env delete prod.tls-cert-path`"
	}
	s.Command.Args = cobra.ExactArgs(1)
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalEnvGetCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
}

func NewTemporalEnvGetCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvGetCommand {
	var s TemporalEnvGetCommand
	s.Parent = parent
	s.Command.Use = "get"
	s.Command.Short = "Print environment properties."
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal env get [environment or property]\x1b[0m\n\nPrint all properties of the 'prod' environment:\n\n\x1b[1mtemporal env get prod\x1b[0m\n\ntls-cert-path  /home/my-user/certs/client.cert\ntls-key-path   /home/my-user/certs/client.key\naddress        temporal.example.com:7233\nnamespace      someNamespace\n\nPrint a single property:\n\n\x1b[1mtemporal env get prod.tls-key-path\x1b[0m\n\ntls-key-path  /home/my-user/certs/cluster.key"
	} else {
		s.Command.Long = "`temporal env get [environment or property]`\n\nPrint all properties of the 'prod' environment:\n\n`temporal env get prod`\n\ntls-cert-path  /home/my-user/certs/client.cert\ntls-key-path   /home/my-user/certs/client.key\naddress        temporal.example.com:7233\nnamespace      someNamespace\n\nPrint a single property:\n\n`temporal env get prod.tls-key-path`\n\ntls-key-path  /home/my-user/certs/cluster.key"
	}
	s.Command.Args = cobra.ExactArgs(1)
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalEnvListCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
}

func NewTemporalEnvListCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvListCommand {
	var s TemporalEnvListCommand
	s.Parent = parent
	s.Command.Use = "list"
	s.Command.Short = "Print all environments."
	s.Command.Long = "List all environments."
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalEnvSetCommand struct {
	Parent  *TemporalEnvCommand
	Command cobra.Command
}

func NewTemporalEnvSetCommand(cctx *CommandContext, parent *TemporalEnvCommand) *TemporalEnvSetCommand {
	var s TemporalEnvSetCommand
	s.Parent = parent
	s.Command.Use = "set"
	s.Command.Short = "Set environment properties."
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal env set [environment.property name] [property value]\x1b[0m\n\nProperty names match CLI option names, for example '--address' and '--tls-cert-path':\n\n\x1b[1mtemporal env set prod.address 127.0.0.1:7233\x1b[0m\n\x1b[1mtemporal env set prod.tls-cert-path  /home/my-user/certs/cluster.cert\x1b[0m"
	} else {
		s.Command.Long = "`temporal env set [environment.property name] [property value]`\n\nProperty names match CLI option names, for example '--address' and '--tls-cert-path':\n\n`temporal env set prod.address 127.0.0.1:7233`\n`temporal env set prod.tls-cert-path  /home/my-user/certs/cluster.cert`"
	}
	s.Command.Args = cobra.ExactArgs(2)
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
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
	s.Command.Short = "Run Temporal Server."
	if hasHighlighting {
		s.Command.Long = "Start a development version of Temporal Server:\n\n\x1b[1mtemporal server start-dev\x1b[0m"
	} else {
		s.Command.Long = "Start a development version of Temporal Server:\n\n`temporal server start-dev`"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalServerStartDevCommand(cctx, &s).Command)
	cctx.BindConfigFlags(s.Command.PersistentFlags())
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
	UiAssetPath        string
	UiCodecEndpoint    string
	SqlitePragma       []string
	DynamicConfigValue []string
	LogConfig          bool
}

func NewTemporalServerStartDevCommand(cctx *CommandContext, parent *TemporalServerCommand) *TemporalServerStartDevCommand {
	var s TemporalServerStartDevCommand
	s.Parent = parent
	s.Command.Use = "start-dev"
	s.Command.Short = "Start Temporal development server."
	if hasHighlighting {
		s.Command.Long = "Start Temporal Server on \x1b[1mlocalhost:7233\x1b[0m with:\n\n\x1b[1mtemporal server start-dev\x1b[0m\n\nView the UI at http://localhost:8233\n\nTo persist Workflows across runs, use:\n\n\x1b[1mtemporal server start-dev --db-filename temporal.db\x1b[0m"
	} else {
		s.Command.Long = "Start Temporal Server on `localhost:7233` with:\n\n`temporal server start-dev`\n\nView the UI at http://localhost:8233\n\nTo persist Workflows across runs, use:\n\n`temporal server start-dev --db-filename temporal.db`"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.DbFilename, "db-filename", "f", "", "File in which to persist Temporal state (by default, Workflows are lost when the process dies).")
	s.Command.Flags().StringArrayVarP(&s.Namespace, "namespace", "n", nil, "Specify namespaces that should be pre-created (namespace \"default\" is always created).")
	s.Command.Flags().IntVarP(&s.Port, "port", "p", 7233, "Port for the frontend gRPC service.")
	s.Command.Flags().IntVar(&s.HttpPort, "http-port", 0, "Port for the frontend HTTP API service. Default is off.")
	s.Command.Flags().IntVar(&s.MetricsPort, "metrics-port", 0, "Port for /metrics. Default is off.")
	s.Command.Flags().IntVar(&s.UiPort, "ui-port", 0, "Port for the Web UI. Default is --port + 1000.")
	s.Command.Flags().BoolVar(&s.Headless, "headless", false, "Disable the Web UI.")
	s.Command.Flags().StringVar(&s.Ip, "ip", "127.0.0.1", "IP address to bind the frontend service to.")
	s.Command.Flags().StringVar(&s.UiIp, "ui-ip", "", "IP address to bind the Web UI to. Default is same as --ip.")
	s.Command.Flags().StringVar(&s.UiAssetPath, "ui-asset-path", "", "UI custom assets path.")
	s.Command.Flags().StringVar(&s.UiCodecEndpoint, "ui-codec-endpoint", "", "UI remote codec HTTP endpoint.")
	s.Command.Flags().StringArrayVar(&s.SqlitePragma, "sqlite-pragma", nil, "Specify SQLite pragma statements in pragma=value format.")
	s.Command.Flags().StringArrayVar(&s.DynamicConfigValue, "dynamic-config-value", nil, "Dynamic config value, as KEY=JSON_VALUE (string values need quotes).")
	s.Command.Flags().BoolVar(&s.LogConfig, "log-config", false, "Log the server config being used in stderr.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type ClientOptions struct {
	Address                    string
	Namespace                  string
	GrpcMeta                   []string
	Tls                        bool
	TlsCertPath                string
	TlsKeyPath                 string
	TlsCaPath                  string
	TlsDisableHostVerification bool
	TlsServerName              string
	ContextTimeout             time.Duration
	CodecEndpoint              string
	CodecAuth                  string
}

func (v *ClientOptions) buildFlags(f *pflag.FlagSet) {
	f.StringVar(&v.Address, "address", "", "Temporal server address.")
	_ = cobra.MarkFlagRequired(f, "address")
	f.StringVarP(&v.Namespace, "namespace", "n", "default", "Temporal server namespace.")
	f.StringArrayVar(&v.GrpcMeta, "grpc-meta", nil, "Contains gRPC metadata to send with requests (formatted as key=value).")
	f.BoolVar(&v.Tls, "tls", false, "Enable TLS encryption without additional options such as mTLS or client certificates.")
	f.StringVar(&v.TlsCertPath, "tls-cert-path", "", "Path to x509 certificate.")
	f.StringVar(&v.TlsKeyPath, "tls-key-path", "", "Path to private certificate key.")
	f.StringVar(&v.TlsCaPath, "tls-ca-path", "", "Path to server CA certificate.")
	f.BoolVar(&v.TlsDisableHostVerification, "tls-disable-host-verification", false, "Disables TLS host-name verification.")
	f.StringVar(&v.TlsServerName, "tls-server-name", "", "Overrides target TLS server name.")
	f.DurationVar(&v.ContextTimeout, "context-timeout", 5000*time.Millisecond, "Optional timeout for the context of an RPC call.")
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "Endpoint for a remote Codec Server.")
	f.StringVar(&v.CodecAuth, "codec-auth", "", "Sets the authorization header on requests to the Codec Server.")
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
	s.Command.Short = "Start, list, and operate on Workflows."
	if hasHighlighting {
		s.Command.Long = "Workflow commands perform operations on \nWorkflow Executions.\n\nWorkflow commands use this syntax:\x1b[1mtemporal workflow COMMAND [ARGS]\x1b[0m."
	} else {
		s.Command.Long = "Workflow commands perform operations on \nWorkflow Executions.\n\nWorkflow commands use this syntax:`temporal workflow COMMAND [ARGS]`."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalWorkflowCancelCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowCountCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowDeleteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowExecuteCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowListCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowQueryCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowResetCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowResetBatchCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowShowCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowSignalCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowStackCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowStartCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowTerminateCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowTraceCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowUpdateCommand(cctx, &s).Command)
	s.ClientOptions.buildFlags(s.Command.PersistentFlags())
	cctx.BindConfigFlags(s.Command.PersistentFlags())
	return &s
}

type TemporalWorkflowCancelCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowCancelCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowCancelCommand {
	var s TemporalWorkflowCancelCommand
	s.Parent = parent
	s.Command.Use = "cancel"
	s.Command.Short = "Cancel a Workflow Execution."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowCountCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowCountCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowCountCommand {
	var s TemporalWorkflowCountCommand
	s.Parent = parent
	s.Command.Use = "count"
	s.Command.Short = "Count Workflow Executions."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowDeleteCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowDeleteCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowDeleteCommand {
	var s TemporalWorkflowDeleteCommand
	s.Parent = parent
	s.Command.Use = "delete"
	s.Command.Short = "Deletes a Workflow Execution."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowDescribeCommand struct {
	Parent      *TemporalWorkflowCommand
	Command     cobra.Command
	WorkflowId  string
	RunId       string
	ResetPoints bool
	Raw         bool
}

func NewTemporalWorkflowDescribeCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowDescribeCommand {
	var s TemporalWorkflowDescribeCommand
	s.Parent = parent
	s.Command.Use = "describe"
	s.Command.Short = "Show information about a Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow describe\x1b[0m command shows information about a given\nWorkflow Execution.\n\nThis information can be used to locate Workflow Executions that weren't able to run successfully.\n\n\x1b[1mtemporal workflow describe --workflow-id=meaningful-business-id\x1b[0m\n\nOutput can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.\n\n\x1b[1mtemporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true\x1b[0m\n\nUse the command options below to change the information returned by this command."
	} else {
		s.Command.Long = "The `temporal workflow describe` command shows information about a given\nWorkflow Execution.\n\nThis information can be used to locate Workflow Executions that weren't able to run successfully.\n\n`temporal workflow describe --workflow-id=meaningful-business-id`\n\nOutput can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.\n\n`temporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true`\n\nUse the command options below to change the information returned by this command."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow Id.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "workflow-id")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run Id.")
	s.Command.Flags().BoolVar(&s.ResetPoints, "reset-points", false, "Only show auto-reset points.")
	s.Command.Flags().BoolVar(&s.Raw, "raw", false, "Print properties without changing their format.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowExecuteCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowStartOptions
	PayloadInputOptions
	EventDetails bool
}

func NewTemporalWorkflowExecuteCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowExecuteCommand {
	var s TemporalWorkflowExecuteCommand
	s.Parent = parent
	s.Command.Use = "execute"
	s.Command.Short = "Start a new Workflow Execution and prints its progress."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow execute\x1b[0m command starts a new Workflow Execution and\nprints its progress. The command completes when the Workflow Execution completes.\n\nSingle quotes('') are used to wrap input as JSON.\n\n\x1b[1mtemporal workflow execute\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m"
	} else {
		s.Command.Long = "The `temporal workflow execute` command starts a new Workflow Execution and\nprints its progress. The command completes when the Workflow Execution completes.\n\nSingle quotes('') are used to wrap input as JSON.\n\n```\ntemporal workflow execute\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowStartOptions.buildFlags(s.Command.Flags())
	s.PayloadInputOptions.buildFlags(s.Command.Flags())
	s.Command.Flags().BoolVar(&s.EventDetails, "event-details", false, "If set, when using text output this will print the event details instead of just the event during workflow progress.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowListCommand struct {
	Parent   *TemporalWorkflowCommand
	Command  cobra.Command
	Query    string
	Archived bool
}

func NewTemporalWorkflowListCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowListCommand {
	var s TemporalWorkflowListCommand
	s.Parent = parent
	s.Command.Use = "list"
	s.Command.Short = "List Workflow Executions based on a Query."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow list\x1b[0m command provides a list of Workflow Executions\nthat meet the criteria of a given Query.\nBy default, this command returns up to 10 closed Workflow Executions.\n\n\x1b[1mtemporal workflow list --query=MyQuery\x1b[0m\n\nThe command can also return a list of archived Workflow Executions.\n\n\x1b[1mtemporal workflow list --archived\x1b[0m\n\nUse the command options below to change the information returned by this command."
	} else {
		s.Command.Long = "The `temporal workflow list` command provides a list of Workflow Executions\nthat meet the criteria of a given Query.\nBy default, this command returns up to 10 closed Workflow Executions.\n\n`temporal workflow list --query=MyQuery`\n\nThe command can also return a list of archived Workflow Executions.\n\n`temporal workflow list --archived`\n\nUse the command options below to change the information returned by this command."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Filter results using a SQL-like query.")
	s.Command.Flags().BoolVar(&s.Archived, "archived", false, "If set, will only query and list archived workflows instead of regular workflows.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowQueryCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowQueryCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowQueryCommand {
	var s TemporalWorkflowQueryCommand
	s.Parent = parent
	s.Command.Use = "query"
	s.Command.Short = "Query a Workflow Execution."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowResetCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowResetCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowResetCommand {
	var s TemporalWorkflowResetCommand
	s.Parent = parent
	s.Command.Use = "reset"
	s.Command.Short = "Resets a Workflow Execution by Event ID or reset type."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowResetBatchCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowResetBatchCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowResetBatchCommand {
	var s TemporalWorkflowResetBatchCommand
	s.Parent = parent
	s.Command.Use = "reset-batch"
	s.Command.Short = "Reset a batch of Workflow Executions by reset type."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowShowCommand struct {
	Parent      *TemporalWorkflowCommand
	Command     cobra.Command
	WorkflowId  string
	RunId       string
	ResetPoints bool
	Follow      bool
}

func NewTemporalWorkflowShowCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowShowCommand {
	var s TemporalWorkflowShowCommand
	s.Parent = parent
	s.Command.Use = "show"
	s.Command.Short = "Show Event History for a Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow show\x1b[0m command provides the Event History for a\nWorkflow Execution.\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow show` command provides the Event History for a\nWorkflow Execution.\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow Id.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "workflow-id")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run Id.")
	s.Command.Flags().BoolVar(&s.ResetPoints, "reset-points", false, "Only show auto-reset points.")
	s.Command.Flags().BoolVar(&s.Follow, "follow", false, "Follow the progress of a Workflow Execution if it goes to a new run.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowSignalCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowSignalCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowSignalCommand {
	var s TemporalWorkflowSignalCommand
	s.Parent = parent
	s.Command.Use = "signal"
	s.Command.Short = "Signal Workflow Execution by Id or List Filter."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowStackCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowStackCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowStackCommand {
	var s TemporalWorkflowStackCommand
	s.Parent = parent
	s.Command.Use = "stack"
	s.Command.Short = "Query a Workflow Execution with __stack_trace as the query type."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type WorkflowStartOptions struct {
	WorkflowId       string
	Type             string
	TaskQueue        string
	RunTimeout       time.Duration
	ExecutionTimeout time.Duration
	TaskTimeout      time.Duration
	Cron             string
	IdReusePolicy    string
	SearchAttribute  []string
	Memo             []string
}

func (v *WorkflowStartOptions) buildFlags(f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow Id.")
	f.StringVar(&v.Type, "type", "", "Workflow Type name.")
	_ = cobra.MarkFlagRequired(f, "type")
	f.StringVarP(&v.TaskQueue, "task-queue", "t", "", "Workflow Task queue.")
	_ = cobra.MarkFlagRequired(f, "task-queue")
	f.DurationVar(&v.RunTimeout, "run-timeout", 0, "Timeout of a Workflow Run.")
	f.DurationVar(&v.ExecutionTimeout, "execution-timeout", 0, "Timeout for a WorkflowExecution, including retries and ContinueAsNew tasks.")
	f.DurationVar(&v.TaskTimeout, "task-timeout", 10000*time.Millisecond, "Start-to-close timeout for a Workflow Task.")
	f.StringVar(&v.Cron, "cron", "", "Cron schedule for the workflow. Deprecated - use schedules instead.")
	f.StringVar(&v.IdReusePolicy, "id-reuse-policy", "", "Allows the same Workflow Id to be used in a new Workflow Execution.")
	f.StringArrayVar(&v.SearchAttribute, "search-attribute", nil, "Passes Search Attribute in key=value format. Use valid JSON formats for value.")
	f.StringArrayVar(&v.Memo, "memo", nil, "Passes Memo in key=value format. Use valid JSON formats for value.")
}

type PayloadInputOptions struct {
	Input       []string
	InputFile   []string
	InputMeta   []string
	InputBase64 bool
}

func (v *PayloadInputOptions) buildFlags(f *pflag.FlagSet) {
	f.StringArrayVarP(&v.Input, "input", "i", nil, "Input value (default JSON unless --input-payload-meta is non-JSON encoding). Can be given multiple times for multiple arguments. Cannot be combined with --input-file.")
	f.StringArrayVar(&v.InputFile, "input-file", nil, "Reads a file as the input (JSON by default unless --input-payload-meta is non-JSON encoding). Can be given multiple times for multiple arguments. Cannot be combined with --input.")
	f.StringArrayVar(&v.InputMeta, "input-meta", nil, "Metadata for the input payload. Expected as key=value. If key is encoding, overrides the default of json/plain.")
	f.BoolVar(&v.InputBase64, "input-base64", false, "If set, assumes --input or --input-file are base64 encoded and attempts to decode.")
}

type TemporalWorkflowStartCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	WorkflowStartOptions
	PayloadInputOptions
}

func NewTemporalWorkflowStartCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowStartCommand {
	var s TemporalWorkflowStartCommand
	s.Parent = parent
	s.Command.Use = "start"
	s.Command.Short = "Starts a new Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow start\x1b[0m command starts a new Workflow Execution. The\nWorkflow and Run IDs are returned after starting the Workflow.\n\n\x1b[1mtemporal workflow start \\\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m"
	} else {
		s.Command.Long = "The `temporal workflow start` command starts a new Workflow Execution. The\nWorkflow and Run IDs are returned after starting the Workflow.\n\n```\ntemporal workflow start \\\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowStartOptions.buildFlags(s.Command.Flags())
	s.PayloadInputOptions.buildFlags(s.Command.Flags())
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowTerminateCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowTerminateCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowTerminateCommand {
	var s TemporalWorkflowTerminateCommand
	s.Parent = parent
	s.Command.Use = "terminate"
	s.Command.Short = "Terminate Workflow Execution by ID or List Filter."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}

type TemporalWorkflowTraceCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
}

func NewTemporalWorkflowTraceCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowTraceCommand {
	var s TemporalWorkflowTraceCommand
	s.Parent = parent
	s.Command.Use = "trace"
	s.Command.Short = "Trace progress of a Workflow Execution and its children."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
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
	s.Command.Short = "Updates a running workflow synchronously."
	s.Command.Long = "TODO"
	s.Command.Args = cobra.NoArgs
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	cctx.BindConfigFlags(s.Command.Flags())
	return &s
}
