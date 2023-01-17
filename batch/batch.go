package batch

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

func NewBatchCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "describe",
			Usage: "Describe a Batch operation job.",
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagJobID,
					Usage:    "Batch Job Id",
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
			Usage:     "List Batch operation jobs.",
			Flags:     common.FlagsForPaginationAndRendering,
			ArgsUsage: " ",
			Action: func(c *cli.Context) error {
				return ListBatchJobs(c)
			},
		},
		{
			Name:  "terminate",
			Usage: "Stop a Batch operation job.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagJobID,
					Usage:    "Batch Job Id",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason to stop the Batch job.",
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
