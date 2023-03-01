package searchattribute

import (
	"fmt"

	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
)

func NewSearchAttributeCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "create",
			Usage:     common.CreateSearchAttributeDefinition,
			UsageText: common.SearchAttributeCreateUsageText,
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:     common.FlagName,
					Required: true,
					Usage:    common.FlagNameSearchAttribute,
					Category: common.CategoryMain,
				},
				&cli.StringSliceFlag{
					Name:     common.FlagType,
					Required: true,
					Usage:    fmt.Sprintf("Search attribute type: %v", common.AllowedEnumValues(enumspb.IndexedValueType_name)),
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return AddSearchAttributes(c)
			},
		},
		{
			Name:      "list",
			Usage:     common.ListSearchAttributesDefinition,
			UsageText: common.SearchAttributeListUsageText,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     output.FlagOutput,
					Aliases:  common.FlagOutputAlias,
					Usage:    output.UsageText,
					Value:    string(output.Table),
					Category: common.CategoryDisplay,
				},
			},
			Action: func(c *cli.Context) error {
				return ListSearchAttributes(c)
			},
		},
		{
			Name:      "remove",
			Usage:     common.RemoveSearchAttributesDefinition,
			UsageText: common.SearchAttributeRemoveUsageText,
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:     common.FlagName,
					Required: true,
					Usage:    common.FlagNameSearchAttribute,
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    common.FlagYesDefinition,
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return RemoveSearchAttributes(c)
			},
		},
	}
}
