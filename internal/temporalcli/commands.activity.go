package temporalcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/internal/printer"
	activitypb "go.temporal.io/api/activity/v1"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
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

func (c *TemporalActivityStartCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	handle, err := startActivity(cctx, cl, &c.ActivityStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}
	return printActivityExecution(cctx, c.ActivityId, handle.GetRunID(), c.Type, c.Parent.Namespace, c.TaskQueue)
}

func (c *TemporalActivityExecuteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	handle, err := startActivity(cctx, cl, &c.ActivityStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}
	if !cctx.JSONOutput {
		if err := printActivityExecution(cctx, c.ActivityId, handle.GetRunID(), c.Type, c.Parent.Namespace, c.TaskQueue); err != nil {
			return err
		}
	}
	return getActivityResult(cctx, cl, c.Parent.Namespace, c.ActivityId, handle.GetRunID())
}

func (c *TemporalActivityResultCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	return getActivityResult(cctx, cl, c.Parent.Namespace, c.ActivityId, c.RunId)
}

func startActivity(
	cctx *CommandContext,
	cl client.Client,
	opts *ActivityStartOptions,
	inputOpts *PayloadInputOptions,
) (client.ActivityHandle, error) {
	startOpts, err := buildStartActivityOptions(opts)
	if err != nil {
		return nil, err
	}
	input, err := inputOpts.buildRawInput()
	if err != nil {
		return nil, err
	}
	cctx.Context, err = contextWithHeaders(cctx.Context, opts.Headers)
	if err != nil {
		return nil, err
	}
	handle, err := cl.ExecuteActivity(cctx, startOpts, opts.Type, input...)
	if err != nil {
		return nil, fmt.Errorf("failed starting activity: %w", err)
	}
	return handle, nil
}

func printActivityExecution(cctx *CommandContext, activityID, runID, activityType, namespace, taskQueue string) error {
	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString("Running execution:"))
	}
	return cctx.Printer.PrintStructured(struct {
		ActivityId string `json:"activityId"`
		RunId      string `json:"runId"`
		Type       string `json:"type"`
		Namespace  string `json:"namespace"`
		TaskQueue  string `json:"taskQueue"`
	}{
		ActivityId: activityID,
		RunId:      runID,
		Type:       activityType,
		Namespace:  namespace,
		TaskQueue:  taskQueue,
	}, printer.StructuredOptions{})
}

func buildStartActivityOptions(opts *ActivityStartOptions) (client.StartActivityOptions, error) {
	o := client.StartActivityOptions{
		ID:                     opts.ActivityId,
		TaskQueue:              opts.TaskQueue,
		ScheduleToCloseTimeout: opts.ScheduleToCloseTimeout.Duration(),
		ScheduleToStartTimeout: opts.ScheduleToStartTimeout.Duration(),
		StartToCloseTimeout:    opts.StartToCloseTimeout.Duration(),
		HeartbeatTimeout:       opts.HeartbeatTimeout.Duration(),
		Summary:                opts.StaticSummary,
		Details:                opts.StaticDetails,
		Priority: temporal.Priority{
			PriorityKey:    opts.PriorityKey,
			FairnessKey:    opts.FairnessKey,
			FairnessWeight: opts.FairnessWeight,
		},
	}
	if opts.RetryInitialInterval.Duration() > 0 || opts.RetryMaximumInterval.Duration() > 0 ||
		opts.RetryBackoffCoefficient > 0 || opts.RetryMaximumAttempts > 0 {
		o.RetryPolicy = &temporal.RetryPolicy{
			InitialInterval:    opts.RetryInitialInterval.Duration(),
			MaximumInterval:    opts.RetryMaximumInterval.Duration(),
			BackoffCoefficient: float64(opts.RetryBackoffCoefficient),
			MaximumAttempts:    int32(opts.RetryMaximumAttempts),
		}
	}
	if opts.IdReusePolicy.Value != "" {
		var err error
		o.ActivityIDReusePolicy, err = stringToProtoEnum[enumspb.ActivityIdReusePolicy](
			opts.IdReusePolicy.Value, enumspb.ActivityIdReusePolicy_shorthandValue, enumspb.ActivityIdReusePolicy_value)
		if err != nil {
			return o, fmt.Errorf("invalid activity ID reuse policy: %w", err)
		}
	}
	if opts.IdConflictPolicy.Value != "" {
		var err error
		o.ActivityIDConflictPolicy, err = stringToProtoEnum[enumspb.ActivityIdConflictPolicy](
			opts.IdConflictPolicy.Value, enumspb.ActivityIdConflictPolicy_shorthandValue, enumspb.ActivityIdConflictPolicy_value)
		if err != nil {
			return o, fmt.Errorf("invalid activity ID conflict policy: %w", err)
		}
	}
	if len(opts.SearchAttribute) > 0 {
		saMap, err := stringKeysJSONValues(opts.SearchAttribute, false)
		if err != nil {
			return o, fmt.Errorf("invalid search attribute values: %w", err)
		}
		if o.TypedSearchAttributes, err = mapToSearchAttributes(saMap); err != nil {
			return o, err
		}
	}
	return o, nil
}

