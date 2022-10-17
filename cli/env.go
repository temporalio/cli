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
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/config"
	"github.com/temporalio/tctl-kit/pkg/output"
)

func newEnvCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "current-env",
			Usage: "Print the current environment name",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return CurrentEnv(c)
			},
		},
		{
			Name:      "show-env",
			Usage:     "Print environment properties",
			ArgsUsage: "env_name",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    output.FlagOutput,
					Aliases: FlagOutputAlias,
					Usage:   output.UsageText,
				},
			},
			Action: func(c *cli.Context) error {
				return ShowEnv(c)
			},
		},
		{
			Name:      "use-env",
			Usage:     "Switch environment",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name",
			Action: func(c *cli.Context) error {
				return UseEnv(c)
			},
		},
		{
			Name:      "remove-env",
			Usage:     "Remove environment",
			Flags:     []cli.Flag{},
			ArgsUsage: "env_name",
			Action: func(c *cli.Context) error {
				return RemoveEnv(c)
			},
		},
	}
}

func CurrentEnv(c *cli.Context) error {
	fmt.Println(tctlConfig.CurrentEnv)

	return nil
}

func ShowEnv(c *cli.Context) error {
	envName := c.Args().Get(0)

	if envName == "" {
		envName = tctlConfig.CurrentEnv
	}

	env := tctlConfig.Env(envName)

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

func UseEnv(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("env name is required")
	}

	envName := c.Args().Get(0)

	if err := tctlConfig.SetCurrentEnv(envName); err != nil {
		return fmt.Errorf("unable to set property %s: %w", config.KeyCurrentEnvironment, err)
	}

	fmt.Printf("%v: %v\n", color.Magenta(c, "%v", config.KeyCurrentEnvironment), envName)

	return nil
}

func RemoveEnv(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("env name is required")
	}

	envName := c.Args().Get(0)

	if err := tctlConfig.RemoveEnv(envName); err != nil {
		return fmt.Errorf("unable to remove env %s: %w", envName, err)
	}

	fmt.Printf("Removed env %v\n", color.Magenta(c, "%v", envName))

	return nil
}
