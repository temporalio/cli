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

	"github.com/temporalio/tctl/pkg/color"
	"github.com/temporalio/tctl/pkg/output"
	"github.com/urfave/cli/v2"
)

func newClusterCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:    "health",
			Aliases: []string{"h"},
			Usage:   "Check health of frontend service",
			Action: func(c *cli.Context) error {
				return HealthCheck(c)
			},
		},
		{
			Name:    "list-search-attributes",
			Usage:   "List search attributes that can be used in list workflow query",
			Aliases: []string{"gsa"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    output.FlagOutput,
					Aliases: []string{"o"},
					Usage:   output.UsageText,
					Value:   string(output.Table),
				},
				&cli.StringFlag{
					Name:  color.FlagColor,
					Usage: fmt.Sprintf("when to use color: %v, %v, %v.", color.Auto, color.Always, color.Never),
					Value: string(color.Auto),
				},
			},
			Action: func(c *cli.Context) error {
				return GetSearchAttributes(c)
			},
		},
	}
}
