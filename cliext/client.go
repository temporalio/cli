package cliext

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"google.golang.org/grpc"
)

// ClientOptionsBuilder contains options for building Temporal client options.
type ClientOptionsBuilder struct {
	// CommonOptions contains common CLI options including profile config.
	CommonOptions CommonOptions
	// ClientOptions contains the client configuration from flags.
	ClientOptions ClientOptions
	// EnvLookup is the environment variable lookup function.
	// If nil, environment variables are not used for profile loading.
	EnvLookup envconfig.EnvLookup
	// Logger is the slog logger to use for the client. If set, it will be
	// wrapped with the SDK's structured logger adapter.
	Logger *slog.Logger
}

// BuildClientOptions creates SDK client options from a ClientOptionsBuilder.
// If OAuth is configured and no APIKey is set, OAuth will be used to obtain an access token.
// Returns the client options and the resolved namespace (which may differ from input if loaded from profile).
func BuildClientOptions(ctx context.Context, opts ClientOptionsBuilder) (client.Options, string, error) {
	cfg := opts.ClientOptions
	common := opts.CommonOptions

	// Load a client config profile if configured
	var profile envconfig.ClientConfigProfile
	if !common.DisableConfigFile || !common.DisableConfigEnv {
		var err error
		profile, err = envconfig.LoadClientConfigProfile(envconfig.LoadClientConfigProfileOptions{
			ConfigFilePath:    common.ConfigFile,
			ConfigFileProfile: common.Profile,
			DisableFile:       common.DisableConfigFile,
			DisableEnv:        common.DisableConfigEnv,
			EnvLookup:         opts.EnvLookup,
		})
		if err != nil {
			return client.Options{}, "", fmt.Errorf("failed loading client config: %w", err)
		}
	}

	// To support legacy TLS environment variables, if they are present, we will
	// have them force-override anything loaded from existing file or env
	if !common.DisableConfigEnv && opts.EnvLookup != nil {
		oldEnvTLSCert, _ := opts.EnvLookup.LookupEnv("TEMPORAL_TLS_CERT")
		oldEnvTLSCertData, _ := opts.EnvLookup.LookupEnv("TEMPORAL_TLS_CERT_DATA")
		oldEnvTLSKey, _ := opts.EnvLookup.LookupEnv("TEMPORAL_TLS_KEY")
		oldEnvTLSKeyData, _ := opts.EnvLookup.LookupEnv("TEMPORAL_TLS_KEY_DATA")
		oldEnvTLSCA, _ := opts.EnvLookup.LookupEnv("TEMPORAL_TLS_CA")
		oldEnvTLSCAData, _ := opts.EnvLookup.LookupEnv("TEMPORAL_TLS_CA_DATA")
		if oldEnvTLSCert != "" || oldEnvTLSCertData != "" ||
			oldEnvTLSKey != "" || oldEnvTLSKeyData != "" ||
			oldEnvTLSCA != "" || oldEnvTLSCAData != "" {
			if profile.TLS == nil {
				profile.TLS = &envconfig.ClientConfigTLS{}
			}
			if oldEnvTLSCert != "" {
				profile.TLS.ClientCertPath = oldEnvTLSCert
			}
			if oldEnvTLSCertData != "" {
				profile.TLS.ClientCertData = []byte(oldEnvTLSCertData)
			}
			if oldEnvTLSKey != "" {
				profile.TLS.ClientKeyPath = oldEnvTLSKey
			}
			if oldEnvTLSKeyData != "" {
				profile.TLS.ClientKeyData = []byte(oldEnvTLSKeyData)
			}
			if oldEnvTLSCA != "" {
				profile.TLS.ServerCACertPath = oldEnvTLSCA
			}
			if oldEnvTLSCAData != "" {
				profile.TLS.ServerCACertData = []byte(oldEnvTLSCAData)
			}
		}
	}

	// Override some values in client config profile that come from flags. Some
	// flags, like address and namespace, have defaults, but we don't want to
	// override the profile version unless it was _explicitly_ set.
	if cfg.FlagSet != nil && cfg.FlagSet.Changed("address") {
		profile.Address = cfg.Address
	}
	resolvedNamespace := profile.Namespace
	if cfg.FlagSet != nil && cfg.FlagSet.Changed("namespace") {
		profile.Namespace = cfg.Namespace
		resolvedNamespace = cfg.Namespace
	} else if profile.Namespace == "" {
		profile.Namespace = cfg.Namespace
		resolvedNamespace = cfg.Namespace
	}

	// Set API key on profile if provided (OAuth credentials are set later on clientOpts)
	if cfg.ApiKey != "" {
		profile.APIKey = cfg.ApiKey
	}

	// Handle gRPC metadata from flags.
	if len(cfg.GrpcMeta) > 0 {
		grpcMetaFromArg, err := parseKeyValuePairs(cfg.GrpcMeta)
		if err != nil {
			return client.Options{}, "", fmt.Errorf("invalid gRPC meta: %w", err)
		}
		if len(profile.GRPCMeta) == 0 {
			profile.GRPCMeta = make(map[string]string, len(cfg.GrpcMeta))
		}
		for k, v := range grpcMetaFromArg {
			profile.GRPCMeta[k] = v
		}
	}

	// If any of these TLS values are present, set TLS if not set, and set values.
	if cfg.Tls ||
		cfg.TlsCertPath != "" || cfg.TlsKeyPath != "" || cfg.TlsCaPath != "" ||
		cfg.TlsCertData != "" || cfg.TlsKeyData != "" || cfg.TlsCaData != "" {
		if profile.TLS == nil {
			profile.TLS = &envconfig.ClientConfigTLS{}
		}
		if cfg.TlsCertPath != "" {
			profile.TLS.ClientCertPath = cfg.TlsCertPath
		}
		if cfg.TlsCertData != "" {
			profile.TLS.ClientCertData = []byte(cfg.TlsCertData)
		}
		if cfg.TlsKeyPath != "" {
			profile.TLS.ClientKeyPath = cfg.TlsKeyPath
		}
		if cfg.TlsKeyData != "" {
			profile.TLS.ClientKeyData = []byte(cfg.TlsKeyData)
		}
		if cfg.TlsCaPath != "" {
			profile.TLS.ServerCACertPath = cfg.TlsCaPath
		}
		if cfg.TlsCaData != "" {
			profile.TLS.ServerCACertData = []byte(cfg.TlsCaData)
		}
		if cfg.TlsServerName != "" {
			profile.TLS.ServerName = cfg.TlsServerName
		}
		if cfg.TlsDisableHostVerification {
			profile.TLS.DisableHostVerification = cfg.TlsDisableHostVerification
		}
	}

	// If TLS is explicitly disabled (--tls=false), we turn it off. Otherwise it may be
	// implicitly enabled if API key or any other TLS setting is set.
	if cfg.FlagSet != nil && cfg.FlagSet.Changed("tls") && !cfg.Tls {
		profile.TLS = &envconfig.ClientConfigTLS{Disabled: true}
	}

	// If codec endpoint is set, create codec setting regardless. But if auth is
	// set, it only overrides if codec is present.
	if cfg.CodecEndpoint != "" {
		if profile.Codec == nil {
			profile.Codec = &envconfig.ClientConfigCodec{}
		}
		profile.Codec.Endpoint = cfg.CodecEndpoint
	}
	if cfg.CodecAuth != "" && profile.Codec != nil {
		profile.Codec.Auth = cfg.CodecAuth
	}

	// Convert profile to client options.
	clientOpts, err := profile.ToClientOptions(envconfig.ToClientOptionsRequest{})
	if err != nil {
		return client.Options{}, "", fmt.Errorf("failed to build client options: %w", err)
	}

	// Set identity if provided.
	if cfg.Identity != "" {
		clientOpts.Identity = cfg.Identity
	}

	// Set logger if provided.
	if opts.Logger != nil {
		clientOpts.Logger = log.NewStructuredLogger(opts.Logger)
	}

	// Set OAuth credentials if configured and no API key is set.
	// OAuth config is loaded on-demand from the config file.
	if cfg.ApiKey == "" {
		clientOpts.Credentials = client.NewAPIKeyDynamicCredentials(
			NewOAuthDynamicTokenProvider(opts))
	}

	// Set client authority if provided.
	if cfg.ClientAuthority != "" {
		clientOpts.ConnectionOptions.Authority = cfg.ClientAuthority
	}

	// Set connect timeout for GetSystemInfo if provided.
	if common.ClientConnectTimeout != 0 {
		clientOpts.ConnectionOptions.GetSystemInfoTimeout = common.ClientConnectTimeout.Duration()
	}

	// Add codec interceptor if codec endpoint is configured.
	if profile.Codec != nil && profile.Codec.Endpoint != "" {
		codecHeaders, err := parseKeyValuePairs(cfg.CodecHeader)
		if err != nil {
			return client.Options{}, "", fmt.Errorf("invalid codec headers: %w", err)
		}
		interceptor, err := newPayloadCodecInterceptor(
			profile.Namespace, profile.Codec.Endpoint, profile.Codec.Auth, codecHeaders)
		if err != nil {
			return client.Options{}, "", fmt.Errorf("failed creating payload codec interceptor: %w", err)
		}
		clientOpts.ConnectionOptions.DialOptions = append(
			clientOpts.ConnectionOptions.DialOptions, grpc.WithChainUnaryInterceptor(interceptor))
	}

	return clientOpts, resolvedNamespace, nil
}

