package temporalcli

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/temporalio/cli/temporalcli/internal/tracer"

	"go.temporal.io/api/enums/v1"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/sdk/client"
)

var foldFlag = map[string]enums.WorkflowExecutionStatus{
	"running":       enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
	"completed":     enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
	"failed":        enums.WORKFLOW_EXECUTION_STATUS_FAILED,
	"canceled":      enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
	"terminated":    enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
	"timedout":      enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
	"continueasnew": enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
}

func getFoldStatuses(foldFlags []string) ([]enums.WorkflowExecutionStatus, error) {
	// defaults
	if len(foldFlags) == 0 {
		return []enums.WorkflowExecutionStatus{
			enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
			enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
		}, nil
	}

	// parse flags
	var values []enums.WorkflowExecutionStatus
	for _, flag := range foldFlags {
		status, ok := foldFlag[flag]
		if !ok {
			return nil, fmt.Errorf("fold status \"%s\" not recognized", flag)
		}
		values = append(values, status)
	}
	return values, nil
}

func (c *TemporalWorkflowTraceCommand) run(cctx *CommandContext, _ []string) error {
	if cctx.JSONOutput {
		return fmt.Errorf("JSON output not supported for trace command")
	}
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	opts := tracer.WorkflowTracerOptions{
		Depth:       c.Depth,
		Concurrency: c.Concurrency,
		NoFold:      c.NoFold,
	}
	opts.FoldStatuses, err = getFoldStatuses(c.Fold)
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
func (c *TemporalWorkflowTraceCommand) printWorkflowTrace(cctx *CommandContext, cl client.Client, wid, rid string, opts tracer.WorkflowTracerOptions) (int, error) {
	// Load templates
	tmpl, err := tracer.NewExecutionTemplate(opts.FoldStatuses, opts.NoFold)
	if err != nil {
		return 1, err
	}

	workflowTracer, err := tracer.NewWorkflowTracer(cl,
		tracer.WithOptions(opts),
		tracer.WithOutput(cctx.Printer.Output),
		tracer.WithInterrupts(os.Interrupt, syscall.SIGTERM, syscall.SIGINT),
	)
	if err != nil {
		return 1, err
	}

	err = workflowTracer.GetExecutionUpdates(cctx, wid, rid)
	if err != nil {
		return 1, err
	}

	cctx.Printer.Println(color.MagentaString("Progress:"))
	return workflowTracer.PrintUpdates(tmpl, time.Second)
}
