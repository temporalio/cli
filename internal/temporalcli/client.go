package temporalcli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"

	"github.com/temporalio/cli/cliext"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// dialClient creates a Temporal client using cliext.ClientOptionsBuilder with CLI-specific customizations.
//
// Note, this call may mutate the ClientOptions.Namespace since it is
// so often used by callers after this call to know the currently configured
// namespace.
func dialClient(cctx *CommandContext, c *cliext.ClientOptions) (client.Client, error) {
	cl, _, err := dialClientWithCodec(cctx, c)
	return cl, err
}

// dialClientWithCodec is like [dialClient] but also returns the configured remote
// payload codec, or nil if no codec is configured. The codec is the same instance
// used by the gRPC interceptor; callers can use it to decode payloads nested inside
// opaque proto bytes (e.g. the request/response of a system Nexus operation).
func dialClientWithCodec(cctx *CommandContext, c *cliext.ClientOptions) (client.Client, converter.PayloadCodec, error) {
	if cctx.RootCommand == nil {
		return nil, nil, fmt.Errorf("root command unexpectedly missing when dialing client")
	}

	// Set default identity if not provided
	if c.Identity == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown-host"
		}
		username := "unknown-user"
		if u, err := user.Current(); err == nil {
			username = u.Username
		}
		c.Identity = "temporal-cli:" + username + "@" + hostname
	}

	// Build client options using cliext
	builder := &cliext.ClientOptionsBuilder{
		CommonOptions: cctx.RootCommand.CommonOptions,
		ClientOptions: *c,
		EnvLookup:     cctx.Options.EnvLookup,
		Logger:        cctx.Logger,
	}
	clientOpts, err := builder.Build(cctx)
	if err != nil {
		// An unreadable TLS/credential file is a connection-setup problem
		// worth a suggestion, even though no dial happened.
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			diag := &connectDiagnosis{
				Address: builder.ResolvedAddress,
				Cause:   causeCertFileUnreadable,
				Detail:  pathErr.Path,
			}
			return nil, nil, newConnectError(diag, connectMetaFromBuilder(cctx, builder), err)
		}
		return nil, nil, err
	}

	// We do not put codec on data converter here, it is applied via
	// interceptor. Same for failure conversion.
	// XXX: If this is altered to be more dynamic, have to also update
	// everywhere DataConverterWithRawValue is used.
	clientOpts.DataConverter = DataConverterWithRawValue

	// Add header propagator.
	clientOpts.ContextPropagators = append(clientOpts.ContextPropagators, headerPropagator{})

	// Fixed header overrides
	clientOpts.ConnectionOptions.DialOptions = append(
		clientOpts.ConnectionOptions.DialOptions, grpc.WithChainUnaryInterceptor(fixedHeaderOverrideInterceptor))

	// Additional gRPC options
	clientOpts.ConnectionOptions.DialOptions = append(
		clientOpts.ConnectionOptions.DialOptions, cctx.Options.AdditionalClientGRPCDialOptions...)

	// Apply context timeout for dial if configured
	dialCtx := context.Context(cctx)
	if cctx.RootCommand.CommonOptions.ClientConnectTimeout != 0 {
		timeout := cctx.RootCommand.CommonOptions.ClientConnectTimeout.Duration()
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeoutCause(cctx, timeout, fmt.Errorf("command timed out after %v", timeout))
		defer cancel()
	}

	cl, err := client.DialContext(dialCtx, clientOpts)
	if err != nil {
		return nil, nil, dialConnectError(cctx, dialCtx, builder, clientOpts, err)
	}

	// Since this namespace value is used by many commands after this call,
	// we are mutating it to be the derived one
	c.Namespace = clientOpts.Namespace

	return cl, builder.PayloadCodec, nil
}

