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

func TestToScheduleActionValidatesPriorityAndFairness(t *testing.T) {
	testCases := []struct {
		name    string
		options SharedWorkflowStartOptions
		wantErr bool
	}{
		{name: "default priority and fairness", options: SharedWorkflowStartOptions{}},
		{name: "minimum priority", options: SharedWorkflowStartOptions{PriorityKey: 1}},
		{name: "server configured priority", options: SharedWorkflowStartOptions{PriorityKey: 6}},
		{name: "maximum representable priority", options: SharedWorkflowStartOptions{PriorityKey: math.MaxInt32}},
		{name: "negative priority", options: SharedWorkflowStartOptions{PriorityKey: -1}, wantErr: true},
		{name: "priority above int32 maximum", options: SharedWorkflowStartOptions{PriorityKey: math.MaxInt32 + 1}, wantErr: true},
		{name: "empty fairness key and zero weight", options: SharedWorkflowStartOptions{}},
		{name: "64 byte fairness key", options: SharedWorkflowStartOptions{FairnessKey: strings.Repeat("a", 64)}},
		{name: "65 byte fairness key", options: SharedWorkflowStartOptions{FairnessKey: strings.Repeat("a", 65)}, wantErr: true},
		{name: "minimum fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: 0.001}},
		{name: "maximum fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: 1000}},
		{name: "negative fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: -1}, wantErr: true},
		{name: "fairness weight below minimum", options: SharedWorkflowStartOptions{FairnessWeight: 0.0009}, wantErr: true},
		{name: "fairness weight above maximum", options: SharedWorkflowStartOptions{FairnessWeight: 1000.1}, wantErr: true},
		{name: "NaN fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: float32(math.NaN())}, wantErr: true},
		{name: "positive infinite fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: float32(math.Inf(1))}, wantErr: true},
		{name: "negative infinite fairness weight", options: SharedWorkflowStartOptions{FairnessWeight: float32(math.Inf(-1))}, wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := toScheduleAction(&testCase.options, &PayloadInputOptions{})
			if (err != nil) != testCase.wantErr {
				t.Fatalf("toScheduleAction error = %v, want error = %v", err, testCase.wantErr)
			}
		})
	}
}
