package temporalcli

import (
	"math"
	"strings"
	"testing"

	"go.temporal.io/sdk/client"
)

func TestToScheduleActionAppliesPriorityKey(t *testing.T) {
	action, err := toScheduleAction(&SharedWorkflowStartOptions{
		PriorityKey: 42,
	}, &PayloadInputOptions{})
	if err != nil {
		t.Fatalf("toScheduleAction returned an unexpected error: %v", err)
	}

	scheduleAction, ok := action.(*client.ScheduleWorkflowAction)
	if !ok {
		t.Fatalf("toScheduleAction returned %T, want *client.ScheduleWorkflowAction", action)
	}
	if scheduleAction.Priority.PriorityKey != 42 {
		t.Errorf("PriorityKey = %d, want 42", scheduleAction.Priority.PriorityKey)
	}
}

func TestToScheduleActionAppliesFairnessFieldsWithoutPriorityKey(t *testing.T) {
	action, err := toScheduleAction(&SharedWorkflowStartOptions{
		FairnessKey:    "tenant-a",
		FairnessWeight: 2.5,
	}, &PayloadInputOptions{})
	if err != nil {
		t.Fatalf("toScheduleAction returned an unexpected error: %v", err)
	}

	scheduleAction, ok := action.(*client.ScheduleWorkflowAction)
	if !ok {
		t.Fatalf("toScheduleAction returned %T, want *client.ScheduleWorkflowAction", action)
	}
	if scheduleAction.Priority.PriorityKey != 0 {
		t.Errorf("PriorityKey = %d, want 0", scheduleAction.Priority.PriorityKey)
	}
	if scheduleAction.Priority.FairnessKey != "tenant-a" {
		t.Errorf("FairnessKey = %q, want %q", scheduleAction.Priority.FairnessKey, "tenant-a")
	}
	if scheduleAction.Priority.FairnessWeight != 2.5 {
		t.Errorf("FairnessWeight = %v, want 2.5", scheduleAction.Priority.FairnessWeight)
	}
}

func TestToScheduleActionForwardsServerPolicyValuesAndValidatesPriorityRepresentation(t *testing.T) {
	testCases := []struct {
		name    string
		options SharedWorkflowStartOptions
		wantErr bool
	}{
		{name: "default priority and fairness", options: SharedWorkflowStartOptions{}},
		{name: "minimum priority", options: SharedWorkflowStartOptions{PriorityKey: math.MinInt32}},
		{name: "negative priority", options: SharedWorkflowStartOptions{PriorityKey: -1}},
		{name: "server configured priority", options: SharedWorkflowStartOptions{PriorityKey: 6}},
		{name: "maximum representable priority", options: SharedWorkflowStartOptions{PriorityKey: math.MaxInt32}},
		{name: "priority below int32 minimum", options: SharedWorkflowStartOptions{PriorityKey: math.MinInt32 - 1}, wantErr: true},
		{name: "priority above int32 maximum", options: SharedWorkflowStartOptions{PriorityKey: math.MaxInt32 + 1}, wantErr: true},
		{name: "empty fairness key and zero weight", options: SharedWorkflowStartOptions{}},
		{name: "fairness key longer than 64 bytes", options: SharedWorkflowStartOptions{FairnessKey: strings.Repeat("a", 65)}},
		{name: "negative fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: -1}},
		{name: "fairness weight below prior minimum", options: SharedWorkflowStartOptions{FairnessWeight: 0.0009}},
		{name: "fairness weight above prior maximum", options: SharedWorkflowStartOptions{FairnessWeight: 1000.1}},
		{name: "NaN fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: float32(math.NaN())}},
		{name: "positive infinite fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: float32(math.Inf(1))}},
		{name: "negative infinite fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: float32(math.Inf(-1))}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			action, err := toScheduleAction(&testCase.options, &PayloadInputOptions{})
			if (err != nil) != testCase.wantErr {
				t.Fatalf("toScheduleAction error = %v, want error = %v", err, testCase.wantErr)
			}
			if testCase.wantErr {
				return
			}
			scheduleAction, ok := action.(*client.ScheduleWorkflowAction)
			if !ok {
				t.Fatalf("toScheduleAction returned %T, want *client.ScheduleWorkflowAction", action)
			}
			if scheduleAction.Priority.PriorityKey != testCase.options.PriorityKey {
				t.Errorf("PriorityKey = %d, want %d", scheduleAction.Priority.PriorityKey, testCase.options.PriorityKey)
			}
			if scheduleAction.Priority.FairnessKey != testCase.options.FairnessKey {
				t.Errorf("FairnessKey = %q, want %q", scheduleAction.Priority.FairnessKey, testCase.options.FairnessKey)
			}
			if math.IsNaN(float64(testCase.options.FairnessWeight)) {
				if !math.IsNaN(float64(scheduleAction.Priority.FairnessWeight)) {
					t.Errorf("FairnessWeight = %v, want NaN", scheduleAction.Priority.FairnessWeight)
				}
				return
			}
			if scheduleAction.Priority.FairnessWeight != testCase.options.FairnessWeight {
				t.Errorf("FairnessWeight = %v, want %v", scheduleAction.Priority.FairnessWeight, testCase.options.FairnessWeight)
			}
		})
	}
}
