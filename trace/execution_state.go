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

package sundial

import (
	"fmt"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/history/v1"
	"strconv"
	"time"
)

// ExecutionState provides a common interface to any execution (Workflows, Activities and Timers in this case) updated through HistoryEvents.
type ExecutionState interface {
	Update(*history.HistoryEvent)
	GetName() string
	GetAttempt() int32
	GetFailure() *failure.Failure
	GetRetryState() enums.RetryState

	GetDuration() *time.Duration
	GetStartTime() *time.Time
}

// WorkflowExecutionState is a snapshot of the state of a WorkflowExecution. It is updated through HistoryEvents.
type WorkflowExecutionState struct {
	// Execution is the workflow's execution (WorkflowId and RunId).
	Execution *common.WorkflowExecution
	// Type is the name/type of Workflow.
	Type *common.WorkflowType
	// StartTime is the time the Execution was started (based on the first Execution's Event).
	StartTime *time.Time
	// CloseTime is the time the Execution was closed (based on the first Execution's Event). Will be nil if the Execution hasn't been closed yet.
	CloseTime *time.Time
	// Status is the Execution's Status based on the last event that was processed.
	Status enums.WorkflowExecutionStatus
	// IsArchived will be true if the workflow has been archived.
	IsArchived bool

	// LastEventId is the EventId of the last processed HistoryEvent.
	LastEventId int64
	// HistoryLength is the number of HistoryEvents available in the server. It will zero for archived workflows and non-zero positive for any other workflow executions.
	HistoryLength int64

	// ChildStates contains all the ExecutionStates contained by this WorkflowExecutionState in order of execution.
	ChildStates []ExecutionState
	// activityMap contains all the activities executed in the Workflow, indexed by the EVENT_TYPE_ACTIVITY_TASK_SCHEDULED event id.
	// Used to retrieve the activities from events.
	activityMap map[int64]*ActivityExecutionState
	// childWfMap contains all the child workflows executed in the Workflow, indexed by the EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED event id.
	// Used to retrieve the child workflows from events.
	childWfMap map[int64]*WorkflowExecutionState
	// timerMap contains all the timers executed in the Workflow, indexed by the EVENT_TYPE_TIMER_STARTED event id.
	// Used to retrieve the timers from events.
	timerMap map[int64]*TimerExecutionState

	// Non-successful closed states
	// Failure contains the last failure that the Execution has reported (if any).
	Failure *failure.Failure
	// Termination contains the last available termination information that the Workflow Execution has reported (if any).
	Termination *history.WorkflowExecutionTerminatedEventAttributes
	// CancelRequest contains the last request that has been made to cancel the Workflow Execution (if any).
	CancelRequest *history.WorkflowExecutionCancelRequestedEventAttributes
	// RetryState contains the reason provided for whether the Task should or shouldn't be retried.
	RetryState enums.RetryState

	// Timeout and retry policies
	// WorkflowExecutionTimeout contains the Workflow Execution's timeout if it has been set.
	WorkflowExecutionTimeout *time.Duration
	// Attempt contains the current Workflow Execution's attempt.
	Attempt int32
	// MaximumAttempts contains the maximum number of times the Workflow Execution is allowed to retry before failing.
	MaximumAttempts int32

	// ParentWorkflowExecution identifies the parent Workflow and the execution run.
	ParentWorkflowExecution *common.WorkflowExecution
}

func NewWorkflowExecutionState(wfId, runId string) *WorkflowExecutionState {
	return &WorkflowExecutionState{
		Execution: &common.WorkflowExecution{WorkflowId: wfId, RunId: runId},
	}
}

func (state *WorkflowExecutionState) GetName() string {
	return state.Type.Name
}

func (state *WorkflowExecutionState) GetAttempt() int32 {
	return state.Attempt
}

func (state *WorkflowExecutionState) GetFailure() *failure.Failure {
	return state.Failure
}

func (state *WorkflowExecutionState) GetRetryState() enums.RetryState {
	return state.RetryState
}

func (state *WorkflowExecutionState) GetStartTime() *time.Time {
	return state.StartTime
}

func (state *WorkflowExecutionState) GetDuration() *time.Duration {
	return getDuration(state.StartTime, state.CloseTime)
}

