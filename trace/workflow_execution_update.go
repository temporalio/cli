package trace

import (
	"context"
	"fmt"
	"github.com/alitto/pond"
	sdkclient "go.temporal.io/sdk/client"
)

type WorkflowExecutionUpdate struct {
	State *WorkflowExecutionState
}

// WorkflowExecutionUpdateIterator is the interface the provides iterative updates, analogous to the HistoryEventIterator interface.
type WorkflowExecutionUpdateIterator interface {
	HasNext() bool
	Next() (*WorkflowExecutionUpdate, error)
}

// WorkflowExecutionUpdateIteratorImpl implements the iterator interface. Keeps information about the last processed update and
// receives new updates through the updateChan channel.
type WorkflowExecutionUpdateIteratorImpl struct {
	updated    bool
	nextUpdate *WorkflowExecutionUpdate
	state      *WorkflowExecutionState
	err        error
	updateChan <-chan struct{}
	doneChan   <-chan struct{}
	errorChan  <-chan error
}

// GetWorkflowExecutionUpdates gets workflow execution updates for a particular workflow
// - workflow ID of the workflow
// - runID can be default (empty string)
// - depth of child workflows to request updates for (-1 for unlimited depth)
// - concurrency of requests (non-zero positive integer)
// Returns iterator (see client.GetWorkflowHistory) that provides updated WorkflowExecutionState snapshots.
// Example:
// To print a workflow's state whenever there's updates
//
//	iter := GetWorkflowExecutionUpdates(ctx, client, wfId, runId, -1, 5)
//	var state *WorkflowExecutionState
//	for iter.HasNext() {
//		update = iter.Next()
//		PrintWorkflowState(update.State)
//	}
func GetWorkflowExecutionUpdates(ctx context.Context, client sdkclient.Client, wfId, runId string, fetchAll bool, depth int, concurrency int) (WorkflowExecutionUpdateIterator, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("invalid value for concurrency (expected non-zero positive integer, got %d)", concurrency)
	}

	pool := pond.New(concurrency, 1000)

	// Using a GroupContext we can use pond's error handling and context cancelling.
	group, poolCtx := pool.GroupContext(ctx)
	updateChan := make(chan struct{})
	doneChan := make(chan struct{})
	errorChan := make(chan error)

	state := NewWorkflowExecutionState(wfId, runId)
	job, err := NewWorkflowStateJob(poolCtx, client, state, fetchAll, depth, updateChan)
	if err != nil {
		return nil, err
	}

	go func() {
		group.Submit(job.Run(group))
		// Wait for all tasks to complete and signal done
		if err = group.Wait(); err != nil {
			errorChan <- err
		} else {
			doneChan <- struct{}{}
		}
	}()

	return &WorkflowExecutionUpdateIteratorImpl{
		updateChan: updateChan,
		doneChan:   doneChan,
		errorChan:  errorChan,
		state:      state,
	}, nil
}

// HasNext checks if there's any more updates in the updateChan channel. Returns false if the channel has been closed.
func (iter *WorkflowExecutionUpdateIteratorImpl) HasNext() bool {
	iter.updated = true
	select {
	case <-iter.updateChan:
		iter.nextUpdate = &WorkflowExecutionUpdate{State: iter.state}
		return true
	case <-iter.doneChan:
		return false
	case err := <-iter.errorChan:
		iter.err = err
		return true
	}
}

// Next return the last processed execution update. HasNext has to be called first (following the HasNext/Next pattern).
func (iter *WorkflowExecutionUpdateIteratorImpl) Next() (*WorkflowExecutionUpdate, error) {
	// Make sure HasNext() has been called first.
	if !iter.updated {
		return nil, fmt.Errorf("please call HasNext() first")
	}
	iter.updated = false

	if err := iter.err; err != nil {
		iter.err = nil
		return nil, err
	}

	update := iter.nextUpdate
	iter.nextUpdate = nil
	return update, nil
}
