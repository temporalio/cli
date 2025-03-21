package temporalcli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strings"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (c *ClientOptions) dialClient(cctx *CommandContext) (client.Client, error) {
	clientOptions := client.Options{
		HostPort:  c.Address,
		Namespace: c.Namespace,
		Logger:    log.NewStructuredLogger(cctx.Logger),
		Identity:  clientIdentity(),
		// We do not put codec on data converter here, it is applied via
		// interceptor. Same for failure conversion.
		// XXX: If this is altered to be more dynamic, have to also update
		// everywhere DataConverterWithRawValue is used.
		DataConverter: DataConverterWithRawValue,
	}

	// API key
	if c.ApiKey != "" {
		clientOptions.Credentials = client.NewAPIKeyStaticCredentials(c.ApiKey)
	}

	// Headers
	if len(c.GrpcMeta) > 0 {
		headers, err := NewStringMapHeaderProvider(c.GrpcMeta)
		if err != nil {
			return nil, fmt.Errorf("grpc-meta %s", err)
		}
		clientOptions.HeadersProvider = headers
	}

	// Remote codec
	if c.CodecEndpoint != "" {
		codecHeaders, err := NewStringMapHeaderProvider(c.CodecHeader)
		if err != nil {
			return nil, fmt.Errorf("codec-header %s", err)
		}

		if c.CodecAuth != "" {
			codecHeaders["Authorization"] = c.CodecAuth
		}

		interceptor, err := payloadCodecInterceptor(c.Namespace, c.CodecEndpoint, codecHeaders)
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

	// TLS
	var err error
	if clientOptions.ConnectionOptions.TLS, err = c.tlsConfig(); err != nil {
		return nil, err
	}

	return client.Dial(clientOptions)
}

func (c *ClientOptions) tlsConfig() (*tls.Config, error) {
	// We need TLS if any of these TLS options are set
	if !c.Tls &&
		c.TlsCaPath == "" && c.TlsCertPath == "" && c.TlsKeyPath == "" &&
		c.TlsCaData == "" && c.TlsCertData == "" && c.TlsKeyData == "" {
		return nil, nil
	}

	conf := &tls.Config{
		ServerName:         c.TlsServerName,
		InsecureSkipVerify: c.TlsDisableHostVerification,
	}

	if c.TlsCertPath != "" {
		if c.TlsCertData != "" {
			return nil, fmt.Errorf("cannot specify both --tls-cert-path and --tls-cert-data")
		}
		clientCert, err := tls.LoadX509KeyPair(c.TlsCertPath, c.TlsKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed loading client cert key pair: %w", err)
		}
		conf.Certificates = append(conf.Certificates, clientCert)
	} else if c.TlsCertData != "" {
		clientCert, err := tls.X509KeyPair([]byte(c.TlsCertData), []byte(c.TlsKeyData))
		if err != nil {
			return nil, fmt.Errorf("failed loading client cert key pair: %w", err)
		}
		conf.Certificates = append(conf.Certificates, clientCert)
	}

	if c.TlsCaPath != "" {
		if c.TlsCaData != "" {
			return nil, fmt.Errorf("cannot specify both --tls-ca-path and --tls-ca-data")
		}
		conf.RootCAs = x509.NewCertPool()
		if b, err := os.ReadFile(c.TlsCaPath); err != nil {
			return nil, fmt.Errorf("failed reading CA cert from %v: %w", c.TlsCaPath, err)
		} else if !conf.RootCAs.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("invalid CA cert from %v", c.TlsCaPath)
		}
	} else if c.TlsCaData != "" {
		conf.RootCAs = x509.NewCertPool()
		if !conf.RootCAs.AppendCertsFromPEM([]byte(c.TlsCaData)) {
			return nil, fmt.Errorf("invalid CA cert data")
		}
	}
	return conf, nil
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

func payloadCodecInterceptor(namespace, codecEndpoint string, codecHeaders stringMapHeadersProvider) (grpc.UnaryClientInterceptor, error) {
	codecEndpoint = strings.ReplaceAll(codecEndpoint, "{namespace}", namespace)

	payloadCodec := converter.NewRemotePayloadCodec(
		converter.RemotePayloadCodecOptions{
			Endpoint: codecEndpoint,
			ModifyRequest: func(req *http.Request) error {
				req.Header.Set("X-Namespace", namespace)
				for headerName, headerValue := range codecHeaders {
					req.Header.Set(headerName, headerValue)
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

func clientIdentity() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
	username := "unknown-user"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	return "temporal-cli:" + username + "@" + hostname
}

type stringMapHeadersProvider map[string]string

func (s stringMapHeadersProvider) GetHeaders(context.Context) (map[string]string, error) {
	return s, nil
}

func NewStringMapHeaderProvider(config []string) (stringMapHeadersProvider, error) {
	headers := make(stringMapHeadersProvider, len(config))
	for _, kv := range config {
		pieces := strings.SplitN(kv, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("%q does not have '='", kv)
		}
		headers[pieces[0]] = pieces[1]
	}
	return headers, nil
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
