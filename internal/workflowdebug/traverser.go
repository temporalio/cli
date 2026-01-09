package workflowdebug

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// ClientProvider creates Temporal clients for different namespaces.
type ClientProvider interface {
	// GetClient returns a client for the given namespace.
	GetClient(ctx context.Context, namespace string) (client.Client, error)
}

// TraverserOptions configures the chain traverser.
type TraverserOptions struct {
	// FollowNamespaces is the set of namespaces to follow when traversing child workflows.
	// If empty, only the starting namespace is used.
	FollowNamespaces []string
	// MaxDepth is the maximum depth to traverse. 0 means unlimited.
	MaxDepth int
}

// ChainTraverser traverses workflow chains to find failures.
type ChainTraverser struct {
	clientProvider ClientProvider
	opts           TraverserOptions
	visited        map[visitedKey]struct{}
}

type visitedKey struct {
	namespace  string
	workflowID string
	runID      string
}

// NewChainTraverser creates a new chain traverser.
func NewChainTraverser(clientProvider ClientProvider, opts TraverserOptions) *ChainTraverser {
	return &ChainTraverser{
		clientProvider: clientProvider,
		opts:           opts,
		visited:        make(map[visitedKey]struct{}),
	}
}

// Trace traces a workflow through its child chain to find the deepest failure.
func (t *ChainTraverser) Trace(ctx context.Context, namespace, workflowID, runID string) (*TraceResult, error) {
	// If run ID not provided, describe the workflow first to get the latest run ID
	actualRunID := runID
	if runID == "" {
		cl, err := t.clientProvider.GetClient(ctx, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get client for namespace %s: %w", namespace, err)
		}
		desc, err := cl.DescribeWorkflowExecution(ctx, workflowID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to describe workflow: %w", err)
		}
		actualRunID = desc.WorkflowExecutionInfo.Execution.GetRunId()
	}

	chain, rootCause, err := t.traceRecursive(ctx, namespace, workflowID, actualRunID, 0)
	if err != nil {
		return nil, err
	}

	// Mark the last node as leaf if chain is non-empty
	if len(chain) > 0 {
		chain[len(chain)-1].IsLeaf = true
	}

	result := &TraceResult{
		Chain:     chain,
		RootCause: rootCause,
		Depth:     len(chain) - 1,
	}
	if result.Depth < 0 {
		result.Depth = 0
	}

	return result, nil
}

