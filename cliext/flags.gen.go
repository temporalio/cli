// Code generated. DO NOT EDIT.

package cliext

import (
	"github.com/mattn/go-isatty"

	"github.com/spf13/pflag"

	"os"
)

var hasHighlighting = isatty.IsTerminal(os.Stdout.Fd())

type CommonOptions struct {
	Env                     string
	EnvFile                 string
	ConfigFile              string
	Profile                 string
	DisableConfigFile       bool
	DisableConfigEnv        bool
	LogLevel                FlagStringEnum
	LogFormat               FlagStringEnum
	Output                  FlagStringEnum
	TimeFormat              FlagStringEnum
	Color                   FlagStringEnum
	NoJsonShorthandPayloads bool
	CommandTimeout          FlagDuration
	ClientConnectTimeout    FlagDuration
	FlagSet                 *pflag.FlagSet
}

func (v *CommonOptions) Description() string {
	return "These options apply to every command. They control output formatting,\nlogging, and which configuration profile and environment to use.\nOptions that accept an environment variable can be set instead of\npassing the flag each time.\n"
}

func (v *CommonOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.Env, "env", "default", "Active environment name (`ENV`). Env: TEMPORAL_ENV.")
	f.StringVar(&v.EnvFile, "env-file", "", "Path to environment settings file. Env: TEMPORAL_ENV_FILE.")
	f.StringVar(&v.ConfigFile, "config-file", "", "TOML config file path. Env: TEMPORAL_CONFIG_FILE.")
	f.StringVar(&v.Profile, "profile", "", "Profile to use for config file. Env: TEMPORAL_PROFILE.")
	f.BoolVar(&v.DisableConfigFile, "disable-config-file", false, "Disable loading config from file.")
	f.BoolVar(&v.DisableConfigEnv, "disable-config-env", false, "Disable loading config from environment variables.")
	v.LogLevel = NewFlagStringEnum([]string{"debug", "info", "warn", "error", "never"}, "never")
	f.Var(&v.LogLevel, "log-level", "Log level. Accepted values: debug, info, warn, error, never.")
	v.LogFormat = NewFlagStringEnum([]string{"text", "json", "pretty"}, "text")
	f.Var(&v.LogFormat, "log-format", "Log format. Accepted values: text, json.")
	v.Output = NewFlagStringEnum([]string{"text", "json", "jsonl", "none"}, "text")
	f.VarP(&v.Output, "output", "o", "Non-logging data output format. Accepted values: text, json, jsonl, none.")
	v.TimeFormat = NewFlagStringEnum([]string{"relative", "iso", "raw"}, "relative")
	f.Var(&v.TimeFormat, "time-format", "Time format. Accepted values: relative, iso, raw.")
	v.Color = NewFlagStringEnum([]string{"always", "never", "auto"}, "auto")
	f.Var(&v.Color, "color", "Output coloring. Accepted values: always, never, auto.")
	f.BoolVar(&v.NoJsonShorthandPayloads, "no-json-shorthand-payloads", false, "Raw payload output, even if the JSON option was used.")
	v.CommandTimeout = 0
	f.Var(&v.CommandTimeout, "command-timeout", "Command execution timeout.")
	v.ClientConnectTimeout = 0
	f.Var(&v.ClientConnectTimeout, "client-connect-timeout", "Client connection timeout.")
}

type ClientOptions struct {
	Address                    string
	ClientAuthority            string
	Namespace                  string
	ApiKey                     string
	GrpcMeta                   []string
	Tls                        bool
	TlsCertPath                string
	TlsCertData                string
	TlsKeyPath                 string
	TlsKeyData                 string
	TlsCaPath                  string
	TlsCaData                  string
	TlsDisableHostVerification bool
	TlsServerName              string
	CodecEndpoint              string
	CodecAuth                  string
	CodecHeader                []string
	Identity                   string
	FlagSet                    *pflag.FlagSet
}

func (v *ClientOptions) Description() string {
	return "These options apply to commands that connect to a Temporal Service\n(workflow, activity, schedule, etc). They specify the server address,\nnamespace, authentication, and TLS settings. Values are resolved in\norder: CLI flag > environment variable > config file.\n\nTo persist these settings, use:\n  temporal config set --prop KEY --value VALUE\n"
}

