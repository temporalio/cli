// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
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

package cli

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/config"
)

func newAliasCommand() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "set",
			Usage: "Create an alias for command",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return SetAlias(c)
			},
		}}
}

func SetAlias(c *cli.Context) error {
	if c.NArg() != 2 {
		return errors.New("invalid number of args, expected 2: property and value")
	}

	name := c.Args().Get(0)
	value := c.Args().Get(1)

	if err := tctlConfig.SetAlias(name, value); err != nil {
		return fmt.Errorf("unable to set property %s: %w", config.KeyAliases, err)
	}

	fmt.Printf("%v: %v\n", color.Magenta(c, "%v", config.KeyAliases), value)
	return nil
}
