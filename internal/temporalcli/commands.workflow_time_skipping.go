package temporalcli

import (
	"context"
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
		return nil
	}
	cctx.Printer.Printlnf("Workflow time skipping updated: %s", workflowID)
	if config.Enabled {
		cctx.Printer.Println("Time skipping: enabled")
	} else {
		cctx.Printer.Println("Time skipping: disabled")
	}
	cctx.Printer.Println("Check status:")
	cmd := fmt.Sprintf("  temporal workflow describe --namespace %s --workflow-id %s", parent.Namespace, workflowID)
	if runID != "" {
		cmd += " --run-id " + runID
	}
	cctx.Printer.Println(cmd)
	if ff := config.FastForwardConfig; ff != nil {
		cctx.Printer.Printlnf("Fast-forward ID: %s", ff.ID)
		cctx.Printer.Println("Poll fast-forward:")
		cmd := fmt.Sprintf("  temporal workflow time-skipping poll --namespace %s --workflow-id %s", parent.Namespace, workflowID)
		if runID != "" {
			cmd += " --run-id " + runID
		}
		cctx.Printer.Println(cmd + " --fast-forward-id " + ff.ID)
	}
	return nil
}

func (c *TemporalWorkflowTimeSkippingPollCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	// The SDK reissues the server long poll until our context ends, so a
	// server-side poll window expiring is handled for us rather than being an
	// outcome the user has to re-run on.
	res, err := timeskipping.PollFastForward(cctx, cl, timeskipping.PollFastForwardOptions{
		WorkflowID:    c.WorkflowId,
		RunID:         c.RunId,
		FastForwardID: c.FastForwardId,
	})
	if err != nil {
		return err
	}
	switch {
	case res.Completed:
		cctx.Printer.Println("completed")
	case errors.Is(cctx.Err(), context.DeadlineExceeded):
		cctx.Printer.Println("poll timed out; try again")
	default:
		cctx.Printer.Printlnf("failed: %s", res.FailedReason)
	}
	return nil
}