// newActivity adds a new ActivityExecutionState to the WorkflowExecutionState's ChildStates.
func (state *WorkflowExecutionState) newActivity(event *history.HistoryEvent) *ActivityExecutionState {
	if state.activityMap == nil {
		state.activityMap = make(map[int64]*ActivityExecutionState)
	}
	activityState := NewActivityExecutionState()
	activityState.Update(event)

	state.activityMap[event.EventId] = activityState
	state.ChildStates = append(state.ChildStates, activityState)

	return activityState
}

// updateActivity updates a child ActivityExecutionState with a HistoryEvent by its scheduleId
func (state *WorkflowExecutionState) updateActivity(scheduledId int64, event *history.HistoryEvent) {
	if activityState, ok := state.activityMap[scheduledId]; ok {
		activityState.Update(event)
	}
}

// FindChildWorkflow searches for a child workflow that matches the given WorkflowExecution. It's searched within the ChildStates list to avoid concurrent map writes.
func (state *WorkflowExecutionState) FindChildWorkflow(execution *common.WorkflowExecution) *WorkflowExecutionState {
	for _, child := range state.ChildStates {
		if wf, ok := child.(*WorkflowExecutionState); ok && wf.Execution == execution {
			return wf
		}
	}
	return nil
}

// newChildWorkflow adds a new WorkflowExecutionState to the WorkflowExecutionState's ChildStates.
func (state *WorkflowExecutionState) newChildWorkflow(event *history.HistoryEvent) *WorkflowExecutionState {
	if state.childWfMap == nil {
		state.childWfMap = make(map[int64]*WorkflowExecutionState)
	}
	attrs := event.GetStartChildWorkflowExecutionInitiatedEventAttributes()
	childWfState := NewWorkflowExecutionState(attrs.GetWorkflowId(), "")
	childWfState.Type = attrs.GetWorkflowType()

	state.childWfMap[event.EventId] = childWfState
	state.ChildStates = append(state.ChildStates, childWfState)

	return childWfState
}

// newTimer adds a new TimerExecutionState to the WorkflowExecutionState's ChildStates.
func (state *WorkflowExecutionState) newTimer(event *history.HistoryEvent) *TimerExecutionState {
	if state.timerMap == nil {
		state.timerMap = make(map[int64]*TimerExecutionState)
	}
	timerState := &TimerExecutionState{}
	timerState.Update(event)

	state.timerMap[event.EventId] = timerState
	state.ChildStates = append(state.ChildStates, timerState)

	return timerState
}

// updateTimer updates a child TimerExecutionState with a HistoryEvent by its startedId
func (state *WorkflowExecutionState) updateTimer(startedId int64, event *history.HistoryEvent) {
	if timerState, ok := state.timerMap[startedId]; ok {
		timerState.Update(event)
	}
}

// IsCompleted returns true when the Workflow Execution is completed in a non-failed state.
// This is useful to know if we should fetch child workflows or fold the information.
// For now this is when the workflow is completed, terminated or canceled.
func (state *WorkflowExecutionState) IsCompleted() bool {
	return state.Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED ||
		state.Status == enums.WORKFLOW_EXECUTION_STATUS_CANCELED ||
		state.Status == enums.WORKFLOW_EXECUTION_STATUS_TERMINATED
}

