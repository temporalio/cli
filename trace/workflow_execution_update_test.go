package trace

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/mocks"
)

type WorkflowExecutionUpdateSuite struct {
	suite.Suite
	ctx context.Context
}

func TestWorkflowExecutionUpdateSuite(t *testing.T) {
	suite.Run(t, new(WorkflowExecutionUpdateSuite))
}

func (s *WorkflowExecutionUpdateSuite) SetupTest() {
	s.ctx = context.Background()
}

// SetDescribeWorkflowMocks mocks DescribeWorkflowExecution to return a list of events for a wfId/runId.
// This is used in the WorkflowStateJob.Run method to retrieve the history length.
func (s *WorkflowExecutionUpdateSuite) SetDescribeWorkflowMocks(client *mocks.Client, wfId, runId string, events []*history.HistoryEvent) {
	var historyLength int64
	if len(events) > 0 {
		historyLength = events[len(events)-1].EventId // Get last event id
	}
	client.On("DescribeWorkflowExecution", mock.Anything, wfId, runId).Return(
		&workflowservice.DescribeWorkflowExecutionResponse{
			WorkflowExecutionInfo: &workflow.WorkflowExecutionInfo{
				HistoryLength: historyLength,
			},
		}, nil)
}

// SetDescribeWorkflowErrorMocks mocks DescribeWorkflowExecution to return an error.
func (s *WorkflowExecutionUpdateSuite) SetDescribeWorkflowErrorMocks(client *mocks.Client, wfId, runId string, err error) {
	client.On("DescribeWorkflowExecution", mock.Anything, wfId, runId).Return(nil, err)
}

// SetWorkflowHistoryMocks mocks GetWorkflowHistory to return an iterator that will return each event in succession.
func (s *WorkflowExecutionUpdateSuite) SetWorkflowHistoryMocks(client *mocks.Client, wfId, runId string, events []*history.HistoryEvent) {
	mockIterator := &mocks.HistoryEventIterator{}
	for _, event := range events {
		mockIterator.On("HasNext").Return(true).Once()
		mockIterator.On("Next").Return(event, nil).Once()
	}
	mockIterator.On("HasNext").Return(false)
	mockIterator.On("Next").Return(nil, fmt.Errorf("no more events available"))
	client.On("GetWorkflowHistory", mock.Anything, wfId, runId, mock.Anything, mock.Anything).Return(mockIterator)
}

// RemoveStateMaps removes caching data from WorkflowExecutionState (i.e. activityMap and childWfMap) so it can be easily asserted upon
func RemoveStateMaps(state *WorkflowExecutionState) *WorkflowExecutionState {
	state.childWfMap = nil
	state.activityMap = nil
	state.timerMap = nil
	return state
}

func (s *WorkflowExecutionUpdateSuite) Test_ParametersAreValid() {
	type args struct {
		concurrency int
	}
	tests := map[string]struct {
		args        args
		assertError assert.ErrorAssertionFunc
	}{
		"sanity": {
			args:        args{concurrency: 1},
			assertError: assert.NoError,
		},
		"zero concurrency": {
			args:        args{concurrency: 0},
			assertError: assert.Error,
		},
		"negative concurrency": {
			args:        args{concurrency: -1},
			assertError: assert.Error,
		},
	}

	for name, tt := range tests {
		s.Run(name, func() {
			client := &mocks.Client{}
			// Setup mocks since valid parameters will start WorkflowStateJobs
			s.SetDescribeWorkflowMocks(client, "foo", "bar", nil)
			s.SetWorkflowHistoryMocks(client, "foo", "bar", nil)
			iter, err := GetWorkflowExecutionUpdates(s.ctx, client, "foo", "bar", true, -1, tt.args.concurrency)

			tt.assertError(s.T(), err)

			// Iterate through all the events to make sure we have no issues due to the assertion coming before errors in the started jobs
			if iter != nil {
				for iter.HasNext() {
					_, err = iter.Next()
					s.NoError(err)
				}
			}
		})
	}
}

