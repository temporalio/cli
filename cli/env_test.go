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

const (
	testEnvName = "tctl-test-env"
)

func ExampleCurrentEnv() {
	tctl := NewCliApp()

	tctl.Run([]string{"", "config", "use-env", testEnvName})
	tctl.Run([]string{"", "config", "current-env"})

	// Output:
	// current-env: tctl-test-env
	// tctl-test-env
}

func (s *cliAppSuite) TestUseEnv() {
	err := s.app.Run([]string{"", "config", "use-env", testEnvName})
	s.NoError(err)

	config := readConfig()
	s.Contains(config, "current-env: "+testEnvName)
}

func ExampleShowEnv() {
	tctl := NewCliApp()

	tctl.Run([]string{"", "config", "use-env", testEnvName})
	tctl.Run([]string{"", "config", "set", "namespace", "tctl-test-namespace"})

	tctl.Run([]string{"", "config", "show-env", testEnvName})
	// Output:
	// current-env: tctl-test-env
	// Set 'namespace' to: tctl-test-namespace
	//   namespace  tctl-test-namespace
}
