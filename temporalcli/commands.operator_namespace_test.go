package temporalcli_test

import (
	"encoding/json"
	"time"

	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *SharedServerSuite) TestOperator_Namespace_Create_WithoutName() {
	res := s.Execute(
		"operator", "namespace", "create",
	)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "required flag(s) \"name\" not set")
}

func (s *SharedServerSuite) TestOperator_Namespace_Create_List_And_Describe() {
	namespaceName := "test_namespace"
	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		"--name", namespaceName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "list",
		"--address", s.Address(),
		"--output", "json",
	)
	s.NoError(res.Err)
	var listResp []workflowservice.DescribeNamespaceResponse
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &listResp))
	var uuid string
	for _, ns := range listResp {
		if ns.NamespaceInfo.Name == namespaceName {
			uuid = ns.NamespaceInfo.Id
		}
	}
	s.NotEmpty(uuid)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--namespace-id", uuid,
		"--output", "json",
	)
	s.NoError(res.Err)
	var describeResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp, true))
	s.Equal(namespaceName, describeResp.NamespaceInfo.Name)

	// Validate default values
	s.Equal("active", describeResp.ReplicationConfig.ActiveClusterName)
	s.Equal("active", describeResp.ReplicationConfig.Clusters[0].ClusterName)
	s.Len(describeResp.NamespaceInfo.Data, 0)
	s.Len(describeResp.NamespaceInfo.Description, 0)
	s.Len(describeResp.NamespaceInfo.OwnerEmail, 0)
	s.Equal(false, describeResp.IsGlobalNamespace)
	s.Equal(enums.ARCHIVAL_STATE_DISABLED, describeResp.Config.HistoryArchivalState)
	s.Len(describeResp.Config.HistoryArchivalUri, 0)
	s.Equal(describeResp.Config.WorkflowExecutionRetentionTtl.AsDuration(), 72*time.Hour)
	s.Equal(enums.ARCHIVAL_STATE_DISABLED, describeResp.Config.VisibilityArchivalState)
	s.Len(describeResp.Config.VisibilityArchivalUri, 0)
}
