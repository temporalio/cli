package tests

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *e2eSuite) TestNamespaceUpdate_Verbose() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	nsName := "test-namespace-update-verbose"
	retentionBefore := time.Duration(24 * time.Hour)
	_, err := c.WorkflowService().RegisterNamespace(
		context.Background(),
		&workflowservice.RegisterNamespaceRequest{
			Namespace:                        nsName,
			Description:                      "description test",
			OwnerEmail:                       "email@test",
			WorkflowExecutionRetentionPeriod: durationpb.New(retentionBefore),
		},
	)
	s.NoError(err)

	// TODO: remove if namespace cache refresh is not an issue anymore
	time.Sleep(10 * time.Second)

	err = app.Run([]string{"", "operator", "namespace", "update", "--description", "description after", "--email", "email@after", "--retention", "48h", "--verbose", nsName})
	s.NoError(err)

	logs := writer.GetContent()
	s.Contains(logs, fmt.Sprintf("namespace\": \"%s", nsName))
	s.Contains(logs, "owner_email\": \"email@after")
	s.Contains(logs, "workflow_execution_retention_ttl\": 172800000000000")
	s.Contains(logs, fmt.Sprintf("Namespace %s update succeeded", nsName))

	nsAfter, err := c.WorkflowService().DescribeNamespace(context.Background(), &workflowservice.DescribeNamespaceRequest{Namespace: nsName})
	s.NoError(err)
	s.Equal("description after", nsAfter.GetNamespaceInfo().GetDescription())
	s.Equal("email@after", nsAfter.GetNamespaceInfo().GetOwnerEmail())
	s.Equal(float64(48*60*60), nsAfter.GetConfig().GetWorkflowExecutionRetentionTtl())
}

func (s *e2eSuite) TestNamespaceUpdate_NonVerbose() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	nsName := "test-namespace-update-non-verbose"
	retentionBefore := time.Duration(24 * time.Hour)
	_, err := c.WorkflowService().RegisterNamespace(
		context.Background(),
		&workflowservice.RegisterNamespaceRequest{
			Namespace:                        nsName,
			Description:                      "description test",
			OwnerEmail:                       "email@test",
			WorkflowExecutionRetentionPeriod: durationpb.New(retentionBefore),
		},
	)
	s.NoError(err)

	// TODO: remove if namespace cache refresh is not an issue anymore
	time.Sleep(10 * time.Second)

	err = app.Run([]string{"", "operator", "namespace", "update", "--description", "description after", "--email", "email@after", "--retention", "48h", nsName})
	s.NoError(err)

	logs := writer.GetContent()
	s.NotContains(logs, "email@after")
	s.NotContains(logs, "description after")
	s.NotContains(logs, "172800000000000")
	s.Contains(logs, fmt.Sprintf("Namespace %s update succeeded", nsName))

	nsAfter, err := c.WorkflowService().DescribeNamespace(context.Background(), &workflowservice.DescribeNamespaceRequest{Namespace: nsName})
	s.NoError(err)
	s.Equal("description after", nsAfter.GetNamespaceInfo().GetDescription())
	s.Equal("email@after", nsAfter.GetNamespaceInfo().GetOwnerEmail())
	s.Equal(float64(48*60*60), nsAfter.GetConfig().GetWorkflowExecutionRetentionTtl())
}

func (s *e2eSuite) TestNamespaceUpdate_Data() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	nsName := "default"
	err := app.Run([]string{"", "operator", "namespace", "update", "--data", "k1=v1", "--data", "k2=v2", nsName})
	s.NoError(err)

	s.Contains(writer.GetContent(), fmt.Sprintf("Namespace %s update succeeded", nsName))

	nsAfter, err := c.WorkflowService().DescribeNamespace(context.Background(), &workflowservice.DescribeNamespaceRequest{Namespace: nsName})
	s.NoError(err)

	s.Equal("v1", nsAfter.GetNamespaceInfo().GetData()["k1"])
	s.Equal("v2", nsAfter.GetNamespaceInfo().GetData()["k2"])
}

func (s *e2eSuite) TestNamespaceUpdate_NamespaceDontExist() {
	s.T().Parallel()

	testserver, app, _ := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	nsName := "missing-namespace"
	err := app.Run([]string{"", "operator", "namespace", "update", "--email", "email@after", nsName})
	s.Error(err)
	s.Contains(err.Error(), "Namespace missing-namespace is not found")
}

func (s *e2eSuite) TestNamespaceUpdate_Cluster() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	nsName := "default"
	err := app.Run([]string{"", "operator", "namespace", "update", "--cluster", "active", nsName})
	s.NoError(err)

	s.Contains(writer.GetContent(), fmt.Sprintf("Namespace %s update succeeded", nsName))

	nsAfter, err := c.WorkflowService().DescribeNamespace(context.Background(), &workflowservice.DescribeNamespaceRequest{Namespace: nsName})
	s.NoError(err)

	s.Equal("active", nsAfter.GetReplicationConfig().GetActiveClusterName())
}
