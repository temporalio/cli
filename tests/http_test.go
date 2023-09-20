package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	liteconfig "github.com/temporalio/cli/server/config"
	"github.com/temporalio/cli/tests/workflows/helloworld"
	"go.temporal.io/sdk/worker"
)

func (s *e2eSuite) TestHTTP() {
	s.T().Parallel()

	// Something else during tests is binding on the gRPC - 1 causing clash with
	// this so we just add 1000 to be safe
	httpPort := strconv.Itoa(liteconfig.NewPortProvider().MustGetFreePort() + 1000)
	testServer, _, _ := s.setUpTestEnvironment("--http-port", httpPort)
	defer func() { _ = testServer.Stop() }()
	httpRoot := "http://127.0.0.1:" + httpPort + "/api/v1/namespaces/default"

	// Start worker
	w := s.newWorker(testServer, testTq, func(r worker.Registry) {
		r.RegisterWorkflow(helloworld.Workflow)
		r.RegisterActivity(helloworld.Activity)
	})
	defer w.Stop()

	// Start workflow
	withRunID := struct {
		RunID string `json:"runId"`
	}{}
	s.httpJSON("POST", httpRoot+"/workflows/my-workflow-id", `{
		"workflowType": { "name": "Workflow" },
		"taskQueue": { "name": "`+testTq+`" },
		"input": ["Temporal"]
	}`, &withRunID)
	s.Require().NotEmpty(withRunID.RunID)

	// Get result
	history := struct {
		History struct {
			Events []struct {
				Attrs struct {
					Result []string `json:"result"`
				} `json:"workflowExecutionCompletedEventAttributes"`
			} `json:"events"`
		} `json:"history"`
	}{}
	s.httpJSON(
		"GET",
		httpRoot+"/workflows/my-workflow-id/history?waitNewEvent=true&historyEventFilterType=2",
		"",
		&history,
	)
	s.Require().Equal("Hello Temporal!", history.History.Events[0].Attrs.Result[0])
}

func (s *e2eSuite) httpJSON(method, url, body string, result any) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	s.Require().NoError(err)
	s.Require().NoError(json.Unmarshal(respBody, result))
}
