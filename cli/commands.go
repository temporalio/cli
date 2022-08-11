// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
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
	"github.com/urfave/cli/v2"
)

var tctlCommands = []*cli.Command{
	{
		Name:        "namespace",
		Aliases:     []string{"n"},
		Usage:       "Operations on namespaces",
		Subcommands: newNamespaceCommands(),
	},
	{
		Name:        "workflow",
		Aliases:     []string{"w"},
		Usage:       "Operations on workflows",
		Subcommands: newWorkflowCommands(),
	},
	{
		Name:        "activity",
		Aliases:     []string{"a"},
		Usage:       "Operations on activities of workflows",
		Subcommands: newActivityCommands(),
	},
	{
		Name:        "task-queue",
		Aliases:     []string{"tq"},
		Usage:       "Operations on task queues",
		Subcommands: newTaskQueueCommands(),
	},
	{
		Name:        "schedule",
		Aliases:     []string{"s"},
		Usage:       "Operations on schedules",
		Subcommands: newScheduleCommands(),
	},
	{
		Name:        "batch",
		Usage:       "Batch operations on a list of workflows from a query",
		Subcommands: newBatchCommands(),
	},
	{
		Name:        "cluster",
		Usage:       "Operations on a Temporal cluster",
		Subcommands: newClusterCommands(),
	},
	{
		Name:        "data-converter",
		Aliases:     []string{"dc"},
		Usage:       "Operations using a custom data converter",
		Subcommands: newDataConverterCommands(),
	},
	{
		Name:        "config",
		Aliases:     []string{"c"},
		Usage:       "Configure tctl",
		Subcommands: newConfigCommands(),
	},
	newAliasCommand(),
}

var sharedFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    FlagNamespace,
		Aliases: FlagNamespaceAlias,
		Value:   "default",
		Usage:   "Namespace to operate on",
		EnvVars: []string{"TEMPORAL_CLI_NAMESPACE"},
	},
}
