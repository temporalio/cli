package temporalcli

import (
	"fmt"

	activitypb "go.temporal.io/api/activity/v1"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
	taskqueuepn "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func (c *TemporalActivityCompleteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	metadata := map[string][]byte{"encoding": []byte("json/plain")}
	resultPayloads, err := CreatePayloads([][]byte{[]byte(c.Result)}, metadata, false)
	if err != nil {
		return err
	}

	_, err = cl.WorkflowService().RespondActivityTaskCompletedById(cctx, &workflowservice.RespondActivityTaskCompletedByIdRequest{
		Namespace:  c.Parent.Namespace,
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		ActivityId: c.ActivityId,
		Result:     resultPayloads,
		Identity:   c.Identity,
	})
	if err != nil {
		return fmt.Errorf("unable to complete Activity: %w", err)
	}
	return nil
}

func (c *TemporalActivityFailCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	var detailPayloads *common.Payloads
	if len(c.Detail) > 0 {
		metadata := map[string][]byte{"encoding": []byte("json/plain")}
		detailPayloads, err = CreatePayloads([][]byte{[]byte(c.Detail)}, metadata, false)
		if err != nil {
			return err
		}
	}
	_, err = cl.WorkflowService().RespondActivityTaskFailedById(cctx, &workflowservice.RespondActivityTaskFailedByIdRequest{
		Namespace:  c.Parent.Namespace,
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		ActivityId: c.ActivityId,
		Failure: &failure.Failure{
			Message: c.Reason,
			Source:  "CLI",
			FailureInfo: &failure.Failure_ApplicationFailureInfo{ApplicationFailureInfo: &failure.ApplicationFailureInfo{
				NonRetryable: true,
				Details:      detailPayloads,
			}},
		},
		Identity: c.Identity,
	})
	if err != nil {
		return fmt.Errorf("unable to fail Activity: %w", err)
	}
	return nil
}

func (c *TemporalActivityUpdateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	updatePath := []string{}
	activityOptions := &activitypb.ActivityOptions{}

	if c.Command.Flags().Changed("task-queue") {
		activityOptions.TaskQueue = &taskqueuepn.TaskQueue{Name: c.TaskQueue}
		updatePath = append(updatePath, "task_queue_name")
	}

	if c.Command.Flags().Changed("schedule-to-close-timeout") {
		activityOptions.ScheduleToCloseTimeout = durationpb.New(c.ScheduleToCloseTimeout.Duration())
		updatePath = append(updatePath, "schedule_to_close_timeout")
	}

	if c.Command.Flags().Changed("schedule-to-start-timeout") {
		activityOptions.ScheduleToStartTimeout = durationpb.New(c.ScheduleToStartTimeout.Duration())
		updatePath = append(updatePath, "schedule_to_start_timeout")
	}

	if c.Command.Flags().Changed("start-to-close-timeout") {
		activityOptions.StartToCloseTimeout = durationpb.New(c.StartToCloseTimeout.Duration())
		updatePath = append(updatePath, "start_to_close_timeout")
	}

	if c.Command.Flags().Changed("heartbeat-timeout") {
		activityOptions.HeartbeatTimeout = durationpb.New(c.HeartbeatTimeout.Duration())
		updatePath = append(updatePath, "heartbeat_timeout")
	}

	if c.Command.Flags().Changed("retry-initial-interval") ||
		c.Command.Flags().Changed("retry-maximum-interval") ||
		c.Command.Flags().Changed("retry-backoff-coefficient") ||
		c.Command.Flags().Changed("retry-maximum-attempts") {
		activityOptions.RetryPolicy = &common.RetryPolicy{}
	}

	if c.Command.Flags().Changed("retry-initial-interval") {
		activityOptions.RetryPolicy.InitialInterval = durationpb.New(c.RetryInitialInterval.Duration())
		updatePath = append(updatePath, "retry_policy.initial_interval")
	}

	if c.Command.Flags().Changed("retry-maximum-interval") {
		activityOptions.RetryPolicy.MaximumInterval = durationpb.New(c.RetryMaximumInterval.Duration())
		updatePath = append(updatePath, "retry_policy.maximum_interval")
	}

	if c.Command.Flags().Changed("retry-backoff-coefficient") {
		activityOptions.RetryPolicy.BackoffCoefficient = float64(c.RetryBackoffCoefficient)
		updatePath = append(updatePath, "retry_policy.backoff_coefficient")
	}

	if c.Command.Flags().Changed("retry-maximum-attempts") {
		activityOptions.RetryPolicy.MaximumAttempts = int32(c.RetryMaximumAttempts)
		updatePath = append(updatePath, "retry_policy.maximum_attempts")
	}

	_, err = cl.WorkflowService().UpdateActivityOptionsById(cctx, &workflowservice.UpdateActivityOptionsByIdRequest{
		Namespace:       c.Parent.Namespace,
		WorkflowId:      c.WorkflowId,
		RunId:           c.RunId,
		ActivityId:      c.ActivityId,
		ActivityOptions: activityOptions,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: updatePath,
		},

		Identity: c.Identity,
	})
	if err != nil {
		return fmt.Errorf("unable to update Activity options: %w", err)
	}
	return nil
}