func (t *ChainTraverser) traceRecursive(
	ctx context.Context,
	namespace, workflowID, runID string,
	depth int,
) ([]WorkflowChainNode, *RootCause, error) {
	// Check depth limit
	if t.opts.MaxDepth > 0 && depth >= t.opts.MaxDepth {
		return nil, nil, nil
	}

	// Check if we've already visited this workflow
	key := visitedKey{namespace, workflowID, runID}
	if _, ok := t.visited[key]; ok {
		return nil, nil, nil
	}
	t.visited[key] = struct{}{}

	// Get client for this namespace
	cl, err := t.clientProvider.GetClient(ctx, namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client for namespace %s: %w", namespace, err)
	}

	// Build the state machine to process events
	sm := newStateMachine()

	// Get workflow history
	iter := cl.GetWorkflowHistory(ctx, workflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get history event: %w", err)
		}
		sm.processEvent(event)
	}

	// Build the current node
	// Use the passed-in runID if provided, otherwise fall back to state machine's runID
	actualRunID := runID
	if actualRunID == "" {
		actualRunID = sm.runID
	}
	node := WorkflowChainNode{
		Namespace:    namespace,
		WorkflowID:   workflowID,
		RunID:        actualRunID,
		WorkflowType: sm.workflowType,
		Status:       WorkflowStatusFromEnum(sm.status),
		Depth:        depth,
	}

	if sm.startTime != nil {
		node.StartTime = sm.startTime
	}
	if sm.closeTime != nil {
		node.CloseTime = sm.closeTime
		if sm.startTime != nil {
			node.DurationMs = sm.closeTime.Sub(*sm.startTime).Milliseconds()
		}
	}
	if sm.failureMessage != "" {
		node.Error = sm.failureMessage
	}

	chain := []WorkflowChainNode{node}

	// Find the deepest failure by following failed children and Nexus operations
	var deepestRootCause *RootCause

	// First check for failed child workflows
	for _, child := range sm.failedChildren {
		if !t.canFollowNamespace(child.namespace) {
			continue
		}

		childChain, childRootCause, err := t.traceRecursive(ctx, child.namespace, child.workflowID, child.runID, depth+1)
		if err != nil {
			return nil, nil, err
		}

		if len(childChain) > 0 {
			chain = append(chain, childChain...)
			if childRootCause != nil {
				deepestRootCause = childRootCause
			}
		}
	}

	// Then check for failed Nexus operations
	for _, nexusOp := range sm.failedNexusOps {
		// Skip if we don't have target workflow info
		if nexusOp.namespace == "" || nexusOp.workflowID == "" {
			continue
		}

		if !t.canFollowNamespace(nexusOp.namespace) {
			continue
		}

		nexusChain, nexusRootCause, err := t.traceRecursive(ctx, nexusOp.namespace, nexusOp.workflowID, nexusOp.runID, depth+1)
		if err != nil {
			// Log but continue - may not have access to target namespace
			continue
		}

		if len(nexusChain) > 0 {
			chain = append(chain, nexusChain...)
			if nexusRootCause != nil {
				deepestRootCause = nexusRootCause
			}
		}
	}

	// If no failed children/Nexus found root cause, check activities
	if deepestRootCause == nil && len(sm.failedActivities) > 0 {
		// Use the first failed activity as root cause
		act := sm.failedActivities[0]
		deepestRootCause = &RootCause{
			Type:      "ActivityFailed",
			Activity:  act.activityType,
			Error:     act.failureMessage,
			Timestamp: act.failureTime,
			Workflow: &WorkflowRef{
				Namespace:  namespace,
				WorkflowID: workflowID,
				RunID:      actualRunID,
			},
		}
	}

	// If no root cause yet but have failed Nexus ops (without target info), use that
	if deepestRootCause == nil && len(sm.failedNexusOps) > 0 {
		nexusOp := sm.failedNexusOps[0]
		deepestRootCause = &RootCause{
			Type:  "NexusOperationFailed",
			Error: fmt.Sprintf("nexus %s/%s: %s", nexusOp.service, nexusOp.operation, nexusOp.failureMessage),
			Workflow: &WorkflowRef{
				Namespace:  namespace,
				WorkflowID: workflowID,
				RunID:      actualRunID,
			},
		}
	}

	// If still no root cause and workflow itself failed, use that
	if deepestRootCause == nil && sm.failureMessage != "" {
		deepestRootCause = &RootCause{
			Type:      "WorkflowFailed",
			Error:     sm.failureMessage,
			Timestamp: sm.closeTime,
			Workflow: &WorkflowRef{
				Namespace:  namespace,
				WorkflowID: workflowID,
				RunID:      actualRunID,
			},
		}
	}

	// Check for timeout
	if deepestRootCause == nil && sm.status == enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT {
		deepestRootCause = &RootCause{
			Type:      "Timeout",
			Error:     "workflow execution timed out",
			Timestamp: sm.closeTime,
			Workflow: &WorkflowRef{
				Namespace:  namespace,
				WorkflowID: workflowID,
				RunID:      actualRunID,
			},
		}
	}

	return chain, deepestRootCause, nil
}

func (t *ChainTraverser) canFollowNamespace(ns string) bool {
	// Always allow the starting namespace (implicit)
	if len(t.opts.FollowNamespaces) == 0 {
		return true
	}
	for _, allowed := range t.opts.FollowNamespaces {
		if allowed == ns {
			return true
		}
	}
	return false
}

