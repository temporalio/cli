package temporalcli

import (
	"context"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/proxy"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/api/workflowservice/v1/workflowservicenexus"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

// systemNexusOpKey identifies a system Nexus operation by its (endpoint, operation) pair.
type systemNexusOpKey struct {
	Endpoint  string
	Operation string
}

// systemNexusOpTypes maps a system Nexus operation to the proto request and response types
// whose bytes are serialized in NexusOperationScheduled.Input and NexusOperationCompleted.Result.
type systemNexusOpTypes struct {
	// NewRequest returns a fresh, zero-valued instance of the request proto.
	NewRequest func() proto.Message
	// NewResponse returns a fresh, zero-valued instance of the response proto.
	NewResponse func() proto.Message
}

// systemNexusOps is the global registry of known system Nexus operations on the
// __temporal_system endpoint. Add new entries here as the server adds support for more
// system operations. The keys' Operation values must match what the server records in
// NexusOperationScheduledEventAttributes.Operation.
// NOTE seankane: Part 2 of the System Operations work is to code generate this map from the
// go.temporal.io/api/workflowservice/v1/workflowservicenexus package.
var systemNexusOps = map[systemNexusOpKey]systemNexusOpTypes{
	{
		Endpoint:  temporalSystemNexusEndpoint,
		Operation: workflowservicenexus.TemporalAPIWorkflowserviceV1WorkflowService.SignalWithStartWorkflowExecution.Name(),
	}: {
		NewRequest:  func() proto.Message { return &workflowservice.SignalWithStartWorkflowExecutionRequest{} },
		NewResponse: func() proto.Message { return &workflowservice.SignalWithStartWorkflowExecutionResponse{} },
	},
}

// decodePayloadsInProto walks a proto message and applies codec.Decode to every Payload
// found inside it (including nested messages). The message is mutated in place.
func decodePayloadsInProto(ctx context.Context, msg proto.Message, codec converter.PayloadCodec) error {
	return proxy.VisitPayloads(ctx, msg, proxy.VisitPayloadsOptions{
		SkipSearchAttributes: true,
		Visitor: func(_ *proxy.VisitPayloadsContext, payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
			return codec.Decode(payloads)
		},
	})
}
