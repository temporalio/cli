package taskqueue

import (
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

func NewTaskQueueCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "describe",
			Usage: common.DescribeTaskQueueDefinition,
			UsageText:common.DescribeTaskQueueUsageText,
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
			}, common.FlagsForFormatting...),
			Action: func(c *cli.Context) error {
				return DescribeTaskQueue(c)
			},
		},
		{
			Name:  "list-partition",
			Usage: common.ListPartitionTaskQueueDefinition,
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
	}
}
