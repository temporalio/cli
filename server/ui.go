// Unless explicitly stated otherwise all files in this repository are licensed under the MIT License.
//
// This product includes software developed at Datadog (https://www.datadoghq.com/). Copyright 2021 Datadog, Inc.

//go:build !headless

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

// Name of the ui-server module, used in tests to verify that it is included/excluded
// as a dependency when building with the `headless` tag enabled.
const UIServerModule = "github.com/temporalio/ui-server/v2"

func newUIOption(frontendAddr string, uiIP string, uiPort int, configDir string) (ServerOption, error) {
	cfg, err := newUIConfig(
		frontendAddr,
		uiIP,
		uiPort,
		configDir,
	)
	if err != nil {
		return nil, err
	}
	return WithUI(uiserver.NewServer(uiserveroptions.WithConfigProvider(cfg))), nil
}

func newUIConfig(frontendAddr string, uiIP string, uiPort int, configDir string) (*uiconfig.Config, error) {
	cfg := &uiconfig.Config{
		Host: uiIP,
		Port: uiPort,
	}
	if configDir != "" {
		if err := provider.Load(configDir, cfg, "temporal-ui"); err != nil {
			if !strings.HasPrefix(err.Error(), "no config files found") {
				return nil, err
			}
		}
	}
	cfg.TemporalGRPCAddress = frontendAddr
	cfg.EnableUI = true
	return cfg, nil
}
