package temporalcli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/printer"
	activitypb "go.temporal.io/api/activity/v1"
	callbackpb "go.temporal.io/api/callback/v1"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

func workflowEventLink() *commonpb.Link {
	return &commonpb.Link{
		Variant: &commonpb.Link_WorkflowEvent_{
			WorkflowEvent: &commonpb.Link_WorkflowEvent{
				Namespace:  "ns",
				WorkflowId: "wf-id",
				RunId:      "run-id",
				Reference: &commonpb.Link_WorkflowEvent_EventRef{
					EventRef: &commonpb.Link_WorkflowEvent_EventReference{
						EventId:   1,
						EventType: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
					},
				},
			},
		},
	}
}

// TestNexusLinkStrings verifies that nexusLinkStrings converts workflow-event,
// nexus-operation, and activity common links into their Nexus link URLs using
// the converters that now live in go.temporal.io/api/temporalnexus, and skips
// variants that have no Nexus equivalent.
func TestNexusLinkStrings(t *testing.T) {
	activityLink := &commonpb.Link{
		Variant: &commonpb.Link_Activity_{
			Activity: &commonpb.Link_Activity{
				Namespace:  "ns",
				ActivityId: "act-id",
				RunId:      "run-id",
			},
		},
	}
	nexusOperationLink := &commonpb.Link{
		Variant: &commonpb.Link_NexusOperation_{
			NexusOperation: &commonpb.Link_NexusOperation{
				Namespace:   "ns",
				OperationId: "op-id",
				RunId:       "run-id",
			},
		},
	}
	// A link with an unset (unsupported) variant, which must be skipped.
	otherLink := &commonpb.Link{}

	t.Run("nil and empty", func(t *testing.T) {
		require.Nil(t, nexusLinkStrings(nil))
		require.Nil(t, nexusLinkStrings([]*commonpb.Link{}))
	})

	t.Run("workflow event link", func(t *testing.T) {
		got := nexusLinkStrings([]*commonpb.Link{workflowEventLink()})
		require.Equal(t, []string{
			"temporal:///namespaces/ns/workflows/wf-id/run-id/history?eventID=1&eventType=WorkflowExecutionStarted&referenceType=EventReference",
		}, got)
	})

	t.Run("activity and nexus operation links", func(t *testing.T) {
		got := nexusLinkStrings([]*commonpb.Link{activityLink, nexusOperationLink})
		require.Len(t, got, 2)
		require.Contains(t, got[0], "temporal://")
		require.Contains(t, got[0], "act-id")
		require.Contains(t, got[1], "temporal://")
		require.Contains(t, got[1], "op-id")
	})

	t.Run("skips unsupported variants and preserves order", func(t *testing.T) {
		got := nexusLinkStrings([]*commonpb.Link{otherLink, workflowEventLink(), otherLink})
		require.Len(t, got, 1)
		require.Contains(t, got[0], "wf-id")
	})
}

// TestPrintActivityDescription_LinksAndCallbacks verifies that activity describe
// renders the activity's Nexus links and its callbacks (including callback URL,
// links, and trigger) like workflow describe does.
func TestPrintActivityDescription_LinksAndCallbacks(t *testing.T) {
	var buf bytes.Buffer
	cctx := &CommandContext{Printer: &printer.Printer{Output: &buf}}

	resp := &workflowservice.DescribeActivityExecutionResponse{
		Info: &activitypb.ActivityExecutionInfo{
			ActivityId:   "act-1",
			RunId:        "run-1",
			ActivityType: &commonpb.ActivityType{Name: "MyActivity"},
			Status:       enumspb.ACTIVITY_EXECUTION_STATUS_RUNNING,
			TaskQueue:    "tq",
			Attempt:      1,
			Links:        []*commonpb.Link{workflowEventLink()},
		},
		Callbacks: []*activitypb.CallbackInfo{
			{
				Trigger: &activitypb.CallbackInfo_Trigger{
					Variant: &activitypb.CallbackInfo_Trigger_ActivityClosed{
						ActivityClosed: &activitypb.CallbackInfo_ActivityClosed{},
					},
				},
				Info: &callbackpb.CallbackInfo{
					Callback: &commonpb.Callback{
						Variant: &commonpb.Callback_Nexus_{
							Nexus: &commonpb.Callback_Nexus{Url: "http://callback.example/cb"},
						},
						Links: []*commonpb.Link{workflowEventLink()},
					},
					State:   enumspb.CALLBACK_STATE_SUCCEEDED,
					Attempt: 2,
				},
			},
		},
	}

	require.NoError(t, printActivityDescription(cctx, resp))
	out := buf.String()

	// Execution info.
	require.Contains(t, out, "act-1")
	require.Contains(t, out, "MyActivity")
	// Top-level activity Links section.
	require.Contains(t, out, "Links: 1")
	require.Contains(t, out, "temporal:///namespaces/ns/workflows/wf-id/run-id/history")
	// Callbacks section with URL, trigger, link, and state.
	require.Contains(t, out, "Callbacks: 1")
	require.Contains(t, out, "http://callback.example/cb")
	require.Contains(t, out, "ActivityClosed")
	require.Contains(t, out, "Succeeded")
}

func TestPrintNexusOperationDescription_Links(t *testing.T) {
	var buf bytes.Buffer
	cctx := &CommandContext{Printer: &printer.Printer{Output: &buf}}

	desc := &client.NexusOperationExecutionDescription{
		RawInfo: &nexuspb.NexusOperationExecutionInfo{
			Links: []*commonpb.Link{workflowEventLink()},
		},
	}
	desc.OperationID = "op-1"
	desc.OperationRunID = "run-1"

	require.NoError(t, printNexusOperationDescription(cctx, desc))
	out := buf.String()
	require.Contains(t, out, "op-1")
	require.Contains(t, out, "Links: 1")
	require.Contains(t, out, "temporal:///namespaces/ns/workflows/wf-id/run-id/history")
}
