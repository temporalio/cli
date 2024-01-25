package temporalcli

import (
	"fmt"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (*TemporalWorkflowCancelCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (*TemporalWorkflowDeleteCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (*TemporalWorkflowQueryCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (*TemporalWorkflowResetCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
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

	// Send signal. We have to use the raw signal service call here because the Go
	// SDK's signal call doesn't accept multiple arguments.
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
	return nil
}

func (*TemporalWorkflowStackCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (*TemporalWorkflowTerminateCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (*TemporalWorkflowTraceCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}

func (*TemporalWorkflowUpdateCommand) run(*CommandContext, []string) error {
	return fmt.Errorf("TODO")
}
