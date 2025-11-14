package temporalcli

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/temporalio/cli/internal/printer"
)

type (
	batchDescribe struct {
		State          string
		Type           string
		StartTime      time.Time
		CloseTime      time.Time `cli:",cardOmitEmpty"`
		CompletedCount string
		FailureCount   string
	}
	batchTableRow struct {
		JobId     string
		State     string
		StartTime time.Time
		CloseTime time.Time
	}
)

func (c TemporalBatchDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().DescribeBatchOperation(cctx, &workflowservice.DescribeBatchOperationRequest{
		Namespace: c.Parent.Namespace,
		JobId:     c.JobId,
	})
	var notFound *serviceerror.NotFound
	if errors.As(err, &notFound) {
		return fmt.Errorf("could not find Batch Job '%v'", c.JobId)
	} else if err != nil {
		return fmt.Errorf("failed to describe batch job: %w", err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	_ = cctx.Printer.PrintStructured(batchDescribe{
		Type:           resp.OperationType.String(),
		State:          resp.State.String(),
		StartTime:      toTime(resp.StartTime),
		CloseTime:      toTime(resp.CloseTime),
		CompletedCount: fmt.Sprintf("%d/%d", resp.CompleteOperationCount, resp.TotalOperationCount),
		FailureCount:   fmt.Sprintf("%d/%d", resp.FailureOperationCount, resp.TotalOperationCount),
	}, printer.StructuredOptions{})

	return nil
}

func (c TemporalBatchListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// This is a listing command subject to json vs jsonl rules
	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	pageFetcher := c.pageFetcher(cctx, cl)
	var nextPageToken []byte
	var jobsProcessed int
	for pageIndex := 0; ; pageIndex++ {
		page, err := pageFetcher(nextPageToken)
		if err != nil {
			return fmt.Errorf("failed to list batch jobs: %w", err)
		}

		if pageIndex == 0 && len(page.GetOperationInfo()) == 0 {
			return nil
		}

		var textTable []batchTableRow
		for _, job := range page.GetOperationInfo() {
			if c.Limit > 0 && jobsProcessed >= c.Limit {
				break
			}
			jobsProcessed++
			// For JSON we are going to dump one line of JSON per execution
			if cctx.JSONOutput {
				_ = cctx.Printer.PrintStructured(job, printer.StructuredOptions{})
			} else {
				// For non-JSON, we are doing a table for each page
				textTable = append(textTable, batchTableRow{
					JobId:     job.JobId,
					State:     job.State.String(),
					StartTime: toTime(job.StartTime),
					CloseTime: toTime(job.CloseTime),
				})
			}
		}
		// Print table, headers only on first table
		if len(textTable) > 0 {
			_ = cctx.Printer.PrintStructured(textTable, printer.StructuredOptions{
				Table: &printer.TableOptions{NoHeader: pageIndex > 0},
			})
		}
		// Stop if next page token non-existing or list reached limit
		nextPageToken = page.GetNextPageToken()
		if len(nextPageToken) == 0 || (c.Limit > 0 && jobsProcessed >= c.Limit) {
			return nil
		}
	}
}

func (c *TemporalBatchListCommand) pageFetcher(
	cctx *CommandContext,
	cl client.Client,
) func(next []byte) (*workflowservice.ListBatchOperationsResponse, error) {
	return func(next []byte) (*workflowservice.ListBatchOperationsResponse, error) {
		return cl.WorkflowService().ListBatchOperations(cctx, &workflowservice.ListBatchOperationsRequest{
			Namespace:     c.Parent.Namespace,
			NextPageToken: next,
		})
	}
}

func (c TemporalBatchTerminateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().StopBatchOperation(cctx, &workflowservice.StopBatchOperationRequest{
		Namespace: c.Parent.Namespace,
		JobId:     c.JobId,
		Reason:    c.Reason,
		Identity:  c.Parent.Identity,
	})

	var notFound *serviceerror.NotFound
	if errors.As(err, &notFound) {
		return fmt.Errorf("could not find Batch Job '%v'", c.JobId)
	} else if err != nil {
		return fmt.Errorf("failed to terminate batch job: %w", err)
	}

	cctx.Printer.Printlnf("Terminated Batch Job '%v'", c.JobId)

	return nil
}

// Converts the timestamp to Go's native time.Time.
// Returns the zero time.Time value for nil timestamp.
func toTime(timestamp *timestamppb.Timestamp) (t time.Time) {
	if timestamp != nil {
		t = timestamp.AsTime()
	}
	return
}
