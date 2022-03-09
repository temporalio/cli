// The MIT License
//
// Copyright (c) 2021 Temporal Technologies Inc.  All rights reserved
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
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

func useDynamicCommands(app *cli.App) {
	app.CommandNotFound = func(ctx *cli.Context, cmdToFind string) {
		// try execute as an alias command
		cmdValue, err := lookupCmdInAliasCommands(ctx, cmdToFind)
		if err == nil {
			_ = executeAliasCommand(ctx, app, cmdValue)
			return
		}

		// execute external binary by path
		pluginName := "tctl-" + cmdToFind
		path, err := exec.LookPath(pluginName)
		if err == nil {
			os.Args = append([]string{pluginName}, os.Args[2:]...)
			_ = executePlugin(ctx, path, os.Args, os.Environ())
		}

		fmt.Fprintf(os.Stderr, "%s is not a command. See '%s --help\n'", cmdToFind, ctx.App.Name)
	}
}

// looks up an alias command in tctl config and returns its value
func lookupCmdInAliasCommands(ctx *cli.Context, cmd string) (string, error) {
	aliases, err := tctlConfig.GetAliases()
	if err != nil {
		return "", err
	}

	for aliasCmd, aliasVal := range aliases {
		if cmd == aliasCmd {
			return aliasVal, nil
		}
	}

	return "", fmt.Errorf("alias command %s not found", cmd)
}

func executeAliasCommand(ctx *cli.Context, app *cli.App, aliasVal string) error {
	passedArgs := ctx.Args().Slice()
	aliasArgs := strings.Split(aliasVal, " ")
	args := append(passedArgs, aliasArgs...)

	return app.Run(args)
}
