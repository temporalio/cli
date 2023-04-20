package trace

import (
	"fmt"
	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/enums/v1"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func GetFoldStatus(c *cli.Context) ([]enums.WorkflowExecutionStatus, error) {
	values := []enums.WorkflowExecutionStatus{}
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

// printWorkflowTrace prints and updates a workflow trace following printWorkflowProgress pattern
func printWorkflowTrace(c *cli.Context, wid, rid string) (int, error) {
	childsDepth := c.Int(common.FlagDepth)
	concurrency := c.Int(common.FlagConcurrency)
	allFlag := c.Bool(common.FlagNoFold)

	foldStatus, err := GetFoldStatus(c)
	if err != nil {
		return 1, err
	}

	sdkClient, err := client.GetSDKClient(c)
	//sdkClient, err := temporal.NewClientFromEnv()
	if err != nil {
		return 1, err
	}

	tcCtx, cancel := common.NewIndefiniteContext(c)
	//tcCtx, cancel := context.WithCancel(c.Context)
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
		iter, err := GetWorkflowExecutionUpdates(tcCtx, sdkClient, wid, rid, allFlag, foldStatus, childsDepth, concurrency)
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

	//var printedRows int
	//var prevString string
	var currentEvents int64
	var totalEvents int64
	var isUpToDate bool
	//fmt.Println(style.Title("Progress:"))
	for {
		select {
		case <-ticker:
			state := update.GetState()
			if state == nil {
				printProgress(currentEvents, totalEvents)
				continue
			}

			if !isUpToDate {
				currentEvents, totalEvents = state.GetNumberOfEvents()
				isUpToDate = currentEvents >= totalEvents && !state.IsArchived
			}

			if isUpToDate {
				//printedRows, prevString = updateWorkflow(state, printedRows, prevString, allFlag, true)
			} else {
				printProgress(currentEvents, totalEvents)
			}
		case <-doneChan:
		case <-sigChan:
			// Print last execution state available, this one doesn't need to be trimmed.
			//updateWorkflow(update.GetState(), printedRows, prevString, allFlag, false)
			return GetExitCode(update.GetState()), nil
		case err = <-errChan:
			//fmt.Println(style.Danger("Error:"), err)
			return 1, err
		}
	}
}

func printProgress(currentEvents int64, totalEvents int64) {
	//if totalEvents == 0 {
	//	fmt.Printf("%sProcessing HistoryEvents (%d)\r", output.CLEAR_LINE, currentEvents)
	//} else {
	//	fmt.Printf("%sProcessing HistoryEvents (%d/%d)\r", output.CLEAR_LINE, currentEvents, totalEvents)
	//}
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
