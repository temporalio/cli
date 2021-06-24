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

package format

import (
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

func PrintTable(c *cli.Context, items []interface{}, opts *PrintOptions) {
	fields := opts.Fields
	if len(fields) == 0 {
		// dynamically examine fields
		if len(items) == 0 {
			return
		}
		e := reflect.ValueOf(items[0])
		for e.Type().Kind() == reflect.Ptr {
			e = e.Elem()
		}
		t := e.Type()
		for i := 0; i < e.NumField(); i++ {
			fields = append(fields, t.Field(i).Name)
		}
	}

	table := tablewriter.NewWriter(opts.Pager)
	table.SetBorder(false)
	table.SetColumnSeparator(opts.Separator)
	if opts.Header {
		table.SetHeader(fields)
	}
	table.SetHeaderLine(false)

	for _, item := range items {
		val := reflect.ValueOf(item)
		var columns []string
		for _, field := range fields {
			nestedFields := strings.Split(field, ".") // results in ex. "Execution", "RunId"
			var col interface{}
			for _, nField := range nestedFields {
				for val.Type().Kind() == reflect.Ptr {
					// we want the struct value to be able to get a field by name
					val = val.Elem()
				}
				val = val.FieldByName(nField)
				col = val.Interface()
				val = reflect.ValueOf(col)
			}
			columns = append(columns, formatField(c, col))
			val = reflect.ValueOf(item)
		}
		table.Append(columns)
	}
	table.Render()
	table.ClearRows()
}
