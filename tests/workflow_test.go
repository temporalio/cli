package tests

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/pborman/uuid"
	"github.com/temporalio/cli/tests/workflows/helloworld"
	"github.com/temporalio/cli/tests/workflows/update"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	testTq        = "test-queue"
	testNamespace = "default"
)

func (s *e2eSuite) TestWorkflowShow_ReplayableHistory() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	w := s.newWorker(testserver, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(helloworld.Workflow)
		r.RegisterActivity(helloworld.Activity)
	})
	defer w.Stop()

	wfr, err := c.ExecuteWorkflow(
		context.Background(),
		sdkclient.StartWorkflowOptions{TaskQueue: testTq},
		helloworld.Workflow,
		"world",
	)
	s.NoError(err)

	var result string
	err = wfr.Get(context.Background(), &result)
	s.NoError(err)

	// show history
	err = app.Run([]string{"", "workflow", "show", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--output", "json"})
	s.NoError(err)

	// save history to file
	historyFile := uuid.New() + ".json"
	err = os.WriteFile(historyFile, []byte(writer.GetContent()), 0644)
	s.NoError(err)
	defer os.Remove(historyFile)

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(helloworld.Workflow)
	err = replayer.ReplayWorkflowHistoryFromJSONFile(nil, historyFile)
	s.NoError(err)
}

func (s *e2eSuite) TestWorkflowUpdate() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	w := s.newWorker(testserver, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(update.Counter)
	})
	defer w.Stop()

	randomInt := rand.Intn(100)
	wfr, err := c.ExecuteWorkflow(
		context.Background(),
		sdkclient.StartWorkflowOptions{TaskQueue: testTq},
		update.Counter,
		randomInt,
	)
	s.NoError(err)
	signalWorkflow := func() {
		// send a Signal to stop the workflow
		err = c.SignalWorkflow(context.Background(), wfr.GetID(), wfr.GetRunID(), update.Done, nil)
		s.NoError(err)
	}

	defer signalWorkflow()

	// successful update with wait policy Completed, should show the result
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "10", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", strconv.Itoa(randomInt)})
	s.NoError(err)
	want := fmt.Sprintf(": %v", randomInt)
	s.Contains(writer.GetContent(), want)

	// successful update with wait policy Completed, passing first-execution-run-id
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "10", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1", "--first-execution-run-id", wfr.GetRunID()})
	s.NoError(err)

	// update rejected, when name is not available
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "10", "--workflow-id", "non-existent-ID", "--run-id", wfr.GetRunID(), "-i", "1"})
	s.ErrorContains(err, "Required flag \"name\" not set")

	// update rejected, wrong workflowID
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "10", "--workflow-id", "non-existent-ID", "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1"})
	s.ErrorContains(err, "update workflow failed")

	// update rejected, wrong update name
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "10", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", "non-existent-name", "-i", "1"})
	s.ErrorContains(err, "update workflow failed: unknown update")
}

