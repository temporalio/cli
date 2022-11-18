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

package namespace

import (
	"github.com/temporalio/temporal-cli/common"
	"github.com/urfave/cli/v2"
)

var (
	registerNamespaceFlags = []cli.Flag{
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
