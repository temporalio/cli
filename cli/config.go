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
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl-kit/pkg/config"
)

func newConfigCommands() []*cli.Command {
	return append([]*cli.Command{
		{
			Name:      "get",
			Usage:     "Print the value of an env property",
			Flags:     []cli.Flag{},
			ArgsUsage: "[env.env_name.]property_name",
			Action: func(c *cli.Context) error {
				return EnvProperty(c)
			},
		},
		{
			Name:      "set",
			Usage:     "Set the value of an env property",
			Flags:     []cli.Flag{},
			ArgsUsage: "[env.env_name.]property_name value",
			Action: func(c *cli.Context) error {
				return SetEnvProperty(c)
			},
		},
	}, newEnvCommands()...)
}

func EnvProperty(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("invalid number of args, expected 1: env property name")
	}

	fullKey := c.Args().Get(0)

	if err := validateEnvKey(fullKey); err != nil {
		return err
	}

	env, key := envKey(fullKey)

	val := tctlConfig.EnvProperty(env, key)
	fmt.Println(val)

	return nil
}

func SetEnvProperty(c *cli.Context) error {
	if c.NArg() != 2 {
		return errors.New("invalid number of args, expected 2: property and value")
	}

	fullKey := c.Args().Get(0)
	val := c.Args().Get(1)

	if fullKey == "version" {
		if err := tctlConfig.SetVersion(val); err != nil {
			return fmt.Errorf("unable to set version: %s", err)
		}

		return nil
	}

	if err := validateEnvKey(fullKey); err != nil {
		return err
	}

	env, key := envKey(fullKey)

	if err := tctlConfig.SetEnvProperty(env, key, val); err != nil {
		return fmt.Errorf("unable to set env property %v. %s", key, err)
	}

	fmt.Printf("Set '%v' to: %v\n", fullKey, val)
	return nil
}

func validateEnvKey(fullKey string) error {
	keys := strings.Split(fullKey, ".")

	if len(keys) != 1 && len(keys) != 3 {
		return fmt.Errorf("invalid env key %v. Env key must be in a format <env-name> or env.<env name>.<property-name>", fullKey)
	}

	if len(keys) == 3 && keys[0] != config.KeyEnvironment {
		return fmt.Errorf("invalid env key %v. Env key must be in a format <env-name> or env.<env name>.<property-name>", fullKey)
	}

	return nil
}

func envKey(fullKey string) (string, string) {
	keys := strings.Split(fullKey, ".")

	var env, key string
	if len(keys) == 1 {
		env = tctlConfig.CurrentEnv
		key = keys[0]
	} else if len(keys) == 3 {
		env = keys[1]
		key = keys[2]
	}

	return env, key
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
					value := tctlConfig.EnvProperty(tctlConfig.CurrentEnv, name)
					if value != "" {
						c.Set(name, value)
					}
				}
			}
		}

		return nil
	}
}
