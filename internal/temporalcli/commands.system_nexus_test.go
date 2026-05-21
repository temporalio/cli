package temporalcli

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/api/temporalproto"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/proto"
)

// markingCodec is a test [converter.PayloadCodec] that prefixes every payload's
// data with "decoded:" on Decode and tracks how many payloads it saw. It is used
// to verify that the codec is actually invoked on payloads nested inside opaque
// system Nexus operation bytes.
type markingCodec struct {
	decodeCalls int
}

func (c *markingCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return payloads, nil
}

func (c *markingCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	out := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		c.decodeCalls++
		out[i] = &commonpb.Payload{
			Metadata: p.Metadata,
			Data:     append([]byte("decoded:"), p.Data...),
		}
	}
	return out, nil
}

// failingCodec always returns an error from Decode; used to verify error propagation.
type failingCodec struct{}

func (failingCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return payloads, nil
}

func (failingCodec) Decode(_ []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return nil, fmt.Errorf("codec decode failure for testing")
}

func signalWithStartRequestPayload(t *testing.T, req *workflowservice.SignalWithStartWorkflowExecutionRequest) *commonpb.Payload {
	t.Helper()
	data, err := proto.Marshal(req)
	require.NoError(t, err)
	return &commonpb.Payload{
		Metadata: map[string][]byte{"encoding": []byte("binary/protobuf")},
		Data:     data,
	}
}

func signalWithStartResponsePayload(t *testing.T, resp *workflowservice.SignalWithStartWorkflowExecutionResponse) *commonpb.Payload {
	t.Helper()
	data, err := proto.Marshal(resp)
	require.NoError(t, err)
	return &commonpb.Payload{
		Metadata: map[string][]byte{"encoding": []byte("binary/protobuf")},
		Data:     data,
	}
}

func TestUnwrapAndInjectRequest_NilPayloadIsNoOp(t *testing.T) {
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.unwrapAndInjectRequest(
		temporalSystemNexusEndpoint, "SignalWithStartWorkflowExecution",
		nil, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields, "nil payload should not inject anything")
}

func TestUnwrapAndInjectRequest_UnknownOperationIsNoOp(t *testing.T) {
	p := &commonpb.Payload{Data: []byte("ignored")}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.unwrapAndInjectRequest(
		temporalSystemNexusEndpoint, "NotARealOperation",
		p, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields, "unknown operation should be a no-op")
}

func TestUnwrapAndInjectRequest_UnknownEndpointIsNoOp(t *testing.T) {
	p := &commonpb.Payload{Data: []byte("ignored")}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.unwrapAndInjectRequest(
		"some-user-endpoint", "SignalWithStartWorkflowExecution",
		p, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields, "non-system endpoint should be a no-op even if operation name matches")
}

func TestUnwrapAndInjectRequest_BadProtoBytesReturnsError(t *testing.T) {
	p := &commonpb.Payload{Data: []byte{0xff, 0xff, 0xff}}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	err := iter.unwrapAndInjectRequest(
		temporalSystemNexusEndpoint, "SignalWithStartWorkflowExecution",
		p, fields, temporalproto.CustomJSONMarshalOptions{})
	require.Error(t, err)
	require.ErrorContains(t, err, "failed unmarshaling system nexus payload")
}

func TestUnwrapAndInjectResponse_NilPayloadIsNoOp(t *testing.T) {
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.unwrapAndInjectResponse(
		temporalSystemNexusEndpoint, "SignalWithStartWorkflowExecution",
		nil, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields)
}

func TestUnwrapAndInjectResponse_UnknownOperationIsNoOp(t *testing.T) {
	p := &commonpb.Payload{Data: []byte("ignored")}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.unwrapAndInjectResponse(
		temporalSystemNexusEndpoint, "NotARealOperation",
		p, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields)
}

func TestUnwrapAndInjectResponse_BadProtoBytesReturnsError(t *testing.T) {
	p := &commonpb.Payload{Data: []byte{0xff, 0xff, 0xff}}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	err := iter.unwrapAndInjectResponse(
		temporalSystemNexusEndpoint, "SignalWithStartWorkflowExecution",
		p, fields, temporalproto.CustomJSONMarshalOptions{})
	require.Error(t, err)
	require.ErrorContains(t, err, "failed unmarshaling system nexus payload")
}

