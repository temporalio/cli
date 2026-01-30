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
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/membership/static"
	"go.temporal.io/server/common/metrics"
	sqliteplugin "go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	"go.temporal.io/server/common/primitives"
	"go.temporal.io/server/components/callbacks"
	"go.temporal.io/server/components/nexusoperations"
	"go.temporal.io/server/schema/sqlite"
	sqliteschema "go.temporal.io/server/schema/sqlite"
	"go.temporal.io/server/temporal"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

const (
	localhost = "127.0.0.1"
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
	PublicPath            string
	DatabaseFile          string
	MetricsPort           int
	PProfPort             int
	SqlitePragmas         map[string]string
	FrontendHTTPPort      int
	EnableGlobalNamespace bool
	DynamicConfigValues   map[string]any
	SearchAttributes      map[string]enums.IndexedValueType
	LogConfig             func([]byte)
	GRPCInterceptors      []grpc.UnaryServerInterceptor
}

type Server struct {
	server   temporal.Server
	ui       *uiserver.Server
	logLevel *slog.LevelVar
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

	if options.FrontendHTTPPort == 0 {
		options.FrontendHTTPPort = MustGetFreePort(options.FrontendIP)
	}

	// Build servers
	var ui *uiserver.Server
	if options.UIIP != "" {
		ui = options.buildUIServer()
	}
	server, logLevel, err := options.buildServer()
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
	return &Server{server: server, ui: ui, logLevel: logLevel}, nil
}

func (s *Server) Stop() {
	if s.ui != nil {
		s.ui.Stop()
	}
	s.server.Stop()
}

func (s *Server) SuppressWarnings() {
	if s.logLevel != nil {
		s.logLevel.Set(slog.LevelError)
	}
}

func (s *StartOptions) buildUIServer() *uiserver.Server {
	return uiserver.NewServer(uiserveroptions.WithConfigProvider(&uiconfig.Config{
		Host:                MaybeEscapeIPv6(s.UIIP),
		Port:                s.UIPort,
		TemporalGRPCAddress: fmt.Sprintf("%v:%v", MaybeEscapeIPv6(s.FrontendIP), s.FrontendPort),
		EnableUI:            true,
		PublicPath:          s.PublicPath,
		UIAssetPath:         s.UIAssetPath,
		Codec:               uiconfig.Codec{Endpoint: s.UICodecEndpoint},
		CORS:                uiconfig.CORS{CookieInsecure: true},
		HideLogs:            true,
	}))
}

func (s *StartOptions) buildServer() (temporal.Server, *slog.LevelVar, error) {
	opts, logLevel, err := s.buildServerOptions()
	if err != nil {
		return nil, nil, err
	}
	server, err := temporal.NewServer(opts...)
	return server, logLevel, err
}

