package namespace

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

var (
	createNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagDescription,
			Usage:    common.FlagDescriptionDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagOwnerEmail,
			Usage:    common.FlagOwnerDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagRetention,
			Usage:    common.FlagRetentionDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagActiveCluster,
			Usage:    common.FlagActiveClusterDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagCluster,
			Usage:    common.FlagClusterDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagIsGlobalNamespace,
			Usage:    common.FlagIsGlobalNamespaceDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagNamespaceData,
			Usage:    common.FlagNamespaceDataDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalState,
			Usage:    common.FlagHistoryArchivalStateDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalURI,
			Usage:    common.FlagHistoryArchivalURIDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalState,
			Usage:    common.FlagVisibilityArchivalStateDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalURI,
			Usage:    common.FlagVisibilityArchivalURIDefinition,
			Category: common.CategoryMain,
		},
	}

	updateNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagDescription,
			Usage:    common.FlagDescriptionDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagOwnerEmail,
			Usage:    common.FlagOwnerDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagRetention,
			Usage:    common.FlagRetentionDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagActiveCluster,
			Usage:    common.FlagActiveClusterDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagCluster,
			Usage:    common.FlagClusterDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagNamespaceData,
			Usage:    common.FlagNamespaceDataDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalState,
			Usage:    common.FlagHistoryArchivalStateDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalURI,
			Usage:    common.FlagHistoryArchivalURIDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalState,
			Usage:    common.FlagVisibilityArchivalStateDefinition,
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalURI,
			Usage:    common.FlagVisibilityArchivalURIDefinition,
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagPromoteNamespace,
			Usage:    common.FlagPromoteNamespaceDefinition,
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagVerbose,
			Aliases:  common.FlagVerboseAlias,
			Usage:    common.FlagNamespaceVerboseDefinition,
			Category: common.CategoryDisplay,
		},
	}

	describeNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagNamespaceID,
			Usage:    "Namespace Id",
			Category: common.CategoryMain,
		},
	}
)
