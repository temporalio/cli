package tests

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/pborman/uuid"
	"github.com/temporalio/cli/app"
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
	s.app = app.BuildApp()
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

	// update
	//TODO: uncomment when update is enabled in the server by default, currently there is no way to enable it to test the command
	//err = s.app.Run([]string{"", "workflow", "update", "--workflow-id", wfr.GetID(), "--run-id", wfr.GetRunID(), "--namespace", "default", "--name", update.FetchAndAdd, "-i", `"1"`, "--wait-policy", "Completed"})
	//s.NoError(err)
}
