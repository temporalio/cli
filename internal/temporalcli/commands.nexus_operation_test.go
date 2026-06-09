package temporalcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	nexus "github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) setupNexusEndpointAndWorker(t *testing.T) (string, *DevWorker) {
	handlerWorkflow := func(ctx workflow.Context, input string) (string, error) {
		return "got: " + input, nil
	}
	return s.setupNexusEndpointWithWorkflow(t, handlerWorkflow, "handler-")
}

func (s *SharedServerSuite) setupNexusEndpointWithWorkflow(
	t *testing.T,
	handlerWorkflow func(workflow.Context, string) (string, error),
	workflowIDPrefix string,
) (string, *DevWorker) {
	endpointName := "test-ep-" + uuid.NewString()[:8]

	op := temporalnexus.NewWorkflowRunOperation(
		"test-op",
		handlerWorkflow,
		func(ctx context.Context, input string, opts nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
			return client.StartWorkflowOptions{
				ID: workflowIDPrefix + opts.RequestID,
			}, nil
		},
	)
	svc := nexus.NewService("test-service")
	require.NoError(t, svc.Register(op))

	w := s.DevServer.StartDevWorker(t, DevWorkerOptions{
		Workflows:     []any{handlerWorkflow},
		NexusServices: []*nexus.Service{svc},
	})

	_, err := s.Client.OperatorService().CreateNexusEndpoint(s.Context, &operatorservice.CreateNexusEndpointRequest{
		Spec: &nexuspb.EndpointSpec{
			Name: endpointName,
			Target: &nexuspb.EndpointTarget{
				Variant: &nexuspb.EndpointTarget_Worker_{
					Worker: &nexuspb.EndpointTarget_Worker{
						Namespace: s.Namespace(),
						TaskQueue: w.Options.TaskQueue,
					},
				},
			},
		},
	})
	require.NoError(t, err)

	return endpointName, w
}

func (s *SharedServerSuite) TestNexusOperationStart() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "start-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Started Nexus Operation")
	s.Contains(res.Stdout.String(), opID)
}

func (s *SharedServerSuite) TestNexusOperationExecute() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "exec-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "execute",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "got: hello")
}

func (s *SharedServerSuite) TestNexusOperationExecute_JSON() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "exec-json-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "execute",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
		"--output", "json",
	)
	s.NoError(res.Err)

	var result struct {
		OperationId string          `json:"operationId"`
		RunId       string          `json:"runId"`
		Result      json.RawMessage `json:"result"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &result))
	s.Equal(opID, result.OperationId)
	s.Contains(string(result.Result), "got: hello")
}

func (s *SharedServerSuite) TestNexusOperationDescribe() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "desc-op-" + uuid.NewString()[:8]

	// Start an operation first.
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Describe it.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "describe",
			"--address", s.Address(),
			"--operation-id", opID,
		)
		return res.Err == nil
	}, 30*time.Second, 500*time.Millisecond)

	s.Contains(res.Stdout.String(), opID)
	s.Contains(res.Stdout.String(), endpointName)
	s.Contains(res.Stdout.String(), "test-service")
}

func (s *SharedServerSuite) TestNexusOperationDescribe_JSON() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "desc-json-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "describe",
			"--address", s.Address(),
			"--operation-id", opID,
			"--output", "json",
		)
		return res.Err == nil
	}, 30*time.Second, 500*time.Millisecond)

	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), opID)
}

func (s *SharedServerSuite) TestNexusOperationCancel() {
	blockingHandler := func(ctx workflow.Context, input string) (string, error) {
		ctx.Done().Receive(ctx, nil)
		return "", ctx.Err()
	}
	endpointName, w := s.setupNexusEndpointWithWorkflow(s.T(), blockingHandler, "cancel-handler-")
	defer w.Stop()

	opID := "cancel-op-" + uuid.NewString()[:8]

	// Start operation.
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Cancel it.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "cancel",
			"--address", s.Address(),
			"--operation-id", opID,
			"--reason", "testing cancellation",
		)
		return res.Err == nil
	}, 30*time.Second, 500*time.Millisecond)

	s.Contains(res.Stdout.String(), "Cancellation requested")
}

func (s *SharedServerSuite) TestNexusOperationTerminate() {
	blockingHandler := func(ctx workflow.Context, input string) (string, error) {
		ctx.Done().Receive(ctx, nil)
		return "", ctx.Err()
	}
	endpointName, w := s.setupNexusEndpointWithWorkflow(s.T(), blockingHandler, "term-handler-")
	defer w.Stop()

	opID := "term-op-" + uuid.NewString()[:8]

	// Start operation.
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Terminate it.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "terminate",
			"--address", s.Address(),
			"--operation-id", opID,
			"--reason", "testing termination",
		)
		return res.Err == nil
	}, 30*time.Second, 500*time.Millisecond)

	s.Contains(res.Stdout.String(), "Nexus Operation terminated")
}

func (s *SharedServerSuite) TestNexusOperationList() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	// Start a couple of operations.
	for i := 0; i < 2; i++ {
		opID := fmt.Sprintf("list-op-%d-%s", i, uuid.NewString()[:8])
		res := s.Execute(
			"nexus", "operation", "start",
			"--address", s.Address(),
			"--endpoint", endpointName,
			"--service", "test-service",
			"--operation", "test-op",
			"--operation-id", opID,
			"--input", `"hello"`,
		)
		s.NoError(res.Err)
	}

	// List operations — wait until both are visible.
	var res *CommandResult
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "list",
			"--address", s.Address(),
		)
		out := res.Stdout.String()
		return res.Err == nil &&
			strings.Contains(out, "list-op-0") &&
			strings.Contains(out, "list-op-1")
	}, 30*time.Second, 500*time.Millisecond)
}

func (s *SharedServerSuite) TestNexusOperationCount() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "count-op-" + uuid.NewString()[:8]
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Count operations.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "count",
			"--address", s.Address(),
		)
		return res.Err == nil && res.Stdout.String() != ""
	}, 30*time.Second, 500*time.Millisecond)

	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Total:")
}

func (s *SharedServerSuite) TestNexusOperationResult() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "result-op-" + uuid.NewString()[:8]

	// Start an operation.
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Get result (blocks until completed).
	res = s.Execute(
		"nexus", "operation", "result",
		"--address", s.Address(),
		"--operation-id", opID,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "got: hello")
}

func (s *SharedServerSuite) TestNexusOperationResult_JSON() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "result-json-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"nexus", "operation", "result",
		"--address", s.Address(),
		"--operation-id", opID,
		"--output", "json",
	)
	s.NoError(res.Err)

	var result struct {
		OperationId string          `json:"operationId"`
		Result      json.RawMessage `json:"result"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &result))
	s.Equal(opID, result.OperationId)
	s.Contains(string(result.Result), "got: hello")
}

