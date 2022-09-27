// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl-kit/pkg/color"

	"github.com/temporalio/tctl/v2/cli/dataconverter"
	"github.com/temporalio/tctl/v2/cli/headersprovider"
	"github.com/temporalio/tctl/v2/cli/plugin"
	"github.com/temporalio/tctl/v2/config"
)

// SetFactory is used to set the ClientFactory global
func SetFactory(factory ClientFactory) {
	cFactory = factory
}

// NewCliApp instantiates a new instance of the CLI application.
func NewCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "tctl"
	app.Usage = "A command-line tool for Temporal users"
	app.Version = "2.0.0-beta.1"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    FlagAddress,
			Value:   "",
			Usage:   "host:port for Temporal frontend service",
			EnvVars: []string{"TEMPORAL_CLI_ADDRESS"},
		},
		&cli.StringFlag{
			Name:    FlagNamespace,
			Aliases: FlagNamespaceAlias,
			Value:   "default",
			Usage:   "Temporal workflow namespace",
			EnvVars: []string{"TEMPORAL_CLI_NAMESPACE"},
		},
		&cli.StringFlag{
			Name:    FlagAuth,
			Value:   "",
			Usage:   "Authorization header to set for gRPC requests",
			EnvVars: []string{"TEMPORAL_CLI_AUTH"},
		},
		&cli.StringFlag{
			Name:    FlagTLSCertPath,
			Value:   "",
			Usage:   "Path to x509 certificate",
			EnvVars: []string{"TEMPORAL_CLI_TLS_CERT"},
		},
		&cli.StringFlag{
			Name:    FlagTLSKeyPath,
			Value:   "",
			Usage:   "Path to private key",
			EnvVars: []string{"TEMPORAL_CLI_TLS_KEY"},
		},
		&cli.StringFlag{
			Name:    FlagTLSCaPath,
			Value:   "",
			Usage:   "Path to server CA certificate",
			EnvVars: []string{"TEMPORAL_CLI_TLS_CA"},
		},
		&cli.BoolFlag{
			Name:    FlagTLSDisableHostVerification,
			Usage:   "Disable tls host name verification (tls must be enabled)",
			EnvVars: []string{"TEMPORAL_CLI_TLS_DISABLE_HOST_VERIFICATION"},
		},
		&cli.StringFlag{
			Name:    FlagTLSServerName,
			Value:   "",
			Usage:   "Override for target server name",
			EnvVars: []string{"TEMPORAL_CLI_TLS_SERVER_NAME"},
		},
		&cli.IntFlag{
			Name:    FlagContextTimeout,
			Value:   defaultContextTimeoutInSeconds,
			Usage:   "Optional timeout for context of RPC call in seconds",
			EnvVars: []string{"TEMPORAL_CONTEXT_TIMEOUT"},
		},
		&cli.StringFlag{
			Name:    FlagHeadersProviderPlugin,
			Value:   "",
			Usage:   "Headers provider plugin executable name",
			EnvVars: []string{"TEMPORAL_CLI_PLUGIN_HEADERS_PROVIDER"},
		},
		&cli.StringFlag{
			Name:    FlagDataConverterPlugin,
			Value:   "",
			Usage:   "Data converter plugin executable name",
			EnvVars: []string{"TEMPORAL_CLI_PLUGIN_DATA_CONVERTER"},
		},
		&cli.StringFlag{
			Name:    FlagCodecEndpoint,
			Value:   "",
			Usage:   "Remote Codec Server Endpoint",
			EnvVars: []string{"TEMPORAL_CLI_CODEC_ENDPOINT"},
		},
		&cli.StringFlag{
			Name:    FlagCodecAuth,
			Value:   "",
			Usage:   "Authorization header to set for requests to Codec Server",
			EnvVars: []string{"TEMPORAL_CLI_CODEC_AUTH"},
		},
		&cli.StringFlag{
			Name:  color.FlagColor,
			Usage: fmt.Sprintf("when to use color: %v, %v, %v.", color.Auto, color.Always, color.Never),
			Value: string(color.Auto),
		},
	}
	app.Commands = tctlCommands
	app.Before = configureSDK
	app.After = stopPlugins
	app.ExitErrHandler = handleError

	// set builder if not customized
	if cFactory == nil {
		SetFactory(NewClientFactory())
	}

	tctlConfig, _ = config.NewTctlConfig()
	populateFlags(app.Commands, app.Flags)
	useDynamicCommands(app)

	return app
}

func configureSDK(ctx *cli.Context) error {
	endpoint := ctx.String(FlagCodecEndpoint)
	if endpoint != "" {
		dataconverter.SetRemoteEndpoint(
			endpoint,
			ctx.String(FlagNamespace),
			ctx.String(FlagCodecAuth),
		)
	}

	if ctx.String(FlagAuth) != "" {
		headersprovider.SetAuthorizationHeader(ctx.String(FlagAuth))
	}

	dcPlugin := ctx.String(FlagDataConverterPlugin)
	if dcPlugin != "" {
		dataConverter, err := plugin.NewDataConverterPlugin(dcPlugin)
		if err != nil {
			return fmt.Errorf("unable to load data converter plugin: %s", err)
		}

		dataconverter.SetCurrent(dataConverter)
	}

	hpPlugin := ctx.String(FlagHeadersProviderPlugin)
	if hpPlugin != "" {
		headersProvider, err := plugin.NewHeadersProviderPlugin(hpPlugin)
		if err != nil {
			return fmt.Errorf("unable to load headers provider plugin: %s", err)
		}

		headersprovider.SetCurrent(headersProvider)
	}

	return nil
}

func stopPlugins(ctx *cli.Context) error {
	plugin.StopPlugins()

	return nil
}

func handleError(c *cli.Context, err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%s %+v\n", color.Red(c, "Error:"), err)
	if os.Getenv(showErrorStackEnv) != `` {
		fmt.Fprintln(os.Stderr, color.Magenta(c, "Stack trace:"))
		debug.PrintStack()
	} else {
		fmt.Fprintf(os.Stderr, "('export %s=1' to see stack traces)\n", showErrorStackEnv)
	}

	cli.OsExiter(1)
}
