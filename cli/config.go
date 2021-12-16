// The MIT License
//
// Copyright (c) 2021 Temporal Technologies Inc.  All rights reserved.
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
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/config"
)

var (
	rootKeys = []string{
		config.KeyActive,
		config.KeyAlias,
		"version",
	}
	envKeys = []string{
		"namespace",
		"address",
	}
)

func newConfigCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "get",
			Usage: "get property",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return GetValue(c)
			},
		},
		{
			Name:  "set",
			Usage: "set property",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return SetValue(c)
			},
		},
	}
}

func GetValue(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("invalid number of args, expected 1: property name")
	}

	key := c.Args().Get(0)

	var val interface{}
	var err error
	if isRootKey(key) {
		val, err = tctlConfig.Get(c, key)
	} else if isEnvKey(key) {
		val, err = tctlConfig.GetByEnvironment(c, key)
	} else {
		return fmt.Errorf("invalid key: %s", key)
	}
	if err != nil {
		return err
	}

	fmt.Printf("%v: %v\n", color.Magenta(c, "%v", key), val)
	return nil
}

func SetValue(c *cli.Context) error {
	if c.NArg() != 2 {
		return fmt.Errorf("invalid number of args, expected 2: property and value")
	}

	key := c.Args().Get(0)
	val := c.Args().Get(1)

	if isRootKey(key) {
		if err := tctlConfig.Set(c, key, val); err != nil {
			return fmt.Errorf("unable to set property %s: %s", key, err)
		}
	} else if isEnvKey(key) {
		if err := tctlConfig.SetByEnvironment(c, key, val); err != nil {
			return fmt.Errorf("unable to set property %s: %s", key, err)
		}
	} else {
		return fmt.Errorf("unable to set property %s: invalid key", key)
	}

	fmt.Printf("%v: %v\n", color.Magenta(c, "%v", key), val)
	return nil
}

func isRootKey(key string) bool {
	for _, k := range rootKeys {
		if strings.Compare(key, k) == 0 {
			return true
		}
	}
	return false
}

func isEnvKey(key string) bool {
	for _, k := range envKeys {
		if strings.Compare(key, k) == 0 {
			return true
		}
	}
	return false
}
