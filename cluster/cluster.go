package cluster

import (
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

func NewClusterCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "health",
			Usage: "Check health of frontend service",
			Action: func(c *cli.Context) error {
				return HealthCheck(c)
			},
		},
		{
			Name:      "describe",
			Usage:     "Show information about the cluster",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     output.FlagOutput,
					Aliases:  common.FlagOutputAlias,
					Usage:    output.UsageText,
					Value:    string(output.Table),
					Category: common.CategoryDisplay,
				},
				&cli.StringFlag{
					Name:     output.FlagFields,
					Usage:    "customize fields to print. Set to 'long' to automatically print more of main fields",
					Category: common.CategoryDisplay,
				},
			},
			Action: func(c *cli.Context) error {
				return DescribeCluster(c)
			},
		},
		{
			Name:      "system",
			Usage:     "Show information about the system and capabilities",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     output.FlagOutput,
					Aliases:  common.FlagOutputAlias,
					Usage:    output.UsageText,
					Value:    string(output.Table),
					Category: common.CategoryDisplay,
				},
				&cli.StringFlag{
					Name:     output.FlagFields,
					Usage:    "customize fields to print. Set to 'long' to automatically print more of main fields",
					Category: common.CategoryDisplay,
				},
			},
			Action: func(c *cli.Context) error {
				return DescribeSystem(c)
			},
		},
		{
			Name:      "upsert",
			Usage:     "Add or update a remote cluster",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagClusterAddress,
					Usage:    "Frontend address of the remote cluster",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagClusterEnableConnection,
					Usage:    "Enable cross cluster connection",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return UpsertCluster(c)
			},
		},
		{
			Name:      "list",
			Usage:     "List all remote clusters",
			ArgsUsage: " ",
			Flags:     common.FlagsForPaginationAndRendering,
			Action: func(c *cli.Context) error {
				return ListClusters(c)
			},
		},
		{
			Name:      "remove",
			Usage:     "Remove a remote cluster",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagName,
					Usage:    "Frontend address of the remote cluster",
					Required: true,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return RemoveCluster(c)
			},
		},
	}
}
