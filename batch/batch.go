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
			UsageText: "This command shows the progress of an ongoing Batch job.",
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
			UsageText: "When used, all Batch operation jobs within the system are listed.",
			Flags:     common.FlagsForPaginationAndRendering,
			ArgsUsage: " ",
			Action: func(c *cli.Context) error {
				return ListBatchJobs(c)
			},
		},
		{
			Name:  "terminate",
			Usage: "Stop a Batch operation job.",
			UsageText: "When used, the Batch job with the provided Batch Id is terminated.",
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
