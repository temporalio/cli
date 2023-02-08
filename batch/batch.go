package batch

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

func NewBatchCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "describe",
			Usage: common.DescribeBatchJobDefinition,
			UsageText: common.DescribeBatchUsageText,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagJobID,
					Usage:    common.FlagJobIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
			}, common.FlagsForFormatting...),
			Action: func(c *cli.Context) error {
				return DescribeBatchJob(c)
			},
		},
		{
			Name:      "list",
			Usage:     common.ListBatchJobsDefinition,
			UsageText: common.ListBatchUsageText,
			Flags:     common.FlagsForPaginationAndRendering,
			ArgsUsage: " ",
			Action: func(c *cli.Context) error {
				return ListBatchJobs(c)
			},
		},
		{
			Name:  "terminate",
			Usage: common.TerminateBatchJobDefinition,
			UsageText: common.TerminateBatchUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagJobID,
					Usage:    common.FlagJobIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return StopBatchJob(c)
			},
		},
	}
}
