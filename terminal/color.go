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
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var (
	colorGreen   = color.New(color.FgGreen).SprintfFunc()
	colorMagenta = color.New(color.FgMagenta).SprintfFunc()
	colorRed     = color.New(color.FgRed).SprintfFunc()
)

func Green(c *cli.Context, format string, a ...interface{}) string {
	checkNoColor(c)
	return colorGreen(format, a...)
}

func Magenta(c *cli.Context, format string, a ...interface{}) string {
	checkNoColor(c)
	return colorMagenta(format, a...)
}

func Yellow(c *cli.Context, format string, a ...interface{}) string {
	checkNoColor(c)
	return color.YellowString(format, a...)
}
func Red(c *cli.Context, format string, a ...interface{}) string {
	checkNoColor(c)
	return colorRed(format, a...)
}

func checkNoColor(c *cli.Context) {
	if c.Bool(FlagNoColor) {
		color.NoColor = true
	}
}
