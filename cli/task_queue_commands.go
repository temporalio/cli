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

	enumspb "go.temporal.io/api/enums/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

// DescribeTaskQueue show pollers info of a given taskqueue
func DescribeTaskQueue(c *cli.Context) error {
	sdkClient, err := getSDKClient(c)
	if err != nil {
		return err
	}
	taskQueue := c.String(FlagTaskQueue)
	taskQueueType := strToTaskQueueType(c.String(FlagTaskQueueType))

	ctx, cancel := newContext(c)
	defer cancel()
	resp, err := sdkClient.DescribeTaskQueue(ctx, taskQueue, taskQueueType)
	if err != nil {
		return fmt.Errorf("unable to describe task queue: %w", err)
	}

	opts := &output.PrintOptions{
		// TODO enable when versioning feature is out
		// Fields: []string{"Identity", "LastAccessTime", "RatePerSecond", "WorkerVersioningId"},
		Fields: []string{"Identity", "LastAccessTime", "RatePerSecond"},
	}
	var items []interface{}
	for _, e := range resp.Pollers {
		items = append(items, e)
	}
	return output.PrintItems(c, items, opts)
}

// ListTaskQueuePartitions gets all the taskqueue partition and host information.
func ListTaskQueuePartitions(c *cli.Context) error {
	frontendClient := cFactory.FrontendClient(c)
	namespace, err := requiredFlag(c, FlagNamespace)
	if err != nil {
		return err
	}
	taskQueue := c.String(FlagTaskQueue)

	ctx, cancel := newContext(c)
	defer cancel()
	request := &workflowservice.ListTaskQueuePartitionsRequest{
		Namespace: namespace,
		TaskQueue: &taskqueuepb.TaskQueue{
			Name: taskQueue,
			Kind: enumspb.TASK_QUEUE_KIND_NORMAL,
		},
	}

	resp, err := frontendClient.ListTaskQueuePartitions(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to list task queues: %w", err)
	}

	optsW := &output.PrintOptions{
		Fields: []string{"Key", "OwnerHostName"},
	}

	var items []interface{}
	fmt.Println(color.Magenta(c, "Workflow Task Queue Partitions\n"))
	for _, e := range resp.WorkflowTaskQueuePartitions {
		items = append(items, e)
	}
	err = output.PrintItems(c, items, optsW)
	if err != nil {
		return err
	}

	optsA := &output.PrintOptions{
		Fields: []string{"Key", "OwnerHostName"},
	}
	items = items[:0]
	fmt.Println(color.Magenta(c, "\nActivity Task Queue Partitions\n"))
	for _, e := range resp.ActivityTaskQueuePartitions {
		items = append(items, e)
	}
	return output.PrintItems(c, items, optsA)
}
