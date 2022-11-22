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
			Usage:     "Print environment properties",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name or env_name.property_name",
			Action: func(c *cli.Context) error {
				return EnvProperty(c)
			},
		},
		{
			Name:      "set",
			Usage:     "Set environment property",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name.property_name value",
			Action: func(c *cli.Context) error {
				return SetEnvProperty(c)
			},
		},
		{
			Name:      "delete",
			Usage:     "Delete environment or environment property",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name or env_name.property_name",
			Action: func(c *cli.Context) error {
				return DeleteEnv(c)
			},
		},
	}
}

func Build(c *cli.Context) {
	ClientConfig, _ = NewClientConfig()

	for _, c := range c.App.Commands {
		common.AddBeforeHandler(c, loadEnv)
	}
}

func EnvProperty(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("invalid number of args, expected 1: env property name")
	}

	fullKey := c.Args().Get(0)

	if err := validateEnvArg(fullKey); err != nil {
		return err
	}

	envName, key := envKey(fullKey)

	type flag struct {
		Flag  string
		Value string
	}
	var flags []interface{}

	if key == "" {
		// print all env properties

		env := ClientConfig.Env(envName)

		for k, v := range env {
			flags = append(flags, flag{Flag: k, Value: v})
		}

	} else {
		// print specific env property
		val, err := ClientConfig.EnvProperty(envName, key)
		if err != nil {
			return err
		}

		flags = append(flags, flag{Flag: key, Value: val})
	}

	po := &output.PrintOptions{OutputFormat: output.Table}
	return output.PrintItems(c, flags, po)
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

	if err := validateEnvArg(fullKey); err != nil {
		return err
	}

	env, key := envKey(fullKey)

	if err := ClientConfig.SetEnvProperty(env, key, val); err != nil {
		return fmt.Errorf("unable to set env property %v: %w", key, err)
	}

	fmt.Printf("Set '%v' to: %v\n", fullKey, val)
	return nil
}

func validateEnvArg(fullArg string) error {
	keys := strings.Split(fullArg, ".")

	if len(keys) != 1 && len(keys) != 2 {
		return fmt.Errorf("invalid env argument %v. Env argument must be in a format <env name> or <env name>.<property-name>", fullArg)
	}

	return nil
}

func DeleteEnv(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("env name is required")
	}

	fullKey := c.Args().Get(0)

	if err := validateEnvArg(fullKey); err != nil {
		return err
	}

	envName, key := envKey(fullKey)

	if key == "" {
		if err := ClientConfig.RemoveEnv(envName); err != nil {
			return fmt.Errorf("unable to delete env %v: %w", envName, err)
		}
		fmt.Printf("Deleted env: %v\n", envName)
	} else {
		if err := ClientConfig.RemoveEnvProperty(envName, key); err != nil {
			return fmt.Errorf("unable to delete env property %v: %w", key, err)
		}
		fmt.Printf("Deleted env property: %v\n", fullKey)
	}

	return nil
}

func envKey(fullKey string) (string, string) {
	keys := strings.Split(fullKey, ".")

	var env, key string

	env = keys[0]

	if len(keys) == 2 {
		key = keys[1]
	}

	return env, key
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
