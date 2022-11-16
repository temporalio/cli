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

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/temporal-cli/activity"
	"github.com/temporalio/temporal-cli/batch"
	"github.com/temporalio/temporal-cli/client"
	"github.com/temporalio/temporal-cli/cluster"
	"github.com/temporalio/temporal-cli/common"
	"github.com/temporalio/temporal-cli/env"
	"github.com/temporalio/temporal-cli/headers"
	"github.com/temporalio/temporal-cli/namespace"
	"github.com/temporalio/temporal-cli/schedule"
	"github.com/temporalio/temporal-cli/searchattribute"
	"github.com/temporalio/temporal-cli/server"
	sconfig "github.com/temporalio/temporal-cli/server/config"
	"github.com/temporalio/temporal-cli/taskqueue"
	"github.com/temporalio/temporal-cli/workflow"
	uiversion "github.com/temporalio/ui-server/v2/server/version"
	"github.com/urfave/cli/v2"
	sheaders "go.temporal.io/server/common/headers"
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite" // load sqlite storage driver
)

func BuildApp(version string) *cli.App {
	defaultCfg, _ := sconfig.NewDefaultConfig()

	app := cli.NewApp()
	app.Name = "temporal"
	app.Usage = "Temporal command-line interface and development server"
	app.Suggest = true
	if version == "" {
		version = headers.CLIVersion
	}
	app.Version = fmt.Sprintf("%s (server %s) (ui %s)", version, sheaders.ServerVersion, uiversion.UIVersion)
	app.Commands = commands(defaultCfg)
	app.Before = configureCLI
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
	client.Build(ctx)

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
		}}, common.WithFlags(clientCommands, common.SharedFlags)...)
}

var clientCommands = []*cli.Command{
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
}
