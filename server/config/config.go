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
	"math/rand"
	"sort"
	"time"

	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	"go.temporal.io/server/temporal"
)

const (
	broadcastAddress     = "127.0.0.1"
	PersistenceStoreName = "sqlite-default"
	DefaultFrontendPort  = 7233
	DefaultMetricsPort   = 0
)

// UIServer abstracts the github.com/temporalio/ui-server project to
// make it an optional import for programs that need web UI support.
//
// A working implementation of this interface is available here:
// https://pkg.go.dev/github.com/temporalio/ui-server/server#Server
type UIServer interface {
	Start() error
	Stop()
}

type noopUIServer struct{}

func (noopUIServer) Start() error {
	return nil
}

func (noopUIServer) Stop() {}

type Config struct {
	Ephemeral        bool
	ClusterID        string
	DatabaseFilePath string
	FrontendPort     int
	MetricsPort      int
	DynamicPorts     bool
	Namespaces       []string
	SQLitePragmas    map[string]string
	Logger           log.Logger
	UpstreamOptions  []temporal.ServerOption
	portProvider     *PortProvider
	FrontendIP       string
	UIServer         UIServer
	BaseConfig       *config.Config
	DynamicConfig    dynamicconfig.StaticClient
}

var SupportedPragmas = map[string]struct{}{
	"journal_mode": {},
	"synchronous":  {},
}

func GetAllowedPragmas() []string {
	var allowedPragmaList []string
	for k := range SupportedPragmas {
		allowedPragmaList = append(allowedPragmaList, k)
	}
	sort.Strings(allowedPragmaList)
	return allowedPragmaList
}

func NewDefaultConfig() (*Config, error) {
	return &Config{
		Ephemeral:        true,
		DatabaseFilePath: "",
		FrontendPort:     0,
		MetricsPort:      0,
		UIServer:         noopUIServer{},
		DynamicPorts:     false,
		Namespaces:       []string{"default"},
		SQLitePragmas:    nil,
		Logger: log.NewZapLogger(log.BuildZapLogger(log.Config{
			Stdout:     true,
			Level:      "info",
			OutputFile: "",
		})),
		portProvider: NewPortProvider(),
		FrontendIP:   "",
		BaseConfig:   &config.Config{},
	}, nil
}

func Convert(cfg *Config) *config.Config {
	defer func() {
		if err := cfg.portProvider.Close(); err != nil {
			panic(err)
		}
	}()

	sqliteConfig := config.SQL{
		PluginName:        sqlite.PluginName,
		ConnectAttributes: make(map[string]string),
		DatabaseName:      cfg.DatabaseFilePath,
	}
	if cfg.Ephemeral {
		sqliteConfig.ConnectAttributes["mode"] = "memory"
		sqliteConfig.ConnectAttributes["cache"] = "shared"
		sqliteConfig.DatabaseName = fmt.Sprintf("%d", rand.Intn(9999999))
	} else {
		sqliteConfig.ConnectAttributes["mode"] = "rwc"
	}

	for k, v := range cfg.SQLitePragmas {
		sqliteConfig.ConnectAttributes["_"+k] = v
	}

	var pprofPort int
	if cfg.DynamicPorts {
		if cfg.FrontendPort == 0 {
			cfg.FrontendPort = cfg.portProvider.MustGetFreePort()
		}
		if cfg.MetricsPort == 0 {
			cfg.MetricsPort = cfg.portProvider.MustGetFreePort()
		}
		pprofPort = cfg.portProvider.MustGetFreePort()
	} else {
		if cfg.FrontendPort == 0 {
			cfg.FrontendPort = DefaultFrontendPort
		}
		if cfg.MetricsPort == 0 {
			cfg.MetricsPort = cfg.FrontendPort + 200
		}
		pprofPort = cfg.FrontendPort + 201
	}

	baseConfig := cfg.BaseConfig
	baseConfig.Global.Membership = config.Membership{
		MaxJoinDuration:  30 * time.Second,
		BroadcastAddress: broadcastAddress,
	}
	baseConfig.Global.Metrics = &metrics.Config{
		Prometheus: &metrics.PrometheusConfig{
			ListenAddress: fmt.Sprintf("%s:%d", cfg.FrontendIP, cfg.MetricsPort),
			HandlerPath:   "/metrics",
		},
	}
	baseConfig.Global.PProf = config.PProf{Port: pprofPort}
	baseConfig.Persistence = config.Persistence{
		DefaultStore:     PersistenceStoreName,
		VisibilityStore:  PersistenceStoreName,
		NumHistoryShards: 1,
		DataStores: map[string]config.DataStore{
			PersistenceStoreName: {SQL: &sqliteConfig},
		},
	}
	baseConfig.ClusterMetadata = &cluster.Config{
		EnableGlobalNamespace:    false,
		FailoverVersionIncrement: 10,
		MasterClusterName:        "active",
		CurrentClusterName:       "active",
		ClusterInformation: map[string]cluster.ClusterInformation{
			"active": {
				Enabled:                true,
				InitialFailoverVersion: 1,
				RPCAddress:             fmt.Sprintf("%s:%d", broadcastAddress, cfg.FrontendPort),
				ClusterID:              cfg.ClusterID,
			},
		},
	}
	baseConfig.DCRedirectionPolicy = config.DCRedirectionPolicy{
		Policy: "noop",
	}
	baseConfig.Services = map[string]config.Service{
		"frontend": cfg.mustGetService(0),
		"history":  cfg.mustGetService(1),
		"matching": cfg.mustGetService(2),
		"worker":   cfg.mustGetService(3),
	}
	baseConfig.Archival = config.Archival{
		History: config.HistoryArchival{
			State:      "disabled",
			EnableRead: false,
			Provider:   nil,
		},
		Visibility: config.VisibilityArchival{
			State:      "disabled",
			EnableRead: false,
			Provider:   nil,
		},
	}
	baseConfig.PublicClient = config.PublicClient{
		HostPort: fmt.Sprintf("%s:%d", broadcastAddress, cfg.FrontendPort),
	}
	baseConfig.NamespaceDefaults = config.NamespaceDefaults{
		Archival: config.ArchivalNamespaceDefaults{
			History: config.HistoryArchivalNamespaceDefaults{
				State: "disabled",
			},
			Visibility: config.VisibilityArchivalNamespaceDefaults{
				State: "disabled",
			},
		},
	}
	return baseConfig
}

func (cfg *Config) mustGetService(frontendPortOffset int) config.Service {
	svc := config.Service{
		RPC: config.RPC{
			GRPCPort:        cfg.FrontendPort + frontendPortOffset,
			MembershipPort:  cfg.FrontendPort + 100 + frontendPortOffset,
			BindOnLocalHost: true,
			BindOnIP:        "",
		},
	}

	// Assign any open port when configured to use dynamic ports
	if cfg.DynamicPorts {
		if frontendPortOffset != 0 {
			svc.RPC.GRPCPort = cfg.portProvider.MustGetFreePort()
		}
		svc.RPC.MembershipPort = cfg.portProvider.MustGetFreePort()
	}

	// Optionally bind frontend to IPv4 address
	if frontendPortOffset == 0 && cfg.FrontendIP != "" {
		svc.RPC.BindOnLocalHost = false
		svc.RPC.BindOnIP = cfg.FrontendIP
	}

	return svc
}
