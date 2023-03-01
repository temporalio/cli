package searchattribute

import (
	"fmt"
	"sort"
	"time"

	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
)

const (
	addSearchAttributesTimeout = 30 * time.Second
)

// ListSearchAttributes lists search attributes
func ListSearchAttributes(c *cli.Context) error {
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	client := client.CFactory.OperatorClient(c)
	ctx, cancel := common.NewContext(c)
	defer cancel()

	request := &operatorservice.ListSearchAttributesRequest{
		Namespace: namespace,
	}
	resp, err := client.ListSearchAttributes(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to list search attributes: %w", err)
	}

	var items []interface{}
	type sa struct {
		Name string
		Type string
	}
	for saName, saType := range resp.GetSystemAttributes() {
		items = append(items, sa{Name: saName, Type: saType.String()})
	}
	for saName, saType := range resp.GetCustomAttributes() {
		items = append(items, sa{Name: saName, Type: saType.String()})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].(sa).Name < items[j].(sa).Name
	})

	opts := &output.PrintOptions{
		Fields: []string{"Name", "Type"},
	}
	return output.PrintItems(c, items, opts)
}

// AddSearchAttributes to add search attributes
func AddSearchAttributes(c *cli.Context) error {
	names := c.StringSlice(common.FlagName)
	typeStrs := c.StringSlice(common.FlagType)

	if len(names) != len(typeStrs) {
		return fmt.Errorf("number of --%s and --%s options should be the same", common.FlagName, common.FlagType)
	}

	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	client := client.CFactory.OperatorClient(c)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	listReq := &operatorservice.ListSearchAttributesRequest{
		Namespace: namespace,
	}
	existingSearchAttributes, err := client.ListSearchAttributes(ctx, listReq)
	if err != nil {
		return fmt.Errorf("unable to get existing search attributes: %w", err)
	}

	searchAttributes := make(map[string]enumspb.IndexedValueType, len(typeStrs))
	for i := 0; i < len(typeStrs); i++ {
		typeStr := typeStrs[i]

		typeInt, err := common.StringToEnum(typeStr, enumspb.IndexedValueType_value)
		if err != nil {
			return fmt.Errorf("unable to parse search attribute type %s: %w", typeStr, err)
		}
		existingSearchAttributeType, searchAttributeExists := existingSearchAttributes.CustomAttributes[names[i]]
		if !searchAttributeExists {
			searchAttributes[names[i]] = enumspb.IndexedValueType(typeInt)
			continue
		}
		if existingSearchAttributeType != enumspb.IndexedValueType(typeInt) {
			return fmt.Errorf("search attribute %s already exists and has different type %s: %w", names[i], existingSearchAttributeType, err)
		}
	}

	if len(searchAttributes) == 0 {
		fmt.Println(color.Yellow(c, "Search attributes already exist"))
		return nil
	}

	request := &operatorservice.AddSearchAttributesRequest{
		SearchAttributes: searchAttributes,
		Namespace:        namespace,
	}

	ctx, cancel = common.NewContextWithTimeout(c, addSearchAttributesTimeout)
	defer cancel()
	_, err = client.AddSearchAttributes(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to add search attributes: %w", err)
	}
	fmt.Println(color.Green(c, "Search attributes have been added"))
	return nil
}

// RemoveSearchAttributes to add search attributes
func RemoveSearchAttributes(c *cli.Context) error {
	names := c.StringSlice(common.FlagName)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(
		"You are about to remove search attributes %s. Continue? Y/N",
		color.Yellow(c, "%v", names),
	)
	if !common.PromptYes(promptMsg, c.Bool(common.FlagYes)) {
		return nil
	}

	client := client.CFactory.OperatorClient(c)
	ctx, cancel := common.NewContext(c)
	defer cancel()
	request := &operatorservice.RemoveSearchAttributesRequest{
		SearchAttributes: names,
		Namespace:        namespace,
	}

	_, err = client.RemoveSearchAttributes(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to remove search attributes: %w", err)
	}
	fmt.Println(color.Green(c, "Search attributes have been removed"))
	return nil
}
