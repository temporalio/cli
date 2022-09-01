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

package cli_curr

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"
)

func newConfigCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "get",
			Usage:     "get config property",
			Flags:     []cli.Flag{},
			ArgsUsage: "version",
			Action: func(c *cli.Context) error {
				return GetValue(c)
			},
		},
		{
			Name:      "set",
			Usage:     "set config property",
			Flags:     []cli.Flag{},
			ArgsUsage: "version 2",
			Action: func(c *cli.Context) error {
				return SetValue(c)
			},
		},
	}
}

func GetValue(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("invalid number of args, expected 1: env property name")
	}

	fullKey := c.Args().Get(0)

	if fullKey != "version" {
		return errors.New("only the version property is supported")
	}

	fmt.Println(tctlConfig.Version)

	return nil
}

func SetValue(c *cli.Context) error {
	if c.NArg() != 2 {
		return errors.New("invalid number of args, expected 2: property and value")
	}

	fullKey := c.Args().Get(0)
	val := c.Args().Get(1)

	if fullKey != "version" {
		return errors.New("only the version property is supported")
	}

	if err := tctlConfig.SetVersion(val); err != nil {
		return fmt.Errorf("unable to set version: %s", err)
	}

	return nil
}
