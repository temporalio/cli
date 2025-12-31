package agent

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
)

// TimelineOptions configures the timeline generator.
type TimelineOptions struct {
	// Compact collapses repetitive events like retries.
	Compact bool
	// IncludePayloads includes payload data in the output.
	IncludePayloads bool
	// EventTypes filters to specific event types. If empty, all events are included.
	EventTypes []string
	// ExcludeEventTypes excludes specific event types.
	ExcludeEventTypes []string
}

// TimelineGenerator generates compact timelines from workflow history.
type TimelineGenerator struct {
	client client.Client
	opts   TimelineOptions
}

// NewTimelineGenerator creates a new timeline generator.
func NewTimelineGenerator(cl client.Client, opts TimelineOptions) *TimelineGenerator {
	return &TimelineGenerator{
		client: cl,
		opts:   opts,
	}
}

// Generate generates a timeline for the given workflow.
func (g *TimelineGenerator) Generate(ctx context.Context, namespace, workflowID, runID string) (*TimelineResult, error) {
	// If run ID not provided, describe the workflow first to get the latest run ID
	actualRunID := runID
	if runID == "" {
		desc, err := g.client.DescribeWorkflowExecution(ctx, workflowID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to describe workflow: %w", err)
		}
		actualRunID = desc.WorkflowExecutionInfo.Execution.GetRunId()
	}

	result := &TimelineResult{
		Workflow: WorkflowRef{
			Namespace:  namespace,
			WorkflowID: workflowID,
			RunID:      actualRunID,
		},
		Events: []TimelineEvent{},
	}

	// Track activities and timers for duration calculation
	activityStarts := make(map[int64]time.Time)       // scheduledEventId -> start time
	activityTypes := make(map[int64]string)           // scheduledEventId -> activity type
	activityIDs := make(map[int64]string)             // scheduledEventId -> activity id
	timerStarts := make(map[int64]time.Time)          // startedEventId -> start time
	childWorkflowStarts := make(map[int64]time.Time)  // initiatedEventId -> start time
	childWorkflowTypes := make(map[int64]string)      // initiatedEventId -> workflow type
	childWorkflowRefs := make(map[int64]*WorkflowRef) // initiatedEventId -> workflow ref

	// For compact mode: track retries
	activityRetries := make(map[string]int) // activityType -> retry count

	// Get workflow history (use actual run ID to ensure consistency)
	iter := g.client.GetWorkflowHistory(ctx, workflowID, actualRunID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

	for iter.HasNext() {
		historyEvent, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get history event: %w", err)
		}

		event := g.processEvent(historyEvent, activityStarts, activityTypes, activityIDs,
			timerStarts, childWorkflowStarts, childWorkflowTypes, childWorkflowRefs, activityRetries)

		if event == nil {
			continue
		}

		// Update workflow-level info from first event
		if historyEvent.GetEventType() == enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED {
			attrs := historyEvent.GetWorkflowExecutionStartedEventAttributes()
			result.WorkflowType = attrs.GetWorkflowType().GetName()
			result.Status = "Running"
			t := historyEvent.GetEventTime().AsTime()
			result.StartTime = &t
		}

		// Update status from close events
		switch historyEvent.GetEventType() {
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
			result.Status = "Completed"
			t := historyEvent.GetEventTime().AsTime()
			result.CloseTime = &t
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
			result.Status = "Failed"
			t := historyEvent.GetEventTime().AsTime()
			result.CloseTime = &t
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
			result.Status = "TimedOut"
			t := historyEvent.GetEventTime().AsTime()
			result.CloseTime = &t
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
			result.Status = "Canceled"
			t := historyEvent.GetEventTime().AsTime()
			result.CloseTime = &t
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
			result.Status = "Terminated"
			t := historyEvent.GetEventTime().AsTime()
			result.CloseTime = &t
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CONTINUED_AS_NEW:
			result.Status = "ContinuedAsNew"
			t := historyEvent.GetEventTime().AsTime()
			result.CloseTime = &t
		}

		// Filter events
		if !g.shouldIncludeEvent(event) {
			continue
		}

		result.Events = append(result.Events, *event)
		result.EventCount++
	}

	// Calculate duration
	if result.StartTime != nil && result.CloseTime != nil {
		result.DurationMs = result.CloseTime.Sub(*result.StartTime).Milliseconds()
	}

	// In compact mode, add summary events for retries
	if g.opts.Compact && len(activityRetries) > 0 {
		// The retry counts are already embedded in the events
	}

	return result, nil
}

