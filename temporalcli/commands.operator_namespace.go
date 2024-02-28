package temporalcli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/namespace/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/replication/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *TemporalOperatorNamespaceCreateCommand) run(cctx *CommandContext, args []string) error {
	nsName := args[0]
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	var clusters []*replication.ClusterReplicationConfig
	for _, clusterName := range c.Cluster {
		clusters = append(clusters, &replication.ClusterReplicationConfig{
			ClusterName: clusterName,
		})
	}

	var data map[string]string
	if len(c.Data) > 0 {
		data, err = stringKeysValues(strings.Split(c.Data, ","))
		if err != nil {
			return err
		}
	}

	_, err = cl.WorkflowService().RegisterNamespace(cctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        nsName,
		Description:                      c.Description,
		OwnerEmail:                       c.Email,
		WorkflowExecutionRetentionPeriod: durationpb.New(c.Retention),
		Clusters:                         clusters,
		ActiveClusterName:                c.ActiveCluster,
		Data:                             data,
		IsGlobalNamespace:                c.Global,
		HistoryArchivalState:             archivalState(c.HistoryArchivalState.Value),
		HistoryArchivalUri:               c.HistoryUri,
		VisibilityArchivalState:          archivalState(c.VisibilityArchivalState.Value),
		VisibilityArchivalUri:            c.VisibilityUri,
	})
	if err != nil {
		return fmt.Errorf("unable to create namespace: %w", err)
	}
	cctx.Printer.Println(color.GreenString("Namespace %s successfully registered.", nsName))
	return nil
}

func (c *TemporalOperatorNamespaceDeleteCommand) run(cctx *CommandContext, args []string) error {
	nsName := args[0]
	yes, err := cctx.promptString(
		color.RedString("Are you sure you want to delete namespace %s? Type namespace name to confirm:", nsName),
		nsName,
		c.Yes)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("user denied confirmation")
	}

	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.OperatorService().DeleteNamespace(cctx, &operatorservice.DeleteNamespaceRequest{
		Namespace: nsName,
	})
	if err != nil {
		return fmt.Errorf("unable to delete namespace: %w", err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString("Namespace %s has been deleted.", nsName))
	return nil
}

func (c *TemporalOperatorNamespaceDescribeCommand) run(cctx *CommandContext, args []string) error {
	nsName, nsID, err := getNamespaceFromIDArgs(cctx, c.NamespaceId, args)
	if err != nil {
		return err
	}

	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().DescribeNamespace(cctx, &workflowservice.DescribeNamespaceRequest{
		Namespace: nsName,
		Id:        nsID,
	})
	if err != nil {
		return fmt.Errorf("unable to describe namespace: %w", err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	return printNamespaces(cctx, resp)
}

func (c *TemporalOperatorNamespaceListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}

	var nextPageToken []byte
	for {
		resp, err := cl.WorkflowService().ListNamespaces(cctx, &workflowservice.ListNamespacesRequest{
			NextPageToken: nextPageToken,
			PageSize:      100,
		})
		if err != nil {
			return fmt.Errorf("failed listing namespaces: %w", err)
		}

		if cctx.JSONOutput {
			// For JSON we are going to dump one line of JSON per execution
			_ = cctx.Printer.PrintStructured(resp.Namespaces, printer.StructuredOptions{})
		} else {
			_ = printNamespaces(cctx, resp.Namespaces...)
		}

		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 {
			return nil
		}
	}
}

func (c *TemporalOperatorNamespaceUpdateCommand) run(cctx *CommandContext, args []string) error {
	nsName := args[0]
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	var updateRequest *workflowservice.UpdateNamespaceRequest

	if c.PromoteGlobal {
		fmt.Printf("Will promote local namespace to global namespace for:%s, other flag will be omitted. "+
			"If it is already global namespace, this will be no-op.\n", nsName)
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:        nsName,
			PromoteNamespace: true,
		}
	} else if len(c.ActiveCluster) > 0 {
		fmt.Printf("Will set active cluster name to: %s, other flag will be omitted.\n", c.ActiveCluster)
		replicationConfig := &replication.NamespaceReplicationConfig{
			ActiveClusterName: c.ActiveCluster,
		}
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:         nsName,
			ReplicationConfig: replicationConfig,
		}
	} else {
		resp, err := cl.WorkflowService().DescribeNamespace(cctx, &workflowservice.DescribeNamespaceRequest{
			Namespace: nsName,
		})
		if err != nil {
			switch err.(type) {
			case *serviceerror.NamespaceNotFound:
				return err
			default:
				return fmt.Errorf("namespace update failed: %w", err)
			}
		}

		description := resp.NamespaceInfo.GetDescription()
		ownerEmail := resp.NamespaceInfo.GetOwnerEmail()
		retention := resp.Config.GetWorkflowExecutionRetentionTtl()

		if len(c.Description) > 0 {
			description = c.Description
		}
		if len(c.Email) > 0 {
			ownerEmail = c.Email
		}

		data := map[string]string{}
		if len(c.Data) > 0 {
			data, err = stringKeysValues(c.Data)
			if err != nil {
				return err
			}
		}

		if c.Retention > 0 {
			retention = durationpb.New(c.Retention)
		}

		var clusters []*replication.ClusterReplicationConfig
		if len(c.Cluster) > 0 {
			for _, clusterName := range c.Cluster {
				clusters = append(clusters, &replication.ClusterReplicationConfig{
					ClusterName: clusterName,
				})
			}
		}

		updateInfo := &namespace.UpdateNamespaceInfo{
			Description: description,
			OwnerEmail:  ownerEmail,
			Data:        data,
		}

		updateConfig := &namespace.NamespaceConfig{
			WorkflowExecutionRetentionTtl: retention,
			HistoryArchivalState:          archivalState(c.HistoryArchivalState.String()),
			HistoryArchivalUri:            c.HistoryUri,
			VisibilityArchivalState:       archivalState(c.VisibilityArchivalState.String()),
			VisibilityArchivalUri:         c.VisibilityUri,
		}
		replicationConfig := &replication.NamespaceReplicationConfig{
			Clusters: clusters,
		}
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:         nsName,
			UpdateInfo:        updateInfo,
			Config:            updateConfig,
			ReplicationConfig: replicationConfig,
		}
	}

	resp, err := cl.WorkflowService().UpdateNamespace(cctx, updateRequest)
	if err != nil {
		switch err.(type) {
		case *serviceerror.NamespaceNotFound:
			return err
		default:
			return fmt.Errorf("namespace update failed: %w", err)
		}
	}
	if cctx.JSONOutput {
		_ = cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString("Namespace %s update succeeded.", nsName))

	return nil
}

