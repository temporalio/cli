package temporalcli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	enumspb "go.temporal.io/api/enums/v1"
	schedpb "go.temporal.io/api/schedule/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/primitives/timestamp"
	"google.golang.org/protobuf/encoding/protojson"
)

type printableSchedule struct {
	ScheduleId       any
	CalendarSpecs    []any `cli:",cardOmitEmpty"`
	IntervalSpecs    []any `cli:",cardOmitEmpty"`
	WorkflowId       any
	WorkflowType     any
	Paused           any
	Notes            any `cli:",cardOmitEmpty"`
	NextRunTime      any
	LastRunTime      any
	RunningWorkflows []any
	SearchAttributes any `cli:",cardOmitEmpty"`
	Memo             any `cli:",cardOmitEmpty"`
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
		// TODO: This is not quite right for either json or text:
		// For json: this doesn't come out as proper protojson, the field names are
		// day_of_month instead of dayOfMonth.
		// For text: this looks fine when there's only one, but there are no {} around each
		// element, so if there's more than one it's confusing.
		out.CalendarSpecs = append(out.CalendarSpecs, formatCalendarSpec(cal))
	}
	for _, int := range spec.Intervals {
		if int.Offset > 0 {
			out.IntervalSpecs = append(out.IntervalSpecs, int)
		} else {
			// hide offset if not present
			out.IntervalSpecs = append(out.IntervalSpecs, struct{ Every time.Duration }{Every: int.Every})
		}
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
		return spec, errors.New("invalid interval string")
	} else if len(parts) == 2 {
		if spec.Offset, err = timestamp.ParseDuration(parts[1]); err != nil {
			return spec, fmt.Errorf("invalid interval string: %w", err)
		}
	}
	if spec.Every, err = timestamp.ParseDuration(parts[0]); err != nil {
		return spec, fmt.Errorf("invalid interval string: %w", err)
	}
	return spec, nil
}

func (c *ScheduleConfigurationOptions) toScheduleSpec() (client.ScheduleSpec, error) {
	spec := client.ScheduleSpec{
		CronExpressions: c.Cron,
		// Skip not supported
		Jitter:       c.Jitter,
		TimeZoneName: c.TimeZone,
		StartAt:      c.StartTime.Time(),
		EndAt:        c.EndTime.Time(),
	}

	var err error
	for _, calPbStr := range c.Calendar {
		var calPb schedpb.CalendarSpec
		if err = protojson.Unmarshal([]byte(calPbStr), &calPb); err != nil {
			return spec, fmt.Errorf("failed to parse json calendar spec: %w", err)
		}
		cron, err := toCronString(&calPb)
		if err != nil {
			return spec, err
		}
		spec.CronExpressions = append(spec.CronExpressions, cron)
	}
	for _, intStr := range c.Interval {
		int, err := toIntervalSpec(intStr)
		if err != nil {
			return spec, err
		}
		spec.Intervals = append(spec.Intervals, int)
	}

	return spec, nil
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

	if opts.Spec, err = c.toScheduleSpec(); err != nil {
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

	if c.Raw {
		res, err := cl.WorkflowService().DescribeSchedule(cctx, &workflowservice.DescribeScheduleRequest{
			Namespace:  c.Parent.Namespace,
			ScheduleId: c.ScheduleId,
		})
		if err != nil {
			return err
		}
		// force JSON output
		cctx.Printer.JSON = true
		cctx.Printer.JSONIndent = "  "
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
	return fmt.Errorf("TODO")
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
