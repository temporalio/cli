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

package output

import (
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/hokaccha/go-prettyjson"
	"github.com/temporalio/shared-go/codec"
	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl/pkg/color"
)

func PrintJSON(c *cli.Context, o interface{}, opts *PrintOptions) {
	json, err := ParseToJSON(c, o, true)

	if err != nil {
		fmt.Printf("Error when try to print pretty: %v\n", err)
		fmt.Fprintln(opts.Pager, o)
	}

	fmt.Fprintln(opts.Pager, json)
}

func ParseToJSON(c *cli.Context, o interface{}, indent bool) (string, error) {
	colorFlag := c.String(color.FlagColor)
	enableColor := colorFlag == string(color.Auto) || colorFlag == string(color.Always)
	var b []byte
	var err error

	if enableColor {
		encoder := prettyjson.NewFormatter()
		if !indent {
			encoder.Indent = 0
			encoder.Newline = ""
		}
		b, err = encoder.Marshal(o)
	} else {
		if pb, ok := o.(proto.Message); ok {
			var encoder *codec.JSONPBEncoder
			if indent {
				encoder = codec.NewJSONPBIndentEncoder("  ")
			} else {
				encoder = codec.NewJSONPBEncoder()
			}
			b, err = encoder.Encode(pb)
		} else {
			if indent {
				b, err = json.MarshalIndent(o, "", "  ")
			} else {
				b, err = json.Marshal(o)
			}
		}
	}

	if err != nil {
		return "", err
	}

	return string(b), nil
}
