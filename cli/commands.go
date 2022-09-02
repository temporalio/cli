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
		Usage:       "Operations on namespaces",
		Subcommands: newNamespaceCommands(),
	},
	{
		Name:        "workflow",
		Usage:       "Operations on workflows",
		Subcommands: newWorkflowCommands(),
	},
	{
		Name:        "activity",
		Usage:       "Operations on activities of workflows",
		Subcommands: newActivityCommands(),
	},
	{
		Name:        "task-queue",
		Usage:       "Operations on task queues",
		Subcommands: newTaskQueueCommands(),
	},
	{
		Name:        "schedule",
		Usage:       "Operations on schedules",
		Subcommands: newScheduleCommands(),
	},
	{
		Name:        "search-attribute",
		Usage:       "Operations on search attributes",
		Subcommands: newSearchAttributeCommands(),
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
		Usage:       "Operations using a custom data converter",
		Subcommands: newDataConverterCommands(),
	},
	{
		Name:        "config",
		Usage:       "Configure tctl",
		Subcommands: newConfigCommands(),
	},
	{
		Name:        "alias",
		Usage:       "Create an alias for a command",
		Subcommands: newAliasCommand(),
	},
}