func printNamespaces(cctx *CommandContext, responses ...*workflowservice.DescribeNamespaceResponse) error {
	namespaces := make([]map[string]any, len(responses))
	for i, resp := range responses {
		namespaces[i] = map[string]any{
			"NamespaceInfo.Name":                   resp.NamespaceInfo.Name,
			"NamespaceInfo.Id":                     resp.NamespaceInfo.Id,
			"NamespaceInfo.Description":            resp.NamespaceInfo.Description,
			"NamespaceInfo.OwnerEmail":             resp.NamespaceInfo.OwnerEmail,
			"NamespaceInfo.State":                  resp.NamespaceInfo.State,
			"NamespaceInfo.Data":                   resp.NamespaceInfo.Data,
			"Config.WorkflowExecutionRetentionTtl": resp.Config.WorkflowExecutionRetentionTtl.AsDuration(),
			"ReplicationConfig.ActiveClusterName":  resp.ReplicationConfig.ActiveClusterName,
			"ReplicationConfig.Clusters":           resp.ReplicationConfig.Clusters,
			"Config.HistoryArchivalState":          resp.Config.HistoryArchivalState,
			"Config.VisibilityArchivalState":       resp.Config.VisibilityArchivalState,
			"IsGlobalNamespace":                    resp.IsGlobalNamespace,
			"FailoverVersion":                      resp.FailoverVersion,
			"FailoverHistory":                      resp.FailoverHistory,
			"Config.HistoryArchivalUri":            resp.Config.HistoryArchivalUri,
			"Config.VisibilityArchivalUri":         resp.Config.VisibilityArchivalUri,
		}
	}

	return cctx.Printer.PrintStructured(
		namespaces,
		printer.StructuredOptions{
			Fields: []string{
				"NamespaceInfo.Name", "NamespaceInfo.Id", "NamespaceInfo.Description",
				"NamespaceInfo.OwnerEmail", "NamespaceInfo.State", "NamespaceInfo.Data",
				"Config.WorkflowExecutionRetentionTtl", "ReplicationConfig.ActiveClusterName",
				"ReplicationConfig.Clusters", "Config.HistoryArchivalState", "Config.VisibilityArchivalState",
				"IsGlobalNamespace", "FailoverVersion", "FailoverHistory", "Config.HistoryArchivalUri",
				"Config.VisibilityArchivalUri",
			},
		},
	)
}

func archivalState(input string) enums.ArchivalState {
	switch input {
	case "disabled":
		return enums.ARCHIVAL_STATE_DISABLED
	case "enabled":
		return enums.ARCHIVAL_STATE_ENABLED
	}
	return enums.ARCHIVAL_STATE_UNSPECIFIED
}

func getNamespaceFromIDArgs(cctx *CommandContext, nsID string, args []string) (string, string, error) {
	if nsID == "" && len(args) == 0 {
		return "", "", fmt.Errorf("provide either namespace-id flag or namespace as an argument")
	}

	if nsID != "" && len(args) > 0 {
		cctx.Printer.Println(color.YellowString("Both namespace-id flag and namespace are provided. Will use namespace Id to describe namespace"))
		return "", nsID, nil
	}

	if nsID != "" {
		return "", nsID, nil
	}

	return args[0], "", nil
}
