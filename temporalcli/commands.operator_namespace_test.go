package temporalcli_test

import (
	"fmt"
	"time"

	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *SharedServerSuite) TestOperator_NamespaceCreateListAndDescribe() {
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
	output := fmt.Sprintf("{\"namespaces\": %s}", res.Stdout.String())
	var listResp workflowservice.ListNamespacesResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions([]byte(output), &listResp, true))
	var uuid string
	for _, ns := range listResp.Namespaces {
		if ns.NamespaceInfo.Name == nsName {
			uuid = ns.NamespaceInfo.Id
		}
	}
	s.NotEmpty(uuid)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		"-n", nsName,
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

func (s *SharedServerSuite) TestNamespaceUpdate() {
	nsName := "test-namespace-update-verbose"

	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		"--description", "description before",
		"--email", "email@before",
		"--retention", "24h",
		"--data", "k1=v0",
		"--data", "k3=v3",
		"-n", nsName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "update",
		"--address", s.Address(),
		"--description", "description after",
		"--email", "email@after",
		"--retention", "2d",
		"--data", "k1=v1",
		"--data", "k2=v2",
		"--output", "json",
		"-n", nsName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		"-n", nsName,
	)
	s.NoError(res.Err)
	var describeResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp, true))

	s.Equal("description after", describeResp.NamespaceInfo.Description)
	s.Equal("email@after", describeResp.NamespaceInfo.OwnerEmail)
	s.Equal(48*time.Hour, describeResp.Config.WorkflowExecutionRetentionTtl.AsDuration())
	s.Equal("v1", describeResp.NamespaceInfo.Data["k1"])
	s.Equal("v2", describeResp.NamespaceInfo.Data["k2"])
	s.Equal("v3", describeResp.NamespaceInfo.Data["k3"])
}

func (s *SharedServerSuite) TestNamespaceUpdate_NamespaceDontExist() {
	nsName := "missing-namespace"
	res := s.Execute(
		"operator", "namespace", "update",
		"--email", "email@after",
		"--address", s.Address(),
		"-n", nsName,
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
		"-n", nsName,
	)
	s.NoError(res.Err)
	var describeResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &describeResp, true))
	s.Equal(nsName, describeResp.NamespaceInfo.Name)

	res = s.Execute(
		"operator", "namespace", "delete",
		"--address", s.Address(),
		"--yes",
		"-n", nsName,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		"-n", nsName,
	)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "Namespace test-namespace is not found")
}

func (s *SharedServerSuite) TestDescribeWithID() {
	res := s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		"-n", "default",
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

func (s *SharedServerSuite) TestDescribeBothNameAndID() {
	res := s.Execute(
		"operator", "namespace", "describe",
		"--address", s.Address(),
		"--output", "json",
		"-n", "asdf",
		"--namespace-id=ad7ef0ce-7139-4333-b8ee-60a79c8fda1d",
	)
	s.Error(res.Err)
	s.ContainsOnSameLine(res.Err.Error(), "provide one of", "but not both")
}

func (s *SharedServerSuite) TestUpdateOldAndNewNSArgs() {
	res := s.Execute(
		"operator", "namespace", "update",
		"--address", s.Address(),
		"--output", "json",
		"--email", "foo@bar",
		"-n", "asdf",
		"asdf",
	)
	s.Error(res.Err)
	s.ContainsOnSameLine(res.Err.Error(), "namespace was provided as both an argument", "and a flag")
}
