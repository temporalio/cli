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

package cli_curr

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/pborman/uuid"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli"
	"go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	schedpb "go.temporal.io/api/schedule/v1"
	"go.temporal.io/api/taskqueue/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/collection"
	"go.temporal.io/server/common/primitives/timestamp"
)

func scheduleBaseArgs(c *cli.Context) (
	frontendClient workflowservice.WorkflowServiceClient,
	namespace string,
	scheduleID string,
	err error,
) {
	frontendClient = cFactory.FrontendClient(c)
	namespace = getRequiredGlobalOption(c, FlagNamespace)
	scheduleID = c.String(FlagScheduleID)
	return frontendClient, namespace, scheduleID, nil
}

func buildCalendarSpec(s string) (*schedpb.CalendarSpec, error) {
	var cal schedpb.CalendarSpec
	err := jsonpb.UnmarshalString(s, &cal)
	if err != nil {
		return nil, err
	}
	return &cal, nil
}

func buildIntervalSpec(s string) (*schedpb.IntervalSpec, error) {
	var interval, phase time.Duration
	var err error
	parts := strings.Split(s, "/")
	if len(parts) > 2 {
		return nil, errors.New("invalid interval string")
	} else if len(parts) == 2 {
		if phase, err = timestamp.ParseDuration(parts[1]); err != nil {
			return nil, err
		}
	}
	if interval, err = timestamp.ParseDuration(parts[0]); err != nil {
		return nil, err
	}
	return &schedpb.IntervalSpec{Interval: &interval, Phase: &phase}, nil
}

func buildScheduleSpec(c *cli.Context) (*schedpb.ScheduleSpec, error) {
	now := time.Now()

	var out schedpb.ScheduleSpec
	for _, s := range c.StringSlice(FlagCalendar) {
		cal, err := buildCalendarSpec(s)
		if err != nil {
			return nil, err
		}
		out.Calendar = append(out.Calendar, cal)
	}
	out.CronString = c.StringSlice(FlagCronSchedule)
	for _, s := range c.StringSlice(FlagInterval) {
		cal, err := buildIntervalSpec(s)
		if err != nil {
			return nil, err
		}
		out.Interval = append(out.Interval, cal)
	}
	if c.IsSet(FlagStartTime) {
		t := parseTime(c.String(FlagStartTime), time.Time{}, now)
		out.StartTime = timestamp.TimePtr(t)
	}
	if c.IsSet(FlagEndTime) {
		t := parseTime(c.String(FlagEndTime), time.Time{}, now)
		out.EndTime = timestamp.TimePtr(t)
	}
	if c.IsSet(FlagJitter) {
		d, err := timestamp.ParseDuration(c.String(FlagJitter))
		if err != nil {
			return nil, err
		}
		out.Jitter = timestamp.DurationPtr(d)
	}
	if c.IsSet(FlagTimeZone) {
		tzName := c.String(FlagTimeZone)
		if _, err := time.LoadLocation(tzName); err != nil {
			return nil, fmt.Errorf("unknown time zone name %q", tzName)
		}
		out.TimezoneName = tzName
	}
	return &out, nil
}

func buildScheduleAction(c *cli.Context) (*schedpb.ScheduleAction, error) {
	taskQueue, workflowType, et, rt, dt, wid := startWorkflowBaseArgs(c)
	inputs := processJSONInput(c)

	// TODO: allow specifying: memo, search attributes, workflow retry policy

	newWorkflow := &workflowpb.NewWorkflowExecutionInfo{
		WorkflowId:               wid,
		WorkflowType:             &common.WorkflowType{Name: workflowType},
		TaskQueue:                &taskqueue.TaskQueue{Name: taskQueue},
		Input:                    inputs,
		WorkflowExecutionTimeout: timestamp.DurationPtr(time.Second * time.Duration(et)),
		WorkflowRunTimeout:       timestamp.DurationPtr(time.Second * time.Duration(rt)),
		WorkflowTaskTimeout:      timestamp.DurationPtr(time.Second * time.Duration(dt)),
	}

	return &schedpb.ScheduleAction{
		Action: &schedpb.ScheduleAction_StartWorkflow{
			StartWorkflow: newWorkflow,
		},
	}, nil
}

func buildScheduleState(c *cli.Context) (*schedpb.ScheduleState, error) {
	var out schedpb.ScheduleState
	out.Notes = c.String(FlagNotes)
	out.Paused = c.Bool(FlagPause)
	if c.IsSet(FlagRemainingActions) {
		out.LimitedActions = true
		out.RemainingActions = int64(c.Int(FlagRemainingActions))
	}
	return &out, nil
}