func (s *StartOptions) buildServerOptions() ([]temporal.ServerOption, *slog.LevelVar, error) {
	// Build config and log it
	conf, err := s.buildServerConfig()
	if err != nil {
		return nil, nil, err
	} else if s.LogConfig != nil {
		// We're going to marshal YAML
		if b, err := yaml.Marshal(conf); err != nil {
			s.Logger.Warn("Failed marshaling config for logging", "error", err)
		} else {
			s.LogConfig(b)
		}
	}

	// Build common opts
	logLevel := &slog.LevelVar{}
	logLevel.Set(s.LogLevel)
	logger := slogLogger{
		log:   s.Logger,
		level: logLevel,
	}
	authorizer, err := authorization.GetAuthorizerFromConfig(&conf.Global.Authorization)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating authorizer: %w", err)
	}
	claimMapper, err := authorization.GetClaimMapperFromConfig(&conf.Global.Authorization, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating claim mapper: %w", err)
	}
	opts := []temporal.ServerOption{
		temporal.WithConfig(conf),
		temporal.ForServices(temporal.DefaultServices),
		temporal.WithStaticHosts(map[primitives.ServiceName]static.Hosts{
			primitives.FrontendService: static.SingleLocalHost(
				fmt.Sprintf("%v:%v", localhost, conf.Services[string(primitives.FrontendService)].RPC.GRPCPort)),
			primitives.MatchingService: static.SingleLocalHost(
				fmt.Sprintf("%v:%v", localhost, conf.Services[string(primitives.MatchingService)].RPC.GRPCPort)),
			primitives.HistoryService: static.SingleLocalHost(
				fmt.Sprintf("%v:%v", localhost, conf.Services[string(primitives.HistoryService)].RPC.GRPCPort)),
			primitives.WorkerService: static.SingleLocalHost(
				fmt.Sprintf("%v:%v", localhost, conf.Services[string(primitives.WorkerService)].RPC.GRPCPort)),
		}),
		temporal.WithLogger(logger),
		temporal.WithAuthorizer(authorizer),
		temporal.WithClaimMapper(func(*config.Config) authorization.ClaimMapper { return claimMapper }),
	}

	dynConf := make(dynamicconfig.StaticClient, len(s.DynamicConfigValues)+1)
	// Setting host level mutable state cache size to 8k.
	dynConf[dynamicconfig.HistoryCacheHostLevelMaxSize.Key()] = 8096
	// Up default visibility RPS
	dynConf[dynamicconfig.FrontendMaxNamespaceVisibilityRPSPerInstance.Key()] = 100
	// Enable the system callback URL for worker targets.
	dynConf[nexusoperations.UseSystemCallbackURL.Key()] = true
	dynConf[callbacks.AllowedAddresses.Key()] = []struct {
		Pattern       string
		AllowInsecure bool
	}{
		{
			Pattern:       fmt.Sprintf("%s:%d", MaybeEscapeIPv6(s.FrontendIP), s.FrontendHTTPPort),
			AllowInsecure: true,
		},
	}

	// Dynamic config if set
	for k, v := range s.DynamicConfigValues {
		dynConf[dynamicconfig.MakeKey(k)] = v
	}
	opts = append(opts, temporal.WithDynamicConfigClient(dynConf))

	// gRPC interceptors if set
	if len(s.GRPCInterceptors) > 0 {
		opts = append(opts, temporal.WithChainedFrontendGrpcInterceptors(s.GRPCInterceptors...))
	}

	return opts, logLevel, nil
}

func (s *StartOptions) buildServerConfig() (*config.Config, error) {
	var conf config.Config
	// Global config
	conf.Global.Membership.MaxJoinDuration = 30 * time.Second
	conf.Global.Membership.BroadcastAddress = s.FrontendIP
	if conf.Global.Metrics == nil && s.MetricsPort > 0 {
		conf.Global.Metrics = &metrics.Config{
			Prometheus: &metrics.PrometheusConfig{
				ListenAddress: fmt.Sprintf("%v:%v", MaybeEscapeIPv6(s.FrontendIP), s.MetricsPort),
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
					RPCAddress:             fmt.Sprintf("%v:%v", MaybeEscapeIPv6(s.FrontendIP), s.FrontendPort),
					HTTPAddress:            fmt.Sprintf("%v:%v", MaybeEscapeIPv6(s.FrontendIP), s.FrontendHTTPPort),
					ClusterID:              s.ClusterID,
				},
			},
		}
	}
	conf.DCRedirectionPolicy.Policy = "noop"
	conf.Services = map[string]config.Service{
		"frontend": s.buildServiceConfig(true),
		"history":  s.buildServiceConfig(false),
		"matching": s.buildServiceConfig(false),
		"worker":   s.buildServiceConfig(false),
	}
	conf.Archival.History.State = "disabled"
	conf.Archival.Visibility.State = "disabled"
	conf.NamespaceDefaults.Archival.History.State = "disabled"
	conf.NamespaceDefaults.Archival.Visibility.State = "disabled"
	conf.PublicClient.HostPort = fmt.Sprintf("%v:%v", MaybeEscapeIPv6(s.FrontendIP), s.FrontendPort)
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
		conf.ConnectAttributes[k] = v
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
		nsConfig, err := sqlite.NewNamespaceConfig(s.CurrentClusterName, ns, false, s.SearchAttributes)
		if err != nil {
			return nil, fmt.Errorf("failed creating namespace config: %w", err)
		}
		namespaces[i] = nsConfig
	}
	if err := sqliteschema.CreateNamespaces(&conf, namespaces...); err != nil {
		return nil, fmt.Errorf("failed creating namespaces: %w", err)
	}
	return &conf, nil
}

func (s *StartOptions) buildServiceConfig(frontend bool) config.Service {
	var conf config.Service
	if frontend {
		conf.RPC.GRPCPort = s.FrontendPort
		conf.RPC.BindOnIP = s.FrontendIP
		conf.RPC.HTTPPort = s.FrontendHTTPPort
	} else {
		conf.RPC.GRPCPort = MustGetFreePort(s.FrontendIP)
		conf.RPC.BindOnIP = s.FrontendIP
	}
	return conf
}