func TestUnwrapAndInjectRequest_DecodesAllNestedPayloads(t *testing.T) {
	// The Input/SignalInput fields hold the user-supplied payloads, which the codec
	// should be applied to. The outer payload bytes are raw proto and are not codec-encoded.
	inner1 := &commonpb.Payload{Metadata: map[string][]byte{"encoding": []byte("binary/plain")}, Data: []byte("hello")}
	inner2 := &commonpb.Payload{Metadata: map[string][]byte{"encoding": []byte("binary/plain")}, Data: []byte("world")}
	signalInner := &commonpb.Payload{Metadata: map[string][]byte{"encoding": []byte("binary/plain")}, Data: []byte("signal-arg")}
	req := &workflowservice.SignalWithStartWorkflowExecutionRequest{
		Namespace:   "ns",
		WorkflowId:  "wf",
		Input:       &commonpb.Payloads{Payloads: []*commonpb.Payload{inner1, inner2}},
		SignalInput: &commonpb.Payloads{Payloads: []*commonpb.Payload{signalInner}},
	}
	p := signalWithStartRequestPayload(t, req)

	codec := &markingCodec{}
	iter := &structuredHistoryIter{ctx: context.Background(), codec: codec}
	fields := map[string]any{}
	require.NoError(t, iter.unwrapAndInjectRequest(
		temporalSystemNexusEndpoint, "SignalWithStartWorkflowExecution",
		p, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Equal(t, 3, codec.decodeCalls, "codec should have been invoked once per nested payload")

	unwrapped, ok := fields["unwrappedInput"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "ns", unwrapped["namespace"])
	require.Equal(t, "wf", unwrapped["workflowId"])
}

func TestUnwrapAndInjectRequest_CodecErrorPropagates(t *testing.T) {
	req := &workflowservice.SignalWithStartWorkflowExecutionRequest{
		Input: &commonpb.Payloads{Payloads: []*commonpb.Payload{{Data: []byte("x")}}},
	}
	p := signalWithStartRequestPayload(t, req)
	iter := &structuredHistoryIter{ctx: context.Background(), codec: failingCodec{}}
	fields := map[string]any{}
	err := iter.unwrapAndInjectRequest(
		temporalSystemNexusEndpoint, "SignalWithStartWorkflowExecution",
		p, fields, temporalproto.CustomJSONMarshalOptions{})
	require.Error(t, err)
	require.Equal(t, "codec decode failure for testing", err.Error())
}

func TestDecodePayloadsInProto_VisitsAllPayloads(t *testing.T) {
	req := &workflowservice.SignalWithStartWorkflowExecutionRequest{
		Input: &commonpb.Payloads{Payloads: []*commonpb.Payload{
			{Data: []byte("a")},
			{Data: []byte("b")},
		}},
		SignalInput: &commonpb.Payloads{Payloads: []*commonpb.Payload{
			{Data: []byte("c")},
		}},
	}
	codec := &markingCodec{}
	require.NoError(t, decodePayloadsInProto(context.Background(), req, codec))
	require.Equal(t, 3, codec.decodeCalls)
	require.Equal(t, []byte("decoded:a"), req.Input.Payloads[0].Data)
	require.Equal(t, []byte("decoded:b"), req.Input.Payloads[1].Data)
	require.Equal(t, []byte("decoded:c"), req.SignalInput.Payloads[0].Data)
}

func TestInjectSystemNexusUnwrapped_ScheduledKnownOp(t *testing.T) {
	req := &workflowservice.SignalWithStartWorkflowExecutionRequest{
		Namespace:  "ns",
		WorkflowId: "wf-xyz",
		SignalName: "ping",
	}
	event := &historypb.HistoryEvent{
		EventId:   5,
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
		Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
			NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
				Endpoint:  temporalSystemNexusEndpoint,
				Operation: "SignalWithStartWorkflowExecution",
				Input:     signalWithStartRequestPayload(t, req),
			},
		},
	}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.injectSystemNexusUnwrapped(event, fields, temporalproto.CustomJSONMarshalOptions{}))

	unwrapped, ok := fields["unwrappedInput"].(map[string]any)
	require.True(t, ok, "expected unwrappedInput map to be set")
	require.Equal(t, "ns", unwrapped["namespace"])
	require.Equal(t, "wf-xyz", unwrapped["workflowId"])
	require.Equal(t, "ping", unwrapped["signalName"])
}

