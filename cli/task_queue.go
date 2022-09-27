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
	"github.com/temporalio/tctl-kit/pkg/flags"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

func newTaskQueueCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "describe",
			Usage: "Describe the Workers that have recently polled on this Task Queue",
			Description: `The Server records the last time of each poll request. Poll requests can last up to a minute, so a LastAccessTime under a minute is normal. If it's over a minute, then likely either the Worker is at capacity (all Workflow and Activity slots are full) or it has shut down. Once it has been 5 minutes since the last poll request, the Worker is removed from the list.

RatePerSecond is the maximum Activities per second the Worker will execute.`,
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     FlagTaskQueue,
					Aliases:  FlagTaskQueueAlias,
					Usage:    "Task Queue name",
					Required: true,
				},
				&cli.StringFlag{
					Name:  FlagTaskQueueType,
					Value: "workflow",
					Usage: "Task Queue type [workflow|activity]",
				},
			}, flags.FlagsForRendering...),
			Action: func(c *cli.Context) error {
				return DescribeTaskQueue(c)
			},
		},
		{
			Name:  "list-partition",
			Usage: "List the Task Queue's partitions and which matching node they are assigned to",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     FlagTaskQueue,
					Aliases:  FlagTaskQueueAlias,
					Usage:    "Task Queue name",
					Required: true,
				},
				&cli.StringFlag{
					Name:    output.FlagOutput,
					Aliases: FlagOutputAlias,
					Usage:   output.UsageText,
					Value:   string(output.Table),
				},
			},
			Action: func(c *cli.Context) error {
				return ListTaskQueuePartitions(c)
			},
		},
	}
}
