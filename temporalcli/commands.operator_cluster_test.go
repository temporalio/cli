package temporalcli_test

import (
	"encoding/json"

	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *SharedServerSuite) TestOperator_Cluster_System() {
	// Text
	res := s.Execute(
		"operator", "cluster", "system",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ServerVersion", "SupportsSchedules", "UpsertMemo", "EagerWorkflowStart")
	s.ContainsOnSameLine(out, "true", "true", "true")

	// JSON
	res = s.Execute(
		"operator", "cluster", "system",
		"-o", "json",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	var jsonOut workflowservice.GetSystemInfoResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonOut, true))
	s.NotEmpty(jsonOut.ServerVersion)
	s.True(jsonOut.Capabilities.SupportsSchedules)
	s.True(jsonOut.Capabilities.UpsertMemo)
	s.True(jsonOut.Capabilities.EagerWorkflowStart)
}

func (s *SharedServerSuite) TestOperator_Cluster_Describe() {
	// Text
	res := s.Execute(
		"operator", "cluster", "describe",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ClusterName", "PersistenceStore", "VisibilityStore")
	s.ContainsOnSameLine(out, "active", "sqlite", "sqlite")

	// JSON
	res = s.Execute(
		"operator", "cluster", "describe",
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res.Err)
	var jsonOut workflowservice.GetClusterInfoResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonOut, true))
	s.Equal("active", jsonOut.ClusterName)
	s.Equal("sqlite", jsonOut.PersistenceStore)
	s.Equal("sqlite", jsonOut.VisibilityStore)
}

func (s *SharedServerSuite) TestOperator_Cluster_Health() {
	// Text
	res := s.Execute(
		"operator", "cluster", "health",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.Equal(out, "SERVING\n")

	// JSON
	res = s.Execute(
		"operator", "cluster", "health",
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res.Err)
	var jsonOut map[string]string
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(jsonOut["status"], "SERVING")
}
