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

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/config"
)

func newConfigCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "get",
			Usage: "Print config values",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return GetValue(c)
			},
		},
		{
			Name:  "set",
			Usage: "Set config values",
			Flags: []cli.Flag{},

			Action: func(c *cli.Context) error {
				return SetValue(c)
			},
		},
	}
}

func GetValue(c *cli.Context) error {
	for i := 0; i < c.Args().Len(); i++ {
		key := c.Args().Get(i)
		val, err := tctlConfig.Get(key)

		if err != nil {
			return err
		}

		fmt.Printf("%v: %v\n", color.Magenta(c, "%v", key), val)
	}

	return nil
}

func SetValue(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("no key specified")
	}

	key := c.Args().Get(0)

	var val string
	if c.Args().Len() > 1 {
		val = c.Args().Get(1)
	}

	if err := tctlConfig.Set(key, val); err != nil {
		return fmt.Errorf("unable to set property %s: %s", key, err)
	}

	fmt.Printf("%v: %v\n", color.Magenta(c, "%v", key), val)

	return nil
}

func newAliasCommand() *cli.Command {
	return &cli.Command{
		Name:  "alias",
		Usage: "Create an alias for command",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "command",
				Usage:    "New command name",
				Required: true,
				Value:    "mycommand",
			},
			&cli.StringFlag{
				Name:     "alias",
				Usage:    "Alias for command",
				Required: true,
				Value:    "workflow list --output json",
			},
		},
		Action: func(c *cli.Context) error {
			return createAlias(c)
		},
	}
}

func createAlias(c *cli.Context) error {
	command := c.String("command")
	alias := c.String("alias")

	fullKey := fmt.Sprintf("%s.%s", config.KeyAliases, command)

	if err := tctlConfig.Set(fullKey, alias); err != nil {
		return fmt.Errorf("unable to set property %s: %s", config.KeyAliases, err)
	}

	fmt.Printf("%v: %v\n", color.Magenta(c, "%v", config.KeyAliases), alias)
	return nil
}

func populateFlags(commands []*cli.Command, globalFlags []cli.Flag) {
	for _, command := range commands {
		populateFlags(command.Subcommands, globalFlags)
		command.Before = populateFlagsFunc(command, globalFlags)
	}
}

func populateFlagsFunc(command *cli.Command, globalFlags []cli.Flag) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		flags := append(command.Flags, globalFlags...)
		for _, flag := range flags {
			name := flag.Names()[0]

			for _, c := range ctx.Lineage() {
				if !c.IsSet(name) {
					value, _ := tctlConfig.GetByCurrentEnvironment(name)
					if value != "" {
						c.Set(name, value)
					}
				}
			}
		}

		return nil
	}
}
