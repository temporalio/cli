package trace

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	sdkclient "go.temporal.io/sdk/client"

	"github.com/fatih/color"
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/enums/v1"
)

var (
	title = color.New(color.FgMagenta)
)

func GetFoldStatus(c *cli.Context) ([]enums.WorkflowExecutionStatus, error) {
	var values []enums.WorkflowExecutionStatus
	flagFold := c.String(common.FlagFold)
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

// PrintWorkflowTrace prints and updates a workflow trace following printWorkflowProgress pattern
func PrintWorkflowTrace(c *cli.Context, sdkClient sdkclient.Client, wid, rid string, foldStatus []enums.WorkflowExecutionStatus) (int, error) {
	childWfsDepth := c.Int(common.FlagDepth)
	concurrency := c.Int(common.FlagConcurrency)
	noFold := c.Bool(common.FlagNoFold)

	tcCtx, cancel := common.NewIndefiniteContext(c)
	defer cancel()

	// Load templates
	tmpl, err := NewExecutionTemplate(foldStatus, noFold)
	if err != nil {
		return 1, err
	}

	opts := WorkflowTracerOptions{
		NoFold:        noFold,
		FoldStatus:    foldStatus,
		ChildWfsDepth: childWfsDepth,
		Concurrency:   concurrency,
		UpdatePeriod:  time.Second,
	}
	tracer, err := NewWorkflowTracer(sdkClient,
		WithOptions(opts),
		WithInterrupts(os.Interrupt, syscall.SIGTERM, syscall.SIGINT),
	)
	if err != nil {
		return 1, err
	}

	err = tracer.GetExecutionUpdates(tcCtx, wid, rid)
	if err != nil {
		return 1, err
	}

	_, _ = title.Println("Progress:")
	return tracer.PrintUpdates(tmpl, time.Second)
}
