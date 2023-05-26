// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package trace

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/history/v1"
)

var events = map[string]*history.HistoryEvent{
	"started": {
		EventId:   1,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &history.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &history.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &common.WorkflowType{Name: "foo"},
				Attempt:      1,
			},
		},
	},
	"completed": {
		EventId:   100,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED,
		Attributes: &history.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
			WorkflowExecutionCompletedEventAttributes: &history.WorkflowExecutionCompletedEventAttributes{
				WorkflowTaskCompletedEventId: 120,
				NewExecutionRunId:            "foobar",
			},
		},
	},
	"failed": {
		EventId:   89,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		Attributes: &history.HistoryEvent_WorkflowExecutionFailedEventAttributes{
			WorkflowExecutionFailedEventAttributes: &history.WorkflowExecutionFailedEventAttributes{
				Failure: &failure.Failure{
					Message: "Totally expected workflow failure",
				},
				RetryState:                   enums.RETRY_STATE_NON_RETRYABLE_FAILURE,
				WorkflowTaskCompletedEventId: 120,
				NewExecutionRunId:            "foobar",
			},
		},
	},
	"cancel requested": {
		EventId:   89,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED,
		Attributes: &history.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes{
			WorkflowExecutionCancelRequestedEventAttributes: &history.WorkflowExecutionCancelRequestedEventAttributes{
				Cause:    "foobar",
				Identity: "unit test",
			},
		},
	},
	"canceled": {
		EventId:   90,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED,
		Attributes: &history.HistoryEvent_WorkflowExecutionCanceledEventAttributes{
			WorkflowExecutionCanceledEventAttributes: &history.WorkflowExecutionCanceledEventAttributes{},
		},
	},
	"activity scheduled": {
		EventId:   10,
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &history.HistoryEvent_ActivityTaskScheduledEventAttributes{
			ActivityTaskScheduledEventAttributes: &history.ActivityTaskScheduledEventAttributes{
				ActivityId:   "abc",
				ActivityType: &common.ActivityType{Name: "Mr ActivityFace"},
			},
		},
	},
	"activity started": {
		EventId:   13,
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_STARTED,
		Attributes: &history.HistoryEvent_ActivityTaskStartedEventAttributes{
			ActivityTaskStartedEventAttributes: &history.ActivityTaskStartedEventAttributes{
				ScheduledEventId: 10,
				Identity:         "worker-baz",
				Attempt:          1,
			},
		},
	},
	"activity failed": {
		EventId:   20, // same Id as completed for testing purposes
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_FAILED,
		Attributes: &history.HistoryEvent_ActivityTaskFailedEventAttributes{
			ActivityTaskFailedEventAttributes: &history.ActivityTaskFailedEventAttributes{
				ScheduledEventId: 10,
				StartedEventId:   13,
				Identity:         "worker-baz",
				Failure:          &failure.Failure{Message: "I was a test"},
			},
		},
	},
	"activity completed": {
		EventId:   20, // same Id as failed for testing purposes
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED,
		Attributes: &history.HistoryEvent_ActivityTaskCompletedEventAttributes{
			ActivityTaskCompletedEventAttributes: &history.ActivityTaskCompletedEventAttributes{
				ScheduledEventId: 10,
				StartedEventId:   13,
				Identity:         "worker-baz",
			},
		},
	},
	"activity cancel requested": {
		EventId:   20, // same Id as failed for testing purposes
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_CANCEL_REQUESTED,
		Attributes: &history.HistoryEvent_ActivityTaskCancelRequestedEventAttributes{
			ActivityTaskCancelRequestedEventAttributes: &history.ActivityTaskCancelRequestedEventAttributes{
				ScheduledEventId: 10,
			},
		},
	},
	"activity canceled": {
		EventId:   21,
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED,
		Attributes: &history.HistoryEvent_ActivityTaskCanceledEventAttributes{
			ActivityTaskCanceledEventAttributes: &history.ActivityTaskCanceledEventAttributes{
				ScheduledEventId:             10,
				LatestCancelRequestedEventId: 20,
				Identity:                     "unit test",
			},
		},
	},
	"second activity scheduled": {
		EventId:   30,
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &history.HistoryEvent_ActivityTaskScheduledEventAttributes{
			ActivityTaskScheduledEventAttributes: &history.ActivityTaskScheduledEventAttributes{
				ActivityId:   "def",
				ActivityType: &common.ActivityType{Name: "Hyperactivity"},
			},
		},
	},
	"child workflow initiated": {
		EventId:   50,
		EventType: enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED,
		Attributes: &history.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes{
			StartChildWorkflowExecutionInitiatedEventAttributes: &history.StartChildWorkflowExecutionInitiatedEventAttributes{
				Namespace:    "default",
				WorkflowId:   "childWfId",
				WorkflowType: &common.WorkflowType{Name: "baz"},
			},
		},
	},
	"child workflow started": {
		EventId:   52,
		EventType: enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED,
		Attributes: &history.HistoryEvent_ChildWorkflowExecutionStartedEventAttributes{
			ChildWorkflowExecutionStartedEventAttributes: &history.ChildWorkflowExecutionStartedEventAttributes{
				Namespace:        "default",
				InitiatedEventId: 50,
				WorkflowExecution: &common.WorkflowExecution{
					WorkflowId: "childWfId", RunId: "childRunId",
				},
				WorkflowType: &common.WorkflowType{Name: "baz"},
			},
		},
	},
	"child workflow completed": {
		EventId:   60,
		EventType: enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED,
		Attributes: &history.HistoryEvent_ChildWorkflowExecutionCompletedEventAttributes{
			ChildWorkflowExecutionCompletedEventAttributes: &history.ChildWorkflowExecutionCompletedEventAttributes{
				Namespace: "default",
				WorkflowExecution: &common.WorkflowExecution{
					WorkflowId: "childWfId", RunId: "childRunId",
				},
				WorkflowType:     &common.WorkflowType{Name: "baz"},
				InitiatedEventId: 50,
				StartedEventId:   52,
			},
		},
	},
	"child workflow failed": {
		EventId:   55,
		EventType: enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED,
		Attributes: &history.HistoryEvent_ChildWorkflowExecutionFailedEventAttributes{
			ChildWorkflowExecutionFailedEventAttributes: &history.ChildWorkflowExecutionFailedEventAttributes{
				Failure: &failure.Failure{
					Message: "This child failed us",
				},
				RetryState: enums.RETRY_STATE_MAXIMUM_ATTEMPTS_REACHED,
				Namespace:  "default",
				WorkflowExecution: &common.WorkflowExecution{
					WorkflowId: "childWfId", RunId: "childRunId",
				},
				WorkflowType:     &common.WorkflowType{Name: "baz"},
				InitiatedEventId: 50,
				StartedEventId:   52,
			},
		},
	},
	"child workflow canceled": {
		EventId:   55,
		EventType: enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED,
		Attributes: &history.HistoryEvent_ChildWorkflowExecutionCanceledEventAttributes{
			ChildWorkflowExecutionCanceledEventAttributes: &history.ChildWorkflowExecutionCanceledEventAttributes{
				Namespace: "default",
				WorkflowExecution: &common.WorkflowExecution{
					WorkflowId: "childWfId", RunId: "childRunId",
				},
				WorkflowType:     &common.WorkflowType{Name: "baz"},
				InitiatedEventId: 50,
				StartedEventId:   52,
			},
		},
	},
	"workflow started child": {
		EventId:   1,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &history.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &history.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &common.WorkflowType{Name: "baz"},
				Attempt:      1,
			},
		},
	},
	"child workflow initiated child": { // Child workflow has a child workflow
		EventId:   50,
		EventType: enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED,
		Attributes: &history.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes{
			StartChildWorkflowExecutionInitiatedEventAttributes: &history.StartChildWorkflowExecutionInitiatedEventAttributes{
				Namespace:    "default",
				WorkflowId:   "depth2child",
				WorkflowType: &common.WorkflowType{Name: "baz"},
			},
		},
	},
	"child workflow started child": {
		EventId:   52,
		EventType: enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED,
		Attributes: &history.HistoryEvent_ChildWorkflowExecutionStartedEventAttributes{
			ChildWorkflowExecutionStartedEventAttributes: &history.ChildWorkflowExecutionStartedEventAttributes{
				Namespace:        "default",
				InitiatedEventId: 50,
				WorkflowExecution: &common.WorkflowExecution{
					WorkflowId: "depth2child", RunId: "depth2childRunId",
				},
				WorkflowType: &common.WorkflowType{Name: "baz"},
			},
		},
	},
	"timer started": {
		EventId:   20,
		EventType: enums.EVENT_TYPE_TIMER_STARTED,
		Attributes: &history.HistoryEvent_TimerStartedEventAttributes{
			TimerStartedEventAttributes: &history.TimerStartedEventAttributes{
				TimerId:            "20", // If TimerId is not set it'll be the same as EventId
				StartToFireTimeout: NewDuration(time.Hour),
			},
		},
	},
	"timer fired": {
		EventId:   21,
		EventType: enums.EVENT_TYPE_TIMER_FIRED,
		Attributes: &history.HistoryEvent_TimerFiredEventAttributes{
			TimerFiredEventAttributes: &history.TimerFiredEventAttributes{
				TimerId:        "20",
				StartedEventId: 20,
			},
		},
	},
	"timer canceled": {
		EventId:   21,
		EventType: enums.EVENT_TYPE_TIMER_CANCELED,
		Attributes: &history.HistoryEvent_TimerCanceledEventAttributes{
			TimerCanceledEventAttributes: &history.TimerCanceledEventAttributes{
				TimerId:        "20",
				StartedEventId: 20,
				Identity:       "test",
			},
		},
	},
}

