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

package env

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/config"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/temporal-cli/common"
)

var (
	ClientConfig *config.Config
)

func NewEnvCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "get",
			Usage:     "Print the value of an env property",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name.property_name",
			Action: func(c *cli.Context) error {
				return EnvProperty(c)
			},
		},
		{
			Name:      "set",
			Usage:     "Set the value of an env property",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name.property_name value",
			Action: func(c *cli.Context) error {
				return SetEnvProperty(c)
			},
		},
		{
			Name:      "describe",
			Usage:     "Print environment properties",
			ArgsUsage: "env_name",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     output.FlagOutput,
					Aliases:  common.FlagOutputAlias,
					Usage:    output.UsageText,
					Category: common.CategoryFormatting,
				},
			},
			Action: func(c *cli.Context) error {
				return DescribeEnv(c)
			},
		},
		{
			Name:      "remove",
			Usage:     "Remove environment",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name",
			Action: func(c *cli.Context) error {
				return RemoveEnv(c)
			},
		},
	}
}

func DescribeEnv(c *cli.Context) error {
	envName := c.Args().Get(0)
	env := ClientConfig.Env(envName)

	type flag struct {
		Flag  string
		Value string
	}

	var flags []interface{}
	for k, v := range env {
		flags = append(flags, flag{Flag: k, Value: v})
	}

	po := &output.PrintOptions{OutputFormat: output.Table}
	return output.PrintItems(c, flags, po)
}

func RemoveEnv(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("env name is required")
	}

	envName := c.Args().Get(0)

	if err := ClientConfig.RemoveEnv(envName); err != nil {
		return fmt.Errorf("unable to remove env %s: %w", envName, err)
	}

	fmt.Printf("Removed env %v\n", color.Magenta(c, "%v", envName))

	return nil
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

	val, err := ClientConfig.EnvProperty(env, key)
	if err != nil {
		return err
	}

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
		if err := ClientConfig.SetVersion(val); err != nil {
			return fmt.Errorf("unable to set version: %w", err)
		}

		return nil
	}

	if err := validateEnvKey(fullKey); err != nil {
		return err
	}

	env, key := envKey(fullKey)

	if err := ClientConfig.SetEnvProperty(env, key, val); err != nil {
		return fmt.Errorf("unable to set env property %v: %w", key, err)
	}

	fmt.Printf("Set '%v' to: %v\n", fullKey, val)
	return nil
}

func validateEnvKey(fullKey string) error {
	keys := strings.Split(fullKey, ".")

	if len(keys) != 2 {
		return fmt.Errorf("invalid env key %v. Env key must be in a format <env name>.<property-name>", fullKey)
	}

	return nil
}

func envKey(fullKey string) (string, string) {
	keys := strings.Split(fullKey, ".")

	var env, key string
	if len(keys) == 2 {
		env = keys[0]
		key = keys[1]
	}

	return env, key
}

func Build(c *cli.Context) {
	ClientConfig, _ = NewClientConfig()

	for _, c := range c.App.Commands {
		common.AddBeforeHandler(c, loadEnv)
	}
}

// loadEnv loads environment options from the config file
func loadEnv(ctx *cli.Context) error {
	cmd := ctx.Command
	env := ctx.String(common.FlagEnv)

	if env == "" {
		return nil
	}

	for _, flag := range cmd.Flags {
		name := flag.Names()[0]

		for _, c := range ctx.Lineage() {
			if !c.IsSet(name) {
				value, err := ClientConfig.EnvProperty(env, name)
				if err != nil {
					return err
				}

				if value != "" {
					c.Set(name, value)
				}
			}
		}
	}

	return nil
}
