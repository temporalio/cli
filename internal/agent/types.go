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
	Failures []FailureReport `json:"failures"`
	// TotalCount is the total number of failures matching the query (may be more than returned).
	TotalCount int `json:"total_count,omitempty"`
	// Query is the visibility query used.
	Query string `json:"query,omitempty"`
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