func (g *TimelineGenerator) processEvent(
	event *history.HistoryEvent,
	activityStarts map[int64]time.Time,
	activityTypes map[int64]string,
	activityIDs map[int64]string,
	timerStarts map[int64]time.Time,
	childWorkflowStarts map[int64]time.Time,
	childWorkflowTypes map[int64]string,
	childWorkflowRefs map[int64]*WorkflowRef,
	activityRetries map[string]int,
) *TimelineEvent {
	ts := event.GetEventTime().AsTime()
	eventType := event.GetEventType().String()

	// Remove the prefix for cleaner output
	if len(eventType) > 11 && eventType[:11] == "EVENT_TYPE_" {
		eventType = eventType[11:]
	}

	te := &TimelineEvent{
		Timestamp: ts,
		EventID:   event.GetEventId(),
		Type:      eventType,
	}

	switch event.GetEventType() {
	// Workflow events
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
		te.Category = "workflow"
		te.Status = "started"
		attrs := event.GetWorkflowExecutionStartedEventAttributes()
		te.Name = attrs.GetWorkflowType().GetName()

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		te.Category = "workflow"
		te.Status = "completed"

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		te.Category = "workflow"
		te.Status = "failed"
		attrs := event.GetWorkflowExecutionFailedEventAttributes()
		te.Error = attrs.GetFailure().GetMessage()

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		te.Category = "workflow"
		te.Status = "timed_out"
		te.Error = "workflow execution timed out"

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		te.Category = "workflow"
		te.Status = "canceled"

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
		te.Category = "workflow"
		te.Status = "terminated"

	// Activity events
	case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
		te.Category = "activity"
		te.Status = "scheduled"
		attrs := event.GetActivityTaskScheduledEventAttributes()
		te.Name = attrs.GetActivityType().GetName()
		te.ActivityID = attrs.GetActivityId()
		activityStarts[event.EventId] = ts
		activityTypes[event.EventId] = attrs.GetActivityType().GetName()
		activityIDs[event.EventId] = attrs.GetActivityId()

	case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
		te.Category = "activity"
		te.Status = "started"
		attrs := event.GetActivityTaskStartedEventAttributes()
		te.Attempt = attrs.GetAttempt()
		if name, ok := activityTypes[attrs.GetScheduledEventId()]; ok {
			te.Name = name
		}
		if id, ok := activityIDs[attrs.GetScheduledEventId()]; ok {
			te.ActivityID = id
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
		te.Category = "activity"
		te.Status = "completed"
		attrs := event.GetActivityTaskCompletedEventAttributes()
		if name, ok := activityTypes[attrs.GetScheduledEventId()]; ok {
			te.Name = name
		}
		if id, ok := activityIDs[attrs.GetScheduledEventId()]; ok {
			te.ActivityID = id
		}
		if startTime, ok := activityStarts[attrs.GetScheduledEventId()]; ok {
			te.DurationMs = ts.Sub(startTime).Milliseconds()
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_FAILED:
		te.Category = "activity"
		te.Status = "failed"
		attrs := event.GetActivityTaskFailedEventAttributes()
		te.Error = attrs.GetFailure().GetMessage()
		if name, ok := activityTypes[attrs.GetScheduledEventId()]; ok {
			te.Name = name
			if g.opts.Compact {
				activityRetries[name]++
				te.RetryCount = activityRetries[name]
			}
		}
		if id, ok := activityIDs[attrs.GetScheduledEventId()]; ok {
			te.ActivityID = id
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
		te.Category = "activity"
		te.Status = "timed_out"
		attrs := event.GetActivityTaskTimedOutEventAttributes()
		te.Error = "activity timed out"
		if attrs.GetFailure() != nil {
			te.Error = attrs.GetFailure().GetMessage()
		}
		if name, ok := activityTypes[attrs.GetScheduledEventId()]; ok {
			te.Name = name
		}
		if id, ok := activityIDs[attrs.GetScheduledEventId()]; ok {
			te.ActivityID = id
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
		te.Category = "activity"
		te.Status = "canceled"
		attrs := event.GetActivityTaskCanceledEventAttributes()
		if name, ok := activityTypes[attrs.GetScheduledEventId()]; ok {
			te.Name = name
		}
		if id, ok := activityIDs[attrs.GetScheduledEventId()]; ok {
			te.ActivityID = id
		}

	// Timer events
	case enums.EVENT_TYPE_TIMER_STARTED:
		te.Category = "timer"
		te.Status = "started"
		attrs := event.GetTimerStartedEventAttributes()
		te.Name = attrs.GetTimerId()
		timerStarts[event.EventId] = ts

	case enums.EVENT_TYPE_TIMER_FIRED:
		te.Category = "timer"
		te.Status = "fired"
		attrs := event.GetTimerFiredEventAttributes()
		te.Name = attrs.GetTimerId()
		if startTime, ok := timerStarts[attrs.GetStartedEventId()]; ok {
			te.DurationMs = ts.Sub(startTime).Milliseconds()
		}

	case enums.EVENT_TYPE_TIMER_CANCELED:
		te.Category = "timer"
		te.Status = "canceled"
		attrs := event.GetTimerCanceledEventAttributes()
		te.Name = attrs.GetTimerId()

	// Child workflow events
	case enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED:
		te.Category = "child_workflow"
		te.Status = "initiated"
		attrs := event.GetStartChildWorkflowExecutionInitiatedEventAttributes()
		te.Name = attrs.GetWorkflowType().GetName()
		childWorkflowStarts[event.EventId] = ts
		childWorkflowTypes[event.EventId] = attrs.GetWorkflowType().GetName()
		childWorkflowRefs[event.EventId] = &WorkflowRef{
			Namespace:  attrs.GetNamespace(),
			WorkflowID: attrs.GetWorkflowId(),
		}
		te.ChildWorkflow = childWorkflowRefs[event.EventId]

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
		te.Category = "child_workflow"
		te.Status = "started"
		attrs := event.GetChildWorkflowExecutionStartedEventAttributes()
		if name, ok := childWorkflowTypes[attrs.GetInitiatedEventId()]; ok {
			te.Name = name
		}
		if ref, ok := childWorkflowRefs[attrs.GetInitiatedEventId()]; ok {
			ref.RunID = attrs.GetWorkflowExecution().GetRunId()
			te.ChildWorkflow = ref
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
		te.Category = "child_workflow"
		te.Status = "completed"
		attrs := event.GetChildWorkflowExecutionCompletedEventAttributes()
		if name, ok := childWorkflowTypes[attrs.GetInitiatedEventId()]; ok {
			te.Name = name
		}
		if ref, ok := childWorkflowRefs[attrs.GetInitiatedEventId()]; ok {
			te.ChildWorkflow = ref
		}
		if startTime, ok := childWorkflowStarts[attrs.GetInitiatedEventId()]; ok {
			te.DurationMs = ts.Sub(startTime).Milliseconds()
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED:
		te.Category = "child_workflow"
		te.Status = "failed"
		attrs := event.GetChildWorkflowExecutionFailedEventAttributes()
		te.Error = attrs.GetFailure().GetMessage()
		if name, ok := childWorkflowTypes[attrs.GetInitiatedEventId()]; ok {
			te.Name = name
		}
		if ref, ok := childWorkflowRefs[attrs.GetInitiatedEventId()]; ok {
			te.ChildWorkflow = ref
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
		te.Category = "child_workflow"
		te.Status = "timed_out"
		attrs := event.GetChildWorkflowExecutionTimedOutEventAttributes()
		te.Error = "child workflow timed out"
		if name, ok := childWorkflowTypes[attrs.GetInitiatedEventId()]; ok {
			te.Name = name
		}
		if ref, ok := childWorkflowRefs[attrs.GetInitiatedEventId()]; ok {
			te.ChildWorkflow = ref
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
		te.Category = "child_workflow"
		te.Status = "canceled"
		attrs := event.GetChildWorkflowExecutionCanceledEventAttributes()
		if name, ok := childWorkflowTypes[attrs.GetInitiatedEventId()]; ok {
			te.Name = name
		}
		if ref, ok := childWorkflowRefs[attrs.GetInitiatedEventId()]; ok {
			te.ChildWorkflow = ref
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TERMINATED:
		te.Category = "child_workflow"
		te.Status = "terminated"
		attrs := event.GetChildWorkflowExecutionTerminatedEventAttributes()
		if name, ok := childWorkflowTypes[attrs.GetInitiatedEventId()]; ok {
			te.Name = name
		}
		if ref, ok := childWorkflowRefs[attrs.GetInitiatedEventId()]; ok {
			te.ChildWorkflow = ref
		}

	// Signal events
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED:
		te.Category = "signal"
		te.Status = "received"
		attrs := event.GetWorkflowExecutionSignaledEventAttributes()
		te.Name = attrs.GetSignalName()

	// Other events - return nil to skip them in compact mode
	default:
		if g.opts.Compact {
			return nil
		}
		te.Category = "other"
	}

	return te
}

func (g *TimelineGenerator) shouldIncludeEvent(event *TimelineEvent) bool {
	// Check include filter
	if len(g.opts.EventTypes) > 0 {
		found := false
		for _, et := range g.opts.EventTypes {
			if et == event.Type || et == event.Category {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check exclude filter
	for _, et := range g.opts.ExcludeEventTypes {
		if et == event.Type || et == event.Category {
			return false
		}
	}

	return true
}
