package temporalcli

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/temporalio/cli/cliext"
	"github.com/temporalio/cli/internal/printer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	schedpb "go.temporal.io/api/schedule/v1"
	sdkpb "go.temporal.io/api/sdk/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/proto"
)

type printableSchedule struct {
	ScheduleId string
	// Schedule.Action
	Action any // list has Workflow only
	// Schedule.Spec
	Spec         []any     `cli:",cardOmitEmpty"` // can contain *schedpb.CalendarSpec or printableInterval
	SkipSpec     []any     `cli:",cardOmitEmpty"`
	StartAt      time.Time `cli:",cardOmitEmpty"`
	EndAt        time.Time `cli:",cardOmitEmpty"`
	Jitter       string    `cli:",cardOmitEmpty"`
	TimeZoneName string    `cli:",cardOmitEmpty"`
	// Schedule.Policy
	OverlapPolicy  enumspb.ScheduleOverlapPolicy // describe only
	CatchupWindow  string                        // describe only
	PauseOnFailure bool                          // describe only
	// Schedule.State
	Notes            string `cli:",cardOmitEmpty"`
	Paused           bool
	LimitedActions   bool   `cli:",cardOmitEmpty"` // describe only
	RemainingActions string `cli:",cardOmitEmpty"` // describe only; string so we can hide it
	// Info
	NextRunTime      time.Time
	LastRunTime      time.Time
	RunningWorkflows []string      // describe only
	CreatedAt        time.Time     `cli:",cardOmitEmpty"` // describe only
	LastUpdateAt     time.Time     `cli:",cardOmitEmpty"` // describe only
	ActionCounts     *actionCounts `cli:",cardOmitEmpty"` // describe only
	// SearchAttributes, Memo
	SearchAttributes *commonpb.SearchAttributes `cli:",cardOmitEmpty"`
	Memo             *commonpb.Memo             `cli:",cardOmitEmpty"`
}

type actionCounts struct {
	Total               int
	MissedCatchupWindow int
	SkippedOverlap      int
}

// Neither protojson nor fmt print structs containing time.Durations nicely, so do it manually
// using a struct of strings.
type printableInterval struct {
	Every  string `json:"every"`
	Offset string `json:"offset,omitempty"`
}

func describeResultToPrintable(id string, desc *client.ScheduleDescription) *printableSchedule {
	// ID, SearchAttributes, Memo
	out := &printableSchedule{
		ScheduleId:       id,
		SearchAttributes: desc.SearchAttributes,
		Memo:             desc.Memo,
	}
	// Schedule.Action
	out.Action = desc.Schedule.Action
	// Schedule.Spec
	specToPrintable(out, desc.Schedule.Spec)
	// Schedule.Policy
	out.OverlapPolicy = desc.Schedule.Policy.Overlap
	out.CatchupWindow = formatDuration(desc.Schedule.Policy.CatchupWindow)
	out.PauseOnFailure = desc.Schedule.Policy.PauseOnFailure
	// Schedule.State
	out.Notes = desc.Schedule.State.Note
	out.Paused = desc.Schedule.State.Paused
	if out.LimitedActions = desc.Schedule.State.LimitedActions; out.LimitedActions {
		out.RemainingActions = strconv.Itoa(desc.Schedule.State.RemainingActions)
	}
	// Info
	if len(desc.Info.NextActionTimes) > 0 {
		out.NextRunTime = desc.Info.NextActionTimes[0]
	}
	if l := len(desc.Info.RecentActions); l > 0 {
		last := desc.Info.RecentActions[l-1]
		out.LastRunTime = last.ScheduleTime
	}
	for _, w := range desc.Info.RunningWorkflows {
		out.RunningWorkflows = append(out.RunningWorkflows, w.WorkflowID)
	}
	out.CreatedAt = desc.Info.CreatedAt
	out.LastUpdateAt = desc.Info.LastUpdateAt
	out.ActionCounts = &actionCounts{
		Total:               desc.Info.NumActions,
		MissedCatchupWindow: desc.Info.NumActionsMissedCatchupWindow,
		SkippedOverlap:      desc.Info.NumActionsSkippedOverlap,
	}

	return out
}

func listEntryToPrintable(ent *client.ScheduleListEntry) *printableSchedule {
	out := &printableSchedule{
		ScheduleId:       ent.ID,
		Paused:           ent.Paused,
		Notes:            ent.Note,
		Action:           struct{ Workflow string }{Workflow: ent.WorkflowType.Name},
		SearchAttributes: ent.SearchAttributes,
		Memo:             ent.Memo,
	}
	specToPrintable(out, ent.Spec)
	if len(ent.NextActionTimes) > 0 {
		out.NextRunTime = ent.NextActionTimes[0]
	}
	if l := len(ent.RecentActions); l > 0 {
		last := ent.RecentActions[l-1]
		out.LastRunTime = last.ScheduleTime
	}
	return out
}

