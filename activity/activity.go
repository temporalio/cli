package activity

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

func NewActivityCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "complete",
			Usage: "Completes an Activity.",
			UsageText: "When used, the Activity is scheduled to be completed.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    "Identifies the Workflow that the Activity is running on.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    "Identifies the current Workflow Run.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagActivityID,
					Usage:    "Identifies the Activity to be completed.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagResult,
					Usage:    "Set the result value of Activity completion.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagIdentity,
					Usage:    "Specify operator's identity.",
					Required: true,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return CompleteActivity(c)
			},
		},
		{
			Name:  "fail",
			Usage: "Fail an Activity.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    "Identifies the Workflow that the Activity is running on.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    "Identifies the current Workflow Run.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagActivityID,
					Usage:    "Identifies the Activity to fail.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason to fail the Activity.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagDetail,
					Usage:    "Detail to fail the Activity.",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagIdentity,
					Usage:    "Specify the operator's identity.",
					Required: true,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return FailActivity(c)
			},
		},
	}
}
