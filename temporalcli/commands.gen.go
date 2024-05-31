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
	s.Command.Short = "Temporal command-line interface and development server"
	if hasHighlighting {
		s.Command.Long = "The Temporal CLI manages, monitors, and debugs Temporal apps. It lets you\nrun a local Temporal Service, start Workflow Executions, pass messages to\nrunning Workflows, inspects state, and more.\n\n* Start a local development service:\n      \x1b[1mtemporal server start-dev\x1b[0m\n* View help: pass --help to any command:\n      \x1b[1mtemporal activity complete --help\x1b[0m"
	} else {
		s.Command.Long = "The Temporal CLI manages, monitors, and debugs Temporal apps. It lets you\nrun a local Temporal Service, start Workflow Executions, pass messages to\nrunning Workflows, inspects state, and more.\n\n* Start a local development service:\n      `temporal server start-dev`\n* View help: pass --help to any command:\n      `temporal activity complete --help`"
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
	s.Command.PersistentFlags().StringVar(&s.EnvFile, "env-file", "", "Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).")
	s.LogLevel = NewStringEnum([]string{"debug", "info", "warn", "error", "never"}, "info")
	s.Command.PersistentFlags().Var(&s.LogLevel, "log-level", "Log level. Default is \"info\" for most commands and \"warn\" for `server start-dev`. Accepted values: debug, info, warn, error, never.")
	s.Command.PersistentFlags().StringVar(&s.LogFormat, "log-format", "text", "Log format. Accepted values: text, json.")
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
	s.Command.Long = "Update Activity state to report completion or failure. This command marks\nan Activity as successfully finished or as having encountered an error."
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
		s.Command.Long = "Complete an Activity, marking it as successfully finished. Specify the ID\nand include a JSON result for the returned value:\n\n\x1b[1mtemporal activity complete \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --result '{\"YourResultKey\": \"YourResultVal\"}'\x1b[0m"
	} else {
		s.Command.Long = "Complete an Activity, marking it as successfully finished. Specify the ID\nand include a JSON result for the returned value:\n\n```\ntemporal activity complete \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId \\\n    --result '{\"YourResultKey\": \"YourResultVal\"}'\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "Activity `ID` to complete.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Result, "result", "", "Result `JSON` for completing the Activity.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "result")
	s.Command.Flags().StringVar(&s.Identity, "identity", "", "Identity of the user submitting this request.")
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
		s.Command.Long = "Fail an Activity, marking it as having encountered an error. Specify the\nActivity and Workflow IDs:\n\n\x1b[1mtemporal activity fail \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\x1b[0m"
	} else {
		s.Command.Long = "Fail an Activity, marking it as having encountered an error. Specify the\nActivity and Workflow IDs:\n\n```\ntemporal activity fail \\\n    --activity-id YourActivityId \\\n    --workflow-id YourWorkflowId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.ActivityId, "activity-id", "", "`ID` of the Activity to be failed.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "activity-id")
	s.Command.Flags().StringVar(&s.Detail, "detail", "", "JSON data with the reason for failing the Activity.")
	s.Command.Flags().StringVar(&s.Identity, "identity", "", "Identity of the user submitting this request.")
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
	s.Command.Short = "Manage batch jobs"
	if hasHighlighting {
		s.Command.Long = "A batch job executes a command affecting multiple Workflow Executions at once.\nSpecify the Workflow Executions to include and the type of batch job to apply.\nFor example, to cancel all running 'YourWorkflow' Workflows:\n\n\x1b[1mtemporal batch workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\x1b[0m\n\nThe \x1b[1mbatch\x1b[0m command keyword is optional when you supply a \x1b[1m--query\x1b[0m. This\ninvocation is identical to the previous one:\n\n\x1b[1mtemporal workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\x1b[0m"
	} else {
		s.Command.Long = "A batch job executes a command affecting multiple Workflow Executions at once.\nSpecify the Workflow Executions to include and the type of batch job to apply.\nFor example, to cancel all running 'YourWorkflow' Workflows:\n\n```\ntemporal batch workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\n```\n\nThe `batch` command keyword is optional when you supply a `--query`. This\ninvocation is identical to the previous one:\n\n```\ntemporal workflow cancel \\\n  --query 'ExecutionStatus = \"Running\" AND WorkflowType=\"YourWorkflow\"' \\\n  --reason \"Testing\"\n```"
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
		s.Command.Long = "Show the progress of an ongoing batch job. Pass a valid job ID to display\nits information:\n\n\x1b[1mtemporal batch describe --job-id YourJobId\x1b[0m"
	} else {
		s.Command.Long = "Show the progress of an ongoing batch job. Pass a valid job ID to display\nits information:\n\n```\ntemporal batch describe --job-id YourJobId\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.JobId, "job-id", "", "Batch job ID.")
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
		s.Command.Long = "Return a list of batch jobs on the Service or within a single Namespace.\nFor example, list the batch jobs for \"YourNamespace\":\n\x1b[1mtemporal batch list --namespace YourNamespace\x1b[0m"
	} else {
		s.Command.Long = "Return a list of batch jobs on the Service or within a single Namespace.\nFor example, list the batch jobs for \"YourNamespace\":\n```\ntemporal batch list --namespace YourNamespace\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Max output count.")
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
	s.Command.Short = "Terminate a batch job"
	if hasHighlighting {
		s.Command.Long = "Terminate a batch job with the provided job ID. Provide a reason for the\ntermination. This explanation is stored with the Service for later reference.\nFor example:\n\n\x1b[1mtemporal batch terminate \\\n    --job-id YourJobId \\\n    --reason YourTerminationReason\x1b[0m"
	} else {
		s.Command.Long = "Terminate a batch job with the provided job ID. Provide a reason for the\ntermination. This explanation is stored with the Service for later reference.\nFor example:\n\n```\ntemporal batch terminate \\\n    --job-id YourJobId \\\n    --reason YourTerminationReason\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.JobId, "job-id", "", "Job Id to terminate.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "job-id")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for terminating the batch job.")
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
		s.Command.Long = "Environments manage key-value presets, auto-configuring Temporal CLI options\nfor you. Set up distinct environments like \"dev\" and \"prod\" for convenience.\nFor example:\n\n\x1b[1mtemporal env set \\\n    --env prod \\\n    --key address \\\n    --value production.f45a2.tmprl.cloud:7233\x1b[0m\n\nEach environment is isolated. Changes to \"prod\" presets won't affect \"dev\".\n\nFor easiest use, set a \x1b[1mTEMPORAL_ENV\x1b[0m environmental variable in your shell.\nThe Temporal CLI checks for an \x1b[1m--env\x1b[0m option first, then checks the shell.\nIf neither is set, most commands use \x1b[1mdefault\x1b[0m environmental presets if\nthey are available."
	} else {
		s.Command.Long = "Environments manage key-value presets, auto-configuring Temporal CLI options\nfor you. Set up distinct environments like \"dev\" and \"prod\" for convenience.\nFor example:\n\n```\ntemporal env set \\\n    --env prod \\\n    --key address \\\n    --value production.f45a2.tmprl.cloud:7233\n```\n\nEach environment is isolated. Changes to \"prod\" presets won't affect \"dev\".\n\nFor easiest use, set a `TEMPORAL_ENV` environmental variable in your shell.\nThe Temporal CLI checks for an `--env` option first, then checks the shell.\nIf neither is set, most commands use `default` environmental presets if\nthey are available."
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
		s.Command.Long = "Remove an environment entirely or remove a key-value pair within an\nenvironment. If you don't specify an environment (with --env or by setting\nthe TEMPORAL_ENV variable), this updates the default environment. \nFor example:\n\n\x1b[1mtemporal env delete --env YourEnvironment\x1b[0m\n\nor\n\n\x1b[1mtemporal env delete --env prod  --key tls-key-path\x1b[0m"
	} else {
		s.Command.Long = "Remove an environment entirely or remove a key-value pair within an\nenvironment. If you don't specify an environment (with --env or by setting\nthe TEMPORAL_ENV variable), this updates the default environment. \nFor example:\n\n```\ntemporal env delete --env YourEnvironment\n```\n\nor\n\n```\ntemporal env delete --env prod  --key tls-key-path\n```"
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
		s.Command.Long = "List the properties for a given environment:\n\n\x1b[1mtemporal env get --env YourEnvironment\x1b[0m\n\nPrint a single property:\n\n\x1b[1mtemporal env get --env YourEnvironment --key YourPropertyKey\x1b[0m\n\nIf you don't specify an environment name, with \x1b[1m--env\x1b[0m or by setting\n\x1b[1mTEMPORAL_ENV\x1b[0m as an environmental variable in your shell, this command\nlists properties in the \x1b[1mdefault\x1b[0m environment."
	} else {
		s.Command.Long = "List the properties for a given environment:\n\n```\ntemporal env get --env YourEnvironment\n```\n\nPrint a single property:\n\n```\ntemporal env get --env YourEnvironment --key YourPropertyKey\n```\n\nIf you don't specify an environment name, with `--env` or by setting\n`TEMPORAL_ENV` as an environmental variable in your shell, this command\nlists properties in the `default` environment."
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
	s.Command.Short = "List environments"
	s.Command.Long = "List the environments you have set up on your local computer. The output \nenumerates the environments stored in the Temporal environment file on\nyour computer (\"$HOME/.config/temporalio/temporal.yaml\")."
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
	s.Command.Short = "Set environment properties"
	if hasHighlighting {
		s.Command.Long = "Assign a value to a property key and store it to an environment:\n\n\x1b[1mtemporal env set \\\n    --env environment \\\n    --key property \\\n    --value value\x1b[0m\n\nIf you don't specify an environment name, with \x1b[1m--env\x1b[0m or by setting\n\x1b[1mTEMPORAL_ENV\x1b[0m as an environmental variable in your shell, this command\nsets properties in the \x1b[1mdefault\x1b[0m environment.\n\nSetting property names lets the CLI automatically set options on your behalf.\nStoring these property values in advance reduces the effort required to \nissue CLI commands and helps avoid typos."
	} else {
		s.Command.Long = "Assign a value to a property key and store it to an environment:\n\n```\ntemporal env set \\\n    --env environment \\\n    --key property \\\n    --value value\n```\n\nIf you don't specify an environment name, with `--env` or by setting\n`TEMPORAL_ENV` as an environmental variable in your shell, this command\nsets properties in the `default` environment.\n\nSetting property names lets the CLI automatically set options on your behalf.\nStoring these property values in advance reduces the effort required to \nissue CLI commands and helps avoid typos."
	}
	s.Command.Args = cobra.MaximumNArgs(2)
	s.Command.Flags().StringVarP(&s.Key, "key", "k", "", "Property name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "key")
	s.Command.Flags().StringVarP(&s.Value, "value", "v", "", "Property value.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "value")
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
		s.Command.Long = "Operator commands manage and fetch information about Namespaces, \nSearch Attributes, and Temporal Clusters:\n\n\x1b[1mtemporal operator [command] [subcommand] [options]\x1b[0m\n\nFor example, to view a cluster-by-cluster overview:\n\n\x1b[1mtemporal operator cluster describe\x1b[0m"
	} else {
		s.Command.Long = "Operator commands manage and fetch information about Namespaces, \nSearch Attributes, and Temporal Clusters:\n\n```\ntemporal operator [command] [subcommand] [options]\n```\n\nFor example, to view a cluster-by-cluster overview:\n\n```\ntemporal operator cluster describe\n```"
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
	s.Command.Short = "Manage a Temporal Cluster"
	if hasHighlighting {
		s.Command.Long = "Perform operator actions on Temporal Clusters.\n\n\x1b[1mtemporal operator cluster [subcommand] [options]\x1b[0m\n\nFor example to check Cluster health:\n\n\x1b[1mtemporal operator cluster health\x1b[0m"
	} else {
		s.Command.Long = "Perform operator actions on Temporal Clusters.\n\n```\ntemporal operator cluster [subcommand] [options]\n```\n\nFor example to check Cluster health:\n\n```\ntemporal operator cluster health\n```"
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
	s.Command.Short = "Describe a Temporal Cluster"
	if hasHighlighting {
		s.Command.Long = "View information about a Temporal Cluster, including Cluster Name,\npersistence store, and visibility store. Add \x1b[1m--detail\x1b[0m for additional\ninformation:\n\n\x1b[1mtemporal operator cluster describe [--detail]\x1b[0m"
	} else {
		s.Command.Long = "View information about a Temporal Cluster, including Cluster Name,\npersistence store, and visibility store. Add `--detail` for additional\ninformation:\n\n```\ntemporal operator cluster describe [--detail]\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVar(&s.Detail, "detail", false, "High-detail output.")
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
	s.Command.Short = "Check Frontend Service health"
	if hasHighlighting {
		s.Command.Long = "View information about the health of a Frontend Service:\n\n\x1b[1mtemporal operator cluster health\x1b[0m"
	} else {
		s.Command.Long = "View information about the health of a Frontend Service:\n\n```\ntemporal operator cluster health\n```"
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
		s.Command.Long = "Print a list of Temporal Clusters registered to this system. Report details\ninclude the Cluster's name, ID, address, History Shard count,\nFailover version, and availability.\n\n\x1b[1mtemporal operator cluster list [--limit max-count]\x1b[0m"
	} else {
		s.Command.Long = "Print a list of Temporal Clusters registered to this system. Report details\ninclude the Cluster's name, ID, address, History Shard count,\nFailover version, and availability.\n\n```\ntemporal operator cluster list [--limit max-count]\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Max count to list.")
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
		s.Command.Long = "Remove a registered Temporal Cluster from this system.\n\n\x1b[1mtemporal operator cluster remove --name YourClusterName\x1b[0m"
	} else {
		s.Command.Long = "Remove a registered Temporal Cluster from this system.\n\n```\ntemporal operator cluster remove --name YourClusterName\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.Name, "name", "", "Cluster name.")
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
		s.Command.Long = "Show Temporal Server information for Temporal Clusters: Server version, \nscheduling support, and more. This information helps diagnose problems with\nthe Temporal Server. This command defaults to the local Service. Otherwise, \nuse the \x1b[1m--frontend-address\x1b[0m option to specify a Cluster endpoint.\n\n\x1b[1mtemporal operator cluster system \\ \n    --frontend-address \"your_remote_endpoint:your_remote_port\"\x1b[0m"
	} else {
		s.Command.Long = "Show Temporal Server information for Temporal Clusters: Server version, \nscheduling support, and more. This information helps diagnose problems with\nthe Temporal Server. This command defaults to the local Service. Otherwise, \nuse the `--frontend-address` option to specify a Cluster endpoint.\n\n```\ntemporal operator cluster system \\ \n    --frontend-address \"your_remote_endpoint:your_remote_port\"\n```"
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
		s.Command.Long = "Add, remove, or update a registered (\"remote\") Temporal Cluster.\n\n\x1b[1mtemporal operator cluster upsert [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator cluster upsert \\\n    --frontend_address \"your_remote_endpoint:your_remote_port\"\n    --enable_connection false\x1b[0m"
	} else {
		s.Command.Long = "Add, remove, or update a registered (\"remote\") Temporal Cluster.\n\n```\ntemporal operator cluster upsert [options]\n```\n\nFor example:\n\n```\ntemporal operator cluster upsert \\\n    --frontend_address \"your_remote_endpoint:your_remote_port\"\n    --enable_connection false\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.FrontendAddress, "frontend-address", "", "Remote endpoint.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "frontend-address")
	s.Command.Flags().BoolVar(&s.EnableConnection, "enable-connection", false, "Enabled/disabled connection.")
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
		s.Command.Long = "Perform operations on Temporal Cluster Namespaces:\n\n\x1b[1mtemporal operator namespace [command] [command options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal operator namespace create \\\n    --namespace YourNewNamespaceName\x1b[0m"
	} else {
		s.Command.Long = "Perform operations on Temporal Cluster Namespaces:\n\n```\ntemporal operator namespace [command] [command options]\n```\n\nFor example:\n\n```\ntemporal operator namespace create \\\n    --namespace YourNewNamespaceName\n```"
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
	s.Command.Short = "Register a new Namespace"
	if hasHighlighting {
		s.Command.Long = "Create a new Namespace on the Temporal Server:\n\n\x1b[1mtemporal operator namespace create \\\n    --namespace YourNewNamespaceName \\\n    [options]\x1b[0m`\n\nTo create a global Namespace, for cross-region data replication:\n\n\x1b[1mtemporal operator namespace create \\\n    --global \\\n    --namespace YourNewNamespaceName\x1b[0m\n\nConfigure other settings like retention and Visibility Archival State\nas needed. For example, the Visibility Archive can be set on a separate URI:\n\n\x1b[1mtemporal operator namespace create \\\n    --retention 5d \\\n    --visibility-archival-state enabled \\\n    --visibility-uri YourURI \\\n    --namespace YourNewNamespaceName\x1b[0m\n\nNote: URI values for archival states can't be changed after being enabled."
	} else {
		s.Command.Long = "Create a new Namespace on the Temporal Server:\n\n```\ntemporal operator namespace create \\\n    --namespace YourNewNamespaceName \\\n    [options]\n````\n\nTo create a global Namespace, for cross-region data replication:\n\n```\ntemporal operator namespace create \\\n    --global \\\n    --namespace YourNewNamespaceName\n```\n\nConfigure other settings like retention and Visibility Archival State\nas needed. For example, the Visibility Archive can be set on a separate URI:\n\n```\ntemporal operator namespace create \\\n    --retention 5d \\\n    --visibility-archival-state enabled \\\n    --visibility-uri YourURI \\\n    --namespace YourNewNamespaceName\n```\n\nNote: URI values for archival states can't be changed after being enabled."
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active cluster name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster names for Namespace creation. Can be passed multiple times.")
	s.Command.Flags().StringVar(&s.Data, "data", "", "Namespace data.")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description. Comma-separated \"KEY=VALUE\" string pairs.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.Global, "global", false, "Enable/disable cross-region data replication.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disabled")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Archive history to this `URI`. Once enabled, cannot be changed.")
	s.Command.Flags().DurationVar(&s.Retention, "retention", 259200000*time.Millisecond, "Time to preserve closed Workflows before deletion.")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "disable")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Archive visibility data to this `URI`. Once enabled, cannot be changed.")
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
	s.Command.Short = "Delete Namespace"
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
	s.Command.Use = "describe [flags] [namespace]"
	s.Command.Short = "Describe Namespace"
	if hasHighlighting {
		s.Command.Long = "Provide long-form information about a Namespace identified by its ID or name:\n\n\x1b[1mtemporal operator namespace describe --namespace-id YourNamespaceId\x1b[0m\n\nor\n\n\x1b[1mtemporal operator namespace describe --namespace YourNamespaceName\x1b[0m"
	} else {
		s.Command.Long = "Provide long-form information about a Namespace identified by its ID or name:\n\n```\ntemporal operator namespace describe --namespace-id YourNamespaceId\n```\n\nor\n\n```\ntemporal operator namespace describe --namespace YourNamespaceName\n```"
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
		s.Command.Long = "A long-form listing for all Namespaces on the Service:\n\n\x1b[1mtemporal operator namespace list\x1b[0m"
	} else {
		s.Command.Long = "A long-form listing for all Namespaces on the Service:\n\n```\ntemporal operator namespace list\n```"
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
	Retention               time.Duration
	VisibilityArchivalState StringEnum
	VisibilityUri           string
}

func NewTemporalOperatorNamespaceUpdateCommand(cctx *CommandContext, parent *TemporalOperatorNamespaceCommand) *TemporalOperatorNamespaceUpdateCommand {
	var s TemporalOperatorNamespaceUpdateCommand
	s.Parent = parent
	s.Command.DisableFlagsInUseLine = true
	s.Command.Use = "update [flags]"
	s.Command.Short = "Update Namespace"
	if hasHighlighting {
		s.Command.Long = "Update a Namespace using the properties you specify.\n\n\x1b[1mtemporal operator namespace update [options]\x1b[0m \n\nAssign a Namespace's active Cluster:\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --active-cluster NewActiveCluster\x1b[0m\n\nPromote a Namespace for data replication:\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --promote-global\x1b[0m\n\nYou may update archives that were previously enabled or disabled. \nNote: URI values for archival states can't be changed after being enabled:\n\n\x1b[1mtemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --history-archival-state enabled \\\n    --visibility-archival-state disabled\x1b[0m"
	} else {
		s.Command.Long = "Update a Namespace using the properties you specify.\n\n```\ntemporal operator namespace update [options]\n``` \n\nAssign a Namespace's active Cluster:\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --active-cluster NewActiveCluster\n```\n\nPromote a Namespace for data replication:\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --promote-global\n```\n\nYou may update archives that were previously enabled or disabled. \nNote: URI values for archival states can't be changed after being enabled:\n\n```\ntemporal operator namespace update \\\n    --namespace YourNamespaceName \\\n    --history-archival-state enabled \\\n    --visibility-archival-state disabled\n```"
	}
	s.Command.Args = cobra.MaximumNArgs(1)
	s.Command.Flags().StringVar(&s.ActiveCluster, "active-cluster", "", "Active cluster name.")
	s.Command.Flags().StringArrayVar(&s.Cluster, "cluster", nil, "Cluster names.")
	s.Command.Flags().StringArrayVar(&s.Data, "data", nil, "Set a \"KEY=VALUE\" string pair in Namespace data. KEY is an unquoted string identifier. Unquoted JSON is recommended for the VALUE. Can be passed multiple times.")
	s.Command.Flags().StringVar(&s.Description, "description", "", "Namespace description.")
	s.Command.Flags().StringVar(&s.Email, "email", "", "Owner email.")
	s.Command.Flags().BoolVar(&s.PromoteGlobal, "promote-global", false, "Enable/disable cross-region data replication.")
	s.HistoryArchivalState = NewStringEnum([]string{"disabled", "enabled"}, "")
	s.Command.Flags().Var(&s.HistoryArchivalState, "history-archival-state", "History archival state. Accepted values: disabled, enabled.")
	s.Command.Flags().StringVar(&s.HistoryUri, "history-uri", "", "Archive history to this `URI`. Once enabled, cannot be changed.")
	s.Command.Flags().DurationVar(&s.Retention, "retention", 0, "Length of time a closed Workflow is preserved before deletion")
	s.VisibilityArchivalState = NewStringEnum([]string{"disabled", "enable"}, "")
	s.Command.Flags().Var(&s.VisibilityArchivalState, "visibility-archival-state", "Visibility archival state. Accepted values: disabled, enable.")
	s.Command.Flags().StringVar(&s.VisibilityUri, "visibility-uri", "", "Archive visibility data to this `URI`. Once enabled, cannot be changed.")
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
		s.Command.Long = "Create, list, or remove Search Attributes, fields stored in a Workflow\nExecution's metadata:\n\n\x1b[1mtemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\x1b[0m\n\nSupported types include: Text, Keyword, Int, Double, Bool, Datetime, and\nKeywordList."
	} else {
		s.Command.Long = "Create, list, or remove Search Attributes, fields stored in a Workflow\nExecution's metadata:\n\n```\ntemporal operator search-attribute create \\\n    --name YourAttributeName \\\n    --type Keyword\n```\n\nSupported types include: Text, Keyword, Int, Double, Bool, Datetime, and\nKeywordList."
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
	Type    []string
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
	s.Command.Flags().StringArrayVar(&s.Name, "name", nil, "Search Attribute name. Required")
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
	s.Command.Short = "List Search Attributes"
	if hasHighlighting {
		s.Command.Long = "Display a list of active Search Attributes that can be assigned or used with \nWorkflow Queries.\n\n\x1b[1mtemporal operator search-attribute list\x1b[0m"
	} else {
		s.Command.Long = "Display a list of active Search Attributes that can be assigned or used with \nWorkflow Queries.\n\n```\ntemporal operator search-attribute list\n```"
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
		s.Command.Long = "Remove custom Search Attributes from the options that can be assigned or\nused with Workflow Queries.\n\n\x1b[1mtemporal operator search-attribute remove \\\n    --name YourAttributeName\x1b[0m\n\nRemove without confirmation:\n\n\x1b[1mtemporal operator search-attribute remove \\\n    --name YourAttributeName \\\n    --yes false\x1b[0m"
	} else {
		s.Command.Long = "Remove custom Search Attributes from the options that can be assigned or\nused with Workflow Queries.\n\n```\ntemporal operator search-attribute remove \\\n    --name YourAttributeName\n```\n\nRemove without confirmation:\n\n```\ntemporal operator search-attribute remove \\\n    --name YourAttributeName \\\n    --yes false\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.Name, "name", nil, "Search Attribute name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Override confirmation request before removal.")
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
		s.Command.Long = "Create, use, and update Schedules that allow Workflow Executions to\nbe created at specified times:\n\n\x1b[1mtemporal schedule [commands] [options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal schedule describe \\\n    --schedule-id \"YourScheduleId\"\x1b[0m"
	} else {
		s.Command.Long = "Create, use, and update Schedules that allow Workflow Executions to\nbe created at specified times:\n\n```\ntemporal schedule [commands] [options]\n```\n\nFor example:\n\n```\ntemporal schedule describe \\\n    --schedule-id \"YourScheduleId\"\n```"
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
	f.StringVarP(&v.ScheduleId, "schedule-id", "s", "", "Schedule ID.")
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
	s.Command.Short = "Backfill past actions"
	if hasHighlighting {
		s.Command.Long = "Batch-execute actions that would have run during a specified time interval.\nUse this command to fill in Workflow Runs from when a Schedule was paused,\nfrom before a Schedule was created, from the future, or to re-process an\ninterval that was previously executed.\n\nBackfills require a Schedule ID and the time period covered by the request. \nIt's best to use the \x1b[1mBufferAll\x1b[0m or \x1b[1mAllowAll\x1b[0m policies.\n\nFor example:\n\n\x1b[1m  temporal schedule backfill \\\n    --schedule-id \"YourScheduleId\" \\\n    --start-time \"2022-05-01T00:00:00Z\" \\\n    --end-time \"2022-05-31T23:59:59Z\" \\\n    --overlap-policy BufferAll\x1b[0m"
	} else {
		s.Command.Long = "Batch-execute actions that would have run during a specified time interval.\nUse this command to fill in Workflow Runs from when a Schedule was paused,\nfrom before a Schedule was created, from the future, or to re-process an\ninterval that was previously executed.\n\nBackfills require a Schedule ID and the time period covered by the request. \nIt's best to use the `BufferAll` or `AllowAll` policies.\n\nFor example:\n\n```\n  temporal schedule backfill \\\n    --schedule-id \"YourScheduleId\" \\\n    --start-time \"2022-05-01T00:00:00Z\" \\\n    --end-time \"2022-05-31T23:59:59Z\" \\\n    --overlap-policy BufferAll\n```"
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
	f.StringArrayVar(&v.Calendar, "calendar", nil, "Calendar specification in JSON. For example: `{\"dayOfWeek\":\"Fri\",\"hour\":\"17\",\"minute\":\"5\"}`.")
	f.DurationVar(&v.CatchupWindow, "catchup-window", 0, "Maximum catch-up time for when the Service is unavailable.")
	f.StringArrayVar(&v.Cron, "cron", nil, "Calendar specification in cron string format. For example: `\"3 11 * * Fri\"`.")
	f.Var(&v.EndTime, "end-time", "Schedule end time")
	f.StringArrayVar(&v.Interval, "interval", nil, "Interval duration. For example, 90m, or 90m/13m to include phase offset.")
	f.DurationVar(&v.Jitter, "jitter", 0, "Per-action jitter range.")
	f.StringVar(&v.Notes, "notes", "", "Initial notes field value.")
	f.BoolVar(&v.Paused, "paused", false, "Initial value for paused state.")
	f.BoolVar(&v.PauseOnFailure, "pause-on-failure", false, "Pause schedule after Workflow failures.")
	f.IntVar(&v.RemainingActions, "remaining-actions", 0, "Total allowed actions. Default is zero (unlimited).")
	f.Var(&v.StartTime, "start-time", "Schedule start time.")
	f.StringVar(&v.TimeZone, "time-zone", "", "Interpret calendar specs with the `TZ` time zone. For a list of time zones, see: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones.")
	f.StringArrayVar(&v.ScheduleSearchAttribute, "schedule-search-attribute", nil, "Set schedule Search Attributes using \"KEY=VALUE\" format. KEY is an unquoted string identifier. VALUE is a JSON string. Can be passed multiple times.")
	f.StringArrayVar(&v.ScheduleMemo, "schedule-memo", nil, "Set a schedule memo using \"KEY=VALUE\" format. KEY is an unquoted string identifier. VALUE is a JSON string. Can be passed multiple times.")
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
		s.Command.Long = "Create a new Schedule on the Frontend Service. A Schedule will\nautomatically start new Workflow Executions at the times you specify.\n\nFor example:\n\n\x1b[1m  temporal schedule create \\\n    --schedule-id \"YourScheduleId\" \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"30\"}' \\\n    --workflow-id \"YourBaseWorkflowId\" \\\n    --task-queue \"YourTaskQueue\" \\\n    --workflow-type \"YourWorkflowType\"\x1b[0m\n\nSchedules support any combination of \x1b[1m--calendar\x1b[0m, \x1b[1m--interval\x1b[0m, \nand \x1b[1m--cron\x1b[0m."
	} else {
		s.Command.Long = "Create a new Schedule on the Frontend Service. A Schedule will\nautomatically start new Workflow Executions at the times you specify.\n\nFor example:\n\n```\n  temporal schedule create \\\n    --schedule-id \"YourScheduleId\" \\\n    --calendar '{\"dayOfWeek\":\"Fri\",\"hour\":\"3\",\"minute\":\"30\"}' \\\n    --workflow-id \"YourBaseWorkflowId\" \\\n    --task-queue \"YourTaskQueue\" \\\n    --workflow-type \"YourWorkflowType\"\n```\n\nSchedules support any combination of `--calendar`, `--interval`, \nand `--cron`."
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
	s.Command.Short = "Delete a Schedule"
	if hasHighlighting {
		s.Command.Long = "Deletes a Schedule on the front end Service:\n\n\x1b[1mtemporal schedule delete --schedule-id YourScheduleId\x1b[0m \n\nRemoving a schedule won't affect the Workflow Executions it started that\nare still running. To cancel or terminate these Workflow Executions, use \n\x1b[1mtemporal workflow delete\x1b[0m with the \x1b[1mTemporalScheduledById\x1b[0m Search Attribute."
	} else {
		s.Command.Long = "Deletes a Schedule on the front end Service:\n\n```\ntemporal schedule delete --schedule-id YourScheduleId\n``` \n\nRemoving a schedule won't affect the Workflow Executions it started that\nare still running. To cancel or terminate these Workflow Executions, use \n`temporal workflow delete` with the `TemporalScheduledById` Search Attribute."
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
	s.Command.Short = "Get Schedule configuration and current state"
	if hasHighlighting {
		s.Command.Long = "Show a Schedule configuration including information about past, current, \nand future Workflow runs:\n\n\x1b[1mtemporal schedule describe --schedule-id YourScheduleId\x1b[0m"
	} else {
		s.Command.Long = "Show a Schedule configuration including information about past, current, \nand future Workflow runs:\n\n```\ntemporal schedule describe --schedule-id YourScheduleId\n```"
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
	s.Command.Short = "List Schedules"
	if hasHighlighting {
		s.Command.Long = "Lists the Schedules hosted by a Namespace:\n\n\x1b[1mtemporal schedule list --namespace YourNamespace\x1b[0m"
	} else {
		s.Command.Long = "Lists the Schedules hosted by a Namespace:\n\n```\ntemporal schedule list --namespace YourNamespace\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().BoolVarP(&s.Long, "long", "l", false, "Show detailed information")
	s.Command.Flags().BoolVar(&s.ReallyLong, "really-long", false, "Show extensive information in non-table form.")
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
		s.Command.Long = "Pause or unpause a Schedule:\n\n\x1b[1mtemporal schedule toggle \\\n    --schedule-id \"YourScheduleId\" \\\n    --pause \\\n    --reason \"YourReason\"\x1b[0m\n\nand\n\n\x1b[1mtemporal schedule toggle\n    --schedule-id \"YourScheduleId\" \\\n    --unpause \\\n    --reason \"YourReason\"\x1b[0m\n\nThe reason updates the Schedule's \x1b[1mnotes\x1b[0m field to support operations\ncommunication. It defaults to \"(no reason provided)\" when omitted."
	} else {
		s.Command.Long = "Pause or unpause a Schedule:\n\n```\ntemporal schedule toggle \\\n    --schedule-id \"YourScheduleId\" \\\n    --pause \\\n    --reason \"YourReason\"\n```\n\nand\n\n```\ntemporal schedule toggle\n    --schedule-id \"YourScheduleId\" \\\n    --unpause \\\n    --reason \"YourReason\"\n```\n\nThe reason updates the Schedule's `notes` field to support operations\ncommunication. It defaults to \"(no reason provided)\" when omitted."
	}
	s.Command.Args = cobra.NoArgs
	s.ScheduleIdOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().BoolVar(&s.Pause, "pause", false, "Pause the schedule.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "\"(no reason provided)", "Reason for pausing/unpausing.")
	s.Command.Flags().BoolVar(&s.Unpause, "unpause", false, "Unpause the schedule.")
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
		s.Command.Long = "Trigger a Schedule to run immediately:\n\n\x1b[1mtemporal schedule trigger --schedule-id \"YourScheduleId\"\x1b[0m"
	} else {
		s.Command.Long = "Trigger a Schedule to run immediately:\n\n```\ntemporal schedule trigger --schedule-id \"YourScheduleId\"\n```"
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
		s.Command.Long = "Update an existing Schedule with new configuration details, including\nspec, action, and policies:\n\n\x1b[1mtemporal schedule update \ntemporal schedule update \\  \n --schedule-id \"YourScheduleId\"   \n --workflow-type \"NewWorkflowType\"\x1b[0m"
	} else {
		s.Command.Long = "Update an existing Schedule with new configuration details, including\nspec, action, and policies:\n\n```\ntemporal schedule update \ntemporal schedule update \\  \n --schedule-id \"YourScheduleId\"   \n --workflow-type \"NewWorkflowType\"\n```"
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
		s.Command.Long = "Run a development Temporal Server on your local system.\nView the Web UI for the default configuration at http://localhost:8233:\n\n\x1b[1mtemporal server start-dev\x1b[0m\n\nAdd persistence for Workflow Executions across runs:\n\n\x1b[1mtemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\x1b[0m\n\nSet the port from the front-end gRPC Service (7233 default):\n\n\x1b[1mtemporal server start-dev \\\n    --port 7000\x1b[0m\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n\x1b[1mtemporal server start-dev \\\n    --ui-port 3000\x1b[0m"
	} else {
		s.Command.Long = "Run a development Temporal Server on your local system.\nView the Web UI for the default configuration at http://localhost:8233:\n\n```\ntemporal server start-dev\n```\n\nAdd persistence for Workflow Executions across runs:\n\n```\ntemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\n```\n\nSet the port from the front-end gRPC Service (7233 default):\n\n```\ntemporal server start-dev \\\n    --port 7000\n```\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n```\ntemporal server start-dev \\\n    --ui-port 3000\n```"
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
	s.Command.Short = "Start Temporal development server"
	if hasHighlighting {
		s.Command.Long = "Run a development Temporal Server on your local system.\nView the Web UI for the default configuration at http://localhost:8233:\n\n\x1b[1mtemporal server start-dev\x1b[0m\n\nAdd persistence for Workflow Executions across runs:\n\n\x1b[1mtemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\x1b[0m\n\nSet the port from the front-end gRPC Service (7233 default):\n\n\x1b[1mtemporal server start-dev \\\n    --port 7000\x1b[0m\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n\x1b[1mtemporal server start-dev \\\n    --ui-port 3000\x1b[0m"
	} else {
		s.Command.Long = "Run a development Temporal Server on your local system.\nView the Web UI for the default configuration at http://localhost:8233:\n\n```\ntemporal server start-dev\n```\n\nAdd persistence for Workflow Executions across runs:\n\n```\ntemporal server start-dev \\\n    --db-filename path-to-your-local-persistent-store\n```\n\nSet the port from the front-end gRPC Service (7233 default):\n\n```\ntemporal server start-dev \\\n    --port 7000\n```\n\nUse a custom port for the Web UI. The default is the gRPC port (7233 default)\nplus 1000 (8233):\n\n```\ntemporal server start-dev \\\n    --ui-port 3000\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.DbFilename, "db-filename", "f", "", "Path to file for persistent Temporal state store. By default, Workflow Executions are lost when the server process dies.")
	s.Command.Flags().StringArrayVarP(&s.Namespace, "namespace", "n", nil, "Namespaces to be created at launch. The \"default\" Namespace is always created automatically.")
	s.Command.Flags().IntVarP(&s.Port, "port", "p", 7233, "Port for the frontend gRPC Service.")
	s.Command.Flags().IntVar(&s.HttpPort, "http-port", 0, "Port for the frontend HTTP API service. Default is off.")
	s.Command.Flags().IntVar(&s.MetricsPort, "metrics-port", 0, "Port for '/metrics'. Default is off.")
	s.Command.Flags().IntVar(&s.UiPort, "ui-port", 0, "Port for the Web UI. Default is '--port' value + 1000.")
	s.Command.Flags().BoolVar(&s.Headless, "headless", false, "Disable the Web UI.")
	s.Command.Flags().StringVar(&s.Ip, "ip", "localhost", "IP address bound to the frontend Service.")
	s.Command.Flags().StringVar(&s.UiIp, "ui-ip", "", "IP address bound to the WebUI. Default is same as '--ip' value.")
	s.Command.Flags().StringVar(&s.UiAssetPath, "ui-asset-path", "", "UI custom assets path.")
	s.Command.Flags().StringVar(&s.UiCodecEndpoint, "ui-codec-endpoint", "", "UI remote codec HTTP endpoint.")
	s.Command.Flags().StringArrayVar(&s.SqlitePragma, "sqlite-pragma", nil, "SQLite pragma statements in \"PRAGMA=VALUE\" format.")
	s.Command.Flags().StringArrayVar(&s.DynamicConfigValue, "dynamic-config-value", nil, "Dynamic configuration value in \"KEY=VALUE\" format. KEY is an unquoted string identifier. VALUE is a JSON string. For example, \"YourKey=\\\"YourStringValue\\\"\".")
	s.Command.Flags().BoolVar(&s.LogConfig, "log-config", false, "Log the server config to stderr.")
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
		s.Command.Long = "Inspect and update Task Queues, the queues that Worker entities poll for\nWorkflow and Activity tasks:\n\n\x1b[1mtemporal task-queue [command] [command options]\x1b[0m\n\nFor example:\n\n\x1b[1mtemporal task-queue describe --task-queue \"YourTaskQueueName\"\x1b[0m"
	} else {
		s.Command.Long = "Inspect and update Task Queues, the queues that Worker entities poll for\nWorkflow and Activity tasks:\n\n```\ntemporal task-queue [command] [command options]\n```\n\nFor example:\n\n```\ntemporal task-queue describe --task-queue \"YourTaskQueueName\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.AddCommand(&NewTemporalTaskQueueDescribeCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueGetBuildIdReachabilityCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueGetBuildIdsCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueListPartitionCommand(cctx, &s).Command)
	s.Command.AddCommand(&NewTemporalTaskQueueUpdateBuildIdsCommand(cctx, &s).Command)
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
	s.Command.Short = "Show active Workers"
	if hasHighlighting {
		s.Command.Long = "Display a list of active Workers that have recently polled a \nTask Queue. The Temporal Server records each poll\nrequest time. A \x1b[1mLastAccessTime\x1b[0m over one minute may indicate the Worker is at \ncapacity or has shut down. Temporal Workers are removed if 5 minutes \nhave passed since the last poll request:\n\n\x1b[1mtemporal task-queue describe \\\n   --task-queue \"YourTaskQueueName\"\x1b[0m\n\nWorkflow and Activity polling use separate Task Queues:\n\n\x1b[1mtemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type \"activity\"`\x1b[0m"
	} else {
		s.Command.Long = "Display a list of active Workers that have recently polled a \nTask Queue. The Temporal Server records each poll\nrequest time. A `LastAccessTime` over one minute may indicate the Worker is at \ncapacity or has shut down. Temporal Workers are removed if 5 minutes \nhave passed since the last poll request:\n\n```\ntemporal task-queue describe \\\n   --task-queue \"YourTaskQueueName\"\n```\n\nWorkflow and Activity polling use separate Task Queues:\n\n```\ntemporal task-queue describe \\\n    --task-queue YourTaskQueue \\\n    --task-queue-type \"activity\"`\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task queue name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.TaskQueueType = NewStringEnum([]string{"workflow", "activity"}, "workflow")
	s.Command.Flags().Var(&s.TaskQueueType, "task-queue-type", "Task Queue type. Accepted values: workflow, activity.")
	s.Command.Flags().IntVar(&s.Partitions, "partitions", 0, "Query partitions 1 through `N`. Experimental/Temporary feature. Default: 1")
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
	s.Command.Short = "Build ID availability"
	if hasHighlighting {
		s.Command.Long = "Shows if a given Build ID can be used for new, existing, or closed workflows.\nYou can specify the '--build-id' and '--task-queue' flags multiple times.\nWhen '--task-queue' is omitted, checks Build ID reachability against all\ntask queues. For example:\n\n\x1b[1mtemporal task-queue get-build-id-reachability --build-id \"2.0\"\x1b[0m"
	} else {
		s.Command.Long = "Shows if a given Build ID can be used for new, existing, or closed workflows.\nYou can specify the '--build-id' and '--task-queue' flags multiple times.\nWhen '--task-queue' is omitted, checks Build ID reachability against all\ntask queues. For example:\n\n```\ntemporal task-queue get-build-id-reachability --build-id \"2.0\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringArrayVar(&s.BuildId, "build-id", nil, "One or more Build ID strings. Can be passed multiple times.")
	s.ReachabilityType = NewStringEnum([]string{"open", "closed", "existing"}, "existing")
	s.Command.Flags().Var(&s.ReachabilityType, "reachability-type", "Reachability filter. `open`: reachable by one or more open workflows. `closed`: reachable by one or more closed workflows. `existing`: reachable by either. New Workflow Executions reachable by a Build ID are always reported. Accepted values: open, closed, existing.")
	s.Command.Flags().StringArrayVarP(&s.TaskQueue, "task-queue", "t", nil, "List of Task Queues. Constrains the reachability search. Can be passed multiple times.")
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
	s.Command.Short = "Build ID versions"
	if hasHighlighting {
		s.Command.Long = "Fetches sets of compatible Build IDs for specified Task Queues and displays\ntheir information:\n\n\x1b[1mtemporal task-queue get-build-ids \\\n    --task-queue \"YourTaskQueue\"\x1b[0m"
	} else {
		s.Command.Long = "Fetches sets of compatible Build IDs for specified Task Queues and displays\ntheir information:\n\n```\ntemporal task-queue get-build-ids \\\n    --task-queue \"YourTaskQueue\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task queue name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "task-queue")
	s.Command.Flags().IntVar(&s.MaxSets, "max-sets", 0, "Max return count. Use 1 for default [major-version](https://en.wikipedia.org/wiki/Software_versioning) set. Use 0 for all sets.")
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
	s.Command.Short = "List partitions"
	if hasHighlighting {
		s.Command.Long = "Display a Task Queue's partition list with assigned matching nodes:\n\n\x1b[1mtemporal task-queue list-partition \\\n    --task-queue \"YourTaskQueue\"\x1b[0m"
	} else {
		s.Command.Long = "Display a Task Queue's partition list with assigned matching nodes:\n\n```\ntemporal task-queue list-partition \\\n    --task-queue \"YourTaskQueue\"\n```"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Task Queue name. Required")
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
	s.Command.Short = "Operations to manage Build ID versions on a Task Queue"
	s.Command.Long = "Provides various commands for adding or changing the sets of compatible build IDs associated with a Task Queue. See the help of each sub-command for more"
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
	s.Command.Short = "Add a new build ID compatible with an existing ID to a Task Queue's version sets"
	s.Command.Long = "The new build ID will become the default for the set containing the existing ID. See per-flag help for more"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "The new build ID to be added. Required")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue. Required")
	s.Command.Flags().StringVar(&s.ExistingCompatibleBuildId, "existing-compatible-build-id", "", "A build ID which must already exist in the version sets known by the task queue. The new ID will be stored in the set containing this ID, marking it as compatible with the versions within. Required")
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
	s.Command.Short = "Add a new default (incompatible) build ID to a Task Queue's 7 sets"
	s.Command.Long = "Creates a new build ID set which will become the new overall default for the queue with the provided build ID as its only member. This new set is incompatible with all previous sets/versions"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "The new build ID to be added. Required")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue. Required")
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
	s.Command.Short = "Promote a build ID to become the default for its containing set"
	s.Command.Long = "New tasks compatible with the set will be dispatched to the default ID"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "An existing build ID which will be promoted to be the default inside its containing set. Required")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue. Required")
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
	s.Command.Short = "Promote a build ID set to become the default for a Task Queue"
	s.Command.Long = "If the set is already the default, this command has no effect"
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "An existing build ID whose containing set will be promoted. Required")
	s.Command.Flags().StringVarP(&s.TaskQueue, "task-queue", "t", "", "Name of the Task Queue. Required")
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
	cctx.BindFlagEnvVar(f.Lookup("namespace"), "TEMPORAL_NAMESPAC")
	f.StringVar(&v.ApiKey, "api-key", "", "Sets the API key on requests.")
	cctx.BindFlagEnvVar(f.Lookup("api-key"), "TEMPORAL_API_KE")
	f.StringArrayVar(&v.GrpcMeta, "grpc-meta", nil, "HTTP headers to send with requests (formatted as \"KEY=VALUE\")")
	f.BoolVar(&v.Tls, "tls", false, "Enable TLS encryption without additional options such as mTLS or client certificates.")
	cctx.BindFlagEnvVar(f.Lookup("tls"), "TEMPORAL_TL")
	f.StringVar(&v.TlsCertPath, "tls-cert-path", "", "Path to x509 certificate.")
	cctx.BindFlagEnvVar(f.Lookup("tls-cert-path"), "TEMPORAL_TLS_CER")
	f.StringVar(&v.TlsKeyPath, "tls-key-path", "", "Path to private certificate key.")
	cctx.BindFlagEnvVar(f.Lookup("tls-key-path"), "TEMPORAL_TLS_KE")
	f.StringVar(&v.TlsCaPath, "tls-ca-path", "", "Path to server CA certificate.")
	cctx.BindFlagEnvVar(f.Lookup("tls-ca-path"), "TEMPORAL_TLS_C")
	f.StringVar(&v.TlsCertData, "tls-cert-data", "", "Data for x509 certificate. Exclusive with -path variant.")
	cctx.BindFlagEnvVar(f.Lookup("tls-cert-data"), "TEMPORAL_TLS_CERT_DAT")
	f.StringVar(&v.TlsKeyData, "tls-key-data", "", "Data for private certificate key. Exclusive with -path variant.")
	cctx.BindFlagEnvVar(f.Lookup("tls-key-data"), "TEMPORAL_TLS_KEY_DAT")
	f.StringVar(&v.TlsCaData, "tls-ca-data", "", "Data for server CA certificate. Exclusive with -path variant.")
	cctx.BindFlagEnvVar(f.Lookup("tls-ca-data"), "TEMPORAL_TLS_CA_DAT")
	f.BoolVar(&v.TlsDisableHostVerification, "tls-disable-host-verification", false, "Disables TLS host-name verification.")
	cctx.BindFlagEnvVar(f.Lookup("tls-disable-host-verification"), "TEMPORAL_TLS_DISABLE_HOST_VERIFICATIO")
	f.StringVar(&v.TlsServerName, "tls-server-name", "", "Overrides target TLS server name.")
	cctx.BindFlagEnvVar(f.Lookup("tls-server-name"), "TEMPORAL_TLS_SERVER_NAM")
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "Endpoint for a remote Codec Server.")
	cctx.BindFlagEnvVar(f.Lookup("codec-endpoint"), "TEMPORAL_CODEC_ENDPOIN")
	f.StringVar(&v.CodecAuth, "codec-auth", "", "Sets the authorization header on requests to the Codec Server.")
	cctx.BindFlagEnvVar(f.Lookup("codec-auth"), "TEMPORAL_CODEC_AUT")
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
		s.Command.Long = "Workflow commands perform operations on Workflow Executions\n\nWorkflow commands use this syntax: \x1b[1mtemporal workflow COMMAND [ARGS]\x1b[0m"
	} else {
		s.Command.Long = "Workflow commands perform operations on Workflow Executions\n\nWorkflow commands use this syntax: `temporal workflow COMMAND [ARGS]`"
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
	s.Command.Short = "Cancel a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow cancel\x1b[0m command is used to cancel a Workflow Execution)\nCanceling a running Workflow Execution records a \x1b[1mWorkflowExecutionCancelRequested\x1b[0m event in the Event History. A new\nCommand Task will be scheduled, and the Workflow Execution will perform cleanup work\n\nExecutions may be cancelled by Workflow ID:\n\x1b[1mtemporal workflow cancel --workflow-id YourWorkflowId\x1b[0m\n\n...or in bulk via a visibility query list filter:\n\x1b[1mtemporal workflow cancel --query YourQuery\x1b[0m\n\nUse the options listed below to change the behavior of this command"
	} else {
		s.Command.Long = "The `temporal workflow cancel` command is used to cancel a Workflow Execution)\nCanceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History. A new\nCommand Task will be scheduled, and the Workflow Execution will perform cleanup work\n\nExecutions may be cancelled by Workflow ID:\n```\ntemporal workflow cancel --workflow-id YourWorkflowId\n```\n\n...or in bulk via a visibility query list filter:\n```\ntemporal workflow cancel --query YourQuery\n```\n\nUse the options listed below to change the behavior of this command"
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
	s.Command.Short = "Count Workflow Executions"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow count\x1b[0m command returns a count of Workflow Executions.\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "The `temporal workflow count` command returns a count of Workflow Executions.\n\nUse the options listed below to change the command's behavior"
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
	s.Command.Short = "Delete a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow delete\x1b[0m command is used to delete a specific Workflow Executions.\nThis asynchronously deletes a workflow's Event History.\nIf the Workflow Executions is Running, it will be terminated before deletion.\n\n\x1b[1mtemporal workflow delete \\\n\t\t--workflow-id YourWorkflowId \\\x1b[0m\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "The `temporal workflow delete` command is used to delete a specific Workflow Executions.\nThis asynchronously deletes a workflow's Event History.\nIf the Workflow Executions is Running, it will be terminated before deletion.\n\n```\ntemporal workflow delete \\\n\t\t--workflow-id YourWorkflowId \\\n```\n\nUse the options listed below to change the command's behavior"
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
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID.")
	_ = cobra.MarkFlagRequired(f, "workflow-id")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID.")
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
	s.Command.Short = "Show information about a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow describe\x1b[0m command shows information about a given\nWorkflow Executions.\n\nThis information can be used to locate Workflow Executions that weren't able to run successfully\n\n\x1b[1mtemporal workflow describe --workflow-id meaningful-business-id\x1b[0m\n\nOutput can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points\n\n\x1b[1mtemporal workflow describe --workflow-id meaningful-business-id --raw true --reset-points true\x1b[0m\n\nUse the command options below to change the information returned by this command"
	} else {
		s.Command.Long = "The `temporal workflow describe` command shows information about a given\nWorkflow Executions.\n\nThis information can be used to locate Workflow Executions that weren't able to run successfully\n\n`temporal workflow describe --workflow-id meaningful-business-id`\n\nOutput can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points\n\n`temporal workflow describe --workflow-id meaningful-business-id --raw true --reset-points true`\n\nUse the command options below to change the information returned by this command"
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
	s.Command.Short = "Start a new Workflow Execution and prints its progress"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow execute\x1b[0m command starts a new Workflow Executions and\nprints its progress. The command completes when the Workflow Execution completes\n\nSingle quotes('') are used to wrap input as JSON.\n\n\x1b[1mtemporal workflow execute\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type YourWorkflow \\\n\t\t--task-queue YourTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m"
	} else {
		s.Command.Long = "The `temporal workflow execute` command starts a new Workflow Executions and\nprints its progress. The command completes when the Workflow Execution completes\n\nSingle quotes('') are used to wrap input as JSON.\n\n```\ntemporal workflow execute\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type YourWorkflow \\\n\t\t--task-queue YourTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```"
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
	s.Command.Short = "Updates an event history JSON file to the current format"
	if hasHighlighting {
		s.Command.Long = "\x1b[1mtemporal workflow fix-history-json \\\n\t--source original.json \\\n\t--target reserialized.json\x1b[0m\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "```\ntemporal workflow fix-history-json \\\n\t--source original.json \\\n\t--target reserialized.json\n```\n\nUse the options listed below to change the command's behavior"
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
	s.Command.Short = "List Workflow Executions based on a Query"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow list\x1b[0m command provides a list of Workflow Executions\nthat meet the criteria of a given Query.\nBy default, this command returns up to 10 closed Workflow Executions\n\n\x1b[1mtemporal workflow list --query YourQuery\x1b[0m\n\nThe command can also return a list of archived Workflow Executions\n\n\x1b[1mtemporal workflow list --archived\x1b[0m\n\nUse the command options below to change the information returned by this command"
	} else {
		s.Command.Long = "The `temporal workflow list` command provides a list of Workflow Executions\nthat meet the criteria of a given Query.\nBy default, this command returns up to 10 closed Workflow Executions\n\n`temporal workflow list --query YourQuery`\n\nThe command can also return a list of archived Workflow Executions\n\n`temporal workflow list --archived`\n\nUse the command options below to change the information returned by this command"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Filter results using a SQL-like query.")
	s.Command.Flags().BoolVar(&s.Archived, "archived", false, "If set, will only query and list archived workflows instead of regular workflows.")
	s.Command.Flags().IntVar(&s.Limit, "limit", 0, "Max count to list.")
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
	s.Command.Short = "Query a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow query\x1b[0m command is used to Query a\nWorkflow Execution\nby Workflow ID.\n\n\x1b[1mtemporal workflow query \\\n\t\t--workflow-id YourWorkflowId \\\n\t\t--name YourQuery \\\n\t\t--input '{\"YourInputKey\": \"YourInputValue\"}'\x1b[0m\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "The `temporal workflow query` command is used to Query a\nWorkflow Execution\nby Workflow ID.\n\n```\ntemporal workflow query \\\n\t\t--workflow-id YourWorkflowId \\\n\t\t--name YourQuery \\\n\t\t--input '{\"YourInputKey\": \"YourInputValue\"}'\n```\n\nUse the options listed below to change the command's behavior"
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
	s.Command.Short = "Reset a Workflow Execution to an older point in history"
	if hasHighlighting {
		s.Command.Long = "The temporal workflow reset command resets a Workflow Execution\nA reset allows the Workflow to resume from a certain point without losing its parameters or Event History\n\nThe Workflow Execution can be set to a given Event Type:\n\x1b[1mtemporal workflow reset --workflow-id meaningful-business-id --type LastContinuedAsNew\x1b[0m\n\n...or a specific any Event after \x1b[1mWorkflowTaskStarted\x1b[0m\n\x1b[1mtemporal workflow reset --workflow-id meaningful-business-id --event-id YourLastEvent\x1b[0m\nFor batch reset only FirstWorkflowTask, LastWorkflowTask or BuildId can be used. Workflow ID, run ID and event ID\nshould not be set\nUse the options listed below to change reset behavior"
	} else {
		s.Command.Long = "The temporal workflow reset command resets a Workflow Execution\nA reset allows the Workflow to resume from a certain point without losing its parameters or Event History\n\nThe Workflow Execution can be set to a given Event Type:\n```\ntemporal workflow reset --workflow-id meaningful-business-id --type LastContinuedAsNew\n```\n\n...or a specific any Event after `WorkflowTaskStarted`\n```\ntemporal workflow reset --workflow-id meaningful-business-id --event-id YourLastEvent\n```\nFor batch reset only FirstWorkflowTask, LastWorkflowTask or BuildId can be used. Workflow ID, run ID and event ID\nshould not be set\nUse the options listed below to change reset behavior"
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow ID. Required for non-batch reset operations.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID.")
	s.Command.Flags().IntVarP(&s.EventId, "event-id", "e", 0, "The Event ID for any Event after `WorkflowTaskStarted` you want to reset to (exclusive). It can be `WorkflowTaskCompleted`, `WorkflowTaskFailed` or others.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "The reason why this workflow is being reset.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "reason")
	s.ReapplyType = NewStringEnum([]string{"All", "Signal", "None"}, "All")
	s.Command.Flags().Var(&s.ReapplyType, "reapply-type", "Event types to reapply after the reset point. Accepted values: All, Signal, None.")
	s.Type = NewStringEnum([]string{"FirstWorkflowTask", "LastWorkflowTask", "LastContinuedAsNew", "BuildId"}, "")
	s.Command.Flags().VarP(&s.Type, "type", "t", "Event type to which you want to reset. Accepted values: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew, BuildId.")
	s.Command.Flags().StringVar(&s.BuildId, "build-id", "", "Only used if type is BuildId. Reset the first workflow task processed by this build ID. Note that by default, this reset is allowed to be to a prior run in a chain of continue-as-new.")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Start a batch reset to operate on Workflow Executions with given List Filter.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Override confirmation request before reset. Only allowed when `--query` is present.")
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
	s.Command.Short = "Show Event History for a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow show\x1b[0m command provides the Event History for a\nWorkflow Execution. With JSON output specified, this output can be given to\nan SDK to perform a replay\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "The `temporal workflow show` command provides the Event History for a\nWorkflow Execution. With JSON output specified, this output can be given to\nan SDK to perform a replay\n\nUse the options listed below to change the command's behavior"
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
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID. Either this or --query must be set.")
	f.StringVarP(&v.RunId, "run-id", "r", "", "Run ID. Cannot be set when --query is set.")
	f.StringVarP(&v.Query, "query", "q", "", "Start a batch to operate on Workflow Executions with given List Filter. Either --query or --workflow-id must be set.")
	f.StringVar(&v.Reason, "reason", "", "Reason to perform batch. Only allowed if query is present unless the command specifies otherwise. Defaults to message with the current user's name.")
	f.BoolVarP(&v.Yes, "yes", "y", false, "Override confirmation request before signalling. Only allowed when `--query` is present.")
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
	s.Command.Short = "Signal a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow signal\x1b[0m command is used to Signal a\nWorkflow Execution by Workflow ID\n\n\x1b[1mtemporal workflow signal \\\n\t\t--workflow-id YourWorkflowId \\\n\t\t--name YourSignal \\\n\t\t--input '{\"YourInputKey\": \"YourInputValue\"}'\x1b[0m\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "The `temporal workflow signal` command is used to Signal a\nWorkflow Execution by Workflow ID\n\n```\ntemporal workflow signal \\\n\t\t--workflow-id YourWorkflowId \\\n\t\t--name YourSignal \\\n\t\t--input '{\"YourInputKey\": \"YourInputValue\"}'\n```\n\nUse the options listed below to change the command's behavior"
	}
	s.Command.Args = cobra.NoArgs
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.Name, "name", "", "Signal Name. Required")
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
	s.Command.Short = "Show the stack trace of a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow stack\x1b[0m command Queries a\nWorkflow Execution with \x1b[1m__stack_trace\x1b[0m as the query type\nThis returns a stack trace of all the threads or routines currently used by the workflow, and is\nuseful for troubleshooting\n\n\x1b[1mtemporal workflow stack --workflow-id YourWorkflowId\x1b[0m\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "The `temporal workflow stack` command Queries a\nWorkflow Execution with `__stack_trace` as the query type\nThis returns a stack trace of all the threads or routines currently used by the workflow, and is\nuseful for troubleshooting\n\n```\ntemporal workflow stack --workflow-id YourWorkflowId\n```\n\nUse the options listed below to change the command's behavior"
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
	f.StringVarP(&v.WorkflowId, "workflow-id", "w", "", "Workflow ID.")
	f.StringVar(&v.Type, "type", "", "Workflow Type name.")
	_ = cobra.MarkFlagRequired(f, "type")
	f.StringVarP(&v.TaskQueue, "task-queue", "t", "", "Workflow Task queue.")
	_ = cobra.MarkFlagRequired(f, "task-queue")
	f.DurationVar(&v.RunTimeout, "run-timeout", 0, "Fail a Workflow Run if it takes longer than `DURATION`.")
	f.DurationVar(&v.ExecutionTimeout, "execution-timeout", 0, "Fail a WorkflowExecution if it takes longer than `DURATION`, including retries and ContinueAsNew tasks.")
	f.DurationVar(&v.TaskTimeout, "task-timeout", 0, "Fail a Workflow Task if it takes longer than `DURATION`. (Start-to-close timeout for a Workflow Task.) Default: 10s.")
	f.StringArrayVar(&v.SearchAttribute, "search-attribute", nil, "Passes Search Attribute in \"KEY=VALUE\" format. Use valid JSON formats for value.")
	f.StringArrayVar(&v.Memo, "memo", nil, "Passes Memo in \"KEY=VALUE\" format. Use valid JSON formats for value.")
}

type WorkflowStartOptions struct {
	Cron          string
	FailExisting  bool
	StartDelay    time.Duration
	IdReusePolicy string
}

func (v *WorkflowStartOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringVar(&v.Cron, "cron", "", "Cron schedule for the Workflow. Deprecated - use schedules instead.")
	f.BoolVar(&v.FailExisting, "fail-existing", false, "Fail if the Workflow already exists.")
	f.DurationVar(&v.StartDelay, "start-delay", 0, "Wait before starting the Workflow. Cannot be used with a cron schedule. If the Workflow receives a signal or update before the delay has elapsed, it will start immediately.")
	f.StringVar(&v.IdReusePolicy, "id-reuse-policy", "", "Allow the same Workflow ID to be used in a new Workflow Execution. Accepted values: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.")
}

type PayloadInputOptions struct {
	Input       []string
	InputFile   []string
	InputMeta   []string
	InputBase64 bool
}

func (v *PayloadInputOptions) buildFlags(cctx *CommandContext, f *pflag.FlagSet) {
	f.StringArrayVarP(&v.Input, "input", "i", nil, "Input value (default JSON unless --input-meta is non-JSON encoding). Can be passed multiple times for multiple arguments. Cannot be combined with --input-file. Can be passed multiple times.")
	f.StringArrayVar(&v.InputFile, "input-file", nil, "Read `PATH` as the input value (JSON by default unless --input-meta is non-JSON encoding). Can be passed multiple times for multiple arguments. Cannot be combined with --input. Can be passed multiple times.")
	f.StringArrayVar(&v.InputMeta, "input-meta", nil, "Metadata for the input payload, specified as a \"KEY=VALUE\" pair. If KEY is \"encoding\", overrides the default of \"json/plain\". Pass multiple --input-meta options to set multiple pairs. Can be passed multiple times.")
	f.BoolVar(&v.InputBase64, "input-base64", false, "Assume inputs are base64-encoded and attempt to decode them.")
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
	s.Command.Short = "Start a new Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow start\x1b[0m command starts a new Workflow Execution. The\nWorkflow and Run IDs are returned after starting the Workflow.\n\n\n\x1b[1mtemporal workflow start \\\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type YourWorkflow \\\n\t\t--task-queue YourTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m"
	} else {
		s.Command.Long = "The `temporal workflow start` command starts a new Workflow Execution. The\nWorkflow and Run IDs are returned after starting the Workflow.\n\n\n```\ntemporal workflow start \\\n\t\t--workflow-id meaningful-business-id \\\n\t\t--type YourWorkflow \\\n\t\t--task-queue YourTaskQueue \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```"
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
	s.Command.Short = "Terminate a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow terminate\x1b[0m command is used to terminate a\nWorkflow Execution. Canceling a running Workflow Execution records a\n\x1b[1mWorkflowExecutionTerminated\x1b[0m event as the closing Event in the workflow's Event History. Workflow code is oblivious to\ntermination. Use \x1b[1mtemporal workflow cancel\x1b[0m if you need to perform cleanup in your workflow\n\nExecutions may be terminated by Workflow ID with an optional reason:\n\x1b[1mtemporal workflow terminate [--reason my-reason] --workflow-id YourWorkflowId\x1b[0m\n\n...or in bulk via a visibility query list filter:\n\x1b[1mtemporal workflow terminate --query YourQuery\x1b[0m\n\nUse the options listed below to change the behavior of this command."
	} else {
		s.Command.Long = "The `temporal workflow terminate` command is used to terminate a\nWorkflow Execution. Canceling a running Workflow Execution records a\n`WorkflowExecutionTerminated` event as the closing Event in the workflow's Event History. Workflow code is oblivious to\ntermination. Use `temporal workflow cancel` if you need to perform cleanup in your workflow\n\nExecutions may be terminated by Workflow ID with an optional reason:\n```\ntemporal workflow terminate [--reason my-reason] --workflow-id YourWorkflowId\n```\n\n...or in bulk via a visibility query list filter:\n```\ntemporal workflow terminate --query YourQuery\n```\n\nUse the options listed below to change the behavior of this command."
	}
	s.Command.Args = cobra.NoArgs
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow ID. Either this or query must be set.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run ID. Cannot be set when query is set.")
	s.Command.Flags().StringVarP(&s.Query, "query", "q", "", "Start a batch to terminate Workflow Executions with the `QUERY` List Filter. Either this or Workflow ID must be set.")
	s.Command.Flags().StringVar(&s.Reason, "reason", "", "Reason for termination. Defaults to message with the current user's name.")
	s.Command.Flags().BoolVarP(&s.Yes, "yes", "y", false, "Override confirmation request before termination. Only allowed when `--query` is present.")
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
	s.Command.Short = "Interactively show the progress of a Workflow Execution"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow trace\x1b[0m command displays the progress of a Workflow Execution and its child workflows with a real-time trace.\nThis view provides a great way to understand the flow of a workflow.\n\nUse the options listed below to change the behavior of this command."
	} else {
		s.Command.Long = "The `temporal workflow trace` command displays the progress of a Workflow Execution and its child workflows with a real-time trace.\nThis view provides a great way to understand the flow of a workflow.\n\nUse the options listed below to change the behavior of this command."
	}
	s.Command.Args = cobra.NoArgs
	s.WorkflowReferenceOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringArrayVar(&s.Fold, "fold", nil, "Fold Child Workflows with the specified `STATUS`. To specify multiple statuses, pass --fold multiple times. This will reduce the amount of information fetched and displayed. Case-insensitive. Ignored if --no-fold supplied. Available values: running, completed, failed, canceled, terminated, timedout, continueasnew. Can be passed multiple times.")
	s.Command.Flags().BoolVar(&s.NoFold, "no-fold", false, "Disable folding. All Child Workflows within the set depth will be fetched and displayed.")
	s.Command.Flags().IntVar(&s.Depth, "depth", -1, "Fetch up to N Child Workflows deep. Use -1 to fetch child workflows at any depth.")
	s.Command.Flags().IntVar(&s.Concurrency, "concurrency", 10, "Fetch up to N Workflow Histories at a time.")
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
	s.Command.Short = "Update a running workflow synchronously"
	if hasHighlighting {
		s.Command.Long = "The \x1b[1mtemporal workflow update\x1b[0m command is used to synchronously Update a\nWorkflow Execution by Workflow ID\n\n\x1b[1mtemporal workflow update \\\n\t\t--workflow-id YourWorkflowId \\\n\t\t--name YourUpdate \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\x1b[0m\n\nUse the options listed below to change the command's behavior"
	} else {
		s.Command.Long = "The `temporal workflow update` command is used to synchronously Update a\nWorkflow Execution by Workflow ID\n\n```\ntemporal workflow update \\\n\t\t--workflow-id YourWorkflowId \\\n\t\t--name YourUpdate \\\n\t\t--input '{\"Input\": \"As-JSON\"}'\n```\n\nUse the options listed below to change the command's behavior"
	}
	s.Command.Args = cobra.NoArgs
	s.PayloadInputOptions.buildFlags(cctx, s.Command.Flags())
	s.Command.Flags().StringVar(&s.Name, "name", "", "Update Name.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "name")
	s.Command.Flags().StringVarP(&s.WorkflowId, "workflow-id", "w", "", "Workflow `ID`.")
	_ = cobra.MarkFlagRequired(s.Command.Flags(), "workflow-id")
	s.Command.Flags().StringVar(&s.UpdateId, "update-id", "", "Update `ID`. If unset, default to a UUID.")
	s.Command.Flags().StringVarP(&s.RunId, "run-id", "r", "", "Run `ID`. If unset, the currently running Workflow Execution receives the Update.")
	s.Command.Flags().StringVar(&s.FirstExecutionRunId, "first-execution-run-id", "", "Send the Update to the last Workflow Execution in the chain that started with `ID`.")
	s.Command.Run = func(c *cobra.Command, args []string) {
		if err := s.run(cctx, args); err != nil {
			cctx.Options.Fail(err)
		}
	}
	return &s
}
