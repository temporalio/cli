package tests

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/pborman/uuid"
	"github.com/temporalio/cli/tests/workflows/helloworld"
	"go.temporal.io/sdk/client"
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
		client.StartWorkflowOptions{TaskQueue: testTq},
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
