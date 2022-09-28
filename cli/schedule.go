// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"

	"github.com/temporalio/tctl-kit/pkg/flags"
)

func newScheduleCommands() []*cli.Command {
	sid := &cli.StringFlag{
		Name:     FlagScheduleID,
		Aliases:  FlagScheduleIDAlias,
		Usage:    "Schedule Id",
		Required: true,
	}
	overlap := &cli.StringFlag{
		Name:  FlagOverlapPolicy,
		Usage: "Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll",
	}

	scheduleSpecFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:  FlagCalendar,
			Usage: `Calendar specification in JSON, e.g. {"dayOfWeek":"Fri","hour":"17","minute":"5"}`,
		},
		&cli.StringSliceFlag{
			Name:  FlagCronSchedule,
			Usage: `Calendar specification as cron string, e.g. "30 2 * * 5" or "@daily"`,
		},
		&cli.StringSliceFlag{
			Name:  FlagInterval,
			Usage: "Interval duration, e.g. 90m, or 90m/13m to include phase offset",
		},
		&cli.StringFlag{
			Name:  FlagStartTime,
			Usage: "Overall schedule start time",
		},
		&cli.StringFlag{
			Name:  FlagEndTime,
			Usage: "Overall schedule end time",
		},
		&cli.StringFlag{
			Name:  FlagJitter,
			Usage: "Jitter duration",
		},
		&cli.StringFlag{
			Name:  FlagTimeZone,
			Usage: "Time zone (IANA name)",
		},
	}

	scheduleStateFlags := []cli.Flag{
		&cli.StringFlag{
			Name:  FlagNotes,
			Usage: "Initial value of notes field",
		},
		&cli.BoolFlag{
			Name:  FlagPause,
			Usage: "Initial value of paused state",
		},
		&cli.IntFlag{
			Name:  FlagRemainingActions,
			Usage: "Total number of actions allowed",
		},
	}

	schedulePolicyFlags := []cli.Flag{
		overlap,
		&cli.StringFlag{
			Name:  FlagCatchupWindow,
			Usage: "Maximum allowed catch-up time if server is down",
		},
		&cli.BoolFlag{
			Name:  FlagPauseOnFailure,
			Usage: "Pause schedule after any workflow failure",
		},
	}

	// These are the same flags as for start workflow, but we need to change the Usage to talk about schedules instead of workflows.
	scheduleVisibilityFlags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:  FlagSearchAttribute,
			Usage: "Set Search Attribute on a schedule. Format: key=value. Use valid JSON formats for value",
		},
		&cli.StringSliceFlag{
			Name:  FlagMemo,
			Usage: "Set a memo on a schedule. Format: key=value. Use valid JSON formats for value",
		},
		&cli.StringFlag{
			Name:  FlagMemoFile,
			Usage: "Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value",
		},
	}

	createFlags := []cli.Flag{sid}
	createFlags = append(createFlags, scheduleSpecFlags...)
	createFlags = append(createFlags, scheduleStateFlags...)
	createFlags = append(createFlags, schedulePolicyFlags...)
	createFlags = append(createFlags, scheduleVisibilityFlags...)
	createFlags = append(createFlags, removeFlags(flagsForStartWorkflowLong,
		FlagCronSchedule, FlagWorkflowIDReusePolicy,
		FlagMemo, FlagMemoFile,
		FlagSearchAttribute,
	)...)

	return []*cli.Command{
		{
			Name:        "create",
			Usage:       "Create a new schedule",
			Description: "Takes a schedule specification plus all the same args as starting a workflow",
			Flags:       createFlags,
			Action:      CreateSchedule,
		},
		{
			Name:        "update",
			Usage:       "Updates a schedule with a new definition (full replacement, not patch)",
			Description: "Takes a schedule specification plus all the same args as starting a workflow",
			Flags:       createFlags,
			Action:      UpdateSchedule,
		},
		{
			Name:  "toggle",
			Usage: "Pauses or unpauses a schedule",
			Flags: []cli.Flag{
				sid,
				&cli.BoolFlag{
					Name:  FlagPause,
					Usage: "Pauses the schedule",
				},
				&cli.BoolFlag{
					Name:  FlagUnpause,
					Usage: "Unpauses the schedule",
				},
				&cli.StringFlag{
					Name:  FlagReason,
					Usage: "Free-form text to describe reason for pause/unpause",
					Value: "(no reason provided)",
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
					Name:     FlagStartTime,
					Usage:    "Backfill start time",
					Required: true,
				},
				&cli.StringFlag{
					Name:     FlagEndTime,
					Usage:    "Backfill end time",
					Required: true,
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
					Name:  FlagPrintRaw,
					Usage: "Print raw data as json (prefer this over -o json for scripting)",
				},
			}, flags.FlagsForRendering...),
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
			Flags:  append([]cli.Flag{}, flags.FlagsForPaginationAndRendering...),
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
