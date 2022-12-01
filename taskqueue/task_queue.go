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
			Usage: "Describe the Workers that have recently polled on this Task Queue",
			Description: `The Server records the last time of each poll request. Poll requests can last up to a minute, so a LastAccessTime under a minute is normal. If it's over a minute, then likely either the Worker is at capacity (all Workflow and Activity slots are full) or it has shut down. Once it has been 5 minutes since the last poll request, the Worker is removed from the list.

RatePerSecond is the maximum Activities per second the Worker will execute.`,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagTaskQueue,
					Aliases:  common.FlagTaskQueueAlias,
					Usage:    "Task Queue name",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagTaskQueueType,
					Value:    "workflow",
					Usage:    "Task Queue type [workflow|activity]",
					Category: common.CategoryMain,
				},
			}, common.FlagsForFormatting...),
			Action: func(c *cli.Context) error {
				return DescribeTaskQueue(c)
			},
		},
		{
			Name:  "list-partition",
			Usage: "List the Task Queue's partitions and which matching node they are assigned to",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagTaskQueue,
					Aliases:  common.FlagTaskQueueAlias,
					Usage:    "Task Queue name",
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
