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
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"

	"github.com/temporalio/shared-go/timestamp"
)

const (
	FlagFormatTime = "format-time"
)

type FormatTimeOption string

const (
	Relative FormatTimeOption = "relative"
	ISO      FormatTimeOption = "iso"
	Raw      FormatTimeOption = "raw"
)

func FormatTime(c *cli.Context, val time.Time) string {
	formatFlag := c.String(FlagFormatTime)

	timeVal := timestamp.TimeValue(&val)
	format := FormatTimeOption(formatFlag)
	switch format {
	case ISO:
		return timeVal.Format(time.RFC3339)
	case Raw:
		return fmt.Sprintf("%v", timeVal)
	case Relative:
		return humanize.Time(timeVal)
	default:
		return humanize.Time(timeVal)
	}
}
