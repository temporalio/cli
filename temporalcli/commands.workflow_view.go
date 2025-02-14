package temporalcli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

func (c *TemporalWorkflowDescribeCommand) run(cctx *CommandContext, args []string) error {
	// Call describe
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	resp, err := cl.DescribeWorkflowExecution(cctx, c.WorkflowId, c.RunId)
	if err != nil {
		return fmt.Errorf("failed describing workflow: %w", err)
	}

	// Print reset points if that is all that is wanted
	if c.ResetPoints {
		points := resp.WorkflowExecutionInfo.AutoResetPoints.GetPoints()
		cctx.Printer.Println(color.MagentaString("Auto Reset Points: %v", len(points)))
		pts := make([]struct {
			BinaryChecksum string
			CreateTime     time.Time
			RunId          string
			EventId        int64
		}, len(points))
		for i, p := range points {
			pts[i].BinaryChecksum = p.BinaryChecksum
			pts[i].CreateTime = timestampToTime(p.CreateTime)
			pts[i].RunId = p.RunId
			pts[i].EventId = p.FirstWorkflowTaskCompletedId
		}
		_ = cctx.Printer.PrintStructured(pts, printer.StructuredOptions{})
		return nil
	}

	// Try to also get the close event if the description says completed. We don't
	// just ask always because we don't want the race where it may have finished
	// between when we called describe and now
	var closeEvent *history.HistoryEvent
	running := resp.WorkflowExecutionInfo.Status == enums.WORKFLOW_EXECUTION_STATUS_RUNNING
	if !running {
		iter := cl.GetWorkflowHistory(cctx,
			resp.WorkflowExecutionInfo.Execution.WorkflowId,
			resp.WorkflowExecutionInfo.Execution.RunId,
			false,
			enums.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT,
		)
		if !iter.HasNext() {
			return fmt.Errorf("missing close event: %w", err)
		} else if closeEvent, err = iter.Next(); err != nil {
			return fmt.Errorf("failed getting close event: %w", err)
		}
	}

	// Print JSON
	if cctx.JSONOutput {
		// We want to inject the "closeEvent" and "result" into the same structure,
		// and in order to do that, we need to serialize the protojson to a map, add
		// the fields, then re-serialize.
		var toPrint any = resp
		if closeEvent != nil {
			var respObj map[string]any
			b, err := cctx.MarshalProtoJSON(resp)
			if err != nil {
				return fmt.Errorf("failed marshaling response object: %w", err)
			}
			if err := json.Unmarshal(b, &respObj); err != nil {
				return fmt.Errorf("failed unmarshaling: %w", err)
			}
			b, err = cctx.MarshalProtoJSON(closeEvent)
			if err != nil {
				return fmt.Errorf("failed marshaling close event: %w", err)
			}
			respObj["closeEvent"] = json.RawMessage(b)
			if attr := closeEvent.GetWorkflowExecutionCompletedEventAttributes(); attr != nil {
				respObj["result"], err = cctx.MarshalFriendlyJSONPayloads(
					closeEvent.GetWorkflowExecutionCompletedEventAttributes().GetResult())
				if err != nil {
					return fmt.Errorf("failed marshaling result: %w", err)
				}
			}
			toPrint = respObj
		}
		return cctx.Printer.PrintStructured(toPrint, printer.StructuredOptions{})
	}

	cctx.Printer.Println(color.MagentaString("Execution Info:"))
	info := resp.WorkflowExecutionInfo
	_ = cctx.Printer.PrintStructured(struct {
		WorkflowId           string
		RunId                string
		Type                 string
		Namespace            string
		TaskQueue            string
		AssignedBuildId      string
		StartTime            time.Time
		CloseTime            time.Time                  `cli:",cardOmitEmpty"`
		ExecutionTime        time.Time                  `cli:",cardOmitEmpty"`
		Memo                 map[string]*common.Payload `cli:",cardOmitEmpty"`
		SearchAttributes     map[string]*common.Payload `cli:",cardOmitEmpty"`
		StateTransitionCount int64
		HistoryLength        int64
		HistorySize          int64
	}{
		WorkflowId:           info.Execution.WorkflowId,
		RunId:                info.Execution.RunId,
		Type:                 info.Type.GetName(),
		Namespace:            c.Parent.Namespace,
		TaskQueue:            info.TaskQueue,
		AssignedBuildId:      info.GetAssignedBuildId(),
		StartTime:            timestampToTime(info.StartTime),
		CloseTime:            timestampToTime(info.CloseTime),
		ExecutionTime:        timestampToTime(info.ExecutionTime),
		Memo:                 info.Memo.GetFields(),
		SearchAttributes:     info.SearchAttributes.GetIndexedFields(),
		StateTransitionCount: info.StateTransitionCount,
		HistoryLength:        info.HistoryLength,
		HistorySize:          info.HistorySizeBytes,
	}, printer.StructuredOptions{})

	staticSummary := resp.GetExecutionConfig().GetUserMetadata().GetSummary()
	staticDetails := resp.GetExecutionConfig().GetUserMetadata().GetDetails()
	if len(staticSummary.GetData()) > 0 || len(staticDetails.GetData()) > 0 {
		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Metadata:"))
		_ = cctx.Printer.PrintStructured(struct {
			StaticSummary *common.Payload
			StaticDetails *common.Payload
		}{
			StaticSummary: staticSummary,
			StaticDetails: staticDetails,
		}, printer.StructuredOptions{})
	}

	if info.VersioningInfo != nil {
		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Versioning Info:"))
		cctx.Printer.Println()
		vInfo := info.VersioningInfo
		_ = cctx.Printer.PrintStructured(struct {
			Behavior                       string
			DeploymentSeriesName           string
			DeploymentBuildID              string
			OverrideBehavior               string `cli:",cardOmitEmpty"`
			OverrideDeploymentSeriesName   string `cli:",cardOmitEmpty"`
			OverrideDeploymentBuildID      string `cli:",cardOmitEmpty"`
			TransitionDeploymentSeriesName string `cli:",cardOmitEmpty"`
			TransitionDeploymentBuildID    string `cli:",cardOmitEmpty"`
		}{
			Behavior:                       vInfo.Behavior.String(),
			DeploymentSeriesName:           vInfo.Deployment.GetSeriesName(),
			DeploymentBuildID:              vInfo.Deployment.GetBuildId(),
			OverrideBehavior:               vInfo.VersioningOverride.GetBehavior().String(),
			OverrideDeploymentSeriesName:   vInfo.VersioningOverride.GetDeployment().GetSeriesName(),
			OverrideDeploymentBuildID:      vInfo.VersioningOverride.GetDeployment().GetBuildId(),
			TransitionDeploymentSeriesName: vInfo.DeploymentTransition.GetDeployment().GetSeriesName(),
			TransitionDeploymentBuildID:    vInfo.DeploymentTransition.GetDeployment().GetBuildId(),
		}, printer.StructuredOptions{})
	}

	if len(resp.Callbacks) > 0 {
		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Callbacks: %v", len(resp.Callbacks)))
		cctx.Printer.Println()
		cbs := make([]struct {
			URL                     string
			Trigger                 string
			State                   enums.CallbackState
			Attempt                 int32
			RegistrationTime        time.Time        `cli:",cardOmitEmpty"`
			NextAttemptScheduleTime time.Time        `cli:",cardOmitEmpty"`
			LastAttemptCompleteTime time.Time        `cli:",cardOmitEmpty"`
			LastAttemptFailure      *failure.Failure `cli:",cardOmitEmpty"`
			BlockedReason           string           `cli:",cardOmitEmpty"`
		}, len(resp.Callbacks))
		for i, cb := range resp.Callbacks {
			cbs[i].URL = cb.GetCallback().GetNexus().GetUrl()
			cbs[i].State = cb.GetState()
			cbs[i].Attempt = cb.GetAttempt()
			cbs[i].LastAttemptFailure = cb.LastAttemptFailure
			cbs[i].LastAttemptCompleteTime = timestampToTime(cb.LastAttemptCompleteTime)
			cbs[i].NextAttemptScheduleTime = timestampToTime(cb.NextAttemptScheduleTime)
			cbs[i].RegistrationTime = timestampToTime(cb.RegistrationTime)
			cbs[i].Trigger = "Unknown"
			if cb.GetTrigger().GetWorkflowClosed() != nil {
				cbs[i].Trigger = "WorkflowClosed"
			}
			cbs[i].BlockedReason = cb.GetBlockedReason()
		}
		_ = cctx.Printer.PrintStructured(cbs, printer.StructuredOptions{})
		cctx.Printer.Println()
	}

	if running {
		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Pending Activities: %v", len(resp.PendingActivities)))
		if len(resp.PendingActivities) > 0 {
			cctx.Printer.Println()
			acts := make([]struct {
				ActivityId           string
				Type                 string
				State                enums.PendingActivityState
				Attempt              int32
				MaximumAttempts      int32
				ScheduledTime        time.Time
				LastStartedTime      time.Time         `cli:",cardOmitEmpty"`
				LastHeartbeatTime    time.Time         `cli:",cardOmitEmpty"`
				ExpirationTime       time.Time         `cli:",cardOmitEmpty"`
				LastFailure          *failure.Failure  `cli:",cardOmitEmpty"`
				LastWorkerIdentity   string            `cli:",cardOmitEmpty"`
				LastHeartbeatDetails []*common.Payload `cli:",cardOmitEmpty"`
			}, len(resp.PendingActivities))
			for i, a := range resp.PendingActivities {
				acts[i].ActivityId = a.ActivityId
				acts[i].Type = a.ActivityType.GetName()
				acts[i].State = a.State
				acts[i].Attempt = a.Attempt
				acts[i].MaximumAttempts = a.MaximumAttempts
				acts[i].ScheduledTime = timestampToTime(a.ScheduledTime)
				acts[i].LastStartedTime = timestampToTime(a.LastStartedTime)
				acts[i].LastHeartbeatTime = timestampToTime(a.LastHeartbeatTime)
				acts[i].ExpirationTime = timestampToTime(a.ExpirationTime)
				acts[i].LastFailure = a.LastFailure
				acts[i].LastWorkerIdentity = a.LastWorkerIdentity
				acts[i].LastHeartbeatDetails = a.HeartbeatDetails.GetPayloads()
			}
			_ = cctx.Printer.PrintStructured(acts, printer.StructuredOptions{})
			cctx.Printer.Println()
		}

		cctx.Printer.Println(color.MagentaString("Pending Child Workflows: %v", len(resp.PendingChildren)))
		if len(resp.PendingChildren) > 0 {
			cctx.Printer.Println()
			_ = cctx.Printer.PrintStructured(resp.PendingChildren, printer.StructuredOptions{})
		}

		cctx.Printer.Println(color.MagentaString("Pending Nexus Operations: %v", len(resp.PendingNexusOperations)))
		if len(resp.PendingNexusOperations) > 0 {
			cctx.Printer.Println()
			ops := make([]struct {
				Endpoint                           string
				Service                            string
				Operation                          string
				OperationID                        string
				State                              enums.PendingNexusOperationState
				Attempt                            int32
				ScheduleToCloseTimeout             string                                `cli:",cardOmitEmpty"`
				NextAttemptScheduleTime            time.Time                             `cli:",cardOmitEmpty"`
				LastAttemptCompleteTime            time.Time                             `cli:",cardOmitEmpty"`
				LastAttemptFailure                 *failure.Failure                      `cli:",cardOmitEmpty"`
				BlockedReason                      string                                `cli:",cardOmitEmpty"`
				CancelationState                   enums.NexusOperationCancellationState `cli:",cardOmitEmpty"`
				CancelationAttempt                 int32                                 `cli:",cardOmitEmpty"`
				CancelationRequestedTime           time.Time                             `cli:",cardOmitEmpty"`
				CancelationNextAttemptScheduleTime time.Time                             `cli:",cardOmitEmpty"`
				CancelationLastAttemptCompleteTime time.Time                             `cli:",cardOmitEmpty"`
				CancelationLastAttemptFailure      *failure.Failure                      `cli:",cardOmitEmpty"`
				CancelationBlockedReason           string                                `cli:",cardOmitEmpty"`
			}, len(resp.PendingNexusOperations))
			for i, op := range resp.PendingNexusOperations {
				ops[i].Endpoint = op.GetEndpoint()
				ops[i].Service = op.GetService()
				ops[i].Operation = op.GetOperation()
				ops[i].OperationID = op.GetOperationId()
				ops[i].State = op.GetState()
				ops[i].Attempt = op.GetAttempt()
				ops[i].LastAttemptFailure = op.LastAttemptFailure
				ops[i].LastAttemptCompleteTime = timestampToTime(op.LastAttemptCompleteTime)
				ops[i].NextAttemptScheduleTime = timestampToTime(op.NextAttemptScheduleTime)
				ops[i].ScheduleToCloseTimeout = formatDuration(op.GetScheduleToCloseTimeout().AsDuration())
				ops[i].BlockedReason = op.GetBlockedReason()
				ops[i].CancelationState = op.GetCancellationInfo().GetState()
				ops[i].CancelationAttempt = op.GetCancellationInfo().GetAttempt()
				ops[i].CancelationLastAttemptFailure = op.GetCancellationInfo().GetLastAttemptFailure()
				ops[i].CancelationLastAttemptCompleteTime = timestampToTime(op.GetCancellationInfo().GetLastAttemptCompleteTime())
				ops[i].CancelationNextAttemptScheduleTime = timestampToTime(op.GetCancellationInfo().GetNextAttemptScheduleTime())
				ops[i].CancelationRequestedTime = timestampToTime(op.GetCancellationInfo().GetRequestedTime())
				ops[i].CancelationBlockedReason = op.GetCancellationInfo().GetBlockedReason()
			}
			_ = cctx.Printer.PrintStructured(ops, printer.StructuredOptions{})
			cctx.Printer.Println()
		}
	} else if closeEvent != nil {
		cctx.Printer.Println()
		var duration time.Duration
		if resp.WorkflowExecutionInfo.StartTime != nil && resp.WorkflowExecutionInfo.CloseTime != nil {
			duration = resp.WorkflowExecutionInfo.CloseTime.AsTime().Sub(resp.WorkflowExecutionInfo.StartTime.AsTime())
		}
		if err := printTextResult(cctx, closeEvent, duration); err != nil {
			return fmt.Errorf("failed printing result: %w", err)
		}
	}
	return nil
}

