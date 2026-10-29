package temporalcli

import (
	"errors"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/timeskipping"
)

func (c *TemporalWorkflowTimeSkippingEnableCommand) run(cctx *CommandContext, args []string) error {
	if c.Command.Flags().Changed("max-skip-per-session") && c.MaxSkipPerSession <= 0 {
		return errors.New("--max-skip-per-session must be greater than zero")
	}
	config, err := timeSkippingConfig(c.FastForwardDuration.Duration(), c.Command.Flags().Changed("fast-forward-duration"), c.FastForwardId, c.DisablePropagation, c.MaxSkipPerSession)
	if err != nil {
		return err
	}
	return updateWorkflowOptionsForTimeSkipping(cctx, c.Parent.Parent, c.WorkflowId, c.RunId, config)
}

func (c *TemporalWorkflowTimeSkippingDisableCommand) run(cctx *CommandContext, args []string) error {
	return updateWorkflowOptionsForTimeSkipping(cctx, c.Parent.Parent, c.WorkflowId, c.RunId, &client.TimeSkippingConfig{Enabled: false})
}

func updateWorkflowOptionsForTimeSkipping(cctx *CommandContext, parent *TemporalWorkflowCommand, workflowID, runID string, config *client.TimeSkippingConfig) error {
	if config == nil {
		return errors.New("time-skipping configuration is required")
	}
	cl, err := dialClient(cctx, &parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	_, err = cl.UpdateWorkflowExecutionOptions(cctx, client.UpdateWorkflowExecutionOptionsRequest{
		WorkflowId: workflowID,
		RunId:      runID,
		WorkflowExecutionOptionsChanges: client.WorkflowExecutionOptionsChanges{
			TimeSkippingConfig: &client.TimeSkippingConfigChange{Value: config},
		},
	})
	if err != nil {
		return err
	}
	if cctx.JSONOutput {
		return printTimeSkippingUpdateJSON(cctx, config)
	}
	cctx.Printer.Printlnf("Workflow time skipping updated: %s", workflowID)
	if config.Enabled {
		cctx.Printer.Println("Time skipping: enabled")
	} else {
		cctx.Printer.Println("Time skipping: disabled")
	}
	if ff := config.FastForwardConfig; ff != nil {
		cctx.Printer.Printlnf("Fast-forward ID: %s", ff.ID)
	}
	return nil
}

func (c *TemporalWorkflowTimeSkippingPollCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	// The SDK retries server poll timeouts until the command context ends.
	res, err := timeskipping.PollFastForward(cctx, cl, timeskipping.PollFastForwardOptions{
		WorkflowID:    c.WorkflowId,
		RunID:         c.RunId,
		FastForwardID: c.FastForwardId,
	})
	if err != nil {
		return err
	}
	if !res.Completed {
		return fmt.Errorf("fast-forward failed: %s", res.FailedReason)
	}
	return printTimeSkippingPollCompleted(cctx)
}
