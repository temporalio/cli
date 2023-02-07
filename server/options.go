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
	sconfig "github.com/temporalio/cli/server/config"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/temporal"
)

// WithLogger overrides the default logger.
func WithLogger(logger log.Logger) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.Logger = logger
	})
}

// WithDatabaseFilePath persists state to the file at the specified path.
func WithDatabaseFilePath(filepath string) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.Ephemeral = false
		cfg.DatabaseFilePath = filepath
	})
}

// WithPersistenceDisabled disables file persistence and uses the in-memory storage driver.
// State will be reset on each process restart.
func WithPersistenceDisabled() ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.Ephemeral = true
	})
}

// WithCustomClusterID explicitly sets the cluster ID to use.
func WithCustomClusterID(clusterID string) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.ClusterID = clusterID
	})
}

// WithUI enables the Temporal web interface.
//
// When unspecified, Temporal will run in headless mode.
//
// This option accepts a UIServer implementation in order to avoid bloating
// programs that do not need to embed the UI.
// See ./cmd/temporal/main.go for an example of usage.
func WithUI(server sconfig.UIServer) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.UIServer = server
	})
}

// WithFrontendPort sets the listening port for the temporal-frontend GRPC service.
//
// When unspecified, the default port number of 7233 is used.
func WithFrontendPort(port int) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.FrontendPort = port
	})
}

// WithMetricsPort sets the listening port for metrics.
//
// When unspecified, the port will be system-chosen.
func WithMetricsPort(port int) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.MetricsPort = port
	})
}

// WithFrontendIP binds the temporal-frontend GRPC service to a specific IP (eg. `0.0.0.0`)
// Check net.ParseIP for supported syntax; only IPv4 is supported.
//
// When unspecified, the frontend service will bind to localhost.
func WithFrontendIP(address string) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.FrontendIP = address
	})
}

// WithDynamicPorts starts Temporal on system-chosen ports.
func WithDynamicPorts() ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.DynamicPorts = true
	})
}

// WithNamespaces registers each namespace on Temporal start.
func WithNamespaces(namespaces ...string) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.Namespaces = append(cfg.Namespaces, namespaces...)
	})
}

// WithSQLitePragmas applies pragma statements to SQLite on Temporal start.
func WithSQLitePragmas(pragmas map[string]string) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		if cfg.SQLitePragmas == nil {
			cfg.SQLitePragmas = make(map[string]string)
		}
		for k, v := range pragmas {
			cfg.SQLitePragmas[k] = v
		}
	})
}

// WithUpstreamOptions registers Temporal server options.
func WithUpstreamOptions(options ...temporal.ServerOption) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.UpstreamOptions = append(cfg.UpstreamOptions, options...)
	})
}

// WithBaseConfig sets the default Temporal server configuration.
//
// Storage and client configuration will always be overridden, however base config can be
// used to enable settings like TLS or authentication.
func WithBaseConfig(base *config.Config) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		cfg.BaseConfig = base
	})
}

// WithDynamicConfigValue sets the given dynamic config key with the given set
// of values. This will overwrite the key if already set.
func WithDynamicConfigValue(key dynamicconfig.Key, value []dynamicconfig.ConstrainedValue) ServerOption {
	return newApplyFuncContainer(func(cfg *sconfig.Config) {
		if cfg.DynamicConfig == nil {
			cfg.DynamicConfig = dynamicconfig.StaticClient{}
		}
		cfg.DynamicConfig[key] = value
	})
}

// WithSearchAttributeCacheDisabled disables search attribute caching. This
// delegates to WithDynamicConfigValue.
func WithSearchAttributeCacheDisabled() ServerOption {
	return WithDynamicConfigValue(
		dynamicconfig.ForceSearchAttributesCacheRefreshOnRead,
		[]dynamicconfig.ConstrainedValue{{Value: true}},
	)
}

type applyFuncContainer struct {
	applyInternal func(*sconfig.Config)
}

func (fso *applyFuncContainer) apply(cfg *sconfig.Config) {
	fso.applyInternal(cfg)
}

func newApplyFuncContainer(apply func(*sconfig.Config)) *applyFuncContainer {
	return &applyFuncContainer{
		applyInternal: apply,
	}
}
