package temporalcli

import (
	"encoding/json"
	"fmt"
  "os/user"
	"github.com/fatih/color"
	"github.com/google/uuid"
  "github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/query/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"	
)

func (c *TemporalWorkflowCancelCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})
	if err != nil {
		return err
	}

	// Run single or batch
	if exec != nil {
		err = cl.CancelWorkflow(cctx, exec.WorkflowId, exec.RunId)
		if err != nil {
			return fmt.Errorf("failed to cancel workflow: %w", err)
		}
		cctx.Printer.Println("Canceled workflow")
	} else if batchReq != nil {
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_CancellationOperation{
			CancellationOperation: &batch.BatchOperationCancellation{
				Identity: clientIdentity(),
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}

func (*TemporalWorkflowDeleteCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (c *TemporalWorkflowQueryCommand) run(cctx *CommandContext, args []string) error {
	return queryHelper(cctx, c.Parent, c.PayloadInputOptions,
		c.Type, c.RejectCondition, c.WorkflowReferenceOptions)
}

func (*TemporalWorkflowResetBatchCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (c *TemporalWorkflowSignalCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Get input payloads
	input, err := c.buildRawInputPayloads()
	if err != nil {
		return err
	}

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})
	if err != nil {
		return err
	}

	// Run single or batch
	if exec != nil {
		// We have to use the raw signal service call here because the Go SDK's
		// signal call doesn't accept multiple arguments.
		_, err = cl.WorkflowService().SignalWorkflowExecution(cctx, &workflowservice.SignalWorkflowExecutionRequest{
			Namespace:         c.Parent.Namespace,
			WorkflowExecution: &common.WorkflowExecution{WorkflowId: c.WorkflowId, RunId: c.RunId},
			SignalName:        c.Name,
			Input:             input,
			Identity:          clientIdentity(),
		})
		if err != nil {
			return fmt.Errorf("failed signalling workflow: %w", err)
		}
		cctx.Printer.Println("Signal workflow succeeded")
	} else if batchReq != nil {
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_SignalOperation{
			SignalOperation: &batch.BatchOperationSignal{
				Signal:   c.Name,
				Input:    input,
				Identity: clientIdentity(),
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}
	return nil
}

func (c *TemporalWorkflowStackCommand) run(cctx *CommandContext, args []string) error {
	return queryHelper(cctx, c.Parent, PayloadInputOptions{},
		"__stack_trace", c.RejectCondition, c.WorkflowReferenceOptions)
}

func (c *TemporalWorkflowTerminateCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// We create a faux SingleWorkflowOrBatchOptions to use the shared logic
	opts := SingleWorkflowOrBatchOptions{
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Query:      c.Query,
		Reason:     c.Reason,
		Yes:        c.Yes,
	}

	exec, batchReq, err := opts.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{
		// You're allowed to specify a reason when terminating a workflow
		AllowReasonWithWorkflowID: true,
	})
	if err != nil {
		return err
	}

	// Run single or batch
	if exec != nil {
		reason := c.Reason
		if reason == "" {
			reason = defaultReason()
		}
		err = cl.TerminateWorkflow(cctx, exec.WorkflowId, exec.RunId, reason)
		if err != nil {
			return fmt.Errorf("failed to terminate workflow: %w", err)
		}
		cctx.Printer.Println("Workflow terminated")
	} else if batchReq != nil {
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_TerminationOperation{
			TerminationOperation: &batch.BatchOperationTermination{
				Identity: clientIdentity(),
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}

func (*TemporalWorkflowTraceCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (c *TemporalWorkflowUpdateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Get raw input
	input, err := c.buildRawInput()
	if err != nil {
		return err
	}

	request := &client.UpdateWorkflowWithOptionsRequest{
		WorkflowID:          c.WorkflowId,
		RunID:               c.RunId,
		UpdateName:          c.Name,
		FirstExecutionRunID: c.FirstExecutionRunId,
		Args:                input,
	}

	updateHandle, err := cl.UpdateWorkflowWithOptions(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to update workflow: %w", err)
	}

	var valuePtr interface{}
	err = updateHandle.Get(cctx, &valuePtr)
	if err != nil {
		return fmt.Errorf("unable to update workflow: %w", err)
	}

	return cctx.Printer.PrintStructured(
		struct {
			Name     string      `json:"name"`
			UpdateID string      `json:"updateId"`
			Result   interface{} `json:"result"`
		}{Name: c.Name, UpdateID: updateHandle.UpdateID(), Result: valuePtr},
		printer.StructuredOptions{})
}

func username() string {
	username := "<unknown-user>"
	if u, err := user.Current(); err != nil && u.Username != "" {
		username = u.Username
	}
	return username
}

func defaultReason() string {
	return "Requested from CLI by " + username()
}

type singleOrBatchOverrides struct {
	AllowReasonWithWorkflowID bool
}

func (s *SingleWorkflowOrBatchOptions) workflowExecOrBatch(
	cctx *CommandContext,
	namespace string,
	cl client.Client,
	overrides singleOrBatchOverrides,
) (*common.WorkflowExecution, *workflowservice.StartBatchOperationRequest, error) {
	// If workflow is set, we return single execution
	if s.WorkflowId != "" {
		if s.Query != "" {
			return nil, nil, fmt.Errorf("cannot set query when workflow ID is set")
		} else if s.Reason != "" && !overrides.AllowReasonWithWorkflowID {
			return nil, nil, fmt.Errorf("cannot set reason when workflow ID is set")
		} else if s.Yes {
			return nil, nil, fmt.Errorf("cannot set 'yes' when workflow ID is set")
		}
		return &common.WorkflowExecution{WorkflowId: s.WorkflowId, RunId: s.RunId}, nil, nil
	}

	// Check query is set properly
	if s.Query == "" {
		return nil, nil, fmt.Errorf("must set either workflow ID or query")
	} else if s.WorkflowId != "" {
		return nil, nil, fmt.Errorf("cannot set workflow ID when query is set")
	} else if s.RunId != "" {
		return nil, nil, fmt.Errorf("cannot set run ID when query is set")
	}

	// Count the workflows that will be affected
	count, err := cl.CountWorkflow(cctx, &workflowservice.CountWorkflowExecutionsRequest{Query: s.Query})
	if err != nil {
		return nil, nil, fmt.Errorf("failed counting workflows from query: %w", err)
	}
	yes, err := cctx.promptYes(
		fmt.Sprintf("Start batch against approximately %v workflow(s)? y/N", count.Count), s.Yes)
	if err != nil {
		return nil, nil, err
	} else if !yes {
		// We consider this a command failure
		return nil, nil, fmt.Errorf("user denied confirmation")
	}

	// Default the reason if not set
	reason := s.Reason
	if reason == "" {
		reason = defaultReason()
	}

	return nil, &workflowservice.StartBatchOperationRequest{
		Namespace:       namespace,
		JobId:           uuid.NewString(),
		VisibilityQuery: s.Query,
		Reason:          reason,
	}, nil
}

func startBatchJob(cctx *CommandContext, cl client.Client, req *workflowservice.StartBatchOperationRequest) error {
	_, err := cl.WorkflowService().StartBatchOperation(cctx, req)
	if err != nil {
		return fmt.Errorf("failed starting batch operation: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			struct {
				BatchJobID string `json:"batchJobId"`
			}{BatchJobID: req.JobId},
			printer.StructuredOptions{})
	}
	cctx.Printer.Printlnf("Started batch for job ID: %v", req.JobId)
	return nil
}

func queryHelper(cctx *CommandContext,
	parent *TemporalWorkflowCommand,
	inputOpts PayloadInputOptions,
	queryType string,
	rejectCondition StringEnum,
	execution WorkflowReferenceOptions,
) error {
	cl, err := parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Get input payloads
	input, err := inputOpts.buildRawInputPayloads()
	if err != nil {
		return err
	}

	queryRejectCond := enums.QUERY_REJECT_CONDITION_UNSPECIFIED
	switch rejectCondition.Value {
	case "":
	case "not_open":
		queryRejectCond = enums.QUERY_REJECT_CONDITION_NOT_OPEN
	case "not_completed_cleanly":
		queryRejectCond = enums.QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
	default:
		return fmt.Errorf("invalid query reject condition: %v, valid values are: 'not_open', 'not_completed_cleanly'", rejectCondition)
	}

	result, err := cl.WorkflowService().QueryWorkflow(cctx, &workflowservice.QueryWorkflowRequest{
		Namespace: parent.Namespace,
		Execution: &common.WorkflowExecution{WorkflowId: execution.WorkflowId, RunId: execution.RunId},
		Query: &query.WorkflowQuery{
			QueryType: queryType,
			QueryArgs: input,
		},
		QueryRejectCondition: queryRejectCond,
	})

	if err != nil {
		return fmt.Errorf("querying workflow failed: %w", err)
	}

	if result.QueryRejected != nil {
		return fmt.Errorf("query was rejected, workflow has status: %v", result.QueryRejected.GetStatus())
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Query result:"))
	output := struct {
		QueryResult json.RawMessage `cli:",cardOmitEmpty"`
	}{}
	output.QueryResult, err = cctx.MarshalFriendlyJSONPayloads(result.QueryResult)
	if err != nil {
		return fmt.Errorf("failed to marshal query result: %w", err)
	}

	return cctx.Printer.PrintStructured(output, printer.StructuredOptions{})
}
