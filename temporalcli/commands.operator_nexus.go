package temporalcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	commonpb "go.temporal.io/api/common/v1"
	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
)

func (c *TemporalOperatorNexusEndpointCreateCommand) run(cctx *CommandContext, _ []string) error {
	description, err := c.Parent.descriptionToPayload(c.Description, c.DescriptionFile)
	if err != nil {
		return err
	}
	target, err := c.Parent.endpointTargetFromArgs(c.TargetNamespace, c.TargetTaskQueue, c.TargetUrl)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("either --target-namespace and --target-task-queue or --target-url are required")
	}

	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.OperatorService().CreateNexusEndpoint(cctx, &operatorservice.CreateNexusEndpointRequest{
		Spec: &nexuspb.EndpointSpec{
			Name:        c.Name,
			Description: description,
			Target:      target,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create endpoint %q: %w", c.Name, err)
	}
	cctx.Printer.Println(color.GreenString("Endpoint %s successfully created.", c.Name))
	return nil
}

func (c *TemporalOperatorNexusEndpointDeleteCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	endpoint, err := c.Parent.getEndpointByName(cctx, cl, c.Name)
	if err != nil {
		return err
	}

	_, err = cl.OperatorService().DeleteNexusEndpoint(cctx, &operatorservice.DeleteNexusEndpointRequest{
		Id:      endpoint.Id,
		Version: endpoint.Version,
	})
	if err != nil {
		return fmt.Errorf("unable to delete endpoint %q: %w", c.Name, err)
	}
	cctx.Printer.Println(color.GreenString("Endpoint %s successfully deleted (ID %s).", c.Name, endpoint.Id))
	return nil
}

func (c *TemporalOperatorNexusEndpointGetCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	endpoint, err := c.Parent.getEndpointByName(cctx, cl, c.Name)
	if err != nil {
		return err
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(endpoint, printer.StructuredOptions{})
	}
	return printNexusEndpoints(cctx, endpoint)
}

func (c *TemporalOperatorNexusEndpointListCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	var nextPageToken []byte
	for {
		resp, err := cl.OperatorService().ListNexusEndpoints(cctx, &operatorservice.ListNexusEndpointsRequest{
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return fmt.Errorf("unable to list endpoints: %w", err)
		}

		if cctx.JSONOutput {
			for _, ep := range resp.GetEndpoints() {
				_ = cctx.Printer.PrintStructured(ep, printer.StructuredOptions{})
			}
		} else {
			return printNexusEndpoints(cctx, resp.GetEndpoints()...)
		}

		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 {
			return nil
		}
	}
}

func (c *TemporalOperatorNexusEndpointUpdateCommand) run(cctx *CommandContext, _ []string) error {
	description, err := c.Parent.descriptionToPayload(c.Description, c.DescriptionFile)
	if err != nil {
		return err
	}
	if description != nil && c.UnsetDescription {
		return fmt.Errorf("--unset-description should not be set if --description or --description-file is set")
	}

	if c.TargetNamespace != "" && c.TargetUrl != "" {
		return fmt.Errorf("provided both --target-namespace and --target-url")
	}
	if c.TargetTaskQueue != "" && c.TargetUrl != "" {
		return fmt.Errorf("provided both --target-task-queue and --target-url")
	}

	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	endpoint, err := c.Parent.getEndpointByName(cctx, cl, c.Name)
	if err != nil {
		return err
	}

	existingDescription := endpoint.GetSpec().GetDescription()
	if description == nil && !c.UnsetDescription {
		description = existingDescription
	}

	target := endpoint.GetSpec().GetTarget()
	if endpoint.GetSpec().GetTarget().GetExternal() != nil &&
		(c.TargetNamespace == "" && c.TargetTaskQueue != "" || c.TargetNamespace != "" && c.TargetTaskQueue == "") {
		return fmt.Errorf("both --target-namespace and --target-task-queue are required when changing target type from external to worker")
	}
	if c.TargetUrl != "" {
		target = &nexuspb.EndpointTarget{
			Variant: &nexuspb.EndpointTarget_External_{External: &nexuspb.EndpointTarget_External{Url: c.TargetUrl}},
		}
	} else if c.TargetNamespace != "" || c.TargetTaskQueue != "" {
		workerTarget := &nexuspb.EndpointTarget_Worker{
			Namespace: endpoint.GetSpec().GetTarget().GetWorker().GetNamespace(),
			TaskQueue: endpoint.GetSpec().GetTarget().GetWorker().GetTaskQueue(),
		}
		if c.TargetNamespace != "" {
			workerTarget.Namespace = c.TargetNamespace
		}
		if c.TargetTaskQueue != "" {
			workerTarget.TaskQueue = c.TargetTaskQueue
		}
		target = &nexuspb.EndpointTarget{
			Variant: &nexuspb.EndpointTarget_Worker_{
				Worker: workerTarget,
			},
		}
	}

	_, err = cl.OperatorService().UpdateNexusEndpoint(cctx, &operatorservice.UpdateNexusEndpointRequest{
		Id:      endpoint.Id,
		Version: endpoint.Version,
		Spec: &nexuspb.EndpointSpec{
			Name:        c.Name,
			Description: description,
			Target:      target,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to update endpoint %q: %w", c.Name, err)
	}
	cctx.Printer.Println(color.GreenString("Endpoint %s successfully updated.", c.Name))
	return nil
}

func (c *TemporalOperatorNexusEndpointCommand) descriptionToPayload(description, descriptionFile string) (*commonpb.Payload, error) {
	if description != "" && descriptionFile != "" {
		return nil, fmt.Errorf("provided both --description and --description-file")
	}
	if descriptionFile != "" {
		b, err := os.ReadFile(descriptionFile)
		if err != nil {
			return nil, fmt.Errorf("failed reading input file %q: %w", descriptionFile, err)
		}
		if len(b) == 0 {
			return nil, fmt.Errorf("empty description file: %q", descriptionFile)
		}
		description = string(b)
	}
	if description == "" {
		return nil, nil
	}
	data, err := json.Marshal(description)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal description to JSON: %w", err)
	}

	return &commonpb.Payload{
		Metadata: map[string][]byte{"encoding": []byte("json/plain")},
		Data:     data,
	}, nil
}

func (c *TemporalOperatorNexusEndpointCommand) endpointTargetFromArgs(namespace, taskQueue, url string) (*nexuspb.EndpointTarget, error) {
	if namespace != "" {
		if url != "" {
			return nil, fmt.Errorf("provided both --target-namespace and --target-url")
		}
		if taskQueue == "" {
			return nil, fmt.Errorf("both --target-namespace and --target-task-queue are required")
		}
		return &nexuspb.EndpointTarget{
			Variant: &nexuspb.EndpointTarget_Worker_{
				Worker: &nexuspb.EndpointTarget_Worker{
					Namespace: namespace,
					TaskQueue: taskQueue,
				},
			},
		}, nil
	} else if url != "" {
		if namespace != "" {
			return nil, fmt.Errorf("provided both --target-namespace and --target-url")
		}
		return &nexuspb.EndpointTarget{
			Variant: &nexuspb.EndpointTarget_External_{
				External: &nexuspb.EndpointTarget_External{
					Url: url,
				},
			},
		}, nil
	}
	return nil, fmt.Errorf("either --target-namespace and --target-task-queue or --target-url are required")
}

func (c *TemporalOperatorNexusEndpointCommand) getEndpointByName(ctx context.Context, cl client.Client, name string) (*nexuspb.Endpoint, error) {
	resp, err := cl.OperatorService().ListNexusEndpoints(ctx, &operatorservice.ListNexusEndpointsRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get endpoint %q: %w", name, err)
	}

	if len(resp.Endpoints) == 0 {
		return nil, fmt.Errorf("endpoint not found: %q", name)
	}

	// We filter by name and are guaranteed to get 0 or 1 results.
	return resp.Endpoints[0], nil
}

func printNexusEndpoints(cctx *CommandContext, endpoints ...*nexuspb.Endpoint) error {
	mapped := make([]map[string]any, len(endpoints))
	for i, ep := range endpoints {
		var description string
		if desc := ep.GetSpec().GetDescription(); desc != nil && string(desc.Metadata["encoding"]) == "json/plain" {
			if err := json.Unmarshal(desc.Data, &description); err != nil {
				return fmt.Errorf("malformed description for endpoint: %q, expected a string encoded as JSON", ep.GetSpec().GetName())
			}
		}
		mapped[i] = map[string]any{
			"ID":                      ep.Id,
			"Name":                    ep.GetSpec().GetName(),
			"CreatedTime":             ep.CreatedTime,
			"LastModifiedTime":        ep.LastModifiedTime,
			"Target.External.URL":     ep.GetSpec().GetTarget().GetExternal().GetUrl(),
			"Target.Worker.Namespace": ep.GetSpec().GetTarget().GetWorker().GetNamespace(),
			"Target.Worker.TaskQueue": ep.GetSpec().GetTarget().GetWorker().GetTaskQueue(),
			"Description":             description,
		}
	}

	return cctx.Printer.PrintStructured(
		mapped,
		printer.StructuredOptions{
			Fields: []string{
				"ID",
				"Name",
				"CreatedTime",
				"LastModifiedTime",
				"Target.External.URL",
				"Target.Worker.Namespace",
				"Target.Worker.TaskQueue",
				"Description",
			},
		},
	)
}
