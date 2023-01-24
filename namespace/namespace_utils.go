package namespace

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

var (
	createNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagDescription,
			Usage:    "Namespace description",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagOwnerEmail,
			Usage:    "Owner email",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagRetention,
			Usage:    "Workflow Execution retention",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagActiveCluster,
			Usage:    "Active cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagCluster,
			Usage:    "Cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagIsGlobalNamespace,
			Usage:    "Flag to indicate whether namespace is a global namespace",
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagNamespaceData,
			Usage:    "Namespace data in a format key=value",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalState,
			Usage:    "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalURI,
			Usage:    "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalState,
			Usage:    "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalURI,
			Usage:    "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
	}

	updateNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagDescription,
			Usage:    "Namespace description",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagOwnerEmail,
			Usage:    "Owner email",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagRetention,
			Usage:    "Workflow Execution retention",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagActiveCluster,
			Usage:    "Active cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagCluster,
			Usage:    "Cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagNamespaceData,
			Usage:    "Namespace data in a format key=value",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalState,
			Usage:    "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalURI,
			Usage:    "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalState,
			Usage:    "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalURI,
			Usage:    "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagReason,
			Usage:    "Reason for the operation",
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagPromoteNamespace,
			Usage:    "Promote local namespace to global namespace",
			Category: common.CategoryMain,
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
