// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Copyright (c) 2021 Datadog, Inc.
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

package app

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/temporalio/cli/activity"
	"github.com/temporalio/cli/batch"
	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/cluster"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/cli/dataconverter"
	"github.com/temporalio/cli/env"
	"github.com/temporalio/cli/headersprovider"
	"github.com/temporalio/cli/namespace"
	"github.com/temporalio/cli/plugin"
	"github.com/temporalio/cli/schedule"
	"github.com/temporalio/cli/searchattribute"
	"github.com/temporalio/cli/server"
	sconfig "github.com/temporalio/cli/server/config"
	"github.com/temporalio/cli/taskqueue"
	"github.com/temporalio/cli/workflow"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/urfave/cli/v2"
	"go.temporal.io/server/common/headers"
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite" // load sqlite storage driver
)

func BuildApp(version string) *cli.App {
	defaultCfg, _ := sconfig.NewDefaultConfig()

	if version == "" {
		version = "(devel)"
	}
	app := cli.NewApp()
	app.Name = "temporal"
	app.Usage = "Temporal single binary"
	app.Version = fmt.Sprintf("%s (server %s)", version, headers.ServerVersion)
	app.Commands = commands(defaultCfg)
	app.Before = configureCLI
	app.After = stopPlugins
	app.ExitErrHandler = HandleError

	// set builder if not customized
	if client.CFactory == nil {
		SetFactory(client.NewClientFactory())
	}

	return app
}

// SetFactory is used to set the ClientFactory global
func SetFactory(factory client.ClientFactory) {
	client.CFactory = factory
}

func configureCLI(ctx *cli.Context) error {
	env.Build(ctx)
	return configureSDK(ctx)
}

func configureSDK(ctx *cli.Context) error {
	endpoint := ctx.String(common.FlagCodecEndpoint)
	if endpoint != "" {
		dataconverter.SetRemoteEndpoint(
			endpoint,
			ctx.String(common.FlagNamespace),
			ctx.String(common.FlagCodecAuth),
		)
	}

	if ctx.String(common.FlagAuth) != "" {
		headersprovider.SetAuthorizationHeader(ctx.String(common.FlagAuth))
	}

	hpPlugin := ctx.String(common.FlagHeadersProviderPlugin)
	if hpPlugin != "" {
		headersProvider, err := plugin.NewHeadersProviderPlugin(hpPlugin)
		if err != nil {
			return fmt.Errorf("unable to load headers provider plugin: %w", err)
		}

		headersprovider.SetCurrent(headersProvider)
	}

	return nil
}

func stopPlugins(ctx *cli.Context) error {
	plugin.StopPlugins()

	return nil
}

func HandleError(c *cli.Context, err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%s %+v\n", color.Red(c, "Error:"), err)
	if os.Getenv(common.ShowErrorStackEnv) != `` {
		fmt.Fprintln(os.Stderr, color.Magenta(c, "Stack trace:"))
		debug.PrintStack()
	} else {
		fmt.Fprintf(os.Stderr, "('export %s=1' to see stack traces)\n", common.ShowErrorStackEnv)
	}

	cli.OsExiter(1)
}

func commands(defaultCfg *sconfig.Config) []*cli.Command {
	return append([]*cli.Command{
		{
			Name:        "server",
			Usage:       "Commands for managing a Temporal server",
			Subcommands: server.NewServerCommands(defaultCfg),
		}}, clientCommands...)
}

var clientCommands = common.WithFlags([]*cli.Command{
	{
		Name:        "workflow",
		Usage:       "Operations on Workflows",
		Subcommands: workflow.NewWorkflowCommands(),
	},
	{
		Name:        "activity",
		Usage:       "Operations on Activities of Workflows",
		Subcommands: activity.NewActivityCommands(),
	},
	{
		Name:        "task-queue",
		Usage:       "Operations on Task Queues",
		Subcommands: taskqueue.NewTaskQueueCommands(),
	},
	{
		Name:        "schedule",
		Usage:       "Operations on Schedules",
		Subcommands: schedule.NewScheduleCommands(),
	},

	{
		Name:        "batch",
		Usage:       "Operations on Batch jobs",
		Subcommands: batch.NewBatchCommands(),
	},
	{
		Name:  "operator",
		Usage: "Operation on Temporal server",
		Subcommands: []*cli.Command{
			{
				Name:        "namespace",
				Usage:       "Operations on namespaces",
				Subcommands: namespace.NewNamespaceCommands(),
			},
			{
				Name:        "search-attribute",
				Usage:       "Operations on search attributes",
				Subcommands: searchattribute.NewSearchAttributeCommands(),
			},
			{
				Name:        "cluster",
				Usage:       "Operations on a Temporal cluster",
				Subcommands: cluster.NewClusterCommands(),
			},
		},
	},
	{
		Name:        "env",
		Usage:       "Manage client environment configurations",
		Subcommands: env.NewEnvCommands(),
	},
}, common.SharedFlags)