// FailuresOptions configures the failures finder.
type FailuresOptions struct {
	// Since is the time window to search for failures.
	Since time.Duration
	// Statuses to filter by. If empty, defaults to Failed and TimedOut.
	Statuses []enums.WorkflowExecutionStatus
	// FollowChildren determines whether to traverse child workflows.
	FollowChildren bool
	// FollowNamespaces is the set of namespaces to follow when traversing child workflows.
	FollowNamespaces []string
	// MaxDepth is the maximum depth to traverse. 0 means unlimited.
	MaxDepth int
	// Limit is the maximum number of failures to return.
	Limit int
	// ErrorContains filters failures to only those containing this substring in the error message.
	// Case-insensitive matching.
	ErrorContains string
	// LeafOnly, when true, shows only leaf failures (workflows with no failing children).
	// Parent workflows that failed due to child workflow failures are excluded.
	LeafOnly bool
	// CompactErrors, when true, extracts the core error message and strips wrapper context.
	CompactErrors bool
	// GroupBy specifies how to group failures. Valid values: "none", "type", "namespace", "status", "error".
	GroupBy string
}

// FailuresFinder finds recent workflow failures.
type FailuresFinder struct {
	clientProvider ClientProvider
	opts           FailuresOptions
}

// NewFailuresFinder creates a new failures finder.
func NewFailuresFinder(clientProvider ClientProvider, opts FailuresOptions) *FailuresFinder {
	return &FailuresFinder{
		clientProvider: clientProvider,
		opts:           opts,
	}
}

// FindFailures finds recent workflow failures in the given namespace.
func (f *FailuresFinder) FindFailures(ctx context.Context, namespace string) (*FailuresResult, error) {
	cl, err := f.clientProvider.GetClient(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for namespace %s: %w", namespace, err)
	}

	// Build the query
	query := f.buildQuery()

	// List workflows
	failures := make([]FailureReport, 0) // Initialize as empty slice, not nil (for JSON serialization)
	var nextPageToken []byte

	// Track parent workflow IDs for leaf-only filtering
	// Key: workflow ID, Value: true if this workflow is a parent of a failing child
	parentWorkflows := make(map[string]bool)

	for {
		resp, err := cl.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			Query:         query,
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list workflows: %w", err)
		}

		for _, exec := range resp.GetExecutions() {
			// For leaf-only, we may need to collect more than the limit initially
			// since we'll filter some out
			if !f.opts.LeafOnly && f.opts.Limit > 0 && len(failures) >= f.opts.Limit {
				break
			}

			report := FailureReport{
				RootWorkflow: WorkflowRef{
					Namespace:  namespace,
					WorkflowID: exec.GetExecution().GetWorkflowId(),
					RunID:      exec.GetExecution().GetRunId(),
				},
				Status: WorkflowStatusFromEnum(exec.GetStatus()),
				Chain:  []string{exec.GetExecution().GetWorkflowId()},
				Depth:  0,
			}

			if exec.GetCloseTime() != nil {
				t := exec.GetCloseTime().AsTime()
				report.Timestamp = &t
			}

			// If following children, trace the workflow
			// Always trace to get detailed root cause (activity name, actual error)
			// Use MaxDepth=1 to only trace the current workflow's activities (no child following)
			// unless FollowChildren is enabled
			maxDepth := 1 // Only the root workflow, don't follow children
			if f.opts.FollowChildren {
				maxDepth = f.opts.MaxDepth // 0 = unlimited, or user-specified limit
			}
			traverser := NewChainTraverser(f.clientProvider, TraverserOptions{
				FollowNamespaces: f.opts.FollowNamespaces,
				MaxDepth:         maxDepth,
			})

			traceResult, err := traverser.Trace(ctx, namespace, exec.GetExecution().GetWorkflowId(), exec.GetExecution().GetRunId())
			if err != nil {
				// Log but continue
				report.RootCause = fmt.Sprintf("failed to trace: %v", err)
			} else if traceResult != nil {
				report.Depth = traceResult.Depth
				report.Chain = make([]string, len(traceResult.Chain))
				for i, node := range traceResult.Chain {
					report.Chain[i] = node.WorkflowID
				}
				if traceResult.RootCause != nil {
					if traceResult.RootCause.Activity != "" {
						report.RootCause = fmt.Sprintf("%s: %s - %s", traceResult.RootCause.Type, traceResult.RootCause.Activity, traceResult.RootCause.Error)
					} else {
						report.RootCause = fmt.Sprintf("%s: %s", traceResult.RootCause.Type, traceResult.RootCause.Error)
					}
					if len(traceResult.Chain) > 0 {
						lastNode := traceResult.Chain[len(traceResult.Chain)-1]
						report.LeafFailure = &WorkflowRef{
							Namespace:  lastNode.Namespace,
							WorkflowID: lastNode.WorkflowID,
							RunID:      lastNode.RunID,
						}
					}
				}

				// Track parent workflows for leaf-only filtering
				// All workflows in the chain except the last one are parents
				if f.opts.LeafOnly && len(traceResult.Chain) > 1 {
					for i := 0; i < len(traceResult.Chain)-1; i++ {
						parentWorkflows[traceResult.Chain[i].WorkflowID] = true
					}
				}
			}

			// Filter by error message if specified
			if f.opts.ErrorContains != "" {
				if !strings.Contains(strings.ToLower(report.RootCause), strings.ToLower(f.opts.ErrorContains)) {
					continue
				}
			}

			// Apply error compaction if requested
			if f.opts.CompactErrors && report.RootCause != "" {
				report.RootCause = CompactErrorWithContext(report.RootCause)
			}

			failures = append(failures, report)
		}

		if !f.opts.LeafOnly && f.opts.Limit > 0 && len(failures) >= f.opts.Limit {
			break
		}

		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 {
			break
		}
	}

	// Apply leaf-only filtering
	if f.opts.LeafOnly {
		var leafFailures []FailureReport
		for _, report := range failures {
			// Skip this failure if it's a parent (has failing children)
			if parentWorkflows[report.RootWorkflow.WorkflowID] {
				continue
			}
			leafFailures = append(leafFailures, report)
			// Apply limit after filtering
			if f.opts.Limit > 0 && len(leafFailures) >= f.opts.Limit {
				break
			}
		}
		failures = leafFailures
	}

	// Apply grouping if requested
	if f.opts.GroupBy != "" && f.opts.GroupBy != "none" {
		groups := f.groupFailures(failures)
		return &FailuresResult{
			Groups:     groups,
			TotalCount: len(failures),
			Query:      query,
			GroupedBy:  f.opts.GroupBy,
		}, nil
	}

	return &FailuresResult{
		Failures:   failures,
		TotalCount: len(failures),
		Query:      query,
	}, nil
}

