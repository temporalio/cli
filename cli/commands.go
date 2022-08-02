package cli

import (
	"github.com/urfave/cli/v2"
)

var tctlCommands = []*cli.Command{
	{
		Name:        "namespace",
		Aliases:     []string{"n"},
		Usage:       "Operations on namespaces",
		Subcommands: newNamespaceCommands(),
	},
	{
		Name:        "workflow",
		Aliases:     []string{"w"},
		Usage:       "Operations on workflows",
		Subcommands: newWorkflowCommands(),
	},
	{
		Name:        "activity",
		Aliases:     []string{"a"},
		Usage:       "Operations on activities of workflows",
		Subcommands: newActivityCommands(),
	},
	{
		Name:        "task-queue",
		Aliases:     []string{"tq"},
		Usage:       "Operations on task queues",
		Subcommands: newTaskQueueCommands(),
	},
	{
		Name:        "schedule",
		Aliases:     []string{"s"},
		Usage:       "Operations on schedules",
		Subcommands: newScheduleCommands(),
	},
	{
		Name:        "batch",
		Usage:       "Batch operations on a list of workflows from a query",
		Subcommands: newBatchCommands(),
	},
	{
		Name:        "cluster",
		Usage:       "Operations on a Temporal cluster",
		Subcommands: newClusterCommands(),
	},
	{
		Name:        "data-converter",
		Aliases:     []string{"dc"},
		Usage:       "Operations using a custom data converter",
		Subcommands: newDataConverterCommands(),
	},
	{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "Configure tctl",
		Subcommands: newConfigCommands(),
	},
	newAliasCommand(),
}

var sharedFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagNamespace,
		Aliases: FlagNamespaceAlias,
		Value:   "default",
		Usage:   "Namespace to operate on",
		EnvVars: []string{"TEMPORAL_CLI_NAMESPACE"},
	},
}
