package workflowdebug

import (
	"testing"

	"go.temporal.io/api/enums/v1"
)

func TestWorkflowStatusFromEnum(t *testing.T) {
	tests := []struct {
		input    enums.WorkflowExecutionStatus
		expected string
	}{
		{enums.WORKFLOW_EXECUTION_STATUS_RUNNING, "Running"},
		{enums.WORKFLOW_EXECUTION_STATUS_COMPLETED, "Completed"},
		{enums.WORKFLOW_EXECUTION_STATUS_FAILED, "Failed"},
		{enums.WORKFLOW_EXECUTION_STATUS_CANCELED, "Canceled"},
		{enums.WORKFLOW_EXECUTION_STATUS_TERMINATED, "Terminated"},
		{enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW, "ContinuedAsNew"},
		{enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT, "TimedOut"},
		{enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := WorkflowStatusFromEnum(tt.input)
			if result != tt.expected {
				t.Errorf("WorkflowStatusFromEnum(%v) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseWorkflowStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected enums.WorkflowExecutionStatus
	}{
		{"Running", enums.WORKFLOW_EXECUTION_STATUS_RUNNING},
		{"Completed", enums.WORKFLOW_EXECUTION_STATUS_COMPLETED},
		{"Failed", enums.WORKFLOW_EXECUTION_STATUS_FAILED},
		{"Canceled", enums.WORKFLOW_EXECUTION_STATUS_CANCELED},
		{"Terminated", enums.WORKFLOW_EXECUTION_STATUS_TERMINATED},
		{"ContinuedAsNew", enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW},
		{"TimedOut", enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT},
		{"Unknown", enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED},
		{"invalid", enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseWorkflowStatus(tt.input)
			if result != tt.expected {
				t.Errorf("ParseWorkflowStatus(%s) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWorkflowStatusRoundTrip(t *testing.T) {
	statuses := []enums.WorkflowExecutionStatus{
		enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
		enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
		enums.WORKFLOW_EXECUTION_STATUS_FAILED,
		enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
		enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
		enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
		enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
	}

	for _, status := range statuses {
		str := WorkflowStatusFromEnum(status)
		parsed := ParseWorkflowStatus(str)
		if parsed != status {
			t.Errorf("Round trip failed for %v: got %v", status, parsed)
		}
	}
}