func (c *TemporalWorkflowListCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// This is a listing command subject to json vs jsonl rules
	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	// Build request and start looping. We always use default page size regardless
	// of user-defined limit, because we're ok w/ extra page data and the default
	// is not clearly defined.
	pageFetcher := c.pageFetcher(cctx, cl)
	var nextPageToken []byte
	var execsProcessed int
	for pageIndex := 0; ; pageIndex++ {
		page, err := pageFetcher(nextPageToken)
		if err != nil {
			return fmt.Errorf("failed listing workflows: %w", err)
		}
		var textTable []map[string]any
		for _, exec := range page.GetExecutions() {
			if c.Limit > 0 && execsProcessed >= c.Limit {
				break
			}
			execsProcessed++
			// For JSON we are going to dump one line of JSON per execution
			if cctx.JSONOutput {
				_ = cctx.Printer.PrintStructured(exec, printer.StructuredOptions{})
			} else {
				// For non-JSON, we are doing a table for each page
				textTable = append(textTable, map[string]any{
					"Status":     exec.Status,
					"WorkflowId": exec.Execution.WorkflowId,
					"Type":       exec.Type.GetName(),
					"StartTime":  exec.StartTime.AsTime(),
				})
			}
		}
		// Print table, headers only on first table
		if len(textTable) > 0 {
			_ = cctx.Printer.PrintStructured(textTable, printer.StructuredOptions{
				Fields: []string{"Status", "WorkflowId", "Type", "StartTime"},
				Table:  &printer.TableOptions{NoHeader: pageIndex > 0},
			})
		}
		// Stop if next page token non-existing or executions reached limit
		nextPageToken = page.GetNextPageToken()
		if len(nextPageToken) == 0 || (c.Limit > 0 && execsProcessed >= c.Limit) {
			return nil
		}
	}
}