func specToPrintable(out *printableSchedule, spec *client.ScheduleSpec) {
	for _, cal := range spec.Calendars {
		out.Spec = append(out.Spec, formatCalendarSpec(cal))
	}
	for _, cal := range spec.Skip {
		out.SkipSpec = append(out.SkipSpec, formatCalendarSpec(cal))
	}
	for _, int := range spec.Intervals {
		pInt := printableInterval{Every: formatDuration(int.Every)}
		if int.Offset > 0 {
			pInt.Offset = formatDuration(int.Offset)
		}
		out.Spec = append(out.Spec, pInt)
	}
	out.StartAt = spec.StartAt
	out.EndAt = spec.EndAt
	out.Jitter = formatDuration(spec.Jitter)
	out.TimeZoneName = spec.TimeZoneName
}

func (c *TemporalScheduleBackfillCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)

	overlap, err := enumspb.ScheduleOverlapPolicyFromString(c.OverlapPolicy.Value)
	if err != nil {
		return err
	}

	err = sch.Backfill(cctx, client.ScheduleBackfillOptions{
		Backfill: []client.ScheduleBackfill{
			{
				Start:   c.StartTime.Time(),
				End:     c.EndTime.Time(),
				Overlap: overlap,
			},
		},
	})
	if err != nil {
		return err
	}
	cctx.Printer.Println("Backfill request sent")
	return nil
}

func toCronString(pb *schedpb.CalendarSpec) (string, error) {
	def := func(a, b string) string {
		if a != "" {
			return a
		}
		return b
	}
	fields := []string{
		def(pb.Second, "0"),
		def(pb.Minute, "0"),
		def(pb.Hour, "0"),
		def(pb.DayOfMonth, "*"),
		def(pb.Month, "*"),
		def(pb.DayOfWeek, "*"),
		def(pb.Year, "*"),
	}
	for _, f := range fields {
		if len(strings.Fields(f)) != 1 {
			return "", fmt.Errorf("invalid CalendarSpec")
		}
	}
	if pb.Comment != "" {
		fields = append(fields, "#", pb.Comment)
	}
	return strings.Join(fields, " "), nil
}

func toIntervalSpec(str string) (client.ScheduleIntervalSpec, error) {
	var spec client.ScheduleIntervalSpec
	var err error
	parts := strings.Split(str, "/")
	if len(parts) > 2 {
		return spec, fmt.Errorf(`invalid interval: must be "<duration>" or "<duration>/<duration>"`)
	} else if len(parts) == 2 {
		if spec.Offset, err = cliext.ParseFlagDuration(parts[1]); err != nil {
			return spec, fmt.Errorf("invalid interval: %w", err)
		}
	}
	if spec.Every, err = cliext.ParseFlagDuration(parts[0]); err != nil {
		return spec, fmt.Errorf("invalid interval: %w", err)
	}
	return spec, nil
}

func (c *ScheduleConfigurationOptions) toScheduleSpec(spec *client.ScheduleSpec) error {
	spec.CronExpressions = c.Cron
	// Skip not supported
	spec.Jitter = c.Jitter.Duration()
	spec.TimeZoneName = c.TimeZone
	spec.StartAt = c.StartTime.Time()
	spec.EndAt = c.EndTime.Time()

	var err error
	for _, calPbStr := range c.Calendar {
		var calPb schedpb.CalendarSpec
		if err = protojson.Unmarshal([]byte(calPbStr), &calPb); err != nil {
			return fmt.Errorf("failed to parse json calendar spec: %w", err)
		}
		cron, err := toCronString(&calPb)
		if err != nil {
			return err
		}
		spec.CronExpressions = append(spec.CronExpressions, cron)
	}
	for _, intStr := range c.Interval {
		int, err := toIntervalSpec(intStr)
		if err != nil {
			return err
		}
		spec.Intervals = append(spec.Intervals, int)
	}

	return nil
}

func toScheduleAction(sw *SharedWorkflowStartOptions, i *PayloadInputOptions) (client.ScheduleAction, error) {
	if len(sw.Headers) > 0 {
		return nil, fmt.Errorf("headers are not supported for schedule actions")
	}
	if sw.PriorityKey < 0 || sw.PriorityKey > 5 {
		return nil, fmt.Errorf("priority key must be between 0 and 5")
	}
	if len(sw.FairnessKey) > 64 {
		return nil, fmt.Errorf("fairness key must be at most 64 bytes")
	}
	if math.IsNaN(float64(sw.FairnessWeight)) ||
		sw.FairnessWeight < 0 ||
		(sw.FairnessWeight > 0 && sw.FairnessWeight < 0.001) ||
		sw.FairnessWeight > 1000 {
		return nil, fmt.Errorf("fairness weight must be between 0.001 and 1000")
	}

	opts, err := buildStartOptions(sw, &WorkflowStartOptions{})
	if err != nil {
		return nil, err
	}
	untypedSearchAttributes, err := encodeMapToPayloads(opts.SearchAttributes)
	if err != nil {
		return nil, err
	}
	action := &client.ScheduleWorkflowAction{
		ID:                       opts.ID,
		Workflow:                 sw.Type,
		TaskQueue:                opts.TaskQueue,
		WorkflowExecutionTimeout: opts.WorkflowExecutionTimeout,
		WorkflowRunTimeout:       opts.WorkflowRunTimeout,
		WorkflowTaskTimeout:      opts.WorkflowTaskTimeout,
		// RetryPolicy not supported yet
		UntypedSearchAttributes: untypedSearchAttributes,
		Memo:                    opts.Memo,
		StaticSummary:           opts.StaticSummary,
		StaticDetails:           opts.StaticDetails,
		Priority:                opts.Priority,
	}
	if action.Args, err = i.buildRawInput(); err != nil {
		return nil, err
	}
	return action, nil
}

