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

func BuildApp(version string) *cli.App {
	defaultCfg, _ := sconfig.NewDefaultConfig()

	app := cli.NewApp()
	app.Name = "temporal"
	app.Usage = "Temporal command-line interface and development server"
	if version == "" {
		version = headers.CLIVersion
	}
	app.Version = fmt.Sprintf("%s (server %s) (ui %s)", version, sheaders.ServerVersion, uiversion.UIVersion)
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
	env.Build(ctx)
	client.Build(ctx)

	return nil
}

func HandleError(c *cli.Context, err error) {
	if err == nil {
		return
	}

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
			Usage:       "Commands for managing Temporal server",
			Subcommands: server.NewServerCommands(defaultCfg),
		},
	}
}

var clientCommands = []*cli.Command{
	{
		Name:        "workflow",
		Usage:       "Operations on Workflows",
		Subcommands: workflow.NewWorkflowCommands(),
	},
	{
		Name:        "activity",
		Usage:       "Operations on Activities of Workflows",
		Subcommands: activity.NewActivityCommands(),
	},
	{
		Name:        "task-queue",
		Usage:       "Operations on Task Queues",
		Subcommands: taskqueue.NewTaskQueueCommands(),
	},
	{
		Name:        "schedule",
		Usage:       "Operations on Schedules",
		Subcommands: schedule.NewScheduleCommands(),
	},

	{
		Name:        "batch",
		Usage:       "Operations on Batch jobs. Use workflow commands with --query flag to start batch jobs",
		Subcommands: batch.NewBatchCommands(),
	},
	{
		Name:  "operator",
		Usage: "Operation on Temporal server",
		Subcommands: []*cli.Command{
			{
				Name:        "namespace",
				Usage:       "Operations on namespaces",
				Subcommands: namespace.NewNamespaceCommands(),
			},
			{
				Name:        "search-attribute",
				Usage:       "Operations on search attributes",
				Subcommands: searchattribute.NewSearchAttributeCommands(),
			},
			{
				Name:        "cluster",
				Usage:       "Operations on a Temporal cluster",
				Subcommands: cluster.NewClusterCommands(),
			},
		},
	},
	{
		Name:        "env",
		Usage:       "Manage client environment configurations",
		Subcommands: env.NewEnvCommands(),
	},
}

var completionCommands = []*cli.Command{
	{
		Name:        "completion",
		Usage:       "Output shell completion code for the specified shell (zsh, bash)",
		Subcommands: completion.NewCompletionCommands(),
	},
}
