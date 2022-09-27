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
	"strings"

	"github.com/pborman/uuid"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/collection"
	"go.temporal.io/server/common/payloads"
	"go.temporal.io/server/common/primitives"
	"go.temporal.io/server/service/worker/batcher"
)

// DescribeBatchJob describe the status of the batch job
func DescribeBatchJob(c *cli.Context) error {
	jobID := c.String(FlagJobID)

	client := cFactory.FrontendClient(c)
	ctx, cancel := newContext(c)
	defer cancel()
	resp, err := client.DescribeBatchOperation(ctx, &workflowservice.DescribeBatchOperationRequest{
		Namespace: primitives.SystemLocalNamespace,
		JobId:     jobID,
	})
	if err != nil {
		return fmt.Errorf("failed to describe batch job: %w", err)
	}

	opts := &output.PrintOptions{
		Fields:     []string{"State", "JobId", "StartTime"},
		FieldsLong: []string{"Identity", "Reason"},
		Pager:      pager.Less,
	}
	output.PrintItems(c, []interface{}{resp}, opts)
	return nil
}

// ListBatchJobs list the started batch jobs
func ListBatchJobs(c *cli.Context) error {
	client := cFactory.FrontendClient(c)

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		var items []interface{}
		var err error

		ctx, cancel := newContext(c)
		defer cancel()
		resp, err := client.ListBatchOperations(ctx, &workflowservice.ListBatchOperationsRequest{
			Namespace: primitives.SystemLocalNamespace,
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

// StartBatchJob starts a batch job
func StartBatchJob(c *cli.Context) error {
	namespace, err := getRequiredGlobalOption(c, FlagNamespace)
	if err != nil {
		return err
	}
	query := c.String(FlagQuery)
	reason := c.String(FlagReason)
	batchType := c.String(FlagType)
	operator := getCurrentUserFromEnv()

	sdk := cFactory.SDKClient(c, namespace)
	tcCtx, cancel := newContext(c)
	defer cancel()
	cResp, err := sdk.CountWorkflow(tcCtx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     query,
	})
	if err != nil {
		return fmt.Errorf("unable to count impacted workflows: %v", err)
	}

	promptMsg := fmt.Sprintf(
		"Will start a batch job operating on %v Workflow Executions. Continue? Y/N",
		color.Yellow(c, "%v", cResp.GetCount()),
	)
	if !promptYes(promptMsg, c.Bool(FlagYes)) {
		return nil
	}

	input := c.String(FlagInput)
	payloads, err := payloads.Encode(input)
	if err != nil {
		return fmt.Errorf("failed to serialize signal value: %w", err)
	}

	jobID := uuid.New()
	req := workflowservice.StartBatchOperationRequest{
		Namespace:       namespace,
		Reason:          reason,
		VisibilityQuery: query,
		JobId:           jobID,
	}

	switch batchType {
	case batcher.BatchTypeSignal:
		sigName := c.String(FlagSignalName)

		if sigName == "" {
			return fmt.Errorf("option %s are required for type %s", FlagSignalName, batcher.BatchTypeSignal)
		}

		req.Operation = &workflowservice.StartBatchOperationRequest_SignalOperation{
			SignalOperation: &batch.BatchOperationSignal{
				Signal:   sigName,
				Input:    payloads,
				Identity: operator,
			},
		}
	case batcher.BatchTypeTerminate:
		req.Operation = &workflowservice.StartBatchOperationRequest_TerminationOperation{
			TerminationOperation: &batch.BatchOperationTermination{
				Identity: operator,
			},
		}
	case batcher.BatchTypeCancel:
		req.Operation = &workflowservice.StartBatchOperationRequest_CancellationOperation{
			CancellationOperation: &batch.BatchOperationCancellation{
				Identity: operator,
			},
		}
	default:
		return fmt.Errorf("unknown batch type. Supported types: %s", strings.Join(allBatchTypes, ","))
	}

	client := cFactory.FrontendClient(c)
	ctx, cancel := newContext(c)
	defer cancel()
	_, err = client.StartBatchOperation(ctx, &req)
	if err != nil {
		return fmt.Errorf("unable to start batch job: %s", err)
	}

	fmt.Printf("Batch job %s is started\n", color.Magenta(c, jobID))
	return nil
}

// StopBatchJob stops a batch job
func StopBatchJob(c *cli.Context) error {
	jobID := c.String(FlagJobID)
	reason := c.String(FlagReason)
	client := cFactory.FrontendClient(c)

	ctx, cancel := newContext(c)
	defer cancel()
	_, err := client.StopBatchOperation(ctx, &workflowservice.StopBatchOperationRequest{
		Namespace: primitives.SystemLocalNamespace,
		JobId:     jobID,
		Reason:    reason,
		Identity:  getCurrentUserFromEnv(),
	})

	if err != nil {
		return fmt.Errorf("unable to stop a batch job %s: %s", color.Magenta(c, jobID), err)
	}

	fmt.Printf("Batch job %s is stopped\n", color.Magenta(c, jobID))
	return nil
}
