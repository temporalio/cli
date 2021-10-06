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

	"github.com/temporalio/tctl/pkg/format"
	"github.com/temporalio/tctl/pkg/output"
	"github.com/temporalio/tctl/pkg/pager"
)

var FlagsForPagination = []cli.Flag{
	&cli.IntFlag{
		Name:  output.FlagLimit,
		Usage: "number of items to print",
	},
	&cli.StringFlag{
		Name:    pager.FlagPager,
		Usage:   "pager to use: less, more, favoritePager..",
		EnvVars: []string{"PAGER"},
	},
	&cli.BoolFlag{
		Name:    pager.FlagNoPager,
		Aliases: []string{"P"},
		Usage:   "disable interactive pager",
	},
}

var FlagsForRendering = []cli.Flag{
	&cli.StringFlag{
		Name:    output.FlagOutput,
		Aliases: []string{"o"},
		Usage:   output.UsageText,
		Value:   string(output.Table),
	},
	&cli.StringFlag{
		Name:  format.FlagTimeFormat,
		Usage: fmt.Sprintf("format time as: %v, %v, %v.", format.Relative, format.ISO, format.Raw),
		Value: string(format.Relative),
	},
	&cli.StringFlag{
		Name:  output.FlagFields,
		Usage: "customize fields to print. Set to 'long' to automatically print more of main fields",
	},
}

var FlagsForPaginationAndRendering = append(FlagsForPagination, FlagsForRendering...)
