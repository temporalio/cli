// The MIT License
//
// Copyright (c) 2023 Temporal Technologies Inc.  All rights reserved.
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

package app_test

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerInterruptRC(t *testing.T) {
	// build and run the binary, not "go run" as it modifies the exit code
	build := exec.Command("go", "build", "-o", "./test-temporal", "../cmd/temporal")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	err := build.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./test-temporal")

	cmd := exec.Command("./test-temporal", "server", "start-dev", "--headless")
	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	// ensure the dev server process is killed, even if SIGTERM fails
	defer cmd.Process.Signal(syscall.SIGKILL)

	// Wait for the app to start
	time.Sleep(time.Second * 2)

	cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Wait()

	assert.Equal(t, 0, cmd.ProcessState.ExitCode())
}