func (c *TemporalScheduleCreateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	opts := client.ScheduleOptions{
		ID:               c.ScheduleId,
		PauseOnFailure:   c.PauseOnFailure,
		Note:             c.Notes,
		Paused:           c.Paused,
		CatchupWindow:    c.CatchupWindow.Duration(),
		RemainingActions: c.RemainingActions,
		// TriggerImmediately not supported
		// ScheduleBackfill not supported
	}

	if err = c.toScheduleSpec(&opts.Spec); err != nil {
		return err
	} else if opts.Action, err = toScheduleAction(&c.SharedWorkflowStartOptions, &c.PayloadInputOptions); err != nil {
		return err
	} else if opts.Overlap, err = enumspb.ScheduleOverlapPolicyFromString(c.OverlapPolicy.Value); err != nil {
		return err
	} else if opts.Memo, err = stringKeysJSONValues(c.ScheduleMemo, false); err != nil {
		return fmt.Errorf("invalid memo values: %w", err)
	} else if opts.SearchAttributes, err = stringKeysJSONValues(c.ScheduleSearchAttribute, false); err != nil {
		return fmt.Errorf("invalid search attribute values: %w", err)
	}

	_, err = cl.ScheduleClient().Create(cctx, opts)
	return err
}

func (c *TemporalScheduleDeleteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)
	err = sch.Delete(cctx)
	if err != nil {
		return err
	}
	cctx.Printer.Println("Schedule deleted")
	return nil
}

func (c *TemporalScheduleDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	if cctx.JSONOutput {
		// Use raw gRPC for stability
		res, err := cl.WorkflowService().DescribeSchedule(cctx, &workflowservice.DescribeScheduleRequest{
			Namespace:  c.Parent.Namespace,
			ScheduleId: c.ScheduleId,
		})
		if err != nil {
			return err
		}
		// TODO: remove this after https://github.com/temporalio/api-go/pull/154
		noShorthand := false
		return cctx.Printer.PrintStructuredErr(res, printer.StructuredOptions{
			OverrideJSONPayloadShorthand: &noShorthand,
		})
	}

	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)
	res, err := sch.Describe(cctx)
	if err != nil {
		return err
	}

	printable := describeResultToPrintable(c.ScheduleId, res)
	return cctx.Printer.PrintStructuredErr(printable, printer.StructuredOptions{})
}

func finishScheduleList(p *printer.Printer, primaryErr error) error {
	return errors.Join(primaryErr, p.EndListErr())
}

func (c *TemporalScheduleListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	if cctx.JSONOutput {
		// Use raw gRPC for stability
		// This is a listing command subject to json vs jsonl rules
		if err := cctx.Printer.StartListErr(); err != nil {
			return err
		}

		var token []byte
		for {
			res, err := cl.WorkflowService().ListSchedules(cctx, &workflowservice.ListSchedulesRequest{
				Namespace:     c.Parent.Namespace,
				NextPageToken: token,
				Query:         c.Query,
			})
			if err != nil {
				return finishScheduleList(cctx.Printer, err)
			}
			// TODO: remove this after https://github.com/temporalio/api-go/pull/154
			noShorthand := false
			for _, entry := range res.Schedules {
				if err := cctx.Printer.PrintStructuredErr(entry, printer.StructuredOptions{
					OverrideJSONPayloadShorthand: &noShorthand,
				}); err != nil {
					return finishScheduleList(cctx.Printer, err)
				}
			}
			if token = res.NextPageToken; len(token) == 0 {
				break
			}
		}

		return finishScheduleList(cctx.Printer, nil)
	}

	res, err := cl.ScheduleClient().List(cctx, client.ScheduleListOptions{
		Query: c.Query,
	})
	if err != nil {
		return err
	}

	// This is a listing command subject to json vs jsonl rules
	if err := cctx.Printer.StartListErr(); err != nil {
		return err
	}

	printOpts := printer.StructuredOptions{
		ExcludeFields: []string{
			// These aren't available in list results
			"OverlapPolicy",
			"CatchupWindow",
			"PauseOnFailure",
			"LimitedActions",
			"RemainingActions",
			"RunningWorkflows",
			"CreatedAt",
			"LastUpdateAt",
			"ActionCounts",
		},
		Table: &printer.TableOptions{},
	}

	if !c.Long && !c.ReallyLong {
		printOpts.ExcludeFields = append(printOpts.ExcludeFields,
			"Spec",
			"Notes",
		)
	}

	if !c.ReallyLong {
		printOpts.ExcludeFields = append(printOpts.ExcludeFields,
			"SkipSpec",
			"StartAt",
			"EndAt",
			"Jitter",
			"TimeZoneName",
			"SearchAttributes",
			"Memo",
		)
	}

	// make artificial "pages" so we get better aligned columns
	page := make([]*printableSchedule, 0, 100)

	for res.HasNext() {
		ent, err := res.Next()
		if err != nil {
			return finishScheduleList(cctx.Printer, err)
		}
		page = append(page, listEntryToPrintable(ent))
		if len(page) == cap(page) {
			if err := cctx.Printer.PrintStructuredErr(page, printOpts); err != nil {
				return finishScheduleList(cctx.Printer, err)
			}
			page = page[:0]
			printOpts.Table.NoHeader = true
		}
	}
	if err := cctx.Printer.PrintStructuredErr(page, printOpts); err != nil {
		return finishScheduleList(cctx.Printer, err)
	}

	return finishScheduleList(cctx.Printer, nil)
}

