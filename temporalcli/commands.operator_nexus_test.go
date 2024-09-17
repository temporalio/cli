package temporalcli_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
)

func (s *SharedServerSuite) TestCreateNexusEndpoint_Target() {
	s.T().Run("BothWorkerAndExternal_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq",
			"--target-url", "http://fake-url-for-test",
		)
		require.ErrorContains(t, res.Err, "provided both --target-namespace and --target-url")
	})

	s.T().Run("NoTarget_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create", "--name", validEndpointName(t))
		require.ErrorContains(t, res.Err, "either --target-namespace and --target-task queue or --target-url are required")
	})

	s.T().Run("NoTaskQueue_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--name", validEndpointName(t),
			"--target-namespace", "default",
		)
		require.ErrorContains(t, res.Err, "both --target-namespace and --target-task-queue are required")
	})

	s.T().Run("NoNamespace_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--name", validEndpointName(t),
			"--target-task-queue", "tq",
		)
		require.ErrorContains(t, res.Err, "either --target-namespace and --target-task queue or --target-url are required")
	})

	s.T().Run("WorkerTarget_Accepted", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq")
		require.NoError(t, res.Err)
		require.Contains(t, res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully created.", validEndpointName(t)))

		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		target := endpoint.Spec.Target.GetWorker()
		require.Equal(t, "default", target.Namespace)
		require.Equal(t, "tq", target.TaskQueue)
	})

	s.T().Run("ExternalTarget_Accepted", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-url", "http://just-a-test")
		require.NoError(t, res.Err)
		require.Contains(t, res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully created.", validEndpointName(t)))

		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		target := endpoint.Spec.Target.GetExternal()
		require.Equal(t, "http://just-a-test", target.Url)
	})
}

func (s *SharedServerSuite) TestCreateNexusEndpoint_Description() {
	s.T().Run("BothDescriptionAndDescriptionFile_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--name", validEndpointName(t),
			"--description", "foo",
			"--description-file", "bar")
		require.ErrorContains(t, res.Err, "provided both --description and --description-file")
	})

	s.T().Run("DescriptionFile_Accepted", func(t *testing.T) {
		p := filepath.Join(s.T().TempDir(), "description.md")

		require.NoError(t, os.WriteFile(p, []byte("markdown"), 0o755))
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq",
			"--description-file", p)
		require.NoError(t, res.Err)
		require.Contains(t, res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully created.", validEndpointName(t)))

		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		require.Equal(t, []byte("json/plain"), endpoint.Spec.Description.Metadata["encoding"])
		require.Equal(t, []byte(`"markdown"`), endpoint.Spec.Description.Data)
	})

	s.T().Run("Description_Accepted", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq",
			"--description", "markdown")
		require.NoError(t, res.Err)
		require.Contains(t, res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully created.", validEndpointName(t)))

		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		require.Equal(t, []byte("json/plain"), endpoint.Spec.Description.Metadata["encoding"])
		require.Equal(t, []byte(`"markdown"`), endpoint.Spec.Description.Data)
	})

	s.T().Run("Description_NotRequired", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq")
		require.NoError(t, res.Err)
		require.Contains(t, res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully created.", validEndpointName(t)))

		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		require.Nil(t, endpoint.Spec.Description)
	})
}

