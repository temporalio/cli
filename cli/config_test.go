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
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func ExampleEnvProperty() {
	tctl := NewCliApp()
	defer setupConfig(tctl)()

	tctl.Run([]string{"", "config", "set", "namespace", "tctl-test-namespace"})

	tctl.Run([]string{"", "config", "get", "namespace"})

	// Output:
	// current-env: tctl-test-env
	// Set 'namespace' to: tctl-test-namespace
	// tctl-test-namespace
	// current-env: local
	// Removed env tctl-test-env
}

func (s *cliAppSuite) TestSetConfigValue() {
	defer setupConfig(s.app)()

	err := s.app.Run([]string{"", "config", "set", "namespace", "tctl-test-namespace"})
	s.NoError(err)
	err = s.app.Run([]string{"", "config", "set", "address", "0.0.0.0:00000"})
	s.NoError(err)

	config := readConfig()
	s.Contains(config, "    tctl-test-env:")
	s.Contains(config, "        address: 0.0.0.0:00000")
	s.Contains(config, "        namespace: tctl-test-namespace")
}

func setupConfig(app *cli.App) func() {
	err := app.Run([]string{"", "config", "use-env", testEnvName})
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		err = app.Run([]string{"", "config", "use-env", "local"})
		if err != nil {
			log.Fatal(err)
		}
		err = app.Run([]string{"", "config", "remove-env", testEnvName})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func readConfig() string {
	path := getConfigPath()
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func getConfigPath() string {
	dpath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(dpath, ".config", "temporalio", "tctl.yaml")

	return path
}
