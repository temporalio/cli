// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Copyright (c) 2021 Datadog, Inc.
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

package liteconfig

import (
	"fmt"
	"net"
)

// Modified from https://github.com/phayes/freeport/blob/95f893ade6f232a5f1511d61735d89b1ae2df543/freeport.go

func NewPortProvider() *PortProvider {
	return &PortProvider{}
}

type PortProvider struct {
	listeners []*net.TCPListener
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func (p *PortProvider) GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		if addr, err = net.ResolveTCPAddr("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("temporal: failed to get free port: %v", err))
		}
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	p.listeners = append(p.listeners, l)

	return l.Addr().(*net.TCPAddr).Port, nil
}

func (p *PortProvider) MustGetFreePort() int {
	port, err := p.GetFreePort()
	if err != nil {
		panic(err)
	}
	return port
}

func (p *PortProvider) Close() error {
	for _, l := range p.listeners {
		if err := l.Close(); err != nil {
			return err
		}
	}
	return nil
}
