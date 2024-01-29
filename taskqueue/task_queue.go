package taskqueue

import (
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

func NewTaskQueueCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "describe",
			Usage:     common.DescribeTaskQueueDefinition,
			UsageText: common.DescribeTaskQueueUsageText,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagTaskQueue,
					Aliases:  common.FlagTaskQueueAlias,
					Usage:    common.FlagTaskQueueName,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagTaskQueueType,
					Value:    "workflow",
					Usage:    common.FlagTaskQueueTypeDefinition,
					Category: common.CategoryMain,
				},
				// TOOD: remove this when the server does partition fan-out
				&cli.IntFlag{
					Name:     common.FlagPartitions,
					Value:    1,
					Usage:    common.FlagPartitionsDefinition,
					Category: common.CategoryMain,
				},
			}, common.FlagsForFormatting...),
			Action: func(c *cli.Context) error {
				return DescribeTaskQueue(c)
			},
		},
		{
			Name:      "list-partition",
			Usage:     common.ListPartitionTaskQueueDefinition,
			UsageText: common.TaskQueueListPartitionUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagTaskQueue,
					Aliases:  common.FlagTaskQueueAlias,
					Usage:    common.FlagTaskQueueName,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     output.FlagOutput,
					Aliases:  common.FlagOutputAlias,
					Usage:    output.UsageText,
					Value:    string(output.Table),
					Category: common.CategoryDisplay,
				},
			},
			Action: func(c *cli.Context) error {
				return ListTaskQueuePartitions(c)
			},
		},
		{
			Name:      "update-build-ids",
			Usage:     common.UpdateBuildIDsDefinition,
			UsageText: common.UpdateBuildIDsDefinitionText,
			Subcommands: []*cli.Command{
				{
					Name:      "add-new-default",
					Usage:     common.AddNewDefaultBuildIDDefinition,
					UsageText: common.AddNewDefaultBuildIDDefinitionUsage,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     common.FlagTaskQueue,
							Aliases:  common.FlagTaskQueueAlias,
							Usage:    common.FlagTaskQueueName,
							Required: true,
							Category: common.CategoryMain,
						},
						&cli.StringFlag{
							Name:     common.FlagBuildID,
							Usage:    common.FlagNewBuildIDUsage,
							Required: true,
							Category: common.CategoryMain,
						},
					},
					Action: func(c *cli.Context) error {
						return BuildIDAddNewDefault(c)
					},
				},
				{
					Name:      "add-new-compatible",
					Usage:     common.AddNewCompatibleBuildIDDefinition,
					UsageText: common.AddNewCompatibleBuildIDDefinitionUsage,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     common.FlagTaskQueue,
							Aliases:  common.FlagTaskQueueAlias,
							Usage:    common.FlagTaskQueueName,
							Required: true,
							Category: common.CategoryMain,
						},
						&cli.StringFlag{
							Name:     common.FlagBuildID,
							Usage:    common.FlagNewBuildIDUsage,
							Required: true,
							Category: common.CategoryMain,
						},
						&cli.StringFlag{
							Name:     common.FlagExistingCompatibleBuildID,
							Usage:    common.FlagExistingCompatibleBuildIDUsage,
							Required: true,
							Category: common.CategoryMain,
						},
						&cli.BoolFlag{
							Name:     common.FlagSetBuildIDAsDefault,
							Usage:    common.FlagSetBuildIDAsDefaultUsage,
							Category: common.CategoryMain,
						},
					},
					Action: func(c *cli.Context) error {
						return BuildIDAddNewCompatible(c)
					},
				},
				{
					Name:      "promote-set",
					Usage:     common.PromoteSetDefinition,
					UsageText: common.PromoteSetDefinitionUsage,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     common.FlagTaskQueue,
							Aliases:  common.FlagTaskQueueAlias,
							Usage:    common.FlagTaskQueueName,
							Required: true,
							Category: common.CategoryMain,
						},
						&cli.StringFlag{
							Name:     common.FlagBuildID,
							Usage:    common.FlagPromoteSetBuildIDUsage,
							Required: true,
							Category: common.CategoryMain,
						},
					},
					Action: func(c *cli.Context) error {
						return BuildIDPromoteSet(c)
					},
				},
				{
					Name:      "promote-id-in-set",
					Usage:     common.PromoteIDInSetDefinition,
					UsageText: common.PromoteIDInSetDefinitionUsage,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     common.FlagTaskQueue,
							Aliases:  common.FlagTaskQueueAlias,
							Usage:    common.FlagTaskQueueName,
							Required: true,
							Category: common.CategoryMain,
						},
						&cli.StringFlag{
							Name:     common.FlagBuildID,
							Usage:    common.FlagPromoteBuildIDUsage,
							Required: true,
							Category: common.CategoryMain,
						},
					},
					Action: func(c *cli.Context) error {
						return BuildIDPromoteInSet(c)
					},
				},
			},
		},
		{
			Name:      "get-build-ids",
			Usage:     common.GetBuildIDsDefinition,
			UsageText: common.GetBuildIDsDefinitionText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagTaskQueue,
					Aliases:  common.FlagTaskQueueAlias,
					Usage:    common.FlagTaskQueueName,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.IntFlag{
					Name:     common.FlagMaxBuildIDSets,
					Usage:    common.FlagMaxBuildIDSetsUsage,
					Category: common.CategoryMain,
					Value:    0,
				},
			},
			Action: func(c *cli.Context) error {
				return GetBuildIDs(c)
			},
		},
		{
			Name:      "get-build-id-reachability",
			Usage:     common.GetBuildIDReachabilityDefinition,
			UsageText: common.GetBuildIDReachabilityDefinitionText,
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:     common.FlagBuildID,
					Usage:    common.FlagBuildIDReachabilityUsage,
					Category: common.CategoryMain,
				},
				&cli.StringSliceFlag{
					Name:     common.FlagTaskQueue,
					Aliases:  common.FlagTaskQueueAlias,
					Usage:    common.FlagTaskQueueForReachabilityUsage,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReachabilityType,
					Usage:    common.FlagReachabilityTypeUsage,
					Category: common.CategoryMain,
					Value:    "existing",
				},
			},
			Action: func(c *cli.Context) error {
				return GetBuildIDReachability(c)
			},
		},
	}
}
