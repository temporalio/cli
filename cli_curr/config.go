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
	"strings"

	"github.com/urfave/cli"
)

func newConfigCommands() []cli.Command {
	return []cli.Command{
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

var (
	validKeys = []string{
		"version",
	}
)

func GetValue(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("invalid number of args, expected 1: property name")
	}

	key := c.Args().Get(0)

	if err := validateKey(key); err != nil {
		return fmt.Errorf("unable to get property %v. %s", key, err)
	}

	val, err := tctlConfig.Get(nil, key)
	if err != nil {
		return fmt.Errorf("unable to get property %v. %s", key, err)
	}
	fmt.Printf("%v: %v\n", key, val)
	return nil
}

func SetValue(c *cli.Context) error {
	if c.NArg() != 2 {
		return errors.New("invalid number of args, expected 2: property and value")
	}

	key := c.Args().Get(0)
	val := c.Args().Get(1)

	if err := validateKey(key); err != nil {
		return fmt.Errorf("unable to set property %v. %s", key, err)
	}

	if err := tctlConfig.Set(nil, key, val); err != nil {
		return fmt.Errorf("unable to set property %v. %s", key, err)
	}

	fmt.Printf("%v: %v\n", key, val)
	return nil
}

func validateKey(key string) error {
	// in composite keys such as alias.mycommand, the first part before dot is configuration property name
	// second part is custom value
	key = strings.Split(key, ".")[0]

	for _, k := range validKeys {
		if strings.Compare(key, k) == 0 {
			return nil
		}
	}

	return fmt.Errorf("unknown key %v", key)
}
