package temporalcli

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/api/proxy"
	"go.temporal.io/api/temporalproto"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/proto"
)

func markedSystemPayload(t *testing.T, msg proto.Message) *commonpb.Payload {
	t.Helper()
	data, err := proto.Marshal(msg)
	require.NoError(t, err)
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			"encoding":                     []byte("binary/protobuf"),
			"messageType":                  []byte(msg.ProtoReflect().Descriptor().FullName()),
			proxy.SystemPayloadMetadataKey: []byte("true"),
		},
		Data: data,
	}
}

func TestUnwrapAndInject_NilPayloadIsNoOp(t *testing.T) {
	fields := map[string]any{}
	require.NoError(t, (&structuredHistoryIter{}).unwrapAndInject(
		nil, fields, "unwrappedInput", temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields)
}

func TestUnwrapAndInject_RejectsInvalidEnvelope(t *testing.T) {
	valid := markedSystemPayload(t, &workflowservice.SignalWithStartWorkflowExecutionRequest{})
	tests := []struct {
		name    string
		mutate  func(*commonpb.Payload)
		wantErr string
	}{
		{
			name: "missing marker",
			mutate: func(payload *commonpb.Payload) {
				delete(payload.Metadata, proxy.SystemPayloadMetadataKey)
			},
			wantErr: "missing the __temporal_system_payload marker",
		},
		{
			name: "wrong encoding",
			mutate: func(payload *commonpb.Payload) {
				payload.Metadata["encoding"] = []byte("json/protobuf")
			},
			wantErr: "must be encoded as binary/protobuf",
		},
		{
			name: "missing message type",
			mutate: func(payload *commonpb.Payload) {
				delete(payload.Metadata, "messageType")
			},
			wantErr: "missing messageType metadata",
		},
		{
			name: "unknown message type",
			mutate: func(payload *commonpb.Payload) {
				payload.Metadata["messageType"] = []byte("temporal.api.unknown.v1.Message")
			},
			wantErr: "references unknown message type",
		},
		{
			name: "invalid protobuf",
			mutate: func(payload *commonpb.Payload) {
				payload.Data = []byte{0xff, 0xff, 0xff}
			},
			wantErr: "failed unmarshaling system nexus payload",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := proto.Clone(valid).(*commonpb.Payload)
			tc.mutate(payload)
			err := (&structuredHistoryIter{}).unwrapAndInject(
				payload, map[string]any{}, "unwrappedInput", temporalproto.CustomJSONMarshalOptions{})
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestUnwrapAndInject_UsesEnvelopeMessageType(t *testing.T) {
	payload := markedSystemPayload(t, &workflowservice.SignalWithStartWorkflowExecutionRequest{
		Namespace:  "ns",
		WorkflowId: "wf",
		SignalName: "signal",
	})
	fields := map[string]any{}
	require.NoError(t, (&structuredHistoryIter{}).unwrapAndInject(
		payload, fields, "unwrappedInput", temporalproto.CustomJSONMarshalOptions{}))

	unwrapped, ok := fields["unwrappedInput"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "ns", unwrapped["namespace"])
	require.Equal(t, "wf", unwrapped["workflowId"])
	require.Equal(t, "signal", unwrapped["signalName"])
}

func TestSystemNexusPayload_DecodedByVisitorAndUnwrappedLocally(t *testing.T) {
	req := &workflowservice.SignalWithStartWorkflowExecutionRequest{
		WorkflowId: "wf",
		Input: &commonpb.Payloads{Payloads: []*commonpb.Payload{
			{Data: []byte("encoded:first")},
			{Data: []byte("encoded:second")},
		}},
	}
	payload := markedSystemPayload(t, req)
	attrs := &historypb.NexusOperationScheduledEventAttributes{Input: payload}

	decodeRequests := 0
	decodedPayloads := 0
	err := proxy.VisitPayloads(context.Background(), attrs, proxy.VisitPayloadsOptions{
		Visitor: func(_ *proxy.VisitPayloadsContext, payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
			decodeRequests++
			decodedPayloads += len(payloads)
			decoded := make([]*commonpb.Payload, len(payloads))
			for i, payload := range payloads {
				require.True(t, bytes.HasPrefix(payload.Data, []byte("encoded:")))
				decoded[i] = proto.Clone(payload).(*commonpb.Payload)
				decoded[i].Data = bytes.TrimPrefix(payload.Data, []byte("encoded:"))
			}
			return decoded, nil
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, decodeRequests)
	require.Equal(t, 2, decodedPayloads)

	fields := map[string]any{}
	require.NoError(t, (&structuredHistoryIter{}).unwrapAndInject(
		attrs.Input, fields, "unwrappedInput", temporalproto.CustomJSONMarshalOptions{}))
	require.Equal(t, 1, decodeRequests, "local rendering must not revisit the codec")
	require.Contains(t, fields, "unwrappedInput")
}

func TestInjectSystemNexusUnwrapped_ScheduledUsesMessageTypeForAnyOperation(t *testing.T) {
	event := &historypb.HistoryEvent{
		EventId:   5,
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
		Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
			NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
				Endpoint:  temporalSystemNexusEndpoint,
				Operation: "FutureSystemOperation",
				Input: markedSystemPayload(t, &workflowservice.SignalWithStartWorkflowExecutionRequest{
					WorkflowId: "wf",
				}),
			},
		},
	}
	fields := map[string]any{}
	require.NoError(t, (&structuredHistoryIter{}).injectSystemNexusUnwrapped(
		event, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Contains(t, fields, "unwrappedInput")
}

func TestInjectSystemNexusUnwrapped_UnknownEndpointIsNoOp(t *testing.T) {
	event := &historypb.HistoryEvent{
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
		Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
			NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
				Endpoint: "user-endpoint",
				Input:    markedSystemPayload(t, &workflowservice.SignalWithStartWorkflowExecutionRequest{}),
			},
		},
	}
	fields := map[string]any{}
	require.NoError(t, (&structuredHistoryIter{}).injectSystemNexusUnwrapped(
		event, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields)
}

func TestInjectSystemNexusUnwrapped_CompletedUsesPriorScheduledEvent(t *testing.T) {
	event := &historypb.HistoryEvent{
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_COMPLETED,
		Attributes: &historypb.HistoryEvent_NexusOperationCompletedEventAttributes{
			NexusOperationCompletedEventAttributes: &historypb.NexusOperationCompletedEventAttributes{
				ScheduledEventId: 5,
				Result: markedSystemPayload(t, &workflowservice.SignalWithStartWorkflowExecutionResponse{
					RunId: "run-id",
				}),
			},
		},
	}
	iter := &structuredHistoryIter{systemNexusOps: map[int64]string{5: "FutureSystemOperation"}}
	fields := map[string]any{}
	require.NoError(t, iter.injectSystemNexusUnwrapped(
		event, fields, temporalproto.CustomJSONMarshalOptions{}))
	unwrapped, ok := fields["unwrappedResult"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "run-id", unwrapped["runId"])
}

func TestInjectSystemNexusUnwrapped_CompletedWithoutPriorScheduledIsNoOp(t *testing.T) {
	event := &historypb.HistoryEvent{
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_COMPLETED,
		Attributes: &historypb.HistoryEvent_NexusOperationCompletedEventAttributes{
			NexusOperationCompletedEventAttributes: &historypb.NexusOperationCompletedEventAttributes{
				ScheduledEventId: 5,
				Result:           markedSystemPayload(t, &workflowservice.SignalWithStartWorkflowExecutionResponse{}),
			},
		},
	}
	fields := map[string]any{}
	require.NoError(t, (&structuredHistoryIter{}).injectSystemNexusUnwrapped(
		event, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields)
}
