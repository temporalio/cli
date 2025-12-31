// Package agent provides structured, agent-optimized views of Temporal workflow
// execution data, designed for AI agents and automated tooling.
package agent

import (
	"time"

	"go.temporal.io/api/enums/v1"
)

// WorkflowRef identifies a workflow execution.
type WorkflowRef struct {
	Namespace  string `json:"namespace"`
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id,omitempty"`
}

// WorkflowChainNode represents a single workflow in a chain of executions.
type WorkflowChainNode struct {
	Namespace    string `json:"namespace"`
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id,omitempty"`
	WorkflowType string `json:"workflow_type,omitempty"`
	Status       string `json:"status"`
	// IsLeaf is true if this is the deepest workflow in the failure chain.
	IsLeaf bool `json:"leaf,omitempty"`
	// Depth indicates how deep this workflow is in the chain (0 = root).
	Depth int `json:"depth"`
	// StartTime when the workflow started.
	StartTime *time.Time `json:"start_time,omitempty"`
	// CloseTime when the workflow closed (if closed).
	CloseTime *time.Time `json:"close_time,omitempty"`
	// Duration of the workflow execution.
	DurationMs int64 `json:"duration_ms,omitempty"`
	// Error message if the workflow failed.
	Error string `json:"error,omitempty"`
}

// RootCause contains information about the root cause of a failure.
type RootCause struct {
	// Type is the type of failure (e.g., "ActivityFailed", "WorkflowFailed", "Timeout").
	Type string `json:"type"`
	// Activity is the name of the failed activity (if applicable).
	Activity string `json:"activity,omitempty"`
	// Error is the error message.
	Error string `json:"error"`
	// Timestamp when the failure occurred.
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Workflow where the root cause was found.
	Workflow *WorkflowRef `json:"workflow,omitempty"`
}

// TraceResult is the output of the trace command.
type TraceResult struct {
	// Chain is the ordered list of workflows from root to leaf.
	Chain []WorkflowChainNode `json:"chain"`
	// RootCause contains the deepest failure information.
	RootCause *RootCause `json:"root_cause,omitempty"`
	// Depth is the maximum depth reached during traversal.
	Depth int `json:"depth"`
}

// FailureReport represents a single failure with its chain and root cause.
type FailureReport struct {
	// RootWorkflow is the top-level workflow that started the chain.
	RootWorkflow WorkflowRef `json:"root_workflow"`
	// LeafFailure is the deepest workflow where the failure originated.
	LeafFailure *WorkflowRef `json:"leaf_failure,omitempty"`
	// Depth is how deep the failure chain goes.
	Depth int `json:"depth"`
	// RootCause is a human-readable summary of the failure.
	RootCause string `json:"root_cause"`
	// Chain is the list of workflow IDs from root to leaf.
	Chain []string `json:"chain"`
	// Timestamp when the failure was detected.
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Status of the root workflow.
	Status string `json:"status"`
}

// FailuresResult is the output of the failures command.
type FailuresResult struct {
	// Failures is the list of failure reports (always present, may be empty).
	Failures []FailureReport `json:"failures"`
	// Groups contains aggregated failure groups (when --group-by is used).
	Groups []FailureGroup `json:"groups,omitempty"`
	// TotalCount is the total number of failures matching the query.
	TotalCount int `json:"total_count"`
	// Query is the visibility query used.
	Query string `json:"query,omitempty"`
	// GroupedBy indicates what field failures were grouped by (if any).
	GroupedBy string `json:"grouped_by,omitempty"`
}

