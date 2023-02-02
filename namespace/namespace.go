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
			Usage:     common.DescribeNamespaceDefinition,
			UsageText: common.NamespaceDescribeUsageText,
			Flags:     describeNamespaceFlags,
			ArgsUsage: "namespace_name",
			Action: func(c *cli.Context) error {
				return DescribeNamespace(c)
			},
		},
		{
			Name:      "list",
			Usage:     common.ListNamespacesDefinition,
			UsageText: common.NamespaceListUsageText,
			ArgsUsage: " ",
			Action: func(c *cli.Context) error {
				return ListNamespaces(c)
			},
		},
		{
			Name:      "create",
			Usage:     common.CreateNamespaceDefinition,
			UsageText: common.NamespaceCreateUsageText,
			Flags:     createNamespaceFlags,
			ArgsUsage: "namespace_name [cluster_name...]",
			Action: func(c *cli.Context) error {
				return createNamespace(c)
			},
		},
		{
			Name:      "update",
			Usage:     common.UpdateNamespaceDefinition,
			UsageText: common.NamespaceUpdateUsageText,
			Flags:     updateNamespaceFlags,
			ArgsUsage: "namespace_name [cluster_name...]",
			Action: func(c *cli.Context) error {
				return UpdateNamespace(c)
			},
		},
		{
			Name:  "delete",
			Usage: common.DeleteNamespaceDefinition,
			UsageText: common.NamespaceDeleteUsageText,
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
