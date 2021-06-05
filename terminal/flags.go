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
package terminal

import (
	"strconv"

	"github.com/urfave/cli/v2"
)

// General command line flags
const (
	FlagAll      = "all"
	FlagDetach   = "detach"
	FlagJSON     = "json"
	FlagPageSize = "pagesize"
)

var FlagsForPagination = []cli.Flag{
	&cli.BoolFlag{
		Name:  FlagAll,
		Usage: "List all pages",
	},
	&cli.BoolFlag{
		Name:  FlagDetach,
		Usage: "Detach after printing first results",
	},
	&cli.IntFlag{
		Name:        FlagPageSize,
		Value:       defaultPageSize,
		Usage:       "Items per page",
		DefaultText: strconv.Itoa(defaultPageSize),
	},
}

var FlagsForRendering = []cli.Flag{
	&cli.BoolFlag{
		Name:  FlagJSON,
		Usage: "Print in json format",
	},
}

var FlagsForPaginationAndRendering = append(FlagsForPagination, FlagsForRendering...)
