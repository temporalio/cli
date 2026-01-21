package tracer

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/require"
	"go.temporal.io/api/common/v1"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/enums/v1"
)

var foldStatus = []enums.WorkflowExecutionStatus{
	enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
	enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
	enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
	enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
}

func GetRelativeTime(d time.Duration) *timestamppb.Timestamp {
	startTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	if d > 0 {
		startTime = startTime.Add(d)
	}
	return timestamppb.New(startTime)
}

func GetDuration(d time.Duration) *durationpb.Duration {
	return durationpb.New(d)
}

func TestExecuteWorkflowTemplate(t *testing.T) {
	tests := map[string]struct {
		state ExecutionState
		depth int
		want  string
	}{
		"simple wf": {
			state: &WorkflowExecutionState{
				Execution:     &common.WorkflowExecution{WorkflowId: "foo", RunId: "bar"},
				Status:        enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
				Type:          &common.WorkflowType{Name: "foo"},
				Attempt:       1,
				LastEventId:   52,
				HistoryLength: 52,
				ChildStates:   []ExecutionState{},
			},
			depth: 0,
			want: " ╪ ▷ foo \n" +
				" │   wfId: foo, runId: bar\n",
		},
		"extras": {
			state: &WorkflowExecutionState{
				Execution:     &common.WorkflowExecution{WorkflowId: "foo", RunId: "bar"},
				Status:        enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				Type:          &common.WorkflowType{Name: "foo"},
				Attempt:       10,
				LastEventId:   52,
				HistoryLength: 52,
				ChildStates:   []ExecutionState{},
				StartTime:     GetRelativeTime(0),
				CloseTime:     GetRelativeTime(time.Hour),
			},
			depth: 0,
			want: " ╪ ✓ foo (1h0m0s, 10 attempts)\n" +
				" │   wfId: foo, runId: bar\n",
		},
		"child wf": {
			state: &WorkflowExecutionState{
				Execution:     &common.WorkflowExecution{WorkflowId: "foo", RunId: "bar"},
				Status:        enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
				Type:          &common.WorkflowType{Name: "foo"},
				Attempt:       1,
				LastEventId:   52,
				HistoryLength: 52,
				ChildStates: []ExecutionState{
					&WorkflowExecutionState{
						LastEventId:   1, // Child workflow has its own events
						HistoryLength: 1,
						Status:        enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
						Type: &common.WorkflowType{
							Name: "baz",
						},
						Execution: &common.WorkflowExecution{
							WorkflowId: "childWfId",
							RunId:      "childRunId",
						},
						Attempt: 1,
					},
				},
			},
			depth: 0,
			want: " ╪ x foo \n" +
				" │   wfId: foo, runId: bar\n" +
				" │   ╪ ▷ baz \n" +
				" │   │   wfId: childWfId, runId: childRunId\n",
		},
		"activity": {
			state: &ActivityExecutionState{
				ActivityId: "1",
				Status:     ACTIVITY_EXECUTION_STATUS_COMPLETED,
				Type:       &common.ActivityType{Name: "foobar"},
				Attempt:    0,
				RetryState: 0,
				StartTime:  GetRelativeTime(0),
				CloseTime:  GetRelativeTime(10 * time.Second),
			},
			want: " ┼ ✓ foobar (10s)\n",
		},
		"timer": {
			state: &TimerExecutionState{
				StartToFireTimeout: GetDuration(time.Hour),
				Status:             TIMER_STATUS_CANCELED,
				Name:               "Timer (1h0m0s)",
				StartTime:          GetRelativeTime(0),
				CloseTime:          GetRelativeTime(10 * time.Minute),
			},
			want: " ┼ ⧖ Timer (1h0m0s) (10m0s)\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			b := new(bytes.Buffer)
			tmpl, err := NewExecutionTemplate(foldStatus, false)
			require.NoError(t, err)

			require.NoError(t, tmpl.Execute(b, tt.state, tt.depth))

			if runtime.GOOS == "windows" {
				tt.want = strings.ReplaceAll(tt.want, "\n", "\r\n")
			}

			assert.Equal(t, tt.want, b.String())
		})
	}
}

func TestShouldFold(t *testing.T) {
	tests := map[string]struct {
		displayAll   bool
		currentDepth int
		state        *WorkflowExecutionState
		want         bool
	}{
		"running": {
			state: &WorkflowExecutionState{
				Status: enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
			},
			displayAll:   false,
			currentDepth: 1,
			want:         false,
		},
		"completed": {
			state: &WorkflowExecutionState{
				Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			},
			displayAll:   false,
			currentDepth: 1,
			want:         true,
		},
		"canceled": {
			state: &WorkflowExecutionState{
				Status: enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
			},
			displayAll:   false,
			currentDepth: 1,
			want:         true,
		},
		"terminate": {
			state: &WorkflowExecutionState{
				Status: enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
			},
			displayAll:   false,
			currentDepth: 1,
			want:         true,
		},
		"not deep enough": {
			state: &WorkflowExecutionState{
				Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			},
			displayAll:   false,
			currentDepth: 0,
			want:         false,
		},
		"infinite depth": {
			state: &WorkflowExecutionState{
				Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			},
			displayAll:   true,
			currentDepth: 10,
			want:         false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res := ShouldFoldStatus(foldStatus, tt.displayAll)(tt.state, tt.currentDepth)

			assert.Equal(t, tt.want, res)
		})
	}
}
