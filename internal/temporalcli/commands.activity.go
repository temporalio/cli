package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/internal/printer"
	activitypb "go.temporal.io/api/activity/v1"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type (
	updateOptionsDescribe struct {
		TaskQueue string

		ScheduleToCloseTimeout time.Duration
		ScheduleToStartTimeout time.Duration
		StartToCloseTimeout    time.Duration
		HeartbeatTimeout       time.Duration

		InitialInterval    time.Duration
		BackoffCoefficient float64
		MaximumInterval    time.Duration
		MaximumAttempts    int32
	}
)

func (c *TemporalActivityCompleteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	metadata := map[string][][]byte{"encoding": {[]byte("json/plain")}}
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
		Identity:   c.Parent.Identity,
	})
	if err != nil {
		return fmt.Errorf("unable to complete Activity: %w", err)
	}
	return nil
}

func (c *TemporalActivityFailCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	var detailPayloads *common.Payloads
	if len(c.Detail) > 0 {
		metadata := map[string][][]byte{"encoding": {[]byte("json/plain")}}
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
		Identity: c.Parent.Identity,
	})
	if err != nil {
		return fmt.Errorf("unable to fail Activity: %w", err)
	}
	return nil
}

func (c *TemporalActivityUpdateOptionsCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	updatePath := []string{}
	activityOptions := &activitypb.ActivityOptions{}

	if c.Command.Flags().Changed("task-queue") {
		activityOptions.TaskQueue = &taskqueuepb.TaskQueue{Name: c.TaskQueue}
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

	// workflowExecOrBatch is defined on SingleWorkflowOrBatchOptions; bridge via
	// manual copy from the embedded SingleActivityOrBatchOptions fields.
	opts := SingleWorkflowOrBatchOptions{
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Query:      c.Query,
		Reason:     c.Reason,
		Yes:        c.Yes,
		Rps:        c.Rps,
	}

	exec, batchReq, err := opts.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})
	if err != nil {
		return err
	}

	if exec != nil {
		if c.ActivityId == "" {
			return fmt.Errorf("either --activity-id and --workflow-id, or --query must be set")
		}
		result, err := cl.WorkflowService().UpdateActivityOptions(cctx, &workflowservice.UpdateActivityOptionsRequest{
			Namespace: c.Parent.Namespace,
			Execution: &common.WorkflowExecution{
				WorkflowId: c.WorkflowId,
				RunId:      c.RunId,
			},
			Activity:        &workflowservice.UpdateActivityOptionsRequest_Id{Id: c.ActivityId},
			ActivityOptions: activityOptions,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: updatePath,
			},
			Identity: c.Parent.Identity,
		})
		if err != nil {
			return fmt.Errorf("unable to update Activity options: %w", err)
		}

		updatedOptions := updateOptionsDescribe{
			TaskQueue: result.GetActivityOptions().TaskQueue.GetName(),

			ScheduleToCloseTimeout: result.GetActivityOptions().ScheduleToCloseTimeout.AsDuration(),
			ScheduleToStartTimeout: result.GetActivityOptions().ScheduleToStartTimeout.AsDuration(),
			StartToCloseTimeout:    result.GetActivityOptions().StartToCloseTimeout.AsDuration(),
			HeartbeatTimeout:       result.GetActivityOptions().HeartbeatTimeout.AsDuration(),

			InitialInterval:    result.GetActivityOptions().RetryPolicy.InitialInterval.AsDuration(),
			BackoffCoefficient: result.GetActivityOptions().RetryPolicy.BackoffCoefficient,
			MaximumInterval:    result.GetActivityOptions().RetryPolicy.MaximumInterval.AsDuration(),
			MaximumAttempts:    result.GetActivityOptions().RetryPolicy.MaximumAttempts,
		}

		_ = cctx.Printer.PrintStructured(updatedOptions, printer.StructuredOptions{})
	} else {
		updateActivitiesOperation := &batch.BatchOperationUpdateActivityOptions{
			Identity: c.Parent.Identity,
			Activity: &batch.BatchOperationUpdateActivityOptions_MatchAll{MatchAll: true},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: updatePath,
			},
			RestoreOriginal: c.RestoreOriginalOptions,
		}

		batchReq.Operation = &workflowservice.StartBatchOperationRequest_UpdateActivityOptionsOperation{
			UpdateActivityOptionsOperation: updateActivitiesOperation,
		}

		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}
	return nil
}

