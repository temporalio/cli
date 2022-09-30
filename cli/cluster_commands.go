// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"fmt"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/collection"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	fullWorkflowServiceName = "temporal.api.workflowservice.v1.WorkflowService"
)

// HealthCheck check frontend health.
func HealthCheck(c *cli.Context) error {
	healthClient := cFactory.HealthClient(c)
	ctx, cancel := newContext(c)
	defer cancel()

	req := &healthpb.HealthCheckRequest{
		Service: fullWorkflowServiceName,
	}
	resp, err := healthClient.Check(ctx, req)

	if err != nil {
		return fmt.Errorf("unable to health check %q service: %w", req.GetService(), err)
	}

	fmt.Printf("%s: ", req.GetService())
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		fmt.Println(color.Red(c, "%v", resp.Status))
		return nil
	}

	fmt.Println(color.Green(c, "%v", resp.Status))
	return nil
}

func DescribeCluster(c *cli.Context) error {
	client := cFactory.FrontendClient(c)
	ctx, cancel := newContext(c)
	defer cancel()

	cluster, err := client.GetClusterInfo(ctx, &workflowservice.GetClusterInfoRequest{})
	if err != nil {
		return fmt.Errorf("unable to get cluster information: %w", err)
	}

	po := &output.PrintOptions{
		Fields:     []string{"ClusterName", "PersistenceStore", "VisibilityStore"},
		FieldsLong: []string{"HistoryShardCount", "VersionInfo"},
	}
	return output.PrintItems(c, []interface{}{cluster}, po)
}

func DescribeSystem(c *cli.Context) error {
	client := cFactory.FrontendClient(c)
	ctx, cancel := newContext(c)
	defer cancel()

	system, err := client.GetSystemInfo(ctx, &workflowservice.GetSystemInfoRequest{})
	if err != nil {
		return fmt.Errorf("unable to get system information: %w", err)
	}

	po := &output.PrintOptions{
		Fields:     []string{"ServerVersion", "Capabilities.SupportsSchedules", "Capabilities.UpsertMemo"},
		FieldsLong: []string{"Capabilities.SignalAndQueryHeader", "Capabilities.ActivityFailureIncludeHeartbeat", "Capabilities.InternalErrorDifferentiation"},
	}
	return output.PrintItems(c, []interface{}{system}, po)
}

func UpsertCluster(c *cli.Context) error {
	client := cFactory.OperatorClient(c)
	ctx, cancel := newContext(c)
	defer cancel()

	address := c.String(FlagClusterAddress)

	_, err := client.AddOrUpdateRemoteCluster(ctx, &operatorservice.AddOrUpdateRemoteClusterRequest{
		FrontendAddress:               address,
		EnableRemoteClusterConnection: c.Bool(FlagClusterEnableConnection),
	})
	if err != nil {
		return fmt.Errorf("unable to upsert cluster: %w", err)
	}

	fmt.Println(color.Green(c, "Upserted cluster %s", address))
	return nil
}

func ListClusters(c *cli.Context) error {
	client := cFactory.OperatorClient(c)

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		var items []interface{}
		var err error

		ctx, cancel := newContext(c)
		defer cancel()
		resp, err := client.ListClusters(ctx, &operatorservice.ListClustersRequest{})
		if err != nil {
			return nil, nil, fmt.Errorf("unable to list clusters: %w", err)
		}

		for _, e := range resp.Clusters {
			items = append(items, e)
		}

		if err != nil {
			return nil, nil, err
		}

		return items, npt, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &output.PrintOptions{
		Fields:     []string{"ClusterName", "Address", "IsConnectionEnabled"},
		FieldsLong: []string{"ClusterId", "InitialFailoverVersion", "HistoryShardCount"},
		Pager:      pager.Less,
	}
	return output.PrintIterator(c, iter, opts)

}

func RemoveCluster(c *cli.Context) error {
	client := cFactory.OperatorClient(c)
	ctx, cancel := newContext(c)
	defer cancel()

	name := c.String(FlagName)

	_, err := client.RemoveRemoteCluster(ctx, &operatorservice.RemoveRemoteClusterRequest{
		ClusterName: c.String(FlagName),
	})
	if err != nil {
		return fmt.Errorf("unable to remove cluster: %w", err)
	}

	fmt.Println(color.Green(c, "Removed cluster %s", name))
	return nil
}
