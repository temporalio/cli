package temporalcli

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	schedpb "go.temporal.io/api/schedule/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/primitives/timestamp"
	"google.golang.org/protobuf/encoding/protojson"
)

type printableSchedule struct {
	ScheduleId       string
	CalendarSpecs    []any  `cli:",cardOmitEmpty"`
	IntervalSpecs    []any  `cli:",cardOmitEmpty"`
	WorkflowId       string // describe only
	WorkflowType     any
	Paused           bool
	Notes            string `cli:",cardOmitEmpty"`
	NextRunTime      time.Time
	LastRunTime      time.Time
	RunningWorkflows []string                   // describe only
	SearchAttributes *commonpb.SearchAttributes `cli:",cardOmitEmpty"`
	Memo             *commonpb.Memo             `cli:",cardOmitEmpty"`
}

// Neither protojson nor fmt print structs containing time.Durations nicely, so do it manually
// using a struct of strings.
type printableInterval struct {
	Every  string `json:"every"`
	Offset string `json:"offset,omitempty"`
}

func describeResultToPrintable(id string, desc *client.ScheduleDescription) *printableSchedule {
	// TODO: should we include any other fields here, e.g. jitter, time zone, start/end time
	out := &printableSchedule{
		ScheduleId:       id,
		Paused:           desc.Schedule.State.Paused,
		Notes:            desc.Schedule.State.Note,
		SearchAttributes: desc.SearchAttributes,
		Memo:             desc.Memo,
	}
	specToPrintable(out, desc.Schedule.Spec)
	if workflowAction, ok := desc.Schedule.Action.(*client.ScheduleWorkflowAction); ok {
		out.WorkflowId = workflowAction.ID
		out.WorkflowType = workflowAction.Workflow
	}
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
	return out
}

func listEntryToPrintable(ent *client.ScheduleListEntry) *printableSchedule {
	out := &printableSchedule{
		ScheduleId:       ent.ID,
		Paused:           ent.Paused,
		Notes:            ent.Note,
		WorkflowType:     ent.WorkflowType.Name,
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
		out.CalendarSpecs = append(out.CalendarSpecs, formatCalendarSpec(cal))
	}
	for _, int := range spec.Intervals {
		pInt := printableInterval{Every: formatDuration(int.Every)}
		if int.Offset > 0 {
			pInt.Offset = formatDuration(int.Offset)
		}
		out.IntervalSpecs = append(out.IntervalSpecs, pInt)
	}
}

func (c *TemporalScheduleBackfillCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
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
		if spec.Offset, err = timestamp.ParseDuration(parts[1]); err != nil {
			return spec, fmt.Errorf("invalid interval: %w", err)
		}
	}
	if spec.Every, err = timestamp.ParseDuration(parts[0]); err != nil {
		return spec, fmt.Errorf("invalid interval: %w", err)
	}
	return spec, nil
}

func (c *ScheduleConfigurationOptions) toScheduleSpec(spec *client.ScheduleSpec) error {
	spec.CronExpressions = c.Cron
	// Skip not supported
	spec.Jitter = c.Jitter
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
	opts, err := buildStartOptions(sw, &WorkflowStartOptions{})
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
		SearchAttributes: opts.SearchAttributes,
		Memo:             opts.Memo,
	}
	if action.Args, err = i.buildRawInput(); err != nil {
		return action, nil
	}
	return action, nil
}

func (c *TemporalScheduleCreateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	opts := client.ScheduleOptions{
		ID:               c.ScheduleId,
		PauseOnFailure:   c.PauseOnFailure,
		Note:             c.Notes,
		Paused:           c.Paused,
		CatchupWindow:    c.CatchupWindow,
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
	} else if opts.Memo, err = stringKeysJSONValues(c.ScheduleMemo); err != nil {
		return fmt.Errorf("invalid memo values: %w", err)
	} else if opts.SearchAttributes, err = stringKeysJSONValues(c.ScheduleSearchAttribute); err != nil {
		return fmt.Errorf("invalid search attribute values: %w", err)
	}

	_, err = cl.ScheduleClient().Create(cctx, opts)
	return err
}

func (c *TemporalScheduleDeleteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
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
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	if cctx.JSONOutput {
		res, err := cl.WorkflowService().DescribeSchedule(cctx, &workflowservice.DescribeScheduleRequest{
			Namespace:  c.Parent.Namespace,
			ScheduleId: c.ScheduleId,
		})
		if err != nil {
			return err
		}
		cctx.Printer.PrintStructured(res, printer.StructuredOptions{})
		return nil
	}

	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)
	res, err := sch.Describe(cctx)
	if err != nil {
		return err
	}

	printable := describeResultToPrintable(c.ScheduleId, res)
	return cctx.Printer.PrintStructured(printable, printer.StructuredOptions{})
}

func (c *TemporalScheduleListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	res, err := cl.ScheduleClient().List(cctx, client.ScheduleListOptions{})
	if err != nil {
		return err
	}

	// This is a listing command subject to json vs jsonl rules
	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	printOpts := printer.StructuredOptions{
		ExcludeFields: []string{
			// These aren't available in list results
			"WorkflowId",
			"RunningWorkflows",
		},
		Table: &printer.TableOptions{},
	}

	if !c.Long {
		printOpts.ExcludeFields = append(printOpts.ExcludeFields,
			"CalendarSpecs",
			"IntervalSpecs",
			"Notes",
			"SearchAttributes",
			"Memo",
		)
	}

	// make artificial "pages" so we get better aligned columns
	page := make([]*printableSchedule, 0, 100)

	for res.HasNext() {
		ent, err := res.Next()
		if err != nil {
			return err
		}
		printable := listEntryToPrintable(ent)
		if cctx.JSONOutput {
			cctx.Printer.PrintStructured(printable, printOpts)
		} else {
			page = append(page, printable)
			if len(page) == cap(page) {
				cctx.Printer.PrintStructured(page, printOpts)
				page = page[:0]
				printOpts.Table.NoHeader = true
			}
		}
	}
	if !cctx.JSONOutput {
		cctx.Printer.PrintStructured(page, printOpts)
	}

	return nil
}

func (c *TemporalScheduleToggleCommand) run(cctx *CommandContext, args []string) error {
	if c.Pause == c.Unpause {
		return errors.New("exactly one of --pause or --unpause is required")
	}

	cl, err := c.Parent.ClientOptions.dialClient(cctx)
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
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
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

func (c *TemporalScheduleUpdateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	newSchedule := client.Schedule{
		Spec: &client.ScheduleSpec{},
		Policy: &client.SchedulePolicies{
			CatchupWindow:  c.CatchupWindow,
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