// Update updates the WorkflowExecutionState and its child states with a HistoryEvent.
func (state *WorkflowExecutionState) Update(event *history.HistoryEvent) {
	if event == nil {
		return
	}

	state.LastEventId = event.EventId
	switch event.EventType {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
		// Always the first event in a workflow history
		state.Status = enums.WORKFLOW_EXECUTION_STATUS_RUNNING

		attrs := event.GetWorkflowExecutionStartedEventAttributes()
		state.StartTime = event.EventTime
		state.Attempt = attrs.GetAttempt()
		state.Type = attrs.GetWorkflowType()

		state.ParentWorkflowExecution = attrs.ParentWorkflowExecution

		// Cleanup the failure/cancel request
		state.Failure = nil
		state.CancelRequest = nil
		state.Termination = nil

		// Get timeout and max retry info
		state.WorkflowExecutionTimeout = attrs.WorkflowExecutionTimeout
		if attrs.RetryPolicy != nil {
			state.MaximumAttempts = attrs.RetryPolicy.MaximumAttempts
		} else {
			state.MaximumAttempts = 0
		}

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		state.Status = enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		state.Status = enums.WORKFLOW_EXECUTION_STATUS_FAILED

		attrs := event.GetWorkflowExecutionFailedEventAttributes()
		state.Failure = attrs.GetFailure()
		state.RetryState = attrs.GetRetryState()
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
		state.Status = enums.WORKFLOW_EXECUTION_STATUS_TERMINATED
		state.Termination = event.GetWorkflowExecutionTerminatedEventAttributes()
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED:
		state.CancelRequest = event.GetWorkflowExecutionCancelRequestedEventAttributes()

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		state.Status = enums.WORKFLOW_EXECUTION_STATUS_CANCELED
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CONTINUED_AS_NEW:
		state.Status = enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		attrs := event.GetWorkflowExecutionTimedOutEventAttributes()
		state.Status = enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT
		state.CloseTime = event.EventTime
		state.RetryState = attrs.GetRetryState()

	// ACTIVITY EVENTS
	case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
		// First activity event
		state.newActivity(event)
	case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
		attrs := event.GetActivityTaskStartedEventAttributes()
		state.updateActivity(attrs.ScheduledEventId, event)
	case enums.EVENT_TYPE_ACTIVITY_TASK_FAILED:
		attrs := event.GetActivityTaskFailedEventAttributes()
		state.updateActivity(attrs.ScheduledEventId, event)
	case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
		attrs := event.GetActivityTaskCompletedEventAttributes()
		state.updateActivity(attrs.ScheduledEventId, event)
	case enums.EVENT_TYPE_ACTIVITY_TASK_CANCEL_REQUESTED:
		attrs := event.GetActivityTaskCancelRequestedEventAttributes()
		state.updateActivity(attrs.ScheduledEventId, event)
	case enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
		attrs := event.GetActivityTaskCanceledEventAttributes()
		state.updateActivity(attrs.ScheduledEventId, event)
	case enums.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
		attrs := event.GetActivityTaskTimedOutEventAttributes()
		state.updateActivity(attrs.ScheduledEventId, event)

	// CHILD WORKFLOW EVENTS
	case enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED:
		// First child workflow
		state.newChildWorkflow(event)
	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
		attrs := event.GetChildWorkflowExecutionStartedEventAttributes()
		if child, ok := state.childWfMap[attrs.InitiatedEventId]; ok {
			child.Status = enums.WORKFLOW_EXECUTION_STATUS_RUNNING
			child.Execution = attrs.GetWorkflowExecution()
			if child.StartTime == nil {
				child.StartTime = event.EventTime
			}
		}
	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
		attrs := event.GetChildWorkflowExecutionCompletedEventAttributes()
		if child, ok := state.childWfMap[attrs.InitiatedEventId]; ok {
			child.Status = enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
			if child.CloseTime == nil {
				child.CloseTime = event.EventTime
			}
		}
	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED:
		attrs := event.GetChildWorkflowExecutionFailedEventAttributes()
		if child, ok := state.childWfMap[attrs.InitiatedEventId]; ok {
			child.Status = enums.WORKFLOW_EXECUTION_STATUS_FAILED
			child.Failure = attrs.GetFailure()
			child.RetryState = attrs.GetRetryState()
			if child.CloseTime == nil {
				child.CloseTime = event.EventTime
			}
		}
	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TERMINATED:
		attrs := event.GetChildWorkflowExecutionTerminatedEventAttributes()
		// We don't have termination reason from this event :(
		if child, ok := state.childWfMap[attrs.InitiatedEventId]; ok {
			child.Status = enums.WORKFLOW_EXECUTION_STATUS_TERMINATED
			if child.CloseTime == nil {
				child.CloseTime = event.EventTime
			}
		}
	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
		attrs := event.GetChildWorkflowExecutionCanceledEventAttributes()
		if child, ok := state.childWfMap[attrs.InitiatedEventId]; ok {
			child.Status = enums.WORKFLOW_EXECUTION_STATUS_CANCELED
			if child.CloseTime == nil {
				child.CloseTime = event.EventTime
			}
		}
	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
		attrs := event.GetChildWorkflowExecutionTimedOutEventAttributes()
		if child, ok := state.childWfMap[attrs.InitiatedEventId]; ok {
			child.Status = enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT
			if child.CloseTime == nil {
				child.CloseTime = event.EventTime
			}
		}

	// TIMER EVENTS
	case enums.EVENT_TYPE_TIMER_STARTED:
		state.newTimer(event)
	case enums.EVENT_TYPE_TIMER_FIRED:
		startedId := event.GetTimerFiredEventAttributes().GetStartedEventId()
		state.updateTimer(startedId, event)
	case enums.EVENT_TYPE_TIMER_CANCELED:
		startedId := event.GetTimerCanceledEventAttributes().GetStartedEventId()
		state.updateTimer(startedId, event)
	}
}

