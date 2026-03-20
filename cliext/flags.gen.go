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

func (v *CommonOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.Env, "env", "default", "Active environment name (`ENV`).")
	f.StringVar(&v.EnvFile, "env-file", "", "Path to environment settings file. Defaults to `$HOME/.config/temporalio/temporal.yaml`.")
	f.StringVar(&v.ConfigFile, "config-file", "", "File path to read TOML config from, defaults to `$CONFIG_PATH/temporalio/temporal.toml` where `$CONFIG_PATH` is defined as `$HOME/.config` on Unix, `$HOME/Library/Application Support` on macOS, and `%AppData%` on Windows. EXPERIMENTAL.")
	f.StringVar(&v.Profile, "profile", "", "Profile to use for config file. EXPERIMENTAL.")
	f.BoolVar(&v.DisableConfigFile, "disable-config-file", false, "If set, disables loading environment config from config file. EXPERIMENTAL.")
	f.BoolVar(&v.DisableConfigEnv, "disable-config-env", false, "If set, disables loading environment config from environment variables. EXPERIMENTAL.")
	v.LogLevel = NewFlagStringEnum([]string{"debug", "info", "warn", "error", "never"}, "never")
	f.Var(&v.LogLevel, "log-level", "Log level. Default is \"never\" for most commands and \"warn\" for \"server start-dev\". Accepted values: debug, info, warn, error, never.")
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
	f.Var(&v.CommandTimeout, "command-timeout", "The command execution timeout. 0s means no timeout.")
	v.ClientConnectTimeout = 0
	f.Var(&v.ClientConnectTimeout, "client-connect-timeout", "The client connection timeout. 0s means no timeout.")
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

func (v *ClientOptions) BuildFlags(f *pflag.FlagSet) {
	v.FlagSet = f
	f.StringVar(&v.Address, "address", "localhost:7233", "Temporal Service gRPC endpoint.")
	f.StringVar(&v.ClientAuthority, "client-authority", "", "Temporal gRPC client :authority pseudoheader.")
	f.StringVarP(&v.Namespace, "namespace", "n", "default", "Temporal Service Namespace.")
	f.StringVar(&v.ApiKey, "api-key", "", "API key for request.")
	f.StringArrayVar(&v.GrpcMeta, "grpc-meta", nil, "HTTP headers for requests. Format as a `KEY=VALUE` pair. May be passed multiple times to set multiple headers. Can also be made available via environment variable as `TEMPORAL_GRPC_META_[name]`.")
	f.BoolVar(&v.Tls, "tls", false, "Enable base TLS encryption. Does not have additional options like mTLS or client certs. This is defaulted to true if api-key or any other TLS options are present. Use --tls=false to explicitly disable.")
	f.StringVar(&v.TlsCertPath, "tls-cert-path", "", "Path to x509 certificate. Can't be used with --tls-cert-data.")
	f.StringVar(&v.TlsCertData, "tls-cert-data", "", "Data for x509 certificate. Can't be used with --tls-cert-path.")
	f.StringVar(&v.TlsKeyPath, "tls-key-path", "", "Path to x509 private key. Can't be used with --tls-key-data.")
	f.StringVar(&v.TlsKeyData, "tls-key-data", "", "Private certificate key data. Can't be used with --tls-key-path.")
	f.StringVar(&v.TlsCaPath, "tls-ca-path", "", "Path to server CA certificate. Can't be used with --tls-ca-data.")
	f.StringVar(&v.TlsCaData, "tls-ca-data", "", "Data for server CA certificate. Can't be used with --tls-ca-path.")
	f.BoolVar(&v.TlsDisableHostVerification, "tls-disable-host-verification", false, "Disable TLS host-name verification.")
	f.StringVar(&v.TlsServerName, "tls-server-name", "", "Override target TLS server name.")
	f.StringVar(&v.CodecEndpoint, "codec-endpoint", "", "Remote Codec Server endpoint.")
	f.StringVar(&v.CodecAuth, "codec-auth", "", "Authorization header for Codec Server requests.")
	f.StringArrayVar(&v.CodecHeader, "codec-header", nil, "HTTP headers for requests to codec server. Format as a `KEY=VALUE` pair. May be passed multiple times to set multiple headers.")
	f.StringVar(&v.Identity, "identity", "", "The identity of the user or client submitting this request. Defaults to \"temporal-cli:$USER@$HOST\".")
}