func NewDuration(d time.Duration) *time.Duration {
	return &d
}

func NewTime(d time.Duration) *time.Time {
	t := time.Time{}.Add(d)
	return &t
}

func TestExecutionState_UpdateWorkflow(t *testing.T) {
	tests := map[string]struct {
		events        []*history.HistoryEvent
		expectedState *WorkflowExecutionState
	}{
		"workflow started": {
			events: []*history.HistoryEvent{events["started"]},
			expectedState: &WorkflowExecutionState{
				LastEventId: 1,
				Execution:   &common.WorkflowExecution{WorkflowId: "foo"},
				Type:        &common.WorkflowType{Name: "foo"},
				Status:      enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
				Attempt:     1,
			},
		},
		"workflow completed": {
			events: []*history.HistoryEvent{events["started"], events["completed"]},
			expectedState: &WorkflowExecutionState{
				LastEventId: 100,
				Execution:   &common.WorkflowExecution{WorkflowId: "foo"},
				Type:        &common.WorkflowType{Name: "foo"},
				Status:      enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				Attempt:     1,
			},
		},
		"workflow failed": {
			events: []*history.HistoryEvent{events["started"], events["failed"]},
			expectedState: &WorkflowExecutionState{
				LastEventId: 89,
				Execution:   &common.WorkflowExecution{WorkflowId: "foo"},
				Type:        &common.WorkflowType{Name: "foo"},
				Status:      enums.WORKFLOW_EXECUTION_STATUS_FAILED,
				Attempt:     1,
				Failure:     &failure.Failure{Message: "Totally expected workflow failure"},
				RetryState:  enums.RETRY_STATE_NON_RETRYABLE_FAILURE,
			},
		},
		"workflow cancel requested": {
			events: []*history.HistoryEvent{events["started"], events["cancel requested"]},
			expectedState: &WorkflowExecutionState{
				LastEventId: 89,
				Execution:   &common.WorkflowExecution{WorkflowId: "foo"},
				Type:        &common.WorkflowType{Name: "foo"},
				Status:      enums.WORKFLOW_EXECUTION_STATUS_RUNNING, // There's no cancel requested status in Temporal's enums
				CancelRequest: &history.WorkflowExecutionCancelRequestedEventAttributes{
					Cause:    "foobar",
					Identity: "unit test",
				},
				Attempt: 1,
			},
		},
		"workflow canceled": {
			events: []*history.HistoryEvent{events["started"], events["cancel requested"], events["canceled"]},
			expectedState: &WorkflowExecutionState{
				LastEventId: 90,
				Execution:   &common.WorkflowExecution{WorkflowId: "foo"},
				Type:        &common.WorkflowType{Name: "foo"},
				Status:      enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
				CancelRequest: &history.WorkflowExecutionCancelRequestedEventAttributes{
					Cause:    "foobar",
					Identity: "unit test",
				},
				Attempt: 1,
			},
		},
		// TODO: Add terminated and canceled events
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			state := NewWorkflowExecutionState("foo", "")
			for _, event := range tt.events {
				state.Update(event)
			}
			assert.Equal(t, tt.expectedState, state)
		})
	}
}