// groupFailures groups failures by the specified field.
func (f *FailuresFinder) groupFailures(failures []FailureReport) []FailureGroup {
	groupMap := make(map[string]*FailureGroup)

	for i := range failures {
		report := &failures[i]
		key := f.getGroupKey(report)

		if group, ok := groupMap[key]; ok {
			group.Count++
			// Update time bounds
			if report.Timestamp != nil {
				if group.FirstSeen == nil || report.Timestamp.Before(*group.FirstSeen) {
					group.FirstSeen = report.Timestamp
				}
				if group.LastSeen == nil || report.Timestamp.After(*group.LastSeen) {
					group.LastSeen = report.Timestamp
				}
			}
		} else {
			groupMap[key] = &FailureGroup{
				Key:       key,
				Count:     1,
				Sample:    report,
				FirstSeen: report.Timestamp,
				LastSeen:  report.Timestamp,
			}
		}
	}

	// Convert to slice and calculate percentages
	total := len(failures)
	groups := make([]FailureGroup, 0, len(groupMap))
	for _, group := range groupMap {
		group.Percentage = float64(group.Count) / float64(total) * 100
		groups = append(groups, *group)
	}

	// Sort by count descending
	for i := 0; i < len(groups); i++ {
		for j := i + 1; j < len(groups); j++ {
			if groups[j].Count > groups[i].Count {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
	}

	return groups
}

// getGroupKey returns the grouping key for a failure report.
func (f *FailuresFinder) getGroupKey(report *FailureReport) string {
	switch f.opts.GroupBy {
	case "type":
		// Extract workflow type from the leaf failure or root workflow
		if report.LeafFailure != nil {
			return report.LeafFailure.WorkflowID
		}
		return report.RootWorkflow.WorkflowID
	case "namespace":
		if report.LeafFailure != nil {
			return report.LeafFailure.Namespace
		}
		return report.RootWorkflow.Namespace
	case "status":
		return report.Status
	case "error":
		// Use the root cause as the key (compacted for grouping)
		return CompactError(report.RootCause)
	default:
		return "unknown"
	}
}

func (f *FailuresFinder) buildQuery() string {
	var parts []string

	// Add status filter
	statuses := f.opts.Statuses
	if len(statuses) == 0 {
		statuses = []enums.WorkflowExecutionStatus{
			enums.WORKFLOW_EXECUTION_STATUS_FAILED,
			enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
		}
	}

	var statusParts []string
	for _, s := range statuses {
		statusParts = append(statusParts, fmt.Sprintf("ExecutionStatus = %q", WorkflowStatusFromEnum(s)))
	}
	if len(statusParts) > 0 {
		parts = append(parts, "("+strings.Join(statusParts, " OR ")+")")
	}

	// Add time filter
	if f.opts.Since > 0 {
		cutoff := time.Now().Add(-f.opts.Since)
		parts = append(parts, fmt.Sprintf("CloseTime > %q", cutoff.Format(time.RFC3339)))
	}

	return strings.Join(parts, " AND ")
}

func (f *FailuresFinder) getWorkflowFailure(ctx context.Context, cl client.Client, workflowID, runID string) string {
	// Get just the close event to find the failure message
	iter := cl.GetWorkflowHistory(ctx, workflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT)
	if iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return "failed to get close event"
		}

		switch event.GetEventType() {
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
			attrs := event.GetWorkflowExecutionFailedEventAttributes()
			return fmt.Sprintf("WorkflowFailed: %s", attrs.GetFailure().GetMessage())
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
			return "Timeout: workflow execution timed out"
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
			return "Canceled: workflow was canceled"
		case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
			return "Terminated: workflow was terminated"
		}
	}
	return ""
}

