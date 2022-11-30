package batch

import (
	"fmt"

	"github.com/pborman/uuid"
	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
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
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	jobID := c.String(common.FlagJobID)

	client := client.CFactory.FrontendClient(c)
	ctx, cancel := common.NewContext(c)
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
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	client := client.CFactory.FrontendClient(c)

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		var items []interface{}
		var err error

		ctx, cancel := common.NewContext(c)
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
	operator := common.GetCurrentUserFromEnv()

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
	operator := common.GetCurrentUserFromEnv()

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
	signalName := c.String(common.FlagName)
	input := c.String(common.FlagInput)
	operator := common.GetCurrentUserFromEnv()

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
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	query := c.String(common.FlagQuery)
	reason := c.String(common.FlagReason)

	sdk := client.CFactory.SDKClient(c, namespace)
	tcCtx, cancel := common.NewContext(c)
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
	if !common.PromptYes(promptMsg, c.Bool(common.FlagYes)) {
		return nil
	}

	jobID := uuid.New()
	req.JobId = jobID
	req.Namespace = namespace
	req.VisibilityQuery = query
	req.Reason = reason

	client := client.CFactory.FrontendClient(c)
	ctx, cancel := common.NewContext(c)
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
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	jobID := c.String(common.FlagJobID)
	reason := c.String(common.FlagReason)
	client := client.CFactory.FrontendClient(c)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	_, err = client.StopBatchOperation(ctx, &workflowservice.StopBatchOperationRequest{
		Namespace: namespace,
		JobId:     jobID,
		Reason:    reason,
		Identity:  common.GetCurrentUserFromEnv(),
	})

	if err != nil {
		return fmt.Errorf("unable to stop a batch job %s: %w", color.Magenta(c, jobID), err)
	}

	fmt.Printf("Batch job %s is stopped\n", color.Magenta(c, jobID))
	return nil
}
