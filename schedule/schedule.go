package schedule

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func NewScheduleCommands() []*cli.Command {
	sid := &cli.StringFlag{
		Name:     common.FlagScheduleID,
		Aliases:  common.FlagScheduleIDAlias,
		Usage:    common.FlagScheduleIDDefinition,
		Required: true,
		Category: common.CategoryMain,
	}
	overlap := &cli.StringFlag{
		Name:     common.FlagOverlapPolicy,
		Usage:    common.FlagOverlapPolicyDefinition,
		Category: common.CategoryMain,
	}

	scheduleSpecFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:     common.FlagCalendar,
			Usage:    common.FlagCalenderDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagCronSchedule,
			Usage:    common.FlagCronScheduleDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagInterval,
			Usage:    common.FlagIntervalDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagStartTime,
			Usage:    common.FlagStartTimeDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagEndTime,
			Usage:    common.FlagEndTimeDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagJitter,
			Usage:    common.FlagJitterDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagTimeZone,
			Usage:    common.FlagTimeZoneDefinition,
			Category: common.CategoryMain,
		},
	}

	scheduleStateFlags := []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagNotes,
			Usage:    common.FlagNotesDefinition,
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagPause,
			Usage:    common.FlagPauseDefinition,
			Category: common.CategoryMain,
		},
		&cli.IntFlag{
			Name:     common.FlagRemainingActions,
			Usage:    common.FlagRemainingActionsDefinition,
			Category: common.CategoryMain,
		},
	}

	schedulePolicyFlags := []cli.Flag{
		overlap,
		&cli.StringFlag{
			Name:     common.FlagCatchupWindow,
			Usage:    common.FlagCatchupWindowDefinition,
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagPauseOnFailure,
			Usage:    common.FlagPauseOnFailureDefinition,
			Category: common.CategoryMain,
		},
	}

	// These are the same flags as for start workflow, but we need to change the Usage to talk about schedules instead of workflows.
	scheduleVisibilityFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:     common.FlagSearchAttribute,
			Usage:    common.FlagSearchAttributeScheduleDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagMemo,
			Usage:    common.FlagMemoScheduleDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagMemoFile,
			Usage:    common.FlagMemoFileScheduleDefinition,
			Category: common.CategoryMain,
		},
	}

	createFlags := []cli.Flag{sid}
	createFlags = append(createFlags, scheduleSpecFlags...)
	createFlags = append(createFlags, scheduleStateFlags...)
	createFlags = append(createFlags, schedulePolicyFlags...)
	createFlags = append(createFlags, scheduleVisibilityFlags...)
	createFlags = append(createFlags, removeFlags(common.FlagsForStartWorkflowLong,
		common.FlagCronSchedule, common.FlagWorkflowIDReusePolicy,
		common.FlagMemo, common.FlagMemoFile,
		common.FlagSearchAttribute,
	)...)

	return []*cli.Command{
		{
			Name:        "create",
			Usage:       common.ScheduleCreateDefinition,
			UsageText: common.ScheduleCreateUsageText,
			Description: common.ScheduleCreateDescription,
			Flags:       createFlags,
			Action:      CreateSchedule,
			Category:    common.CategoryMain,
		},
		{
			Name:        "update",
			Usage:       common.ScheduleUpdateDefinition,
			UsageText: common.ScheduleUpdateUsageText,
			Description: common.ScheduleUpdateDescription,
			Flags:       createFlags,
			Action:      UpdateSchedule,
			Category:    common.CategoryMain,
		},
		{
			Name:  "toggle",
			Usage: common.ScheduleToggleDefinition,
			UsageText: common.ScheduleToggleUsageText,
			Flags: []cli.Flag{
				sid,
				&cli.BoolFlag{
					Name:     common.FlagPause,
					Usage:    common.FlagPauseScheduleDefinition,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagUnpause,
					Usage:    common.FlagUnpauseDefinition,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    common.FlagReasonDefinition,
					Value:    "(no reason provided)",
					Category: common.CategoryMain,
				},
			},
			Action: ToggleSchedule,
		},
		{
			Name:  "trigger",
			Usage: common.ScheduleTriggerDefinition,
			UsageText: common.ScheduleTriggerUsageText,
			Flags: []cli.Flag{
				sid,
				overlap,
			},
			Action: TriggerSchedule,
		},
		{
			Name:  "backfill",
			Usage: common.ScheduleBackfillDefinition,
			UsageText: common.ScheduleBackfillUsageText,
			Flags: []cli.Flag{
				sid,
				overlap,
				&cli.StringFlag{
					Name:     common.FlagStartTime,
					Usage:    common.FlagBackfillStartTime,
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagEndTime,
					Usage:    common.FlagBackfillEndTime,
					Required: true,
					Category: common.CategoryMain,
				},
			},
			Action: BackfillSchedule,
		},
		{
			Name:  "describe",
			Usage: common.ScheduleDescribeDefinition,
			UsageText: common.ScheduleDescribeUsageText,
			Flags: append([]cli.Flag{
				sid,
				&cli.BoolFlag{
					Name:     common.FlagPrintRaw,
					Usage:    common.FlagPrintRawDefinition,
					Category: common.CategoryMain,
				},
			}, common.FlagsForFormatting...),
			Action: DescribeSchedule,
		},
		{
			Name:  "delete",
			Usage: common.ScheduleDeleteDefinition,
			UsageText: common.ScheduleDeleteUsageText,
			Flags: []cli.Flag{
				sid,
			},
			Action: DeleteSchedule,
		},
		{
			Name:   "list",
			Usage:  common.ScheduleListDefinition,
			UsageText: common.ScheduleListUsageText,
			Flags:  common.FlagsForPaginationAndRendering,
			Action: ListSchedules,
		},
	}
}

func removeFlags(flags []cli.Flag, remove ...string) []cli.Flag {
	out := make([]cli.Flag, 0, len(flags))
	for _, f := range flags {
		// Names[0] is always the primary name
		if !slices.Contains(remove, f.Names()[0]) {
			out = append(out, f)
		}
	}
	return out
}
