package temporalcli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"

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
		// everywhere dataConverter is used.
		DataConverter: dataConverter,
	}

	// API key
	if c.ApiKey != "" {
		clientOptions.Credentials = client.NewAPIKeyStaticCredentials(c.ApiKey)
	}

	// Cloud options
	if c.Cloud {
		if err := c.applyCloudOptions(cctx, &clientOptions); err != nil {
			return nil, err
		}
	}

	// Headers
	if len(c.GrpcMeta) > 0 {
		headers := make(stringMapHeadersProvider, len(c.GrpcMeta))
		for _, kv := range c.GrpcMeta {
			pieces := strings.SplitN(kv, "=", 2)
			if len(pieces) != 2 {
				return nil, fmt.Errorf("gRPC meta of %q does not have '='", kv)
			}
			headers[pieces[0]] = pieces[1]
		}
		clientOptions.HeadersProvider = headers
	}

	// Remote codec
	if c.CodecEndpoint != "" {
		interceptor, err := payloadCodecInterceptor(c.Namespace, c.CodecEndpoint, c.CodecAuth)
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

func (c *ClientOptions) applyCloudOptions(cctx *CommandContext, clientOptions *client.Options) error {
	// Must have non-default namespace with single dot
	if strings.Count(c.Namespace, ".") != 1 {
		return fmt.Errorf("namespace must be provided and be a cloud namespace")
	}
	// Address must have been left at default or be expected address
	// TODO(cretz): This endpoint is not currently working
	clientOptions.HostPort = c.Namespace + ".tmprl.cloud:7233"
	if c.Address != "127.0.0.1:7233" && c.Address != clientOptions.HostPort {
		return fmt.Errorf("address should not be provided for cloud")
	}
	// If there is no API key and no TLS auth, try to use login token or fail
	if c.ApiKey == "" && c.TlsCertData == "" && c.TlsCertPath == "" {
		file := defaultCloudLoginTokenFile()
		if file == "" {
			return fmt.Errorf("no auth provided and unable to find home dir for cloud token file")
		}
		resp, err := readCloudLoginTokenFile(file)
		if err != nil {
			return fmt.Errorf("failed reading cloud token file: %w", err)
		} else if resp == nil {
			return fmt.Errorf("no auth provided and no cloud token present")
		}
		// Help the user out with a simple expiration check, but never fail if
		// unable to parse
		if t := getJWTExpiry(resp.AccessToken); !t.IsZero() {
			if t.Before(time.Now()) {
				cctx.Logger.Warn("Cloud token expired", "expiration", t)
			} else {
				cctx.Logger.Debug("Cloud token expires", "expiration", t)
			}
		}
		// TODO(cretz): Use gRPC OAuth creds with refresh token
		clientOptions.Credentials = client.NewAPIKeyStaticCredentials(resp.AccessToken)
	}
	return nil
}

// Zero time if unable to get
func getJWTExpiry(token string) time.Time {
	if tokenPieces := strings.Split(token, "."); len(tokenPieces) == 3 {
		if b, err := base64.RawURLEncoding.DecodeString(tokenPieces[1]); err == nil {
			var withExp struct {
				Exp int64 `json:"exp"`
			}
			if json.Unmarshal(b, &withExp) == nil && withExp.Exp > 0 {
				return time.Unix(withExp.Exp, 0)
			}
		}
	}
	return time.Time{}
}

func (c *ClientOptions) tlsConfig() (*tls.Config, error) {
	// We need TLS if any of these TLS options are set
	if !c.Cloud && !c.Tls &&
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

func payloadCodecInterceptor(namespace, codecEndpoint, codecAuth string) (grpc.UnaryClientInterceptor, error) {
	codecEndpoint = strings.ReplaceAll(codecEndpoint, "{namespace}", namespace)

	payloadCodec := converter.NewRemotePayloadCodec(
		converter.RemotePayloadCodecOptions{
			Endpoint: codecEndpoint,
			ModifyRequest: func(req *http.Request) error {
				req.Header.Set("X-Namespace", namespace)
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

var dataConverter = converter.NewCompositeDataConverter(
	rawValuePayloadConverter{},
	converter.NewNilPayloadConverter(),
	converter.NewByteSlicePayloadConverter(),
	converter.NewProtoJSONPayloadConverter(),
	converter.NewProtoPayloadConverter(),
	converter.NewJSONPayloadConverter(),
)

type rawValue struct{ payload *common.Payload }

type rawValuePayloadConverter struct{}

func (rawValuePayloadConverter) ToPayload(value any) (*common.Payload, error) {
	// Only convert if value is a raw value
	if r, ok := value.(rawValue); ok {
		return r.payload, nil
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
