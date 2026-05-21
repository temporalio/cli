package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
)

func TestSystemNexusOpDisplayName(t *testing.T) {
	const op = "SignalWithStartWorkflowExecution" // present in the global systemNexusOps registry
	const futureOp = "FutureUnregisteredOperation" // intentionally NOT in the global registry

	newScheduled := func(eventID int64, endpoint, operation string) *historypb.HistoryEvent {
		return &historypb.HistoryEvent{
			EventId:   eventID,
			EventType: enums.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
			Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
				NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
					Endpoint: endpoint, Operation: operation,
				},
			},
		}
	}
	// terminalEvent returns a NexusOperation* event of the given type referencing scheduledEventID.
	// Returns nil for unsupported event types.
	terminalEvent := func(eventType enums.EventType, scheduledEventID int64) *historypb.HistoryEvent {
		e := &historypb.HistoryEvent{EventId: scheduledEventID + 1, EventType: eventType}
		switch eventType {
		case enums.EVENT_TYPE_NEXUS_OPERATION_STARTED:
			e.Attributes = &historypb.HistoryEvent_NexusOperationStartedEventAttributes{
				NexusOperationStartedEventAttributes: &historypb.NexusOperationStartedEventAttributes{ScheduledEventId: scheduledEventID},
			}
		case enums.EVENT_TYPE_NEXUS_OPERATION_COMPLETED:
			e.Attributes = &historypb.HistoryEvent_NexusOperationCompletedEventAttributes{
				NexusOperationCompletedEventAttributes: &historypb.NexusOperationCompletedEventAttributes{ScheduledEventId: scheduledEventID},
			}
		case enums.EVENT_TYPE_NEXUS_OPERATION_FAILED:
			e.Attributes = &historypb.HistoryEvent_NexusOperationFailedEventAttributes{
				NexusOperationFailedEventAttributes: &historypb.NexusOperationFailedEventAttributes{ScheduledEventId: scheduledEventID},
			}
		case enums.EVENT_TYPE_NEXUS_OPERATION_TIMED_OUT:
			e.Attributes = &historypb.HistoryEvent_NexusOperationTimedOutEventAttributes{
				NexusOperationTimedOutEventAttributes: &historypb.NexusOperationTimedOutEventAttributes{ScheduledEventId: scheduledEventID},
			}
		case enums.EVENT_TYPE_NEXUS_OPERATION_CANCELED:
			e.Attributes = &historypb.HistoryEvent_NexusOperationCanceledEventAttributes{
				NexusOperationCanceledEventAttributes: &historypb.NexusOperationCanceledEventAttributes{ScheduledEventId: scheduledEventID},
			}
		}
		return e
	}

	// priorScheduled returns a setup that creates an iter and pre-processes the given Scheduled
	// event. In forward traversal this matches the natural event order; in reverse traversal it
	// simulates the iter having pre-scanned the buffered history.
	priorScheduled := func(scheduled *historypb.HistoryEvent) func(reverse bool) *structuredHistoryIter {
		return func(reverse bool) *structuredHistoryIter {
			s := &structuredHistoryIter{reverse: reverse}
			s.systemNexusOpDisplayName(scheduled)
			return s
		}
	}
	empty := func(reverse bool) *structuredHistoryIter {
		return &structuredHistoryIter{reverse: reverse}
	}

	schedSystem := newScheduled(5, temporalSystemNexusEndpoint, op)
	schedSystemFutureOp := newScheduled(5, temporalSystemNexusEndpoint, futureOp)
	schedNonSystem := newScheduled(7, "some-other-endpoint", op)

	tests := []struct {
		name      string
		setup     func(reverse bool) *structuredHistoryIter
		event     *historypb.HistoryEvent
		wantName  string
		wantEmpty bool
	}{
		// Scheduled cases
		{
			name:     "scheduled system endpoint records op and returns name",
			setup:    empty,
			event:    schedSystem,
			wantName: op + "Scheduled",
		},
		{
			name:      "scheduled non-system endpoint returns empty",
			setup:     empty,
			event:     schedNonSystem,
			wantEmpty: true,
		},
		{
			// Display name does not consult the global systemNexusOps registry; any op on the
			// __temporal_system endpoint produces a display name. (Payload unwrap-and-inject is
			// the layer that requires registry membership; see TestUnwrapAndInjectRequest_*.)
			name:     "scheduled system endpoint with op not in registry still returns name",
			setup:    empty,
			event:    schedSystemFutureOp,
			wantName: futureOp + "Scheduled",
		},
		// Terminal cases with a prior Scheduled in the instance map
		{
			name:     "started after system scheduled",
			setup:    priorScheduled(schedSystem),
			event:    terminalEvent(enums.EVENT_TYPE_NEXUS_OPERATION_STARTED, 5),
			wantName: op + "Started",
		},
		{
			// Confirms terminal events use the iter's instance map (which captured the unregistered
			// op from its Scheduled event), regardless of global-registry membership.
			name:     "started after system scheduled with op not in registry",
			setup:    priorScheduled(schedSystemFutureOp),
			event:    terminalEvent(enums.EVENT_TYPE_NEXUS_OPERATION_STARTED, 5),
			wantName: futureOp + "Started",
		},
		{
			name:     "completed after system scheduled",
			setup:    priorScheduled(schedSystem),
			event:    terminalEvent(enums.EVENT_TYPE_NEXUS_OPERATION_COMPLETED, 5),
			wantName: op + "Completed",
		},
		{
			name:      "completed with no prior scheduled returns empty",
			setup:     empty,
			event:     terminalEvent(enums.EVENT_TYPE_NEXUS_OPERATION_COMPLETED, 5),
			wantEmpty: true,
		},
		{
			name:     "failed after system scheduled",
			setup:    priorScheduled(schedSystem),
			event:    terminalEvent(enums.EVENT_TYPE_NEXUS_OPERATION_FAILED, 5),
			wantName: op + "Failed",
		},
		{
			name:     "timed out after system scheduled",
			setup:    priorScheduled(schedSystem),
			event:    terminalEvent(enums.EVENT_TYPE_NEXUS_OPERATION_TIMED_OUT, 5),
			wantName: op + "TimedOut",
		},
		{
			name:     "canceled after system scheduled",
			setup:    priorScheduled(schedSystem),
			event:    terminalEvent(enums.EVENT_TYPE_NEXUS_OPERATION_CANCELED, 5),
			wantName: op + "Canceled",
		},
	}

	// The function is direction-agnostic: given the same prior state, it returns the same
	// display name whether iter.reverse is true or false. Running each case in both modes
	// pins that contract so a future change can't silently introduce direction-sensitive
	// branching without breaking these tests.
	for _, reverse := range []bool{false, true} {
		modeName := "forward"
		if reverse {
			modeName = "reverse"
		}
		t.Run(modeName, func(t *testing.T) {
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					iter := tc.setup(reverse)
					require.Equal(t, reverse, iter.reverse, "test setup must mark iter reverse=%v", reverse)
					name := iter.systemNexusOpDisplayName(tc.event)
					if tc.wantEmpty {
						require.Empty(t, name)
					} else {
						require.Equal(t, tc.wantName, name)
					}
				})
			}
		})
	}
}
