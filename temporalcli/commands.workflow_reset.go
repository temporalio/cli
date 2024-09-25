package temporalcli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/cli/temporalcli/internal/printer"
)

func (c *TemporalWorkflowResetCommand) run(cctx *CommandContext, _ []string) error {
	validateArguments, doReset := c.getResetOperations()
	if err := validateArguments(); err != nil {
		return err
	}
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	return doReset(cctx, cl)
}

func (c *TemporalWorkflowResetCommand) getResetOperations() (validate func() error, doReset func(*CommandContext, client.Client) error) {
	if c.WorkflowId != "" {
		validate = c.validateWorkflowResetArguments
		doReset = c.doWorkflowReset
	} else {
		validate = c.validateBatchResetArguments
		doReset = c.runBatchReset
	}
	return validate, doReset
}

func (c *TemporalWorkflowResetCommand) validateWorkflowResetArguments() error {
	if c.Type.Value == "" && c.EventId <= 0 {
		return errors.New("must specify either valid event id or reset type")
	}
	if c.WorkflowId == "" {
		return errors.New("must specify workflow id")
	}
	return nil
}

func (c *TemporalWorkflowResetCommand) validateBatchResetArguments() error {
	if c.Type.Value == "" {
		return errors.New("must specify reset type")
	}
	if c.RunId != "" {
		return errors.New("must not specify run Id")
	}
	if c.EventId != 0 {
		return errors.New("must not specify event Id")
	}
	if c.Type.Value == "BuildId" && c.BuildId == "" {
		return errors.New("must specify build Id for BuildId based batch reset")
	}
	return nil
}
func (c *TemporalWorkflowResetCommand) doWorkflowReset(cctx *CommandContext, cl client.Client) error {

	var err error
	resetBaseRunID := c.RunId
	eventID := int64(c.EventId)
	if c.Type.Value != "" {
		resetBaseRunID, eventID, err = c.getResetEventIDByType(cctx, cl)
		if err != nil {
			return fmt.Errorf("getting reset event ID by type failed: %w", err)
		}
	}

	reapplyExcludes := make([]enums.ResetReapplyExcludeType, 0)
	for _, exclude := range c.ReapplyExclude.Values {
		if strings.ToLower(exclude) == "all" {
			for _, excludeType := range enums.ResetReapplyExcludeType_value {
				if excludeType == 0 {
					continue
				}
				reapplyExcludes = append(reapplyExcludes, enums.ResetReapplyExcludeType(excludeType))
			}
			break
		}
		excludeType, err := enums.ResetReapplyExcludeTypeFromString(exclude)
		if err != nil {
			return err
		}
		reapplyExcludes = append(reapplyExcludes, excludeType)
	}

	reapplyType := enums.RESET_REAPPLY_TYPE_SIGNAL
	if c.ReapplyType.Value != "All" {
		if len(c.ReapplyExclude.Values) > 0 {
			return errors.New("cannot specify --reapply-type and --reapply-exclude at the same time")
		}
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
		ResetReapplyExcludeTypes:  reapplyExcludes,
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

func (c *TemporalWorkflowResetCommand) runBatchReset(cctx *CommandContext, cl client.Client) error {
	request := workflowservice.StartBatchOperationRequest{
		Namespace:       c.Parent.Namespace,
		JobId:           uuid.NewString(),
		VisibilityQuery: c.Query,
		Reason:          c.Reason,
	}
	request.Operation = &workflowservice.StartBatchOperationRequest_ResetOperation{
		ResetOperation: &batch.BatchOperationReset{
			Identity: clientIdentity(),
			Options:  c.batchResetOptions(c.Type.Value),
		},
	}
	count, err := cl.CountWorkflow(cctx, &workflowservice.CountWorkflowExecutionsRequest{Query: c.Query})
	if err != nil {
		return fmt.Errorf("failed counting workflows from query: %w", err)
	}
	yes, err := cctx.promptYes(
		fmt.Sprintf("Start batch against approximately %v workflow(s)? y/N", count.Count), c.Yes)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("user denied confirmation")
	}

	return startBatchJob(cctx, cl, &request)
}

func (c *TemporalWorkflowResetCommand) batchResetOptions(resetType string) *common.ResetOptions {
	switch resetType {
	case "FirstWorkflowTask":
		return &common.ResetOptions{
			Target: &common.ResetOptions_FirstWorkflowTask{},
		}
	case "LastWorkflowTask":
		return &common.ResetOptions{
			Target: &common.ResetOptions_LastWorkflowTask{},
		}
	case "BuildId":
		return &common.ResetOptions{
			Target: &common.ResetOptions_BuildId{
				BuildId: c.BuildId,
			},
		}
	default:
		panic("unsupported operation type was filtered by cli framework")
	}
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
