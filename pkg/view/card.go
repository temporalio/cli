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

package view

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl/pkg/color"
	"github.com/temporalio/tctl/pkg/process"
)

func PrintCards(c *cli.Context, items []interface{}, opts *PrintOptions) {
	rows, err := extractFieldValues(items, opts.Fields)
	if err != nil {
		process.ErrorAndExit("unable to print card", err)
	}

	if opts.Pager == nil {
		pager, close := newPagerWithDefault(c)
		opts.Pager = pager
		defer close()
	}

	fieldNames := make([]string, len(opts.Fields))
	for i, field := range opts.Fields {
		nestedFields := splitFieldPath(field)
		fieldNames[i] = nestedFields[len(nestedFields)-1]
	}

	w := opts.Pager
	for _, row := range rows {
		fmt.Fprintf(w, "---------------------------------------------------\n")
		for j, col := range row {
			field := fieldNames[j]
			val := formatField(c, col)
			fmt.Fprintf(w, "%v \t\t%v\n", color.Magenta(c, field), val)
		}
	}
}
