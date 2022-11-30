package searchattribute

import (
	"fmt"

	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/temporal-cli/common"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
)

func NewSearchAttributeCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "create",
			Usage: "Add custom search attributes",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:     common.FlagName,
					Required: true,
					Usage:    "Search attribute name",
					Category: common.CategoryMain,
				},
				&cli.StringSliceFlag{
					Name:     common.FlagType,
					Required: true,
					Usage:    fmt.Sprintf("Search attribute type: %v", common.AllowedEnumValues(enumspb.IndexedValueType_name)),
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    "Confirm all prompts",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return AddSearchAttributes(c)
			},
		},
		{
			Name:  "list",
			Usage: "List search attributes that can be used in list workflow query",
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
			Name:  "remove",
			Usage: "Remove custom search attributes metadata only (Elasticsearch index schema is not modified)",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:     common.FlagName,
					Required: true,
					Usage:    "Search attribute name",
					Category: common.CategoryMain,
				},
				&cli.BoolFlag{
					Name:     common.FlagYes,
					Aliases:  common.FlagYesAlias,
					Usage:    "Confirm all prompts",
					Category: common.CategoryMain,
				},
			},
			Action: func(c *cli.Context) error {
				return RemoveSearchAttributes(c)
			},
		},
	}
}
