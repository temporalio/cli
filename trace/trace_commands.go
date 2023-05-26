package trace

import (
	"fmt"
	"os"
	"os/signal"
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

	doneChan := make(chan bool)
	errChan := make(chan error)
	ticker := time.NewTicker(time.Second).C

	// Capture interrupt signals to do a last print
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start update fetching
	var update *WorkflowExecutionUpdate
	go func() {
		iter, err := GetWorkflowExecutionUpdates(tcCtx, sdkClient, wid, rid, noFold, foldStatus, childWfsDepth, concurrency)
		if err != nil {
			errChan <- err
			return
		}
		for iter.HasNext() {
			if update, err = iter.Next(); err != nil {
				errChan <- err
			}
		}
		doneChan <- true
	}()

	var currentEvents int64
	var totalEvents int64
	var isUpToDate bool

	// Load templates
	writer := NewTermWriter().WithTerminalSize()
	tmpl, err := NewExecutionTemplate(writer, foldStatus, noFold)
	if err != nil {
		return 1, err
	}

	_, _ = title.Println("Progress:")
	for {
		select {
		case <-ticker:
			state := update.GetState()
			if state == nil {
				_, _ = writer.WriteString(ProgressString(currentEvents, totalEvents))
				writer.Flush(true)
				continue
			}

			if !isUpToDate {
				currentEvents, totalEvents = state.GetNumberOfEvents()
				isUpToDate = totalEvents > 0 && currentEvents >= totalEvents && !state.IsArchived
			}

			if isUpToDate {
				if err := tmpl.Execute(update.GetState(), 0); err != nil {
					_, _ = writer.WriteString(fmt.Sprintf("%s %s", color.RedString("Error:"), err.Error()))
				}
			} else {
				_, _ = writer.WriteString(ProgressString(currentEvents, totalEvents))
			}
			if err := writer.Flush(true); err != nil {
				return 1, err
			}
		case <-doneChan:
			return PrintAndExit(writer, tmpl, update)
		case <-sigChan:
			return PrintAndExit(writer, tmpl, update)
		case err = <-errChan:
			return 1, err
		}
	}
}

func ProgressString(currentEvents int64, totalEvents int64) string {
	if totalEvents == 0 {
		if currentEvents == 0 {
			return "Processing HistoryEvents"
		}
		return fmt.Sprintf("Processing HistoryEvents (%d)", currentEvents)
	} else {
		return fmt.Sprintf("Processing HistoryEvents (%d/%d)", currentEvents, totalEvents)
	}
}

func PrintAndExit(writer *TermWriter, tmpl *ExecutionTemplate, update *WorkflowExecutionUpdate) (int, error) {
	if err := tmpl.Execute(update.GetState(), 0); err != nil {
		return 1, err
	}
	if err := writer.Flush(false); err != nil {
		return 1, err
	}
	return GetExitCode(update.GetState()), nil
}

// GetExitCode returns the exit code for a given workflow execution status.
func GetExitCode(exec *WorkflowExecutionState) int {
	if exec == nil {
		// Don't panic if the state is missing.
		return 0
	}
	switch exec.Status {
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return 2
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return 3
	case enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED:
		return 4
	}
	return 0
}