// stateMachine processes workflow history events to extract state.
type stateMachine struct {
	runID          string
	workflowType   string
	status         enums.WorkflowExecutionStatus
	startTime      *time.Time
	closeTime      *time.Time
	failureMessage string

	// Track events for lookups
	events []*history.HistoryEvent

	// Pending and failed activities
	pendingActivities map[int64]*activityState
	failedActivities  []*activityState

	// Pending and failed child workflows
	pendingChildren map[int64]*childWorkflowState
	failedChildren  []*childWorkflowState

	// Pending and failed Nexus operations
	pendingNexusOps map[int64]*nexusOperationState
	failedNexusOps  []*nexusOperationState
}

type activityState struct {
	activityID     string
	activityType   string
	scheduledTime  *time.Time
	startedTime    *time.Time
	failureTime    *time.Time
	failureMessage string
	attempt        int32
}

type childWorkflowState struct {
	namespace      string
	workflowID     string
	runID          string
	workflowType   string
	failureMessage string
}

// nexusOperationState tracks a Nexus operation's state.
type nexusOperationState struct {
	endpoint       string
	service        string
	operation      string
	namespace      string // Target namespace (from links)
	workflowID     string // Target workflow ID (from links)
	runID          string // Target workflow run ID (from links)
	failureMessage string
}

func newStateMachine() *stateMachine {
	return &stateMachine{
		status:            enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED,
		pendingActivities: make(map[int64]*activityState),
		pendingChildren:   make(map[int64]*childWorkflowState),
		pendingNexusOps:   make(map[int64]*nexusOperationState),
	}
}

