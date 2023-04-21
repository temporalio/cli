package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"github.com/pborman/uuid"
	"github.com/temporalio/cli/tests/workflows/helloworld"
	"github.com/temporalio/cli/tests/workflows/update"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	testTq = "test-queue"
)

func (s *e2eSuite) TestWorkflowShow_ReplayableHistory() {
	c := s.ts.Client()

	s.NewWorker(testTq, func(r worker.Registry) {
		r.RegisterWorkflow(helloworld.Workflow)
		r.RegisterActivity(helloworld.Activity)
	})

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
	err = s.app.Run([]string{"", "workflow", "show", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--output", "json"})
	s.NoError(err)

	// save history to file
	historyFile := uuid.New() + ".json"
	logs := s.writer.GetContent()
	err = ioutil.WriteFile(historyFile, []byte(logs), 0644)
	s.NoError(err)
	defer os.Remove(historyFile)

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(helloworld.Workflow)
	err = replayer.ReplayWorkflowHistoryFromJSONFile(nil, historyFile)
	s.NoError(err)
}

func (s *e2eSuite) TestWorkflowUpdate() {
	c := s.ts.Client()

	s.NewWorker(testTq, func(r worker.Registry) {
		r.RegisterWorkflow(update.Counter)
	})
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
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", strconv.Itoa(randomInt)})
	s.NoError(err)
	want := fmt.Sprintf(": %v", randomInt)
	s.Contains(s.writer.GetContent(), want)

	// successful update with wait policy Completed, passing first-execution-run-id
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1", "--first-execution-run-id", wfr.GetRunID()})
	s.NoError(err)

	// update rejected, when name or update-id is not available
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", "non-existent-ID", "--run-id", wfr.GetRunID(), "-i", "1"})
	s.Error(err)

	// update rejected, wrong workflowID
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", "non-existent-ID", "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1"})
	s.Error(err)

	// update rejected, wrong update name
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", "non-existent-name", "-i", "1"})
	s.Error(err)

}
