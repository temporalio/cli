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

package devserver

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	uiserver "github.com/temporalio/ui-server/v2/server"
	uiconfig "github.com/temporalio/ui-server/v2/server/config"
	uiserveroptions "github.com/temporalio/ui-server/v2/server/server_options"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/metrics"
	sqliteplugin "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	"go.temporal.io/server/schema/sqlite"
	sqliteschema "go.temporal.io/server/schema/sqlite"
	"go.temporal.io/server/temporal"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type StartOptions struct {
	// Required fields
	FrontendIP             string
	FrontendPort           int
	Namespaces             []string
	ClusterID              string
	MasterClusterName      string
	CurrentClusterName     string
	InitialFailoverVersion int
	Logger                 *slog.Logger
	LogLevel               slog.Level

	// Optional fields
	UIIP                  string // Empty means no UI
	UIPort                int    // Required if UIIP is non-empty
	UIAssetPath           string
	UICodecEndpoint       string
	DatabaseFile          string
	MetricsPort           int
	PProfPort             int
	SqlitePragmas         map[string]string
	FrontendHTTPPort      int
	EnableGlobalNamespace bool
	DynamicConfigValues   map[string]any
	LogConfig             func([]byte)
	GRPCInterceptors      []grpc.UnaryServerInterceptor
}

type Server struct {
	server temporal.Server
	ui     *uiserver.Server
}

func Start(options StartOptions) (*Server, error) {
	// Validate
	if options.FrontendIP == "" {
		return nil, fmt.Errorf("missing frontend IP")
	} else if options.FrontendPort == 0 {
		return nil, fmt.Errorf("missing frontend port")
	} else if len(options.Namespaces) == 0 {
		return nil, fmt.Errorf("missing namespaces")
	} else if options.Logger == nil {
		return nil, fmt.Errorf("missing logger")
	} else if options.UIIP != "" && options.UIPort == 0 {
		return nil, fmt.Errorf("must provide UI port if UI IP is provided")
	} else if options.ClusterID == "" {
		return nil, fmt.Errorf("missing cluster ID")
	} else if options.MasterClusterName == "" {
		return nil, fmt.Errorf("missing master cluster name")
	} else if options.CurrentClusterName == "" {
		return nil, fmt.Errorf("missing current cluster name")
	} else if options.InitialFailoverVersion == 0 {
		return nil, fmt.Errorf("missing initial failover version")
	}

	// Build servers
	var ui *uiserver.Server
	if options.UIIP != "" {
		ui = options.buildUIServer()
	}
	server, err := options.buildServer()
	if err != nil {
		return nil, err
	}

	// Start. We have to start UI server in background because it's start call is
	// blocking. Therefore we have no way to relay error out to users, so we just
	// log and panic.
	if ui != nil {
		go func() {
			if err := ui.Start(); err != nil {
				options.Logger.Error("failed running UI server", "error", err)
				panic(err)
			}
		}()
	}
	if err := server.Start(); err != nil {
		// Stop UI before returning to avoid leaks
		if ui != nil {
			ui.Stop()
		}
		return nil, err
	}
	return &Server{server, ui}, nil
}

func (s *Server) Stop() {
	if s.ui != nil {
		s.ui.Stop()
	}
	s.server.Stop()
}

func (s *StartOptions) buildUIServer() *uiserver.Server {
	return uiserver.NewServer(uiserveroptions.WithConfigProvider(&uiconfig.Config{
		Host:                s.UIIP,
		Port:                s.UIPort,
		TemporalGRPCAddress: fmt.Sprintf("%v:%v", s.FrontendIP, s.FrontendPort),
		EnableUI:            true,
		UIAssetPath:         s.UIAssetPath,
		Codec:               uiconfig.Codec{Endpoint: s.UICodecEndpoint},
		HideLogs:            true,
	}))
}

func (s *StartOptions) buildServer() (temporal.Server, error) {
	opts, err := s.buildServerOptions()
	if err != nil {
		return nil, err
	}
	return temporal.NewServer(opts...)
}

func (s *StartOptions) buildServerOptions() ([]temporal.ServerOption, error) {
	// Build config and log it
	conf, err := s.buildServerConfig()
	if err != nil {
		return nil, err
	} else if s.LogConfig != nil {
		// We're going to marshal YAML
		if b, err := yaml.Marshal(conf); err != nil {
			s.Logger.Warn("Failed marshaling config for logging", "error", err)
		} else {
			s.LogConfig(b)
		}
	}

	// Build common opts
	logger := slogLogger{
		log:   s.Logger,
		level: s.LogLevel,
	}
	authorizer, err := authorization.GetAuthorizerFromConfig(&conf.Global.Authorization)
	if err != nil {
		return nil, fmt.Errorf("failed creating authorizer: %w", err)
	}
	claimMapper, err := authorization.GetClaimMapperFromConfig(&conf.Global.Authorization, logger)
	if err != nil {
		return nil, fmt.Errorf("failed creating claim mapper: %w", err)
	}
	opts := []temporal.ServerOption{
		temporal.WithConfig(conf),
		temporal.ForServices(temporal.DefaultServices),
		temporal.WithLogger(logger),
		temporal.WithAuthorizer(authorizer),
		temporal.WithClaimMapper(func(*config.Config) authorization.ClaimMapper { return claimMapper }),
	}

	// Dynamic config if set
	if len(s.DynamicConfigValues) > 0 {
		dynConf := make(dynamicconfig.StaticClient, len(s.DynamicConfigValues))
		for k, v := range s.DynamicConfigValues {
			dynConf[dynamicconfig.Key(k)] = v
		}
		opts = append(opts, temporal.WithDynamicConfigClient(dynConf))
	}

	// gRPC interceptors if set
	if len(s.GRPCInterceptors) > 0 {
		opts = append(opts, temporal.WithChainedFrontendGrpcInterceptors(s.GRPCInterceptors...))
	}

	return opts, nil
}