func (v *ClientOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.Address, "address", "localhost:7233", "Temporal Service gRPC endpoint. Env: TEMPORAL_ADDRESS. Config: address.")
	f.StringVar(&v.ClientAuthority, "client-authority", "", "Temporal gRPC client :authority pseudoheader.")
	f.StringVarP(&v.Namespace, "namespace", "n", "default", "Temporal Service Namespace. Env: TEMPORAL_NAMESPACE. Config: namespace.")
	f.StringVar(&v.ApiKey, "api-key", "", "API key for request. Env: TEMPORAL_API_KEY. Config: api_key.")
	f.StringArrayVar(&v.GrpcMeta, "grpc-meta", nil, "HTTP headers for requests (KEY=VALUE, repeatable). Config: grpc_meta.<key>.")
	f.BoolVar(&v.Tls, "tls", false, "Enable base TLS encryption. Auto-enabled when api-key or TLS options are set. Env: TEMPORAL_TLS. Config: tls.")
	f.StringVar(&v.TlsCertPath, "tls-cert-path", "", "Path to x509 certificate. Env: TEMPORAL_TLS_CLIENT_CERT_PATH. Config: tls.client_cert_path.")
	f.StringVar(&v.TlsCertData, "tls-cert-data", "", "Inline x509 certificate data. Env: TEMPORAL_TLS_CLIENT_CERT_DATA. Config: tls.client_cert_data.")
	f.StringVar(&v.TlsKeyPath, "tls-key-path", "", "Path to x509 private key. Env: TEMPORAL_TLS_CLIENT_KEY_PATH. Config: tls.client_key_path.")
	f.StringVar(&v.TlsKeyData, "tls-key-data", "", "Inline x509 private key data. Env: TEMPORAL_TLS_CLIENT_KEY_DATA. Config: tls.client_key_data.")
	f.StringVar(&v.TlsCaPath, "tls-ca-path", "", "Path to server CA certificate. Env: TEMPORAL_TLS_SERVER_CA_CERT_PATH. Config: tls.server_ca_cert_path.")
	f.StringVar(&v.TlsCaData, "tls-ca-data", "", "Inline server CA certificate data. Env: TEMPORAL_TLS_SERVER_CA_CERT_DATA. Config: tls.server_ca_cert_data.")
	f.BoolVar(&v.TlsDisableHostVerification, "tls-disable-host-verification", false, "Disable TLS host-name verification. Env: TEMPORAL_TLS_DISABLE_HOST_VERIFICATION. Config: tls.disable_host_verification.")
	f.StringVar(&v.TlsServerName, "tls-server-name", "", "Override target TLS server name. Env: TEMPORAL_TLS_SERVER_NAME. Config: tls.server_name.")
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "Remote Codec Server endpoint. Env: TEMPORAL_CODEC_ENDPOINT. Config: codec.endpoint.")
	f.StringVar(&v.CodecAuth, "codec-auth", "", "Authorization header for Codec Server requests. Env: TEMPORAL_CODEC_AUTH. Config: codec.auth.")
	f.StringArrayVar(&v.CodecHeader, "codec-header", nil, "HTTP headers for codec server (KEY=VALUE, repeatable).")
	f.StringVar(&v.Identity, "identity", "", "Identity of the client submitting requests.")
}

func (v *ClientOptions) HideFlags() {
	if v.FlagSet == nil {
		return
	}
	v.FlagSet.Lookup("address").Hidden = true
	v.FlagSet.Lookup("client-authority").Hidden = true
	v.FlagSet.Lookup("namespace").Hidden = true
	v.FlagSet.Lookup("api-key").Hidden = true
	v.FlagSet.Lookup("grpc-meta").Hidden = true
	v.FlagSet.Lookup("tls").Hidden = true
	v.FlagSet.Lookup("tls-cert-path").Hidden = true
	v.FlagSet.Lookup("tls-cert-data").Hidden = true
	v.FlagSet.Lookup("tls-key-path").Hidden = true
	v.FlagSet.Lookup("tls-key-data").Hidden = true
	v.FlagSet.Lookup("tls-ca-path").Hidden = true
	v.FlagSet.Lookup("tls-ca-data").Hidden = true
	v.FlagSet.Lookup("tls-disable-host-verification").Hidden = true
	v.FlagSet.Lookup("tls-server-name").Hidden = true
	v.FlagSet.Lookup("codec-endpoint").Hidden = true
	v.FlagSet.Lookup("codec-auth").Hidden = true
	v.FlagSet.Lookup("codec-header").Hidden = true
	v.FlagSet.Lookup("identity").Hidden = true
}
