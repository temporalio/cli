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

package cli_curr

import (
	"fmt"

	"github.com/temporalio/tctl/config"
	"github.com/urfave/cli"

	"github.com/temporalio/tctl/cli/headers"
	"github.com/temporalio/tctl/cli_curr/dataconverter"
	"github.com/temporalio/tctl/cli_curr/headersprovider"
	"github.com/temporalio/tctl/cli_curr/plugin"
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
	app.Version = headers.CLIVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   FlagAddressWithAlias,
			Value:  "",
			Usage:  "host:port for Temporal frontend service",
			EnvVar: "TEMPORAL_CLI_ADDRESS",
		},
		cli.StringFlag{
			Name:   FlagNamespaceWithAlias,
			Value:  "default",
			Usage:  "Temporal workflow namespace",
			EnvVar: "TEMPORAL_CLI_NAMESPACE",
		},
		cli.StringFlag{
			Name:   FlagAuth,
			Value:  "",
			Usage:  "Authorization header to set for GRPC requests",
			EnvVar: "TEMPORAL_CLI_AUTH",
		},
		cli.IntFlag{
			Name:   FlagContextTimeoutWithAlias,
			Value:  defaultContextTimeoutInSeconds,
			Usage:  "Optional timeout for context of RPC call in seconds",
			EnvVar: "TEMPORAL_CONTEXT_TIMEOUT",
		},
		cli.BoolFlag{
			Name:  FlagAutoConfirm,
			Usage: "Automatically confirm all prompts",
		},
		cli.StringFlag{
			Name:   FlagTLSCertPath,
			Value:  "",
			Usage:  "Path to x509 certificate",
			EnvVar: "TEMPORAL_CLI_TLS_CERT",
		},
		cli.StringFlag{
			Name:   FlagTLSKeyPath,
			Value:  "",
			Usage:  "Path to private key",
			EnvVar: "TEMPORAL_CLI_TLS_KEY",
		},
		cli.StringFlag{
			Name:   FlagTLSCaPath,
			Value:  "",
			Usage:  "Path to server CA certificate",
			EnvVar: "TEMPORAL_CLI_TLS_CA",
		},
		cli.BoolFlag{
			Name:   FlagTLSDisableHostVerification,
			Usage:  "Disable tls host name verification (tls must be enabled)",
			EnvVar: "TEMPORAL_CLI_TLS_DISABLE_HOST_VERIFICATION",
		},
		cli.StringFlag{
			Name:   FlagTLSServerName,
			Value:  "",
			Usage:  "Override for target server name",
			EnvVar: "TEMPORAL_CLI_TLS_SERVER_NAME",
		},
		cli.StringFlag{
			Name:   FlagHeadersProviderPluginWithAlias,
			Value:  "",
			Usage:  "Headers provider plugin executable name",
			EnvVar: "TEMPORAL_CLI_PLUGIN_HEADERS_PROVIDER",
		},
		cli.StringFlag{
			Name:   FlagDataConverterPluginWithAlias,
			Value:  "",
			Usage:  "Data converter plugin executable name",
			EnvVar: "TEMPORAL_CLI_PLUGIN_DATA_CONVERTER",
		},
		cli.StringFlag{
			Name:   FlagCodecEndpoint,
			Value:  "",
			Usage:  "Codec Server Endpoint",
			EnvVar: "TEMPORAL_CLI_CODEC_ENDPOINT",
		},
		cli.StringFlag{
			Name:   FlagCodecAuth,
			Value:  "",
			Usage:  "Authorization header to set for requests to Codec Server",
			EnvVar: "TEMPORAL_CLI_CODEC_AUTH",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:        "namespace",
			Aliases:     []string{"n"},
			Usage:       "Operate Temporal namespace",
			Subcommands: newNamespaceCommands(),
		},
		{
			Name:        "workflow",
			Aliases:     []string{"wf"},
			Usage:       "Operate Temporal workflow",
			Subcommands: newWorkflowCommands(),
		},
		{
			Name:        "activity",
			Aliases:     []string{"act"},
			Usage:       "Operate activities of workflow",
			Subcommands: newActivityCommands(),
		},
		{
			Name:        "taskqueue",
			Aliases:     []string{"tq"},
			Usage:       "Operate Temporal task queue",
			Subcommands: newTaskQueueCommands(),
		},
		{
			Name:        "schedule",
			Usage:       "Operate schedules",
			Subcommands: newScheduleCommands(),
		},
		{
			Name:        "batch",
			Usage:       "Batch operation on a list of workflows from query",
			Subcommands: newBatchCommands(),
		},
		{
			Name:        "batch-v2",
			Usage:       "Batch operation on a list of workflows from query",
			Subcommands: newBatchV2Commands(),
		},
		{
			Name:    "admin",
			Aliases: []string{"adm"},
			Usage:   "Run admin operation",
			Subcommands: []cli.Command{
				{
					Name:        "workflow",
					Aliases:     []string{"wf"},
					Usage:       "Run admin operation on workflow",
					Subcommands: newAdminWorkflowCommands(),
				},
				{
					Name:        "shard",
					Aliases:     []string{"shar"},
					Usage:       "Run admin operation on specific shard",
					Subcommands: newAdminShardManagementCommands(),
				},
				{
					Name:        "history_host",
					Aliases:     []string{"hist"},
					Usage:       "Run admin operation on history host",
					Subcommands: newAdminHistoryHostCommands(),
				},
				{
					Name:        "taskqueue",
					Aliases:     []string{"tq"},
					Usage:       "Run admin operation on taskQueue",
					Subcommands: newAdminTaskQueueCommands(),
				},
				{
					Name:        "membership",
					Usage:       "Run admin operation on membership",
					Subcommands: newAdminMembershipCommands(),
				},
				{
					Name:        "cluster",
					Aliases:     []string{"cl"},
					Usage:       "Run admin operation on cluster",
					Subcommands: newAdminClusterCommands(),
				},
				{
					Name:        "dlq",
					Aliases:     []string{"dlq"},
					Usage:       "Run admin operation on DLQ",
					Subcommands: newAdminDLQCommands(),
				},
				{
					Name:        "db",
					Aliases:     []string{"db"},
					Usage:       "Run admin operations on database",
					Subcommands: newDBCommands(),
				},
				{
					Name:        "decode",
					Usage:       "Decode payload",
					Subcommands: newDecodeCommands(),
				},
			},
		},
		{
			Name:        "cluster",
			Aliases:     []string{"cl"},
			Usage:       "Operate Temporal cluster",
			Subcommands: newClusterCommands(),
		},
		{
			Name:        "dataconverter",
			Aliases:     []string{"dc"},
			Usage:       "Operate Custom Data Converter",
			Subcommands: newDataConverterCommands(),
		},
		{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Configure tctl",
			Subcommands: newConfigCommands(),
		},
	}
	app.Before = configureSDK
	app.After = stopPlugins

	// set builder if not customized
	if cFactory == nil {
		SetFactory(NewClientFactory())
	}

	if tctlConfig == nil {
		var err error
		if tctlConfig, err = config.NewTctlConfig(); err != nil {
			fmt.Printf("unable to load tctl config: %v", err)
		}
	}

	return app
}

func configureSDK(c *cli.Context) error {
	endpoint := c.String(FlagCodecEndpoint)
	if endpoint != "" {
		dataconverter.SetRemoteEndpoint(
			endpoint,
			c.String(FlagNamespace),
			c.String(FlagCodecAuth),
		)
	}

	if c.String(FlagAuth) != "" {
		headersprovider.SetAuthorizationHeader(c.String(FlagAuth))
	}

	dcPlugin := c.String(FlagDataConverterPlugin)
	if dcPlugin != "" {
		dataConverter, err := plugin.NewDataConverterPlugin(dcPlugin)
		if err != nil {
			ErrorAndExit("unable to load data converter plugin", err)
		}

		dataconverter.SetCurrent(dataConverter)
	}

	hpPlugin := c.String(FlagHeadersProviderPlugin)
	if hpPlugin != "" {
		headersProvider, err := plugin.NewHeadersProviderPlugin(hpPlugin)
		if err != nil {
			ErrorAndExit("unable to load headers provider plugin", err)
		}

		headersprovider.SetCurrent(headersProvider)
	}

	return nil
}

func stopPlugins(c *cli.Context) error {
	plugin.StopPlugins()

	return nil
}
