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
			Name:  FlagDescription,
			Usage: "Namespace description",
		},
		&cli.StringFlag{
			Name:  FlagOwnerEmail,
			Usage: "Owner email",
		},
		&cli.StringFlag{
			Name:  FlagRetention,
			Usage: "Workflow Execution retention",
		},
		&cli.StringFlag{
			Name:  FlagActiveCluster,
			Usage: "Active cluster name",
		},
		&cli.StringSliceFlag{
			Name:  FlagCluster,
			Usage: "Cluster name",
		},
		&cli.StringFlag{
			Name:  FlagIsGlobalNamespace,
			Usage: "Flag to indicate whether namespace is a global namespace",
		},
		&cli.StringSliceFlag{
			Name:  FlagNamespaceData,
			Usage: "Namespace data in a format key=value",
		},
		&cli.StringFlag{
			Name:  FlagHistoryArchivalState,
			Usage: "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:  FlagHistoryArchivalURI,
			Usage: "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
		},
		&cli.StringFlag{
			Name:  FlagVisibilityArchivalState,
			Usage: "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:  FlagVisibilityArchivalURI,
			Usage: "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
		},
	}

	updateNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  FlagDescription,
			Usage: "Namespace description",
		},
		&cli.StringFlag{
			Name:  FlagOwnerEmail,
			Usage: "Owner email",
		},
		&cli.StringFlag{
			Name:  FlagRetention,
			Usage: "Workflow Execution retention",
		},
		&cli.StringFlag{
			Name:  FlagActiveCluster,
			Usage: "Active cluster name",
		},
		&cli.StringFlag{
			Name:  FlagCluster,
			Usage: "Cluster name",
		},
		&cli.StringSliceFlag{
			Name:  FlagNamespaceData,
			Usage: "Namespace data in a format key=value",
		},
		&cli.StringFlag{
			Name:  FlagHistoryArchivalState,
			Usage: "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:  FlagHistoryArchivalURI,
			Usage: "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
		},
		&cli.StringFlag{
			Name:  FlagVisibilityArchivalState,
			Usage: "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
		},
		&cli.StringFlag{
			Name:  FlagVisibilityArchivalURI,
			Usage: "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
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
			Usage: "Namespace Id",
		},
	}

	listNamespacesFlags []cli.Flag

	deleteNamespacesFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:    FlagYes,
			Aliases: FlagYesAlias,
			Usage:   "Confirm all prompts",
		},
	}
)
