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

package server

import (
	"testing"

	uiconfig "github.com/temporalio/ui-server/v2/server/config"
)

func TestNewUIConfig(t *testing.T) {
	c := &uiconfig.Config{
		Host:                "localhost",
		Port:                8233,
		TemporalGRPCAddress: "localhost:7233",
		EnableUI:            true,
	}
	cfg, err := MergeWithConfigFile(c, "")
	if err != nil {
		t.Errorf("cannot create config: %s", err)
		return
	}
	if err = cfg.Validate(); err != nil {
		t.Errorf("config not valid: %s", err)
	}
}

func TestNewUIConfigWithMissingConfigFile(t *testing.T) {
	c := &uiconfig.Config{
		Host:                "localhost",
		Port:                8233,
		TemporalGRPCAddress: "localhost:7233",
		EnableUI:            true,
	}
	cfg, err := MergeWithConfigFile(c, "wibble")
	if err != nil {
		t.Errorf("cannot create config: %s", err)
		return
	}
	if err = cfg.Validate(); err != nil {
		t.Errorf("config not valid: %s", err)
	}
}

func TestNewUIConfigWithPresentConfigFile(t *testing.T) {
	c := &uiconfig.Config{
		Host:                "localhost",
		Port:                8233,
		TemporalGRPCAddress: "localhost:7233",
		EnableUI:            true,
	}
	cfg, err := MergeWithConfigFile(c, "testdata")
	if err != nil {
		t.Errorf("cannot create config: %s", err)
		return
	}
	if err = cfg.Validate(); err != nil {
		t.Errorf("config not valid: %s", err)
	}
	if cfg.TLS.ServerName != "local.dev" {
		t.Errorf("did not load expected config file")
	}
}