func (c *TemporalScheduleToggleCommand) run(cctx *CommandContext, args []string) error {
	if c.Pause == c.Unpause {
		return errors.New("exactly one of --pause or --unpause is required")
	}

	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)

	if c.Pause {
		return sch.Pause(cctx, client.SchedulePauseOptions{
			Note: c.Reason,
		})
	} else {
		return sch.Unpause(cctx, client.ScheduleUnpauseOptions{
			Note: c.Reason,
		})
	}
}

func (c *TemporalScheduleTriggerCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)

	overlap, err := enumspb.ScheduleOverlapPolicyFromString(c.OverlapPolicy.Value)
	if err != nil {
		return err
	}

	err = sch.Trigger(cctx, client.ScheduleTriggerOptions{
		Overlap: overlap,
	})
	if err != nil {
		return err
	}
	cctx.Printer.Println("Trigger request sent")
	return nil
}

type schedulePatchIntent struct {
	notesSet              bool
	notesUnset            bool
	notes                 string
	overlapSet            bool
	overlap               enumspb.ScheduleOverlapPolicy
	catchupSet            bool
	catchup               time.Duration
	catchupUnset          bool
	pauseOnFailureSet     bool
	pauseOnFailure        bool
	pausedSet             bool
	paused                bool
	remainingActionsSet   bool
	remainingActions      int
	calendarSet           bool
	cronSet               bool
	intervalSet           bool
	calendar              []string
	cron                  []string
	interval              []*schedpb.IntervalSpec
	cadenceClearAll       bool
	startTimeSet          bool
	startTimeUnset        bool
	startTime             *timestamppb.Timestamp
	endTimeSet            bool
	endTimeUnset          bool
	endTime               *timestamppb.Timestamp
	jitterSet             bool
	jitterUnset           bool
	jitter                time.Duration
	timeZoneSet           bool
	timeZoneUnset         bool
	timeZone              string
	workflowIDSet         bool
	workflowID            string
	workflowTypeSet       bool
	workflowType          string
	taskQueueSet          bool
	taskQueue             string
	executionTimeoutSet   bool
	executionTimeoutUnset bool
	executionTimeout      time.Duration
	runTimeoutSet         bool
	runTimeoutUnset       bool
	runTimeout            time.Duration
	taskTimeoutSet        bool
	taskTimeoutUnset      bool
	taskTimeout           time.Duration
	staticSummarySet      bool
	staticSummaryUnset    bool
	staticSummary         *commonpb.Payload
	staticDetailsSet      bool
	staticDetailsUnset    bool
	staticDetails         *commonpb.Payload
}

