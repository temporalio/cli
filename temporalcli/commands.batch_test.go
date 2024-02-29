package temporalcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *SharedServerSuite) TestBatchJob_Describe() {
	s.t.Run("non-existing job id", func(t *testing.T) {
		t.Run("as text", func(t *testing.T) {
			res := s.Execute(
				"batch", "describe",
				"--address", s.Address(),
				"--job-id", "not-found")
			s.EqualError(res.Err, "could not find Batch Job 'not-found'")
		})

		t.Run("as json", func(t *testing.T) {
			res := s.Execute(
				"batch", "describe",
				"--address", s.Address(),
				"--job-id", "not-found",
				"-o", "json")
			s.EqualError(res.Err, "could not find Batch Job 'not-found'")
		})
	})

	s.t.Run("existing job id", func(t *testing.T) {
		// kickstart batch job (using Client directly to provide a `JobId`)
		jobId := "TestBatchJob_Describe"
		s.startBatchJob(jobId, s.Namespace())

		t.Run("as text", func(t *testing.T) {
			res := s.Execute(
				"batch", "describe",
				"--address", s.Address(),
				"--job-id", jobId)
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())

			out := res.Stdout.String()
			s.ContainsOnSameLine(out, "State", "Running")
			s.ContainsOnSameLine(out, "Type", "Terminate")
			s.ContainsOnSameLine(out, "CompletedCount", "0/0")
			s.ContainsOnSameLine(out, "FailureCount", "0/0")
		})

		t.Run("as json", func(t *testing.T) {
			res := s.Execute(
				"batch", "describe",
				"--address", s.Address(),
				"--job-id", jobId,
				"-o", "json")
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())

			spew.Dump(res.Stdout.String())

			var jsonOut map[string]any
			s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
			s.Equal(jobId, jsonOut["jobId"])
			s.Equal("BATCH_OPERATION_TYPE_TERMINATE", jsonOut["operationType"])
			s.Equal("REASON", jsonOut["reason"])
		})
	})
}

func (s *SharedServerSuite) TestBatchJob_List() {
	// NOTE: this test is the only test to use the "batch-empty" namespace;
	// ie it is guaranteed to be empty at the start

	s.t.Run("no batch jobs", func(t *testing.T) {
		t.Run("as text", func(t *testing.T) {
			res := s.Execute(
				"batch", "list",
				"--namespace", "batch-empty",
				"--address", s.Address())
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())
			s.Empty(res.Stdout.String())
		})

		t.Run("as json", func(t *testing.T) {
			res := s.Execute(
				"batch", "list",
				"--address", s.Address(),
				"--namespace", "batch-empty",
				"-o", "json",
			)
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())
			s.Equal("[\n]\n", res.Stdout.String())
		})
	})

	s.t.Run("a few batch jobs", func(t *testing.T) {
		// start a few batch jobs
		for i := 0; i < 3; i++ {
			s.startBatchJob(fmt.Sprintf("TestBatchJob_List_%d", i), "batch-empty")
		}

		t.Run("as text", func(t *testing.T) {
			var res *CommandResult
			s.Eventually(func() bool {
				res = s.Execute(
					"batch", "list",
					"--address", s.Address(),
					"--namespace", "batch-empty")
				s.NoError(res.Err)
				s.Empty(res.Stderr.String())
				return strings.Count(res.Stdout.String(), "Completed") == 3
			}, 5*time.Second, 100*time.Millisecond)

			out := res.Stdout.String()
			s.Equal(4, strings.Count(out, "\n"), "expect 3 data rows + 1 header row")
			s.ContainsOnSameLine(out, "JobId", "State", "StartTime", "CloseTime") // header
			s.ContainsOnSameLine(out, "TestBatchJob_List_2", "Completed")
			s.ContainsOnSameLine(out, "TestBatchJob_List_1", "Completed")
			s.ContainsOnSameLine(out, "TestBatchJob_List_0", "Completed")
		})

		t.Run("as text with limit", func(t *testing.T) {
			res := s.Execute(
				"batch", "list",
				"--address", s.Address(),
				"--namespace", "batch-empty",
				"--limit", "1")
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())

			out := res.Stdout.String()
			s.Equal(2, strings.Count(out, "\n"), "expect 1 data row + 1 header row")
			s.ContainsOnSameLine(out, "JobId", "State", "StartTime", "CloseTime") // header
		})

		t.Run("as json", func(t *testing.T) {
			res := s.Execute(
				"batch", "list",
				"--address", s.Address(),
				"--namespace", "batch-empty",
				"-o", "json")
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())

			out := res.Stdout.String()
			s.ContainsOnSameLine(out, "\"jobId\": \"TestBatchJob_List_2\"")
			s.ContainsOnSameLine(out, "\"jobId\": \"TestBatchJob_List_1\"")
			s.ContainsOnSameLine(out, "\"jobId\": \"TestBatchJob_List_0\"")
		})
	})
}

func (s *SharedServerSuite) TestBatchJob_Terminate() {
	s.t.Run("non-existing job id", func(t *testing.T) {
		t.Run("as text", func(t *testing.T) {
			res := s.Execute(
				"batch", "terminate",
				"--address", s.Address(),
				"--job-id", "not-found",
				"--reason", "testing")
			s.EqualError(res.Err, "could not find Batch Job 'not-found'")
		})

		t.Run("as json", func(t *testing.T) {
			res := s.Execute(
				"batch", "terminate",
				"--address", s.Address(),
				"--job-id", "not-found",
				"--reason", "testing",
				"-o", "json")
			s.EqualError(res.Err, "could not find Batch Job 'not-found'")
		})
	})

	s.t.Run("existing job id", func(t *testing.T) {
		t.Run("as text", func(t *testing.T) {
			jobId := "TestBatchJob_Terminate_Text"
			s.startBatchJob(jobId, s.Namespace())

			res := s.Execute(
				"batch", "terminate",
				"--address", s.Address(),
				"--job-id", jobId,
				"--reason", "testing")
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())
			s.Equal("Terminated Batch Job '"+jobId+"'\n", res.Stdout.String())
		})

		t.Run("as json", func(t *testing.T) {
			jobId := "TestBatchJob_Terminate_JSON"
			s.startBatchJob(jobId, s.Namespace())

			res := s.Execute(
				"batch", "terminate",
				"--address", s.Address(),
				"--job-id", jobId,
				"--reason", "testing",
				"-o", "json")
			s.NoError(res.Err)
			s.Empty(res.Stderr.String())
			s.Empty(res.Stdout.String())
		})
	})
}

// kickstart batch job (using Client directly to provide a `JobId`)
func (s *SharedServerSuite) startBatchJob(jobId, namespace string) {
	_, err := s.Client.WorkflowService().StartBatchOperation(
		context.Background(),
		&workflowservice.StartBatchOperationRequest{
			JobId:           jobId,
			Namespace:       namespace,
			VisibilityQuery: "WorkflowType=\"BATCH-TEST\"",
			Reason:          "REASON",
			Operation: &workflowservice.StartBatchOperationRequest_TerminationOperation{
				TerminationOperation: &batch.BatchOperationTermination{},
			},
		},
	)
	s.NoError(err)
}
