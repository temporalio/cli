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

package batch

import (
	"github.com/temporalio/temporal-cli/common"
	"github.com/urfave/cli/v2"
)

func NewBatchCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "describe",
			Usage: "Describe a batch operation job",
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagJobID,
					Usage:    "Batch Job Id",
					Required: true,
					Category: common.CategoryMain,
				},
			}, common.FlagsForFormatting...),
			Action: func(c *cli.Context) error {
				return DescribeBatchJob(c)
			},
		},
		{
			Name:      "list",
			Usage:     "List batch operation jobs",
			Flags:     common.FlagsForPaginationAndRendering,
			ArgsUsage: " ",
			Action: func(c *cli.Context) error {
				return ListBatchJobs(c)
			},
		},
		{
			Name:  "terminate",
			Usage: "Stop a batch operation job",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     common.FlagJobID,
					Usage:    "Batch Job Id",
					Required: true,
					Category: common.CategoryMain,
				},
				&cli.StringFlag{
					Name:     common.FlagReason,
					Usage:    "Reason to stop the batch job",
					Required: true,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return StopBatchJob(c)
			},
		},
	}
}
