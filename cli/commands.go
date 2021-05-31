package cli

import (
	"github.com/urfave/cli/v2"
)

var tctlCommands = []*cli.Command{
	// {
	// 	Name:        "namespace",
	// 	Aliases:     []string{"n"},
	// 	Usage:       "Operate Temporal namespace",
	// 	Subcommands: newNamespaceCommands(),
	// },
	{
		Name:        "workflow",
		Aliases:     []string{"w"},
		Usage:       "Operate Temporal workflow",
		Subcommands: newWorkflowCommands(),
	},
	// {
	// 	Name:        "activity",
	// 	Aliases:     []string{"act"},
	// 	Usage:       "operate activities of workflow",
	// 	Subcommands: newActivityCommands(),
	// },
	// {
	// 	Name:        "taskqueue",
	// 	Aliases:     []string{"tq"},
	// 	Usage:       "Operate Temporal task queue",
	// 	Subcommands: newTaskQueueCommands(),
	// },
	// {
	// 	Name:        "batch",
	// 	Usage:       "batch operation on a list of workflows from query.",
	// 	Subcommands: newBatchCommands(),
	// },
	// {
	// 	Name:    "admin",
	// 	Aliases: []string{"adm"},
	// 	Usage:   "Run admin operation",
	// 	Subcommands: []cli.Command{
	// 		{
	// 			Name:        "workflow",
	// 			Aliases:     []string{"wf"},
	// 			Usage:       "Run admin operation on workflow",
	// 			Subcommands: newAdminWorkflowCommands(),
	// 		},
	// 		{
	// 			Name:        "shard",
	// 			Aliases:     []string{"shar"},
	// 			Usage:       "Run admin operation on specific shard",
	// 			Subcommands: newAdminShardManagementCommands(),
	// 		},
	// 		{
	// 			Name:        "history_host",
	// 			Aliases:     []string{"hist"},
	// 			Usage:       "Run admin operation on history host",
	// 			Subcommands: newAdminHistoryHostCommands(),
	// 		},
	// 		{
	// 			Name:        "namespace",
	// 			Aliases:     []string{"d"},
	// 			Usage:       "Run admin operation on namespace",
	// 			Subcommands: newAdminNamespaceCommands(),
	// 		},
	// 		{
	// 			Name:        "elasticsearch",
	// 			Aliases:     []string{"es"},
	// 			Usage:       "Run admin operation on ElasticSearch",
	// 			Subcommands: newAdminElasticSearchCommands(),
	// 		},
	// 		{
	// 			Name:        "taskqueue",
	// 			Aliases:     []string{"tq"},
	// 			Usage:       "Run admin operation on taskQueue",
	// 			Subcommands: newAdminTaskQueueCommands(),
	// 		},
	// 		{
	// 			Name:        "membership",
	// 			Usage:       "Run admin operation on membership",
	// 			Subcommands: newAdminMembershipCommands(),
	// 		},
	// 		{
	// 			Name:        "cluster",
	// 			Aliases:     []string{"cl"},
	// 			Usage:       "Run admin operation on cluster",
	// 			Subcommands: newAdminClusterCommands(),
	// 		},
	// 		{
	// 			Name:        "dlq",
	// 			Aliases:     []string{"dlq"},
	// 			Usage:       "Run admin operation on DLQ",
	// 			Subcommands: newAdminDLQCommands(),
	// 		},
	// 		{
	// 			Name:        "db",
	// 			Aliases:     []string{"db"},
	// 			Usage:       "Run admin operations on database",
	// 			Subcommands: newDBCommands(),
	// 		},
	// 		{
	// 			Name:        "decode",
	// 			Usage:       "Decode payload",
	// 			Subcommands: newDecodeCommands(),
	// 		},
	// 	},
	// },
	// {
	// 	Name:        "cluster",
	// 	Aliases:     []string{"cl"},
	// 	Usage:       "Operate Temporal cluster",
	// 	Subcommands: newClusterCommands(),
	// },
	{
		Name:        "dataconverter",
		Aliases:     []string{"dc"},
		Usage:       "Operate Custom Data Converter",
		Subcommands: newDataConverterCommands(),
	},
}
