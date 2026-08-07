package temporalcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

func (c *TemporalNexusOperationStartCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	handle, err := startNexusOperation(cctx, cl, &c.NexusOperationStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}
	return printNexusOperationExecution(cctx, &c.NexusOperationStartOptions, handle.GetID(), handle.GetRunID(), c.Parent.Parent.Namespace)
}

func (c *TemporalNexusOperationExecuteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	handle, err := startNexusOperation(cctx, cl, &c.NexusOperationStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}
	if !cctx.JSONOutput {
		if err := printNexusOperationExecution(cctx, &c.NexusOperationStartOptions, handle.GetID(), handle.GetRunID(), c.Parent.Parent.Namespace); err != nil {
			cctx.Logger.Error("Failed printing execution info", "error", err)
		}
	}
	return getNexusOperationResult(cctx, cl, c.Parent.Parent.Namespace, handle.GetID(), handle.GetRunID())
}

func (c *TemporalNexusOperationResultCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	return getNexusOperationResult(cctx, cl, c.Parent.Parent.Namespace, c.OperationId, c.RunId)
}

func startNexusOperation(
	cctx *CommandContext,
	cl client.Client,
	opts *NexusOperationStartOptions,
	inputOpts *PayloadInputOptions,
) (client.NexusOperationHandle, error) {
	nexusCl, startOpts, err := buildNexusStartOptions(opts, inputOpts)
	if err != nil {
		return nil, err
	}
	nCl, err := cl.NewNexusClient(nexusCl)
	if err != nil {
		return nil, err
	}
	handle, err := nCl.ExecuteOperation(cctx, startOpts.operation, startOpts.input, startOpts.options)
	if err != nil {
		return nil, fmt.Errorf("failed starting nexus operation: %w", err)
	}
	return handle, nil
}

func printNexusOperationExecution(cctx *CommandContext, opts *NexusOperationStartOptions, operationID, runID, namespace string) error {
	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString("Started Nexus Operation:"))
	}
	return cctx.Printer.PrintStructured(struct {
		Endpoint    string `json:"endpoint"`
		Service     string `json:"service"`
		Operation   string `json:"operation"`
		OperationId string `json:"operationId"`
		RunId       string `json:"runId"`
		Namespace   string `json:"namespace"`
	}{
		Endpoint:    opts.Endpoint,
		Service:     opts.Service,
		Operation:   opts.Operation,
		OperationId: operationID,
		RunId:       runID,
		Namespace:   namespace,
	}, printer.StructuredOptions{})
}

func getNexusOperationResult(cctx *CommandContext, cl client.Client, namespace, operationID, runID string) error {
	resp, err := pollNexusOperationOutcome(cctx, cl, namespace, operationID, runID)
	if err != nil {
		var notFound *serviceerror.NotFound
		if errors.As(err, &notFound) {
			return fmt.Errorf("nexus operation not found: %s", operationID)
		}
		return fmt.Errorf("failed polling nexus operation result: %w", err)
	}

	// Use the run ID from the response if the caller didn't supply one.
	if runID == "" {
		runID = resp.GetRunId()
	}

	switch v := resp.GetOutcome().(type) {
	case *workflowservice.PollNexusOperationExecutionResponse_Result:
		return printNexusOperationResult(cctx, operationID, runID, v.Result)
	case *workflowservice.PollNexusOperationExecutionResponse_Failure:
		if err := printNexusOperationFailure(cctx, operationID, runID, v.Failure); err != nil {
			cctx.Logger.Error("Nexus operation failed, and printing the output also failed", "error", err)
		}
		return fmt.Errorf("nexus operation failed")
	default:
		return fmt.Errorf("unexpected nexus operation outcome type: %T", v)
	}
}

// Matches the SDK's pollNexusOperationTimeout in internal_nexus_client.go.
const pollNexusOperationTimeout = 60 * time.Second