func (i schedulePatchIntent) validate() error {
	if !i.notesSet && !i.notesUnset && !i.overlapSet && !i.catchupSet && !i.catchupUnset && !i.pauseOnFailureSet && !i.pausedSet && !i.remainingActionsSet && !i.calendarSet && !i.cronSet && !i.intervalSet && !i.cadenceClearAll && !i.startTimeSet && !i.startTimeUnset && !i.endTimeSet && !i.endTimeUnset && !i.jitterSet && !i.jitterUnset && !i.timeZoneSet && !i.timeZoneUnset && !i.workflowIDSet && !i.workflowTypeSet && !i.taskQueueSet && !i.executionTimeoutSet && !i.executionTimeoutUnset && !i.runTimeoutSet && !i.runTimeoutUnset && !i.taskTimeoutSet && !i.taskTimeoutUnset && !i.staticSummarySet && !i.staticSummaryUnset && !i.staticDetailsSet && !i.staticDetailsUnset {
		return errors.New("at least one patch operation is required")
	}
	if i.notesSet && i.notesUnset {
		return errors.New("--notes and --unset-notes are mutually exclusive")
	}
	if i.catchupSet && i.catchupUnset {
		return errors.New("--catchup-window and --unset-catchup-window are mutually exclusive")
	}
	if i.catchupSet && i.catchup < 10*time.Second {
		return errors.New("catchup window must be at least 10s")
	}
	if i.remainingActionsSet && i.remainingActions < 0 {
		return errors.New("remaining actions must not be negative")
	}
	if i.startTimeSet && i.startTimeUnset {
		return errors.New("--start-time and --unset-start-time are mutually exclusive")
	}
	if i.startTimeSet {
		if err := i.startTime.CheckValid(); err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}
	}
	if i.endTimeSet && i.endTimeUnset {
		return errors.New("--end-time and --unset-end-time are mutually exclusive")
	}
	if i.endTimeSet {
		if err := i.endTime.CheckValid(); err != nil {
			return fmt.Errorf("invalid end time: %w", err)
		}
	}
	if i.jitterSet && i.jitterUnset {
		return errors.New("--jitter and --unset-jitter are mutually exclusive")
	}
	if i.timeZoneSet && i.timeZoneUnset {
		return errors.New("--time-zone and --unset-time-zone are mutually exclusive")
	}
	if i.timeZoneSet && strings.TrimSpace(i.timeZone) == "" {
		return errors.New("--time-zone requires a non-empty value; use --unset-time-zone to clear")
	}
	if i.workflowIDSet && i.workflowID == "" {
		return errors.New("workflow ID must not be empty")
	}
	if i.workflowTypeSet && i.workflowType == "" {
		return errors.New("workflow type must not be empty")
	}
	if i.taskQueueSet && i.taskQueue == "" {
		return errors.New("task queue must not be empty")
	}
	if i.executionTimeoutSet && i.executionTimeoutUnset {
		return errors.New("--execution-timeout and --unset-execution-timeout are mutually exclusive")
	}
	if i.executionTimeoutSet && i.executionTimeout < 0 {
		return errors.New("execution timeout must not be negative")
	}
	if i.runTimeoutSet && i.runTimeoutUnset {
		return errors.New("--run-timeout and --unset-run-timeout are mutually exclusive")
	}
	if i.runTimeoutSet && i.runTimeout < 0 {
		return errors.New("run timeout must not be negative")
	}
	if i.taskTimeoutSet && i.taskTimeoutUnset {
		return errors.New("--task-timeout and --unset-task-timeout are mutually exclusive")
	}
	if i.taskTimeoutSet && i.taskTimeout < 0 {
		return errors.New("task timeout must not be negative")
	}
	if i.staticSummarySet && i.staticSummaryUnset {
		return errors.New("--static-summary and --unset-static-summary are mutually exclusive")
	}
	if i.staticDetailsSet && i.staticDetailsUnset {
		return errors.New("--static-details and --unset-static-details are mutually exclusive")
	}
	if i.cadenceClearAll && (i.calendarSet || i.cronSet || i.intervalSet) {
		return errors.New("--cadence-clear-all cannot be combined with a cadence source")
	}
	if i.jitterSet && i.jitter < 0 {
		return errors.New("jitter must not be negative")
	}
	return nil
}

func (i schedulePatchIntent) validateResult(schedule *schedpb.Schedule) error {
	if i.cadenceClearAll && !schedule.GetState().GetPaused() {
		return errors.New("--cadence-clear-all requires the Schedule to be paused; use --paused=true to pause explicitly")
	}
	return nil
}

