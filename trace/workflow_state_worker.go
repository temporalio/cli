package trace

import (
	"context"
	"errors"
	"fmt"
	"github.com/alitto/pond"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	sdkclient "go.temporal.io/sdk/client"
)

// WorkflowStateJob implements a WorkerJob to retrieve updates for a WorkflowExecutionState and its child workflows.
type WorkflowStateJob struct {
	ctx    context.Context
	client sdkclient.Client

	state      *WorkflowExecutionState
	depth      int
	fetchAll   bool
	childJobs  []*WorkflowStateJob
	isUpToDate bool

	updateChan chan struct{}
}

// NewWorkflowStateJob returns a new WorkflowStateJob. It requires an updateChan to signal when there's updates.
func NewWorkflowStateJob(ctx context.Context, client sdkclient.Client, state *WorkflowExecutionState, fetchAll bool, depth int, updateChan chan struct{}) (*WorkflowStateJob, error) {
	if state == nil {
		return nil, errors.New("workflow state cannot be nil for a workflow state job")
	}
	if updateChan == nil {
		return nil, errors.New("updateChan cannot be nil for a workflow state job")
	}

	return &WorkflowStateJob{
		ctx:        ctx,
		client:     client,
		state:      state,
		depth:      depth,
		fetchAll:   fetchAll,
		updateChan: updateChan,
		childJobs:  []*WorkflowStateJob{},
	}, nil
}

// Run starts the WorkflowStateJob, which retrieves the workflow's events and spawns new jobs for the child workflows once it's up-to-date.
// New jobs are submitted to the pool when the job is up-to-date to reduce the amount of unnecessary history fetches (e.g. when the child workflow is already completed).
func (job *WorkflowStateJob) Run(group *pond.TaskGroupWithContext) func() error {
	return func() error {
		state := job.state
		wfId := state.Execution.GetWorkflowId()
		runId := state.Execution.GetRunId()

		// Get workflow's history length, so we know when we're up-to-date.
		if execInfo, err := GetWorkflowExecutionInfo(job.ctx, job.client, state.Execution.GetWorkflowId(), state.Execution.GetRunId()); err != nil {
			return err
		} else {
			state.HistoryLength = execInfo.HistoryLength
			state.IsArchived = execInfo.HistoryLength == 0 // TODO: Find a better way to identify archived workflows
		}

		// Make sure to not long poll archived workflows since GetWorkflowHistory fails under those circumstances.
		isLongPoll := !state.IsArchived
		historyIterator := job.client.GetWorkflowHistory(job.ctx, wfId, runId, isLongPoll, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

		for historyIterator.HasNext() {
			event, err := historyIterator.Next()
			if err != nil {
				return err
			}
			if event == nil {
				continue
			}

			// Update state with new event and signal
			job.state.Update(event)
			job.updateChan <- struct{}{}

			// Create child jobs if we're on a child workflow execution started event and we haven't hit depth 0.
			if event.EventType == enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED && job.depth != 0 {
				childJob, err := job.GetChildJob(event)
				if err != nil {
					// TODO: Consider if we want to error out if a child workflow cannot be updated by itself.
					return err
				}
				job.childJobs = append(job.childJobs, childJob)

				if job.isUpToDate {
					group.Submit(childJob.Run(group))
				}
			}

			// Start child jobs when we're up-to-date
			if !job.isUpToDate && event.EventId >= state.HistoryLength {
				job.isUpToDate = true
				for _, childJob := range job.childJobs {
					if childJob.ShouldStart() {
						group.Submit(childJob.Run(group))
					}
				}
			}
		}

		return nil
	}
}

// GetChildJob gets a new child job and appends it to the list of childJobs. These jobs don't start until the parent is catched up.
func (job *WorkflowStateJob) GetChildJob(event *history.HistoryEvent) (*WorkflowStateJob, error) {
	// Retrieve child workflow from parent and create a new job to fetch events for it
	childAttrs := event.GetChildWorkflowExecutionStartedEventAttributes()
	wf, ok := job.state.GetChildWorkflowByEventId(childAttrs.GetInitiatedEventId())
	if !ok {
		exec := childAttrs.GetWorkflowExecution()
		return nil, fmt.Errorf("child workflow (%s, %s) initiated in event %d not found in parent workflow's events", exec.GetWorkflowId(), exec.GetRunId(), childAttrs.GetInitiatedEventId())
	}

	// Create child job
	childJob, err := NewWorkflowStateJob(job.ctx, job.client, wf, job.fetchAll, job.depth-1, job.updateChan)
	if err != nil {
		return nil, err
	}
	return childJob, nil
}

// ShouldStart will return true if the state is in a status that requires requesting its event history.
// This will help reduce the amount of event histories requested when they're not needed.
func (job *WorkflowStateJob) ShouldStart() bool {
	switch job.state.Status {
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
		enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
		enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return true
	default:
		return job.fetchAll
	}
}
