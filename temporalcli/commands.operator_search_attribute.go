package temporalcli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
)

func (c *TemporalOperatorSearchAttributeCreateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Name and Type are required parameters, and there must be the same number of them.

	listReq := &operatorservice.ListSearchAttributesRequest{
		Namespace: c.Parent.Parent.Namespace,
	}
	existingSearchAttributes, err := cl.OperatorService().ListSearchAttributes(cctx, listReq)
	if err != nil {
		return fmt.Errorf("unable to get existing search attributes: %w", err)
	}

	searchAttributes := make(map[string]enums.IndexedValueType, len(c.Type.Values))
	for i, saType := range c.Type.Values {
		saName := c.Name[i]
		typeInt, err := searchAttributeTypeStringToEnum(saType)
		if err != nil {
			return fmt.Errorf("unable to parse search attribute type %s: %w", saType, err)
		}
		existingSearchAttributeType, searchAttributeExists := existingSearchAttributes.CustomAttributes[saName]
		if searchAttributeExists && existingSearchAttributeType != typeInt {
			return fmt.Errorf("search attribute %s already exists and has different type %s", saName, existingSearchAttributeType.String())
		}
		searchAttributes[saName] = typeInt
	}

	request := &operatorservice.AddSearchAttributesRequest{
		SearchAttributes: searchAttributes,
		Namespace:        c.Parent.Parent.Namespace,
	}

	_, err = cl.OperatorService().AddSearchAttributes(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to add search attributes: %w", err)
	}
	cctx.Printer.Println(color.GreenString("Search attributes have been added"))
	return nil
}

func searchAttributeTypeStringToEnum(search string) (enums.IndexedValueType, error) {
	for k, v := range enums.IndexedValueType_shorthandValue {
		if strings.EqualFold(search, k) {
			return enums.IndexedValueType(v), nil
		}
	}
	return enums.INDEXED_VALUE_TYPE_UNSPECIFIED, fmt.Errorf("unsupported search attribute type: %v", search)
}

func (c *TemporalOperatorSearchAttributeRemoveCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	yes, err := cctx.promptYes(
		fmt.Sprintf("You are about to remove search attribute %q? y/N", c.Name), c.Yes)
	if err != nil {
		return err
	} else if !yes {
		// We consider this a command failure
		return fmt.Errorf("user denied confirmation, operation aborted")
	}

	names := c.Name
	namespace := c.Parent.Parent.Namespace
	if err != nil {
		return err
	}

	request := &operatorservice.RemoveSearchAttributesRequest{
		SearchAttributes: names,
		Namespace:        namespace,
	}

	_, err = cl.OperatorService().RemoveSearchAttributes(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to remove search attributes: %w", err)
	}

	// response contains nothing
	cctx.Printer.Println(color.GreenString("Search attributes have been removed"))
	return nil
}

func (c *TemporalOperatorSearchAttributeListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	request := &operatorservice.ListSearchAttributesRequest{
		Namespace: c.Parent.Parent.Namespace,
	}
	resp, err := cl.OperatorService().ListSearchAttributes(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to list search attributes: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	type saNameType struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	var sas []saNameType
	for saName, saType := range resp.SystemAttributes {
		sas = append(sas, saNameType{
			Name: saName,
			Type: saType.String(),
		})
	}
	for saName, saType := range resp.CustomAttributes {
		sas = append(sas, saNameType{
			Name: saName,
			Type: saType.String(),
		})
	}

	return cctx.Printer.PrintStructured(sas, printer.StructuredOptions{Table: &printer.TableOptions{}})
}