func mapToSearchAttributes(m map[string]any) (temporal.SearchAttributes, error) {
	updates := make([]temporal.SearchAttributeUpdate, 0, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			updates = append(updates, temporal.NewSearchAttributeKeyKeyword(k).ValueSet(val))
		case float64:
			updates = append(updates, temporal.NewSearchAttributeKeyFloat64(k).ValueSet(val))
		case bool:
			updates = append(updates, temporal.NewSearchAttributeKeyBool(k).ValueSet(val))
		case []any:
			strs := make([]string, len(val))
			for i, s := range val {
				strs[i] = fmt.Sprint(s)
			}
			updates = append(updates, temporal.NewSearchAttributeKeyKeywordList(k).ValueSet(strs))
		default:
			return temporal.SearchAttributes{}, fmt.Errorf("unsupported search attribute type for key %q: %T", k, v)
		}
	}
	return temporal.NewSearchAttributes(updates...), nil
}

func getActivityResult(cctx *CommandContext, cl client.Client, namespace, activityID, runID string) error {
	outcome, err := pollActivityOutcome(cctx, cl, namespace, activityID, runID)
	if err != nil {
		var notFound *serviceerror.NotFound
		if errors.As(err, &notFound) {
			return fmt.Errorf("activity not found: %s", activityID)
		}
		return fmt.Errorf("failed polling activity result: %w", err)
	}

	resolvedRunID := runID
	if resolvedRunID == "" {
		handle := cl.GetActivityHandle(client.GetActivityHandleOptions{ActivityID: activityID})
		if desc, descErr := handle.Describe(cctx, client.DescribeActivityOptions{}); descErr == nil {
			resolvedRunID = desc.RawExecutionInfo.GetRunId()
		}
	}

	switch v := outcome.GetValue().(type) {
	case *activitypb.ActivityExecutionOutcome_Result:
		return printActivityResult(cctx, activityID, resolvedRunID, v.Result)
	case *activitypb.ActivityExecutionOutcome_Failure:
		return printActivityFailure(cctx, activityID, resolvedRunID, v.Failure)
	default:
		return fmt.Errorf("unexpected activity outcome type: %T", v)
	}
}

func pollActivityOutcome(cctx *CommandContext, cl client.Client, namespace, activityID, runID string) (*activitypb.ActivityExecutionOutcome, error) {
	for {
		resp, err := cl.WorkflowService().PollActivityExecution(cctx, &workflowservice.PollActivityExecutionRequest{
			Namespace:  namespace,
			ActivityId: activityID,
			RunId:      runID,
		})
		if err != nil {
			return nil, err
		}
		if resp.GetOutcome() != nil {
			return resp.GetOutcome(), nil
		}
	}
}