func getOverlapPolicy(c *cli.Context) (enumspb.ScheduleOverlapPolicy, error) {
	i, err := stringToEnum(c.String(FlagOverlapPolicy), enumspb.ScheduleOverlapPolicy_value)
	if err != nil {
		return 0, err
	}
	return enumspb.ScheduleOverlapPolicy(i), nil
}

func buildSchedulePolicies(c *cli.Context) (*schedpb.SchedulePolicies, error) {
	var out schedpb.SchedulePolicies
	var err error
	out.OverlapPolicy, err = getOverlapPolicy(c)
	if err != nil {
		return nil, err
	}
	if c.IsSet(FlagCatchupWindow) {
		d, err := timestamp.ParseDuration(c.String(FlagCatchupWindow))
		if err != nil {
			return nil, err
		}
		out.CatchupWindow = timestamp.DurationPtr(d)
	}
	out.PauseOnFailure = c.Bool(FlagPauseOnFailure)
	return &out, nil
}

func buildSchedule(c *cli.Context) (*schedpb.Schedule, error) {
	sched := &schedpb.Schedule{}
	var err error
	if sched.Spec, err = buildScheduleSpec(c); err != nil {
		return nil, err
	}
	if sched.Action, err = buildScheduleAction(c); err != nil {
		return nil, err
	}
	if sched.Policies, err = buildSchedulePolicies(c); err != nil {
		return nil, err
	}
	if sched.State, err = buildScheduleState(c); err != nil {
		return nil, err
	}
	return sched, nil
}

func getMemoAndSearchAttributesForSchedule(c *cli.Context) (*common.Memo, *common.SearchAttributes) {
	memoMap := unmarshalMemoFromCLI(c)
	memo := encodeMemo(memoMap)

	saMap := unmarshalSearchAttrFromCLI(c)
	sa := encodeSearchAttributes(saMap)

	return memo, sa
}

func CreateSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	sched, err := buildSchedule(c)
	if err != nil {
		return err
	}
	memo, sa := getMemoAndSearchAttributesForSchedule(c)
	if err != nil {
		return err
	}

	req := &workflowservice.CreateScheduleRequest{
		Namespace:        namespace,
		ScheduleId:       scheduleID,
		Schedule:         sched,
		Identity:         getCliIdentity(),
		RequestId:        uuid.New(),
		Memo:             memo,
		SearchAttributes: sa,
	}

	_, err = frontendClient.CreateSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create schedule: %w", err)
	}

	fmt.Println(color.GreenString("Schedule created"))
	return nil
}

func UpdateSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	sched, err := buildSchedule(c)
	if err != nil {
		return err
	}

	req := &workflowservice.UpdateScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Schedule:   sched,
		Identity:   getCliIdentity(),
		RequestId:  uuid.New(),
	}

	_, err = frontendClient.UpdateSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to update schedule: %w", err)
	}

	fmt.Println(color.GreenString("Schedule updated"))
	return nil
}

func ToggleSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	pause, unpause := c.Bool(FlagPause), c.Bool(FlagUnpause)
	if pause && unpause {
		return errors.New("Cannot specify both --pause and --unpause")
	} else if !pause && !unpause {
		return errors.New("Must specify one of --pause and --unpause")
	}
	patch := &schedpb.SchedulePatch{}
	if pause {
		patch.Pause = c.String(FlagReason)
	} else if unpause {
		patch.Unpause = c.String(FlagReason)
	}

	req := &workflowservice.PatchScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Patch:      patch,
		Identity:   getCliIdentity(),
		RequestId:  uuid.New(),
	}
	_, err = frontendClient.PatchSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to toggle schedule: %w", err)
	}

	fmt.Println(color.GreenString("Schedule updated"))
	return nil
}

func TriggerSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	overlap, err := getOverlapPolicy(c)
	if err != nil {
		return err
	}

	req := &workflowservice.PatchScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Patch: &schedpb.SchedulePatch{
			TriggerImmediately: &schedpb.TriggerImmediatelyRequest{
				OverlapPolicy: overlap,
			},
		},
		Identity:  getCliIdentity(),
		RequestId: uuid.New(),
	}
	_, err = frontendClient.PatchSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to trigger schedule: %w", err)
	}

	fmt.Println(color.GreenString("Trigger request sent"))
	return nil
}

func BackfillSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	now := time.Now()
	startTime := parseTime(c.String(FlagStartTime), time.Time{}, now)
	endTime := parseTime(c.String(FlagEndTime), time.Time{}, now)

	overlap, err := getOverlapPolicy(c)
	if err != nil {
		return err
	}

	req := &workflowservice.PatchScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Patch: &schedpb.SchedulePatch{
			BackfillRequest: []*schedpb.BackfillRequest{
				{
					StartTime:     timestamp.TimePtr(startTime),
					EndTime:       timestamp.TimePtr(endTime),
					OverlapPolicy: overlap,
				},
			},
		},
		Identity:  getCliIdentity(),
		RequestId: uuid.New(),
	}
	_, err = frontendClient.PatchSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to backfill schedule: %w", err)
	}

	fmt.Println(color.GreenString("Backfill request sent"))
	return nil
}

func DescribeSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	req := &workflowservice.DescribeScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
	}
	resp, err := frontendClient.DescribeSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to describe schedule: %w", err)
	}

	if c.Bool(FlagPrintRaw) {
		prettyPrintJSONObject(resp)
		return nil
	}

	// output.PrintItems gets confused by nested fields of nil values, because it uses
	// reflection. ensure the first level is non-nil to avoid runtime errors.
	ensureNonNil(&resp.Schedule)
	ensureNonNil(&resp.Schedule.Spec)
	ensureNonNil(&resp.Schedule.Action)
	ensureNonNil(&resp.Schedule.Policies)
	ensureNonNil(&resp.Schedule.State)
	ensureNonNil(&resp.Info)

	// reform resp into more convenient shape
	var item struct {
		ScheduleId string

		Specification *schedpb.ScheduleSpec

		StartWorkflow *workflowpb.NewWorkflowExecutionInfo
		WorkflowType  string   // copy just string to reduce noise
		Input         []string // copy so we can decode it

		Policies *schedpb.SchedulePolicies
		State    *schedpb.ScheduleState
		Info     *schedpb.ScheduleInfo

		// more convenient copies of values from Info
		NextRunTime       *time.Time
		LastRunTime       *time.Time
		LastRunExecution  *common.WorkflowExecution
		LastRunActualTime *time.Time

		Memo             map[string]string // json only
		SearchAttributes map[string]string // json only
	}

	s, i := resp.Schedule, resp.Info
	item.ScheduleId = scheduleID
	item.Specification = s.Spec
	uncanonicalizeSpec(item.Specification)
	if sw := s.Action.GetStartWorkflow(); sw != nil {
		item.StartWorkflow = sw
		item.WorkflowType = sw.WorkflowType.GetName()
		item.Input = customDataConverter().ToStrings(sw.Input)
	}
	item.Policies = s.Policies
	if item.Policies.OverlapPolicy == enumspb.SCHEDULE_OVERLAP_POLICY_UNSPECIFIED {
		item.Policies.OverlapPolicy = enumspb.SCHEDULE_OVERLAP_POLICY_SKIP
	}
	item.State = s.State
	item.Info = i
	if fas := i.FutureActionTimes; len(fas) > 0 {
		item.NextRunTime = fas[0]
	}
	if ras := i.RecentActions; len(ras) > 0 {
		ra := ras[len(ras)-1]
		item.LastRunTime = ra.ScheduleTime
		item.LastRunActualTime = ra.ActualTime
		item.LastRunExecution = ra.StartWorkflowResult
	}
	if fields := resp.Memo.GetFields(); len(fields) > 0 {
		item.Memo = make(map[string]string, len(fields))
		for k, payload := range fields {
			item.Memo[k] = customDataConverter().ToString(payload)
		}
	}
	if fields := resp.SearchAttributes.GetIndexedFields(); len(fields) > 0 {
		item.SearchAttributes = make(map[string]string, len(fields))
		for k, payload := range fields {
			item.SearchAttributes[k] = defaultDataConverter().ToString(payload)
		}
	}

	opts := &output.PrintOptions{
		Fields: []string{
			"ScheduleId",
			"WorkflowType",
			"State.Paused",
			"State.Notes",
			"Info.RunningWorkflows",
			"NextRunTime",
			"LastRunTime",
			"Specification",
		},
		FieldsLong: []string{
			"StartWorkflow.WorkflowId",
			"StartWorkflow.TaskQueue",
			"Input",
			"Policies.OverlapPolicy",
			"Policies.PauseOnFailure",
			"Info.ActionCount",
			"Info.MissedCatchupWindow",
			"Info.OverlapSkipped",
			"LastRunExecution",
			"LastRunActualTime",
			"Info.CreateTime",
			"Info.UpdateTime",
			"Info.InvalidScheduleError",
		},
	}
	return output.PrintItems(nil, []interface{}{item}, opts)
}

func DeleteSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	req := &workflowservice.DeleteScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Identity:   getCliIdentity(),
	}
	_, err = frontendClient.DeleteSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to delete schedule: %w", err)
	}

	fmt.Println(color.GreenString("Schedule deleted"))
	return nil
}

func ListSchedules(c *cli.Context) error {
	frontendClient := cFactory.FrontendClient(c)
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	ctx, cancel := newContext(c)
	defer cancel()

	missingExtendedInfo := false

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		req := &workflowservice.ListSchedulesRequest{
			Namespace:     namespace,
			NextPageToken: npt,
		}
		resp, err := frontendClient.ListSchedules(ctx, req)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to list schedules: %w", err)
		}
		items := make([]interface{}, len(resp.Schedules))
		for i, sch := range resp.Schedules {
			var item struct {
				ScheduleId    string
				Specification *schedpb.ScheduleSpec
				StartWorkflow struct {
					WorkflowType string
				}
				State struct {
					Paused bool
					Notes  string
				}
				Info struct {
					NextRunTime       *time.Time
					LastRunTime       *time.Time
					LastRunExecution  *common.WorkflowExecution
					LastRunActualTime *time.Time
				}
			}
			info := sch.GetInfo()
			if info == nil {
				missingExtendedInfo = true
			}
			item.ScheduleId = sch.ScheduleId
			item.StartWorkflow.WorkflowType = info.GetWorkflowType().GetName()
			item.State.Paused = info.GetPaused()
			item.State.Notes = info.GetNotes()
			if fas := info.GetFutureActionTimes(); len(fas) > 0 {
				item.Info.NextRunTime = fas[0]
			}
			if ras := info.GetRecentActions(); len(ras) > 0 {
				ra := ras[len(ras)-1]
				item.Info.LastRunTime = ra.ScheduleTime
				item.Info.LastRunActualTime = ra.ActualTime
				item.Info.LastRunExecution = ra.StartWorkflowResult
			}
			item.Specification = info.GetSpec()
			uncanonicalizeSpec(item.Specification)
			items[i] = item
		}
		return items, resp.NextPageToken, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &output.PrintOptions{
		Fields:     []string{"ScheduleId", "StartWorkflow.WorkflowType", "State.Paused", "State.Notes", "Info.NextRunTime", "Info.LastRunTime"},
		FieldsLong: []string{"Info.LastRunActualTime", "Info.LastRunExecution", "Specification"},
		Pager:      pager.Less,
	}
	if missingExtendedInfo {
		fmt.Println(color.YellowString("Note: Extended schedule information is not available without Elasticsearch"))
		opts.Fields = []string{"ScheduleId"}
		opts.FieldsLong = nil
	}
	return output.PrintIterator(nil, iter, opts)
}

func uncanonicalizeSpec(spec *schedpb.ScheduleSpec) {
	if spec == nil {
		return
	}
	processField := func(ranges []*schedpb.Range) string {
		var out []string
		for _, r := range ranges {
			s := fmt.Sprintf("%d", r.Start)
			if r.End > r.Start {
				s += fmt.Sprintf("-%d", r.End)
			}
			if r.Step > 1 {
				s += fmt.Sprintf("/%d", r.Step)
			}
			out = append(out, s)
		}
		return strings.Join(out, ",")
	}
	// Turn StructuredCalenderSpec into CalendarSpec for ease of reading
	for _, scs := range spec.StructuredCalendar {
		spec.Calendar = append(spec.Calendar, &schedpb.CalendarSpec{
			Second:     processField(scs.Second),
			Minute:     processField(scs.Minute),
			Hour:       processField(scs.Hour),
			DayOfMonth: processField(scs.DayOfMonth),
			Month:      processField(scs.Month),
			Year:       processField(scs.Year),
			DayOfWeek:  processField(scs.DayOfWeek),
			Comment:    scs.Comment,
		})
	}
	spec.StructuredCalendar = nil
}

func startWorkflowBaseArgs(c *cli.Context) (
	taskQueue string,
	workflowType string,
	et, rt, dt int,
	wid string,
) {
	taskQueue = c.String(FlagTaskQueue)
	workflowType = c.String(FlagWorkflowType)
	et = c.Int(FlagWorkflowExecutionTimeout)
	rt = c.Int(FlagWorkflowRunTimeout)
	dt = c.Int(FlagWorkflowTaskTimeout)
	wid = c.String(FlagWorkflowID)
	if len(wid) == 0 {
		wid = uuid.New()
	}
	return
}
