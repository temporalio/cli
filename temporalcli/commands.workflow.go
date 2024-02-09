package temporalcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/user"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/query/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

var (
	resetTypesMap = map[string]interface{}{
		"FirstWorkflowTask":  "",
		"LastWorkflowTask":   "",
		"LastContinuedAsNew": "",
	}
	resetReapplyTypesMap = map[string]interface{}{
		"":       enums.RESET_REAPPLY_TYPE_SIGNAL, // default value
		"Signal": enums.RESET_REAPPLY_TYPE_SIGNAL,
		"None":   enums.RESET_REAPPLY_TYPE_NONE,
	}
)

func (c *TemporalWorkflowCancelCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})
	if err != nil {
		return err
	}

	// Run single or batch
	if exec != nil {
		err = cl.CancelWorkflow(cctx, exec.WorkflowId, exec.RunId)
		if err != nil {
			return fmt.Errorf("failed to cancel workflow: %w", err)
		}
		cctx.Printer.Println("Canceled workflow")
	} else if batchReq != nil {
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_CancellationOperation{
			CancellationOperation: &batch.BatchOperationCancellation{
				Identity: clientIdentity(),
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}

func (*TemporalWorkflowDeleteCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (c *TemporalWorkflowQueryCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Get input payloads
	input, err := c.buildRawInputPayloads()
	if err != nil {
		return err
	}

	queryRejectCond := enums.QUERY_REJECT_CONDITION_UNSPECIFIED
	switch c.RejectCondition.Value {
	case "":
	case "not_open":
		queryRejectCond = enums.QUERY_REJECT_CONDITION_NOT_OPEN
	case "not_completed_cleanly":
		queryRejectCond = enums.QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
	default:
		return fmt.Errorf("invalid query reject condition: %v, valid values are: 'not_open', 'not_completed_cleanly'", c.RejectCondition)
	}

	result, err := cl.WorkflowService().QueryWorkflow(cctx, &workflowservice.QueryWorkflowRequest{
		Namespace: c.Parent.Namespace,
		Execution: &common.WorkflowExecution{WorkflowId: c.WorkflowId, RunId: c.RunId},
		Query: &query.WorkflowQuery{
			QueryType: c.Type,
			QueryArgs: input,
		},
		QueryRejectCondition: queryRejectCond,
	})

	if err != nil {
		return fmt.Errorf("querying workflow failed: %w", err)
	}

	if result.QueryRejected != nil {
		return fmt.Errorf("query was rejected, workflow has status: %v\n", result.QueryRejected.GetStatus())
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Query result:"))
	output := struct {
		QueryResult json.RawMessage `cli:",cardOmitEmpty"`
	}{}
	output.QueryResult, err = cctx.MarshalFriendlyJSONPayloads(result.QueryResult)
	if err != nil {
		return fmt.Errorf("failed to marshal query result: %w", err)
	}
	return cctx.Printer.PrintStructured(output, printer.StructuredOptions{})
}

func (c *TemporalWorkflowResetCommand) run(cctx *CommandContext, _ []string) error {
	if c.Type.Value == "" && c.EventId <= 0 {
		return errors.New("must specify either valid event id or reset type")
	}
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	conn, err := c.Parent.ClientOptions.dialGRPC(cctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	wfsvc := workflowservice.NewWorkflowServiceClient(conn)
	resetBaseRunID := c.RunId
	eventID := int64(c.EventId)
	if c.Type.Value != "" {
		resetBaseRunID, eventID, err = getResetEventIDByType(cctx, c.Type.Value, c.Parent.Namespace, c.WorkflowId, c.RunId, wfsvc)
		if err != nil {
			return fmt.Errorf("getting reset event ID by type failed: %w", err)
		}
	}
	username := "<unknown-user>"
	if u, err := user.Current(); err != nil && u.Username != "" {
		username = u.Username
	}
	reapplyType := enums.RESET_REAPPLY_TYPE_SIGNAL
	if c.ReapplyType.Value != "All" {
		reapplyType, err = enums.ResetReapplyTypeFromString(c.ReapplyType.Value)
		if err != nil {
			return err
		}
	}

	if eventID > 0 {
		cctx.Printer.Printlnf("Resetting workflow %s to event ID %d", c.WorkflowId, eventID)
	}

	resp, err := cl.ResetWorkflowExecution(cctx, &workflowservice.ResetWorkflowExecutionRequest{
		Namespace: c.Parent.Namespace,
		WorkflowExecution: &common.WorkflowExecution{
			WorkflowId: c.WorkflowId,
			RunId:      resetBaseRunID,
		},
		Reason:                    fmt.Sprintf("%s: %s", username, c.Reason),
		WorkflowTaskFinishEventId: eventID,
		RequestId:                 uuid.NewString(),
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

func (*TemporalWorkflowResetBatchCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (c *TemporalWorkflowSignalCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Get input payloads
	input, err := c.buildRawInputPayloads()
	if err != nil {
		return err
	}

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})
	if err != nil {
		return err
	}

	// Run single or batch
	if exec != nil {
		// We have to use the raw signal service call here because the Go SDK's
		// signal call doesn't accept multiple arguments.
		_, err = cl.WorkflowService().SignalWorkflowExecution(cctx, &workflowservice.SignalWorkflowExecutionRequest{
			Namespace:         c.Parent.Namespace,
			WorkflowExecution: &common.WorkflowExecution{WorkflowId: c.WorkflowId, RunId: c.RunId},
			SignalName:        c.Name,
			Input:             input,
			Identity:          clientIdentity(),
		})
		if err != nil {
			return fmt.Errorf("failed signalling workflow: %w", err)
		}
		cctx.Printer.Println("Signal workflow succeeded")
	} else if batchReq != nil {
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_SignalOperation{
			SignalOperation: &batch.BatchOperationSignal{
				Signal:   c.Name,
				Input:    input,
				Identity: clientIdentity(),
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}
	return nil
}

func (*TemporalWorkflowStackCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (c *TemporalWorkflowTerminateCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// We create a faux SingleWorkflowOrBatchOptions to use the shared logic
	opts := SingleWorkflowOrBatchOptions{
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Query:      c.Query,
		Reason:     c.Reason,
		Yes:        c.Yes,
	}

	exec, batchReq, err := opts.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{
		// You're allowed to specify a reason when terminating a workflow
		AllowReasonWithWorkflowID: true,
	})
	if err != nil {
		return err
	}

	// Run single or batch
	if exec != nil {
		reason := c.Reason
		if reason == "" {
			reason = defaultReason()
		}
		err = cl.TerminateWorkflow(cctx, exec.WorkflowId, exec.RunId, reason)
		if err != nil {
			return fmt.Errorf("failed to terminate workflow: %w", err)
		}
		cctx.Printer.Println("Workflow terminated")
	} else if batchReq != nil {
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_TerminationOperation{
			TerminationOperation: &batch.BatchOperationTermination{
				Identity: clientIdentity(),
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}

func (*TemporalWorkflowTraceCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (*TemporalWorkflowUpdateCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func defaultReason() string {
	username := "<unknown-user>"
	if u, err := user.Current(); err != nil && u.Username != "" {
		username = u.Username
	}
	return "Requested from CLI by " + username
}

type singleOrBatchOverrides struct {
	AllowReasonWithWorkflowID bool
}

func (s *SingleWorkflowOrBatchOptions) workflowExecOrBatch(
	cctx *CommandContext,
	namespace string,
	cl client.Client,
	overrides singleOrBatchOverrides,
) (*common.WorkflowExecution, *workflowservice.StartBatchOperationRequest, error) {
	// If workflow is set, we return single execution
	if s.WorkflowId != "" {
		if s.Query != "" {
			return nil, nil, fmt.Errorf("cannot set query when workflow ID is set")
		} else if s.Reason != "" && !overrides.AllowReasonWithWorkflowID {
			return nil, nil, fmt.Errorf("cannot set reason when workflow ID is set")
		} else if s.Yes {
			return nil, nil, fmt.Errorf("cannot set 'yes' when workflow ID is set")
		}
		return &common.WorkflowExecution{WorkflowId: s.WorkflowId, RunId: s.RunId}, nil, nil
	}

	// Check query is set properly
	if s.Query == "" {
		return nil, nil, fmt.Errorf("must set either workflow ID or query")
	} else if s.WorkflowId != "" {
		return nil, nil, fmt.Errorf("cannot set workflow ID when query is set")
	} else if s.RunId != "" {
		return nil, nil, fmt.Errorf("cannot set run ID when query is set")
	}

	// Count the workflows that will be affected
	count, err := cl.CountWorkflow(cctx, &workflowservice.CountWorkflowExecutionsRequest{Query: s.Query})
	if err != nil {
		return nil, nil, fmt.Errorf("failed counting workflows from query: %w", err)
	}
	yes, err := cctx.promptYes(
		fmt.Sprintf("Start batch against approximately %v workflow(s)? y/N", count.Count), s.Yes)
	if err != nil {
		return nil, nil, err
	} else if !yes {
		// We consider this a command failure
		return nil, nil, fmt.Errorf("user denied confirmation")
	}

	// Default the reason if not set
	reason := s.Reason
	if reason == "" {
		reason = defaultReason()
	}

	return nil, &workflowservice.StartBatchOperationRequest{
		Namespace:       namespace,
		JobId:           uuid.NewString(),
		VisibilityQuery: s.Query,
		Reason:          reason,
	}, nil
}

func startBatchJob(cctx *CommandContext, cl client.Client, req *workflowservice.StartBatchOperationRequest) error {
	_, err := cl.WorkflowService().StartBatchOperation(cctx, req)
	if err != nil {
		return fmt.Errorf("failed starting batch operation: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			struct {
				BatchJobID string `json:"batchJobId"`
			}{BatchJobID: req.JobId},
			printer.StructuredOptions{})
	}
	cctx.Printer.Printlnf("Started batch for job ID: %v", req.JobId)
	return nil
}

func getResetEventIDByType(ctx context.Context, resetType, namespace, wid, rid string, wfsvc workflowservice.WorkflowServiceClient) (string, int64, error) {
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

	for {
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
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
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
	for {
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
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
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
	for {
		resp, err := wfsvc.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return "", 0, fmt.Errorf("failed to get workflow execution history of original execution (run id %s): %w", resetBaseRunID, err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskCompletedID = e.GetEventId()
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if workflowTaskCompletedID == 0 {
		return "", 0, errors.New("unable to find WorkflowTaskCompleted event for original workflow")
	}
	return
}