func (s *SharedServerSuite) TestNexusOperationStart_JSON() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "start-json-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
		"--output", "json",
	)
	s.NoError(res.Err)

	var result struct {
		Endpoint    string `json:"endpoint"`
		Service     string `json:"service"`
		Operation   string `json:"operation"`
		OperationId string `json:"operationId"`
		RunId       string `json:"runId"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &result))
	s.Equal(endpointName, result.Endpoint)
	s.Equal("test-service", result.Service)
	s.Equal("test-op", result.Operation)
	s.Equal(opID, result.OperationId)
	s.NotEmpty(result.RunId)
}
func (s *SharedServerSuite) TestNexusOperationStart_OperationIDRequired() {
	res := s.Execute(
		"nexus", "operation", "start",
		"--endpoint", "test-ep",
		"--service", "test-service",
		"--operation", "test-op",
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "operation-id")
}

func (s *SharedServerSuite) TestNexusOperationStart_ScheduleToCloseTimeout() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "timeout-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--schedule-to-close-timeout", "30m",
		"--input", `"hello"`,
		"--output", "json",
	)
	s.NoError(res.Err)

	var result struct {
		OperationId string `json:"operationId"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &result))
	s.Equal(opID, result.OperationId)

	// Verify the timeout is reflected in describe.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "describe",
			"--address", s.Address(),
			"--operation-id", opID,
			"--output", "json",
		)
		return res.Err == nil
	}, 30*time.Second, 500*time.Millisecond)
	s.Contains(res.Stdout.String(), opID)
}

func (s *SharedServerSuite) TestNexusOperationStart_IdConflictPolicy() {
	// Use a blocking handler so the operation stays running across the three
	// start calls — IDConflictPolicy only triggers against an open operation.
	blockingHandler := func(ctx workflow.Context, input string) (string, error) {
		ctx.Done().Receive(ctx, nil)
		return "", ctx.Err()
	}
	endpointName, w := s.setupNexusEndpointWithWorkflow(s.T(), blockingHandler, "conflict-handler-")
	defer w.Stop()

	opID := "conflict-op-" + uuid.NewString()[:8]

	// Start first operation.
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Start again with same ID and UseExisting policy — should succeed.
	res = s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--id-conflict-policy", "UseExisting",
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Start again with same ID and Fail policy — should fail.
	res = s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--id-conflict-policy", "Fail",
		"--input", `"hello"`,
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestNexusOperationList_JSON() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "list-json-op-" + uuid.NewString()[:8]
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "list",
			"--address", s.Address(),
			"--output", "json",
		)
		return res.Err == nil && strings.Contains(res.Stdout.String(), opID)
	}, 30*time.Second, 500*time.Millisecond)
}

func (s *SharedServerSuite) TestNexusOperationList_Limit() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	// Start 3 operations.
	for i := 0; i < 3; i++ {
		opID := fmt.Sprintf("limit-op-%d-%s", i, uuid.NewString()[:8])
		res := s.Execute(
			"nexus", "operation", "start",
			"--address", s.Address(),
			"--endpoint", endpointName,
			"--service", "test-service",
			"--operation", "test-op",
			"--operation-id", opID,
			"--input", `"hello"`,
		)
		s.NoError(res.Err)
	}

	// List with limit=1, JSONL output so we can count entries.
	var res *CommandResult
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "list",
			"--address", s.Address(),
			"--limit", "1",
			"--output", "jsonl",
		)
		return res.Err == nil && len(res.Stdout.String()) > 0
	}, 30*time.Second, 500*time.Millisecond)

	// Count JSON objects — each on its own line.
	lines := 0
	for _, line := range strings.Split(strings.TrimSpace(res.Stdout.String()), "\n") {
		if len(strings.TrimSpace(line)) > 0 {
			lines++
		}
	}
	s.Equal(1, lines, "limit=1 should return exactly 1 operation")
}

