package trace

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	sdkclient "go.temporal.io/sdk/client"

	"github.com/fatih/color"
	"go.temporal.io/api/enums/v1"
)

var (
	title = color.New(color.FgMagenta)
)

// TODO: Remove if it's not longer being used
func GetFoldStatus(flagFold string) ([]enums.WorkflowExecutionStatus, error) {
	var values []enums.WorkflowExecutionStatus
	for _, v := range strings.Split(flagFold, ",") {
		var status enums.WorkflowExecutionStatus
		switch strings.ToLower(v) {
		case "running":
			status = enums.WORKFLOW_EXECUTION_STATUS_RUNNING
		case "completed":
			status = enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
		case "failed":
			status = enums.WORKFLOW_EXECUTION_STATUS_FAILED
		case "canceled":
			status = enums.WORKFLOW_EXECUTION_STATUS_CANCELED
		case "terminated":
			status = enums.WORKFLOW_EXECUTION_STATUS_TERMINATED
		case "timedout":
			status = enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT
		case "continueasnew":
			status = enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW
		default:
			return nil, fmt.Errorf("fold status \"%s\" not recognized", v)
		}

		values = append(values, status)
	}
	return values, nil
}

type WorkflowTraceOptions struct {
	NoFold      bool
	FoldStatus  []enums.WorkflowExecutionStatus
	Depth       int
	Concurrency int

	UpdatePeriod time.Duration
}

// PrintWorkflowTrace prints and updates a workflow trace following printWorkflowProgress pattern
func PrintWorkflowTrace(ctx context.Context, sdkClient sdkclient.Client, wid, rid string, opts WorkflowTraceOptions) (int, error) {
	// Load templates
	tmpl, err := NewExecutionTemplate(opts.FoldStatus, opts.NoFold)
	if err != nil {
		return 1, err
	}

	tracer, err := NewWorkflowTracer(sdkClient,
		WithOptions(opts),
		WithInterrupts(os.Interrupt, syscall.SIGTERM, syscall.SIGINT),
		// WithWriter(NewTermWriter().WithTerminalSize()),
	)
	if err != nil {
		return 1, err
	}

	err = tracer.GetExecutionUpdates(ctx, wid, rid)
	if err != nil {
		return 1, err
	}

	_, _ = title.Println("Progress:")
	return tracer.PrintUpdates(tmpl, time.Second)
}