// GetNumberOfEvents returns a count of the number of events processed and the total for a workflow execution.
// This method iteratively sums the LastEventId (the sequential id of the last event processed) and the HistoryLength for all child workflows
func (state *WorkflowExecutionState) GetNumberOfEvents() (int64, int64) {
	var current, total int64
	if state.ChildStates == nil {
		return 0, 0
	}
	for _, child := range state.ChildStates {
		if childWf, ok := child.(*WorkflowExecutionState); ok {
			c, t := childWf.GetNumberOfEvents()
			current += c
			total += t
		}
	}
	return current + state.LastEventId, total + state.HistoryLength
}

// ActivityExecutionStatus is the Status of an ActivityExecution, analogous to enums.WorkflowExecutionStatus.
type ActivityExecutionStatus int32

var (
	ACTIVITY_EXECUTION_STATUS_UNSPECIFIED      ActivityExecutionStatus = 0
	ACTIVITY_EXECUTION_STATUS_SCHEDULED        ActivityExecutionStatus = 1
	ACTIVITY_EXECUTION_STATUS_RUNNING          ActivityExecutionStatus = 2
	ACTIVITY_EXECUTION_STATUS_COMPLETED        ActivityExecutionStatus = 3
	ACTIVITY_EXECUTION_STATUS_FAILED           ActivityExecutionStatus = 4
	ACTIVITY_EXECUTION_STATUS_TIMED_OUT        ActivityExecutionStatus = 5
	ACTIVITY_EXECUTION_STATUS_CANCEL_REQUESTED ActivityExecutionStatus = 6
	ACTIVITY_EXECUTION_STATUS_CANCELED         ActivityExecutionStatus = 7
)

// ActivityExecutionState is a snapshot of the state of an Activity's Execution.
// It implements the ExecutionState interface so it can be referenced as a WorkflowExecutionState's child state.
type ActivityExecutionState struct {
	// ActivityId is the Activity's id, which will usually be the EventId of the Event it was scheduled with.
	ActivityId string
	// Status is the Execution's Status based on the last event that was processed.
	Status ActivityExecutionStatus
	// Type is the name/type of Activity.
	Type *common.ActivityType
	// Attempt contains the current Activity Execution's attempt.
	// Since Activities' events aren't reported until the Activity is closed, this will always be the last attempt.
	Attempt int32
	// Failure contains the last failure that the Execution has reported (if any).
	Failure *failure.Failure
	// RetryState contains the reason provided for whether the Task should or shouldn't be retried.
	RetryState enums.RetryState

	// StartTime is the time the Execution was started (based on the start Event).
	StartTime *time.Time
	// CloseTime is the time the Execution was closed (based on the closing Event). Will be nil if the Execution hasn't been closed yet.
	CloseTime *time.Time
}

func NewActivityExecutionState() *ActivityExecutionState {
	return &ActivityExecutionState{
		Status: ACTIVITY_EXECUTION_STATUS_UNSPECIFIED,
	}
}

func (state *ActivityExecutionState) GetName() string {
	return state.Type.Name
}

func (state *ActivityExecutionState) GetAttempt() int32 {
	return state.Attempt
}

func (state *ActivityExecutionState) GetFailure() *failure.Failure {
	return state.Failure
}

func (state *ActivityExecutionState) GetRetryState() enums.RetryState {
	return state.RetryState
}

func (state *ActivityExecutionState) GetStartTime() *time.Time {
	return state.StartTime
}

func (state *ActivityExecutionState) GetDuration() *time.Duration {
	return getDuration(state.StartTime, state.CloseTime)
}