func (i schedulePatchIntent) apply(schedule *schedpb.Schedule) error {
	if i.overlapSet || i.catchupSet || i.catchupUnset || i.pauseOnFailureSet {
		if schedule.Policies == nil {
			schedule.Policies = &schedpb.SchedulePolicies{}
		}
	}
	if i.overlapSet {
		schedule.Policies.OverlapPolicy = i.overlap
	}
	if i.catchupSet {
		schedule.Policies.CatchupWindow = durationpb.New(i.catchup)
	}
	if i.catchupUnset {
		schedule.Policies.CatchupWindow = nil
	}
	if i.pauseOnFailureSet {
		schedule.Policies.PauseOnFailure = i.pauseOnFailure
	}
	if i.pausedSet {
		if schedule.State == nil {
			schedule.State = &schedpb.ScheduleState{}
		}
		schedule.State.Paused = i.paused
	}
	if i.remainingActionsSet {
		if schedule.State == nil {
			schedule.State = &schedpb.ScheduleState{}
		}
		schedule.State.RemainingActions = int64(i.remainingActions)
		schedule.State.LimitedActions = i.remainingActions > 0
	}
	if i.calendarSet || i.cronSet || i.intervalSet || i.cadenceClearAll || i.startTimeSet || i.startTimeUnset || i.endTimeSet || i.endTimeUnset || i.jitterSet || i.jitterUnset || i.timeZoneSet || i.timeZoneUnset {
		if schedule.Spec == nil {
			schedule.Spec = &schedpb.ScheduleSpec{}
		}
	}
	if i.calendarSet || i.cronSet || i.intervalSet || i.cadenceClearAll {
		schedule.Spec.StructuredCalendar = nil
		schedule.Spec.Calendar = nil
		schedule.Spec.CronString = nil
		schedule.Spec.Interval = nil
	}
	if i.calendarSet || i.cronSet || i.intervalSet {
		schedule.Spec.CronString = append(append([]string(nil), i.calendar...), i.cron...)
		schedule.Spec.Interval = append([]*schedpb.IntervalSpec(nil), i.interval...)
	}
	if i.startTimeSet {
		schedule.Spec.StartTime = i.startTime
	}
	if i.startTimeUnset {
		schedule.Spec.StartTime = nil
	}
	if i.endTimeSet {
		schedule.Spec.EndTime = i.endTime
	}
	if i.endTimeUnset {
		schedule.Spec.EndTime = nil
	}
	if i.jitterSet {
		schedule.Spec.Jitter = durationpb.New(i.jitter)
	}
	if i.jitterUnset {
		schedule.Spec.Jitter = nil
	}
	if i.timeZoneSet {
		schedule.Spec.TimezoneName = i.timeZone
		schedule.Spec.TimezoneData = nil
	}
	if i.timeZoneUnset {
		schedule.Spec.TimezoneName = ""
		schedule.Spec.TimezoneData = nil
	}
	if i.workflowIDSet || i.workflowTypeSet || i.taskQueueSet || i.executionTimeoutSet || i.executionTimeoutUnset || i.runTimeoutSet || i.runTimeoutUnset || i.taskTimeoutSet || i.taskTimeoutUnset || i.staticSummarySet || i.staticSummaryUnset || i.staticDetailsSet || i.staticDetailsUnset {
		startWorkflow := schedule.GetAction().GetStartWorkflow()
		if startWorkflow == nil {
			return errors.New("Schedule action does not contain a StartWorkflow action")
		}
		if i.workflowIDSet {
			startWorkflow.WorkflowId = i.workflowID
		}
		if i.workflowTypeSet {
			if startWorkflow.WorkflowType == nil {
				startWorkflow.WorkflowType = &commonpb.WorkflowType{}
			}
			startWorkflow.WorkflowType.Name = i.workflowType
		}
		if i.taskQueueSet {
			if startWorkflow.TaskQueue == nil {
				startWorkflow.TaskQueue = &taskqueuepb.TaskQueue{}
			}
			startWorkflow.TaskQueue.Name = i.taskQueue
		}
		if i.executionTimeoutSet {
			startWorkflow.WorkflowExecutionTimeout = durationpb.New(i.executionTimeout)
		}
		if i.executionTimeoutUnset {
			startWorkflow.WorkflowExecutionTimeout = nil
		}
		if i.runTimeoutSet {
			startWorkflow.WorkflowRunTimeout = durationpb.New(i.runTimeout)
		}
		if i.runTimeoutUnset {
			startWorkflow.WorkflowRunTimeout = nil
		}
		if i.taskTimeoutSet {
			startWorkflow.WorkflowTaskTimeout = durationpb.New(i.taskTimeout)
		}
		if i.taskTimeoutUnset {
			startWorkflow.WorkflowTaskTimeout = durationpb.New(10 * time.Second)
		}
		if i.staticSummarySet || i.staticDetailsSet {
			if startWorkflow.UserMetadata == nil {
				startWorkflow.UserMetadata = &sdkpb.UserMetadata{}
			}
		}
		if i.staticSummarySet {
			startWorkflow.UserMetadata.Summary = i.staticSummary
		}
		if i.staticDetailsSet {
			startWorkflow.UserMetadata.Details = i.staticDetails
		}
		if i.staticSummaryUnset && startWorkflow.UserMetadata != nil {
			startWorkflow.UserMetadata.Summary = nil
		}
		if i.staticDetailsUnset && startWorkflow.UserMetadata != nil {
			startWorkflow.UserMetadata.Details = nil
		}
	}
	if !i.notesSet && !i.notesUnset {
		return nil
	}
	if schedule.State == nil {
		schedule.State = &schedpb.ScheduleState{}
	}
	if i.notesSet {
		schedule.State.Notes = i.notes
		return nil
	}
	schedule.State.Notes = ""
	return nil
}

func schedulePatchCadence(calendar, cron, intervals []string) ([]string, []*schedpb.IntervalSpec, error) {
	calendarCron := make([]string, 0, len(calendar))
	for _, calendarJSON := range calendar {
		var calendarSpec schedpb.CalendarSpec
		if err := protojson.Unmarshal([]byte(calendarJSON), &calendarSpec); err != nil {
			return nil, nil, fmt.Errorf("failed to parse json calendar spec: %w", err)
		}
		calendarCronString, err := toCronString(&calendarSpec)
		if err != nil {
			return nil, nil, err
		}
		calendarCron = append(calendarCron, calendarCronString)
	}
	for _, cronString := range cron {
		trimmedCronString := strings.TrimSpace(cronString)
		if strings.HasPrefix(trimmedCronString, "TZ=") || strings.HasPrefix(trimmedCronString, "CRON_TZ=") {
			return nil, nil, errors.New("cron time zones are not supported; use --time-zone")
		}
	}
	intervalSpecs := make([]*schedpb.IntervalSpec, 0, len(intervals))
	for _, intervalString := range intervals {
		interval, err := toIntervalSpec(intervalString)
		if err != nil {
			return nil, nil, err
		}
		if interval.Every < time.Second {
			return nil, nil, errors.New("interval must be at least 1s")
		}
		if interval.Offset < 0 {
			return nil, nil, errors.New("interval phase must not be negative")
		}
		if interval.Offset >= interval.Every {
			return nil, nil, errors.New("interval phase must be less than the interval")
		}
		intervalSpecs = append(intervalSpecs, &schedpb.IntervalSpec{
			Interval: durationpb.New(interval.Every),
			Phase:    durationpb.New(interval.Offset),
		})
	}
	return calendarCron, intervalSpecs, nil
}

