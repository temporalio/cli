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

package cli_curr

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pborman/uuid"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/collection"
	"go.temporal.io/server/common/payloads"
)

// DescribeBatchJobV2 describe the status of the batch job
func DescribeBatchJobV2(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	jobID := c.String(FlagJobID)

	client := cFactory.FrontendClient(c)
	ctx, cancel := newContext(c)
	defer cancel()
	resp, err := client.DescribeBatchOperation(ctx, &workflowservice.DescribeBatchOperationRequest{
		Namespace: namespace,
		JobId:     jobID,
	})
	if err != nil {
		ErrorAndExit("unable to describe batch job", err)
	}

	opts := &output.PrintOptions{
		OutputFormat: output.JSON,
	}
	output.PrintItems(nil, []interface{}{resp}, opts)
}

// ListBatchJobs list the started batch jobs
func ListBatchJobsV2(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	client := cFactory.FrontendClient(c)

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		var items []interface{}
		var err error

		ctx, cancel := newContext(c)
		defer cancel()
		resp, err := client.ListBatchOperations(ctx, &workflowservice.ListBatchOperationsRequest{
			Namespace: namespace,
		})

		for _, e := range resp.OperationInfo {
			items = append(items, e)
		}

		if err != nil {
			return nil, nil, err
		}

		return items, npt, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &output.PrintOptions{
		Fields: []string{"State", "JobId", "StartTime", "CloseTime"},
		Pager:  pager.Less,
	}
	output.PrintIterator(nil, iter, opts)
}

// BatchTerminateV2 terminate a list of workflows
func BatchTerminateV2(c *cli.Context) {
	operator := getCurrentUserFromEnv()

	req := workflowservice.StartBatchOperationRequest{
		Operation: &workflowservice.StartBatchOperationRequest_TerminationOperation{
			TerminationOperation: &batch.BatchOperationTermination{
				Identity: operator,
			},
		},
	}

	startBatchJob(c, &req)
}

// BatchCancelV2 cancel a list of workflows
func BatchCancelV2(c *cli.Context) {
	operator := getCurrentUserFromEnv()

	req := workflowservice.StartBatchOperationRequest{
		Operation: &workflowservice.StartBatchOperationRequest_CancellationOperation{
			CancellationOperation: &batch.BatchOperationCancellation{
				Identity: operator,
			},
		},
	}

	startBatchJob(c, &req)
}

// BatchSignalV2 send a signal to a list of workflows
func BatchSignalV2(c *cli.Context) {
	signalName := c.String(FlagName)
	input := c.String(FlagInput)
	operator := getCurrentUserFromEnv()

	inputP, err := payloads.Encode(input)
	if err != nil {
		ErrorAndExit("unable to serialize signal input", err)
	}

	req := workflowservice.StartBatchOperationRequest{
		Operation: &workflowservice.StartBatchOperationRequest_SignalOperation{
			SignalOperation: &batch.BatchOperationSignal{
				Signal:   signalName,
				Identity: operator,
				Input:    inputP,
			},
		},
	}

	startBatchJob(c, &req)
}

// startBatchJob starts a batch job
func startBatchJob(c *cli.Context, req *workflowservice.StartBatchOperationRequest) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	query := c.String(FlagListQuery)
	reason := c.String(FlagReason)

	sdk := cFactory.SDKClient(c, namespace)
	tcCtx, cancel := newContext(c)
	defer cancel()
	count, err := sdk.CountWorkflow(tcCtx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     query,
	})
	if err != nil {
		ErrorAndExit("unable to count impacted workflows", err)
	}

	msg := fmt.Sprintf("Will start a batch job operating on %v Workflow Executions. Continue? Y/N", count.GetCount())
	prompt(msg, c.Bool(FlagYes))

	jobID := uuid.New()
	req.JobId = jobID
	req.Namespace = namespace
	req.VisibilityQuery = query
	req.Reason = reason

	client := cFactory.FrontendClient(c)
	ctx, cancel := newContext(c)
	defer cancel()
	_, err = client.StartBatchOperation(ctx, req)
	if err != nil {
		ErrorAndExit("unable to start batch job", err)
	}

	fmt.Printf("Batch job %s is started\n", color.MagentaString(jobID))
}

// StopBatchJobV2 stops a batch job
func StopBatchJobV2(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	jobID := c.String(FlagJobID)
	reason := c.String(FlagReason)
	client := cFactory.FrontendClient(c)

	ctx, cancel := newContext(c)
	defer cancel()
	_, err := client.StopBatchOperation(ctx, &workflowservice.StopBatchOperationRequest{
		Namespace: namespace,
		JobId:     jobID,
		Reason:    reason,
		Identity:  getCurrentUserFromEnv(),
	})

	if err != nil {
		ErrorAndExit("unable to stop a batch job", err)
	}

	fmt.Printf("Batch job %s is stopped\n", color.MagentaString(jobID))
}
