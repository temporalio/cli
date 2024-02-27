package temporalcli_test

import (
	"encoding/json"
	"time"

	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *SharedServerSuite) TestOperator_Namespace_Create_List_And_Describe() {
	nsName := "test_namespace"
	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		nsName,
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
		if ns.NamespaceInfo.Name == nsName {
			uuid = ns.NamespaceInfo.Id
		}
	}
	s.NotEmpty(uuid)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		nsName,
	)
	s.NoError(res.Err)
	var describeResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp, true))
	s.Equal(nsName, describeResp.NamespaceInfo.Name)

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

func (s *SharedServerSuite) TestNamespaceUpdate_Verbose() {
	nsName := "test-namespace-update-verbose"

	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		"--description", "description before",
		"--email", "email@before",
		"--retention", "1",
		nsName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "update",
		"--address", s.Address(),
		"--description", "description after",
		"--email", "email@after",
		"--retention", "2",
		"--output", "json",
		"--verbose",
		nsName,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "workflowExecutionRetentionTtl", "172800s")
	s.ContainsOnSameLine(res.Stdout.String(), "description", "description after")
	s.ContainsOnSameLine(res.Stdout.String(), "ownerEmail", "email@after")

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		nsName,
	)
	s.NoError(res.Err)
	var describeResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp, true))

	s.Equal("description after", describeResp.NamespaceInfo.Description)
	s.Equal("email@after", describeResp.NamespaceInfo.OwnerEmail)
	s.Equal(48*time.Hour, describeResp.Config.WorkflowExecutionRetentionTtl.AsDuration())
}

func (s *SharedServerSuite) TestNamespaceUpdate_Data() {
	nsName := "default"
	res := s.Execute(
		"operator", "namespace", "update",
		"--address", s.Address(),
		"--data", "k1=v1",
		"--data", "k2=v2",
		nsName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		nsName,
	)
	s.NoError(res.Err)
	var describeResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp, true))

	s.Equal("v1", describeResp.NamespaceInfo.Data["k1"])
	s.Equal("v2", describeResp.NamespaceInfo.Data["k2"])
}

func (s *SharedServerSuite) TestNamespaceUpdate_NamespaceDontExist() {
	nsName := "missing-namespace"
	res := s.Execute(
		"operator", "namespace", "update",
		"--email", "email@after",
		"--address", s.Address(),
		nsName,
	)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "Namespace missing-namespace is not found")
}

func (s *SharedServerSuite) TestDeleteNamespace() {
	nsName := "test-namespace"
	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		nsName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		nsName,
	)
	s.NoError(res.Err)
	var describeResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp, true))
	s.Equal(nsName, describeResp.NamespaceInfo.Name)

	res = s.Execute(
		"operator", "namespace", "delete",
		"--address", s.Address(),
		"--yes",
		nsName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		nsName,
	)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "Namespace test-namespace is not found")
}

func (s *SharedServerSuite) TestDescribeWithID() {
	res := s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		"default",
	)
	s.NoError(res.Err)

	var describeResp1 workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp1, true))
	nsID := describeResp1.NamespaceInfo.Id

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		"--namespace-id", nsID,
	)
	s.NoError(res.Err)
	var describeResp2 workflowservice.DescribeNamespaceResponse

	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp2, true))
	s.Equal(describeResp1.NamespaceInfo, describeResp2.NamespaceInfo)
}