func (c *TemporalSchedulePatchCommand) run(cctx *CommandContext, args []string) error {
	const maxAttempts = 3

	var err error
	intent := schedulePatchIntent{
		notesSet:              c.Command.Flags().Changed("notes"),
		notesUnset:            c.UnsetNotes,
		notes:                 c.Notes,
		overlapSet:            c.Command.Flags().Changed("overlap-policy"),
		catchupSet:            c.Command.Flags().Changed("catchup-window"),
		catchup:               c.CatchupWindow.Duration(),
		catchupUnset:          c.UnsetCatchupWindow,
		pauseOnFailureSet:     c.Command.Flags().Changed("pause-on-failure"),
		pauseOnFailure:        c.PauseOnFailure,
		pausedSet:             c.Command.Flags().Changed("paused"),
		paused:                c.Paused,
		remainingActionsSet:   c.Command.Flags().Changed("remaining-actions"),
		remainingActions:      c.RemainingActions,
		calendarSet:           c.Command.Flags().Changed("calendar"),
		cronSet:               c.Command.Flags().Changed("cron"),
		intervalSet:           c.Command.Flags().Changed("interval"),
		cron:                  c.Cron,
		cadenceClearAll:       c.CadenceClearAll,
		startTimeSet:          c.Command.Flags().Changed("start-time"),
		startTimeUnset:        c.UnsetStartTime,
		startTime:             timestamppb.New(c.StartTime.Time()),
		endTimeSet:            c.Command.Flags().Changed("end-time"),
		endTimeUnset:          c.UnsetEndTime,
		endTime:               timestamppb.New(c.EndTime.Time()),
		jitterSet:             c.Command.Flags().Changed("jitter"),
		jitterUnset:           c.UnsetJitter,
		jitter:                c.Jitter.Duration(),
		timeZoneSet:           c.Command.Flags().Changed("time-zone"),
		timeZoneUnset:         c.UnsetTimeZone,
		timeZone:              c.TimeZone,
		workflowIDSet:         c.Command.Flags().Changed("workflow-id"),
		workflowID:            c.WorkflowId,
		workflowTypeSet:       c.Command.Flags().Changed("type"),
		workflowType:          c.Type,
		taskQueueSet:          c.Command.Flags().Changed("task-queue"),
		taskQueue:             c.TaskQueue,
		executionTimeoutSet:   c.Command.Flags().Changed("execution-timeout"),
		executionTimeoutUnset: c.UnsetExecutionTimeout,
		executionTimeout:      c.ExecutionTimeout.Duration(),
		runTimeoutSet:         c.Command.Flags().Changed("run-timeout"),
		runTimeoutUnset:       c.UnsetRunTimeout,
		runTimeout:            c.RunTimeout.Duration(),
		taskTimeoutSet:        c.Command.Flags().Changed("task-timeout"),
		taskTimeoutUnset:      c.UnsetTaskTimeout,
		taskTimeout:           c.TaskTimeout.Duration(),
		staticSummarySet:      c.Command.Flags().Changed("static-summary"),
		staticSummaryUnset:    c.UnsetStaticSummary,
		staticDetailsSet:      c.Command.Flags().Changed("static-details"),
		staticDetailsUnset:    c.UnsetStaticDetails,
	}
	if intent.staticSummarySet {
		intent.staticSummary, err = DataConverterWithRawValue.ToPayload(c.StaticSummary)
		if err != nil {
			return fmt.Errorf("failed to encode static summary: %w", err)
		}
	}
	if intent.staticDetailsSet {
		intent.staticDetails, err = DataConverterWithRawValue.ToPayload(c.StaticDetails)
		if err != nil {
			return fmt.Errorf("failed to encode static details: %w", err)
		}
	}
	if intent.calendarSet || intent.cronSet || intent.intervalSet {
		intent.calendar, intent.interval, err = schedulePatchCadence(c.Calendar, c.Cron, c.Interval)
		if err != nil {
			return err
		}
	}
	if intent.overlapSet {
		intent.overlap, err = enumspb.ScheduleOverlapPolicyFromString(c.OverlapPolicy.Value)
		if err != nil {
			return err
		}
	}
	if err := intent.validate(); err != nil {
		return err
	}
	if c.ScheduleId == "" {
		return errors.New("schedule ID is required")
	}

	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	for attempt := 0; attempt < maxAttempts; attempt++ {
		describeResponse, err := cl.WorkflowService().DescribeSchedule(cctx, &workflowservice.DescribeScheduleRequest{
			Namespace:  c.Parent.Namespace,
			ScheduleId: c.ScheduleId,
		})
		if err != nil {
			return err
		}

		schedule := proto.Clone(describeResponse.Schedule).(*schedpb.Schedule)
		if err := intent.apply(schedule); err != nil {
			return err
		}
		if err := intent.validateResult(schedule); err != nil {
			return err
		}
		_, err = cl.WorkflowService().UpdateSchedule(cctx, &workflowservice.UpdateScheduleRequest{
			Namespace:     c.Parent.Namespace,
			ScheduleId:    c.ScheduleId,
			Schedule:      schedule,
			ConflictToken: describeResponse.ConflictToken,
			Identity:      c.Parent.Identity,
			RequestId:     uuid.NewString(),
		})
		if err == nil {
			break
		}
		conflictErr, ok := err.(*serviceerror.FailedPrecondition)
		if !ok || conflictErr.Message != "mismatched conflict token" || attempt == maxAttempts-1 {
			return err
		}
	}
	if err := cctx.Printer.PrintlnStrictErr("Schedule patch submitted"); err != nil {
		fmt.Fprintln(cctx.Options.Stderr, "Schedule patch may already have been submitted")
		return err
	}
	return nil
}

