package temporalcli

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/sdk/client"
)

func (c *TemporalTaskQueueGetBuildIdsCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	sets, err := cl.GetWorkerBuildIdCompatibility(cctx, &client.GetWorkerBuildIdCompatibilityOptions{
		TaskQueue: c.TaskQueue,
		MaxSets:   c.MaxSets,
	})
	if err != nil {
		return fmt.Errorf("unable to get task queue build ids: %w", err)
	}

	type rowtype struct {
		BuildIds      []string
		DefaultForSet string
		IsDefaultSet  bool
	}
	var items []rowtype
	for ix, s := range sets.Sets {
		row := rowtype{
			BuildIds:      s.BuildIDs,
			IsDefaultSet:  ix == len(sets.Sets)-1,
			DefaultForSet: s.BuildIDs[len(s.BuildIDs)-1],
		}
		items = append(items, row)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(items, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Version Sets:"))
	return cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

// Missing in the SDK?
func taskReachabilityToString(x client.TaskReachability) string {
	switch x {
	case client.TaskReachabilityUnspecified:
		return "Unspecified"
	case client.TaskReachabilityNewWorkflows:
		return "NewWorkflows"
	case client.TaskReachabilityExistingWorkflows:
		return "ExistingWorkflows"
	case client.TaskReachabilityOpenWorkflows:
		return "OpenWorkflows"
	case client.TaskReachabilityClosedWorkflows:
		return "ClosedWorkflows"
	default:
		return strconv.Itoa(int(x))
	}
}

func (c *TemporalTaskQueueGetBuildIdReachabilityCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	reachability := client.TaskReachabilityUnspecified
	if c.ReachabilityType.Value != "" {
		if c.ReachabilityType.Value == "open" {
			reachability = client.TaskReachabilityOpenWorkflows
		} else if c.ReachabilityType.Value == "closed" {
			reachability = client.TaskReachabilityClosedWorkflows
		} else if c.ReachabilityType.Value == "existing" {
			reachability = client.TaskReachabilityExistingWorkflows
		} else {
			return fmt.Errorf("invalid reachability type: %v", c.ReachabilityType)
		}
	}

	reach, err := cl.GetWorkerTaskReachability(cctx, &client.GetWorkerTaskReachabilityOptions{
		BuildIDs:     c.BuildId,
		TaskQueues:   c.TaskQueue,
		Reachability: client.TaskReachability(reachability),
	})
	if err != nil {
		return fmt.Errorf("unable to get Build ID reachability: %w", err)
	}

	type rowtype struct {
		BuildId      string
		TaskQueue    string
		Reachability []string
	}

	var items []rowtype

	for bid, e := range reach.BuildIDReachability {
		for tq, r := range e.TaskQueueReachable {
			reachability := make([]string, len(r.TaskQueueReachability))
			for i, v := range r.TaskQueueReachability {
				reachability[i] = taskReachabilityToString(v)
			}
			row := rowtype{
				BuildId:      bid,
				TaskQueue:    tq,
				Reachability: reachability,
			}
			items = append(items, row)
		}
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(items, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Reachability:"))
	return cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})
}
