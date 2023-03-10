package app

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/temporalio/cli/activity"
	"github.com/temporalio/cli/batch"
	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/cluster"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/cli/completion"
	"github.com/temporalio/cli/env"
	"github.com/temporalio/cli/headers"
	"github.com/temporalio/cli/namespace"
	"github.com/temporalio/cli/schedule"
	"github.com/temporalio/cli/searchattribute"
	"github.com/temporalio/cli/server"
	sconfig "github.com/temporalio/cli/server/config"
	"github.com/temporalio/cli/taskqueue"
	"github.com/temporalio/cli/workflow"
	"github.com/temporalio/tctl-kit/pkg/color"
	uiversion "github.com/temporalio/ui-server/v2/server/version"
	"github.com/urfave/cli/v2"
	sheaders "go.temporal.io/server/common/headers"
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite" // load sqlite storage driver
)

func BuildApp() *cli.App {
	defaultCfg, _ := sconfig.NewDefaultConfig()

	app := cli.NewApp()
	app.Name = "temporal"
	app.Usage = "Temporal command-line interface and development server"
	app.Version = fmt.Sprintf("%s (server %s) (ui %s)", headers.Version,
		sheaders.ServerVersion, uiversion.UIVersion)
	app.Suggest = true
	app.EnableBashCompletion = true
	app.DisableSliceFlagSeparator = true
	app.Commands = commands(defaultCfg)
	app.Before = configureCLI
	app.ExitErrHandler = HandleError

	// set builder if not customized
	if client.CFactory == nil {
		SetFactory(client.NewClientFactory())
	}

	return app
}

// SetFactory is used to set the ClientFactory global
func SetFactory(factory client.ClientFactory) {
	client.CFactory = factory
}

func configureCLI(ctx *cli.Context) error {
	env.Init(ctx)
	client.Init(ctx)
	headers.Init()

	return nil
}

func HandleError(c *cli.Context, err error) {
	if err == nil {
		return
	}

	cli.HandleExitCoder(err)

	fmt.Fprintf(os.Stderr, "%s %+v\n", color.Red(c, "Error:"), err)
	if os.Getenv(common.ShowErrorStackEnv) != `` {
		fmt.Fprintln(os.Stderr, color.Magenta(c, "Stack trace:"))
		debug.PrintStack()
	} else {
		fmt.Fprintf(os.Stderr, "('export %s=1' to see stack traces)\n", common.ShowErrorStackEnv)
	}

	cli.OsExiter(1)
}

func commands(defaultCfg *sconfig.Config) []*cli.Command {
	return append(append(serverCommands(defaultCfg), common.WithFlags(clientCommands, common.SharedFlags)...), completionCommands...)
}

func serverCommands(defaultCfg *sconfig.Config) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "server",
			Usage:       "Commands for managing the Temporal Server.",
			UsageText:   common.ServerUsageText,
			Subcommands: server.NewServerCommands(defaultCfg),
		},
	}
}

var clientCommands = []*cli.Command{
	{
		Name:        "workflow",
		Usage:       common.WorkflowDefinition,
		UsageText:   common.WorkflowUsageText,
		Subcommands: workflow.NewWorkflowCommands(),
	},
	{
		Name:        "activity",
		Usage:       common.ActivityDefinition,
		UsageText:   common.ActivityUsageText,
		Subcommands: activity.NewActivityCommands(),
	},
	{
		Name:        "task-queue",
		Usage:       common.TaskQueueDefinition,
		UsageText:   common.TaskQueueUsageText,
		Subcommands: taskqueue.NewTaskQueueCommands(),
	},
	{
		Name:        "schedule",
		Usage:       common.ScheduleDefinition,
		UsageText:   common.ScheduleUsageText,
		Subcommands: schedule.NewScheduleCommands(),
	},
	{
		Name:        "batch",
		Usage:       common.BatchDefinition,
		UsageText:   common.BatchUsageText,
		Subcommands: batch.NewBatchCommands(),
	},
	{
		Name:      "operator",
		Usage:     common.OperatorDefinition,
		UsageText: common.OperatorUsageText,
		Subcommands: []*cli.Command{
			{
				Name:        "namespace",
				Usage:       common.NamespaceDefinition,
				UsageText:   common.NamespaceUsageText,
				Subcommands: namespace.NewNamespaceCommands(),
			},
			{
				Name:        "search-attribute",
				Usage:       common.SearchAttributeDefinition,
				UsageText:   common.SearchAttributeUsageText,
				Subcommands: searchattribute.NewSearchAttributeCommands(),
			},
			{
				Name:        "cluster",
				Usage:       common.ClusterDefinition,
				UsageText:   common.ClusterUsageText,
				Subcommands: cluster.NewClusterCommands(),
			},
		},
	},
	{
		Name:        "env",
		Usage:       common.EnvDefinition,
		UsageText:   common.EnvUsageText,
		Subcommands: env.NewEnvCommands(),
	},
}

var completionCommands = []*cli.Command{
	{
		Name:        "completion",
		Usage:       "Output shell completion code for the specified shell (zsh, bash).",
		Subcommands: completion.NewCompletionCommands(),
	},
}