func (s *SharedServerSuite) TestUpdateNexusEndpoint() {
	s.T().Run("BothDescriptionAndDescriptionFile_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "update",
			"--name", validEndpointName(t),
			"--description", "foo",
			"--description-file", "bar")
		require.ErrorContains(t, res.Err, "provided both --description and --description-file")
	})

	s.T().Run("BothDescriptionAndUnsetDescription_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "update",
			"--name", validEndpointName(t),
			"--description", "foo",
			"--unset-description")
		require.ErrorContains(t, res.Err, "--unset-description should not be set if --description or --description-file is set")
	})

	s.T().Run("BothWorkerAndExternal_FailsValidation", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "update",
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-url", "http://fake-url-for-test",
		)
		require.ErrorContains(t, res.Err, "provided both --target-namespace and --target-url")

		res = s.Execute("operator", "nexus", "endpoint", "update",
			"--name", validEndpointName(t),
			"--target-task-queue", "tq",
			"--target-url", "http://fake-url-for-test",
		)
		require.ErrorContains(t, res.Err, "provided both --target-task-queue and --target-url")
	})

	s.T().Run("EndpointNotFound_ReturnsClearError", func(t *testing.T) {
		res := s.Execute("operator", "nexus", "endpoint", "update",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--description", "markdown",
			"--target-namespace", "default",
			"--target-task-queue", "tq")
		require.ErrorContains(t, res.Err, "endpoint not found")
	})

	s.T().Run("ExternalToWorker", func(t *testing.T) {
		// First create an endpoint.
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-url", "http://fake-url-for-test")
		require.NoError(t, res.Err)

		// Verify both namespace and task queue are required.
		res = s.Execute("operator", "nexus", "endpoint", "update",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-task-queue", "new-tq")
		require.ErrorContains(t, res.Err, "both --target-namespace and --target-task-queue are required when changing target type from external to worker")

		// Verify endpoint is updated.
		res = s.Execute("operator", "nexus", "endpoint", "update",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq")
		require.NoError(t, res.Err)
		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		require.Equal(t, "default", endpoint.Spec.Target.GetWorker().Namespace)
		require.Equal(t, "tq", endpoint.Spec.Target.GetWorker().TaskQueue)
	})

	s.T().Run("PartialWorker", func(t *testing.T) {
		// First create an endpoint.
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq")
		require.NoError(t, res.Err)

		// Verify endpoint is updated.
		res = s.Execute("operator", "nexus", "endpoint", "update",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-task-queue", "new-tq")
		require.NoError(t, res.Err)
		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		require.Equal(t, "default", endpoint.Spec.Target.GetWorker().Namespace)
		require.Equal(t, "new-tq", endpoint.Spec.Target.GetWorker().TaskQueue)
	})

	s.T().Run("External", func(t *testing.T) {
		// First create an endpoint.
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-namespace", "default",
			"--target-task-queue", "tq")
		require.NoError(t, res.Err)

		// Verify endpoint is updated.
		res = s.Execute("operator", "nexus", "endpoint", "update",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-url", "http://fake-url-for-test")
		require.NoError(t, res.Err)
		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		require.Equal(t, "http://fake-url-for-test", endpoint.Spec.Target.GetExternal().Url)
	})

	s.T().Run("UpdateDescription", func(t *testing.T) {
		// First create an endpoint.
		res := s.Execute("operator", "nexus", "endpoint", "create",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--target-url", "http://fake-url-for-test",
			"--description", "v1")
		require.NoError(t, res.Err)

		res = s.Execute("operator", "nexus", "endpoint", "update",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--description", "v2")
		require.NoError(t, res.Err)
		require.Contains(t, res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully updated.", validEndpointName(t)))

		res = s.Execute("operator", "nexus", "endpoint", "update",
			"--address", s.Address(),
			"--name", validEndpointName(t),
			"--unset-description")
		require.NoError(t, res.Err)
		require.Contains(t, res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully updated.", validEndpointName(t)))

		endpoint := s.getNexusEndpointByName(validEndpointName(t))
		require.Nil(t, endpoint.Spec.Description)
	})
}

func (s *SharedServerSuite) TestDeleteNexusEndpoint() {
	// First create an endpoint.
	res := s.Execute("operator", "nexus", "endpoint", "create",
		"--address", s.Address(),
		"--name", validEndpointName(s.T()),
		"--description", "markdown",
		"--target-namespace", "default",
		"--target-task-queue", "tq")
	s.NoError(res.Err)

	res = s.Execute("operator", "nexus", "endpoint", "delete",
		"--address", s.Address(),
		"--name", validEndpointName(s.T()),
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), fmt.Sprintf("Endpoint %s successfully deleted (ID", validEndpointName(s.T())))
}

func (s *SharedServerSuite) TestGetNexusEndpoint() {
	// First create an endpoint.
	res := s.Execute("operator", "nexus", "endpoint", "create",
		"--address", s.Address(),
		"--name", validEndpointName(s.T()),
		"--description", "markdown",
		"--target-namespace", "default",
		"--target-task-queue", "tq")
	s.NoError(res.Err)

	s.T().Run("Text", func(t *testing.T) {
		res = s.Execute("operator", "nexus", "endpoint", "get",
			"--address", s.Address(),
			"--name", validEndpointName(s.T()))
		require.NoError(t, res.Err)

		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Name", validEndpointName(s.T())))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Target.Worker.Namespace", "default"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Target.Worker.TaskQueue", "tq"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Description", "markdown"))
	})

	s.T().Run("JSON", func(t *testing.T) {
		res = s.Execute("operator", "nexus", "endpoint", "get",
			"--address", s.Address(),
			"--name", validEndpointName(s.T()),
			"--output", "json",
		)
		require.NoError(t, res.Err)

		var endpoint nexus.Endpoint
		require.NoError(t, temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &endpoint, true))
		require.Equal(t, validEndpointName(s.T()), endpoint.Spec.Name)
		require.Equal(t, "default", endpoint.Spec.Target.GetWorker().Namespace)
		require.Equal(t, "tq", endpoint.Spec.Target.GetWorker().TaskQueue)
		require.Equal(t, `"markdown"`, string(endpoint.Spec.Description.Data))
	})
}

func (s *SharedServerSuite) TestListNexusEndpoints() {
	// Create a couple of endpoints.
	res := s.Execute("operator", "nexus", "endpoint", "create",
		"--address", s.Address(),
		"--name", validEndpointName(s.T())+"-1",
		"--description", "markdown-1",
		"--target-namespace", "default",
		"--target-task-queue", "tq-1")
	s.NoError(res.Err)

	res = s.Execute("operator", "nexus", "endpoint", "create",
		"--address", s.Address(),
		"--name", validEndpointName(s.T())+"-2",
		"--description", "markdown-2",
		"--target-namespace", "default",
		"--target-task-queue", "tq-2")
	s.NoError(res.Err)

	s.T().Run("Text", func(t *testing.T) {
		res = s.Execute("operator", "nexus", "endpoint", "list", "--address", s.Address())
		require.NoError(t, res.Err)

		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Name", validEndpointName(s.T())+"-1"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Target.Worker.Namespace", "default"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Target.Worker.TaskQueue", "tq-1"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Description", "markdown-1"))

		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Name", validEndpointName(s.T())+"-2"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Target.Worker.Namespace", "default"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Target.Worker.TaskQueue", "tq-2"))
		require.NoError(t, AssertContainsOnSameLine(res.Stdout.String(), "Description", "markdown-2"))
	})

	s.T().Run("JSON", func(t *testing.T) {
		res = s.Execute("operator", "nexus", "endpoint", "list",
			"--address", s.Address(),
			"--output", "json")
		require.NoError(t, res.Err)

		output := fmt.Sprintf("{\"endpoints\": %s}", res.Stdout.String())
		var listResp operatorservice.ListNexusEndpointsResponse
		s.NoError(temporalcli.UnmarshalProtoJSONWithOptions([]byte(output), &listResp, true))
		endpoints := listResp.Endpoints
		// There may be endpoints created by other tests, ignore those.
		require.GreaterOrEqual(t, len(endpoints), 2)

		ep1Idx := slices.IndexFunc(endpoints, func(e *nexus.Endpoint) bool {
			return e.Spec.Name == validEndpointName(s.T())+"-1"
		})
		ep2Idx := slices.IndexFunc(endpoints, func(e *nexus.Endpoint) bool {
			return e.Spec.Name == validEndpointName(s.T())+"-2"
		})
		require.Equal(t, "default", endpoints[ep1Idx].Spec.Target.GetWorker().Namespace)
		require.Equal(t, "tq-1", endpoints[ep1Idx].Spec.Target.GetWorker().TaskQueue)
		require.Equal(t, `"markdown-1"`, string(endpoints[ep1Idx].Spec.Description.Data))

		require.Equal(t, "default", endpoints[ep2Idx].Spec.Target.GetWorker().Namespace)
		require.Equal(t, "tq-2", endpoints[ep2Idx].Spec.Target.GetWorker().TaskQueue)
		require.Equal(t, `"markdown-2"`, string(endpoints[ep2Idx].Spec.Description.Data))
	})
}

func (s *SharedServerSuite) getNexusEndpointByName(name string) *nexus.Endpoint {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	resp, err := s.Client.OperatorService().ListNexusEndpoints(ctx, &operatorservice.ListNexusEndpointsRequest{})
	s.NoError(err)
	idx := slices.IndexFunc(resp.Endpoints, func(ep *nexus.Endpoint) bool {
		return ep.Spec.Name == name
	})
	s.Greater(idx, -1)
	return resp.Endpoints[idx]
}

func validEndpointName(t *testing.T) string {
	re := regexp.MustCompile("[/_]")
	return re.ReplaceAllString(t.Name(), "-")
}
