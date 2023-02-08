package activity

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

func NewActivityCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "complete",
			Usage: common.CompleteActivityDefinition,
			UsageText: common.CompleteActivityUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    common.FlagWorkflowIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    common.FlagRunIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagActivityID,
					Usage:    common.FlagActivityIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagResult,
					Usage:    common.FlagResultDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagIdentity,
					Usage:    common.FlagIdentityDefinition,
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
			Usage: common.FailActivityDefinition,
			UsageText: common.FailActivityUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagWorkflowID,
					Aliases:  common.FlagWorkflowIDAlias,
					Usage:    common.FlagWorkflowIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagRunID,
					Aliases:  common.FlagRunIDAlias,
					Usage:    common.FlagRunIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagActivityID,
					Usage:    common.FlagActivityIDDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagDetail,
					Usage:    common.FlagDetailDefinition,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagIdentity,
					Usage:    common.FlagIdentityDefinition,
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
