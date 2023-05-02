package tests

import (
	"context"
	"time"

	"go.temporal.io/api/workflowservice/v1"
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
			OwnerEmail:                       "email test",
			WorkflowExecutionRetentionPeriod: &retentionBefore,
		},
	)
	s.NoError(err)

	// TODO: remove if namespace cache refresh is not an issue anymore
	time.Sleep(10 * time.Second)

	err = app.Run([]string{"", "operator", "namespace", "update", "--description", "description after", "--email", "email after", "--retention", "48h", "--verbose", nsName})
	s.NoError(err)
	logs := writer.GetContent()
	s.Contains(logs, "NamespaceInfo.Description")
	s.Contains(logs, "description after")
	s.Contains(logs, "NamespaceInfo.OwnerEmail")
	s.Contains(logs, "email after")
	s.Contains(logs, "Config.WorkflowExecutionRetentionTtl")
	s.Contains(logs, "48h0m0s")
	nsAfter, err := c.WorkflowService().DescribeNamespace(context.Background(), &workflowservice.DescribeNamespaceRequest{Namespace: nsName})
	s.NoError(err)
	s.Equal("description after", nsAfter.GetNamespaceInfo().GetDescription())
	s.Equal("email after", nsAfter.GetNamespaceInfo().GetOwnerEmail())
	s.Equal(float64(48*60*60), nsAfter.GetConfig().GetWorkflowExecutionRetentionTtl().Seconds())
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
			OwnerEmail:                       "email test",
			WorkflowExecutionRetentionPeriod: &retentionBefore,
		},
	)
	s.NoError(err)

	// TODO: remove if namespace cache refresh is not an issue anymore
	time.Sleep(10 * time.Second)

	err = app.Run([]string{"", "operator", "namespace", "update", "--description", "description after", "--email", "email after", "--retention", "48h", nsName})
	s.NoError(err)
	s.NotContains(writer.GetContent(), "NamespaceInfo.Description")
	s.NotContains(writer.GetContent(), "description after")
	s.NotContains(writer.GetContent(), "NamespaceInfo.OwnerEmail")
	nsAfter, err := c.WorkflowService().DescribeNamespace(context.Background(), &workflowservice.DescribeNamespaceRequest{Namespace: nsName})
	s.NoError(err)
	s.Equal("description after", nsAfter.GetNamespaceInfo().GetDescription())
	s.Equal("email after", nsAfter.GetNamespaceInfo().GetOwnerEmail())
	s.Equal(float64(48*60*60), nsAfter.GetConfig().GetWorkflowExecutionRetentionTtl().Seconds())
}

func (s *e2eSuite) TestNamespaceUpdate_NoChanges() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	c := testserver.Client()

	nsName := "test-namespace-update-no-changes"
	retentionBefore := time.Duration(24 * time.Hour)
	_, err := c.WorkflowService().RegisterNamespace(
		context.Background(),
		&workflowservice.RegisterNamespaceRequest{
			Namespace:                        nsName,
			Description:                      "description test",
			OwnerEmail:                       "email test",
			WorkflowExecutionRetentionPeriod: &retentionBefore,
		},
	)
	s.NoError(err)

	// TODO: remove if namespace cache refresh is not an issue anymore
	time.Sleep(10 * time.Second)

	err = app.Run([]string{"", "operator", "namespace", "update", "--description", "description test", "--email", "email test", "--retention", "24h", "--verbose", nsName})
	s.NoError(err)
	s.Contains(writer.GetContent(), "No namespace fields are updated")
}
