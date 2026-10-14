package temporalcli

import (
	"context"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/proxy"
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