func (c *TemporalActivityPauseCommand) run(cctx *CommandContext, args []string) error {
	if c.ActivityId == "" {
		return fmt.Errorf("Activity Id must be specified")
	}

	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	request := &workflowservice.PauseActivityRequest{
		Namespace: c.Parent.Namespace,
		Execution: &common.WorkflowExecution{
			WorkflowId: c.WorkflowId,
			RunId:      c.RunId,
		},
		Identity: c.Identity,
		Reason:   c.Reason,
		Activity: &workflowservice.PauseActivityRequest_Id{Id: c.ActivityId},
	}
	if request.Identity == "" {
		request.Identity = c.Parent.Identity
	}

	_, err = cl.WorkflowService().PauseActivity(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to pause Activity: %w", err)
	}

	return nil
}

func (c *TemporalActivityUnpauseCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	// workflowExecOrBatch is defined on SingleWorkflowOrBatchOptions; bridge via
	// manual copy from the embedded SingleActivityOrBatchOptions fields.
	opts := SingleWorkflowOrBatchOptions{
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Query:      c.Query,
		Reason:     c.Reason,
		Yes:        c.Yes,
		Rps:        c.Rps,
	}

	exec, batchReq, err := opts.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})
	if err != nil {
		return err
	}

	if exec != nil { // single workflow operation
		if c.ActivityId == "" {
			return fmt.Errorf("either --activity-id and --workflow-id, or --query must be set")
		}

		request := &workflowservice.UnpauseActivityRequest{
			Namespace: c.Parent.Namespace,
			Execution: &common.WorkflowExecution{
				WorkflowId: c.WorkflowId,
				RunId:      c.RunId,
			},
			ResetAttempts:  c.ResetAttempts,
			ResetHeartbeat: c.ResetHeartbeats,
			Jitter:         durationpb.New(c.Jitter.Duration()),
			Identity:       c.Parent.Identity,
			Activity:       &workflowservice.UnpauseActivityRequest_Id{Id: c.ActivityId},
		}

		_, err = cl.WorkflowService().UnpauseActivity(cctx, request)
		if err != nil {
			return fmt.Errorf("unable to unpause an Activity: %w", err)
		}
	} else { // batch operation
		unpauseActivitiesOperation := &batch.BatchOperationUnpauseActivities{
			Identity:       c.Parent.Identity,
			ResetAttempts:  c.ResetAttempts,
			ResetHeartbeat: c.ResetHeartbeats,
			Jitter:         durationpb.New(c.Jitter.Duration()),
			Activity:       &batch.BatchOperationUnpauseActivities_MatchAll{MatchAll: true},
		}

		batchReq.Operation = &workflowservice.StartBatchOperationRequest_UnpauseActivitiesOperation{
			UnpauseActivitiesOperation: unpauseActivitiesOperation,
		}

		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}

