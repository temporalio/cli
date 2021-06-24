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
package flags

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl/terminal/color"
	"github.com/temporalio/tctl/terminal/format"
	"github.com/temporalio/tctl/terminal/pager"
	"github.com/temporalio/tctl/terminal/timeformat"
)

var FlagsForPagination = []cli.Flag{
	&cli.BoolFlag{
		Name:  pager.FlagNoPager,
		Usage: "disable interactive paging",
	},
	&cli.StringFlag{
		Name:    pager.FlagPager,
		Usage:   "pager to use: cat, less, favoritePager..",
		EnvVars: []string{"PAGER"},
	},
	&cli.IntFlag{
		Name:  pager.FlagPageSize,
		Value: pager.DefaultListPageSize,
		Usage: "items per page",
	},
}

var FlagsForRendering = []cli.Flag{
	&cli.StringFlag{
		Name:    format.FlagFormat,
		Aliases: []string{"f"},
		Usage:   fmt.Sprintf("format output as: %v, %v, %v.", format.Table, format.JSON, format.Card),
		Value:   string(format.Table),
	},
	&cli.StringFlag{
		Name:  timeformat.FlagTimeFormat,
		Usage: fmt.Sprintf("format time as: %v, %v, %v.", timeformat.Relative, timeformat.ISO, timeformat.Raw),
		Value: string(timeformat.Relative),
	},
	&cli.StringFlag{
		Name:    color.FlagColor,
		Aliases: []string{"c"},
		Usage:   fmt.Sprintf("when to use color: %v, %v, %v.", color.Auto, color.Always, color.Never),
		Value:   string(color.Auto),
	},
}

var FlagsForPaginationAndRendering = append(FlagsForPagination, FlagsForRendering...)
