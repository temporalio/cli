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
		"version",
	}
	envKeys = []string{
		FlagNamespace,
		FlagAddress,
		FlagTLSCertPath,
		FlagTLSKeyPath,
		FlagTLSCaPath,
		FlagTLSDisableHostVerification,
		FlagTLSServerName,
	}
	keys = append(rootKeys, envKeys...)
)

func newConfigCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "get",
			Usage: "Print config values",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  config.KeyActive,
					Usage: "Print active environment",
				},
				&cli.BoolFlag{
					Name:  config.KeyAlias,
					Usage: "Print command aliases",
				},
				&cli.BoolFlag{
					Name:  FlagNamespace,
					Usage: "Print default namespace (for active environment)",
				},
				&cli.BoolFlag{
					Name:  FlagAddress,
					Usage: "Print Temporal address (for active environment)",
				},
				&cli.StringFlag{
					Name:  FlagTLSCertPath,
					Value: "",
					Usage: "Print path to x509 certificate",
				},
				&cli.StringFlag{
					Name:  FlagTLSKeyPath,
					Value: "",
					Usage: "Print path to private key",
				},
				&cli.StringFlag{
					Name:  FlagTLSCaPath,
					Value: "",
					Usage: "Print path to server CA certificate",
				},
				&cli.BoolFlag{
					Name:  FlagTLSDisableHostVerification,
					Usage: "Print whether tls host name verification is disabled",
				},
				&cli.StringFlag{
					Name:  FlagTLSServerName,
					Value: "",
					Usage: "Print target TLS server name",
				},
			},
			Action: func(c *cli.Context) error {
				return GetValue(c)
			},
		},
		{
			Name:  "set",
			Usage: "Set config values",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  config.KeyActive,
					Usage: "Activate environment",
					Value: "local",
				},
				&cli.StringFlag{
					Name:  config.KeyAlias,
					Usage: "Create an alias command",
				},
				&cli.StringFlag{
					Name:  "version",
					Usage: "Opt-in to a new TCTL UX, values: (current, next)",
				},
				&cli.StringFlag{
					Name:  FlagAddress,
					Usage: "host:port for Temporal frontend service",
					Value: localHostPort,
				},
				&cli.StringFlag{
					Name:  FlagTLSCertPath,
					Value: "",
					Usage: "Path to x509 certificate",
				},
				&cli.StringFlag{
					Name:  FlagTLSKeyPath,
					Value: "",
					Usage: "Path to private key",
				},
				&cli.StringFlag{
					Name:  FlagTLSCaPath,
					Value: "",
					Usage: "Path to server CA certificate",
				},
				&cli.BoolFlag{
					Name:  FlagTLSDisableHostVerification,
					Usage: "Disable TLS host name verification (TLS must be enabled)",
				},
				&cli.StringFlag{
					Name:  FlagTLSServerName,
					Value: "",
					Usage: "Override for target server name",
				},
			},
			Action: func(c *cli.Context) error {
				return SetValue(c)
			},
		},
	}
}

func GetValue(c *cli.Context) error {
	for _, key := range keys {
		if !c.IsSet(key) {
			continue
		}

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
	}
	return nil
}

func SetValue(c *cli.Context) error {
	for _, key := range keys {
		if !c.IsSet(key) {
			continue
		}

		val := c.String(key)
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
	}

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

func createAlias(c *cli.Context) error {
	command := c.String("command")
	alias := c.String("alias")

	fullKey := fmt.Sprintf("%s.%s", config.KeyAlias, command)

	if err := tctlConfig.Set(c, fullKey, alias); err != nil {
		return fmt.Errorf("unable to set property %s: %s", config.KeyAlias, err)
	}

	fmt.Printf("%v: %v\n", color.Magenta(c, "%v", config.KeyAlias), alias)
	return nil
}