func TestExecutionState_UpdateActivities(t *testing.T) {
	tests := map[string]struct {
		events           []*history.HistoryEvent
		expectedChildren []ExecutionState
	}{
		"activity scheduled": {
			events: []*history.HistoryEvent{events["started"], events["activity scheduled"]},
			expectedChildren: []ExecutionState{
				&ActivityExecutionState{
					ActivityId: "abc",
					Type:       &common.ActivityType{Name: "Mr ActivityFace"},
					Status:     ACTIVITY_EXECUTION_STATUS_SCHEDULED,
				},
			},
		},
		"activity started": {
			events: []*history.HistoryEvent{
				events["started"],
				events["activity scheduled"],
				events["activity started"],
			},
			expectedChildren: []ExecutionState{
				&ActivityExecutionState{
					ActivityId: "abc",
					Type:       &common.ActivityType{Name: "Mr ActivityFace"},
					Status:     ACTIVITY_EXECUTION_STATUS_RUNNING,
					Attempt:    1,
				},
			},
		},
		"activity failed": {
			events: []*history.HistoryEvent{
				events["started"],
				events["activity scheduled"],
				events["activity started"],
				events["activity failed"],
			},
			expectedChildren: []ExecutionState{
				&ActivityExecutionState{
					ActivityId: "abc",
					Type:       &common.ActivityType{Name: "Mr ActivityFace"},
					Status:     ACTIVITY_EXECUTION_STATUS_FAILED,
					Attempt:    1,
					Failure:    &failure.Failure{Message: "I was a test"},
				},
			},
		},
		"activity completed": {
			events: []*history.HistoryEvent{
				events["started"],
				events["activity scheduled"],
				events["activity started"],
				events["activity completed"],
			},
			expectedChildren: []ExecutionState{
				&ActivityExecutionState{
					ActivityId: "abc",
					Type:       &common.ActivityType{Name: "Mr ActivityFace"},
					Status:     ACTIVITY_EXECUTION_STATUS_COMPLETED,
					Attempt:    1,
				},
			},
		},
		"second activity scheduled": {
			events: []*history.HistoryEvent{
				events["started"],
				events["activity scheduled"],
				events["second activity scheduled"],
				events["activity started"],
			},
			expectedChildren: []ExecutionState{
				&ActivityExecutionState{
					ActivityId: "abc",
					Type:       &common.ActivityType{Name: "Mr ActivityFace"},
					Status:     ACTIVITY_EXECUTION_STATUS_RUNNING,
					Attempt:    1,
				},
				&ActivityExecutionState{
					ActivityId: "def",
					Type:       &common.ActivityType{Name: "Hyperactivity"},
					Status:     ACTIVITY_EXECUTION_STATUS_SCHEDULED,
				},
			},
		},
		"activity cancel requested": {
			events: []*history.HistoryEvent{
				events["started"],
				events["activity scheduled"],
				events["activity started"],
				events["activity cancel requested"],
			},
			expectedChildren: []ExecutionState{
				&ActivityExecutionState{
					ActivityId: "abc",
					Type:       &common.ActivityType{Name: "Mr ActivityFace"},
					Status:     ACTIVITY_EXECUTION_STATUS_CANCEL_REQUESTED,
					Attempt:    1,
				},
			},
		},
		"activity canceled": {
			events: []*history.HistoryEvent{
				events["started"],
				events["activity scheduled"],
				events["activity started"],
				events["activity cancel requested"],
				events["activity canceled"],
			},
			expectedChildren: []ExecutionState{
				&ActivityExecutionState{
					ActivityId: "abc",
					Type:       &common.ActivityType{Name: "Mr ActivityFace"},
					Status:     ACTIVITY_EXECUTION_STATUS_CANCELED,
					Attempt:    1,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			state := NewWorkflowExecutionState("foo", "")
			for _, event := range tt.events {
				state.Update(event)
			}
			assert.Equal(t, tt.expectedChildren, state.ChildStates)
		})
	}
}

func TestExecutionState_UpdateChildWorkflows(t *testing.T) {
	tests := map[string]struct {
		events           []*history.HistoryEvent
		expectedChildren []ExecutionState
	}{
		"child workflow initiated": {
			events: []*history.HistoryEvent{events["started"], events["child workflow initiated"]},
			expectedChildren: []ExecutionState{
				&WorkflowExecutionState{
					Type: &common.WorkflowType{Name: "baz"},
					Execution: &common.WorkflowExecution{
						WorkflowId: "childWfId", RunId: "",
					},
					ParentWorkflowExecution: &common.WorkflowExecution{
						WorkflowId: "foo",
					},
				},
			},
		},
		"child workflow started": {
			events: []*history.HistoryEvent{
				events["started"],
				events["child workflow initiated"],
				events["child workflow started"],
			},
			expectedChildren: []ExecutionState{
				&WorkflowExecutionState{
					Type:   &common.WorkflowType{Name: "baz"},
					Status: enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
					Execution: &common.WorkflowExecution{
						WorkflowId: "childWfId", RunId: "childRunId",
					},
					ParentWorkflowExecution: &common.WorkflowExecution{
						WorkflowId: "foo",
					},
				},
			},
		},
		"child workflow completed": {
			events: []*history.HistoryEvent{
				events["started"],
				events["child workflow initiated"],
				events["child workflow started"],
				events["child workflow completed"],
			},
			expectedChildren: []ExecutionState{
				&WorkflowExecutionState{
					Type:   &common.WorkflowType{Name: "baz"},
					Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
					Execution: &common.WorkflowExecution{
						WorkflowId: "childWfId", RunId: "childRunId",
					},
					ParentWorkflowExecution: &common.WorkflowExecution{
						WorkflowId: "foo",
					},
				},
			},
		},
		"child workflow failed": {
			events: []*history.HistoryEvent{
				events["started"],
				events["child workflow initiated"],
				events["child workflow started"],
				events["child workflow failed"],
			},
			expectedChildren: []ExecutionState{
				&WorkflowExecutionState{
					Type:   &common.WorkflowType{Name: "baz"},
					Status: enums.WORKFLOW_EXECUTION_STATUS_FAILED,
					Execution: &common.WorkflowExecution{
						WorkflowId: "childWfId", RunId: "childRunId",
					},
					ParentWorkflowExecution: &common.WorkflowExecution{
						WorkflowId: "foo",
					},
					Failure: &failure.Failure{
						Message: "This child failed us",
					},
					RetryState: enums.RETRY_STATE_MAXIMUM_ATTEMPTS_REACHED,
				},
			},
		},
		"child workflow canceled": {
			events: []*history.HistoryEvent{
				events["started"],
				events["child workflow initiated"],
				events["child workflow started"],
				events["child workflow canceled"],
			},
			expectedChildren: []ExecutionState{
				&WorkflowExecutionState{
					Type:   &common.WorkflowType{Name: "baz"},
					Status: enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
					Execution: &common.WorkflowExecution{
						WorkflowId: "childWfId", RunId: "childRunId",
					},
					ParentWorkflowExecution: &common.WorkflowExecution{
						WorkflowId: "foo",
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			state := NewWorkflowExecutionState("foo", "")
			for _, event := range tt.events {
				state.Update(event)
			}
			assert.Equal(t, tt.expectedChildren, state.ChildStates)
		})
	}
}

func TestExecutionState_UpdateTimers(t *testing.T) {
	tests := map[string]struct {
		events           []*history.HistoryEvent
		expectedChildren []ExecutionState
	}{
		"timer started": {
			events: []*history.HistoryEvent{events["started"], events["timer started"]},
			expectedChildren: []ExecutionState{
				&TimerExecutionState{
					Name:               "Timer (1h0m0s)",
					TimerId:            "20",
					StartToFireTimeout: NewDuration(time.Hour),
					Status:             TIMER_STATUS_WAITING,
				},
			},
		},
		"timer fired": {
			events: []*history.HistoryEvent{events["started"], events["timer started"], events["timer fired"]},
			expectedChildren: []ExecutionState{
				&TimerExecutionState{
					Name:               "Timer (1h0m0s)",
					TimerId:            "20",
					StartToFireTimeout: NewDuration(time.Hour),
					Status:             TIMER_STATUS_FIRED,
				},
			},
		},
		"timer canceled": {
			events: []*history.HistoryEvent{events["started"], events["timer started"], events["timer canceled"]},
			expectedChildren: []ExecutionState{
				&TimerExecutionState{
					Name:               "Timer (1h0m0s)",
					TimerId:            "20",
					StartToFireTimeout: NewDuration(time.Hour),
					Status:             TIMER_STATUS_CANCELED,
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			state := NewWorkflowExecutionState("foo", "")
			for _, event := range tt.events {
				state.Update(event)
			}
			assert.Equal(t, tt.expectedChildren, state.ChildStates)
		})
	}
}

func TestExecutionState_TimerExecutionStateImplementation(t *testing.T) {
	type expectations struct {
		Name       string
		Attempt    int32
		Failure    *failure.Failure
		RetryState enums.RetryState
		Duration   *time.Duration
		StartTime  *time.Time
	}
	tests := map[string]struct {
		State    *TimerExecutionState
		Expected expectations
	}{
		"waiting timer": {
			State: &TimerExecutionState{
				TimerId:            "12",
				StartToFireTimeout: NewDuration(time.Hour),
				Status:             TIMER_STATUS_WAITING,
				StartTime:          NewTime(0),
			},
			Expected: expectations{
				Attempt:    1,
				Failure:    nil,
				RetryState: 0,
				Duration:   nil,
				StartTime:  NewTime(0),
			},
		},
		"fired timer": {
			State: &TimerExecutionState{
				TimerId:            "12",
				StartToFireTimeout: NewDuration(time.Hour),
				Status:             TIMER_STATUS_FIRED,
				StartTime:          NewTime(0),
				CloseTime:          NewTime(time.Hour),
			},
			Expected: expectations{
				Attempt:    1,
				Failure:    nil,
				RetryState: 0,
				Duration:   NewDuration(time.Hour),
				StartTime:  NewTime(0),
			},
		},
		"canceled timer ": {
			State: &TimerExecutionState{
				TimerId:            "12",
				StartToFireTimeout: NewDuration(time.Hour),
				Status:             TIMER_STATUS_CANCELED,
				StartTime:          NewTime(0),
				CloseTime:          NewTime(30 * time.Minute),
			},
			Expected: expectations{
				Attempt:    1,
				Failure:    nil,
				RetryState: 0,
				Duration:   NewDuration(30 * time.Minute),
				StartTime:  NewTime(0),
			},
		},
		"named timer": {
			State: &TimerExecutionState{
				TimerId:            "12",
				Name:               "TestTimer",
				StartToFireTimeout: NewDuration(time.Hour),
				Status:             TIMER_STATUS_WAITING,
				StartTime:          NewTime(0),
			},
			Expected: expectations{
				Name:       "TestTimer",
				Attempt:    1,
				Failure:    nil,
				RetryState: 0,
				Duration:   nil,
				StartTime:  NewTime(0),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.Expected.Name, tt.State.GetName(), "GetName")
			assert.Equal(t, tt.Expected.Attempt, tt.State.GetAttempt(), "GetAttempt")
			assert.Equal(t, tt.Expected.Failure, tt.State.GetFailure(), "GetFailure")
			assert.Equal(t, tt.Expected.RetryState, tt.State.GetRetryState(), "GetRetryState")
			assert.Equal(t, tt.Expected.Duration, tt.State.GetDuration(), fmt.Sprintf("GetDuration missmatch (expected %s, got %s)", tt.Expected.Duration, tt.State.GetDuration()))
			assert.Equal(t, tt.Expected.StartTime, tt.State.GetStartTime(), "GetStartTime")
		})
	}
}

func TestWorkflowExecutionState_GetNumberOfEvents(t *testing.T) {
	tests := map[string]struct {
		state       *WorkflowExecutionState
		wantCurrent int64
		wantTotal   int64
	}{
		"no children": {
			state: &WorkflowExecutionState{
				LastEventId:   1,
				HistoryLength: 5,
			},
			wantCurrent: 1,
			wantTotal:   5,
		},
		"one child": {
			state: &WorkflowExecutionState{
				LastEventId:   5,
				HistoryLength: 10,
				ChildStates: []ExecutionState{
					&WorkflowExecutionState{
						LastEventId:   1,
						HistoryLength: 3,
					},
				},
			},
			wantCurrent: 6,
			wantTotal:   13,
		},
		"multiple children": {
			state: &WorkflowExecutionState{
				LastEventId:   5,
				HistoryLength: 10,
				ChildStates: []ExecutionState{
					&WorkflowExecutionState{
						LastEventId:   1,
						HistoryLength: 3,
					},
					&ActivityExecutionState{},
					&TimerExecutionState{},
					&WorkflowExecutionState{
						LastEventId:   1,
						HistoryLength: 3,
					},
				},
			},
			wantCurrent: 7,
			wantTotal:   16,
		},
		"multiple depths": {
			state: &WorkflowExecutionState{
				LastEventId:   5,
				HistoryLength: 10,
				ChildStates: []ExecutionState{
					&WorkflowExecutionState{
						LastEventId:   1,
						HistoryLength: 3,
					},
					&WorkflowExecutionState{
						LastEventId:   5,
						HistoryLength: 10,
						ChildStates: []ExecutionState{
							&WorkflowExecutionState{
								LastEventId:   1,
								HistoryLength: 3,
							},
						},
					},
				},
			},
			wantCurrent: 12,
			wantTotal:   26,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resCurrent, resTotal := tt.state.GetNumberOfEvents()
			assert.Equalf(t, tt.wantCurrent, resCurrent, "GetNumberOfEvents() current")
			assert.Equalf(t, tt.wantTotal, resTotal, "GetNumberOfEvents() total")
		})
	}
}
