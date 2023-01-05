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

// This file should be the only one to import ui-server packages.
// This is to avoid embedding the UI's static assets in the binary when the `headless` build tag is enabled.
import (
	"strings"

	provider "github.com/temporalio/ui-server/v2/plugins/fs_config_provider"
	uiserver "github.com/temporalio/ui-server/v2/server"
	uiconfig "github.com/temporalio/ui-server/v2/server/config"
	uiserveroptions "github.com/temporalio/ui-server/v2/server/server_options"
)

func newUIOption(c *uiconfig.Config, configDir string) (ServerOption, error) {
	cfg, err := MergeWithConfigFile(
		c,
		configDir,
	)
	if err != nil {
		return nil, err
	}
	return WithUI(uiserver.NewServer(uiserveroptions.WithConfigProvider(cfg))), nil
}

func MergeWithConfigFile(cfg *uiconfig.Config, configDir string) (*uiconfig.Config, error) {
	if configDir != "" {
		if err := provider.Load(configDir, cfg, "temporal-ui"); err != nil {
			if !strings.HasPrefix(err.Error(), "no config files found") {
				return nil, err
			}
		}
	}

	return cfg, nil
}
