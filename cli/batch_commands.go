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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/server/common/collection"
	"go.temporal.io/server/common/primitives"

	"go.temporal.io/server/common/payloads"
	"go.temporal.io/server/common/searchattribute"
	"go.temporal.io/server/service/worker/batcher"
)

// DescribeBatchJob describe the status of the batch job
func DescribeBatchJob(c *cli.Context) error {
	jobID := c.String(FlagJobID)

	client := cFactory.SDKClient(c, primitives.SystemLocalNamespace)
	tcCtx, cancel := newContext(c)
	defer cancel()
	wf, err := client.DescribeWorkflowExecution(tcCtx, jobID, "")
	if err != nil {
		return fmt.Errorf("failed to describe batch job: %w", err)
	}

	output := map[string]interface{}{}
	if wf.WorkflowExecutionInfo.GetStatus() != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		if wf.WorkflowExecutionInfo.GetStatus() != enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED {
			output["msg"] = "batch job stopped status: " + wf.WorkflowExecutionInfo.GetStatus().String()
		} else {
			output["msg"] = "batch job is finished successfully"
		}
	} else {
		output["msg"] = "batch job is running"
		if len(wf.PendingActivities) > 0 {
			hbdPayload := wf.PendingActivities[0].HeartbeatDetails
			var hbd batcher.HeartBeatDetails
			err := payloads.Decode(hbdPayload, &hbd)
			if err != nil {
				return fmt.Errorf("failed to describe batch job: %w", err)
			}
			output["progress"] = hbd
		}
	}
	prettyPrintJSONObject(output)
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
	query := c.String(FlagListQuery)
	reason := c.String(FlagReason)
	batchType := c.String(FlagType)
	if !validateBatchType(batchType) {
		return fmt.Errorf("unknown batch type, supported types: %s", strings.Join(allBatchTypes, ","))
	}
	operator := getCurrentUserFromEnv()
	var sigName, sigVal string
	if batchType == batcher.BatchTypeSignal {
		sigName = c.String(FlagSignalName)
		sigVal = c.String(FlagInput)

		if len(sigName) == 0 || len(sigVal) == 0 {
			return fmt.Errorf("options %s and %s are required for type %s", FlagSignalName, FlagInput, batcher.BatchTypeSignal)
		}
	}
	rps := c.Int(FlagRPS)
	concurrency := c.Int(FlagConcurrency)

	client := cFactory.SDKClient(c, primitives.SystemLocalNamespace)
	tcCtx, cancel := newContext(c)
	defer cancel()
	resp, err := client.CountWorkflow(tcCtx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     query,
	})
	if err != nil {
		return fmt.Errorf("failed to count impacted workflows: %w", err)
	}
	fmt.Printf("This batch job will be operating on %v workflows, with max RPS of %v and concurrency of %v.\n",
		resp.GetCount(), rps, concurrency)
	if !c.Bool(FlagYes) {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Please confirm[Yes/No]:")
			text, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to get confirmation to start a batch job: %w", err)
			}
			if strings.EqualFold(strings.TrimSpace(text), "yes") {
				break
			} else {
				fmt.Println("Batch job is not started")
				return nil
			}
		}

	}
	tcCtx, cancel = newContext(c)
	defer cancel()
	options := sdkclient.StartWorkflowOptions{
		TaskQueue: "temporal-sys-batcher-taskqueue",
		Memo: map[string]interface{}{
			"Reason": reason,
		},
		SearchAttributes: map[string]interface{}{
			searchattribute.BatcherNamespace: namespace,
			searchattribute.BatcherUser:      operator,
		},
	}

	sigInput, err := payloads.Encode(sigVal)
	if err != nil {
		return fmt.Errorf("failed to serialize signal value: %w", err)
	}

	params := batcher.BatchParams{
		Namespace: namespace,
		Query:     query,
		Reason:    reason,
		BatchType: batchType,
		SignalParams: batcher.SignalParams{
			SignalName: sigName,
			Input:      sigInput,
		},
		RPS:         rps,
		Concurrency: concurrency,
	}
	wf, err := client.ExecuteWorkflow(tcCtx, options, batcher.BatchWFTypeName, params)
	if err != nil {
		return fmt.Errorf("failed to start batch job: %w", err)
	}
	output := map[string]interface{}{
		"msg":   "batch job is started",
		"jobId": wf.GetID(),
	}
	prettyPrintJSONObject(output)
	return nil
}

// TerminateBatchJob stops abatch job
func TerminateBatchJob(c *cli.Context) error {
	jobID := c.String(FlagJobID)
	reason := c.String(FlagReason)
	client := cFactory.SDKClient(c, primitives.SystemLocalNamespace)
	tcCtx, cancel := newContext(c)
	defer cancel()
	err := client.TerminateWorkflow(tcCtx, jobID, "", reason, nil)
	if err != nil {
		return fmt.Errorf("failed to terminate batch job: %w", err)
	}
	output := map[string]interface{}{
		"msg": "batch job is terminated",
	}
	prettyPrintJSONObject(output)
	return nil
}

func validateBatchType(bt string) bool {
	for _, b := range allBatchTypes {
		if b == bt {
			return true
		}
	}
	return false
}
