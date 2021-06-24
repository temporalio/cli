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

package pager

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

const (
	DefaultListPageSize = 20
)

func NewPager(c *cli.Context, defaultPager string) (io.Writer, func()) {
	noPager := c.Bool(FlagNoPager)
	if noPager {
		return os.Stdout, func() {}
	}

	pager, err := lookupPager(c, defaultPager)
	if err != nil {
		return os.Stdout, func() {}
	}

	exe, _ := exec.LookPath(pager)
	cmd := exec.Command(exe)

	if pager == string(Less) {
		env := os.Environ()
		env = append(env, "LESS=FRX")
		cmd.Env = env
	}

	signal.Ignore(syscall.SIGPIPE)

	reader, writer := io.Pipe()
	cmd.Stdin = reader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	done := make(chan struct{})
	go func() {
		defer close(done)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}()

	return writer, func() {
		writer.Close()
		<-done
	}
}

func lookupPager(c *cli.Context, defaultPager string) (string, error) {
	pagerFlag := c.String("pager")
	if pagerFlag == "" {
		pagerFlag = defaultPager
	}

	if pagerFlag != "" {
		if pager, err := exec.LookPath(pagerFlag); err == nil {
			return pager, nil
		}
	}

	if pager, err := exec.LookPath(string(Less)); err == nil {
		return pager, nil
	}

	if pager, err := exec.LookPath(string(More)); err == nil {
		return pager, nil
	}

	if pager, err := exec.LookPath(string(Cat)); err == nil {
		return pager, nil
	}

	return "", errors.New("no pager available. Set $PAGER env variable or install 'less', 'more' or 'cat'")
}
