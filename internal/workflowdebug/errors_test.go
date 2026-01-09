package workflowdebug

import (
	"testing"
)

func TestCompactError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple error",
			input:    "database connection refused",
			expected: "database connection refused",
		},
		{
			name:     "workflow failed prefix",
			input:    "WorkflowFailed: database connection refused",
			expected: "database connection refused",
		},
		{
			name:     "nested child workflow error",
			input:    "WorkflowFailed: child workflow at depth 0 failed: child workflow execution error (type: NestedFailureWorkflow, workflowID: nested-level-1, runID: abc123, initiatedEventID: 5, startedEventID: 6): child workflow at depth 1 failed: child workflow execution error (type: NestedFailureWorkflow, workflowID: nested-level-2, runID: def456, initiatedEventID: 5, startedEventID: 6): activity error (type: FailingActivity, scheduledEventID: 5, startedEventID: 6, identity: worker@host): critical failure at depth 3: database connection refused",
			expected: "database connection refused",
		},
		{
			name:     "activity error with metadata",
			input:    "activity error (type: ProcessPaymentActivity, scheduledEventID: 5, startedEventID: 6, identity: 74744@host): payment gateway connection timeout",
			expected: "payment gateway connection timeout",
		},
		{
			name:     "activity name prefix",
			input:    "ProcessPaymentActivity - payment gateway connection timeout",
			expected: "payment gateway connection timeout",
		},
		{
			name:     "activity failed prefix with name",
			input:    "ActivityFailed: ProcessPaymentActivity - payment gateway connection timeout",
			expected: "payment gateway connection timeout",
		},
		{
			name:     "order workflow chain",
			input:    "WorkflowFailed: order failed: payment error: child workflow execution error (type: PaymentWorkflow, workflowID: payment-123, runID: abc): activity error (type: ProcessPaymentActivity, scheduledEventID: 5): payment gateway timeout",
			expected: "payment gateway timeout",
		},
		{
			name:     "validation error",
			input:    "WorkflowFailed: validation failed: child workflow execution error (type: ValidationWorkflow, workflowID: validation-123, runID: abc): activity error (type: ValidationActivity, scheduledEventID: 8): validation failed: order contains invalid product SKU 'INVALID-123'",
			expected: "order contains invalid product SKU 'INVALID-123'",
		},
		{
			name:     "timeout error",
			input:    "Timeout: workflow execution timed out",
			expected: "workflow execution timed out",
		},
		{
			name:     "retry exhaustion",
			input:    "WorkflowFailed: retry exhaustion: all 5 attempts failed: activity error (type: AlwaysFailsActivity, scheduledEventID: 8): transient error: service unavailable",
			expected: "service unavailable",
		},
		{
			name:     "error with retryable annotation",
			input:    "critical failure (type: wrapError, retryable: true): database connection refused",
			expected: "database connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompactError(tt.input)
			if result != tt.expected {
				t.Errorf("CompactError(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompactErrorWithContext(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "workflow execution timeout",
			input:    "WorkflowFailed: workflow execution timed out",
			expected: "workflow timeout",
		},
		{
			name:     "activity timeout",
			input:    "activity error (type: SlowActivity, scheduledEventID: 8): activity StartToClose timeout",
			expected: "activity StartToClose timeout",
		},
		{
			name:     "regular error",
			input:    "database connection refused",
			expected: "database connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompactErrorWithContext(tt.input)
			if result != tt.expected {
				t.Errorf("CompactErrorWithContext(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
