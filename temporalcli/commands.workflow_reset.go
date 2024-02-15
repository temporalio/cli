package temporalcli

import (
	"context"
	"errors"
	"fmt"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

var (
	resetTypesMap = map[string]interface{}{
		"FirstWorkflowTask":  "",
		"LastWorkflowTask":   "",
		"LastContinuedAsNew": "",
	}
)

func (c *TemporalWorkflowResetCommand) run(cctx *CommandContext, _ []string) error {
	if c.Type.Value == "" && c.EventId <= 0 {
		return errors.New("must specify either valid event id or reset type")
	}
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	resetBaseRunID := c.RunId
	eventID := int64(c.EventId)
	if c.Type.Value != "" {
		resetBaseRunID, eventID, err = c.getResetEventIDByType(cctx, cl)
		if err != nil {
			return fmt.Errorf("getting reset event ID by type failed: %w", err)
		}
	}
	reapplyType := enums.RESET_REAPPLY_TYPE_SIGNAL
	if c.ReapplyType.Value != "All" {
		reapplyType, err = enums.ResetReapplyTypeFromString(c.ReapplyType.Value)
		if err != nil {
			return err
		}
	}

	cctx.Printer.Printlnf("Resetting workflow %s to event ID %d", c.WorkflowId, eventID)

	resp, err := cl.ResetWorkflowExecution(cctx, &workflowservice.ResetWorkflowExecutionRequest{
		Namespace: c.Parent.Namespace,
		WorkflowExecution: &common.WorkflowExecution{
			WorkflowId: c.WorkflowId,
			RunId:      resetBaseRunID,
		},
		Reason:                    fmt.Sprintf("%s: %s", username(), c.Reason),
		WorkflowTaskFinishEventId: eventID,
		ResetReapplyType:          reapplyType,
	})
	if err != nil {
		return fmt.Errorf("failed to reset workflow: %w", err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			resp,
			printer.StructuredOptions{})
	}
	return nil
}

func (c *TemporalWorkflowResetCommand) getResetEventIDByType(ctx context.Context, cl client.Client) (string, int64, error) {
	resetType, namespace, wid, rid := c.Type.Value, c.Parent.Namespace, c.WorkflowId, c.RunId
	wfsvc := cl.WorkflowService()
	switch resetType {
	case "LastWorkflowTask":
		return getLastWorkflowTaskEventID(ctx, namespace, wid, rid, wfsvc)
	case "LastContinuedAsNew":
		return getLastContinueAsNewID(ctx, namespace, wid, rid, wfsvc)
	case "FirstWorkflowTask":
		return getFirstWorkflowTaskEventID(ctx, namespace, wid, rid, wfsvc)
	default:
		return "", -1, fmt.Errorf("invalid reset type: %s", resetType)
	}
}

// Returns event id of the last completed task or id of the next event after scheduled task.
func getLastWorkflowTaskEventID(ctx context.Context, namespace, wid, rid string, wfsvc workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskEventID int64, err error) {
	resetBaseRunID = rid
	req := workflowservice.GetWorkflowExecutionHistoryReverseRequest{
		Namespace: namespace,
		Execution: &common.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		MaximumPageSize: 250,
		NextPageToken:   nil,
	}

	for more := true; more; more = len(req.NextPageToken) != 0 {
		resp, err := wfsvc.GetWorkflowExecutionHistoryReverse(ctx, &req)
		if err != nil {
			return "", 0, fmt.Errorf("failed to get workflow execution history: %w", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskEventID = e.GetEventId()
				break
			} else if e.GetEventType() == enums.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED {
				// if there is no task completed event, set it to first scheduled event + 1
				workflowTaskEventID = e.GetEventId() + 1
			}
		}
		req.NextPageToken = resp.NextPageToken
	}
	if workflowTaskEventID == 0 {
		return "", 0, errors.New("unable to find any scheduled or completed task")
	}
	return
}

// Returns id of the first workflow task completed event or if it doesn't exist then id of the event after task scheduled event.
func getFirstWorkflowTaskEventID(ctx context.Context, namespace, wid, rid string, wfsvc workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskEventID int64, err error) {
	resetBaseRunID = rid
	req := workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: namespace,
		Execution: &common.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		MaximumPageSize: 250,
		NextPageToken:   nil,
	}
	for more := true; more; more = len(req.NextPageToken) != 0 {
		resp, err := wfsvc.GetWorkflowExecutionHistory(ctx, &req)
		if err != nil {
			return "", 0, fmt.Errorf("failed to get workflow execution history: %w", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskEventID = e.GetEventId()
				return resetBaseRunID, workflowTaskEventID, nil
			}
			if e.GetEventType() == enums.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED {
				if workflowTaskEventID == 0 {
					workflowTaskEventID = e.GetEventId() + 1
				}
			}
		}
		req.NextPageToken = resp.NextPageToken
	}
	if workflowTaskEventID == 0 {
		return "", 0, errors.New("unable to find any scheduled or completed task")
	}
	return
}

func getLastContinueAsNewID(ctx context.Context, namespace, wid, rid string, wfsvc workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskCompletedID int64, err error) {
	// get first event
	req := &workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: namespace,
		Execution: &common.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		MaximumPageSize: 1,
		NextPageToken:   nil,
	}
	resp, err := wfsvc.GetWorkflowExecutionHistory(ctx, req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get workflow execution history: %w", err)
	}
	firstEvent := resp.History.Events[0]
	resetBaseRunID = firstEvent.GetWorkflowExecutionStartedEventAttributes().GetContinuedExecutionRunId()
	if resetBaseRunID == "" {
		return "", 0, errors.New("cannot use LastContinuedAsNew for workflow; workflow was not continued from another")
	}

	req = &workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: namespace,
		Execution: &common.WorkflowExecution{
			WorkflowId: wid,
			RunId:      resetBaseRunID,
		},
		MaximumPageSize: 250,
		NextPageToken:   nil,
	}
	for more := true; more; more = len(req.NextPageToken) != 0 {
		resp, err := wfsvc.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return "", 0, fmt.Errorf("failed to get workflow execution history of previous execution (run id %s): %w", resetBaseRunID, err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskCompletedID = e.GetEventId()
			}
		}
		req.NextPageToken = resp.NextPageToken
	}
	if workflowTaskCompletedID == 0 {
		return "", 0, errors.New("unable to find WorkflowTaskCompleted event for previous execution")
	}
	return
}
