package temporalcli

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"github.com/temporalio/cli/temporalcli/internal/trace"
	"go.temporal.io/sdk/client"
)

func (c *TemporalWorkflowTraceCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	ctx := cctx.Context

	opts := trace.WorkflowTraceOptions{
		Depth:       c.FlagDepth,
		Concurrency: c.FlagConcurrency,
		NoFold:      c.FlagNoFold,
	}
	opts.FoldStatus, err = trace.GetFoldStatus(c.FlagFold)
	if err != nil {
		return err
	}

	if err = c.printWorkflowSummary(cctx, cl, c.WorkflowId, c.RunId); err != nil {
		return err
	}
	_, err = trace.PrintWorkflowTrace(ctx, cl, c.WorkflowId, c.RunId, opts)

	return err
}

type workflowSummary struct {
	WorkflowId string `json:"workflowId"`
	RunId      string `json:"runId"`
	Type       string `json:"uype"`
	Namespace  string `json:"namespace"`
	TaskQueue  string `json:"taskQueue"`
}

// printWorkflowSummary prints a summary of the workflow execution, similar to the one available when starting a workflow.
func (c *TemporalWorkflowTraceCommand) printWorkflowSummary(cctx *CommandContext, cl client.Client, wfId, runId string) error {
	res, err := cl.DescribeWorkflowExecution(cctx, wfId, runId)
	if err != nil {
		return err
	}

	info := res.GetWorkflowExecutionInfo()

	cctx.Printer.Println(color.MagentaString("Execution summary:"))

	_ = cctx.Printer.PrintStructured(workflowSummary{
		WorkflowId: info.GetExecution().GetWorkflowId(),
		RunId:      info.GetExecution().GetRunId(),
		Type:       info.GetType().GetName(),
		Namespace:  c.Parent.Namespace,
		TaskQueue:  info.GetTaskQueue(),
	}, printer.StructuredOptions{})
	_, _ = fmt.Println()

	return err
}
