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
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl/pkg/process"
)

var (
	headerColor = tablewriter.Colors{tablewriter.FgHiBlueColor}
)

func PrintTable(c *cli.Context, items []interface{}, opts *PrintOptions) {
	fields := opts.Fields
	table := tablewriter.NewWriter(opts.Pager)
	table.SetBorder(false)
	table.SetColumnSeparator(opts.Separator)

	if !opts.NoHeader {
		table.SetHeader(fields)
		headerColors := make([]tablewriter.Colors, len(fields))
		for i := range headerColors {
			headerColors[i] = headerColor
		}
		table.SetHeaderColor(headerColors...)
		table.SetHeaderLine(false)
	}

	rows, err := extractFieldValues(items, fields)
	if err != nil {
		process.ErrorAndExit("unable to print table", err)
	}

	for _, row := range rows {
		columns := make([]string, len(row))
		for j, column := range row {
			columns[j] = formatField(c, column)
		}
		table.Append(columns)
	}
	table.Render()
	table.ClearRows()
}
