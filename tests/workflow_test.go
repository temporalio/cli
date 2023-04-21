package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"

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

	wfr, err := c.ExecuteWorkflow(
		context.Background(),
		sdkclient.StartWorkflowOptions{TaskQueue: testTq},
		update.Counter,
	)
	s.NoError(err)
	signalWorkflow := func() {
		// send a Signal to stop the workflow
		err = c.SignalWorkflow(context.Background(), wfr.GetID(), wfr.GetRunID(), update.Done, nil)
		s.NoError(err)
	}

	defer signalWorkflow()

	// successful update with wait policy Completed, should show the result
	rand.Seed(time.Now().UnixNano())
	randomInt := strconv.Itoa(rand.Intn(100))
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", randomInt})
	s.NoError(err)
	s.Contains(s.writer.GetContent(), ": 0")

	// successful update with wait policy Completed, make sure previous val is returned and printed
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1"})
	s.NoError(err)
	want := fmt.Sprintf(": %s", randomInt)
	s.Contains(s.writer.GetContent(), want)

	// successful update with wait policy Completed, passing first-execution-run-id
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1", "--first-execution-run-id", wfr.GetRunID()})
	s.NoError(err)

	// update rejected, wrong workflowID
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", "non-existent-ID", "--run-id", wfr.GetRunID(), "--name", update.FetchAndAdd, "-i", "1"})
	s.Error(err)

	// update rejected, wrong update name
	err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--name", "non-existent-name", "-i", "1"})
	s.Error(err)

}
