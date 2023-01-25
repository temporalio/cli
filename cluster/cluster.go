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
			Usage: common.HealthDefinition,
			UsageText: common.HealthUsageText,
			Action: func(c *cli.Context) error {
				return HealthCheck(c)
			},
		},
		{
			Name:      "describe",
			Usage:     common.DescribeDefinition,
			UsageText: common.ClusterDescribeUsageText,
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
					Usage:    "Customize fields to print. Set to 'long' to automatically print more of main fields.",
					Category: common.CategoryDisplay,
				},
			},
			Action: func(c *cli.Context) error {
				return DescribeCluster(c)
			},
		},
		{
			Name:      "system",
			Usage:     common.SystemDefinition,
			UsageText: common.ClusterSystemUsageText,
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
					Usage:    "Customize fields to print. Set to 'long' to automatically print more of main fields.",
					Category: common.CategoryDisplay,
				},
			},
			Action: func(c *cli.Context) error {
				return DescribeSystem(c)
			},
		},
		{
			Name:      "upsert",
			Usage:     common.UpsertDefinition,
			UsageText: common.ClusterUpsertUsageText,
			ArgsUsage: " ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagClusterAddress,
					Usage:    "Frontend address of the remote Cluster.",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagClusterEnableConnection,
					Usage:    "Enable cross-cluster connection.",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return UpsertCluster(c)
			},
		},
		{
			Name:      "list",
			Usage:     common.ListDefinition,
			UsageText: common.ClusterListUsageText,
			ArgsUsage: " ",
			Flags:     common.FlagsForPaginationAndRendering,
			Action: func(c *cli.Context) error {
				return ListClusters(c)
			},
		},
		{
			Name:      "remove",
			Usage:     common.RemoveDefinition,
			UsageText: common.ClusterRemoveUsageText,
			ArgsUsage: " ",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagName,
					Usage:    "Frontend address of the remote Cluster.",
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