func printActivityResult(cctx *CommandContext, activityID, runID string, result *common.Payloads) error {
	if cctx.JSONOutput {
		resultJSON, err := marshalActivityPayloads(cctx, result)
		if err != nil {
			return fmt.Errorf("failed marshaling result: %w", err)
		}
		return cctx.Printer.PrintStructured(struct {
			ActivityId string          `json:"activityId"`
			RunId      string          `json:"runId"`
			Status     string          `json:"status"`
			Result     json.RawMessage `json:"result"`
		}{
			ActivityId: activityID,
			RunId:      runID,
			Status:     "COMPLETED",
			Result:     resultJSON,
		}, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Results:"))
	var valuePtr interface{}
	if err := converter.GetDefaultDataConverter().FromPayloads(result, &valuePtr); err != nil {
		return fmt.Errorf("failed decoding result: %w", err)
	}
	resultJSON, err := json.Marshal(valuePtr)
	if err != nil {
		return fmt.Errorf("failed marshaling result: %w", err)
	}
	return cctx.Printer.PrintStructured(struct {
		Status string
		Result json.RawMessage `cli:",cardOmitEmpty"`
	}{
		Status: color.GreenString("COMPLETED"),
		Result: resultJSON,
	}, printer.StructuredOptions{})
}

func marshalActivityPayloads(cctx *CommandContext, payloads *common.Payloads) (json.RawMessage, error) {
	if cctx.JSONShorthandPayloads {
		var valuePtr interface{}
		if err := converter.GetDefaultDataConverter().FromPayloads(payloads, &valuePtr); err != nil {
			return nil, err
		}
		return json.Marshal(valuePtr)
	}
	return cctx.MarshalProtoJSON(payloads)
}

func printActivityFailure(cctx *CommandContext, activityID, runID string, f *failure.Failure) error {
	if cctx.JSONOutput {
		failureJSON, err := cctx.MarshalProtoJSON(f)
		if err != nil {
			return fmt.Errorf("failed marshaling failure: %w", err)
		}
		_ = cctx.Printer.PrintStructured(struct {
			ActivityId string          `json:"activityId"`
			RunId      string          `json:"runId"`
			Status     string          `json:"status"`
			Failure    json.RawMessage `json:"failure"`
		}{
			ActivityId: activityID,
			RunId:      runID,
			Status:     "FAILED",
			Failure:    failureJSON,
		}, printer.StructuredOptions{})
		return fmt.Errorf("activity failed")
	}

	cctx.Printer.Println(color.MagentaString("Results:"))
	_ = cctx.Printer.PrintStructured(struct {
		Status  string
		Failure string `cli:",cardOmitEmpty"`
	}{
		Status:  color.RedString("FAILED"),
		Failure: cctx.MarshalFriendlyFailureBodyText(f, "    "),
	}, printer.StructuredOptions{})
	return fmt.Errorf("activity failed")
}

func (c *TemporalActivityDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	handle := cl.GetActivityHandle(client.GetActivityHandleOptions{
		ActivityID: c.ActivityId,
		RunID:      c.RunId,
	})
	desc, err := handle.Describe(cctx, client.DescribeActivityOptions{})
	if err != nil {
		return fmt.Errorf("failed describing activity: %w", err)
	}
	if c.Raw || cctx.JSONOutput {
		return cctx.Printer.PrintStructured(desc.RawExecutionInfo, printer.StructuredOptions{})
	}
	return printActivityDescription(cctx, desc.RawExecutionInfo)
}

func printActivityDescription(cctx *CommandContext, info *activitypb.ActivityExecutionInfo) error {
	d := struct {
		ActivityId              string
		RunId                   string
		Type                    string
		Status                  string
		RunState                string `cli:",cardOmitEmpty"`
		TaskQueue               string
		ScheduleToCloseTimeout  time.Duration `cli:",cardOmitEmpty"`
		ScheduleToStartTimeout  time.Duration `cli:",cardOmitEmpty"`
		StartToCloseTimeout     time.Duration `cli:",cardOmitEmpty"`
		HeartbeatTimeout        time.Duration `cli:",cardOmitEmpty"`
		LastStartedTime         time.Time     `cli:",cardOmitEmpty"`
		Attempt                 int32
		ExecutionDuration       time.Duration `cli:",cardOmitEmpty"`
		ScheduleTime            time.Time     `cli:",cardOmitEmpty"`
		CloseTime               time.Time     `cli:",cardOmitEmpty"`
		LastFailure             string        `cli:",cardOmitEmpty"`
		LastWorkerIdentity      string        `cli:",cardOmitEmpty"`
		LastAttemptCompleteTime time.Time     `cli:",cardOmitEmpty"`
		StateTransitionCount    int64
	}{
		ActivityId:              info.GetActivityId(),
		RunId:                   info.GetRunId(),
		Type:                    info.GetActivityType().GetName(),
		Status:                  activityStatusShorthand(info.GetStatus()),
		RunState:                pendingActivityStateShorthand(info.GetRunState()),
		TaskQueue:               info.GetTaskQueue(),
		ScheduleToCloseTimeout:  info.GetScheduleToCloseTimeout().AsDuration(),
		ScheduleToStartTimeout:  info.GetScheduleToStartTimeout().AsDuration(),
		StartToCloseTimeout:     info.GetStartToCloseTimeout().AsDuration(),
		HeartbeatTimeout:        info.GetHeartbeatTimeout().AsDuration(),
		LastStartedTime:         timestampToTime(info.GetLastStartedTime()),
		Attempt:                 info.GetAttempt(),
		ExecutionDuration:       info.GetExecutionDuration().AsDuration(),
		ScheduleTime:            timestampToTime(info.GetScheduleTime()),
		CloseTime:               timestampToTime(info.GetCloseTime()),
		LastWorkerIdentity:      info.GetLastWorkerIdentity(),
		LastAttemptCompleteTime: timestampToTime(info.GetLastAttemptCompleteTime()),
		StateTransitionCount:    info.GetStateTransitionCount(),
	}
	if f := info.GetLastFailure(); f != nil {
		d.LastFailure = cctx.MarshalFriendlyFailureBodyText(f, "    ")
	}
	return cctx.Printer.PrintStructured(d, printer.StructuredOptions{})
}

func activityStatusShorthand(s enumspb.ActivityExecutionStatus) string {
	for name, val := range enumspb.ActivityExecutionStatus_shorthandValue {
		if int32(s) == val {
			return name
		}
	}
	return s.String()
}

func pendingActivityStateShorthand(s enumspb.PendingActivityState) string {
	for name, val := range enumspb.PendingActivityState_shorthandValue {
		if int32(s) == val && name != "Unspecified" {
			return name
		}
	}
	return ""
}

func (c *TemporalActivityListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
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

	result, err := cl.CountActivities(cctx, client.CountActivitiesOptions{Query: c.Query})
	if err != nil {
		return fmt.Errorf("failed counting activities: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}
	cctx.Printer.Printlnf("Total: %v", result.Count)
	for _, group := range result.Groups {
		var valueStr string
		for _, v := range group.GroupValues {
			if valueStr != "" {
				valueStr += ", "
			}
			valueStr += fmt.Sprintf("%v", v)
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

	handle := cl.GetActivityHandle(client.GetActivityHandleOptions{
		ActivityID: c.ActivityId,
		RunID:      c.RunId,
	})
	if err := handle.Cancel(cctx, client.CancelActivityOptions{Reason: c.Reason}); err != nil {
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

	// CONSIDER(dan): defaultReason is applied for terminate but not cancel, matching
	// the workflow pattern. It may be worth making this consistent across both.
	reason := c.Reason
	if reason == "" {
		reason = defaultReason()
	}
	handle := cl.GetActivityHandle(client.GetActivityHandleOptions{
		ActivityID: c.ActivityId,
		RunID:      c.RunId,
	})
	// Terminate may fail if the activity doesn't exist or has already completed.
	if err := handle.Terminate(cctx, client.TerminateActivityOptions{Reason: reason}); err != nil {
		return fmt.Errorf("failed to terminate activity: %w", err)
	}
	cctx.Printer.Println("Activity terminated")
	return nil
}

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
