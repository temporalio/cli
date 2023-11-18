package tests

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pborman/uuid"
	"github.com/temporalio/cli/tests/workflows/awaitsignal"
	"github.com/temporalio/cli/tests/workflows/encodejson"
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

func (s *e2eSuite) TestWorkflowExecute_Input() {
	s.T().Parallel()

	server, cli, _ := s.setUpTestEnvironment()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	w := s.newWorker(server, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(encodejson.Workflow)
	})
	defer w.Stop()

	// Run the workflow to completion using the CLI.  (TODO: We unfortunately
	// don't have a way to check the CLI output directly to make sure it prints
	// the right result...)
	err := cli.Run([]string{"", "workflow", "execute",
		"--input", `1`, "--input", `"two"`, "--input", `{"three": 3}`,
		"--input", `["a", "b", "c"]`,
		"--type", "Workflow", "--task-queue", testTq, "--workflow-id", "test"})
	s.NoError(err)

	// Check that the workflow produced the result we expect--if it did, that
	// means the CLI passed the arguments correctly.
	var result interface{}
	wf := client.GetWorkflow(context.Background(), "test", "")
	err = wf.Get(context.Background(), &result)
	s.NoError(err)

	s.Assert().Equal(`[1,"two",{"three":3},["a","b","c"]]`, result)
}

func (s *e2eSuite) TestWorkflowExecute_InputFile() {
	s.T().Parallel()

	tempDir := s.T().TempDir()
	argFiles := []string{
		filepath.Join(tempDir, "arg1.json"), filepath.Join(tempDir, "arg2.json"),
		filepath.Join(tempDir, "arg3.json"), filepath.Join(tempDir, "arg4.json"),
	}
	s.NoError(os.WriteFile(argFiles[0], []byte("1"), 0700))
	s.NoError(os.WriteFile(argFiles[1], []byte(`"two"`), 0700))
	s.NoError(os.WriteFile(argFiles[2], []byte(`{"three": 3}`), 0700))
	s.NoError(os.WriteFile(argFiles[3], []byte(`["a", "b", "c"]`), 0700))

	server, cli, _ := s.setUpTestEnvironment()
	defer func() { _ = server.Stop() }()

	client := server.Client()

	w := s.newWorker(server, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(encodejson.Workflow)
	})
	defer w.Stop()

	// Run the workflow to completion using the CLI.  (TODO: We unfortunately
	// don't have a way to check the CLI output directly to make sure it prints
	// the right result...)
	err := cli.Run([]string{"", "workflow", "execute",
		"--input-file", argFiles[0], "--input-file", argFiles[1],
		"--input-file", argFiles[2], "--input-file", argFiles[3],
		"--type", "Workflow", "--task-queue", testTq, "--workflow-id", "test"})
	s.NoError(err)

	// Check that the workflow produced the result we expect--if it did, that
	// means the CLI passed the arguments correctly.
	var result interface{}
	wf := client.GetWorkflow(context.Background(), "test", "")
	err = wf.Get(context.Background(), &result)
	s.NoError(err)

	s.Assert().Equal(`[1,"two",{"three":3},["a","b","c"]]`, result)
}

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
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "30", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", strconv.Itoa(randomInt)})
	s.NoError(err)
	want := fmt.Sprintf(": %v", randomInt)
	s.Contains(writer.GetContent(), want)

	// successful update with wait policy Completed, passing first-execution-run-id
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "30", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1", "--first-execution-run-id", wfr.GetRunID()})
	s.NoError(err)

	// update rejected, when name is not available
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "30", "--workflow-id", "non-existent-ID", "--run-id", wfr.GetRunID(), "-i", "1"})
	s.ErrorContains(err, "Required flag \"name\" not set")

	// update rejected, wrong workflowID
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "30", "--workflow-id", "non-existent-ID", "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1"})
	s.ErrorContains(err, "unable to update workflow")

	// update rejected, wrong update name
	err = app.Run([]string{"", "workflow", "update", "--context-timeout", "30", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", "non-existent-name", "-i", "1"})
	s.ErrorContains(err, "unable to update workflow")
}

func (s *e2eSuite) TestWorkflowCancel_Batch() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	w := s.newWorker(testserver, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(awaitsignal.Workflow)
	})
	defer w.Stop()

	c := testserver.Client()

	ids := []string{"1", "2", "3"}
	for _, id := range ids {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			sdkclient.StartWorkflowOptions{ID: id, TaskQueue: testTq},
			awaitsignal.Workflow,
		)
		s.NoError(err)
	}

	err := app.Run([]string{"", "workflow", "cancel", "--query", "WorkflowId = '1' OR WorkflowId = '2'", "--reason", "test", "--yes", "--namespace", testNamespace})
	s.NoError(err)

	awaitTaskQueuePoller(s, c, testTq)
	awaitBatchJob(s, c, testNamespace)

	s.Eventually(func() bool {
		w1 := c.GetWorkflowHistory(context.Background(), "1", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		if expected := checkForEventType(w1, enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED); !expected {
			return false
		}

		w2 := c.GetWorkflowHistory(context.Background(), "2", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		if expected := checkForEventType(w2, enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED); !expected {
			return false
		}

		w3 := c.GetWorkflowHistory(context.Background(), "3", "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		if expected := !checkForEventType(w3, enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED); !expected {
			return false
		}

		return true
	}, 10*time.Second, time.Second, "timed out awaiting for workflows cancellation")
}

func (s *e2eSuite) TestWorkflowSignal_Batch() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	defer s.newWorker(testserver, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(awaitsignal.Workflow)
	}).Stop()

	c := testserver.Client()

	ids := []string{"1", "2", "3"}
	for _, id := range ids {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			sdkclient.StartWorkflowOptions{ID: id, TaskQueue: testTq},
			awaitsignal.Workflow,
		)
		s.NoError(err)
	}

	s.Eventually(func() bool {
		wfs, err := c.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: testNamespace,
		})
		if err != nil {
			return false
		}
		return len(wfs.GetExecutions()) == 3
	}, 10*time.Second, time.Second)

	i1 := "\"" + awaitsignal.Input1 + "\""
	q := "WorkflowId = '1' OR WorkflowId = '2'"
	err := app.Run([]string{"", "workflow", "signal", "--name", awaitsignal.Done, "--input", i1, "--query", q, "--reason", "test", "--yes", "--namespace", testNamespace})
	s.NoError(err)

	awaitTaskQueuePoller(s, c, testTq)
	awaitBatchJob(s, c, testNamespace)

	s.Eventually(func() bool {
		wfs, err := c.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: testNamespace,
		})
		s.NoError(err)

		for _, wf := range wfs.GetExecutions() {
			switch wf.GetExecution().GetWorkflowId() {
			case "1", "2":
				if wf.GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_COMPLETED {
					return false
				}
			case "3":
				if wf.GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
					return false
				}
			}
		}

		return true
	}, 3*time.Second, time.Second, "timed out awaiting for workflows completion after signal")
}

