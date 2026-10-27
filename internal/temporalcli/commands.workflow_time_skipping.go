package temporalcli

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

func workflowTimeSkippingConfig(duration time.Duration, hasDuration bool, fastForwardID string, disablePropagation bool, maxSkips int) (*client.TimeSkippingConfig, error) {
	if hasDuration && duration <= 0 {
		return nil, errors.New("--fast-forward-duration must be greater than zero")
	}
	if !hasDuration && fastForwardID != "" {
		return nil, errors.New("--fast-forward-id requires --fast-forward-duration")
	}
	if maxSkips < 0 || maxSkips > math.MaxInt32 {
		return nil, fmt.Errorf("--max-skip-per-session must be between 0 and %d", math.MaxInt32)
	}
	config := &client.TimeSkippingConfig{
		Enabled:             true,
		DisablePropagation:  disablePropagation,
		MaxSessionSkipCount: int32(maxSkips),
	}
	if hasDuration {
		if fastForwardID == "" {
			fastForwardID = uuid.NewString()
		}
		config.FastForwardConfig = &client.FastForwardConfig{ID: fastForwardID, Duration: duration}
	}
	return config, nil
}

func (c *TemporalWorkflowTimeSkippingEnableCommand) run(cctx *CommandContext, args []string) error {
	if c.Command.Flags().Changed("max-skip-per-session") && c.MaxSkipPerSession <= 0 {
		return errors.New("--max-skip-per-session must be greater than zero")
	}
	config, err := workflowTimeSkippingConfig(c.FastForwardDuration.Duration(), c.Command.Flags().Changed("fast-forward-duration"), c.FastForwardId, c.DisablePropagation, c.MaxSkipPerSession)
	if err != nil {
		return err
	}
	return patchWorkflowTimeSkipping(cctx, c.Parent.Parent, c.WorkflowId, c.RunId, config)
}

func (c *TemporalWorkflowTimeSkippingDisableCommand) run(cctx *CommandContext, args []string) error {
	return patchWorkflowTimeSkipping(cctx, c.Parent.Parent, c.WorkflowId, c.RunId, &client.TimeSkippingConfig{Enabled: false})
}

func patchWorkflowTimeSkipping(cctx *CommandContext, parent *TemporalWorkflowCommand, workflowID, runID string, config *client.TimeSkippingConfig) error {
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

func workflowTimeSkippingPollResult(resp *workflowservice.PollWorkflowExecutionTimeSkippingResponse) (string, string, error) {
	switch resp.GetFastForwardPollingResult() {
	case enumspb.FAST_FORWARD_POLLING_RESULT_FAST_FORWARD_COMPLETED:
		return "completed", "", nil
	case enumspb.FAST_FORWARD_POLLING_RESULT_POLL_TIMEOUT:
		return "poll timed out", "", nil
	case enumspb.FAST_FORWARD_POLLING_RESULT_FAST_FORWARD_FAILED:
		return "failed", resp.GetFailedReason(), nil
	default:
		return "", "", fmt.Errorf("unknown fast-forward polling result %v", resp.GetFastForwardPollingResult())
	}
}

func (c *TemporalWorkflowTimeSkippingPollCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()
	resp, err := cl.WorkflowService().PollWorkflowExecutionTimeSkipping(cctx, &workflowservice.PollWorkflowExecutionTimeSkippingRequest{
		Namespace: c.Parent.Parent.Namespace,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: c.WorkflowId,
			RunId:      c.RunId,
		},
		FastForwardId: c.FastForwardId,
	})
	if err != nil {
		return err
	}
	status, reason, err := workflowTimeSkippingPollResult(resp)
	if err != nil {
		return err
	}
	if reason == "" {
		cctx.Printer.Println(status)
	} else {
		cctx.Printer.Printlnf("%s: %s", status, reason)
	}
	return nil
}
