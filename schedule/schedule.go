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
		Usage:    "Schedule Id",
		Required: true,
		Category: common.CategoryMain,
	}
	overlap := &cli.StringFlag{
		Name:     common.FlagOverlapPolicy,
		Usage:    "Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll",
		Category: common.CategoryMain,
	}

	scheduleSpecFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:     common.FlagCalendar,
			Usage:    `Calendar specification in JSON, e.g. {"dayOfWeek":"Fri","hour":"17","minute":"5"}`,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagCronSchedule,
			Usage:    `Calendar specification as cron string, e.g. "30 2 * * 5" or "@daily"`,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagInterval,
			Usage:    "Interval duration, e.g. 90m, or 90m/13m to include phase offset",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagStartTime,
			Usage:    "Overall schedule start time",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagEndTime,
			Usage:    "Overall schedule end time",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagJitter,
			Usage:    "Jitter duration",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagTimeZone,
			Usage:    "Time zone (IANA name)",
			Category: common.CategoryMain,
		},
	}

	scheduleStateFlags := []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagNotes,
			Usage:    "Initial value of notes field",
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagPause,
			Usage:    "Initial value of paused state",
			Category: common.CategoryMain,
		},
		&cli.IntFlag{
			Name:     common.FlagRemainingActions,
			Usage:    "Total number of actions allowed",
			Category: common.CategoryMain,
		},
	}

	schedulePolicyFlags := []cli.Flag{
		overlap,
		&cli.StringFlag{
			Name:     common.FlagCatchupWindow,
			Usage:    "Maximum allowed catch-up time if server is down",
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagPauseOnFailure,
			Usage:    "Pause schedule after any workflow failure",
			Category: common.CategoryMain,
		},
	}

	// These are the same flags as for start workflow, but we need to change the Usage to talk about schedules instead of workflows.
	scheduleVisibilityFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:     common.FlagSearchAttribute,
			Usage:    "Set Search Attribute on a schedule. Format: key=value. Use valid JSON formats for value",
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagMemo,
			Usage:    "Set a memo on a schedule. Format: key=value. Use valid JSON formats for value",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagMemoFile,
			Usage:    "Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value",
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
			Usage:       "Create a new schedule",
			Description: "Takes a schedule specification plus all the same args as starting a workflow",
			Flags:       createFlags,
			Action:      CreateSchedule,
			Category:    common.CategoryMain,
		},
		{
			Name:        "update",
			Usage:       "Updates a schedule with a new definition (full replacement, not patch)",
			Description: "Takes a schedule specification plus all the same args as starting a workflow",
			Flags:       createFlags,
			Action:      UpdateSchedule,
			Category:    common.CategoryMain,
		},
		{
			Name:  "toggle",
			Usage: "Pauses or unpauses a schedule",
			Flags: []cli.Flag{
				sid,
				&cli.BoolFlag{
					Name:     common.FlagPause,
					Usage:    "Pauses the schedule",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagUnpause,
					Usage:    "Unpauses the schedule",
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Free-form text to describe reason for pause/unpause",
					Value:    "(no reason provided)",
					Category: common.CategoryMain,
				},
			},
			Action: ToggleSchedule,
		},
		{
			Name:  "trigger",
			Usage: "Triggers an immediate action",
			Flags: []cli.Flag{
				sid,
				overlap,
			},
			Action: TriggerSchedule,
		},
		{
			Name:  "backfill",
			Usage: "Backfills a past time range of actions",
			Flags: []cli.Flag{
				sid,
				overlap,
				&cli.StringFlag{
					Name:     common.FlagStartTime,
					Usage:    "Backfill start time",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagEndTime,
					Usage:    "Backfill end time",
					Required: true,
					Category: common.CategoryMain,
				},
			},
			Action: BackfillSchedule,
		},
		{
			Name:  "describe",
			Usage: "Get schedule configuration and current state",
			Flags: append([]cli.Flag{
				sid,
				&cli.BoolFlag{
					Name:     common.FlagPrintRaw,
					Usage:    "Print raw data as json (prefer this over -o json for scripting)",
					Category: common.CategoryMain,
				},
			}, common.FlagsForFormatting...),
			Action: DescribeSchedule,
		},
		{
			Name:  "delete",
			Usage: "Deletes a schedule",
			Flags: []cli.Flag{
				sid,
			},
			Action: DeleteSchedule,
		},
		{
			Name:   "list",
			Usage:  "Lists schedules",
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