func (s *SharedServerSuite) TestNexusOperationCount_JSON() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "count-json-op-" + uuid.NewString()[:8]
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "count",
			"--address", s.Address(),
			"--output", "json",
		)
		if res.Err != nil {
			return false
		}
		// Proto JSON encodes int64 as a string; accept either form.
		var result struct {
			Count json.Number `json:"count"`
		}
		if err := json.Unmarshal(res.Stdout.Bytes(), &result); err != nil {
			return false
		}
		n, err := result.Count.Int64()
		return err == nil && n > 0
	}, 30*time.Second, 500*time.Millisecond)
}

func (s *SharedServerSuite) TestNexusOperationCount_Query() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "count-query-op-" + uuid.NewString()[:8]
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// Count with a query that matches this specific operation.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "count",
			"--address", s.Address(),
			"--query", fmt.Sprintf(`OperationId = "%s"`, opID),
		)
		return res.Err == nil && res.Stdout.String() != ""
	}, 30*time.Second, 500*time.Millisecond)

	s.Contains(res.Stdout.String(), "Total:")
}

func (s *SharedServerSuite) TestNexusOperationList_Query() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "query-op-" + uuid.NewString()[:8]
	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
	)
	s.NoError(res.Err)

	// List with query filter — wait until the operation appears.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "list",
			"--address", s.Address(),
			"--query", fmt.Sprintf(`OperationId = "%s"`, opID),
		)
		return res.Err == nil && strings.Contains(res.Stdout.String(), opID)
	}, 30*time.Second, 500*time.Millisecond)
}

func (s *SharedServerSuite) TestNexusOperationStart_MissingRequiredFlags() {
	// Missing --endpoint
	res := s.Execute(
		"nexus", "operation", "start",
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", "some-id",
	)
	s.Error(res.Err)

	// Missing --service
	res = s.Execute(
		"nexus", "operation", "start",
		"--endpoint", "test-ep",
		"--operation", "test-op",
		"--operation-id", "some-id",
	)
	s.Error(res.Err)

	// Missing --operation
	res = s.Execute(
		"nexus", "operation", "start",
		"--endpoint", "test-ep",
		"--service", "test-service",
		"--operation-id", "some-id",
	)
	s.Error(res.Err)

	// Missing --operation-id
	res = s.Execute(
		"nexus", "operation", "start",
		"--endpoint", "test-ep",
		"--service", "test-service",
		"--operation", "test-op",
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestNexusOperationExecute_MissingOperationID() {
	res := s.Execute(
		"nexus", "operation", "execute",
		"--endpoint", "test-ep",
		"--service", "test-service",
		"--operation", "test-op",
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestNexusOperationStart_StaticSummary() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "summary-op-" + uuid.NewString()[:8]
	summary := "this is the operation summary"

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
		"--static-summary", summary,
	)
	s.NoError(res.Err)

	// Describe and verify the summary is reported back.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "describe",
			"--address", s.Address(),
			"--operation-id", opID,
		)
		return res.Err == nil && strings.Contains(res.Stdout.String(), summary)
	}, 30*time.Second, 500*time.Millisecond)
}

func (s *SharedServerSuite) TestNexusOperationStart_SearchAttribute() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "sa-op-" + uuid.NewString()[:8]
	uniqueKW := "nexus-sa-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
		"--search-attribute", fmt.Sprintf(`CustomKeywordField="%s"`, uniqueKW),
	)
	s.NoError(res.Err)

	// List with a query filter on the search attribute — confirms the SA was
	// attached and indexed.
	s.Eventually(func() bool {
		res = s.Execute(
			"nexus", "operation", "list",
			"--address", s.Address(),
			"--query", fmt.Sprintf(`CustomKeywordField = "%s"`, uniqueKW),
		)
		return res.Err == nil && strings.Contains(res.Stdout.String(), opID)
	}, 30*time.Second, 500*time.Millisecond)
}

func (s *SharedServerSuite) TestNexusOperationStart_Timeouts() {
	endpointName, w := s.setupNexusEndpointAndWorker(s.T())
	defer w.Stop()

	opID := "timeouts-op-" + uuid.NewString()[:8]

	res := s.Execute(
		"nexus", "operation", "start",
		"--address", s.Address(),
		"--endpoint", endpointName,
		"--service", "test-service",
		"--operation", "test-op",
		"--operation-id", opID,
		"--input", `"hello"`,
		"--schedule-to-close-timeout", "1m",
		"--schedule-to-start-timeout", "30s",
		"--start-to-close-timeout", "30s",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), opID)
}
