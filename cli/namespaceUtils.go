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
	"github.com/urfave/cli/v2"
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
)
