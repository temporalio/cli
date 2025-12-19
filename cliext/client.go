package cliext

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"google.golang.org/grpc"
)

// ClientOptionsBuilder contains options for building SDK client.Options
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
	// oauthConfig is initialized during Build() if OAuth is configured.
	oauthConfig *OAuthConfig
}

// Build creates SDK client.Options
func (b *ClientOptionsBuilder) Build(ctx context.Context) (client.Options, error) {
	cfg := b.ClientOptions
	common := b.CommonOptions

	// Load a client config profile if configured
	var profile envconfig.ClientConfigProfile
	if !common.DisableConfigFile || !common.DisableConfigEnv {
		var err error
		profile, err = envconfig.LoadClientConfigProfile(envconfig.LoadClientConfigProfileOptions{
			ConfigFilePath:    common.ConfigFile,
			ConfigFileProfile: common.Profile,
			DisableFile:       common.DisableConfigFile,
			DisableEnv:        common.DisableConfigEnv,
			EnvLookup:         b.EnvLookup,
		})
		if err != nil {
			return client.Options{}, fmt.Errorf("failed loading client config: %w", err)
		}
	}

	// To support legacy TLS environment variables, if they are present, we will
	// have them force-override anything loaded from existing file or env
	if !common.DisableConfigEnv && b.EnvLookup != nil {
		oldEnvTLSCert, _ := b.EnvLookup.LookupEnv("TEMPORAL_TLS_CERT")
		oldEnvTLSCertData, _ := b.EnvLookup.LookupEnv("TEMPORAL_TLS_CERT_DATA")
		oldEnvTLSKey, _ := b.EnvLookup.LookupEnv("TEMPORAL_TLS_KEY")
		oldEnvTLSKeyData, _ := b.EnvLookup.LookupEnv("TEMPORAL_TLS_KEY_DATA")
		oldEnvTLSCA, _ := b.EnvLookup.LookupEnv("TEMPORAL_TLS_CA")
		oldEnvTLSCAData, _ := b.EnvLookup.LookupEnv("TEMPORAL_TLS_CA_DATA")
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
	if cfg.FlagSet != nil && cfg.FlagSet.Changed("namespace") {
		profile.Namespace = cfg.Namespace
	} else if profile.Namespace == "" {
		profile.Namespace = cfg.Namespace
	}

	// Set API key on profile if provided (OAuth credentials are set later on clientOpts)
	if cfg.ApiKey != "" {
		profile.APIKey = cfg.ApiKey
	}

	// Handle gRPC metadata from flags.
	if len(cfg.GrpcMeta) > 0 {
		grpcMetaFromArg, err := parseKeyValuePairs(cfg.GrpcMeta)
		if err != nil {
			return client.Options{}, fmt.Errorf("invalid gRPC meta: %w", err)
		}
		if len(profile.GRPCMeta) == 0 {
			profile.GRPCMeta = make(map[string]string, len(cfg.GrpcMeta))
		}
		for k, v := range grpcMetaFromArg {
			profile.GRPCMeta[k] = v
		}
	}

	// If any of these TLS values are present, set TLS if not set, and set values.
	// NOTE: This means that tls=false does not explicitly disable TLS when set
	// via envconfig.
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
		return client.Options{}, fmt.Errorf("failed to build client options: %w", err)
	}

	// Set client authority if provided.
	if cfg.ClientAuthority != "" {
		clientOpts.ConnectionOptions.Authority = cfg.ClientAuthority
	}

	// Set identity if provided.
	if cfg.Identity != "" {
		clientOpts.Identity = cfg.Identity
	}

	// Set logger if provided.
	if b.Logger != nil {
		clientOpts.Logger = log.NewStructuredLogger(b.Logger)
	}

	// Attempt to configure OAuth config if no API key is set.
	if cfg.ApiKey == "" {
		result, err := LoadClientOAuth(LoadClientOAuthOptions{
			ConfigFilePath: common.ConfigFile,
			ProfileName:    common.Profile,
			EnvLookup:      b.EnvLookup,
		})
		if err != nil {
			return client.Options{}, fmt.Errorf("failed to load OAuth config: %w", err)
		}
		// Only set credentials if OAuth is configured with an access token
		if result.OAuth != nil && result.OAuth.AccessToken != "" {
			b.oauthConfig = result.OAuth
			clientOpts.Credentials = client.NewAPIKeyDynamicCredentials(b.getOAuthToken)
		}
	}

	// Remote codec
	if profile.Codec != nil && profile.Codec.Endpoint != "" {
		codecHeaders, err := parseKeyValuePairs(cfg.CodecHeader)
		if err != nil {
			return client.Options{}, fmt.Errorf("invalid codec headers: %w", err)
		}
		interceptor, err := newPayloadCodecInterceptor(
			profile.Namespace, profile.Codec.Endpoint, profile.Codec.Auth, codecHeaders)
		if err != nil {
			return client.Options{}, fmt.Errorf("failed creating payload codec interceptor: %w", err)
		}
		clientOpts.ConnectionOptions.DialOptions = append(
			clientOpts.ConnectionOptions.DialOptions, grpc.WithChainUnaryInterceptor(interceptor))
	}

	// Set connect timeout for GetSystemInfo if provided.
	if common.ClientConnectTimeout != 0 {
		clientOpts.ConnectionOptions.GetSystemInfoTimeout = common.ClientConnectTimeout.Duration()
	}

	return clientOpts, nil
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

// getOAuthToken returns a valid OAuth access token from the builder's configuration.
// It uses oauth2.TokenSource to automatically refresh the token when needed.
func (b *ClientOptionsBuilder) getOAuthToken(ctx context.Context) (string, error) {
	tokenSource := b.oauthConfig.newTokenSource(ctx)
	token, err := tokenSource.Token()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}
