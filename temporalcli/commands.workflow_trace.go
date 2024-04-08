package temporalcli

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"github.com/temporalio/cli/temporalcli/internal/trace"
	"go.temporal.io/sdk/client"
)

func (c *TemporalWorkflowTraceCommand) run(cctx *CommandContext, _ []string) error {
	if cctx.JSONOutput {
		return fmt.Errorf("JSON output not supported for trace command")
	}
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	opts := trace.WorkflowTraceOptions{
		Depth:       c.Depth,
		Concurrency: c.Concurrency,
		NoFold:      c.NoFold,
	}
	opts.FoldStatus, err = trace.GetFoldStatus(c.Fold)
	if err != nil {
		return err
	}

	if err = c.printWorkflowSummary(cctx, cl, c.WorkflowId, c.RunId); err != nil {
		return err
	}
	_, err = c.printWorkflowTrace(cctx, cl, c.WorkflowId, c.RunId, opts)

	return err
}

type workflowTraceSummary struct {
	WorkflowId string `json:"workflowId"`
	RunId      string `json:"runId"`
	Type       string `json:"type"`
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

	_ = cctx.Printer.PrintStructured(workflowTraceSummary{
		WorkflowId: info.GetExecution().GetWorkflowId(),
		RunId:      info.GetExecution().GetRunId(),
		Type:       info.GetType().GetName(),
		Namespace:  c.Parent.Namespace,
		TaskQueue:  info.GetTaskQueue(),
	}, printer.StructuredOptions{})
	cctx.Printer.Println()

	return err
}

// PrintWorkflowTrace prints and updates a workflow trace following printWorkflowProgress pattern
func (c *TemporalWorkflowTraceCommand) printWorkflowTrace(cctx *CommandContext, cl client.Client, wid, rid string, opts trace.WorkflowTraceOptions) (int, error) {
	// Load templates
	tmpl, err := trace.NewExecutionTemplate(opts.FoldStatus, opts.NoFold)
	if err != nil {
		return 1, err
	}

	tracer, err := trace.NewWorkflowTracer(cl,
		trace.WithOptions(opts),
		trace.WithOutput(cctx.Printer.Output),
		trace.WithInterrupts(os.Interrupt, syscall.SIGTERM, syscall.SIGINT),
	)
	if err != nil {
		return 1, err
	}

	err = tracer.GetExecutionUpdates(cctx, wid, rid)
	if err != nil {
		return 1, err
	}

	cctx.Printer.Println(color.MagentaString("Progress:"))
	return tracer.PrintUpdates(tmpl, time.Second)
}
