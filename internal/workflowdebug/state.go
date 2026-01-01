package workflowdebug

import (
	"context"
	"encoding/json"
	"fmt"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// StateOptions configures the state extraction.
type StateOptions struct {
	// IncludeDetails includes detailed information about pending items.
	IncludeDetails bool
}

// StateExtractor extracts the current state of a workflow execution.
type StateExtractor struct {
	client client.Client
	opts   StateOptions
}

// NewStateExtractor creates a new StateExtractor.
func NewStateExtractor(c client.Client, opts StateOptions) *StateExtractor {
	return &StateExtractor{
		client: c,
		opts:   opts,
	}
}

// GetState retrieves the current state of a workflow execution.
func (s *StateExtractor) GetState(ctx context.Context, namespace, workflowID, runID string) (*WorkflowStateResult, error) {
	// Describe the workflow execution
	desc, err := s.client.DescribeWorkflowExecution(ctx, workflowID, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to describe workflow: %w", err)
	}

	execInfo := desc.WorkflowExecutionInfo
	if execInfo == nil {
		return nil, fmt.Errorf("workflow execution info is nil")
	}

	result := &WorkflowStateResult{
		Workflow: WorkflowRef{
			Namespace:  namespace,
			WorkflowID: workflowID,
			RunID:      execInfo.Execution.GetRunId(),
		},
		WorkflowType:               execInfo.Type.GetName(),
		Status:                     WorkflowStatusFromEnum(execInfo.Status),
		IsRunning:                  execInfo.Status == enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
		TaskQueue:                  execInfo.GetTaskQueue(),
		HistoryLength:              execInfo.HistoryLength,
		PendingActivityCount:       len(desc.PendingActivities),
		PendingChildWorkflowCount:  len(desc.PendingChildren),
		PendingNexusOperationCount: len(desc.PendingNexusOperations),
	}

	// Set times
	if execInfo.StartTime != nil {
		t := execInfo.StartTime.AsTime()
		result.StartTime = &t
	}
	if execInfo.CloseTime != nil {
		t := execInfo.CloseTime.AsTime()
		result.CloseTime = &t
	}

	// Process pending activities
	for _, pa := range desc.PendingActivities {
		pendingAct := PendingActivity{
			ActivityID:   pa.ActivityId,
			ActivityType: pa.ActivityType.GetName(),
			State:        pendingActivityStateToString(pa.State),
			Attempt:      pa.Attempt,
		}

		if pa.MaximumAttempts > 0 {
			pendingAct.MaxAttempts = pa.MaximumAttempts
		}

		if pa.ScheduledTime != nil {
			t := pa.ScheduledTime.AsTime()
			pendingAct.ScheduledTime = &t
		}

		if pa.LastStartedTime != nil {
			t := pa.LastStartedTime.AsTime()
			pendingAct.LastStartedTime = &t
		}

		if pa.LastFailure != nil && pa.LastFailure.Message != "" {
			pendingAct.LastFailure = pa.LastFailure.Message
		}

		result.PendingActivities = append(result.PendingActivities, pendingAct)
	}

	// Process pending child workflows
	for _, pc := range desc.PendingChildren {
		pendingChild := PendingChildWorkflow{
			WorkflowID:       pc.WorkflowId,
			RunID:            pc.RunId,
			WorkflowType:     pc.WorkflowTypeName,
			InitiatedEventID: pc.InitiatedId,
		}

		if pc.ParentClosePolicy != enums.PARENT_CLOSE_POLICY_UNSPECIFIED {
			pendingChild.ParentClosePolicy = parentClosePolicyToString(pc.ParentClosePolicy)
		}

		result.PendingChildWorkflows = append(result.PendingChildWorkflows, pendingChild)
	}

	// Process pending Nexus operations
	for _, pn := range desc.PendingNexusOperations {
		pendingNexus := PendingNexusOperation{
			Endpoint:         pn.Endpoint,
			Service:          pn.Service,
			Operation:        pn.Operation,
			OperationToken:   pn.OperationToken,
			State:            pendingNexusOperationStateToString(pn.State),
			Attempt:          pn.Attempt,
			ScheduledEventID: pn.ScheduledEventId,
			BlockedReason:    pn.BlockedReason,
		}

		if pn.ScheduledTime != nil {
			t := pn.ScheduledTime.AsTime()
			pendingNexus.ScheduledTime = &t
		}

		if pn.LastAttemptCompleteTime != nil {
			t := pn.LastAttemptCompleteTime.AsTime()
			pendingNexus.LastAttemptCompleteTime = &t
		}

		if pn.NextAttemptScheduleTime != nil {
			t := pn.NextAttemptScheduleTime.AsTime()
			pendingNexus.NextAttemptScheduleTime = &t
		}

		if pn.LastAttemptFailure != nil && pn.LastAttemptFailure.Message != "" {
			pendingNexus.LastFailure = pn.LastAttemptFailure.Message
		}

		if pn.ScheduleToCloseTimeout != nil {
			pendingNexus.ScheduleToCloseTimeoutSec = int64(pn.ScheduleToCloseTimeout.AsDuration().Seconds())
		}

		result.PendingNexusOperations = append(result.PendingNexusOperations, pendingNexus)
	}

	// Process memo if present
	if s.opts.IncludeDetails && execInfo.Memo != nil && len(execInfo.Memo.Fields) > 0 {
		result.Memo = make(map[string]any)
		for k, v := range execInfo.Memo.Fields {
			var val any
			if err := json.Unmarshal(v.Data, &val); err == nil {
				result.Memo[k] = val
			} else {
				result.Memo[k] = string(v.Data)
			}
		}
	}

	// Process search attributes if present
	if s.opts.IncludeDetails && execInfo.SearchAttributes != nil && len(execInfo.SearchAttributes.IndexedFields) > 0 {
		result.SearchAttributes = make(map[string]any)
		for k, v := range execInfo.SearchAttributes.IndexedFields {
			var val any
			if err := json.Unmarshal(v.Data, &val); err == nil {
				result.SearchAttributes[k] = val
			} else {
				result.SearchAttributes[k] = string(v.Data)
			}
		}
	}

	return result, nil
}

// pendingActivityStateToString converts a pending activity state to a string.
func pendingActivityStateToString(state enums.PendingActivityState) string {
	switch state {
	case enums.PENDING_ACTIVITY_STATE_SCHEDULED:
		return "Scheduled"
	case enums.PENDING_ACTIVITY_STATE_STARTED:
		return "Started"
	case enums.PENDING_ACTIVITY_STATE_CANCEL_REQUESTED:
		return "CancelRequested"
	default:
		return "Unknown"
	}
}

// parentClosePolicyToString converts a parent close policy to a string.
func parentClosePolicyToString(policy enums.ParentClosePolicy) string {
	switch policy {
	case enums.PARENT_CLOSE_POLICY_TERMINATE:
		return "Terminate"
	case enums.PARENT_CLOSE_POLICY_ABANDON:
		return "Abandon"
	case enums.PARENT_CLOSE_POLICY_REQUEST_CANCEL:
		return "RequestCancel"
	default:
		return "Unspecified"
	}
}

// pendingNexusOperationStateToString converts a pending Nexus operation state to a string.
func pendingNexusOperationStateToString(state enums.PendingNexusOperationState) string {
	switch state {
	case enums.PENDING_NEXUS_OPERATION_STATE_SCHEDULED:
		return "Scheduled"
	case enums.PENDING_NEXUS_OPERATION_STATE_BACKING_OFF:
		return "BackingOff"
	case enums.PENDING_NEXUS_OPERATION_STATE_STARTED:
		return "Started"
	case enums.PENDING_NEXUS_OPERATION_STATE_BLOCKED:
		return "Blocked"
	default:
		return "Unknown"
	}
}
