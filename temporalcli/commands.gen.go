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
	Command                 cobra.Command
	Env                     string
	EnvFile                 string
	LogLevel                StringEnum
	LogFormat               string
	Output                  StringEnum
	TimeFormat              StringEnum
	Color                   StringEnum
	NoJsonShorthandPayloads bool
}

func NewTemporalCommand(cctx *CommandContext) *TemporalCommand {
	var s TemporalCommand
	s.Command.Use = "temporal"
	s.Command.Short = "Temporal command-line interface and development server."
	s.Command.Long = ""
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalActivityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalBatchCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalEnvCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalScheduleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalServerCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalWorkflowCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVar(&s.Env, "env", "default", "Environment to read environment-specific flags from.")
	cctx.BindFlagEnvVar(s.Command.PersistentFlags().Lookup("env"), "TEMPORAL_ENV")
	s.Command.PersistentFlags().StringVar(&s.EnvFile, "env-file", "", "File to read all environments (defaults to `$HOME/.config/temporalio/temporal.yaml`).")
	s.LogLevel = NewStringEnum([]string{"debug", "info", "warn", "error", "never"}, "info")
	s.Command.PersistentFlags().Var(&s.LogLevel, "log-level", "Log level. Default is \"info\" for most commands and \"warn\" for `server start-dev`. Accepted values: debug, info, warn, error, never.")
	s.Command.PersistentFlags().StringVar(&s.LogFormat, "log-format", "", "Log format. Options are \"text\" and \"json\". Default is \"text\".")
	s.Output = NewStringEnum([]string{"text", "json", "jsonl", "none"}, "text")
	s.Command.PersistentFlags().VarP(&s.Output, "output", "o", "Data output format. Note, this does not affect logging. Accepted values: text, json, jsonl, none.")
	s.TimeFormat = NewStringEnum([]string{"relative", "iso", "raw"}, "relative")
	s.Command.PersistentFlags().Var(&s.TimeFormat, "time-format", "Time format. Accepted values: relative, iso, raw.")
	s.Color = NewStringEnum([]string{"always", "never", "auto"}, "auto")
	s.Command.PersistentFlags().Var(&s.Color, "color", "Set coloring. Accepted values: always, never, auto.")
	s.Command.PersistentFlags().BoolVar(&s.NoJsonShorthandPayloads, "no-json-shorthand-payloads", false, "Always show all payloads as raw payloads even if they are JSON.")
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
	s.Command.Short = "Complete or fail an Activity."
	s.Command.Long = ""
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
	Identity   string
	Result     string
}

