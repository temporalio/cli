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

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/temporalio/tctl/pkg/config"
)

func useAliasCommands(app *cli.App) {
	aliases, err := config.GetSequence("alias")
	if err != nil {
		fmt.Print(err)
		return
	}

	app.CommandNotFound = func(ctx *cli.Context, cmdToFind string) {
		found := false
		for aliasCmd, aliasVal := range aliases {
			if strings.Compare(cmdToFind, aliasCmd) == 0 {
				found = true

				passedArgs := ctx.Args().Slice()
				aliasArgs := strings.Split(aliasVal, " ")
				args := append(passedArgs, aliasArgs...)

				app.Run(args)
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "%s is not a command. See '%s --help\n'", cmdToFind, ctx.App.Name)
		}
	}
}
