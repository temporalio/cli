package temporalcli

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/temporalio/cli/internal/printer"
	activitypb "go.temporal.io/api/activity/v1"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	sdkpb "go.temporal.io/api/sdk/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
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
			Activity: &batch.BatchOperationUpdateActivityOptions_Type{Type: c.ActivityType},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: updatePath,
			},
			RestoreOriginal: c.RestoreOriginalOptions,
		}

		if c.ActivityType != "" {
			updateActivitiesOperation.Activity = &batch.BatchOperationUpdateActivityOptions_Type{Type: c.ActivityType}
		} else if c.MatchAll {
			updateActivitiesOperation.Activity = &batch.BatchOperationUpdateActivityOptions_MatchAll{MatchAll: true}
		} else {
			return fmt.Errorf("either Activity Type must be provided or MatchAll must be set to true")
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
		Identity: c.Parent.Identity,
	}

	if c.ActivityId != "" && c.ActivityType != "" {
		return fmt.Errorf("either Activity Type or Activity Id, but not both")
	} else if c.ActivityType != "" {
		request.Activity = &workflowservice.PauseActivityRequest_Type{Type: c.ActivityType}
	} else if c.ActivityId != "" {
		request.Activity = &workflowservice.PauseActivityRequest_Id{Id: c.ActivityId}
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
		}

		if c.ActivityId != "" && c.ActivityType != "" {
			return fmt.Errorf("either Activity Type or Activity Id, but not both")
		} else if c.ActivityType != "" {
			request.Activity = &workflowservice.UnpauseActivityRequest_Type{Type: c.ActivityType}
		} else if c.ActivityId != "" {
			request.Activity = &workflowservice.UnpauseActivityRequest_Id{Id: c.ActivityId}
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
		}
		if c.ActivityType != "" {
			unpauseActivitiesOperation.Activity = &batch.BatchOperationUnpauseActivities_Type{Type: c.ActivityType}
		} else if c.MatchAll {
			unpauseActivitiesOperation.Activity = &batch.BatchOperationUnpauseActivities_MatchAll{MatchAll: true}
		} else {
			return fmt.Errorf("either Activity Type must be provided or MatchAll must be set to true")
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

func (c *TemporalActivityResetCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

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
		request := &workflowservice.ResetActivityRequest{
			Namespace: c.Parent.Namespace,
			Execution: &common.WorkflowExecution{
				WorkflowId: c.WorkflowId,
				RunId:      c.RunId,
			},
			Identity:       c.Parent.Identity,
			KeepPaused:     c.KeepPaused,
			ResetHeartbeat: c.ResetHeartbeats,
		}

		if c.ActivityId != "" && c.ActivityType != "" {
			return fmt.Errorf("either Activity Type or Activity Id, but not both")
		} else if c.ActivityType != "" {
			request.Activity = &workflowservice.ResetActivityRequest_Type{Type: c.ActivityType}
		} else if c.ActivityId != "" {
			request.Activity = &workflowservice.ResetActivityRequest_Id{Id: c.ActivityId}
		} else {
			return fmt.Errorf("either Activity Type or Activity Id must be provided")
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
		}
		if c.ActivityType != "" {
			resetActivitiesOperation.Activity = &batch.BatchOperationResetActivities_Type{Type: c.ActivityType}
		} else if c.MatchAll {
			resetActivitiesOperation.Activity = &batch.BatchOperationResetActivities_MatchAll{MatchAll: true}
		} else {
			return fmt.Errorf("either Activity Type must be provided or MatchAll must be set to true")
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

func (c *TemporalActivityStartCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	req, err := buildStartActivityRequest(cctx, c.Parent, &c.ActivityStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}
	resp, err := cl.WorkflowService().StartActivityExecution(cctx, req)
	if err != nil {
		return fmt.Errorf("failed starting activity: %w", err)
	}
	return cctx.Printer.PrintStructured(struct {
		ActivityId string `json:"activityId"`
		RunId      string `json:"runId"`
		Started    bool   `json:"started"`
	}{
		ActivityId: c.ActivityId,
		RunId:      resp.RunId,
		Started:    resp.Started,
	}, printer.StructuredOptions{})
}

func (c *TemporalActivityExecuteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	req, err := buildStartActivityRequest(cctx, c.Parent, &c.ActivityStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}
	startResp, err := cl.WorkflowService().StartActivityExecution(cctx, req)
	if err != nil {
		return fmt.Errorf("failed starting activity: %w", err)
	}
	return pollActivityOutcome(cctx, cl.WorkflowService(), c.Parent.Namespace, c.ActivityId, startResp.RunId)
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
		RunId:          c.RunId,
		IncludeInput:   true,
		IncludeOutcome: true,
	})
	if err != nil {
		return fmt.Errorf("failed describing activity: %w", err)
	}
	if c.Raw || cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	return cctx.Printer.PrintStructured(resp.Info, printer.StructuredOptions{})
}

func (c *TemporalActivityListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	var nextPageToken []byte
	var execsProcessed int
	for pageIndex := 0; ; pageIndex++ {
		resp, err := cl.WorkflowService().ListActivityExecutions(cctx, &workflowservice.ListActivityExecutionsRequest{
			Namespace:     c.Parent.Namespace,
			PageSize:      int32(c.PageSize),
			NextPageToken: nextPageToken,
			Query:         c.Query,
		})
		if err != nil {
			return fmt.Errorf("failed listing activities: %w", err)
		}
		var textTable []map[string]any
		for _, exec := range resp.Executions {
			if c.Limit > 0 && execsProcessed >= c.Limit {
				break
			}
			execsProcessed++
			if cctx.JSONOutput {
				_ = cctx.Printer.PrintStructured(exec, printer.StructuredOptions{})
			} else {
				textTable = append(textTable, map[string]any{
					"Status":     exec.Status,
					"ActivityId": exec.ActivityId,
					"Type":       exec.ActivityType.GetName(),
					"StartTime":  exec.ScheduleTime.AsTime(),
				})
			}
		}
		if len(textTable) > 0 {
			_ = cctx.Printer.PrintStructured(textTable, printer.StructuredOptions{
				Fields: []string{"Status", "ActivityId", "Type", "StartTime"},
				Table:  &printer.TableOptions{NoHeader: pageIndex > 0},
			})
		}
		nextPageToken = resp.NextPageToken
		if len(nextPageToken) == 0 || (c.Limit > 0 && execsProcessed >= c.Limit) {
			return nil
		}
	}
}

func (c *TemporalActivityCountCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().CountActivityExecutions(cctx, &workflowservice.CountActivityExecutionsRequest{
		Namespace: c.Parent.Namespace,
		Query:     c.Query,
	})
	if err != nil {
		return fmt.Errorf("failed counting activities: %w", err)
	}
	if cctx.JSONOutput {
		for _, group := range resp.Groups {
			for _, payload := range group.GroupValues {
				delete(payload.GetMetadata(), "type")
			}
		}
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	cctx.Printer.Printlnf("Total: %v", resp.Count)
	for _, group := range resp.Groups {
		var valueStr string
		for _, payload := range group.GroupValues {
			var value any
			if err := converter.GetDefaultDataConverter().FromPayload(payload, &value); err != nil {
				value = fmt.Sprintf("<failed converting: %v>", err)
			}
			if valueStr != "" {
				valueStr += ", "
			}
			valueStr += fmt.Sprintf("%v", value)
		}
		cctx.Printer.Printlnf("Group total: %v, values: %v", group.Count, valueStr)
	}
	return nil
}

func (c *TemporalActivityCancelCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().RequestCancelActivityExecution(cctx, &workflowservice.RequestCancelActivityExecutionRequest{
		Namespace:  c.Parent.Namespace,
		ActivityId: c.ActivityId,
		RunId:      c.RunId,
		Identity:   c.Parent.Identity,
		RequestId:  uuid.New().String(),
		Reason:     c.Reason,
	})
	if err != nil {
		return fmt.Errorf("failed to cancel activity: %w", err)
	}
	cctx.Printer.Println("Cancellation requested")
	return nil
}

func (c *TemporalActivityTerminateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	reason := c.Reason
	if reason == "" {
		reason = defaultReason()
	}
	_, err = cl.WorkflowService().TerminateActivityExecution(cctx, &workflowservice.TerminateActivityExecutionRequest{
		Namespace:  c.Parent.Namespace,
		ActivityId: c.ActivityId,
		RunId:      c.RunId,
		Identity:   c.Parent.Identity,
		RequestId:  uuid.New().String(),
		Reason:     reason,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate activity: %w", err)
	}
	cctx.Printer.Println("Activity terminated")
	return nil
}

func (c *TemporalActivityResultCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	return pollActivityOutcome(cctx, cl.WorkflowService(), c.Parent.Namespace, c.ActivityId, c.RunId)
}

// pollActivityOutcome long-polls for the activity's outcome, re-issuing the
// poll when the server returns an empty response (its signal to keep polling).
// Each iteration uses a per-RPC timeout as a safety net against network hangs,
// consistent with the Go SDK's long-poll pattern.
func pollActivityOutcome(cctx *CommandContext, svc workflowservice.WorkflowServiceClient, ns, activityID, runID string) error {
	for {
		rpcCtx, cancel := context.WithTimeout(cctx, longPollPerRPCTimeout)
		resp, err := svc.PollActivityExecution(rpcCtx, &workflowservice.PollActivityExecutionRequest{
			Namespace:  ns,
			ActivityId: activityID,
			RunId:      runID,
		})
		cancel()
		if err != nil {
			if cctx.Err() != nil {
				return cctx.Err()
			}
			return fmt.Errorf("failed polling activity result: %w", err)
		}
		if resp.Outcome != nil {
			return printActivityOutcome(cctx, resp.Outcome)
		}
	}
}

const longPollPerRPCTimeout = 70 * time.Second

func buildStartActivityRequest(
	cctx *CommandContext,
	parent *TemporalActivityCommand,
	opts *ActivityStartOptions,
	inputOpts *PayloadInputOptions,
) (*workflowservice.StartActivityExecutionRequest, error) {
	input, err := inputOpts.buildRawInputPayloads()
	if err != nil {
		return nil, err
	}

	req := &workflowservice.StartActivityExecutionRequest{
		Namespace:  parent.Namespace,
		Identity:   parent.Identity,
		RequestId:  uuid.New().String(),
		ActivityId: opts.ActivityId,
		ActivityType: &common.ActivityType{
			Name: opts.Type,
		},
		TaskQueue: &taskqueuepb.TaskQueue{
			Name: opts.TaskQueue,
		},
		ScheduleToCloseTimeout: durationpb.New(opts.ScheduleToCloseTimeout.Duration()),
		ScheduleToStartTimeout: durationpb.New(opts.ScheduleToStartTimeout.Duration()),
		StartToCloseTimeout:    durationpb.New(opts.StartToCloseTimeout.Duration()),
		HeartbeatTimeout:       durationpb.New(opts.HeartbeatTimeout.Duration()),
		Input:                  input,
	}

	if opts.RetryInitialInterval.Duration() > 0 || opts.RetryMaximumInterval.Duration() > 0 ||
		opts.RetryBackoffCoefficient > 0 || opts.RetryMaximumAttempts > 0 {
		req.RetryPolicy = &common.RetryPolicy{}
		if opts.RetryInitialInterval.Duration() > 0 {
			req.RetryPolicy.InitialInterval = durationpb.New(opts.RetryInitialInterval.Duration())
		}
		if opts.RetryMaximumInterval.Duration() > 0 {
			req.RetryPolicy.MaximumInterval = durationpb.New(opts.RetryMaximumInterval.Duration())
		}
		if opts.RetryBackoffCoefficient > 0 {
			req.RetryPolicy.BackoffCoefficient = float64(opts.RetryBackoffCoefficient)
		}
		if opts.RetryMaximumAttempts > 0 {
			req.RetryPolicy.MaximumAttempts = int32(opts.RetryMaximumAttempts)
		}
	}

	if opts.IdReusePolicy.Value != "" {
		v, err := stringToProtoEnum[enumspb.ActivityIdReusePolicy](
			opts.IdReusePolicy.Value, enumspb.ActivityIdReusePolicy_shorthandValue, enumspb.ActivityIdReusePolicy_value)
		if err != nil {
			return nil, fmt.Errorf("invalid activity ID reuse policy: %w", err)
		}
		req.IdReusePolicy = v
	}
	if opts.IdConflictPolicy.Value != "" {
		v, err := stringToProtoEnum[enumspb.ActivityIdConflictPolicy](
			opts.IdConflictPolicy.Value, enumspb.ActivityIdConflictPolicy_shorthandValue, enumspb.ActivityIdConflictPolicy_value)
		if err != nil {
			return nil, fmt.Errorf("invalid activity ID conflict policy: %w", err)
		}
		req.IdConflictPolicy = v
	}

	if len(opts.SearchAttribute) > 0 {
		saMap, err := stringKeysJSONValues(opts.SearchAttribute, false)
		if err != nil {
			return nil, fmt.Errorf("invalid search attribute values: %w", err)
		}
		saPayloads, err := encodeMapToPayloads(saMap)
		if err != nil {
			return nil, fmt.Errorf("failed encoding search attributes: %w", err)
		}
		req.SearchAttributes = &common.SearchAttributes{IndexedFields: saPayloads}
	}

	if len(opts.Headers) > 0 {
		headerMap, err := stringKeysJSONValues(opts.Headers, false)
		if err != nil {
			return nil, fmt.Errorf("invalid header values: %w", err)
		}
		headerPayloads, err := encodeMapToPayloads(headerMap)
		if err != nil {
			return nil, fmt.Errorf("failed encoding headers: %w", err)
		}
		req.Header = &common.Header{Fields: headerPayloads}
	}

	if opts.StaticSummary != "" || opts.StaticDetails != "" {
		req.UserMetadata = &sdkpb.UserMetadata{}
		if opts.StaticSummary != "" {
			req.UserMetadata.Summary = &common.Payload{
				Metadata: map[string][]byte{"encoding": []byte("json/plain")},
				Data:     []byte(fmt.Sprintf("%q", opts.StaticSummary)),
			}
		}
		if opts.StaticDetails != "" {
			req.UserMetadata.Details = &common.Payload{
				Metadata: map[string][]byte{"encoding": []byte("json/plain")},
				Data:     []byte(fmt.Sprintf("%q", opts.StaticDetails)),
			}
		}
	}

	if opts.PriorityKey > 0 || opts.FairnessKey != "" || opts.FairnessWeight > 0 {
		req.Priority = &common.Priority{
			PriorityKey:    int32(opts.PriorityKey),
			FairnessKey:    opts.FairnessKey,
			FairnessWeight: float32(opts.FairnessWeight),
		}
	}

	return req, nil
}

func printActivityOutcome(cctx *CommandContext, outcome *activitypb.ActivityExecutionOutcome) error {
	if outcome == nil {
		return fmt.Errorf("activity outcome not available")
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(outcome, printer.StructuredOptions{})
	}
	if result := outcome.GetResult(); result != nil {
		for _, payload := range result.Payloads {
			var value any
			if err := converter.GetDefaultDataConverter().FromPayload(payload, &value); err != nil {
				cctx.Printer.Printlnf("Result: <failed converting: %v>", err)
			} else {
				cctx.Printer.Printlnf("Result: %v", value)
			}
		}
		return nil
	}
	if f := outcome.GetFailure(); f != nil {
		return fmt.Errorf("activity failed: %s", f.GetMessage())
	}
	return fmt.Errorf("activity completed with unknown outcome")
}