func (s *StartOptions) buildServerConfig() (*config.Config, error) {
	var conf config.Config
	// Global config
	conf.Global.Membership.MaxJoinDuration = 30 * time.Second
	conf.Global.Membership.BroadcastAddress = "127.0.0.1"
	if conf.Global.Metrics == nil && s.MetricsPort > 0 {
		conf.Global.Metrics = &metrics.Config{
			Prometheus: &metrics.PrometheusConfig{
				ListenAddress: fmt.Sprintf("%v:%v", s.FrontendIP, s.MetricsPort),
				HandlerPath:   "/metrics",
			},
		}
	}
	conf.Global.PProf.Port = s.PProfPort

	// Persistence config
	conf.Persistence.DefaultStore = "sqlite-default"
	conf.Persistence.VisibilityStore = "sqlite-default"
	conf.Persistence.NumHistoryShards = 1
	sqlConf, err := s.buildSQLConfig()
	if err != nil {
		return nil, err
	}
	conf.Persistence.DataStores = map[string]config.DataStore{"sqlite-default": {SQL: sqlConf}}

	// Other config
	if conf.ClusterMetadata == nil {
		conf.ClusterMetadata = &cluster.Config{
			EnableGlobalNamespace:    s.EnableGlobalNamespace,
			FailoverVersionIncrement: 10,
			MasterClusterName:        s.MasterClusterName,
			CurrentClusterName:       s.CurrentClusterName,
			ClusterInformation: map[string]cluster.ClusterInformation{
				s.CurrentClusterName: {
					Enabled:                true,
					InitialFailoverVersion: int64(s.InitialFailoverVersion),
					RPCAddress:             fmt.Sprintf("127.0.0.1:%v", s.FrontendPort),
					ClusterID:              s.ClusterID,
				},
			},
		}
	}
	conf.DCRedirectionPolicy.Policy = "noop"
	portProvider := NewPortProvider()
	defer portProvider.Close()
	conf.Services = map[string]config.Service{
		"frontend": s.buildServiceConfig(portProvider, true),
		"history":  s.buildServiceConfig(portProvider, false),
		"matching": s.buildServiceConfig(portProvider, false),
		"worker":   s.buildServiceConfig(portProvider, false),
	}
	conf.Archival.History.State = "disabled"
	conf.Archival.Visibility.State = "disabled"
	conf.NamespaceDefaults.Archival.History.State = "disabled"
	conf.NamespaceDefaults.Archival.Visibility.State = "disabled"
	conf.PublicClient.HostPort = fmt.Sprintf("127.0.0.1:%v", s.FrontendPort)
	return &conf, nil
}

func (s *StartOptions) buildSQLConfig() (*config.SQL, error) {
	conf := config.SQL{PluginName: sqliteplugin.PluginName}
	if s.DatabaseFile == "" {
		conf.ConnectAttributes = map[string]string{"mode": "memory", "cache": "shared"}
		conf.DatabaseName = strconv.Itoa(rand.Intn(9999999))
	} else {
		conf.ConnectAttributes = map[string]string{"mode": "rwc"}
		conf.DatabaseName = s.DatabaseFile
	}
	for k, v := range s.SqlitePragmas {
		// Only some pragmas allowed
		switch k {
		case "journal_mode", "synchronous":
		default:
			return nil, fmt.Errorf("unrecognized pragma %q, only 'journal_mode' and 'synchronous' allowed", k)
		}
		conf.ConnectAttributes["_"+k] = v
	}

	// Apply migrations to sqlite if using file but it does not exist
	if s.DatabaseFile != "" {
		if _, err := os.Stat(s.DatabaseFile); os.IsNotExist(err) {
			// Eagerly check parent dir
			if _, err := os.Stat(filepath.Dir(s.DatabaseFile)); err != nil {
				return nil, fmt.Errorf("failed checking dir for database file: %w", err)
			}
			if err := sqliteschema.SetupSchema(&conf); err != nil {
				return nil, fmt.Errorf("failed setting up schema: %w", err)
			}
		}
	}

	// Create namespaces
	namespaces := make([]*sqliteschema.NamespaceConfig, len(s.Namespaces))
	for i, ns := range s.Namespaces {
		namespaces[i] = sqlite.NewNamespaceConfig(s.CurrentClusterName, ns, false)
	}
	if err := sqliteschema.CreateNamespaces(&conf, namespaces...); err != nil {
		return nil, fmt.Errorf("failed creating namespaces: %w", err)
	}
	return &conf, nil
}

func (s *StartOptions) buildServiceConfig(p *PortProvider, frontend bool) config.Service {
	var conf config.Service
	if frontend {
		conf.RPC.GRPCPort = s.FrontendPort
		conf.RPC.BindOnIP = s.FrontendIP
		if s.FrontendHTTPPort > 0 {
			conf.RPC.HTTPPort = s.FrontendHTTPPort
		}
	} else {
		conf.RPC.GRPCPort = p.MustGetFreePort()
		conf.RPC.BindOnLocalHost = true
	}
	conf.RPC.MembershipPort = p.MustGetFreePort()
	return conf
}
