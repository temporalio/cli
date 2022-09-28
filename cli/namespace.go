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

// by default we don't require any namespace data. But this can be overridden by calling SetRequiredNamespaceDataKeys()
var requiredNamespaceDataKeys []string

// SetRequiredNamespaceDataKeys will set requiredNamespaceDataKeys
func SetRequiredNamespaceDataKeys(keys []string) {
	requiredNamespaceDataKeys = keys
}

func newNamespaceCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "describe",
			Usage:     "Describe a Namespace by name or Id",
			Flags:     describeNamespaceFlags,
			ArgsUsage: "namespace_name",
			Action: func(c *cli.Context) error {
				return DescribeNamespace(c)
			},
		},
		{
			Name:      "list",
			Usage:     "List all Namespaces",
			Flags:     listNamespacesFlags,
			ArgsUsage: " ",
			Action: func(c *cli.Context) error {
				return ListNamespaces(c)
			},
		},
		{
			Name:      "register",
			Usage:     "Register a new Namespace",
			Flags:     registerNamespaceFlags,
			ArgsUsage: "namespace_name [cluster_name...]",
			Action: func(c *cli.Context) error {
				return RegisterNamespace(c)
			},
		},
		{
			Name:      "update",
			Usage:     "Update a Namespace",
			Flags:     updateNamespaceFlags,
			ArgsUsage: "namespace_name [cluster_name...]",
			Action: func(c *cli.Context) error {
				return UpdateNamespace(c)
			},
		},
		{
			Name:      "delete",
			Usage:     "Delete existing Namespace",
			Flags:     deleteNamespacesFlags,
			ArgsUsage: "namespace_name",
			Action: func(c *cli.Context) error {
				return DeleteNamespace(c)
			},
		},
	}
}
