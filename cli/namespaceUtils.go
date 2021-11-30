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

package cli

import (
	"fmt"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/uber-go/tally/v4"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/config"

	"go.temporal.io/server/common"
	"go.temporal.io/server/common/archiver"
	"go.temporal.io/server/common/archiver/provider"
	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/common/namespace"
	"go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/persistence/client"
	"go.temporal.io/server/common/resolver"
)

const (
	dependencyMaxQPS = 100
)

var (
	registerNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    FlagDescription,
			Aliases: FlagDescriptionAlias,
			Usage:   "Namespace description",
		},
		&cli.StringFlag{
			Name:    FlagOwnerEmail,
			Aliases: FlagOwnerEmailAlias,
			Usage:   "Owner email",
		},
		&cli.StringFlag{
			Name:    FlagRetention,
			Aliases: FlagRetentionAlias,
			Usage:   "Workflow execution retention",
		},
		&cli.StringFlag{
			Name:    FlagActiveClusterName,
			Aliases: FlagActiveClusterNameAlias,
			Usage:   "Active cluster name",
		},
		&cli.StringFlag{
			// use StringFlag instead of buggy StringSliceFlag
			// TODO when https://github.com/urfave/cli/pull/392 & v2 is released
			//  consider update urfave/cli
			Name:    FlagClusters,
			Aliases: FlagClustersAlias,
			Usage:   "Clusters",
		},
		&cli.StringFlag{
			Name:    FlagIsGlobalNamespace,
			Aliases: FlagIsGlobalNamespaceAlias,
			Usage:   "Flag to indicate whether namespace is a global namespace",
		},
		&cli.StringFlag{
			Name:    FlagNamespaceData,
			Aliases: FlagNamespaceDataAlias,
			Usage:   "Namespace data of key value pairs, in format of k1:v1,k2:v2,k3:v3",
		},
		&cli.StringFlag{
			Name:    FlagHistoryArchivalState,
			Aliases: FlagHistoryArchivalStateAlias,
			Usage:   "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:    FlagHistoryArchivalURI,
			Aliases: FlagHistoryArchivalURIAlias,
			Usage:   "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
		},
		&cli.StringFlag{
			Name:    FlagVisibilityArchivalState,
			Aliases: FlagVisibilityArchivalStateAlias,
			Usage:   "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:    FlagVisibilityArchivalURI,
			Aliases: FlagVisibilityArchivalURIAlias,
			Usage:   "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
		},
	}

	updateNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    FlagDescription,
			Aliases: FlagDescriptionAlias,
			Usage:   "Namespace description",
		},
		&cli.StringFlag{
			Name:    FlagOwnerEmail,
			Aliases: FlagOwnerEmailAlias,
			Usage:   "Owner email",
		},
		&cli.StringFlag{
			Name:  FlagState,
			Usage: "Namespace state",
		},
		&cli.StringFlag{
			Name:    FlagRetention,
			Aliases: FlagRetentionAlias,
			Usage:   "Workflow execution retention",
		},
		&cli.StringFlag{
			Name:    FlagActiveClusterName,
			Aliases: FlagActiveClusterNameAlias,
			Usage:   "Active cluster name",
		},
		&cli.StringFlag{
			// use StringFlag instead of buggy StringSliceFlag
			// TODO when https://github.com/urfave/cli/pull/392 & v2 is released
			//  consider update urfave/cli
			Name:    FlagClusters,
			Aliases: FlagClustersAlias,
			Usage:   "Clusters",
		},
		&cli.StringFlag{
			Name:    FlagNamespaceData,
			Aliases: FlagNamespaceDataAlias,
			Usage:   "Namespace data of key value pairs, in format of k1:v1,k2:v2,k3:v3 ",
		},
		&cli.StringFlag{
			Name:    FlagHistoryArchivalState,
			Aliases: FlagHistoryArchivalStateAlias,
			Usage:   "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:    FlagHistoryArchivalURI,
			Aliases: FlagHistoryArchivalURIAlias,
			Usage:   "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
		},
		&cli.StringFlag{
			Name:    FlagVisibilityArchivalState,
			Aliases: FlagVisibilityArchivalStateAlias,
			Usage:   "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:    FlagVisibilityArchivalURI,
			Aliases: FlagVisibilityArchivalURIAlias,
			Usage:   "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
		},
		&cli.StringFlag{
			Name:  FlagAddBadBinary,
			Usage: "Binary checksum to add for resetting workflow",
		},
		&cli.StringFlag{
			Name:  FlagRemoveBadBinary,
			Usage: "Binary checksum to remove for resetting workflow",
		},
		&cli.StringFlag{
			Name:  FlagReason,
			Usage: "Reason for the operation",
		},
		&cli.BoolFlag{
			Name:  FlagPromoteNamespace,
			Usage: "Promote local namespace to global namespace",
		},
	}

	describeNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  FlagNamespaceID,
			Usage: "Namespace Id (required if not specify namespace)",
		},
	}

	listNamespacesFlags = []cli.Flag{}

	adminNamespaceCommonFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    FlagServiceConfigDir,
			Aliases: FlagServiceConfigDirAlias,
			Usage:   "Required service configuration dir",
		},
		&cli.StringFlag{
			Name:    FlagServiceEnv,
			Aliases: FlagServiceEnvAlias,
			Usage:   "Optional service env for loading service configuration",
		},
		&cli.StringFlag{
			Name:    FlagServiceZone,
			Aliases: FlagServiceZoneAlias,
			Usage:   "Optional service zone for loading service configuration",
		},
	}

	adminRegisterNamespaceFlags = append(
		registerNamespaceFlags,
		adminNamespaceCommonFlags...,
	)

	adminUpdateNamespaceFlags = append(
		updateNamespaceFlags,
		adminNamespaceCommonFlags...,
	)

	adminDescribeNamespaceFlags = append(
		updateNamespaceFlags,
		adminNamespaceCommonFlags...,
	)
)

