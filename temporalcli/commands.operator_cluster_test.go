package temporalcli_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/temporalio/cli/temporalcli"
	"github.com/temporalio/cli/temporalcli/devserver"
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

func (s *SharedServerSuite) TestOperator_Cluster_Operations() {
	// Create some clusters
	standbyCluster1 := StartDevServer(s.Suite.T(), DevServerOptions{
		StartOptions: devserver.StartOptions{
			MasterClusterName:      "standby1",
			CurrentClusterName:     "standby1",
			InitialFailoverVersion: 2,
			EnableGlobalNamespace:  true,
		},
	})
	defer standbyCluster1.Stop()

	standbyCluster2 := StartDevServer(s.Suite.T(), DevServerOptions{
		StartOptions: devserver.StartOptions{
			MasterClusterName:      "standby2",
			CurrentClusterName:     "standby2",
			InitialFailoverVersion: 3,
			EnableGlobalNamespace:  true,
		},
	})
	defer standbyCluster2.Stop()

	// Upsert the clusters

	// Text
	res := s.Execute(
		"operator", "cluster", "upsert",
		"--frontend-address", standbyCluster1.Address(),
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.Equal("Upserted cluster "+standbyCluster1.Address()+"\n", out)

	// JSON
	res = s.Execute(
		"operator", "cluster", "upsert",
		"--frontend-address", standbyCluster2.Address(),
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res.Err)
	var jsonOut map[string]string
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(jsonOut["frontendAddress"], standbyCluster2.Address())

	// List the clusters

	// Text
	res = s.Execute(
		"operator", "cluster", "list",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "Name", "ClusterId", "Address", "HistoryShardCount", "InitialFailoverVersion", "IsConnectionEnabled")
	s.ContainsOnSameLine(out, "active", s.DevServer.Options.ClusterID, s.DevServer.Address(), "1", "1", "true")
	s.ContainsOnSameLine(out, standbyCluster1.Options.CurrentClusterName, standbyCluster1.Options.ClusterID, standbyCluster1.Address(), "1", strconv.Itoa(standbyCluster1.Options.InitialFailoverVersion), "false")
	s.ContainsOnSameLine(out, standbyCluster2.Options.CurrentClusterName, standbyCluster2.Options.ClusterID, standbyCluster2.Address(), "1", strconv.Itoa(standbyCluster2.Options.InitialFailoverVersion), "false")

	// JSON
	res = s.Execute(
		"operator", "cluster", "list",
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	// If we improve the output of list commands https://github.com/temporalio/cli/issues/448
	// we can do more detailed checks here.
	s.ContainsOnSameLine(out, fmt.Sprintf("\"clusterId\": \"%s\"", s.DevServer.Options.ClusterID))
	s.ContainsOnSameLine(out, fmt.Sprintf("\"clusterId\": \"%s\"", standbyCluster1.Options.ClusterID))
	s.ContainsOnSameLine(out, fmt.Sprintf("\"clusterId\": \"%s\"", standbyCluster2.Options.ClusterID))

	// Need to wait for the cluster cache to be updated
	// clusterMetadataRefreshInterval is 1 millisecond, waiting
	// 5 milliseconds should allow for a sufficient buffer
	time.Sleep(5 * time.Millisecond)

	// Remove the clusters

	// Test
	res = s.Execute(
		"operator", "cluster", "remove",
		"--name", standbyCluster1.Options.CurrentClusterName,
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	s.Equal(fmt.Sprintf("Removed cluster %s\n", standbyCluster1.Options.CurrentClusterName), res.Stdout.String())

	// JSON
	res = s.Execute(
		"operator", "cluster", "remove",
		"--name", standbyCluster2.Options.CurrentClusterName,
		"--address", s.Address(),
		"-o", "json",
	)
	jsonOut = make(map[string]string)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(jsonOut["clusterName"], standbyCluster2.Options.CurrentClusterName)
}
