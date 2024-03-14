package temporalcli

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	enumspb "go.temporal.io/api/enums/v1"
	schedpb "go.temporal.io/api/schedule/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
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
}

func describeResultToPrintable(id string, desc *client.ScheduleDescription) *printableSchedule {
	// TODO: should we include any other fields here, e.g. jitter, time zone, start/end time
	out := &printableSchedule{
		ScheduleId: id,
		Paused:     desc.Schedule.State.Paused,
		Notes:      desc.Schedule.State.Note,
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

func listResultToPrintable(ent *client.ScheduleListEntry) *printableSchedule {
	out := &printableSchedule{
		ScheduleId:   ent.ID,
		Paused:       ent.Paused,
		Notes:        ent.Note,
		WorkflowType: ent.WorkflowType,
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

	startTime, err := time.Parse(time.RFC3339, c.StartTime)
	if err != nil {
		return err
	}
	endTime, err := time.Parse(time.RFC3339, c.EndTime)
	if err != nil {
		return err
	}
	overlap, err := enumspb.ScheduleOverlapPolicyFromString(c.OverlapPolicy.Value)
	if err != nil {
		return err
	}

	err = sch.Backfill(cctx, client.ScheduleBackfillOptions{
		Backfill: []client.ScheduleBackfill{
			{
				Start:   startTime,
				End:     endTime,
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

func (c *TemporalScheduleCreateCommand) run(cctx *CommandContext, args []string) error {
	return fmt.Errorf("TODO")
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

type scheduleListEntryIterAdapter struct {
	i client.ScheduleListIterator
}

func (i scheduleListEntryIterAdapter) Next() (any, error) {
	if !i.i.HasNext() {
		return nil, nil
	}
	next, err := i.i.Next()
	if err != nil {
		return nil, err
	}
	return listResultToPrintable(next), nil
}

func (c *TemporalScheduleListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// // This is a listing command subject to json vs jsonl rules
	// cctx.Printer.StartList()
	// defer cctx.Printer.EndList()

	res, err := cl.ScheduleClient().List(cctx, client.ScheduleListOptions{})
	if err != nil {
		return err
	}

	typ := reflect.TypeOf(client.ScheduleListEntry{})
	iter := scheduleListEntryIterAdapter{i: res}

	fields := []string{
		"ScheduleId",
		"CalendarSpecs",
		"IntervalSpecs",
		"WorkflowType",
		"Paused",
		"Notes",
		"NextRunTime",
		"LastRunTime",
	}
	return cctx.Printer.PrintStructuredIter(typ, iter, printer.StructuredOptions{Fields: fields})
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