type workflowPage interface {
	GetExecutions() []*workflow.WorkflowExecutionInfo
	GetNextPageToken() []byte
}

func (c *TemporalWorkflowListCommand) pageFetcher(
	cctx *CommandContext,
	cl client.Client,
) func(next []byte) (workflowPage, error) {
	return func(next []byte) (workflowPage, error) {
		if c.Archived {
			return cl.ListArchivedWorkflow(cctx, &workflowservice.ListArchivedWorkflowExecutionsRequest{
				Query:         c.Query,
				NextPageToken: next,
			})
		}
		return cl.ListWorkflow(cctx, &workflowservice.ListWorkflowExecutionsRequest{
			Query:         c.Query,
			NextPageToken: next,
		})
	}
}

func (c *TemporalWorkflowCountCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().CountWorkflowExecutions(cctx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: c.Parent.Namespace,
		Query:     c.Query,
	})
	if err != nil {
		return err
	}

	// Just dump response on JSON, otherwise print total and groups
	if cctx.JSONOutput {
		// Shorthand does not apply to search attributes currently, so we're going
		// to remove the "type" from the metadata encoding on group values to make
		// it apply
		for _, group := range resp.Groups {
			for _, payload := range group.GroupValues {
				delete(payload.GetMetadata(), "type")
			}
		}
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	cctx.Printer.Printlnf("Total: %v", resp.Count)
	for _, group := range resp.Groups {
		// Payload values are search attributes, so we can use the default converter
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

func (c *TemporalWorkflowResultCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	run := cl.GetWorkflow(cctx, c.WorkflowId, c.RunId)

	// Get the close event, following continue as new
	var closeEvent *history.HistoryEvent
	for runID := c.RunId; closeEvent == nil; {
		iter := cl.GetWorkflowHistory(cctx, c.WorkflowId, runID, true, enums.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT)
		if !iter.HasNext() {
			return fmt.Errorf("missing close event")
		} else if closeEvent, err = iter.Next(); err != nil {
			return fmt.Errorf("failed getting close event: %w", err)
		} else if canAttr := closeEvent.GetWorkflowExecutionContinuedAsNewEventAttributes(); canAttr != nil {
			closeEvent, runID = nil, canAttr.NewExecutionRunId
		}
	}

	// Print result
	if cctx.JSONOutput {
		result := &workflowJSONResult{
			WorkflowId:     run.GetID(),
			RunId:          run.GetRunID(),
			Type:           "",
			Namespace:      c.Parent.Namespace,
			TaskQueue:      "",
			DurationMillis: 0,
			Status:         "<unknown>",
			CloseEvent:     json.RawMessage("null"),
		}
		if err := result.setCloseEvent(cctx, closeEvent); err != nil {
			return err
		}
		err = cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	} else {
		err = printTextResult(cctx, closeEvent, 0)
	}
	// Log print failure and return workflow failure if workflow failed
	if closeEvent.EventType != enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED {
		if err != nil {
			cctx.Logger.Error("Workflow failed, and printing the output also failed", "error", err)
		}
		err = fmt.Errorf("workflow failed")
	}
	return err
}

func (c *TemporalWorkflowShowCommand) run(cctx *CommandContext, _ []string) error {
	// Call describe
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Print history
	iter := &structuredHistoryIter{
		ctx:            cctx,
		client:         cl,
		workflowID:     c.WorkflowId,
		runID:          c.RunId,
		includeDetails: c.Detailed,
		follow:         c.Follow,
	}
	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString("Progress:"))
		if err := iter.print(cctx); err != nil {
			return fmt.Errorf("displaying history failed: %w", err)
		}
		cctx.Printer.Println()
		if err := printTextResult(cctx, iter.wfResult, 0); err != nil {
			return err
		}
	} else {
		events := make([]*history.HistoryEvent, 0)
		for {
			e, err := iter.NextRawEvent()
			if err != nil {
				return fmt.Errorf("failed getting next history event: %w", err)
			}
			if e == nil {
				break
			}
			events = append(events, e)
		}
		outStruct := history.History{}
		outStruct.Events = events
		// We intentionally disable shorthand because "workflow show" for JSON needs
		// to support SDK replayers which do not work with shorthand
		jsonPayloadShorthand := false
		err = cctx.Printer.PrintStructured(&outStruct, printer.StructuredOptions{
			OverrideJSONPayloadShorthand: &jsonPayloadShorthand,
		})
		if err != nil {
			return fmt.Errorf("failed printing structured output: %w", err)
		}
	}
	return nil
}