// pollNexusOperationOutcome polls for a nexus operation result using a
// hand-rolled loop rather than handle.Get() because handle.Get() deserializes
// the result into a Go value and converts failures to Go errors, losing the
// raw proto payloads.
func pollNexusOperationOutcome(cctx *CommandContext, cl client.Client, namespace, operationID, runID string) (*workflowservice.PollNexusOperationExecutionResponse, error) {
	for {
		pollCtx, cancel := context.WithTimeout(cctx, pollNexusOperationTimeout)
		resp, err := cl.WorkflowService().PollNexusOperationExecution(pollCtx, &workflowservice.PollNexusOperationExecutionRequest{
			Namespace:   namespace,
			OperationId: operationID,
			RunId:       runID,
			WaitStage:   enumspb.NEXUS_OPERATION_WAIT_STAGE_CLOSED,
		})
		if err != nil {
			// check pollCtx.Err() first because it is set by cancel()
			pollTimedOut := pollCtx.Err() != nil
			cancel()
			if cctx.Err() != nil {
				return nil, cctx.Err()
			}
			if pollTimedOut {
				continue
			}
			return nil, err
		}
		cancel()
		if resp.GetOutcome() != nil {
			return resp, nil
		}
	}
}

func printNexusOperationResult(cctx *CommandContext, operationID, runID string, result *common.Payload) error {
	if cctx.JSONOutput {
		var resultJSON json.RawMessage
		var err error
		if cctx.JSONShorthandPayloads {
			var valuePtr any
			if err = converter.GetDefaultDataConverter().FromPayload(result, &valuePtr); err != nil {
				return fmt.Errorf("nexus operation completed, but failed decoding result for json output: %w", err)
			}
			resultJSON, err = json.Marshal(valuePtr)
		} else {
			resultJSON, err = cctx.MarshalProtoJSON(result)
		}
		if err != nil {
			return fmt.Errorf("nexus operation completed, but failed marshaling result for json output: %w", err)
		}
		return cctx.Printer.PrintStructured(struct {
			OperationId string          `json:"operationId"`
			RunId       string          `json:"runId"`
			Status      string          `json:"status"`
			Result      json.RawMessage `json:"result"`
		}{
			OperationId: operationID,
			RunId:       runID,
			Status:      "COMPLETED",
			Result:      resultJSON,
		}, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Results:"))
	var valuePtr any
	if err := converter.GetDefaultDataConverter().FromPayload(result, &valuePtr); err != nil {
		return fmt.Errorf("nexus operation completed, but failed decoding result: %w", err)
	}
	resultJSON, err := json.Marshal(valuePtr)
	if err != nil {
		return fmt.Errorf("nexus operation completed, but failed marshaling result: %w", err)
	}
	return cctx.Printer.PrintStructured(struct {
		Status string
		Result json.RawMessage `cli:",cardOmitEmpty"`
	}{
		Status: color.GreenString("COMPLETED"),
		Result: resultJSON,
	}, printer.StructuredOptions{})
}

func printNexusOperationFailure(cctx *CommandContext, operationID, runID string, f *failure.Failure) error {
	if cctx.JSONOutput {
		failureJSON, err := cctx.MarshalProtoJSON(f)
		if err != nil {
			return fmt.Errorf("nexus operation failed, but failed marshaling failure for json output: %w", err)
		}
		return cctx.Printer.PrintStructured(struct {
			OperationId string          `json:"operationId"`
			RunId       string          `json:"runId"`
			Status      string          `json:"status"`
			Failure     json.RawMessage `json:"failure"`
		}{
			OperationId: operationID,
			RunId:       runID,
			Status:      "FAILED",
			Failure:     failureJSON,
		}, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Results:"))
	return cctx.Printer.PrintStructured(struct {
		Status  string
		Failure string `cli:",cardOmitEmpty"`
	}{
		Status:  color.RedString("FAILED"),
		Failure: cctx.MarshalFriendlyFailureBodyText(f, "    "),
	}, printer.StructuredOptions{})
}

func (c *TemporalNexusOperationDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	handle := cl.GetNexusOperationHandle(client.GetNexusOperationHandleOptions{
		OperationID: c.OperationId,
		RunID:       c.RunId,
	})

	desc, err := handle.Describe(cctx, client.DescribeNexusOperationOptions{})
	if err != nil {
		return fmt.Errorf("failed describing nexus operation: %w", err)
	}

	if c.Raw || cctx.JSONOutput {
		return cctx.Printer.PrintStructured(desc.RawInfo, printer.StructuredOptions{})
	}
	return printNexusOperationDescription(cctx, desc)
}

func printNexusOperationDescription(cctx *CommandContext, desc *client.NexusOperationExecutionDescription) error {
	summary, _ := desc.GetSummary()
	d := struct {
		OperationId            string
		RunId                  string
		Endpoint               string
		Service                string
		Operation              string
		Status                 string
		State                  string `cli:",cardOmitEmpty"`
		Attempt                int32
		ScheduleToCloseTimeout time.Duration `cli:",cardOmitEmpty"`
		ScheduledTime          time.Time     `cli:",cardOmitEmpty"`
		CloseTime              time.Time     `cli:",cardOmitEmpty"`
		ExpirationTime         time.Time     `cli:",cardOmitEmpty"`
		BlockedReason          string        `cli:",cardOmitEmpty"`
		OperationToken         string        `cli:",cardOmitEmpty"`
		Identity               string        `cli:",cardOmitEmpty"`
		Summary                string        `cli:",cardOmitEmpty"`
	}{
		OperationId:            desc.OperationID,
		RunId:                  desc.OperationRunID,
		Endpoint:               desc.Endpoint,
		Service:                desc.Service,
		Operation:              desc.Operation,
		Status:                 desc.Status.String(),
		State:                  desc.State.String(),
		Attempt:                desc.Attempt,
		ScheduleToCloseTimeout: desc.ScheduleToCloseTimeout,
		ScheduledTime:          desc.ScheduledTime,
		CloseTime:              desc.CloseTime,
		ExpirationTime:         desc.ExpirationTime,
		BlockedReason:          desc.BlockedReason,
		OperationToken:         desc.OperationToken,
		Identity:               desc.Identity,
		Summary:                summary,
	}
	if err := cctx.Printer.PrintStructured(d, printer.StructuredOptions{}); err != nil {
		return err
	}
	return printLinks(cctx, desc.RawInfo.GetLinks())
}

func (c *TemporalNexusOperationCancelCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	handle := cl.GetNexusOperationHandle(client.GetNexusOperationHandleOptions{
		OperationID: c.OperationId,
		RunID:       c.RunId,
	})

	err = handle.Cancel(cctx, client.CancelNexusOperationOptions{
		Reason: c.Reason,
	})
	if err != nil {
		return fmt.Errorf("failed to request nexus operation cancellation: %w", err)
	}
	cctx.Printer.Println("Cancellation requested")
	return nil
}

func (c *TemporalNexusOperationTerminateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	reason := c.Reason
	if reason == "" {
		reason = defaultReason()
	}
	handle := cl.GetNexusOperationHandle(client.GetNexusOperationHandleOptions{
		OperationID: c.OperationId,
		RunID:       c.RunId,
	})
	if err := handle.Terminate(cctx, client.TerminateNexusOperationOptions{Reason: reason}); err != nil {
		return fmt.Errorf("failed to terminate nexus operation: %w", err)
	}
	cctx.Printer.Println("Nexus Operation terminated")
	return nil
}

func (c *TemporalNexusOperationListCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	if c.Limit > 0 && c.Limit < c.PageSize {
		c.PageSize = c.Limit
	}

	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	var nextPageToken []byte
	var opsProcessed int
	for pageIndex := 0; ; pageIndex++ {
		resp, err := cl.WorkflowService().ListNexusOperationExecutions(cctx, &workflowservice.ListNexusOperationExecutionsRequest{
			Namespace:     c.Parent.Parent.Namespace,
			PageSize:      int32(c.PageSize),
			NextPageToken: nextPageToken,
			Query:         c.Query,
		})
		if err != nil {
			return fmt.Errorf("failed listing nexus operations: %w", err)
		}
		var textTable []map[string]any
		for _, op := range resp.GetOperations() {
			if c.Limit > 0 && opsProcessed >= c.Limit {
				break
			}
			opsProcessed++
			if cctx.JSONOutput {
				_ = cctx.Printer.PrintStructured(op, printer.StructuredOptions{})
			} else {
				textTable = append(textTable, map[string]any{
					"Status":      op.Status,
					"OperationId": op.OperationId,
					"Endpoint":    op.Endpoint,
					"Service":     op.Service,
					"Operation":   op.Operation,
					"StartTime":   op.ScheduleTime.AsTime(),
				})
			}
		}
		if len(textTable) > 0 {
			_ = cctx.Printer.PrintStructured(textTable, printer.StructuredOptions{
				Fields: []string{"Status", "OperationId", "Endpoint", "Service", "Operation", "StartTime"},
				Table:  &printer.TableOptions{NoHeader: pageIndex > 0},
			})
		}
		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 || (c.Limit > 0 && opsProcessed >= c.Limit) {
			return nil
		}
	}
}

func (c *TemporalNexusOperationCountCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().CountNexusOperationExecutions(cctx, &workflowservice.CountNexusOperationExecutionsRequest{
		Namespace: c.Parent.Parent.Namespace,
		Query:     c.Query,
	})
	if err != nil {
		return fmt.Errorf("failed counting nexus operations: %w", err)
	}
	groups := make([]countGroup, len(resp.Groups))
	for i, g := range resp.Groups {
		groups[i] = g
	}
	if cctx.JSONOutput {
		stripCountGroupMetadataType(groups)
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	cctx.Printer.Printlnf("Total: %v", resp.Count)
	printCountGroupsText(cctx, groups)
	return nil
}

// nexusStartInput holds the parsed inputs for starting a Nexus operation.
type nexusStartInput struct {
	operation string
	input     any
	options   client.StartNexusOperationOptions
}

func buildNexusStartOptions(s *NexusOperationStartOptions, p *PayloadInputOptions) (client.NexusClientOptions, *nexusStartInput, error) {
	nexusCl := client.NexusClientOptions{
		Endpoint: s.Endpoint,
		Service:  s.Service,
	}

	opts := client.StartNexusOperationOptions{
		ID:                     s.OperationId,
		ScheduleToCloseTimeout: s.ScheduleToCloseTimeout.Duration(),
		ScheduleToStartTimeout: s.ScheduleToStartTimeout.Duration(),
		StartToCloseTimeout:    s.StartToCloseTimeout.Duration(),
		Summary:                s.StaticSummary,
	}

	if len(s.SearchAttribute) > 0 {
		saMap, err := stringKeysJSONValues(s.SearchAttribute, false)
		if err != nil {
			return nexusCl, nil, fmt.Errorf("invalid search attribute values: %w", err)
		}
		if opts.SearchAttributes, err = mapToSearchAttributes(saMap); err != nil {
			return nexusCl, nil, err
		}
	}

	if s.IdConflictPolicy.Value != "" {
		v, err := stringToProtoEnum[enumspb.NexusOperationIdConflictPolicy](
			s.IdConflictPolicy.Value,
			enumspb.NexusOperationIdConflictPolicy_shorthandValue,
			enumspb.NexusOperationIdConflictPolicy_value)
		if err != nil {
			return nexusCl, nil, fmt.Errorf("invalid id-conflict-policy: %w", err)
		}
		opts.IDConflictPolicy = v
	}

	if s.IdReusePolicy.Value != "" {
		v, err := stringToProtoEnum[enumspb.NexusOperationIdReusePolicy](
			s.IdReusePolicy.Value,
			enumspb.NexusOperationIdReusePolicy_shorthandValue,
			enumspb.NexusOperationIdReusePolicy_value)
		if err != nil {
			return nexusCl, nil, fmt.Errorf("invalid id-reuse-policy: %w", err)
		}
		opts.IDReusePolicy = v
	}

	// Build input payload
	var input any
	rawInput, err := p.buildRawInput()
	if err != nil {
		return nexusCl, nil, err
	}
	if len(rawInput) > 1 {
		return nexusCl, nil, fmt.Errorf("nexus operations accept at most one input argument, got %d", len(rawInput))
	}
	if len(rawInput) == 1 {
		input = rawInput[0]
	}

	return nexusCl, &nexusStartInput{
		operation: s.Operation,
		input:     input,
		options:   opts,
	}, nil
}
