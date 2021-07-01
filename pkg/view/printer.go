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
	"io"
	"time"

	"github.com/temporalio/shared-go/collection"
	"github.com/temporalio/tctl/pkg/format"
	"github.com/temporalio/tctl/pkg/pager"
	"github.com/urfave/cli/v2"
)

const (
	DefaultListPageSize = 20
)

type PrintOptions struct {
	Fields    []string
	NoHeader  bool
	Separator string
	All       bool
	Pager     io.Writer
}

func PrintItems(c *cli.Context, items []interface{}, opts *PrintOptions) {
	outputFlag := c.String(FlagOutput)
	output := OutputOption(outputFlag)

	if opts.Pager == nil {
		pager, close := newPagerWithDefault(c)
		opts.Pager = pager
		defer close()
	}

	switch output {
	case Table:
		PrintTable(c, items, opts)
	case JSON:
		PrintJSON(items, opts)
	case Card:
		PrintCards(c, items, opts)
	default:
	}
}

// Paginate creates an interactive CLI mode to control the printing of items
func Paginate(c *cli.Context, iter collection.Iterator, opts *PrintOptions) error {
	noPager := c.Bool(pager.FlagNoPager)
	pageSize := c.Int(pager.FlagPageSize)

	pager, close := newPagerWithDefault(c)
	defer close()

	if opts == nil {
		opts = &PrintOptions{}
	}
	opts.Pager = pager

	var pageItems []interface{}
	for iter.HasNext() {
		item, err := iter.Next()
		if err != nil {
			return err
		}

		pageItems = append(pageItems, item)
		shouldPrintPage := len(pageItems) == pageSize || !iter.HasNext()
		if shouldPrintPage {
			PrintItems(c, pageItems, opts)
			pageItems = pageItems[:0]
			opts.NoHeader = true
			if noPager {
				break
			}
		}
	}

	return nil
}

func newPagerWithDefault(c *cli.Context) (io.Writer, func()) {
	outputFlag := c.String(FlagOutput)
	output := OutputOption(outputFlag)

	var defaultPager string
	if output == Table {
		defaultPager = string(pager.Less)
	} else {
		defaultPager = string(pager.More)
	}
	return pager.NewPager(c, defaultPager)
}

func formatField(c *cli.Context, i interface{}) string {
	switch v := i.(type) {
	case time.Time:
		return format.FormatTime(c, &v)
	case *time.Time:
		return format.FormatTime(c, v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
