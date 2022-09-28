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

	"github.com/pborman/uuid"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/collection"
	"go.temporal.io/server/common/payloads"
)

// DescribeBatchJob describe the status of the batch job
func DescribeBatchJob(c *cli.Context) error {
	namespace, err := requiredFlag(c, FlagNamespace)
	if err != nil {
		return err
	}
	jobID := c.String(FlagJobID)

	client := cFactory.FrontendClient(c)
	ctx, cancel := newContext(c)
	defer cancel()
	resp, err := client.DescribeBatchOperation(ctx, &workflowservice.DescribeBatchOperationRequest{
		Namespace: namespace,
		JobId:     jobID,
	})
	if err != nil {
		return fmt.Errorf("unable to describe batch job: %w", err)
	}

	opts := &output.PrintOptions{
		Fields:     []string{"State", "JobId", "StartTime"},
		FieldsLong: []string{"Identity", "Reason"},
		Pager:      pager.Less,
	}
	return output.PrintItems(c, []interface{}{resp}, opts)
}

// ListBatchJobs list the started batch jobs
func ListBatchJobs(c *cli.Context) error {
	namespace, err := requiredFlag(c, FlagNamespace)
	if err != nil {
		return err
	}
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
		Fields:     []string{"State", "JobId", "StartTime"},
		FieldsLong: []string{"CloseTime"},
		Pager:      pager.Less,
	}
	return output.PrintIterator(c, iter, opts)
}

// BatchTerminate terminate a list of workflows
func BatchTerminate(c *cli.Context) error {
	operator := getCurrentUserFromEnv()

	req := workflowservice.StartBatchOperationRequest{
		Operation: &workflowservice.StartBatchOperationRequest_TerminationOperation{
			TerminationOperation: &batch.BatchOperationTermination{
				Identity: operator,
			},
		},
	}

	return startBatchJob(c, &req)
}

// BatchCancel cancel a list of workflows
func BatchCancel(c *cli.Context) error {
	operator := getCurrentUserFromEnv()

	req := workflowservice.StartBatchOperationRequest{
		Operation: &workflowservice.StartBatchOperationRequest_CancellationOperation{
			CancellationOperation: &batch.BatchOperationCancellation{
				Identity: operator,
			},
		},
	}

	return startBatchJob(c, &req)
}

// BatchSignal send a signal to a list of workflows
func BatchSignal(c *cli.Context) error {
	signalName := c.String(FlagName)
	input := c.String(FlagInput)
	operator := getCurrentUserFromEnv()

	inputP, err := payloads.Encode(input)
	if err != nil {
		return fmt.Errorf("unable to serialize signal input: %w", err)
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

	return startBatchJob(c, &req)
}

// startBatchJob starts a batch job
func startBatchJob(c *cli.Context, req *workflowservice.StartBatchOperationRequest) error {
	namespace, err := requiredFlag(c, FlagNamespace)
	if err != nil {
		return err
	}
	query := c.String(FlagQuery)
	reason := c.String(FlagReason)

	sdk := cFactory.SDKClient(c, namespace)
	tcCtx, cancel := newContext(c)
	defer cancel()
	count, err := sdk.CountWorkflow(tcCtx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     query,
	})
	if err != nil {
		return fmt.Errorf("unable to count impacted workflows: %w", err)
	}

	promptMsg := fmt.Sprintf(
		"Will start a batch job operating on %v Workflow Executions. Continue? Y/N",
		color.Yellow(c, "%v", count.GetCount()),
	)
	if !promptYes(promptMsg, c.Bool(FlagYes)) {
		return nil
	}

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
		return fmt.Errorf("unable to start batch job: %w", err)
	}

	fmt.Printf("Batch job %s is started\n", color.Magenta(c, jobID))
	return nil
}

// StopBatchJob stops a batch job
func StopBatchJob(c *cli.Context) error {
	namespace, err := requiredFlag(c, FlagNamespace)
	if err != nil {
		return err
	}
	jobID := c.String(FlagJobID)
	reason := c.String(FlagReason)
	client := cFactory.FrontendClient(c)

	ctx, cancel := newContext(c)
	defer cancel()
	_, err = client.StopBatchOperation(ctx, &workflowservice.StopBatchOperationRequest{
		Namespace: namespace,
		JobId:     jobID,
		Reason:    reason,
		Identity:  getCurrentUserFromEnv(),
	})

	if err != nil {
		return fmt.Errorf("unable to stop a batch job %s: %w", color.Magenta(c, jobID), err)
	}

	fmt.Printf("Batch job %s is stopped\n", color.Magenta(c, jobID))
	return nil
}