func (sm *stateMachine) processEvent(event *history.HistoryEvent) {
	sm.events = append(sm.events, event)

	switch event.GetEventType() {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
		attrs := event.GetWorkflowExecutionStartedEventAttributes()
		sm.status = enums.WORKFLOW_EXECUTION_STATUS_RUNNING
		sm.workflowType = attrs.GetWorkflowType().GetName()
		sm.runID = attrs.GetOriginalExecutionRunId()
		t := event.GetEventTime().AsTime()
		sm.startTime = &t

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		sm.status = enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
		t := event.GetEventTime().AsTime()
		sm.closeTime = &t

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		attrs := event.GetWorkflowExecutionFailedEventAttributes()
		sm.status = enums.WORKFLOW_EXECUTION_STATUS_FAILED
		sm.failureMessage = attrs.GetFailure().GetMessage()
		t := event.GetEventTime().AsTime()
		sm.closeTime = &t

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		sm.status = enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT
		t := event.GetEventTime().AsTime()
		sm.closeTime = &t

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		sm.status = enums.WORKFLOW_EXECUTION_STATUS_CANCELED
		t := event.GetEventTime().AsTime()
		sm.closeTime = &t

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
		sm.status = enums.WORKFLOW_EXECUTION_STATUS_TERMINATED
		t := event.GetEventTime().AsTime()
		sm.closeTime = &t

	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CONTINUED_AS_NEW:
		sm.status = enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW
		t := event.GetEventTime().AsTime()
		sm.closeTime = &t

	// Activity events
	case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
		attrs := event.GetActivityTaskScheduledEventAttributes()
		t := event.GetEventTime().AsTime()
		sm.pendingActivities[event.EventId] = &activityState{
			activityID:    attrs.GetActivityId(),
			activityType:  attrs.GetActivityType().GetName(),
			scheduledTime: &t,
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
		attrs := event.GetActivityTaskStartedEventAttributes()
		if act, ok := sm.pendingActivities[attrs.GetScheduledEventId()]; ok {
			t := event.GetEventTime().AsTime()
			act.startedTime = &t
			act.attempt = attrs.GetAttempt()
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
		attrs := event.GetActivityTaskCompletedEventAttributes()
		delete(sm.pendingActivities, attrs.GetScheduledEventId())

	case enums.EVENT_TYPE_ACTIVITY_TASK_FAILED:
		attrs := event.GetActivityTaskFailedEventAttributes()
		if act, ok := sm.pendingActivities[attrs.GetScheduledEventId()]; ok {
			t := event.GetEventTime().AsTime()
			act.failureTime = &t
			act.failureMessage = attrs.GetFailure().GetMessage()
			delete(sm.pendingActivities, attrs.GetScheduledEventId())
			sm.failedActivities = append(sm.failedActivities, act)
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
		attrs := event.GetActivityTaskTimedOutEventAttributes()
		if act, ok := sm.pendingActivities[attrs.GetScheduledEventId()]; ok {
			t := event.GetEventTime().AsTime()
			act.failureTime = &t
			act.failureMessage = "activity timed out"
			if attrs.GetFailure() != nil {
				act.failureMessage = attrs.GetFailure().GetMessage()
			}
			delete(sm.pendingActivities, attrs.GetScheduledEventId())
			sm.failedActivities = append(sm.failedActivities, act)
		}

	case enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
		attrs := event.GetActivityTaskCanceledEventAttributes()
		delete(sm.pendingActivities, attrs.GetScheduledEventId())

	// Child workflow events
	case enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED:
		attrs := event.GetStartChildWorkflowExecutionInitiatedEventAttributes()
		sm.pendingChildren[event.EventId] = &childWorkflowState{
			namespace:    attrs.GetNamespace(),
			workflowID:   attrs.GetWorkflowId(),
			workflowType: attrs.GetWorkflowType().GetName(),
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
		attrs := event.GetChildWorkflowExecutionStartedEventAttributes()
		if child, ok := sm.pendingChildren[attrs.GetInitiatedEventId()]; ok {
			child.runID = attrs.GetWorkflowExecution().GetRunId()
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
		attrs := event.GetChildWorkflowExecutionCompletedEventAttributes()
		delete(sm.pendingChildren, attrs.GetInitiatedEventId())

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED:
		attrs := event.GetChildWorkflowExecutionFailedEventAttributes()
		if child, ok := sm.pendingChildren[attrs.GetInitiatedEventId()]; ok {
			child.failureMessage = attrs.GetFailure().GetMessage()
			child.runID = attrs.GetWorkflowExecution().GetRunId()
			delete(sm.pendingChildren, attrs.GetInitiatedEventId())
			sm.failedChildren = append(sm.failedChildren, child)
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
		attrs := event.GetChildWorkflowExecutionTimedOutEventAttributes()
		if child, ok := sm.pendingChildren[attrs.GetInitiatedEventId()]; ok {
			child.failureMessage = "child workflow timed out"
			child.runID = attrs.GetWorkflowExecution().GetRunId()
			delete(sm.pendingChildren, attrs.GetInitiatedEventId())
			sm.failedChildren = append(sm.failedChildren, child)
		}

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
		attrs := event.GetChildWorkflowExecutionCanceledEventAttributes()
		delete(sm.pendingChildren, attrs.GetInitiatedEventId())

	case enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TERMINATED:
		attrs := event.GetChildWorkflowExecutionTerminatedEventAttributes()
		if child, ok := sm.pendingChildren[attrs.GetInitiatedEventId()]; ok {
			child.failureMessage = "child workflow terminated"
			child.runID = attrs.GetWorkflowExecution().GetRunId()
			delete(sm.pendingChildren, attrs.GetInitiatedEventId())
			sm.failedChildren = append(sm.failedChildren, child)
		}

	// Nexus operation events
	case enums.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED:
		attrs := event.GetNexusOperationScheduledEventAttributes()
		sm.pendingNexusOps[event.EventId] = &nexusOperationState{
			endpoint:  attrs.GetEndpoint(),
			service:   attrs.GetService(),
			operation: attrs.GetOperation(),
		}

	case enums.EVENT_TYPE_NEXUS_OPERATION_STARTED:
		attrs := event.GetNexusOperationStartedEventAttributes()
		if nexusOp, ok := sm.pendingNexusOps[attrs.GetScheduledEventId()]; ok {
			// Extract target workflow from links
			for _, link := range event.GetLinks() {
				if wfEvent := link.GetWorkflowEvent(); wfEvent != nil {
					nexusOp.namespace = wfEvent.GetNamespace()
					nexusOp.workflowID = wfEvent.GetWorkflowId()
					nexusOp.runID = wfEvent.GetRunId()
					break
				}
			}
		}

	case enums.EVENT_TYPE_NEXUS_OPERATION_COMPLETED:
		attrs := event.GetNexusOperationCompletedEventAttributes()
		delete(sm.pendingNexusOps, attrs.GetScheduledEventId())

	case enums.EVENT_TYPE_NEXUS_OPERATION_FAILED:
		attrs := event.GetNexusOperationFailedEventAttributes()
		if nexusOp, ok := sm.pendingNexusOps[attrs.GetScheduledEventId()]; ok {
			nexusOp.failureMessage = attrs.GetFailure().GetMessage()
			// Extract cause if available
			if cause := attrs.GetFailure().GetCause(); cause != nil {
				nexusOp.failureMessage = cause.GetMessage()
			}
			delete(sm.pendingNexusOps, attrs.GetScheduledEventId())
			sm.failedNexusOps = append(sm.failedNexusOps, nexusOp)
		}

	case enums.EVENT_TYPE_NEXUS_OPERATION_CANCELED:
		attrs := event.GetNexusOperationCanceledEventAttributes()
		delete(sm.pendingNexusOps, attrs.GetScheduledEventId())

	case enums.EVENT_TYPE_NEXUS_OPERATION_TIMED_OUT:
		attrs := event.GetNexusOperationTimedOutEventAttributes()
		if nexusOp, ok := sm.pendingNexusOps[attrs.GetScheduledEventId()]; ok {
			nexusOp.failureMessage = "nexus operation timed out"
			if attrs.GetFailure() != nil {
				nexusOp.failureMessage = attrs.GetFailure().GetMessage()
			}
			delete(sm.pendingNexusOps, attrs.GetScheduledEventId())
			sm.failedNexusOps = append(sm.failedNexusOps, nexusOp)
		}
	}
}