func initializeFrontendClient(
	context *cli.Context,
) workflowservice.WorkflowServiceClient {
	return cFactory.FrontendClient(context)
}

func initializeAdminNamespaceHandler(
	context *cli.Context,
) (namespace.Handler, error) {
	configuration, err := loadConfig(context)
	if err != nil {
		return nil, err
	}

	metricsClient := initializeMetricsClient()
	logger := log.NewZapLogger(log.BuildZapLogger(configuration.Log))

	factory := initializePersistenceFactory(
		configuration,
		metricsClient,
		logger,
	)

	metadataMgr, err := factory.NewMetadataManager()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize metadata manager: %s", err)
	}

	clusterMetadata := initializeClusterMetadata(configuration)

	dynamicConfig, err := initializeDynamicConfig(configuration, logger)
	if err != nil {
		return nil, err
	}

	archProvider, err := initializeArchivalProvider(configuration, clusterMetadata, metricsClient, logger)
	if err != nil {
		return nil, err
	}

	return initializeNamespaceHandler(
		logger,
		metadataMgr,
		clusterMetadata,
		initializeArchivalMetadata(configuration, dynamicConfig),
		archProvider,
	), nil
}

func loadConfig(
	context *cli.Context,
) (*config.Config, error) {
	env := getEnvironment(context)
	zone := getZone(context)
	configDir, err := getConfigDir(context)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	err = config.Load(env, configDir, zone, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %s", err)
	}
	return &cfg, nil
}

func initializeNamespaceHandler(
	logger log.Logger,
	metadataMgr persistence.MetadataManager,
	clusterMetadata cluster.Metadata,
	archivalMetadata archiver.ArchivalMetadata,
	archiverProvider provider.ArchiverProvider,
) namespace.Handler {
	return namespace.NewHandler(
		dynamicconfig.GetIntPropertyFilteredByNamespace(namespace.MaxBadBinaries),
		logger,
		metadataMgr,
		clusterMetadata,
		initializeNamespaceReplicator(logger),
		archivalMetadata,
		archiverProvider,
	)
}

