package namespace

import (
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
)

// by default we don't require any namespace data. But this can be overridden by calling SetRequiredNamespaceDataKeys()
var requiredNamespaceDataKeys []string

// SetRequiredNamespaceDataKeys will set requiredNamespaceDataKeys
func SetRequiredNamespaceDataKeys(keys []string) {
	requiredNamespaceDataKeys = keys
}

func NewNamespaceCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "describe",
			Usage:     "Describe a Namespace by its name or Id.",
			Flags:     describeNamespaceFlags,
			ArgsUsage: "namespace_name",
			Action: func(c *cli.Context) error {
				return DescribeNamespace(c)
			},
		},
		{
			Name:      "list",
			Usage:     "List all Namespaces.",
			ArgsUsage: " ",
			Action: func(c *cli.Context) error {
				return ListNamespaces(c)
			},
		},
		{
			Name:      "register",
			Usage:     "Registers a new Namespace.",
			Flags:     registerNamespaceFlags,
			ArgsUsage: "namespace_name [cluster_name...]",
			Action: func(c *cli.Context) error {
				return RegisterNamespace(c)
			},
		},
		{
			Name:      "update",
			Usage:     "Updates a Namespace.",
			Flags:     updateNamespaceFlags,
			ArgsUsage: "namespace_name [cluster_name...]",
			Action: func(c *cli.Context) error {
				return UpdateNamespace(c)
			},
		},
		{
			Name:  "delete",
			Usage: "Deletes an existing Namespace.",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    "Confirm all prompts.",
					Category: common.CategoryMain,
				},
			},
			ArgsUsage: "namespace_name",
			Action: func(c *cli.Context) error {
				return DeleteNamespace(c)
			},
		},
	}
}
