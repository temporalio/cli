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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sconfig "github.com/temporalio/cli/server/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/schema/sqlite"
	"go.temporal.io/server/temporal"
)

// Server wraps temporal.Server.
type Server struct {
	internal         temporal.Server
	ui               sconfig.UIServer
	frontendHostPort string
	config           *sconfig.Config
}

type ServerOption interface {
	apply(*sconfig.Config)
}

// NewServer returns a new instance of Server.
func NewServer(opts ...ServerOption) (*Server, error) {
	c, err := sconfig.NewDefaultConfig()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	for pragma := range c.SQLitePragmas {
		if _, ok := sconfig.SupportedPragmas[strings.ToLower(pragma)]; !ok {
			return nil, fmt.Errorf("ERROR: unsupported pragma %q, %v allowed", pragma, sconfig.GetAllowedPragmas())
		}
	}

	cfg := sconfig.Convert(c)
	sqlConfig := cfg.Persistence.DataStores[sconfig.PersistenceStoreName].SQL

	if !c.Ephemeral {
		// Apply migrations if file does not already exist
		if _, err := os.Stat(c.DatabaseFilePath); os.IsNotExist(err) {
			// Check if any of the parent dirs are missing
			dir := filepath.Dir(c.DatabaseFilePath)
			if _, err := os.Stat(dir); err != nil {
				return nil, fmt.Errorf("error setting up schema: %w", err)
			}

			if err := sqlite.SetupSchema(sqlConfig); err != nil {
				return nil, fmt.Errorf("error setting up schema: %w", err)
			}
		}
	}
	// Pre-create namespaces
	var namespaces []*sqlite.NamespaceConfig
	for _, ns := range c.Namespaces {
		namespaces = append(namespaces, sqlite.NewNamespaceConfig(cfg.ClusterMetadata.CurrentClusterName, ns, false))
	}
	if err := sqlite.CreateNamespaces(sqlConfig, namespaces...); err != nil {
		return nil, fmt.Errorf("error creating namespaces: %w", err)
	}

	authorizer, err := authorization.GetAuthorizerFromConfig(&cfg.Global.Authorization)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate authorizer: %w", err)
	}

	claimMapper, err := authorization.GetClaimMapperFromConfig(&cfg.Global.Authorization, c.Logger)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate claim mapper: %w", err)
	}

	serverOpts := []temporal.ServerOption{
		temporal.WithConfig(cfg),
		temporal.ForServices(temporal.DefaultServices),
		temporal.WithLogger(c.Logger),
		temporal.WithAuthorizer(authorizer),
		temporal.WithClaimMapper(func(cfg *config.Config) authorization.ClaimMapper {
			return claimMapper
		}),
	}

	if len(c.DynamicConfig) > 0 {
		// To prevent having to code fall-through semantics right now, we currently
		// eagerly fail if dynamic config is being configured in two ways
		if cfg.DynamicConfigClient != nil {
			return nil, fmt.Errorf("unable to have file-based dynamic config and individual dynamic config values")
		}
		serverOpts = append(serverOpts, temporal.WithDynamicConfigClient(c.DynamicConfig))
	}

	if len(c.UpstreamOptions) > 0 {
		serverOpts = append(serverOpts, c.UpstreamOptions...)
	}

	si, err := temporal.NewServer(serverOpts...)
	if err != nil {
		return nil, err
	}

	s := &Server{
		internal:         si,
		ui:               c.UIServer,
		frontendHostPort: cfg.PublicClient.HostPort,
		config:           c,
	}

	return s, nil
}

// Start temporal server.
func (s *Server) Start() error {
	go func() {
		if err := s.ui.Start(); err != nil {
			panic(err)
		}
	}()
	return s.internal.Start()
}

// Stop the server.
func (s *Server) Stop() {
	s.ui.Stop()
	s.internal.Stop()
}

// NewClient initializes a client ready to communicate with the Temporal
// server in the target namespace.
func (s *Server) NewClient(ctx context.Context, namespace string) (client.Client, error) {
	return s.NewClientWithOptions(ctx, client.Options{Namespace: namespace})
}

// NewClientWithOptions is the same as NewClient but allows further customization.
//
// To set the client's namespace, use the corresponding field in client.Options.
//
// Note that the HostPort and ConnectionOptions fields of client.Options will always be overridden.
func (s *Server) NewClientWithOptions(ctx context.Context, options client.Options) (client.Client, error) {
	options.HostPort = s.frontendHostPort
	return client.NewClient(options)
}

// FrontendHostPort returns the host:port for this server.
//
// When constructing a Temporal client from within the same process,
// NewClient or NewClientWithOptions should be used instead.
func (s *Server) FrontendHostPort() string {
	return s.frontendHostPort
}