// FailureGroup represents a group of failures aggregated by a common field.
type FailureGroup struct {
	// Key is the grouping key value.
	Key string `json:"key"`
	// Count is the number of failures in this group.
	Count int `json:"count"`
	// Percentage of total failures.
	Percentage float64 `json:"percentage"`
	// Sample is a sample failure from this group.
	Sample *FailureReport `json:"sample,omitempty"`
	// FirstSeen is the timestamp of the earliest failure in this group.
	FirstSeen *time.Time `json:"first_seen,omitempty"`
	// LastSeen is the timestamp of the most recent failure in this group.
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

// TimelineEvent represents a single event in a workflow timeline.
type TimelineEvent struct {
	// Timestamp when the event occurred.
	Timestamp time.Time `json:"ts"`
	// EventID is the sequential event ID.
	EventID int64 `json:"event_id"`
	// Type is the event type (e.g., "WorkflowExecutionStarted", "ActivityTaskScheduled").
	Type string `json:"type"`
	// Category groups related events (workflow, activity, timer, child_workflow, signal, etc.).
	Category string `json:"category,omitempty"`
	// Name is the name of the related entity (activity type, timer name, child workflow type, etc.).
	Name string `json:"name,omitempty"`
	// Status for events that represent a state change.
	Status string `json:"status,omitempty"`
	// ActivityID for activity events.
	ActivityID string `json:"activity_id,omitempty"`
	// Error for failure events.
	Error string `json:"error,omitempty"`
	// Result for completed events (if include-payloads is set).
	Result any `json:"result,omitempty"`
	// Input for start events (if include-payloads is set).
	Input any `json:"input,omitempty"`
	// DurationMs for events that represent a completed operation.
	DurationMs int64 `json:"duration_ms,omitempty"`
	// Attempt number for retried operations.
	Attempt int32 `json:"attempt,omitempty"`
	// ChildWorkflow reference for child workflow events.
	ChildWorkflow *WorkflowRef `json:"child_workflow,omitempty"`
	// RetryCount for compacted retry events.
	RetryCount int `json:"retry_count,omitempty"`
}

// TimelineResult is the output of the timeline command.
type TimelineResult struct {
	// Workflow identifies the workflow this timeline is for.
	Workflow WorkflowRef `json:"workflow"`
	// WorkflowType is the type name of the workflow.
	WorkflowType string `json:"workflow_type,omitempty"`
	// Status is the current status of the workflow.
	Status string `json:"status"`
	// StartTime when the workflow started.
	StartTime *time.Time `json:"start_time,omitempty"`
	// CloseTime when the workflow closed (if closed).
	CloseTime *time.Time `json:"close_time,omitempty"`
	// DurationMs of the workflow execution.
	DurationMs int64 `json:"duration_ms,omitempty"`
	// Events is the list of timeline events.
	Events []TimelineEvent `json:"events"`
	// EventCount is the total number of events (may differ from len(Events) if filtered).
	EventCount int `json:"event_count"`
}

// PendingActivity represents an activity that is currently pending execution.
type PendingActivity struct {
	// ActivityID is the unique ID of the activity.
	ActivityID string `json:"activity_id"`
	// ActivityType is the type name of the activity.
	ActivityType string `json:"activity_type"`
	// State is the current state (SCHEDULED, STARTED, CANCEL_REQUESTED).
	State string `json:"state"`
	// Attempt is the current attempt number (1-based).
	Attempt int32 `json:"attempt"`
	// MaxAttempts is the maximum number of attempts (0 = unlimited).
	MaxAttempts int32 `json:"max_attempts,omitempty"`
	// ScheduledTime when the activity was scheduled.
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
	// LastStartedTime when the activity was last started.
	LastStartedTime *time.Time `json:"last_started_time,omitempty"`
	// HeartbeatTimeout is the heartbeat timeout duration.
	HeartbeatTimeoutSec int64 `json:"heartbeat_timeout_sec,omitempty"`
	// LastHeartbeatTime when the last heartbeat was received.
	LastHeartbeatTime *time.Time `json:"last_heartbeat_time,omitempty"`
	// LastFailure contains the last failure message if retrying.
	LastFailure string `json:"last_failure,omitempty"`
	// Input data (if include-details is set).
	Input any `json:"input,omitempty"`
}

// PendingChildWorkflow represents a child workflow that is currently pending.
type PendingChildWorkflow struct {
	// WorkflowID of the child workflow.
	WorkflowID string `json:"workflow_id"`
	// RunID of the child workflow.
	RunID string `json:"run_id,omitempty"`
	// WorkflowType is the type name of the child workflow.
	WorkflowType string `json:"workflow_type"`
	// InitiatedEventID is the event ID that initiated this child workflow.
	InitiatedEventID int64 `json:"initiated_event_id"`
	// ParentClosePolicy determines what happens when parent closes.
	ParentClosePolicy string `json:"parent_close_policy,omitempty"`
}

// PendingSignal represents an external signal that hasn't been delivered yet.
// Note: Temporal doesn't expose pending signals directly; this is for future use.
type PendingSignal struct {
	// SignalName is the name of the signal.
	SignalName string `json:"signal_name"`
}

// PendingNexusOperation represents a Nexus operation that is currently pending.
type PendingNexusOperation struct {
	// Endpoint is the Nexus endpoint name.
	Endpoint string `json:"endpoint"`
	// Service is the Nexus service name.
	Service string `json:"service"`
	// Operation is the Nexus operation name.
	Operation string `json:"operation"`
	// OperationToken is the async operation token (for async operations).
	OperationToken string `json:"operation_token,omitempty"`
	// State is the current state (Scheduled, Started, BackingOff, Blocked).
	State string `json:"state"`
	// Attempt is the current attempt number.
	Attempt int32 `json:"attempt"`
	// ScheduledTime when the operation was scheduled.
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
	// ScheduledEventID is the event ID of the NexusOperationScheduled event.
	ScheduledEventID int64 `json:"scheduled_event_id,omitempty"`
	// LastAttemptCompleteTime when the last attempt completed.
	LastAttemptCompleteTime *time.Time `json:"last_attempt_complete_time,omitempty"`
	// NextAttemptScheduleTime when the next attempt is scheduled.
	NextAttemptScheduleTime *time.Time `json:"next_attempt_schedule_time,omitempty"`
	// LastFailure contains the last attempt's failure message if retrying.
	LastFailure string `json:"last_failure,omitempty"`
	// BlockedReason provides additional information if the operation is blocked.
	BlockedReason string `json:"blocked_reason,omitempty"`
	// ScheduleToCloseTimeoutSec is the timeout for the operation in seconds.
	ScheduleToCloseTimeoutSec int64 `json:"schedule_to_close_timeout_sec,omitempty"`
}

// WorkflowStateResult is the output of the state command.
type WorkflowStateResult struct {
	// Workflow identifies the workflow.
	Workflow WorkflowRef `json:"workflow"`
	// WorkflowType is the type name of the workflow.
	WorkflowType string `json:"workflow_type,omitempty"`
	// Status is the current status of the workflow.
	Status string `json:"status"`
	// StartTime when the workflow started.
	StartTime *time.Time `json:"start_time,omitempty"`
	// CloseTime when the workflow closed (if closed).
	CloseTime *time.Time `json:"close_time,omitempty"`
	// IsRunning is true if the workflow is currently running.
	IsRunning bool `json:"is_running"`
	// PendingActivities is the list of currently pending activities.
	PendingActivities []PendingActivity `json:"pending_activities,omitempty"`
	// PendingActivityCount is the count of pending activities.
	PendingActivityCount int `json:"pending_activity_count"`
	// PendingChildWorkflows is the list of currently pending child workflows.
	PendingChildWorkflows []PendingChildWorkflow `json:"pending_child_workflows,omitempty"`
	// PendingChildWorkflowCount is the count of pending child workflows.
	PendingChildWorkflowCount int `json:"pending_child_workflow_count"`
	// PendingNexusOperations is the list of currently pending Nexus operations.
	PendingNexusOperations []PendingNexusOperation `json:"pending_nexus_operations,omitempty"`
	// PendingNexusOperationCount is the count of pending Nexus operations.
	PendingNexusOperationCount int `json:"pending_nexus_operation_count"`
	// TaskQueue is the task queue the workflow is running on.
	TaskQueue string `json:"task_queue,omitempty"`
	// HistoryLength is the number of events in the workflow history.
	HistoryLength int64 `json:"history_length,omitempty"`
	// Memo contains workflow memo data.
	Memo map[string]any `json:"memo,omitempty"`
	// SearchAttributes contains indexed search attributes.
	SearchAttributes map[string]any `json:"search_attributes,omitempty"`
}

// WorkflowStatusFromEnum converts an enums.WorkflowExecutionStatus to a string.
func WorkflowStatusFromEnum(status enums.WorkflowExecutionStatus) string {
	switch status {
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return "Running"
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return "Completed"
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return "Failed"
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return "Canceled"
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return "Terminated"
	case enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return "ContinuedAsNew"
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return "TimedOut"
	default:
		return "Unknown"
	}
}

// ParseWorkflowStatus converts a string status to an enums.WorkflowExecutionStatus.
func ParseWorkflowStatus(s string) enums.WorkflowExecutionStatus {
	switch s {
	case "Running":
		return enums.WORKFLOW_EXECUTION_STATUS_RUNNING
	case "Completed":
		return enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
	case "Failed":
		return enums.WORKFLOW_EXECUTION_STATUS_FAILED
	case "Canceled":
		return enums.WORKFLOW_EXECUTION_STATUS_CANCELED
	case "Terminated":
		return enums.WORKFLOW_EXECUTION_STATUS_TERMINATED
	case "ContinuedAsNew":
		return enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW
	case "TimedOut":
		return enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT
	default:
		return enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED
	}
}