// dialConnectError enriches a client.DialContext failure with a staged
// connection diagnosis and a suggested fix. See connectdiag.go.
func dialConnectError(
	cctx *CommandContext,
	dialCtx context.Context,
	builder *cliext.ClientOptionsBuilder,
	clientOpts client.Options,
	origErr error,
) error {
	// If a CLI-side timeout fired, surface its descriptive cause, which is
	// otherwise lost ("context deadline exceeded" wins over "command timed
	// out after 5s").
	if dialCtx.Err() != nil {
		if cause := context.Cause(dialCtx); cause != nil && !errors.Is(cause, dialCtx.Err()) {
			origErr = fmt.Errorf("%w (%v)", origErr, cause)
		}
	}
	// Never probe after an interrupt or when the whole command timed out.
	if cctx.Err() != nil {
		return origErr
	}
	if v, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_CLI_DISABLE_CONNECT_DIAGNOSIS"); v != "" {
		return fmt.Errorf("failed connecting to Temporal server at %v: %w", clientOpts.HostPort, origErr)
	}
	// The probe has its own internal budget, but never let it exceed the
	// user's explicit connect timeout.
	probeCtx := context.Context(cctx)
	if timeout := cctx.RootCommand.CommonOptions.ClientConnectTimeout.Duration(); timeout > 0 {
		var cancel context.CancelFunc
		probeCtx, cancel = context.WithTimeout(cctx, timeout)
		defer cancel()
	}
	diag := diagnoseConnection(probeCtx, clientOpts.HostPort, clientOpts.ConnectionOptions.TLS, origErr)
	meta := connectMetaFromBuilder(cctx, builder)
	meta.TLSConfigured = clientOpts.ConnectionOptions.TLS != nil
	return newConnectError(diag, meta, origErr)
}

func connectMetaFromBuilder(cctx *CommandContext, builder *cliext.ClientOptionsBuilder) connectMeta {
	return connectMeta{
		Args:          cctx.Options.Args,
		Address:       builder.ResolvedAddress,
		AddressSource: builder.ResolvedAddressSource,
		ProfileName:   builder.ResolvedProfileName,
		HasAPIKey:     builder.HasAPIKey,
	}
}

func fixedHeaderOverrideInterceptor(
	ctx context.Context,
	method string, req, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	// The SDK sets some values on the outgoing metadata that we can't override
	// via normal headers, so we have to replace directly on the metadata
	md, _ := metadata.FromOutgoingContext(ctx)
	if md == nil {
		md = metadata.MD{}
	}
	md.Set("client-name", "temporal-cli")
	md.Set("client-version", Version)
	md.Set("supported-server-versions", ">=1.0.0 <2.0.0")
	md.Set("caller-type", "operator")
	ctx = metadata.NewOutgoingContext(ctx, md)
	return invoker(ctx, method, req, reply, cc, opts...)
}

var DataConverterWithRawValue = converter.NewCompositeDataConverter(
	rawValuePayloadConverter{},
	converter.NewNilPayloadConverter(),
	converter.NewByteSlicePayloadConverter(),
	converter.NewProtoJSONPayloadConverter(),
	converter.NewProtoPayloadConverter(),
	converter.NewJSONPayloadConverter(),
)

type RawValue struct{ Payload *common.Payload }

type rawValuePayloadConverter struct{}

func (rawValuePayloadConverter) ToPayload(value any) (*common.Payload, error) {
	// Only convert if value is a raw value
	if r, ok := value.(RawValue); ok {
		return r.Payload, nil
	}
	return nil, nil
}

func (rawValuePayloadConverter) FromPayload(payload *common.Payload, valuePtr any) error {
	return fmt.Errorf("raw value unsupported from payload")
}

func (rawValuePayloadConverter) ToString(p *common.Payload) string {
	return fmt.Sprintf("<raw payload %v bytes>", len(p.Data))
}

func (rawValuePayloadConverter) Encoding() string {
	// Should never be used
	return "raw-value-encoding"
}

type headerPropagator struct{}

type cliHeaderContextKey struct{}

func (headerPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	if headers, ok := ctx.Value(cliHeaderContextKey{}).(map[string]any); ok {
		for k, v := range headers {
			p, err := converter.GetDefaultDataConverter().ToPayload(v)
			if err != nil {
				return err
			}
			writer.Set(k, p)
		}
	}
	return nil
}

func (headerPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	return nil
}

func (headerPropagator) Extract(ctx context.Context, _ workflow.HeaderReader) (context.Context, error) {
	return ctx, nil
}

func (headerPropagator) ExtractToWorkflow(ctx workflow.Context, _ workflow.HeaderReader) (workflow.Context, error) {
	return ctx, nil
}

func contextWithHeaders(ctx context.Context, headers []string) (context.Context, error) {
	if len(headers) == 0 {
		return ctx, nil
	}
	out, err := stringKeysJSONValues(headers, false)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, cliHeaderContextKey{}, out), nil
}