func initializePersistenceFactory(
	serviceConfig *config.Config,
	metricsClient metrics.Client,
	logger log.Logger,
) client.Factory {

	pConfig := serviceConfig.Persistence
	pFactory := client.NewFactory(
		&pConfig,
		resolver.NewNoopResolver(),
		dynamicconfig.GetIntPropertyFn(dependencyMaxQPS),
		nil, // TODO propagate abstract datastore factory from the CLI.
		"",
		metricsClient,
		logger,
	)
	return pFactory
}

func initializeClusterMetadata(
	serviceConfig *config.Config,
) cluster.Metadata {

	clusterMetadata := serviceConfig.ClusterMetadata
	return cluster.NewMetadata(
		clusterMetadata.EnableGlobalNamespace,
		clusterMetadata.FailoverVersionIncrement,
		clusterMetadata.MasterClusterName,
		clusterMetadata.CurrentClusterName,
		clusterMetadata.ClusterInformation,
	)
}

func initializeArchivalMetadata(
	serviceConfig *config.Config,
	dynamicConfig *dynamicconfig.Collection,
) archiver.ArchivalMetadata {

	return archiver.NewArchivalMetadata(
		dynamicConfig,
		serviceConfig.Archival.History.State,
		serviceConfig.Archival.History.EnableRead,
		serviceConfig.Archival.Visibility.State,
		serviceConfig.Archival.Visibility.EnableRead,
		&serviceConfig.NamespaceDefaults.Archival,
	)
}

func initializeArchivalProvider(
	serviceConfig *config.Config,
	clusterMetadata cluster.Metadata,
	metricsClient metrics.Client,
	logger log.Logger,
) (provider.ArchiverProvider, error) {

	archiverProvider := provider.NewArchiverProvider(
		serviceConfig.Archival.History.Provider,
		serviceConfig.Archival.Visibility.Provider,
	)

	historyArchiverBootstrapContainer := &archiver.HistoryBootstrapContainer{
		ExecutionManager: nil, // not used
		Logger:           logger,
		MetricsClient:    metricsClient,
		ClusterMetadata:  clusterMetadata,
	}
	visibilityArchiverBootstrapContainer := &archiver.VisibilityBootstrapContainer{
		Logger:          logger,
		MetricsClient:   metricsClient,
		ClusterMetadata: clusterMetadata,
	}

	err := archiverProvider.RegisterBootstrapContainer(
		common.FrontendServiceName,
		historyArchiverBootstrapContainer,
		visibilityArchiverBootstrapContainer,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing archival provider: %s", err)
	}
	return archiverProvider, nil
}

func initializeNamespaceReplicator(
	logger log.Logger,
) namespace.Replicator {

	namespaceReplicationQueue := &persistence.MockNamespaceReplicationQueue{}
	namespaceReplicationQueue.EXPECT().Publish(gomock.Any()).Return(nil)
	return namespace.NewNamespaceReplicator(namespaceReplicationQueue, logger)
}

func initializeDynamicConfig(
	serviceConfig *config.Config,
	logger log.Logger,
) (*dynamicconfig.Collection, error) {

	// the done channel is used by dynamic config to stop refreshing
	// and CLI does not need that, so just close the done channel
	doneChan := make(chan interface{})
	close(doneChan)
	dynamicConfigClient, err := dynamicconfig.NewFileBasedClient(
		serviceConfig.DynamicConfigClient,
		logger,
		doneChan,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize dynamic config: %s", err)
	}
	return dynamicconfig.NewCollection(dynamicConfigClient, logger), nil
}

func initializeMetricsClient() metrics.Client {
	return metrics.NewClient(&metrics.ClientConfig{}, tally.NoopScope, metrics.Common)
}

func getEnvironment(c *cli.Context) string {
	return strings.TrimSpace(c.String(FlagServiceEnv))
}

func getZone(c *cli.Context) string {
	return strings.TrimSpace(c.String(FlagServiceZone))
}

func getConfigDir(c *cli.Context) (string, error) {
	dirPath := c.String(FlagServiceConfigDir)
	if len(dirPath) == 0 {
		return "", fmt.Errorf("provide service configuration dir path")
	}
	return dirPath, nil
}