func (s *e2eSuite) TestWorkflowTerminate_Batch() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	w := s.newWorker(testserver, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(awaitsignal.Workflow)
	})
	defer w.Stop()

	c := testserver.Client()

	ids := []string{"1", "2", "3"}
	for _, id := range ids {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			sdkclient.StartWorkflowOptions{ID: id, TaskQueue: testTq},
			awaitsignal.Workflow,
		)
		s.NoError(err)
	}

	s.Eventually(func() bool {
		wfs, err := c.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: testNamespace,
		})
		if err != nil {
			return false
		}
		return len(wfs.GetExecutions()) == 3
	}, 10*time.Second, time.Second)

	err := app.Run([]string{"", "workflow", "terminate", "--query", "WorkflowId = '1' OR WorkflowId = '2'", "--reason", "test", "--yes", "--namespace", testNamespace})
	s.NoError(err)

	awaitTaskQueuePoller(s, c, testTq)
	awaitBatchJob(s, c, testNamespace)

	s.Eventually(func() bool {
		wfs, err := c.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: testNamespace,
		})
		s.NoError(err)

		for _, wf := range wfs.GetExecutions() {
			switch wf.GetExecution().GetWorkflowId() {
			case "1", "2":
				if wf.GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_TERMINATED {
					return false
				}
			case "3":
				if wf.GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
					return false
				}
			}
		}

		return true
	}, 10*time.Second, time.Second, "timed out awaiting for workflows termination")
}

func (s *e2eSuite) TestWorkflowDelete_Batch() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	w := s.newWorker(testserver, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(awaitsignal.Workflow)
	})
	defer w.Stop()

	c := testserver.Client()

	ids := []string{"1", "2", "3"}
	for _, id := range ids {
		_, err := c.ExecuteWorkflow(
			context.Background(),
			sdkclient.StartWorkflowOptions{ID: id, TaskQueue: testTq},
			awaitsignal.Workflow,
		)
		s.NoError(err)
	}

	err := app.Run([]string{"", "workflow", "delete", "--query", "WorkflowId = '1' OR WorkflowId = '2'", "--reason", "test", "--yes", "--namespace", testNamespace})
	s.NoError(err)

	awaitTaskQueuePoller(s, c, testTq)
	awaitBatchJob(s, c, testNamespace)

	s.Eventually(func() bool {
		wfs, err := c.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: testNamespace,
		})
		s.NoError(err)

		if len(wfs.GetExecutions()) == 1 && wfs.GetExecutions()[0].GetExecution().GetWorkflowId() == "3" {
			return true
		}

		return false
	}, 10*time.Second, time.Second, "timed out awaiting for workflows termination")
}

// awaitTaskQueuePoller used mostly for more explicit failure message
func awaitTaskQueuePoller(s *e2eSuite, c sdkclient.Client, taskqueue string) {
	s.Eventually(func() bool {
		resp, err := c.DescribeTaskQueue(context.Background(), taskqueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
		if err != nil {
			return false
		}
		return len(resp.GetPollers()) > 0
	}, 10*time.Second, time.Second, "no worker started for taskqueue "+taskqueue)
}

func awaitBatchJob(s *e2eSuite, c sdkclient.Client, ns string) {
	s.Eventually(func() bool {
		resp, err := c.WorkflowService().ListBatchOperations(context.Background(),
			&workflowservice.ListBatchOperationsRequest{Namespace: ns})
		if err != nil {
			return false
		}
		if len(resp.OperationInfo) == 0 {
			return false
		}

		batchJob, err := c.WorkflowService().DescribeBatchOperation(context.Background(),
			&workflowservice.DescribeBatchOperationRequest{
				JobId:     resp.OperationInfo[0].JobId,
				Namespace: ns,
			})
		if err != nil {
			return false
		}

		return batchJob.State == enums.BATCH_OPERATION_STATE_COMPLETED
	}, 10*time.Second, time.Second, "cancellation batch job timed out")
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