func (s *WorkflowExecutionUpdateSuite) Test_GetWorkflowExecutionUpdates() {
	tests := map[string]struct {
		depth         int
		initialState  *WorkflowExecutionState
		rootEvents    []*history.HistoryEvent
		childEvents   []*history.HistoryEvent
		expectedState *WorkflowExecutionState
	}{
		"sanity": {
			depth:        -1,
			initialState: NewWorkflowExecutionState("foo", "bar"),
			rootEvents: []*history.HistoryEvent{
				events["started"],
			},
			childEvents: []*history.HistoryEvent{},
			expectedState: &WorkflowExecutionState{
				Execution:     &common.WorkflowExecution{WorkflowId: "foo", RunId: "bar"},
				Status:        enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
				Type:          &common.WorkflowType{Name: "foo"},
				Attempt:       1,
				LastEventId:   1,
				HistoryLength: 1,
			},
		},
		"child workflow started": {
			depth:        -1,
			initialState: NewWorkflowExecutionState("foo", "bar"),
			rootEvents: []*history.HistoryEvent{
				events["started"],
				events["child workflow initiated"],
				events["child workflow started"],
			},
			childEvents: []*history.HistoryEvent{
				events["workflow started child"],
			},
			expectedState: &WorkflowExecutionState{
				Execution:     &common.WorkflowExecution{WorkflowId: "foo", RunId: "bar"},
				Status:        enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
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
		},
		"not enough depth for child": {
			depth:        0,
			initialState: NewWorkflowExecutionState("foo", "bar"),
			rootEvents: []*history.HistoryEvent{
				events["started"],
				events["child workflow initiated"],
				events["child workflow started"],
			},
			childEvents: []*history.HistoryEvent{
				events["workflow started child"],
			},
			expectedState: &WorkflowExecutionState{
				Execution:     &common.WorkflowExecution{WorkflowId: "foo", RunId: "bar"},
				Status:        enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
				Type:          &common.WorkflowType{Name: "foo"},
				Attempt:       1,
				LastEventId:   52,
				HistoryLength: 52,
				ChildStates: []ExecutionState{
					&WorkflowExecutionState{
						// Child workflow doesn't have its own events (no LastEventId)
						Status: enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
						Type: &common.WorkflowType{
							Name: "baz",
						},
						Execution: &common.WorkflowExecution{
							WorkflowId: "childWfId",
							RunId:      "childRunId",
						},
					},
				},
			},
		},
	}

	for name, tt := range tests {
		s.Run(name, func() {
			client := &mocks.Client{}

			// Setup HistoryEventIterator mocks
			exec := tt.initialState.Execution
			s.SetWorkflowHistoryMocks(client, exec.GetWorkflowId(), exec.GetRunId(), tt.rootEvents)
			s.SetWorkflowHistoryMocks(client, "childWfId", "childRunId", tt.childEvents)

			// Setup DescribeWorkflowExecution mocks
			s.SetDescribeWorkflowMocks(client, exec.GetWorkflowId(), exec.GetRunId(), tt.rootEvents)
			s.SetDescribeWorkflowMocks(client, "childWfId", "childRunId", tt.childEvents)

			// Execute what we're testing
			iter, _ := GetWorkflowExecutionUpdates(s.ctx, client, exec.GetWorkflowId(), exec.GetRunId(), true, tt.depth, 5)

			// Iterate over all the updates to get the final state
			var update *WorkflowExecutionUpdate
			for iter.HasNext() {
				update, _ = iter.Next()
				// Update should never be nil
				if update == nil {
					assert.FailNow(s.T(), "Update cannot be nil")
				}
			}
			// Remove helper maps to make comparison easier
			// TODO: Find a way to make this comparison clearer and not depend on some adhoc cleanup (maybe using state.Equal?)
			cleanState := RemoveStateMaps(update.State)

			assert.Equal(s.T(), tt.expectedState, cleanState)
		})
	}
}

func (s *WorkflowExecutionUpdateSuite) Test_GetWorkflowExecutionUpdatesErrors() {
	client := &mocks.Client{}

	exec := NewWorkflowExecutionState("foo", "bar").Execution
	s.SetWorkflowHistoryMocks(client, exec.GetWorkflowId(), exec.GetRunId(), []*history.HistoryEvent{events["started"]})

	// Setup DescribeWorkflowExecution mocks
	s.SetDescribeWorkflowErrorMocks(client, exec.GetWorkflowId(), exec.GetRunId(), fmt.Errorf("hey, I'm an error"))

	// Execute what we're testing
	iter, updateErr := GetWorkflowExecutionUpdates(s.ctx, client, exec.GetWorkflowId(), exec.GetRunId(), true, 0, 5)
	require.NoError(s.T(), updateErr)

	// Only call once since we only want to check on the first error.
	iter.HasNext()
	update, err := iter.Next()

	require.Nil(s.T(), update)
	require.Error(s.T(), err)
}
