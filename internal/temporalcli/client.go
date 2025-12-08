package temporalcli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strings"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/server/common/payload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Dial a client.
//
// Note, this call may mutate the receiver [ClientOptions.Namespace] since it is
// so often used by callers after this call to know the currently configured
// namespace.
func (c *ClientOptions) dialClient(cctx *CommandContext) (client.Client, error) {
	if cctx.RootCommand == nil {
		return nil, fmt.Errorf("root command unexpectedly missing when dialing client")
	}

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

	// Load a client config profile
	var clientProfile envconfig.ClientConfigProfile
	if !cctx.RootCommand.DisableConfigFile || !cctx.RootCommand.DisableConfigEnv {
		var err error
		clientProfile, err = envconfig.LoadClientConfigProfile(envconfig.LoadClientConfigProfileOptions{
			ConfigFilePath:    cctx.RootCommand.ConfigFile,
			ConfigFileProfile: cctx.RootCommand.Profile,
			DisableFile:       cctx.RootCommand.DisableConfigFile,
			DisableEnv:        cctx.RootCommand.DisableConfigEnv,
			EnvLookup:         cctx.Options.EnvLookup,
		})
		if err != nil {
			return nil, fmt.Errorf("failed loading client config: %w", err)
		}
	}

	// To support legacy TLS environment variables, if they are present, we will
	// have them force-override anything loaded from existing file or env
	if !cctx.RootCommand.DisableConfigEnv {
		oldEnvTLSCert, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_TLS_CERT")
		oldEnvTLSCertData, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_TLS_CERT_DATA")
		oldEnvTLSKey, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_TLS_KEY")
		oldEnvTLSKeyData, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_TLS_KEY_DATA")
		oldEnvTLSCA, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_TLS_CA")
		oldEnvTLSCAData, _ := cctx.Options.EnvLookup.LookupEnv("TEMPORAL_TLS_CA_DATA")
		if oldEnvTLSCert != "" || oldEnvTLSCertData != "" ||
			oldEnvTLSKey != "" || oldEnvTLSKeyData != "" ||
			oldEnvTLSCA != "" || oldEnvTLSCAData != "" {
			if clientProfile.TLS == nil {
				clientProfile.TLS = &envconfig.ClientConfigTLS{}
			}
			if oldEnvTLSCert != "" {
				clientProfile.TLS.ClientCertPath = oldEnvTLSCert
			}
			if oldEnvTLSCertData != "" {
				clientProfile.TLS.ClientCertData = []byte(oldEnvTLSCertData)
			}
			if oldEnvTLSKey != "" {
				clientProfile.TLS.ClientKeyPath = oldEnvTLSKey
			}
			if oldEnvTLSKeyData != "" {
				clientProfile.TLS.ClientKeyData = []byte(oldEnvTLSKeyData)
			}
			if oldEnvTLSCA != "" {
				clientProfile.TLS.ServerCACertPath = oldEnvTLSCA
			}
			if oldEnvTLSCAData != "" {
				clientProfile.TLS.ServerCACertData = []byte(oldEnvTLSCAData)
			}
		}
	}

	// Override some values in client config profile that come from CLI args. Some
	// flags, like address and namespace, have CLI defaults, but we don't want to
	// override the profile version unless it was _explicitly_ set.
	var addressExplicitlySet, namespaceExplicitlySet bool
	if cctx.CurrentCommand != nil {
		addressExplicitlySet = cctx.CurrentCommand.Flags().Changed("address")
		namespaceExplicitlySet = cctx.CurrentCommand.Flags().Changed("namespace")
	}
	if addressExplicitlySet {
		clientProfile.Address = c.Address
	}
	if namespaceExplicitlySet {
		clientProfile.Namespace = c.Namespace
	} else if clientProfile.Namespace != "" {
		// Since this namespace value is used by many commands after this call,
		// we are mutating it to be the derived one
		c.Namespace = clientProfile.Namespace
	}
	if c.ApiKey != "" {
		clientProfile.APIKey = c.ApiKey
	}
	if len(c.GrpcMeta) > 0 {
		// Append meta to the client profile
		grpcMetaFromArg, err := stringKeysValues(c.GrpcMeta)
		if err != nil {
			return nil, fmt.Errorf("invalid gRPC meta: %w", err)
		}
		if len(clientProfile.GRPCMeta) == 0 {
			clientProfile.GRPCMeta = make(map[string]string, len(c.GrpcMeta))
		}
		for k, v := range grpcMetaFromArg {
			clientProfile.GRPCMeta[k] = v
		}
	}

	// If any of these values are present, set TLS if not set, and set values.
	// NOTE: This means that tls=false does not explicitly disable TLS when set
	// via envconfig.
	if c.Tls ||
		c.TlsCertPath != "" || c.TlsKeyPath != "" || c.TlsCaPath != "" ||
		c.TlsCertData != "" || c.TlsKeyData != "" || c.TlsCaData != "" {
		if clientProfile.TLS == nil {
			clientProfile.TLS = &envconfig.ClientConfigTLS{}
		}
		if c.TlsCertPath != "" {
			clientProfile.TLS.ClientCertPath = c.TlsCertPath
		}
		if c.TlsCertData != "" {
			clientProfile.TLS.ClientCertData = []byte(c.TlsCertData)
		}
		if c.TlsKeyPath != "" {
			clientProfile.TLS.ClientKeyPath = c.TlsKeyPath
		}
		if c.TlsKeyData != "" {
			clientProfile.TLS.ClientKeyData = []byte(c.TlsKeyData)
		}
		if c.TlsCaPath != "" {
			clientProfile.TLS.ServerCACertPath = c.TlsCaPath
		}
		if c.TlsCaData != "" {
			clientProfile.TLS.ServerCACertData = []byte(c.TlsCaData)
		}
		if c.TlsServerName != "" {
			clientProfile.TLS.ServerName = c.TlsServerName
		}
		if c.TlsDisableHostVerification {
			clientProfile.TLS.DisableHostVerification = c.TlsDisableHostVerification
		}
	}

	// If TLS is explicitly disabled, we turn it off. Otherwise it may be
	// implicitly enabled if API key or any other TLS setting is set.
	if cctx.CurrentCommand.Flags().Changed("tls") && !c.Tls {
		clientProfile.TLS = &envconfig.ClientConfigTLS{Disabled: true}
	}

	// If codec endpoint is set, create codec setting regardless. But if auth is
	// set, it only overrides if codec is present.
	if c.CodecEndpoint != "" {
		if clientProfile.Codec == nil {
			clientProfile.Codec = &envconfig.ClientConfigCodec{}
		}
		clientProfile.Codec.Endpoint = c.CodecEndpoint
	}
	if c.CodecAuth != "" && clientProfile.Codec != nil {
		clientProfile.Codec.Auth = c.CodecAuth
	}

	// Now load client options from the profile
	clientOptions, err := clientProfile.ToClientOptions(envconfig.ToClientOptionsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed creating client options: %w", err)
	}

	if c.ClientAuthority != "" {
		clientOptions.ConnectionOptions.Authority = c.ClientAuthority
	}

	clientOptions.Logger = log.NewStructuredLogger(cctx.Logger)
	// We do not put codec on data converter here, it is applied via
	// interceptor. Same for failure conversion.
	// XXX: If this is altered to be more dynamic, have to also update
	// everywhere DataConverterWithRawValue is used.
	clientOptions.DataConverter = DataConverterWithRawValue

	// Remote codec
	if clientProfile.Codec != nil && clientProfile.Codec.Endpoint != "" {
		codecHeaders, err := stringKeysValues(c.CodecHeader)
		if err != nil {
			return nil, fmt.Errorf("invalid codec headers: %w", err)
		}
		interceptor, err := payloadCodecInterceptor(
			clientProfile.Namespace, clientProfile.Codec.Endpoint, clientProfile.Codec.Auth, codecHeaders)
		if err != nil {
			return nil, fmt.Errorf("failed creating payload codec interceptor: %w", err)
		}
		clientOptions.ConnectionOptions.DialOptions = append(
			clientOptions.ConnectionOptions.DialOptions, grpc.WithChainUnaryInterceptor(interceptor))
	}

	// Fixed header overrides
	clientOptions.ConnectionOptions.DialOptions = append(
		clientOptions.ConnectionOptions.DialOptions, grpc.WithChainUnaryInterceptor(fixedHeaderOverrideInterceptor))

	// Additional gRPC options
	clientOptions.ConnectionOptions.DialOptions = append(
		clientOptions.ConnectionOptions.DialOptions, cctx.Options.AdditionalClientGRPCDialOptions...)

	if cctx.Options.ClientConnectTimeout != 0 {
		// This is needed because the Go SDK overwrites the contextTimeout for GetSystemInfo, if not set
		clientOptions.ConnectionOptions.GetSystemInfoTimeout = cctx.Options.ClientConnectTimeout

		ctxWithTimeout, cancel := context.WithTimeoutCause(cctx, cctx.Options.ClientConnectTimeout,
			fmt.Errorf("command timed out after %v", cctx.Options.ClientConnectTimeout))
		defer cancel()
		return client.DialContext(ctxWithTimeout, clientOptions)
	}

	if len(c.Headers) > 0 {
		headerFields := map[string]*common.Payload{}
		for _, h := range c.Headers {
			parts := strings.SplitN(h, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid temporal headers %q — expected KEY=VALUE", h)
			}
			headerFields[parts[0]] = payload.EncodeString(parts[1])
		}
		clientOptions.ContextPropagators = []workflow.ContextPropagator{
			&workflowContextPropagator{Headers: headerFields},
		}
	}

	return client.DialContext(cctx, clientOptions)
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

func payloadCodecInterceptor(
	namespace string,
	codecEndpoint string,
	codecAuth string,
	codecHeaders map[string]string,
) (grpc.UnaryClientInterceptor, error) {
	codecEndpoint = strings.ReplaceAll(codecEndpoint, "{namespace}", namespace)

	payloadCodec := converter.NewRemotePayloadCodec(
		converter.RemotePayloadCodecOptions{
			Endpoint: codecEndpoint,
			ModifyRequest: func(req *http.Request) error {
				req.Header.Set("X-Namespace", namespace)
				for headerName, headerValue := range codecHeaders {
					req.Header.Set(headerName, headerValue)
				}
				if codecAuth != "" {
					req.Header.Set("Authorization", codecAuth)
				}
				return nil
			},
		},
	)
	return converter.NewPayloadCodecGRPCClientInterceptor(
		converter.PayloadCodecGRPCClientInterceptorOptions{
			Codecs: []converter.PayloadCodec{payloadCodec},
		},
	)
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

type workflowContextPropagator struct {
	Headers map[string]*common.Payload
}

func (w *workflowContextPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	for k, v := range w.Headers {
		writer.Set(k, v)
	}
	return nil
}

func (w *workflowContextPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	for k, v := range w.Headers {
		writer.Set(k, v)
	}
	return nil
}

func (w *workflowContextPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	return ctx, nil
}

func (w *workflowContextPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	return ctx, nil
}