// parseKeyValuePairs parses a slice of "KEY=VALUE" strings into a map.
func parseKeyValuePairs(pairs []string) (map[string]string, error) {
	result := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format %q, expected KEY=VALUE", pair)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}

// newPayloadCodecInterceptor creates a gRPC interceptor for remote payload codec.
func newPayloadCodecInterceptor(
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

// BuildTLSConfig creates a TLS configuration from the ClientOptions TLS settings.
// This is useful when you need the TLS config separately from the full client options.
func BuildTLSConfig(cfg ClientOptions) (*tls.Config, error) {
	if !cfg.Tls && cfg.TlsCertPath == "" && cfg.TlsKeyPath == "" && cfg.TlsCaPath == "" &&
		cfg.TlsCertData == "" && cfg.TlsKeyData == "" && cfg.TlsCaData == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		ServerName:         cfg.TlsServerName,
		InsecureSkipVerify: cfg.TlsDisableHostVerification,
	}

	// Load client certificate.
	if cfg.TlsCertPath != "" && cfg.TlsKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TlsCertPath, cfg.TlsKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if cfg.TlsCertData != "" && cfg.TlsKeyData != "" {
		cert, err := tls.X509KeyPair([]byte(cfg.TlsCertData), []byte(cfg.TlsKeyData))
		if err != nil {
			return nil, fmt.Errorf("failed to parse client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate.
	if cfg.TlsCaPath != "" || cfg.TlsCaData != "" {
		pool := x509.NewCertPool()
		var caData []byte
		if cfg.TlsCaPath != "" {
			var err error
			caData, err = os.ReadFile(cfg.TlsCaPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
		} else {
			caData = []byte(cfg.TlsCaData)
		}
		if !pool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = pool
	}

	return tlsConfig, nil
}
