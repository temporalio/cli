package temporalcli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

func (c *TemporalOperatorClusterHealthCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	_, err = cl.CheckHealth(cctx, &client.CheckHealthRequest{})
	if err != nil {
		return fmt.Errorf("failed checking cluster health: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			struct {
				Status string `json:"status"`
			}{"SERVING"},
			printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString("SERVING"))
	return nil
}

func (c *TemporalOperatorClusterSystemCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	resp, err := cl.WorkflowService().GetSystemInfo(cctx, &workflowservice.GetSystemInfoRequest{})
	if err != nil {
		return fmt.Errorf("unable to get system information: %w", err)
	}
	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	return cctx.Printer.PrintStructured(
		struct {
			ServerVersion      string
			SupportsSchedules  bool
			UpsertMemo         bool
			EagerWorkflowStart bool
		}{
			ServerVersion:      resp.GetServerVersion(),
			SupportsSchedules:  resp.GetCapabilities().SupportsSchedules,
			UpsertMemo:         resp.GetCapabilities().UpsertMemo,
			EagerWorkflowStart: resp.GetCapabilities().EagerWorkflowStart,
		},
		printer.StructuredOptions{
			Table: &printer.TableOptions{},
		})
}

func (c *TemporalOperatorClusterDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	resp, err := cl.WorkflowService().GetClusterInfo(cctx, &workflowservice.GetClusterInfoRequest{})
	if err != nil {
		return fmt.Errorf("unable to get cluster information: %w", err)
	}
	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	fields := []string{"ClusterName", "PersistenceStore", "VisibilityStore"}
	if c.Detail {
		fields = append(fields, "HistoryShardCount", "VersionInfo")
	}

	return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{
		Fields: fields,
		Table:  &printer.TableOptions{},
	})
}