func NewTemporalActivityCompleteCommand(cctx *CommandContext, parent *TemporalActivityCommand) *TemporalActivityCompleteCommand {
	var s TemporalActivityCompleteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "complete [flags]"
	s.Command.Short = "Complete an Activity."
	if hasHighlighting {
		s.Command.Long = "Complete an Activity.\n\n\x1b[1mtemporal activity complete --activity-id=MyActivityId --workflow-id=MyWorkflowId --result='{\"MyResultKey\": \"MyResultVal\"}'\x1b[0m"
	} else {
		s.Command.Long = "Complete an Activity.\n\n`temporal activity complete --activity-id=MyActivityId --workflow-id=MyWorkflowId --result='{\"MyResultKey\": \"MyResultVal\"}'`"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "The Activity to be completed.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Identity, "identity", "", "Identity of user submitting this request.")
	s.Command.Flags().StringVar(&s.Result, "result", "", "The result with which to complete the Activity (JSON).")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "result")
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
	s.Command.Short = "Fail an Activity."
	if hasHighlighting {
		s.Command.Long = "Fail an Activity.\n\n\x1b[1mtemporal activity fail --activity-id=MyActivityId --workflow-id=MyWorkflowId\x1b[0m"
	} else {
		s.Command.Long = "Fail an Activity.\n\n`temporal activity fail --activity-id=MyActivityId --workflow-id=MyWorkflowId`"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "The Activity to be failed.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Detail, "detail", "", "JSON data describing reason for failing the Activity.")
	s.Command.Flags().StringVar(&s.Identity, "identity", "", "Identity of user submitting this request.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for failing the Activity.")
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
	s.Command.Short = "Manage Batch Jobs"
	s.Command.Long = "Batch commands change multiple Workflow Executions."
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
	s.Command.Short = "Show Batch Job progress."
	if hasHighlighting {
		s.Command.Long = "The temporal batch describe command shows the progress of an ongoing Batch Job.\n\n\x1b[1mtemporal batch describe --job-id=MyJobId\x1b[0m"
	} else {
		s.Command.Long = "The temporal batch describe command shows the progress of an ongoing Batch Job.\n\n`temporal batch describe --job-id=MyJobId`"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.JobId, "job-id", "", "The Batch Job Id to describe.")
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
	s.Command.Short = "List all Batch Jobs"
	if hasHighlighting {
		s.Command.Long = "The temporal batch list command returns all Batch Jobs.\nBatch Jobs can be returned for an entire Cluster or a single Namespace.\n\n\x1b[1mtemporal batch list --namespace=MyNamespace\x1b[0m"
	} else {
		s.Command.Long = "The temporal batch list command returns all Batch Jobs.\nBatch Jobs can be returned for an entire Cluster or a single Namespace.\n\n`temporal batch list --namespace=MyNamespace`"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Limit the number of items to print.")
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
	s.Command.Short = "Terminate a Batch Job"
	if hasHighlighting {
		s.Command.Long = "The temporal batch terminate command terminates a Batch Job with the provided Job Id.\nFor future reference, provide a reason for terminating the Batch Job.\n\n\x1b[1mtemporal batch terminate --job-id=MyJobId --reason=JobReason\x1b[0m"
	} else {
		s.Command.Long = "The temporal batch terminate command terminates a Batch Job with the provided Job Id.\nFor future reference, provide a reason for terminating the Batch Job.\n\n`temporal batch terminate --job-id=MyJobId --reason=JobReason`"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.JobId, "job-id", "", "The Batch Job Id to terminate.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "job-id")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for terminating the Batch Job.")
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
	s.Command.Short = "Manage environments."
	s.Command.Long = "Use the '--env <env name>' option with other commands to point the CLI at a different Temporal Server instance. If --env\nis not passed, the 'default' environment is used."
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
	s.Command.Short = "Delete an environment or environment property."
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal env delete --env environment [-k property]\x1b[0m\n\nDelete an environment or just a single property:\n\n\x1b[1mtemporal env delete --env prod\x1b[0m\n\x1b[1mtemporal env delete --env prod -k tls-cert-path\x1b[0m\n\nIf the environment is not specified, the \x1b[1mdefault\x1b[0m environment is deleted:\n\n\x1b[1mtemporal env delete -k tls-cert-path\x1b[0m"
	} else {
		s.Command.Long = "`temporal env delete --env environment [-k property]`\n\nDelete an environment or just a single property:\n\n`temporal env delete --env prod`\n`temporal env delete --env prod -k tls-cert-path`\n\nIf the environment is not specified, the `default` environment is deleted:\n\n`temporal env delete -k tls-cert-path`"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVarP(&s.Key, "key", "k", "", "The name of the property.")
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
	s.Command.Short = "Print environment properties."
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal env get --env environment\x1b[0m\n\nPrint all properties of the 'prod' environment:\n\n\x1b[1mtemporal env get prod\x1b[0m\n\n\x1b[1mtls-cert-path  /home/my-user/certs/client.cert\ntls-key-path   /home/my-user/certs/client.key\naddress        temporal.example.com:7233\nnamespace      someNamespace\x1b[0m\n\nPrint a single property:\n\n\x1b[1mtemporal env get --env prod -k tls-key-path\x1b[0m\n\n\x1b[1mtls-key-path  /home/my-user/certs/cluster.key\x1b[0m\n\nIf the environment is not specified, the \x1b[1mdefault\x1b[0m environment is used."
	} else {
		s.Command.Long = "`temporal env get --env environment`\n\nPrint all properties of the 'prod' environment:\n\n`temporal env get prod`\n\n```\ntls-cert-path  /home/my-user/certs/client.cert\ntls-key-path   /home/my-user/certs/client.key\naddress        temporal.example.com:7233\nnamespace      someNamespace\n```\n\nPrint a single property:\n\n`temporal env get --env prod -k tls-key-path`\n\n```\ntls-key-path  /home/my-user/certs/cluster.key\n```\n\nIf the environment is not specified, the `default` environment is used."
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVarP(&s.Key, "key", "k", "", "The name of the property.")
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
	s.Command.Short = "Print all environments."
	s.Command.Long = "List all environments."
	s.Command.Args = cobra.NoArgs
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
	s.Command.Short = "Set environment properties."
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal env set --env environment -k property -v value\x1b[0m\n\nProperty names match CLI option names, for example '--address' and '--tls-cert-path':\n\n\x1b[1mtemporal env set --env prod -k address -v 127.0.0.1:7233\x1b[0m\n\x1b[1mtemporal env set --env prod -k tls-cert-path -v /home/my-user/certs/cluster.cert\x1b[0m\n\nIf the environment is not specified, the \x1b[1mdefault\x1b[0m environment is used."
	} else {
		s.Command.Long = "`temporal env set --env environment -k property -v value`\n\nProperty names match CLI option names, for example '--address' and '--tls-cert-path':\n\n`temporal env set --env prod -k address -v 127.0.0.1:7233`\n`temporal env set --env prod -k tls-cert-path -v /home/my-user/certs/cluster.cert`\n\nIf the environment is not specified, the `default` environment is used."
	}
	s.Command.Args = cobra.MaximumNArgs(2)
	s.Command.Flags().StringVarP(&s.Key, "key", "k", "", "The name of the property.")
	s.Command.Flags().StringVarP(&s.Value, "value", "v", "", "The value to set the property to.")
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
	s.Command.Short = "Manage a Temporal deployment."
	if hasHighlighting {
		s.Command.Long = "Operator commands enable actions on Namespaces, Search Attributes, and Temporal Clusters. These actions are performed through subcommands.\n\nTo run an Operator command, \x1b[1mrun temporal operator [command] [subcommand] [command options]\x1b[0m"
	} else {
		s.Command.Long = "Operator commands enable actions on Namespaces, Search Attributes, and Temporal Clusters. These actions are performed through subcommands.\n\nTo run an Operator command, `run temporal operator [command] [subcommand] [command options]`"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalOperatorClusterCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalOperatorNamespaceCommand(cctx, &s).Command)
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
	s.Command.Short = "Operations for running a Temporal Cluster."
	if hasHighlighting {
		s.Command.Long = "Cluster commands enable actions on Temporal Clusters.\n\nCluster commands follow this syntax: \x1b[1mtemporal operator cluster [command] [command options]\x1b[0m"
	} else {
		s.Command.Long = "Cluster commands enable actions on Temporal Clusters.\n\nCluster commands follow this syntax: `temporal operator cluster [command] [command options]`"
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
	s.Command.Short = "Describe a cluster"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator cluster describe\x1b[0m command shows information about the Cluster."
	} else {
		s.Command.Long = "`temporal operator cluster describe` command shows information about the Cluster."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.Detail, "detail", false, "Prints extra details.")
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
	s.Command.Short = "Checks the health of a cluster"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator cluster health\x1b[0m command checks the health of the Frontend Service."
	} else {
		s.Command.Long = "`temporal operator cluster health` command checks the health of the Frontend Service."
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
	s.Command.Short = "List all clusters"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator cluster list\x1b[0m command prints a list of all remote Clusters on the system."
	} else {
		s.Command.Long = "`temporal operator cluster list` command prints a list of all remote Clusters on the system."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Limit the number of items to print.")
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
	s.Command.Short = "Remove a cluster"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator cluster remove\x1b[0m command removes a remote Cluster from the system."
	} else {
		s.Command.Long = "`temporal operator cluster remove` command removes a remote Cluster from the system."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Name of cluster.")
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
	s.Command.Short = "Provide system info"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator cluster system\x1b[0m command provides information about the system the Cluster is running on. This information can be used to diagnose problems occurring in the Temporal Server."
	} else {
		s.Command.Long = "`temporal operator cluster system` command provides information about the system the Cluster is running on. This information can be used to diagnose problems occurring in the Temporal Server."
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
	s.Command.Short = "Add a remote"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator cluster upsert\x1b[0m command allows the user to add or update a remote Cluster."
	} else {
		s.Command.Long = "`temporal operator cluster upsert` command allows the user to add or update a remote Cluster."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.FrontendAddress, "frontend-address", "", "IP address to bind the frontend service to.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "frontend-address")
	s.Command.Flags().BoolVar(&s.EnableConnection, "enable-connection", false, "enable cross cluster connection.")
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
	s.Command.Short = "Operations performed on Namespaces."
	if hasHighlighting {
		s.Command.Long = "Namespace commands perform operations on Namespaces contained in the Temporal Cluster.\n\nCluster commands follow this syntax: \x1b[1mtemporal operator namespace [command] [command options]\x1b[0m"
	} else {
		s.Command.Long = "Namespace commands perform operations on Namespaces contained in the Temporal Cluster.\n\nCluster commands follow this syntax: `temporal operator namespace [command] [command options]`"
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
	Data                    string
	Description             string
	Email                   string
	Global                  bool
	HistoryArchivalState    StringEnum
	HistoryUri              string
	Retention               time.Duration
	VisibilityArchivalState StringEnum
	VisibilityUri           string
}

func NewTemporalOperatorNamespaceCreateCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceCreateCommand {
	var s TemporalOperatorNamespaceCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Registers a new Namespace."
	if hasHighlighting {
		s.Command.Long = "The temporal operator namespace create command creates a new Namespace on the Server.\nNamespaces can be created on the active Cluster, or any named Cluster.\n\x1b[1mtemporal operator namespace create --cluster=MyCluster -n example-1\x1b[0m\n\nGlobal Namespaces can also be created.\n\x1b[1mtemporal operator namespace create --global -n example-2\x1b[0m\n\nOther settings, such as retention and Visibility Archival State, can be configured as needed.\nFor example, the Visibility Archive can be set on a separate URI.\n\x1b[1mtemporal operator namespace create --retention=5 --visibility-archival-state=enabled --visibility-uri=some-uri -n example-3\x1b[0m"
	} else {
		s.Command.Long = "The temporal operator namespace create command creates a new Namespace on the Server.\nNamespaces can be created on the active Cluster, or any named Cluster.\n`temporal operator namespace create --cluster=MyCluster -n example-1`\n\nGlobal Namespaces can also be created.\n`temporal operator namespace create --global -n example-2`\n\nOther settings, such as retention and Visibility Archival State, can be configured as needed.\nFor example, the Visibility Archive can be set on a separate URI.\n`temporal operator namespace create --retention=5 --visibility-archival-state=enabled --visibility-uri=some-uri -n example-3`"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active cluster name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster names.")
	s.Command.Flags().StringVar(&s.Data, "data", "", "Namespace data in key=value format. Use JSON for values.")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.Global, "global", false, "Whether the namespace is a global namespace.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disabled")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Optionally specify history archival URI (cannot be changed after first time archival is enabled).")
	s.Command.Flags().DurationVar(&s.Retention, "retention", 259200000*time.Millisecond, "Length of time a closed Workflow is preserved before deletion.")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disabled")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled).")
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
	s.Command.Use = "delete [flags] [namespace]"
	s.Command.Short = "Deletes an existing Namespace."
	s.Command.Long = "The temporal operator namespace delete command deletes a given Namespace from the system."
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Confirm prompt to perform deletion.")
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
	s.Command.Use = "describe [flags] [namespace]"
	s.Command.Short = "Describe a Namespace by its name or ID."
	if hasHighlighting {
		s.Command.Long = "The temporal operator namespace describe command provides Namespace information.\nNamespaces are identified either by Namespace ID or by name.\n\n\x1b[1mtemporal operator namespace describe --namespace-id=some-namespace-id\x1b[0m\n\x1b[1mtemporal operator namespace describe -n example-namespace-name\x1b[0m"
	} else {
		s.Command.Long = "The temporal operator namespace describe command provides Namespace information.\nNamespaces are identified either by Namespace ID or by name.\n\n`temporal operator namespace describe --namespace-id=some-namespace-id`\n`temporal operator namespace describe -n example-namespace-name`"
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
	s.Command.Short = "List all Namespaces."
	s.Command.Long = "The temporal operator namespace list command lists all Namespaces on the Server."
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
	Retention               time.Duration
	VisibilityArchivalState StringEnum
	VisibilityUri           string
}

func NewTemporalOperatorNamespaceUpdateCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceUpdateCommand {
	var s TemporalOperatorNamespaceUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Updates a Namespace."
	if hasHighlighting {
		s.Command.Long = "The temporal operator namespace update command updates a Namespace.\n\nNamespaces can be assigned a different active Cluster.\n\x1b[1mtemporal operator namespace update -n namespace --active-cluster=NewActiveCluster\x1b[0m\n\nNamespaces can also be promoted to global Namespaces.\n\x1b[1mtemporal operator namespace update -n namespace --promote-global\x1b[0m\n\nAny Archives that were previously enabled or disabled can be changed through this command.\nHowever, URI values for archival states cannot be changed after the states are enabled.\n\x1b[1mtemporal operator namespace update -n namespace --history-archival-state=enabled --visibility-archival-state=disabled\x1b[0m"
	} else {
		s.Command.Long = "The temporal operator namespace update command updates a Namespace.\n\nNamespaces can be assigned a different active Cluster.\n`temporal operator namespace update -n namespace --active-cluster=NewActiveCluster`\n\nNamespaces can also be promoted to global Namespaces.\n`temporal operator namespace update -n namespace --promote-global`\n\nAny Archives that were previously enabled or disabled can be changed through this command.\nHowever, URI values for archival states cannot be changed after the states are enabled.\n`temporal operator namespace update -n namespace --history-archival-state=enabled --visibility-archival-state=disabled`"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active cluster name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster names.")
	s.Command.Flags().StringArrayVar(&s.Data, "data", nil, "Namespace data in key=value format. Use JSON for values.")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.PromoteGlobal, "promote-global", false, "Promote local namespace to global namespace.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Optionally specify history archival URI (cannot be changed after first time archival is enabled).")
	s.Command.Flags().DurationVar(&s.Retention, "retention", 0, "Length of time a closed Workflow is preserved before deletion.")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled).")
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
	s.Command.Short = "Operations applying to Search Attributes"
	s.Command.Long = "Search Attribute commands enable operations for the creation, listing, and removal of Search Attributes."
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
	Type    []string
}

func NewTemporalOperatorSearchAttributeCreateCommand(cctx *CommandContext, parent *TemporalOperatorSearchAttributeCommand) *TemporalOperatorSearchAttributeCreateCommand {
	var s TemporalOperatorSearchAttributeCreateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "create [flags]"
	s.Command.Short = "Adds one or more custom Search Attributes"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator search-attribute create\x1b[0m command adds one or more custom Search Attributes."
	} else {
		s.Command.Long = "`temporal operator search-attribute create` command adds one or more custom Search Attributes."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Name, "name", nil, "Search Attribute name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringArrayVar(&s.Type, "type", nil, "Search Attribute type. Accepted values: Text, Keyword, Int, Double, Bool, Datetime, KeywordList.")
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
	s.Command.Short = "Lists all Search Attributes that can be used in list Workflow Queries"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator search-attribute list\x1b[0m displays a list of all Search Attributes."
	} else {
		s.Command.Long = "`temporal operator search-attribute list` displays a list of all Search Attributes."
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
	s.Command.Short = "Removes custom search attribute metadata only"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal operator search-attribute remove\x1b[0m command removes custom Search Attribute metadata."
	} else {
		s.Command.Long = "`temporal operator search-attribute remove` command removes custom Search Attribute metadata."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Name, "name", nil, "Search Attribute name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Confirm prompt to perform deletion.")
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
	s.Command.Short = "Perform operations on Schedules."
	s.Command.Long = "Schedule commands allow the user to create, use, and update Schedules.\nSchedules allow starting Workflow Execution at regular times."
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

type OverlapPolicyOptions struct {
	OverlapPolicy StringEnum
}

func (v *OverlapPolicyOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	v.OverlapPolicy = NewStringEnum([]string{"Skip", "BufferOne", "BufferAll", "CancelOther", "TerminateOther", "AllowAll"}, "Skip")
	f.Var(&v.OverlapPolicy, "overlap-policy", "Overlap policy. Accepted values: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll.")
}

type ScheduleIdOptions struct {
	ScheduleId string
}

func (v *ScheduleIdOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.ScheduleId, "schedule-id", "s", "", "Schedule id.")
	_ = cobra.MarkFlagRequired(f, "schedule-id")
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
	s.Command.Short = "Backfills a past time range of actions."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal schedule backfill\x1b[0m command runs the Actions that would have been run in a given time\ninterval, all at once.\n\n You can use backfill to fill in Workflow Runs from a time period when the Schedule was paused, from\nbefore the Schedule was created, from the future, or to re-process an interval that was processed.\n\nSchedule backfills require a Schedule ID, along with the time in which to run the Schedule. You can\noptionally override the overlap policy. It usually only makes sense to run backfills with either\n\x1b[1mBufferAll\x1b[0m or \x1b[1mAllowAll\x1b[0m (other policies will only let one or two runs actually happen).\n\nExample:\n\n\x1b[1m  temporal schedule backfill           \\\n    --schedule-id 'your-schedule-id'   \\\n    --overlap-policy BufferAll         \\\n    --start-time 2022-05-01T00:00:00Z  \\\n    --end-time   2022-05-31T23:59:59Z\x1b[0m"
	} else {
		s.Command.Long = "The `temporal schedule backfill` command runs the Actions that would have been run in a given time\ninterval, all at once.\n\n You can use backfill to fill in Workflow Runs from a time period when the Schedule was paused, from\nbefore the Schedule was created, from the future, or to re-process an interval that was processed.\n\nSchedule backfills require a Schedule ID, along with the time in which to run the Schedule. You can\noptionally override the overlap policy. It usually only makes sense to run backfills with either\n`BufferAll` or `AllowAll` (other policies will only let one or two runs actually happen).\n\nExample:\n\n```\n  temporal schedule backfill           \\\n    --schedule-id 'your-schedule-id'   \\\n    --overlap-policy BufferAll         \\\n    --start-time 2022-05-01T00:00:00Z  \\\n    --end-time   2022-05-31T23:59:59Z\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.OverlapPolicyOptions.buildFlags(cctx, s.Command.Flags())
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().Var(&s.EndTime, "end-time", "Backfill end time.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "end-time")
	s.Command.Flags().Var(&s.StartTime, "start-time", "Backfill start time.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "start-time")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type ScheduleConfigurationOptions struct {
	Calendar                []string
	CatchupWindow           time.Duration
	Cron                    []string
	EndTime                 Timestamp
	Interval                []string
	Jitter                  time.Duration
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
	f.StringArrayVar(&v.Calendar, "calendar", nil, "Calendar specification in JSON, e.g. `{\"dayOfWeek\":\"Fri\",\"hour\":\"17\",\"minute\":\"5\"}`.")
	f.DurationVar(&v.CatchupWindow, "catchup-window", 0, "Maximum allowed catch-up time if server is down.")
	f.StringArrayVar(&v.Cron, "cron", nil, "Calendar spec in cron string format, e.g. `3 11 * * Fri`.")
	f.Var(&v.EndTime, "end-time", "Overall schedule end time.")
	f.StringArrayVar(&v.Interval, "interval", nil, "Interval duration, e.g. 90m, or 90m/13m to include phase offset.")
	f.DurationVar(&v.Jitter, "jitter", 0, "Per-action jitter range.")
	f.StringVar(&v.Notes, "notes", "", "Initial value of notes field.")
	f.BoolVar(&v.Paused, "paused", false, "Initial value of paused state.")
	f.BoolVar(&v.PauseOnFailure, "pause-on-failure", false, "Pause schedule after any workflow failure.")
	f.IntVar(&v.RemainingActions, "remaining-actions", 0, "Total number of actions allowed. Zero (default) means unlimited.")
	f.Var(&v.StartTime, "start-time", "Overall schedule start time.")
	f.StringVar(&v.TimeZone, "time-zone", "", "Time zone to interpret all calendar specs in (IANA name).")
	f.StringArrayVar(&v.ScheduleSearchAttribute, "schedule-search-attribute", nil, "Search Attribute for the _schedule_ in key=value format. Use valid JSON formats for value.")
	f.StringArrayVar(&v.ScheduleMemo, "schedule-memo", nil, "Memo for the _schedule_ in key=value format. Use valid JSON formats for value.")
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
	s.Command.Short = "Create a new Schedule."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal schedule create\x1b[0m command creates a new Schedule.\n\nExample:\n\n\x1b[1m  temporal schedule create                                    \\\n    --schedule-id 'your-schedule-id'                          \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"11\"}' \\\n    --workflow-id 'your-base-workflow-id'                     \\\n    --task-queue 'your-task-queue'                            \\\n    --workflow-type 'YourWorkflowType'\x1b[0m\n\nAny combination of \x1b[1m--calendar\x1b[0m, \x1b[1m--interval\x1b[0m, and \x1b[1m--cron\x1b[0m is supported.\nActions will be executed at any time specified in the Schedule."
	} else {
		s.Command.Long = "The `temporal schedule create` command creates a new Schedule.\n\nExample:\n\n```\n  temporal schedule create                                    \\\n    --schedule-id 'your-schedule-id'                          \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"11\"}' \\\n    --workflow-id 'your-base-workflow-id'                     \\\n    --task-queue 'your-task-queue'                            \\\n    --workflow-type 'YourWorkflowType'\n```\n\nAny combination of `--calendar`, `--interval`, and `--cron` is supported.\nActions will be executed at any time specified in the Schedule."
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleConfigurationOptions.buildFlags(cctx, s.Command.Flags())
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.OverlapPolicyOptions.buildFlags(cctx, s.Command.Flags())
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
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
	s.Command.Short = "Deletes a Schedule."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal schedule delete\x1b[0m command deletes a Schedule.\nDeleting a Schedule does not affect any Workflows started by the Schedule.\n\nIf you do also want to cancel or terminate Workflows started by a Schedule, consider using \x1b[1mtemporal\nworkflow delete\x1b[0m with the \x1b[1mTemporalScheduledById\x1b[0m Search Attribute."
	} else {
		s.Command.Long = "The `temporal schedule delete` command deletes a Schedule.\nDeleting a Schedule does not affect any Workflows started by the Schedule.\n\nIf you do also want to cancel or terminate Workflows started by a Schedule, consider using `temporal\nworkflow delete` with the `TemporalScheduledById` Search Attribute."
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
	s.Command.Short = "Get Schedule configuration and current state."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal schedule describe\x1b[0m command shows the current configuration of one Schedule,\nincluding information about past, current, and future Workflow Runs."
	} else {
		s.Command.Long = "The `temporal schedule describe` command shows the current configuration of one Schedule,\nincluding information about past, current, and future Workflow Runs."
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
}

func NewTemporalScheduleListCommand(cctx *CommandContext, parent *TemporalScheduleCommand) *TemporalScheduleListCommand {
	var s TemporalScheduleListCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "list [flags]"
	s.Command.Short = "Lists Schedules."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal schedule list\x1b[0m command lists all Schedules in a namespace."
	} else {
		s.Command.Long = "The `temporal schedule list` command lists all Schedules in a namespace."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVarP(&s.Long, "long", "l", false, "Include detailed information.")
	s.Command.Flags().BoolVar(&s.ReallyLong, "really-long", false, "Include even more detailed information that's not really usable in table form.")
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
	s.Command.Short = "Pauses or unpauses a Schedule."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal schedule toggle\x1b[0m command can pause and unpause a Schedule.\n\nToggling a Schedule takes a reason. The reason will be set as the \x1b[1mnotes\x1b[0m field of the Schedule,\nto help with operations communication.\n\nExamples:\n\n* \x1b[1mtemporal schedule toggle --schedule-id 'your-schedule-id' --pause --reason \"paused because the database is down\"\x1b[0m\n* \x1b[1mtemporal schedule toggle --schedule-id 'your-schedule-id' --unpause --reason \"the database is back up\"\x1b[0m"
	} else {
		s.Command.Long = "The `temporal schedule toggle` command can pause and unpause a Schedule.\n\nToggling a Schedule takes a reason. The reason will be set as the `notes` field of the Schedule,\nto help with operations communication.\n\nExamples:\n\n* `temporal schedule toggle --schedule-id 'your-schedule-id' --pause --reason \"paused because the database is down\"`\n* `temporal schedule toggle --schedule-id 'your-schedule-id' --unpause --reason \"the database is back up\"`"
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().BoolVar(&s.Pause, "pause", false, "Pauses the schedule.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "\"(no reason provided)\"", "Reason for pausing/unpausing.")
	s.Command.Flags().BoolVar(&s.Unpause, "unpause", false, "Pauses the schedule.")
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
	s.Command.Short = "Triggers a schedule to take an action immediately."
	s.Command.Long = ""
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
	s.Command.Short = "Updates a Schedule with a new definition."
	s.Command.Long = "The temporal schedule update command updates an existing Schedule. It replaces the entire\nconfiguration of the schedule, including spec, action, and policies."
	s.Command.Args = cobra.NoArgs
	s.ScheduleConfigurationOptions.buildFlags(cctx, s.Command.Flags())
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.OverlapPolicyOptions.buildFlags(cctx, s.Command.Flags())
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
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
	s.Command.Short = "Run Temporal Server."
	if hasHighlighting {
		s.Command.Long = "Start a development version of Temporal Server:\n\n\x1b[1mtemporal server start-dev\x1b[0m"
	} else {
		s.Command.Long = "Start a development version of Temporal Server:\n\n`temporal server start-dev`"
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
	UiAssetPath        string
	UiCodecEndpoint    string
	SqlitePragma       []string
	DynamicConfigValue []string
	LogConfig          bool
}

func NewTemporalServerStartDevCommand(cctx *CommandContext, parent *TemporalServerCommand) *TemporalServerStartDevCommand {
	var s TemporalServerStartDevCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "start-dev [flags]"
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
	s.Command.Flags().StringVar(&s.Ip, "ip", "localhost", "IP address to bind the frontend service to.")
	s.Command.Flags().StringVar(&s.UiIp, "ui-ip", "", "IP address to bind the Web UI to. Default is same as --ip.")
	s.Command.Flags().StringVar(&s.UiAssetPath, "ui-asset-path", "", "UI custom assets path.")
	s.Command.Flags().StringVar(&s.UiCodecEndpoint, "ui-codec-endpoint", "", "UI remote codec HTTP endpoint.")
	s.Command.Flags().StringArrayVar(&s.SqlitePragma, "sqlite-pragma", nil, "Specify SQLite pragma statements in pragma=value format.")
	s.Command.Flags().StringArrayVar(&s.DynamicConfigValue, "dynamic-config-value", nil, "Dynamic config value, as KEY=JSON_VALUE (string values need quotes).")
	s.Command.Flags().BoolVar(&s.LogConfig, "log-config", false, "Log the server config being used to stderr.")
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
	s.Command.Short = "Manage Task Queues."
	if hasHighlighting {
		s.Command.Long = "Task Queue commands allow operations to be performed on Task Queues. To run a Task\nQueue command, run \x1b[1mtemporal task-queue [command] [command options]\x1b[0m."
	} else {
		s.Command.Long = "Task Queue commands allow operations to be performed on Task Queues. To run a Task\nQueue command, run `temporal task-queue [command] [command options]`."
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
	Parent        *TemporalTaskQueueCommand
	Command       cobra.Command
	TaskQueue     string
	TaskQueueType StringEnum
	Partitions    int
}

func NewTemporalTaskQueueDescribeCommand(cctx *CommandContext, parent *TemporalTaskQueueCommand) *TemporalTaskQueueDescribeCommand {
	var s TemporalTaskQueueDescribeCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "describe [flags]"
	s.Command.Short = "Provides information for Workers that have recently polled on this Task Queue."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal task-queue describe\x1b[0m command provides poller\ninformation for a given Task Queue.\n\nThe Server records the last time of each poll request. A \x1b[1mLastAccessTime\x1b[0m value\nin excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the\nWorker has shut down. Workers are removed if 5 minutes have passed since the last poll\nrequest.\n\nInformation about the Task Queue can be returned to troubleshoot server issues.\n\n\x1b[1mtemporal task-queue describe --task-queue=MyTaskQueue --task-queue-type=\"activity\"\x1b[0m\n\nUse the options listed below to modify what this command returns."
	} else {
		s.Command.Long = "The `temporal task-queue describe` command provides poller\ninformation for a given Task Queue.\n\nThe Server records the last time of each poll request. A `LastAccessTime` value\nin excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the\nWorker has shut down. Workers are removed if 5 minutes have passed since the last poll\nrequest.\n\nInformation about the Task Queue can be returned to troubleshoot server issues.\n\n`temporal task-queue describe --task-queue=MyTaskQueue --task-queue-type=\"activity\"`\n\nUse the options listed below to modify what this command returns."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task queue name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.TaskQueueType = NewStringEnum([]string{"workflow", "activity"}, "workflow")
	s.Command.Flags().Var(&s.TaskQueueType, "task-queue-type", "Task Queue type. Accepted values: workflow, activity.")
	s.Command.Flags().IntVar(&s.Partitions, "partitions", 1, "Query for all partitions up to this number (experimental+temporary feature).")
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
	s.Command.Short = "Retrieves information about the reachability of Build IDs on one or more Task Queues (Deprecated)."
	s.Command.Long = "This command can tell you whether or not Build IDs may be used for new, existing, or closed workflows. Both the '--build-id' and '--task-queue' flags may be specified multiple times. If you do not provide a task queue, reachability for the provided Build IDs will be checked against all task queues."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.BuildId, "build-id", nil, "Which Build ID to get reachability information for. May be specified multiple times.")
	s.ReachabilityType = NewStringEnum([]string{"open", "closed", "existing"}, "existing")
	s.Command.Flags().Var(&s.ReachabilityType, "reachability-type", "Specify how you'd like to filter the reachability of Build IDs. Valid choices are `open` (reachable by one or more open workflows), `closed` (reachable by one or more closed workflows), or `existing` (reachable by either). If a Build ID is reachable by new workflows, that is always reported. Accepted values: open, closed, existing.")
	s.Command.Flags().StringArrayVarP(&s.TaskQueue, "task-queue", "t", nil, "Which Task Queue(s) to constrain the reachability search to. May be specified multiple times.")
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
	s.Command.Short = "Fetch the sets of worker Build ID versions on the Task Queue (Deprecated)."
	s.Command.Long = "Fetch the sets of compatible build IDs associated with a Task Queue and associated information."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task queue name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Flags().IntVar(&s.MaxSets, "max-sets", 0, "Limits how many compatible sets will be returned. Specify 1 to only return the current default major version set. 0 returns all sets. (default: 0).")
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
	s.Command.Short = "Lists the Task Queue's partitions and the matching nodes they are assigned to."
	s.Command.Long = "The temporal task-queue list-partition command displays the partitions of a Task Queue, along with the matching node they are assigned to."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task queue name.")
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
	s.Command.Short = "Operations to update the sets of worker Build ID versions on the Task Queue (Deprecated)."
	s.Command.Long = "Provides various commands for adding or changing the sets of compatible build IDs associated with a Task Queue. See the help of each sub-command for more."
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
	s.Command.Short = "Add a new build ID compatible with an existing ID to the Task Queue version sets."
	s.Command.Long = "The new build ID will become the default for the set containing the existing ID. See per-flag help for more."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "The new build id to be added.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Flags().StringVar(&s.ExistingCompatibleBuildId, "existing-compatible-build-id", "", "A build id which must already exist in the version sets known by the task queue. The new id will be stored in the set containing this id, marking it as compatible with the versions within.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "existing-compatible-build-id")
	s.Command.Flags().BoolVar(&s.SetAsDefault, "set-as-default", false, "When set, establishes the compatible set being targeted as the overall default for the queue. If a different set was the current default, the targeted set will replace it as the new default. Defaults to false.")
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
	s.Command.Short = "Add a new default (incompatible) build ID to the Task Queue version sets."
	s.Command.Long = "Creates a new build id set which will become the new overall default for the queue with the provided build id as its only member. This new set is incompatible with all previous sets/versions."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "The new build id to be added.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue.")
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
	s.Command.Short = "Promote an existing build ID to become the default for its containing set."
	s.Command.Long = "New tasks compatible with the set will be dispatched to the default id."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "An existing build id which will be promoted to be the default inside its containing set.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue.")
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
	s.Command.Short = "Promote an existing build ID set to become the default for the Task Queue."
	s.Command.Long = "If the set is already the default, this command has no effect."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "An existing build id whose containing set will be promoted.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue.")
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
	s.Command.Short = "Updates or retrieves the worker Build ID assignment and redirect rules on the Task Queue."
	s.Command.Long = "Provides various commands for adding, listing, removing, or replacing worker Build ID assignment and redirect rules associated with a Task Queue. See the help of each sub-command for more."
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningAddRedirectRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningCommitBuildIdCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningDeleteAssignmentRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningDeleteRedirectRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningGetRulesCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningInsertAssignmentRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningReplaceAssignmentRuleCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueVersioningReplaceRedirectRuleCommand(cctx, &s).Command)
	s.Command.PersistentFlags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task queue name.")
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
	s.Command.Short = "Adds the rule to the list of redirect rules for this Task Queue."
	s.Command.Long = "Adds a new redirect rule for this Task Queue. There can be at most one redirect rule for each distinct source build ID."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.SourceBuildId, "source-build-id", "", "The source build ID for this redirect rule.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source-build-id")
	s.Command.Flags().StringVar(&s.TargetBuildId, "target-build-id", "", "The target build ID for this redirect rule.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "target-build-id")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Skip confirmation.")
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
	s.Command.Short = "Completes the rollout of a Build ID for this Task Queue."
	if hasHighlighting {
		s.Command.Long = "Completes  the rollout of a BuildID and cleanup unnecessary rules possibly\ncreated during a gradual rollout. Specifically, this command will make the\nfollowing changes atomically:\n\t1. Adds an unconditional assignment rule for the target Build ID at the end of the list.\n\t2. Removes all previously added assignment rules to the given target Build ID.\n\t3. Removes any unconditional assignment rules for other Build IDs.\n\nTo prevent committing invalid Build IDs, we reject the request if no pollers\nhave been seen recently for this Build ID. Use the \x1b[1mforce\x1b[0m option to disable this validation."
	} else {
		s.Command.Long = "Completes  the rollout of a BuildID and cleanup unnecessary rules possibly\ncreated during a gradual rollout. Specifically, this command will make the\nfollowing changes atomically:\n\t1. Adds an unconditional assignment rule for the target Build ID at the end of the list.\n\t2. Removes all previously added assignment rules to the given target Build ID.\n\t3. Removes any unconditional assignment rules for other Build IDs.\n\nTo prevent committing invalid Build IDs, we reject the request if no pollers\nhave been seen recently for this Build ID. Use the `force` option to disable this validation."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "The target build ID to be committed.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().BoolVar(&s.Force, "force", false, "Bypass the validation that pollers have been recently seen for this build ID.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Skip confirmation.")
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
	Yes       bool
	Force     bool
}

func NewTemporalTaskQueueVersioningDeleteAssignmentRuleCommand(cctx *CommandContext, parent *TemporalTaskQueueVersioningCommand) *TemporalTaskQueueVersioningDeleteAssignmentRuleCommand {
	var s TemporalTaskQueueVersioningDeleteAssignmentRuleCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "delete-assignment-rule [flags]"
	s.Command.Short = "Deletes the rule at a given index in the list of assignment rules for this Task Queue."
	if hasHighlighting {
		s.Command.Long = "Deletes an assignment rule for this Task Queue. By default presence of one\nunconditional rule, i.e., no hint filter or percentage, is enforced, otherwise\nthe delete operation will be rejected. Set \x1b[1mforce\x1b[0m to true to bypass this\nvalidation."
	} else {
		s.Command.Long = "Deletes an assignment rule for this Task Queue. By default presence of one\nunconditional rule, i.e., no hint filter or percentage, is enforced, otherwise\nthe delete operation will be rejected. Set `force` to true to bypass this\nvalidation."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Position of the assignment rule to be replaced.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "rule-index")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Skip confirmation.")
	s.Command.Flags().BoolVar(&s.Force, "force", false, "Bypass the validation that one unconditional rule remains.")
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
	s.Command.Short = "Deletes the rule with the given build ID for this Task Queue."
	s.Command.Long = "Deletes the routing rule with the given source Build ID."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.SourceBuildId, "source-build-id", "", "The source build ID for this redirect rule.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source-build-id")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Skip confirmation.")
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
	s.Command.Short = "Retrieves the worker Build ID assignment and redirect rules on the Task Queue."
	s.Command.Long = "Fetch the worker build ID assignment and redirect rules associated with a Task Queue."
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
	s.Command.Short = "Inserts the rule to the list of assignment rules for this Task Queue."
	s.Command.Long = "Inserts a new assignment rule for this Task Queue. The rules are evaluated in order, starting from index 0. The first applicable rule will be applied and the rest will be ignored."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "The target build ID for this assignment rule.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Insertion position in the assignment rule list. An index 0 means insert at the beginning of the list. If the given index is larger than the list size, the rule will be appended at the end of the list.")
	s.Command.Flags().IntVar(&s.Percentage, "percentage", 100, "Percentage of traffic sent to the target build ID.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Skip confirmation.")
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
	s.Command.Short = "Replaces the rule at a given index in the list of assignment rules for this Task Queue."
	if hasHighlighting {
		s.Command.Long = "Replaces an assignment rule for this Task Queue. By default presence of one\nunconditional rule, i.e., no hint filter or percentage, is enforced, otherwise\nthe delete operation will be rejected. Set \x1b[1mforce\x1b[0m to true to bypass this\nvalidation."
	} else {
		s.Command.Long = "Replaces an assignment rule for this Task Queue. By default presence of one\nunconditional rule, i.e., no hint filter or percentage, is enforced, otherwise\nthe delete operation will be rejected. Set `force` to true to bypass this\nvalidation."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "The target build ID for this assignment rule.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "build-id")
	s.Command.Flags().IntVarP(&s.RuleIndex, "rule-index", "i", 0, "Position of the assignment rule to be replaced.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "rule-index")
	s.Command.Flags().IntVar(&s.Percentage, "percentage", 100, "Percentage of traffic sent to the target build ID.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Skip confirmation.")
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
	s.Command.Short = "Replaces the redirect rule with the given source build ID for this Task Queue."
	s.Command.Long = "Replaces the redirect rule with the given source build ID for this Task Queue."
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.SourceBuildId, "source-build-id", "", "The source build ID for this redirect rule.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source-build-id")
	s.Command.Flags().StringVar(&s.TargetBuildId, "target-build-id", "", "The target build ID for this redirect rule.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "target-build-id")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Skip confirmation.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type ClientOptions struct {
	Address                    string
	Namespace                  string
	ApiKey                     string
	GrpcMeta                   []string
	Tls                        bool
	TlsCertPath                string
	TlsKeyPath                 string
	TlsCaPath                  string
	TlsCertData                string
	TlsKeyData                 string
	TlsCaData                  string
	TlsDisableHostVerification bool
	TlsServerName              string
	CodecEndpoint              string
	CodecAuth                  string
}

func (v *ClientOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Address, "address", "127.0.0.1:7233", "Temporal server address.")
	cctx.BindFlagEnvVar(f.Lookup("address"), "TEMPORAL_ADDRESS")
	f.StringVarP(&v.Namespace, "namespace", "n", "default", "Temporal server namespace.")
	cctx.BindFlagEnvVar(f.Lookup("namespace"), "TEMPORAL_NAMESPACE")
	f.StringVar(&v.ApiKey, "api-key", "", "Sets the API key on requests.")
	cctx.BindFlagEnvVar(f.Lookup("api-key"), "TEMPORAL_API_KEY")
	f.StringArrayVar(&v.GrpcMeta, "grpc-meta", nil, "HTTP headers to send with requests (formatted as key=value).")
	f.BoolVar(&v.Tls, "tls", false, "Enable TLS encryption without additional options such as mTLS or client certificates.")
	cctx.BindFlagEnvVar(f.Lookup("tls"), "TEMPORAL_TLS")
	f.StringVar(&v.TlsCertPath, "tls-cert-path", "", "Path to x509 certificate.")
	cctx.BindFlagEnvVar(f.Lookup("tls-cert-path"), "TEMPORAL_TLS_CERT")
	f.StringVar(&v.TlsKeyPath, "tls-key-path", "", "Path to private certificate key.")
	cctx.BindFlagEnvVar(f.Lookup("tls-key-path"), "TEMPORAL_TLS_KEY")
	f.StringVar(&v.TlsCaPath, "tls-ca-path", "", "Path to server CA certificate.")
	cctx.BindFlagEnvVar(f.Lookup("tls-ca-path"), "TEMPORAL_TLS_CA")
	f.StringVar(&v.TlsCertData, "tls-cert-data", "", "Data for x509 certificate. Exclusive with -path variant.")
	cctx.BindFlagEnvVar(f.Lookup("tls-cert-data"), "TEMPORAL_TLS_CERT_DATA")
	f.StringVar(&v.TlsKeyData, "tls-key-data", "", "Data for private certificate key. Exclusive with -path variant.")
	cctx.BindFlagEnvVar(f.Lookup("tls-key-data"), "TEMPORAL_TLS_KEY_DATA")
	f.StringVar(&v.TlsCaData, "tls-ca-data", "", "Data for server CA certificate. Exclusive with -path variant.")
	cctx.BindFlagEnvVar(f.Lookup("tls-ca-data"), "TEMPORAL_TLS_CA_DATA")
	f.BoolVar(&v.TlsDisableHostVerification, "tls-disable-host-verification", false, "Disables TLS host-name verification.")
	cctx.BindFlagEnvVar(f.Lookup("tls-disable-host-verification"), "TEMPORAL_TLS_DISABLE_HOST_VERIFICATION")
	f.StringVar(&v.TlsServerName, "tls-server-name", "", "Overrides target TLS server name.")
	cctx.BindFlagEnvVar(f.Lookup("tls-server-name"), "TEMPORAL_TLS_SERVER_NAME")
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "Endpoint for a remote Codec Server.")
	cctx.BindFlagEnvVar(f.Lookup("codec-endpoint"), "TEMPORAL_CODEC_ENDPOINT")
	f.StringVar(&v.CodecAuth, "codec-auth", "", "Sets the authorization header on requests to the Codec Server.")
	cctx.BindFlagEnvVar(f.Lookup("codec-auth"), "TEMPORAL_CODEC_AUTH")
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
		s.Command.Long = "Workflow commands perform operations on Workflow Executions.\n\nWorkflow commands use this syntax: \x1b[1mtemporal workflow COMMAND [ARGS]\x1b[0m."
	} else {
		s.Command.Long = "Workflow commands perform operations on Workflow Executions.\n\nWorkflow commands use this syntax: `temporal workflow COMMAND [ARGS]`."
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
	s.Command.Short = "Cancel a Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow cancel\x1b[0m command is used to cancel a Workflow Execution.\nCanceling a running Workflow Execution records a \x1b[1mWorkflowExecutionCancelRequested\x1b[0m event in the Event History. A new\nCommand Task will be scheduled, and the Workflow Execution will perform cleanup work.\n\nExecutions may be cancelled by ID:\n\x1b[1mtemporal workflow cancel --workflow-id MyWorkflowId\x1b[0m\n\n...or in bulk via a visibility query list filter:\n\x1b[1mtemporal workflow cancel --query=MyQuery\x1b[0m\n\nUse the options listed below to change the behavior of this command."
	} else {
		s.Command.Long = "The `temporal workflow cancel` command is used to cancel a Workflow Execution.\nCanceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History. A new\nCommand Task will be scheduled, and the Workflow Execution will perform cleanup work.\n\nExecutions may be cancelled by ID:\n```\ntemporal workflow cancel --workflow-id MyWorkflowId\n```\n\n...or in bulk via a visibility query list filter:\n```\ntemporal workflow cancel --query=MyQuery\n```\n\nUse the options listed below to change the behavior of this command."
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
	s.Command.Short = "Count Workflow Executions."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow count\x1b[0m command returns a count of Workflow Executions.\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow count` command returns a count of Workflow Executions.\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Filter results using a SQL-like query.")
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
	s.Command.Short = "Deletes a Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow delete\x1b[0m command is used to delete a specific Workflow Execution.\nThis asynchronously deletes a workflow's Event History.\nIf the Workflow Execution is Running, it will be terminated before deletion.\n\n\x1b[1mtemporal workflow delete \\\n\t\t--workflow-id MyWorkflowId \\\x1b[0m\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow delete` command is used to delete a specific Workflow Execution.\nThis asynchronously deletes a workflow's Event History.\nIf the Workflow Execution is Running, it will be terminated before deletion.\n\n```\ntemporal workflow delete \\\n\t\t--workflow-id MyWorkflowId \\\n```\n\nUse the options listed below to change the command's behavior."
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

type WorkflowReferenceOptions struct {
	WorkflowId string
	RunId      string
}

func (v *WorkflowReferenceOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow Id.")
	_ = cobra.MarkFlagRequired(f, "workflow-id")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run Id.")
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
	s.Command.Short = "Show information about a Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow describe\x1b[0m command shows information about a given\nWorkflow Execution.\n\nThis information can be used to locate Workflow Executions that weren't able to run successfully.\n\n\x1b[1mtemporal workflow describe --workflow-id=meaningful-business-id\x1b[0m\n\nOutput can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.\n\n\x1b[1mtemporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true\x1b[0m\n\nUse the command options below to change the information returned by this command."
	} else {
		s.Command.Long = "The `temporal workflow describe` command shows information about a given\nWorkflow Execution.\n\nThis information can be used to locate Workflow Executions that weren't able to run successfully.\n\n`temporal workflow describe --workflow-id=meaningful-business-id`\n\nOutput can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.\n\n`temporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true`\n\nUse the command options below to change the information returned by this command."
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().BoolVar(&s.ResetPoints, "reset-points", false, "Only show auto-reset points.")
	s.Command.Flags().BoolVar(&s.Raw, "raw", false, "Print properties without changing their format.")
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
	EventDetails bool
}

func NewTemporalWorkflowExecuteCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowExecuteCommand {
	var s TemporalWorkflowExecuteCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "execute [flags]"
	s.Command.Short = "Start a new Workflow Execution and prints its progress."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow execute\x1b[0m command starts a new Workflow Execution and\nprints its progress. The command completes when the Workflow Execution completes.\n\nSingle quotes('') are used to wrap input as JSON.\n\n\x1b[1mtemporal workflow execute\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m"
	} else {
		s.Command.Long = "The `temporal workflow execute` command starts a new Workflow Execution and\nprints its progress. The command completes when the Workflow Execution completes.\n\nSingle quotes('') are used to wrap input as JSON.\n\n```\ntemporal workflow execute\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().BoolVar(&s.EventDetails, "event-details", false, "If set when using text output, this will print the event details instead of just the event during workflow progress. If set when using JSON output, this will include the entire \"history\" JSON key of the started run (does not follow runs).")
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
	s.Command.Short = "Updates an event history JSON file to the current format."
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal workflow fix-history-json \\\n\t--source original.json \\\n\t--target reserialized.json\x1b[0m\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "```\ntemporal workflow fix-history-json \\\n\t--source original.json \\\n\t--target reserialized.json\n```\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Source, "source", "s", "", "Path to the input file.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "source")
	s.Command.Flags().StringVarP(&s.Target, "target", "t", "", "Path to the output file, or standard output if not set.")
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
	s.Command.Short = "List Workflow Executions based on a Query."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow list\x1b[0m command provides a list of Workflow Executions\nthat meet the criteria of a given Query.\nBy default, this command returns up to 10 closed Workflow Executions.\n\n\x1b[1mtemporal workflow list --query=MyQuery\x1b[0m\n\nThe command can also return a list of archived Workflow Executions.\n\n\x1b[1mtemporal workflow list --archived\x1b[0m\n\nUse the command options below to change the information returned by this command."
	} else {
		s.Command.Long = "The `temporal workflow list` command provides a list of Workflow Executions\nthat meet the criteria of a given Query.\nBy default, this command returns up to 10 closed Workflow Executions.\n\n`temporal workflow list --query=MyQuery`\n\nThe command can also return a list of archived Workflow Executions.\n\n`temporal workflow list --archived`\n\nUse the command options below to change the information returned by this command."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Filter results using a SQL-like query.")
	s.Command.Flags().BoolVar(&s.Archived, "archived", false, "If set, will only query and list archived workflows instead of regular workflows.")
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Limit the number of items to print.")
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
	Type            string
	RejectCondition StringEnum
}

func NewTemporalWorkflowQueryCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowQueryCommand {
	var s TemporalWorkflowQueryCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "query [flags]"
	s.Command.Short = "Query a Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow query\x1b[0m command is used to Query a\nWorkflow Execution\nby ID.\n\n\x1b[1mtemporal workflow query \\\n\t\t--workflow-id MyWorkflowId \\\n\t\t--name MyQuery \\\n\t\t--input '{\"MyInputKey\": \"MyInputValue\"}'\x1b[0m\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow query` command is used to Query a\nWorkflow Execution\nby ID.\n\n```\ntemporal workflow query \\\n\t\t--workflow-id MyWorkflowId \\\n\t\t--name MyQuery \\\n\t\t--input '{\"MyInputKey\": \"MyInputValue\"}'\n```\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.Type, "type", "", "Query Type/Name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "type")
	s.RejectCondition = NewStringEnum([]string{"not_open", "not_completed_cleanly"}, "")
	s.Command.Flags().Var(&s.RejectCondition, "reject-condition", "Optional flag for rejecting Queries based on Workflow state. Accepted values: not_open, not_completed_cleanly.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type TemporalWorkflowResetCommand struct {
	Parent      *TemporalWorkflowCommand
	Command     cobra.Command
	WorkflowId  string
	RunId       string
	EventId     int
	Reason      string
	ReapplyType StringEnum
	Type        StringEnum
	BuildId     string
	Query       string
	Yes         bool
}

func NewTemporalWorkflowResetCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowResetCommand {
	var s TemporalWorkflowResetCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "reset [flags]"
	s.Command.Short = "Resets a Workflow Execution by Event ID or reset type."
	if hasHighlighting {
		s.Command.Long = "The temporal workflow reset command resets a Workflow Execution.\nA reset allows the Workflow to resume from a certain point without losing its parameters or Event History.\n\nThe Workflow Execution can be set to a given Event Type:\n\x1b[1mtemporal workflow reset --workflow-id=meaningful-business-id --type=LastContinuedAsNew\x1b[0m\n\n...or a specific any Event after \x1b[1mWorkflowTaskStarted\x1b[0m.\n\x1b[1mtemporal workflow reset --workflow-id=meaningful-business-id --event-id=MyLastEvent\x1b[0m\nFor batch reset only FirstWorkflowTask, LastWorkflowTask or BuildId can be used. Workflow Id, run Id and event Id\nshould not be set.\nUse the options listed below to change reset behavior."
	} else {
		s.Command.Long = "The temporal workflow reset command resets a Workflow Execution.\nA reset allows the Workflow to resume from a certain point without losing its parameters or Event History.\n\nThe Workflow Execution can be set to a given Event Type:\n```\ntemporal workflow reset --workflow-id=meaningful-business-id --type=LastContinuedAsNew\n```\n\n...or a specific any Event after `WorkflowTaskStarted`.\n```\ntemporal workflow reset --workflow-id=meaningful-business-id --event-id=MyLastEvent\n```\nFor batch reset only FirstWorkflowTask, LastWorkflowTask or BuildId can be used. Workflow Id, run Id and event Id\nshould not be set.\nUse the options listed below to change reset behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow Id. Required for non-batch reset operations.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run Id.")
	s.Command.Flags().IntVarP(&s.EventId, "event-id", "e", 0, "The Event Id for any Event after `WorkflowTaskStarted` you want to reset to (exclusive). It can be `WorkflowTaskCompleted`, `WorkflowTaskFailed` or others.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "The reason why this workflow is being reset.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "reason")
	s.ReapplyType = NewStringEnum([]string{"All", "Signal", "None"}, "All")
	s.Command.Flags().Var(&s.ReapplyType, "reapply-type", "Event types to reapply after the reset point. Accepted values: All, Signal, None.")
	s.Type = NewStringEnum([]string{"FirstWorkflowTask", "LastWorkflowTask", "LastContinuedAsNew", "BuildId"}, "")
	s.Command.Flags().VarP(&s.Type, "type", "t", "Event type to which you want to reset. Accepted values: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew, BuildId.")
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Only used if type is BuildId. Reset the first workflow task processed by this build id. Note that by default, this reset is allowed to be to a prior run in a chain of continue-as-new.")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Start a batch reset to operate on Workflow Executions with given List Filter.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Confirm prompt to perform batch. Only allowed if query is present.")
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
	Follow bool
}

func NewTemporalWorkflowShowCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowShowCommand {
	var s TemporalWorkflowShowCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "show [flags]"
	s.Command.Short = "Show Event History for a Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow show\x1b[0m command provides the Event History for a\nWorkflow Execution. With JSON output specified, this output can be given to\nan SDK to perform a replay.\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow show` command provides the Event History for a\nWorkflow Execution. With JSON output specified, this output can be given to\nan SDK to perform a replay.\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().BoolVarP(&s.Follow, "follow", "f", false, "Follow the progress of a Workflow Execution in real time (does not apply to JSON output).")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type SingleWorkflowOrBatchOptions struct {
	WorkflowId string
	RunId      string
	Query      string
	Reason     string
	Yes        bool
}

func (v *SingleWorkflowOrBatchOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow Id. Either this or query must be set.")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run Id. Cannot be set when query is set.")
	f.StringVarP(&v.Query, "query", "q", "", "Start a batch to operate on Workflow Executions with given List Filter. Either this or Workflow Id must be set.")
	f.StringVar(&v.Reason, "reason", "", "Reason to perform batch. Only allowed if query is present unless the command specifies otherwise. Defaults to message with the current user's name.")
	f.BoolVarP(&v.Yes, "yes", "y", false, "Confirm prompt to perform batch. Only allowed if query is present.")
}

type TemporalWorkflowSignalCommand struct {
	Parent  *TemporalWorkflowCommand
	Command cobra.Command
	PayloadInputOptions
	Name string
	SingleWorkflowOrBatchOptions
}

func NewTemporalWorkflowSignalCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowSignalCommand {
	var s TemporalWorkflowSignalCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "signal [flags]"
	s.Command.Short = "Signal Workflow Execution by Id."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow signal\x1b[0m command is used to Signal a\nWorkflow Execution by ID.\n\n\x1b[1mtemporal workflow signal \\\n\t\t--workflow-id MyWorkflowId \\\n\t\t--name MySignal \\\n\t\t--input '{\"MyInputKey\": \"MyInputValue\"}'\x1b[0m\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow signal` command is used to Signal a\nWorkflow Execution by ID.\n\n```\ntemporal workflow signal \\\n\t\t--workflow-id MyWorkflowId \\\n\t\t--name MySignal \\\n\t\t--input '{\"MyInputKey\": \"MyInputValue\"}'\n```\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.Name, "name", "", "Signal Name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.SingleWorkflowOrBatchOptions.buildFlags(cctx, s.Command.Flags())
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
	s.Command.Short = "Query a Workflow Execution for its stack trace."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow stack\x1b[0m command Queries a\nWorkflow Execution with \x1b[1m__stack_trace\x1b[0m as the query type.\nThis returns a stack trace of all the threads or routines currently used by the workflow, and is\nuseful for troubleshooting.\n\n\x1b[1mtemporal workflow stack --workflow-id MyWorkflowId\x1b[0m\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow stack` command Queries a\nWorkflow Execution with `__stack_trace` as the query type.\nThis returns a stack trace of all the threads or routines currently used by the workflow, and is\nuseful for troubleshooting.\n\n```\ntemporal workflow stack --workflow-id MyWorkflowId\n```\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.RejectCondition = NewStringEnum([]string{"not_open", "not_completed_cleanly"}, "")
	s.Command.Flags().Var(&s.RejectCondition, "reject-condition", "Optional flag for rejecting Queries based on Workflow state. Accepted values: not_open, not_completed_cleanly.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}

type SharedWorkflowStartOptions struct {
	WorkflowId       string
	Type             string
	TaskQueue        string
	RunTimeout       time.Duration
	ExecutionTimeout time.Duration
	TaskTimeout      time.Duration
	SearchAttribute  []string
	Memo             []string
}

func (v *SharedWorkflowStartOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow Id.")
	f.StringVar(&v.Type, "type", "", "Workflow Type name.")
	_ = cobra.MarkFlagRequired(f, "type")
	f.StringVarP(&v.TaskQueue, "task-queue", "t", "", "Workflow Task queue.")
	_ = cobra.MarkFlagRequired(f, "task-queue")
	f.DurationVar(&v.RunTimeout, "run-timeout", 0, "Timeout of a Workflow Run.")
	f.DurationVar(&v.ExecutionTimeout, "execution-timeout", 0, "Timeout for a WorkflowExecution, including retries and ContinueAsNew tasks.")
	f.DurationVar(&v.TaskTimeout, "task-timeout", 10000*time.Millisecond, "Start-to-close timeout for a Workflow Task.")
	f.StringArrayVar(&v.SearchAttribute, "search-attribute", nil, "Passes Search Attribute in key=value format. Use valid JSON formats for value.")
	f.StringArrayVar(&v.Memo, "memo", nil, "Passes Memo in key=value format. Use valid JSON formats for value.")
}

type WorkflowStartOptions struct {
	Cron          string
	FailExisting  bool
	StartDelay    time.Duration
	IdReusePolicy string
}

func (v *WorkflowStartOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Cron, "cron", "", "Cron schedule for the workflow. Deprecated - use schedules instead.")
	f.BoolVar(&v.FailExisting, "fail-existing", false, "Fail if the workflow already exists.")
	f.DurationVar(&v.StartDelay, "start-delay", 0, "Specify a delay before the workflow starts. Cannot be used with a cron schedule. If the workflow receives a signal or update before the delay has elapsed, it will begin immediately.")
	f.StringVar(&v.IdReusePolicy, "id-reuse-policy", "", "Allows the same Workflow Id to be used in a new Workflow Execution. Accepted values: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.")
}

type PayloadInputOptions struct {
	Input       []string
	InputFile   []string
	InputMeta   []string
	InputBase64 bool
}

func (v *PayloadInputOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringArrayVarP(&v.Input, "input", "i", nil, "Input value (default JSON unless --input-payload-meta is non-JSON encoding). Can be given multiple times for multiple arguments. Cannot be combined with --input-file.")
	f.StringArrayVar(&v.InputFile, "input-file", nil, "Reads a file as the input (JSON by default unless --input-payload-meta is non-JSON encoding). Can be given multiple times for multiple arguments. Cannot be combined with --input.")
	f.StringArrayVar(&v.InputMeta, "input-meta", nil, "Metadata for the input payload. Expected as key=value. If key is encoding, overrides the default of json/plain.")
	f.BoolVar(&v.InputBase64, "input-base64", false, "If set, assumes --input or --input-file are base64 encoded and attempts to decode.")
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
	s.Command.Short = "Starts a new Workflow Execution."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow start\x1b[0m command starts a new Workflow Execution. The\nWorkflow and Run IDs are returned after starting the Workflow.\n\n\x1b[1mtemporal workflow start \\\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m"
	} else {
		s.Command.Long = "The `temporal workflow start` command starts a new Workflow Execution. The\nWorkflow and Run IDs are returned after starting the Workflow.\n\n```\ntemporal workflow start \\\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type MyWorkflow \\\n\t\t--task-queue MyTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.SharedWorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.WorkflowStartOptions.buildFlags(cctx, s.Command.Flags())
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
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
	RunId      string
	Query      string
	Reason     string
	Yes        bool
}

func NewTemporalWorkflowTerminateCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowTerminateCommand {
	var s TemporalWorkflowTerminateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "terminate [flags]"
	s.Command.Short = "Terminate Workflow Execution by ID or List Filter."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow terminate\x1b[0m command is used to terminate a\nWorkflow Execution. Canceling a running Workflow Execution records a\n\x1b[1mWorkflowExecutionTerminated\x1b[0m event as the closing Event in the workflow's Event History. Workflow code is oblivious to\ntermination. Use \x1b[1mtemporal workflow cancel\x1b[0m if you need to perform cleanup in your workflow.\n\nExecutions may be terminated by ID with an optional reason:\n\x1b[1mtemporal workflow terminate [--reason my-reason] --workflow-id MyWorkflowId\x1b[0m\n\n...or in bulk via a visibility query list filter:\n\x1b[1mtemporal workflow terminate --query=MyQuery\x1b[0m\n\nUse the options listed below to change the behavior of this command."
	} else {
		s.Command.Long = "The `temporal workflow terminate` command is used to terminate a\nWorkflow Execution. Canceling a running Workflow Execution records a\n`WorkflowExecutionTerminated` event as the closing Event in the workflow's Event History. Workflow code is oblivious to\ntermination. Use `temporal workflow cancel` if you need to perform cleanup in your workflow.\n\nExecutions may be terminated by ID with an optional reason:\n```\ntemporal workflow terminate [--reason my-reason] --workflow-id MyWorkflowId\n```\n\n...or in bulk via a visibility query list filter:\n```\ntemporal workflow terminate --query=MyQuery\n```\n\nUse the options listed below to change the behavior of this command."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow Id. Either this or query must be set.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run Id. Cannot be set when query is set.")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Start a batch to terminate Workflow Executions with given List Filter. Either this or Workflow Id must be set.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for termination. Defaults to message with the current user's name.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Confirm prompt to perform batch. Only allowed if query is present.")
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
	s.Command.Short = "Terminate Workflow Execution by ID or List Filter."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow trace\x1b[0m command display the progress of a Workflow Execution and its child workflows with a trace.\nThis view provides a great way to understand the flow of a workflow.\n\nUse the options listed below to change the behavior of this command."
	} else {
		s.Command.Long = "The `temporal workflow trace` command display the progress of a Workflow Execution and its child workflows with a trace.\nThis view provides a great way to understand the flow of a workflow.\n\nUse the options listed below to change the behavior of this command."
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringArrayVar(&s.Fold, "fold", nil, "Statuses for which Child Workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if no-fold supplied. Available values: running, completed, failed, canceled, terminated, timedout, continueasnew.")
	s.Command.Flags().BoolVar(&s.NoFold, "no-fold", false, "Disable folding. All Child Workflows within the set depth will be fetched and displayed.")
	s.Command.Flags().IntVar(&s.Depth, "depth", -1, "Depth of child workflows to fetch. Use -1 to fetch child workflows at any depth.")
	s.Command.Flags().IntVar(&s.Concurrency, "concurrency", 10, "Number of concurrent workflow histories that will be requested at any given time.")
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
	PayloadInputOptions
	Name                string
	WorkflowId          string
	UpdateId            string
	RunId               string
	FirstExecutionRunId string
}

func NewTemporalWorkflowUpdateCommand(cctx *CommandContext, parent *TemporalWorkflowCommand) *TemporalWorkflowUpdateCommand {
	var s TemporalWorkflowUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Updates a running workflow synchronously."
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow update\x1b[0m command is used to synchronously Update a\nWorkflowExecution by ID.\n\n\x1b[1mtemporal workflow update \\\n\t\t--workflow-id MyWorkflowId \\\n\t\t--name MyUpdate \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m\n\nUse the options listed below to change the command's behavior."
	} else {
		s.Command.Long = "The `temporal workflow update` command is used to synchronously Update a\nWorkflowExecution by ID.\n\n```\ntemporal workflow update \\\n\t\t--workflow-id MyWorkflowId \\\n\t\t--name MyUpdate \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```\n\nUse the options listed below to change the command's behavior."
	}
	s.Command.Args = cobra.NoArgs
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.Name, "name", "", "Update Name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow Id.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "workflow-id")
	s.Command.Flags().StringVar(&s.UpdateId, "update-id", "", "Update ID. If unset, default to a UUID.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run Id. If unset, the currently running Workflow Execution receives the Update.")
	s.Command.Flags().StringVar(&s.FirstExecutionRunId, "first-execution-run-id", "", "Send the Update to the last Workflow Execution in the chain that started with this Run Id.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}