func TestInjectSystemNexusUnwrapped_ScheduledUnknownEndpointSkipped(t *testing.T) {
	req := &workflowservice.SignalWithStartWorkflowExecutionRequest{Namespace: "ns"}
	event := &historypb.HistoryEvent{
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
		Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
			NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
				Endpoint:  "user-endpoint",
				Operation: "SignalWithStartWorkflowExecution",
				Input:     signalWithStartRequestPayload(t, req),
			},
		},
	}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.injectSystemNexusUnwrapped(event, fields, temporalproto.CustomJSONMarshalOptions{}))
	_, ok := fields["unwrappedInput"]
	require.False(t, ok, "non-system endpoint must not produce unwrappedInput")
}

func TestInjectSystemNexusUnwrapped_CompletedUsesPriorScheduled(t *testing.T) {
	resp := &workflowservice.SignalWithStartWorkflowExecutionResponse{
		RunId:   "run-abc",
		Started: true,
	}
	completed := &historypb.HistoryEvent{
		EventId:   6,
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_COMPLETED,
		Attributes: &historypb.HistoryEvent_NexusOperationCompletedEventAttributes{
			NexusOperationCompletedEventAttributes: &historypb.NexusOperationCompletedEventAttributes{
				ScheduledEventId: 5,
				Result:           signalWithStartResponsePayload(t, resp),
			},
		},
	}
	iter := &structuredHistoryIter{
		ctx:            context.Background(),
		systemNexusOps: map[int64]string{5: "SignalWithStartWorkflowExecution"},
	}
	fields := map[string]any{}
	require.NoError(t, iter.injectSystemNexusUnwrapped(completed, fields, temporalproto.CustomJSONMarshalOptions{}))

	unwrapped, ok := fields["unwrappedResult"].(map[string]any)
	require.True(t, ok, "expected unwrappedResult map to be set")
	require.Equal(t, "run-abc", unwrapped["runId"])
	require.Equal(t, true, unwrapped["started"])
}

func TestInjectSystemNexusUnwrapped_CompletedWithoutPriorScheduledSkipped(t *testing.T) {
	completed := &historypb.HistoryEvent{
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_COMPLETED,
		Attributes: &historypb.HistoryEvent_NexusOperationCompletedEventAttributes{
			NexusOperationCompletedEventAttributes: &historypb.NexusOperationCompletedEventAttributes{
				ScheduledEventId: 5,
				Result:           &commonpb.Payload{Data: []byte("garbage")},
			},
		},
	}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.injectSystemNexusUnwrapped(completed, fields, temporalproto.CustomJSONMarshalOptions{}))
	_, ok := fields["unwrappedResult"]
	require.False(t, ok, "no prior scheduled means we don't know the op, so no unwrap")
}

func TestInjectSystemNexusUnwrapped_NonNexusEventNoOp(t *testing.T) {
	event := &historypb.HistoryEvent{
		EventType: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &historypb.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &historypb.WorkflowExecutionStartedEventAttributes{},
		},
	}
	iter := &structuredHistoryIter{ctx: context.Background()}
	fields := map[string]any{}
	require.NoError(t, iter.injectSystemNexusUnwrapped(event, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Empty(t, fields)
}

func TestInjectSystemNexusUnwrapped_AppliesCodecToScheduledInput(t *testing.T) {
	req := &workflowservice.SignalWithStartWorkflowExecutionRequest{
		WorkflowId: "wf",
		Input: &commonpb.Payloads{Payloads: []*commonpb.Payload{
			{Data: []byte("inner-input")},
		}},
		SignalInput: &commonpb.Payloads{Payloads: []*commonpb.Payload{
			{Data: []byte("inner-signal")},
		}},
	}
	event := &historypb.HistoryEvent{
		EventType: enumspb.EVENT_TYPE_NEXUS_OPERATION_SCHEDULED,
		Attributes: &historypb.HistoryEvent_NexusOperationScheduledEventAttributes{
			NexusOperationScheduledEventAttributes: &historypb.NexusOperationScheduledEventAttributes{
				Endpoint:  temporalSystemNexusEndpoint,
				Operation: "SignalWithStartWorkflowExecution",
				Input:     signalWithStartRequestPayload(t, req),
			},
		},
	}
	codec := &markingCodec{}
	iter := &structuredHistoryIter{ctx: context.Background(), codec: codec}
	fields := map[string]any{}
	require.NoError(t, iter.injectSystemNexusUnwrapped(event, fields, temporalproto.CustomJSONMarshalOptions{}))
	require.Equal(t, 2, codec.decodeCalls, "codec should run on both nested payloads")
}