// Update updates the ActivityExecutionState with a HistoryEvent.
func (state *ActivityExecutionState) Update(event *history.HistoryEvent) {
	switch event.EventType {
	case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
		state.Status = ACTIVITY_EXECUTION_STATUS_SCHEDULED

		attrs := event.GetActivityTaskScheduledEventAttributes()
		state.ActivityId = attrs.GetActivityId()
		state.Type = attrs.GetActivityType()

	case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
		state.Status = ACTIVITY_EXECUTION_STATUS_RUNNING

		attrs := event.GetActivityTaskStartedEventAttributes()
		state.Attempt = attrs.GetAttempt()

		// This is the best guess we have for when the activity was started
		state.StartTime = event.EventTime

		// Clear failures
		state.Failure = nil

	case enums.EVENT_TYPE_ACTIVITY_TASK_FAILED:
		state.Status = ACTIVITY_EXECUTION_STATUS_FAILED

		attrs := event.GetActivityTaskFailedEventAttributes()
		state.Failure = attrs.GetFailure()
		state.RetryState = attrs.GetRetryState()
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
		state.Status = ACTIVITY_EXECUTION_STATUS_COMPLETED
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_ACTIVITY_TASK_CANCEL_REQUESTED:
		state.Status = ACTIVITY_EXECUTION_STATUS_CANCEL_REQUESTED

	case enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
		state.Status = ACTIVITY_EXECUTION_STATUS_CANCELED
		state.CloseTime = event.EventTime

	case enums.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
		state.Status = ACTIVITY_EXECUTION_STATUS_TIMED_OUT

		attrs := event.GetActivityTaskTimedOutEventAttributes()
		state.Failure = attrs.GetFailure()
		state.RetryState = attrs.GetRetryState()
		state.CloseTime = event.EventTime
	}
}

// TimerExecutionState contains information about a Timer as an execution.
// It implements the ExecutionState interface so it can be referenced as a WorkflowExecutionState's child state.
type TimerExecutionState struct {
	TimerId string
	// Name is the name of the Timer (if any has been given to it)
	Name string
	// StartToFireTimeout is the amount of time to elapse before the timer fires.
	StartToFireTimeout *time.Duration
	// Status is the Execution's Status based on the last event that was processed.
	Status TimerExecutionStatus
	// StartTime is the time the Execution was started (based on the start Event).
	StartTime *time.Time
	// CloseTime is the time the Execution was closed (based on the closing Event). Will be nil if the Execution hasn't been closed yet.
	CloseTime *time.Time
}

// TimerExecutionStatus is the Status of a TimerExecution, analogous to enums.WorkflowExecutionStatus.
type TimerExecutionStatus int32

var (
	TIMER_STATUS_WAITING  TimerExecutionStatus = 0
	TIMER_STATUS_FIRED    TimerExecutionStatus = 1
	TIMER_STATUS_CANCELED TimerExecutionStatus = 2
)

// Update updates the TimerExecutionState with a HistoryEvent.
func (t *TimerExecutionState) Update(event *history.HistoryEvent) {
	switch event.EventType {
	case enums.EVENT_TYPE_TIMER_STARTED:
		attrs := event.GetTimerStartedEventAttributes()

		t.StartToFireTimeout = attrs.StartToFireTimeout
		t.TimerId = attrs.TimerId
		if attrs.TimerId != strconv.FormatInt(event.EventId, 10) {
			// If the user has set a custom id, we can use it for the name
			t.Name = fmt.Sprintf("%s (%s)", attrs.TimerId, t.StartToFireTimeout.String())
		} else {
			t.Name = fmt.Sprintf("Timer (%s)", t.StartToFireTimeout.String())
		}
		t.Status = TIMER_STATUS_WAITING
		t.StartTime = event.EventTime
	case enums.EVENT_TYPE_TIMER_FIRED:
		t.Status = TIMER_STATUS_FIRED
		t.CloseTime = event.EventTime
	case enums.EVENT_TYPE_TIMER_CANCELED:
		t.Status = TIMER_STATUS_CANCELED
		t.CloseTime = event.EventTime
	}
}

func (t *TimerExecutionState) GetName() string {
	return t.Name
}

func (t *TimerExecutionState) GetAttempt() int32 {
	return 1
}

func (t *TimerExecutionState) GetFailure() *failure.Failure {
	return nil
}

// GetRetryState will always return RETRY_STATE_UNSPECIFIED since Timers don't retry.
func (t *TimerExecutionState) GetRetryState() enums.RetryState {
	return enums.RETRY_STATE_UNSPECIFIED
}

func (t *TimerExecutionState) GetDuration() *time.Duration {
	return getDuration(t.StartTime, t.CloseTime)
}

func (t *TimerExecutionState) GetStartTime() *time.Time {
	return t.StartTime
}

// Utilities
// getDuration converts a start and completed time to a duration.
func getDuration(started, completed *time.Time) *time.Duration {
	if started == nil || completed == nil {
		return nil
	}
	duration := completed.Sub(*started)
	return &duration
}
