package temporalcli_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func TestNamespaceUpdate_ActiveClusterRejectsOtherUpdates(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError string
	}{
		{name: "cluster", args: []string{"--cluster", "cluster-a"}, wantError: "--active-cluster cannot be combined with --cluster"},
		{name: "data", args: []string{"--data", "key=value"}, wantError: "--active-cluster cannot be combined with --data"},
		{name: "description", args: []string{"--description", "description"}, wantError: "--active-cluster cannot be combined with --description"},
		{name: "email", args: []string{"--email", "owner@example.com"}, wantError: "--active-cluster cannot be combined with --email"},
		{name: "promote global", args: []string{"--promote-global"}, wantError: "both --promote-global and --active-cluster flags cannot be set together"},
		{name: "history archival state", args: []string{"--history-archival-state", "enabled"}, wantError: "--active-cluster cannot be combined with --history-archival-state"},
		{name: "history URI", args: []string{"--history-uri", "file:///tmp/history"}, wantError: "--active-cluster cannot be combined with --history-uri"},
		{name: "replication state", args: []string{"--replication-state", "normal"}, wantError: "--active-cluster cannot be combined with --replication-state"},
		{name: "retention", args: []string{"--retention", "24h"}, wantError: "--active-cluster cannot be combined with --retention"},
		{name: "visibility archival state", args: []string{"--visibility-archival-state", "enabled"}, wantError: "--active-cluster cannot be combined with --visibility-archival-state"},
		{name: "visibility URI", args: []string{"--visibility-uri", "file:///tmp/visibility"}, wantError: "--active-cluster cannot be combined with --visibility-uri"},
	}

	for _, output := range []string{"text", "json"} {
		t.Run(output, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					h := NewCommandHarness(t)
					args := []string{
						"operator", "namespace", "update",
						"--address", "127.0.0.1:1",
						"--namespace", "test-namespace",
						"--active-cluster", "cluster-b",
						"--output", output,
					}
					res := h.Execute(append(args, test.args...)...)

					require.ErrorContains(t, res.Err, test.wantError)
					require.Empty(t, res.Stdout.String())
				})
			}
		})
	}
}

func TestNamespaceUpdate_ActiveClusterReportsAllConflicts(t *testing.T) {
	h := NewCommandHarness(t)
	res := h.Execute(
		"operator", "namespace", "update",
		"--address", "127.0.0.1:1",
		"--namespace", "test-namespace",
		"--active-cluster", "cluster-b",
		"--cluster", "cluster-a",
		"--description", "description",
	)

	require.ErrorContains(t, res.Err, "--active-cluster cannot be combined with --cluster, --description")
	require.Empty(t, res.Stdout.String())
}

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

func (s *SharedServerSuite) TestNamespaceUpdate_ActiveClusterAlone() {
	nsName := "test-namespace-update-active-cluster"
	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		"--namespace", nsName,
		"--global",
		"--active-cluster", "active",
		"--cluster", "active",
	)
	require.NoError(s.T(), res.Err)

	res = s.Execute(
		"operator", "namespace", "update",
		"--address", s.Address(),
		"--namespace", nsName,
		"--active-cluster", "active",
	)
	require.NoError(s.T(), res.Err)
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

func (s *SharedServerSuite) TestOperatorNamespace_EnvConfigResolution() {
	// Create a test namespace to use in envconfig
	testNS := "envconfig-test-namespace"
	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		"-n", testNS,
	)
	s.NoError(res.Err)

	// Create temp config file with namespace
	f, err := os.CreateTemp("", "temporal-test-*.toml")
	s.NoError(err)
	defer os.Remove(f.Name())

	_, err = fmt.Fprintf(f, `
[profile.default]
address = "%s"
namespace = "%s"
`, s.Address(), testNS)
	s.NoError(err)
	f.Close()

	// Set environment to use config file
	s.CommandHarness.Options.EnvLookup = EnvLookupMap{
		"TEMPORAL_CONFIG_FILE": f.Name(),
	}

	// Test 1: Describe should use envconfig namespace (no -n flag)
	res = s.Execute(
		"operator", "namespace", "describe",
		"--output", "json",
	)
	s.NoError(res.Err)
	var descResp workflowservice.DescribeNamespaceResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &descResp, true))
	s.Equal(testNS, descResp.NamespaceInfo.Name, "Should use namespace from envconfig")

	// Test 2: Update should use envconfig namespace
	res = s.Execute(
		"operator", "namespace", "update",
		"--description", "Updated via envconfig",
		"--output", "json",
	)
	s.NoError(res.Err)

	// Verify update was applied to correct namespace
	res = s.Execute(
		"operator", "namespace", "describe",
		"--output", "json",
	)
	s.NoError(res.Err)
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &descResp, true))
	s.Equal("Updated via envconfig", descResp.NamespaceInfo.Description)
	s.Equal(testNS, descResp.NamespaceInfo.Name)

	// Test 3: CLI flag should override envconfig
	res = s.Execute(
		"operator", "namespace", "describe",
		"--output", "json",
		"-n", "default",
	)
	s.NoError(res.Err)
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &descResp, true))
	s.Equal("default", descResp.NamespaceInfo.Name, "Explicit -n flag should override envconfig")

	// Test 4: Delete should use envconfig namespace
	res = s.Execute(
		"operator", "namespace", "delete",
		"--yes",
	)
	s.NoError(res.Err)

	// Verify namespace was deleted
	res = s.Execute(
		"operator", "namespace", "describe",
		"--output", "json",
		"-n", testNS,
	)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "is not found")
}
