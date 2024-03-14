package temporalcli

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

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
	sch := cl.ScheduleClient().GetHandle(cctx, c.ScheduleId)
	res, err := sch.Describe(cctx)
	if err != nil {
		return err
	}

	// TODO: print stuff here
	_ = res

	return fmt.Errorf("TODO")
}

type scheduleListEntry struct {
	ID string
	// TODO: more fields here
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
	return &scheduleListEntry{
		ID: next.ID,
		// TODO: copy more fields from next into here
	}, nil
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

	typ := reflect.TypeOf(client.ScheduleListEntry{})
	iter := scheduleListEntryIterAdapter{i: res}

	return cctx.Printer.PrintStructuredIter(typ, iter, printer.StructuredOptions{})
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