func (c *TemporalActivityDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().DescribeActivityExecution(cctx, &workflowservice.DescribeActivityExecutionRequest{
		Namespace:      c.Parent.Namespace,
		ActivityId:     c.ActivityId,
		RunId:          c.ActivityRunId,
		IncludeInput:   c.IncludeInput,
		IncludeOutcome: c.IncludeOutcome,
	})
	if err != nil {
		return fmt.Errorf("unable to describe Activity: %w", err)
	}

	// JSON output: emit the raw proto response
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	// Text output: card-style sections
	info := resp.Info
	cctx.Printer.Println(color.MagentaString("Activity Info:"))
	_ = cctx.Printer.PrintStructured(struct {
		ActivityId             string
		RunId                  string
		Type                   string
		Status                 string
		RunState               string `cli:",cardOmitEmpty"`
		TaskQueue              string `cli:",cardOmitEmpty"`
		Attempt                int32
		ScheduleTime           time.Time     `cli:",cardOmitEmpty"`
		LastStartedTime        time.Time     `cli:",cardOmitEmpty"`
		LastHeartbeatTime      time.Time     `cli:",cardOmitEmpty"`
		CloseTime              time.Time     `cli:",cardOmitEmpty"`
		ExpirationTime         time.Time     `cli:",cardOmitEmpty"`
		ScheduleToCloseTimeout time.Duration `cli:",cardOmitEmpty"`
		StartToCloseTimeout    time.Duration `cli:",cardOmitEmpty"`
		HeartbeatTimeout       time.Duration `cli:",cardOmitEmpty"`
		LastWorkerIdentity     string        `cli:",cardOmitEmpty"`
		CanceledReason         string        `cli:",cardOmitEmpty"`
		StateTransitionCount   int64         `cli:",cardOmitEmpty"`
	}{
		ActivityId:             info.GetActivityId(),
		RunId:                  resp.GetRunId(),
		Type:                   info.GetActivityType().GetName(),
		Status:                 info.GetStatus().String(),
		RunState:               info.GetRunState().String(),
		TaskQueue:              info.GetTaskQueue(),
		Attempt:                info.GetAttempt(),
		ScheduleTime:           timestampToTime(info.GetScheduleTime()),
		LastStartedTime:        timestampToTime(info.GetLastStartedTime()),
		LastHeartbeatTime:      timestampToTime(info.GetLastHeartbeatTime()),
		CloseTime:              timestampToTime(info.GetCloseTime()),
		ExpirationTime:         timestampToTime(info.GetExpirationTime()),
		ScheduleToCloseTimeout: info.GetScheduleToCloseTimeout().AsDuration(),
		StartToCloseTimeout:    info.GetStartToCloseTimeout().AsDuration(),
		HeartbeatTimeout:       info.GetHeartbeatTimeout().AsDuration(),
		LastWorkerIdentity:     info.GetLastWorkerIdentity(),
		CanceledReason:         info.GetCanceledReason(),
		StateTransitionCount:   info.GetStateTransitionCount(),
	}, printer.StructuredOptions{})

	if resp.Input != nil {
		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Input:"))
		_ = cctx.Printer.PrintStructured(resp.Input, printer.StructuredOptions{})
	}

	if resp.Outcome != nil {
		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Outcome:"))
		_ = cctx.Printer.PrintStructured(resp.Outcome, printer.StructuredOptions{})
	}

	return nil
}

func (c *TemporalActivityResetCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	// workflowExecOrBatch is defined on SingleWorkflowOrBatchOptions; bridge via
	// manual copy from the embedded SingleActivityOrBatchOptions fields.
	opts := SingleWorkflowOrBatchOptions{
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Query:      c.Query,
		Reason:     c.Reason,
		Yes:        c.Yes,
		Rps:        c.Rps,
	}

	exec, batchReq, err := opts.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})
	if err != nil {
		return err
	}

	if exec != nil { // single workflow operation
		if c.ActivityId == "" {
			return fmt.Errorf("either --activity-id and --workflow-id, or --query must be set")
		}

		request := &workflowservice.ResetActivityRequest{
			Activity:  &workflowservice.ResetActivityRequest_Id{Id: c.ActivityId},
			Namespace: c.Parent.Namespace,
			Execution: &common.WorkflowExecution{
				WorkflowId: c.WorkflowId,
				RunId:      c.RunId,
			},
			Identity:       c.Parent.Identity,
			KeepPaused:     c.KeepPaused,
			ResetHeartbeat: c.ResetHeartbeats,
		}

		resp, err := cl.WorkflowService().ResetActivity(cctx, request)
		if err != nil {
			return fmt.Errorf("unable to reset an Activity: %w", err)
		}

		resetResponse := struct {
			KeepPaused      bool `json:"keepPaused"`
			ResetHeartbeats bool `json:"resetHeartbeats"`
			ServerResponse  bool `json:"-"`
		}{
			ServerResponse:  resp != nil,
			KeepPaused:      c.KeepPaused,
			ResetHeartbeats: c.ResetHeartbeats,
		}

		_ = cctx.Printer.PrintStructured(resetResponse, printer.StructuredOptions{})
	} else { // batch operation
		resetActivitiesOperation := &batch.BatchOperationResetActivities{
			Identity:               c.Parent.Identity,
			ResetAttempts:          c.ResetAttempts,
			ResetHeartbeat:         c.ResetHeartbeats,
			KeepPaused:             c.KeepPaused,
			Jitter:                 durationpb.New(c.Jitter.Duration()),
			RestoreOriginalOptions: c.RestoreOriginalOptions,
			Activity:               &batch.BatchOperationResetActivities_MatchAll{MatchAll: true},
		}

		batchReq.Operation = &workflowservice.StartBatchOperationRequest_ResetActivitiesOperation{
			ResetActivitiesOperation: resetActivitiesOperation,
		}

		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}