func (c *TemporalScheduleUpdateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	newSchedule := client.Schedule{
		Spec: &client.ScheduleSpec{},
		Policy: &client.SchedulePolicies{
			CatchupWindow:  c.CatchupWindow.Duration(),
			PauseOnFailure: c.PauseOnFailure,
		},
		State: &client.ScheduleState{
			Note:   c.Notes,
			Paused: c.Paused,
		},
	}

	if newSchedule.Policy.Overlap, err = enumspb.ScheduleOverlapPolicyFromString(c.OverlapPolicy.Value); err != nil {
		return err
	}

	if c.RemainingActions > 0 {
		newSchedule.State.LimitedActions = true
		newSchedule.State.RemainingActions = c.RemainingActions
	}

	if err = c.toScheduleSpec(newSchedule.Spec); err != nil {
		return err
	} else if newSchedule.Action, err = toScheduleAction(&c.SharedWorkflowStartOptions, &c.PayloadInputOptions); err != nil {
		return err
	}

	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)
	return sch.Update(cctx, client.ScheduleUpdateOptions{
		DoUpdate: func(u client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			// replace whole schedule
			return &client.ScheduleUpdate{
				Schedule: &newSchedule,
			}, nil
		},
	})
}

func formatCalendarSpec(spec client.ScheduleCalendarSpec) *schedpb.CalendarSpec {
	processField := func(ranges []client.ScheduleRange) string {
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
	return &schedpb.CalendarSpec{
		Second:     processField(spec.Second),
		Minute:     processField(spec.Minute),
		Hour:       processField(spec.Hour),
		DayOfMonth: processField(spec.DayOfMonth),
		Month:      processField(spec.Month),
		Year:       processField(spec.Year),
		DayOfWeek:  processField(spec.DayOfWeek),
		Comment:    spec.Comment,
	}
}

var reHours = regexp.MustCompile(`\d+h`)
var reLetters = regexp.MustCompile(`[a-z]`)

func formatDuration(d time.Duration) string {
	// Start with time.Duration standard formatting
	s := d.String()
	// Turn "72h" into "3d"
	s = reHours.ReplaceAllStringFunc(s, func(v string) string {
		hours, err := strconv.ParseInt(strings.TrimSuffix(v, "h"), 10, 64)
		if err != nil || hours < 24 {
			return v
		}
		days := hours / 24
		hours -= days * 24
		return fmt.Sprintf("%dd%dh", days, hours)
	})
	// Insert spaces between fields for readability
	s = reLetters.ReplaceAllString(s, "$0 ")
	// Remove last space
	s = strings.TrimSpace(s)
	return s
}

func (c *TemporalScheduleListMatchingTimesCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	res, err := cl.WorkflowService().ListScheduleMatchingTimes(cctx, &workflowservice.ListScheduleMatchingTimesRequest{
		Namespace:  c.Parent.Namespace,
		ScheduleId: c.ScheduleId,
		StartTime:  timestamppb.New(c.StartTime.Time()),
		EndTime:    timestamppb.New(c.EndTime.Time()),
	})
	if err != nil {
		return err
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(res, printer.StructuredOptions{})
	}

	type matchingTime struct {
		Time string `json:"time"`
	}
	var rows []matchingTime
	for _, t := range res.StartTime {
		rows = append(rows, matchingTime{Time: t.AsTime().Format(time.RFC3339)})
	}
	return cctx.Printer.PrintStructured(rows, printer.StructuredOptions{
		Table: &printer.TableOptions{},
	})
}
