package temporalcli

import (
	"go.temporal.io/api/workflowservice/v1"
)

func (c *TemporalWorkflowPauseCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().PauseWorkflowExecution(cctx, &workflowservice.PauseWorkflowExecutionRequest{
		Namespace:  c.Parent.Namespace,
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Identity:   c.Parent.Identity,
		Reason:     c.Reason,
	})
	if err != nil {
		return err
	}

	cctx.Printer.Println("Workflow Execution paused")
	return nil
}

func (c *TemporalWorkflowUnpauseCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().UnpauseWorkflowExecution(cctx, &workflowservice.UnpauseWorkflowExecutionRequest{
		Namespace:  c.Parent.Namespace,
		Reason:     c.Reason,
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Identity:   c.Parent.Identity,
	})
	if err != nil {
		return err
	}

	return nil
}
