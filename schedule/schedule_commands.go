package schedule

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pborman/uuid"
	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/cli/workflow"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	schedpb "go.temporal.io/api/schedule/v1"
	"go.temporal.io/api/taskqueue/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/server/common/collection"
	"go.temporal.io/server/common/primitives/timestamp"
)

func scheduleBaseArgs(c *cli.Context) (
	frontendClient workflowservice.WorkflowServiceClient,
	namespace string,
	scheduleID string,
	err error,
) {
	frontendClient = client.Factory(c.App).FrontendClient(c)
	namespace, err = common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return nil, "", "", err
	}
	scheduleID, err = common.RequiredFlag(c, common.FlagScheduleID)
	if err != nil {
		return nil, "", "", err
	}
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
	for _, s := range c.StringSlice(common.FlagCalendar) {
		cal, err := buildCalendarSpec(s)
		if err != nil {
			return nil, err
		}
		out.Calendar = append(out.Calendar, cal)
	}
	out.CronString = c.StringSlice(common.FlagCronSchedule)
	for _, s := range c.StringSlice(common.FlagInterval) {
		cal, err := buildIntervalSpec(s)
		if err != nil {
			return nil, err
		}
		out.Interval = append(out.Interval, cal)
	}
	if c.IsSet(common.FlagStartTime) {
		t, err := common.ParseTime(c.String(common.FlagStartTime), time.Time{}, now)
		if err != nil {
			return nil, err
		}
		out.StartTime = timestamp.TimePtr(t)
	}
	if c.IsSet(common.FlagEndTime) {
		t, err := common.ParseTime(c.String(common.FlagEndTime), time.Time{}, now)
		if err != nil {
			return nil, err
		}
		out.EndTime = timestamp.TimePtr(t)
	}
	if c.IsSet(common.FlagJitter) {
		d, err := timestamp.ParseDuration(c.String(common.FlagJitter))
		if err != nil {
			return nil, err
		}
		out.Jitter = timestamp.DurationPtr(d)
	}
	if c.IsSet(common.FlagTimeZone) {
		tzName := c.String(common.FlagTimeZone)
		if _, err := time.LoadLocation(tzName); err != nil {
			return nil, fmt.Errorf("unknown time zone name %q", tzName)
		}
		out.TimezoneName = tzName
	}
	return &out, nil
}

func buildScheduleAction(c *cli.Context) (*schedpb.ScheduleAction, error) {
	taskQueue, workflowType, et, rt, dt, wid := workflow.StartWorkflowBaseArgs(c)
	inputs, err := common.ProcessJSONInput(c)
	if err != nil {
		return nil, err
	}

	// TODO: allow specifying: memo, search attributes, workflow retry policy

	newWorkflow := &workflowpb.NewWorkflowExecutionInfo{
		WorkflowId:               wid,
		WorkflowType:             &commonpb.WorkflowType{Name: workflowType},
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
	out.Notes = c.String(common.FlagNotes)
	out.Paused = c.Bool(common.FlagPause)
	if c.IsSet(common.FlagRemainingActions) {
		out.LimitedActions = true
		out.RemainingActions = int64(c.Int(common.FlagRemainingActions))
	}
	return &out, nil
}

func getOverlapPolicy(c *cli.Context) (enumspb.ScheduleOverlapPolicy, error) {
	i, err := common.StringToEnum(c.String(common.FlagOverlapPolicy), enumspb.ScheduleOverlapPolicy_value)
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
	if c.IsSet(common.FlagCatchupWindow) {
		d, err := timestamp.ParseDuration(c.String(common.FlagCatchupWindow))
		if err != nil {
			return nil, err
		}
		out.CatchupWindow = timestamp.DurationPtr(d)
	}
	out.PauseOnFailure = c.Bool(common.FlagPauseOnFailure)
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

func getMemoAndSearchAttributesForSchedule(c *cli.Context) (*commonpb.Memo, *commonpb.SearchAttributes, error) {
	if memoMap, err := workflow.UnmarshalMemoFromCLI(c); err != nil {
		return nil, nil, err
	} else if memo, err := encodeMemo(memoMap); err != nil {
		return nil, nil, err
	} else if saMap, err := workflow.UnmarshalSearchAttrFromCLI(c); err != nil {
		return nil, nil, err
	} else if sa, err := encodeSearchAttributes(saMap); err != nil {
		return nil, nil, err
	} else {
		return memo, sa, nil
	}
}

func CreateSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
	defer cancel()

	sched, err := buildSchedule(c)
	if err != nil {
		return err
	}
	memo, sa, err := getMemoAndSearchAttributesForSchedule(c)
	if err != nil {
		return err
	}

	req := &workflowservice.CreateScheduleRequest{
		Namespace:        namespace,
		ScheduleId:       scheduleID,
		Schedule:         sched,
		Identity:         common.GetCliIdentity(),
		RequestId:        uuid.New(),
		Memo:             memo,
		SearchAttributes: sa,
	}

	_, err = frontendClient.CreateSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create schedule: %w", err)
	}

	fmt.Println(color.Green(c, "Schedule created"))
	return nil
}

func UpdateSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
	defer cancel()

	sched, err := buildSchedule(c)
	if err != nil {
		return err
	}

	req := &workflowservice.UpdateScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Schedule:   sched,
		Identity:   common.GetCliIdentity(),
		RequestId:  uuid.New(),
	}

	_, err = frontendClient.UpdateSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to update schedule: %w", err)
	}

	fmt.Println(color.Green(c, "Schedule updated"))
	return nil
}

func ToggleSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
	defer cancel()

	pause, unpause := c.Bool(common.FlagPause), c.Bool(common.FlagUnpause)
	if pause && unpause {
		return errors.New("specify either --pause or --unpause")
	} else if !pause && !unpause {
		return errors.New("specify either --pause or --unpause")
	}
	patch := &schedpb.SchedulePatch{}
	if pause {
		patch.Pause = c.String(common.FlagReason)
	} else if unpause {
		patch.Unpause = c.String(common.FlagReason)
	}

	req := &workflowservice.PatchScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Patch:      patch,
		Identity:   common.GetCliIdentity(),
		RequestId:  uuid.New(),
	}
	_, err = frontendClient.PatchSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to toggle schedule: %w", err)
	}

	fmt.Println(color.Green(c, "Schedule updated"))
	return nil
}

func TriggerSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
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
		Identity:  common.GetCliIdentity(),
		RequestId: uuid.New(),
	}
	_, err = frontendClient.PatchSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to trigger schedule: %w", err)
	}

	fmt.Println(color.Green(c, "Trigger request sent"))
	return nil
}

func BackfillSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
	defer cancel()

	now := time.Now()
	startTime, err := common.ParseTime(c.String(common.FlagStartTime), time.Time{}, now)
	if err != nil {
		return err
	}
	endTime, err := common.ParseTime(c.String(common.FlagEndTime), time.Time{}, now)
	if err != nil {
		return err
	}
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
		Identity:  common.GetCliIdentity(),
		RequestId: uuid.New(),
	}
	_, err = frontendClient.PatchSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to backfill schedule: %w", err)
	}

	fmt.Println(color.Green(c, "Backfill request sent"))
	return nil
}

func DescribeSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
	defer cancel()

	req := &workflowservice.DescribeScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
	}
	resp, err := frontendClient.DescribeSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to describe schedule: %w", err)
	}

	if c.Bool(common.FlagPrintRaw) {
		common.PrettyPrintJSONObject(c, resp)
		return nil
	}

	// output.PrintItems gets confused by nested fields of nil values, because it uses
	// reflection. ensure the first level is non-nil to avoid runtime errors.
	common.EnsureNonNil(&resp.Schedule)
	common.EnsureNonNil(&resp.Schedule.Spec)
	common.EnsureNonNil(&resp.Schedule.Action)
	common.EnsureNonNil(&resp.Schedule.Policies)
	common.EnsureNonNil(&resp.Schedule.State)
	common.EnsureNonNil(&resp.Info)

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
		LastRunExecution  *commonpb.WorkflowExecution
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
		item.Input = converter.GetDefaultDataConverter().ToStrings(sw.Input)
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
			item.Memo[k] = converter.GetDefaultDataConverter().ToString(payload)
		}
	}
	if fields := resp.SearchAttributes.GetIndexedFields(); len(fields) > 0 {
		item.SearchAttributes = make(map[string]string, len(fields))
		for k, payload := range fields {
			item.SearchAttributes[k] = converter.GetDefaultDataConverter().ToString(payload)
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
	return output.PrintItems(c, []interface{}{item}, opts)
}

func DeleteSchedule(c *cli.Context) error {
	frontendClient, namespace, scheduleID, err := scheduleBaseArgs(c)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
	defer cancel()

	req := &workflowservice.DeleteScheduleRequest{
		Namespace:  namespace,
		ScheduleId: scheduleID,
		Identity:   common.GetCliIdentity(),
	}
	_, err = frontendClient.DeleteSchedule(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to delete schedule: %w", err)
	}

	fmt.Println(color.Green(c, "Schedule deleted"))
	return nil
}

func ListSchedules(c *cli.Context) error {
	frontendClient := client.Factory(c.App).FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	ctx, cancel := common.NewContext(c)
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
					LastRunExecution  *commonpb.WorkflowExecution
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
		fmt.Println(color.Yellow(c, "Note: Extended schedule information is not available without Elasticsearch"))
		opts.Fields = []string{"ScheduleId"}
		opts.FieldsLong = nil
	}
	return output.PrintIterator(c, iter, opts)
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

func encodeMemo(memo map[string]interface{}) (*commonpb.Memo, error) {
	if len(memo) == 0 {
		return nil, nil
	}
	dc := converter.GetDefaultDataConverter()
	fields := make(map[string]*commonpb.Payload, len(memo))
	var err error
	for k, v := range memo {
		fields[k], err = dc.ToPayload(v)
		if err != nil {
			return nil, err
		}
	}
	return &commonpb.Memo{Fields: fields}, nil
}

func encodeSearchAttributes(sa map[string]interface{}) (*commonpb.SearchAttributes, error) {
	if len(sa) == 0 {
		return nil, nil
	}
	dc := converter.GetDefaultDataConverter()
	fields := make(map[string]*commonpb.Payload, len(sa))
	var err error
	for k, v := range sa {
		fields[k], err = dc.ToPayload(v)
		if err != nil {
			return nil, err
		}
	}
	return &commonpb.SearchAttributes{IndexedFields: fields}, nil
}
