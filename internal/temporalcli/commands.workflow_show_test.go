package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
)

func TestSystemNexusOpDisplayName(t *testing.T) {
	const op = "SignalWithStartWorkflowExecution"

	// priorScheduled returns an iter that has already seen a __temporal_system scheduled event at ID 5.
	priorScheduled := func() *structuredHistoryIter {
		s := &structuredHistoryIter{}
		s.systemNexusOpDisplayName(&historypb.HistoryEvent{
			EventId:   5,
			EventType: enums.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
			Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
				NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
					Endpoint: temporalSystemNexusEndpoint, Operation: op,
				},
			},
		})
		return s
	}

	tests := []struct {
		name      string
		setup     func() *structuredHistoryIter
		event     *historypb.HistoryEvent
		wantName  string // expected suffix after op, or exact value if wantEmpty
		wantEmpty bool
	}{
		{
			name:  "scheduled system endpoint records op and returns name",
			setup: func() *structuredHistoryIter { return &structuredHistoryIter{} },
			event: &historypb.HistoryEvent{
				EventId:   5,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
				Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
					NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
						Endpoint: temporalSystemNexusEndpoint, Operation: op,
					},
				},
			},
			wantName: op + "Scheduled",
		},
		{
			name:  "scheduled non-system endpoint returns empty",
			setup: func() *structuredHistoryIter { return &structuredHistoryIter{} },
			event: &historypb.HistoryEvent{
				EventId:   7,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
				Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
					NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
						Endpoint: "some-other-endpoint", Operation: op,
					},
				},
			},
			wantEmpty: true,
		},
		{
			name:  "started after system scheduled",
			setup: priorScheduled,
			event: &historypb.HistoryEvent{
				EventId:   6,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_STARTED,
				Attributes: &historypb.HistoryEvent_NexusOperationStartedEventAttributes{
					NexusOperationStartedEventAttributes: &historypb.NexusOperationStartedEventAttributes{ScheduledEventId: 5},
				},
			},
			wantName: op + "Started",
		},
		{
			name:  "completed after system scheduled",
			setup: priorScheduled,
			event: &historypb.HistoryEvent{
				EventId:   6,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_COMPLETED,
				Attributes: &historypb.HistoryEvent_NexusOperationCompletedEventAttributes{
					NexusOperationCompletedEventAttributes: &historypb.NexusOperationCompletedEventAttributes{ScheduledEventId: 5},
				},
			},
			wantName: op + "Completed",
		},
		{
			name:  "completed with no prior scheduled returns empty",
			setup: func() *structuredHistoryIter { return &structuredHistoryIter{} },
			event: &historypb.HistoryEvent{
				EventId:   6,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_COMPLETED,
				Attributes: &historypb.HistoryEvent_NexusOperationCompletedEventAttributes{
					NexusOperationCompletedEventAttributes: &historypb.NexusOperationCompletedEventAttributes{ScheduledEventId: 5},
				},
			},
			wantEmpty: true,
		},
		{
			name:  "failed after system scheduled",
			setup: priorScheduled,
			event: &historypb.HistoryEvent{
				EventId:   6,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_FAILED,
				Attributes: &historypb.HistoryEvent_NexusOperationFailedEventAttributes{
					NexusOperationFailedEventAttributes: &historypb.NexusOperationFailedEventAttributes{ScheduledEventId: 5},
				},
			},
			wantName: op + "Failed",
		},
		{
			name:  "timed out after system scheduled",
			setup: priorScheduled,
			event: &historypb.HistoryEvent{
				EventId:   6,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_TIMED_OUT,
				Attributes: &historypb.HistoryEvent_NexusOperationTimedOutEventAttributes{
					NexusOperationTimedOutEventAttributes: &historypb.NexusOperationTimedOutEventAttributes{ScheduledEventId: 5},
				},
			},
			wantName: op + "TimedOut",
		},
		{
			name:  "canceled after system scheduled",
			setup: priorScheduled,
			event: &historypb.HistoryEvent{
				EventId:   6,
				EventType: enums.EVENT_TYPE_NEXUS_OPERATION_CANCELED,
				Attributes: &historypb.HistoryEvent_NexusOperationCanceledEventAttributes{
					NexusOperationCanceledEventAttributes: &historypb.NexusOperationCanceledEventAttributes{ScheduledEventId: 5},
				},
			},
			wantName: op + "Canceled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name := tc.setup().systemNexusOpDisplayName(tc.event)
			if tc.wantEmpty {
				require.Empty(t, name)
			} else {
				require.Equal(t, tc.wantName, name)
			}
		})
	}
}