func (s *e2eSuite) TestWorkflowCancel_Batch() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	ids := []string{"1", "2", "3"}

	for _, id := range ids {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			sdkclient.StartWorkflowOptions{ID: id, TaskQueue: testTq},
			"non-existing-workflow-type",
		)
		s.NoError(err)
	}

	err := app.Run([]string{"", "workflow", "cancel", "--query", "WorkflowId = '1' OR WorkflowId = '2'", "--reason", "test", "--yes", "--namespace", testNamespace})
	s.NoError(err)

	s.Eventually(func() bool {
		resp, err := c.WorkflowService().ListBatchOperations(context.Background(),
			&workflowservice.ListBatchOperationsRequest{Namespace: testNamespace})
		if err != nil {
			return false
		}
		if len(resp.OperationInfo) == 0 {
			return false
		}

		batchJob, err := c.WorkflowService().DescribeBatchOperation(context.Background(),
			&workflowservice.DescribeBatchOperationRequest{
				JobId:     resp.OperationInfo[0].JobId,
				Namespace: testNamespace,
			})
		if err != nil {
			return false
		}

		return batchJob.State == enums.BATCH_OPERATION_STATE_COMPLETED
	}, time.Second, 10*time.Second)

	w1 := c.GetWorkflowHistory(context.Background(), "1", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	s.True(checkForEventType(w1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED), "Workflow 1 should have a cancellation event")

	w2 := c.GetWorkflowHistory(context.Background(), "2", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	s.True(checkForEventType(w2, enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED), "Workflow 2 should have a cancellation event")

	w3 := c.GetWorkflowHistory(context.Background(), "3", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	s.False(checkForEventType(w3, enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED), "Workflow 3 should not have a cancellation event")
}

func (s *e2eSuite) TestWorkflowSignal_Batch() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	ids := []string{"1", "2", "3"}

	for _, id := range ids {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			sdkclient.StartWorkflowOptions{ID: id, TaskQueue: testTq},
			"non-existing-workflow-type",
		)
		s.NoError(err)
	}

	err := app.Run([]string{"", "workflow", "signal", "--input", "\"testvalue\"", "--name", "test-signal", "--query", "WorkflowId = '1' OR WorkflowId = '2'", "--reason", "test", "--yes", "--namespace", testNamespace})
	s.NoError(err)

	s.Eventually(func() bool {
		resp, err := c.WorkflowService().ListBatchOperations(context.Background(),
			&workflowservice.ListBatchOperationsRequest{Namespace: testNamespace})
		if err != nil {
			return false
		}
		if len(resp.OperationInfo) == 0 {
			return false
		}

		batchJob, err := c.WorkflowService().DescribeBatchOperation(context.Background(),
			&workflowservice.DescribeBatchOperationRequest{
				JobId:     resp.OperationInfo[0].JobId,
				Namespace: testNamespace,
			})
		if err != nil {
			return false
		}

		return batchJob.State == enums.BATCH_OPERATION_STATE_COMPLETED
	}, time.Second, 10*time.Second)

	w1 := c.GetWorkflowHistory(context.Background(), "1", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	s.True(checkForEventType(w1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED), "Workflow 1 should have received a signal")

	w2 := c.GetWorkflowHistory(context.Background(), "2", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	s.True(checkForEventType(w2, enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED), "Workflow 2 should have received a signal")

	w3 := c.GetWorkflowHistory(context.Background(), "3", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	s.False(checkForEventType(w3, enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED), "Workflow 3 should not have received a signal")
}

func (s *e2eSuite) TestWorkflowTerminate_Batch() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	ids := []string{"1", "2", "3"}

	for _, id := range ids {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			sdkclient.StartWorkflowOptions{ID: id, TaskQueue: testTq},
			"non-existing-workflow-type",
		)
		s.NoError(err)
	}

	err := app.Run([]string{"", "workflow", "terminate", "--query", "WorkflowId = '1' OR WorkflowId = '2'", "--reason", "test", "--yes", "--namespace", testNamespace})
	s.NoError(err)

	s.Eventually(func() bool {
		resp, err := c.WorkflowService().ListBatchOperations(context.Background(),
			&workflowservice.ListBatchOperationsRequest{Namespace: testNamespace})
		if err != nil {
			return false
		}
		if len(resp.OperationInfo) == 0 {
			return false
		}

		batchJob, err := c.WorkflowService().DescribeBatchOperation(context.Background(),
			&workflowservice.DescribeBatchOperationRequest{
				JobId:     resp.OperationInfo[0].JobId,
				Namespace: testNamespace,
			})
		if err != nil {
			return false
		}

		return batchJob.State == enums.BATCH_OPERATION_STATE_COMPLETED
	}, time.Second, 10*time.Second)

	w1, err := c.DescribeWorkflowExecution(context.Background(), "1", "")
	s.NoError(err)
	s.Equal(enums.WORKFLOW_EXECUTION_STATUS_TERMINATED, w1.GetWorkflowExecutionInfo().GetStatus())

	w2, err := c.DescribeWorkflowExecution(context.Background(), "2", "")
	s.NoError(err)
	s.Equal(enums.WORKFLOW_EXECUTION_STATUS_TERMINATED, w2.GetWorkflowExecutionInfo().GetStatus())

	w3, err := c.DescribeWorkflowExecution(context.Background(), "3", "")
	s.NoError(err)
	s.Equal(enums.WORKFLOW_EXECUTION_STATUS_RUNNING, w3.GetWorkflowExecutionInfo().GetStatus())
}

func checkForEventType(events sdkclient.HistoryEventIterator, eventType enums.EventType) bool {
	for events.HasNext() {
		event, err := events.Next()
		if err != nil {
			break
		}
		if event.GetEventType() == eventType {
			return true
		}
	}
	return false
}
