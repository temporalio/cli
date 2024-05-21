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
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *TemporalOperatorCommand) getNSFromFlagOrArg0(cctx *CommandContext, args []string) (string, error) {
	if len(args) > 0 && c.Namespace != "default" {
		return "", fmt.Errorf("namespace was provided as both an argument (%s) and a flag (-n %s); please specify namespace only with -n", args[0], c.Namespace)
	}

	if len(args) > 0 {
		cctx.Logger.Warn("Passing the namespace as an argument is now deprecated; please switch to using -n instead")
		return args[0], nil
	}
	return c.Namespace, nil
}

func (c *TemporalOperatorNamespaceCreateCommand) run(cctx *CommandContext, args []string) error {
	nsName, err := c.Parent.Parent.getNSFromFlagOrArg0(cctx, args)
	if err != nil {
		return err
	}

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
		WorkflowExecutionRetentionPeriod: durationpb.New(c.Retention.Duration()),
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
		return fmt.Errorf("unable to create namespace %s: %w", nsName, err)
	}
	cctx.Printer.Println(color.GreenString("Namespace %s successfully registered.", nsName))
	return nil
}

func (c *TemporalOperatorNamespaceDeleteCommand) run(cctx *CommandContext, args []string) error {
	nsName, err := c.Parent.Parent.getNSFromFlagOrArg0(cctx, args)
	if err != nil {
		return err
	}

	yes, err := cctx.promptString(
		color.RedString("Are you sure you want to delete namespace %s? Type namespace name to confirm:", nsName),
		nsName,
		c.Yes)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("user denied confirmation or mistyped the namespace name")
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
		return fmt.Errorf("unable to delete namespace %s: %w", nsName, err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString("Namespace %s has been deleted.", nsName))
	return nil
}

func (c *TemporalOperatorNamespaceDescribeCommand) run(cctx *CommandContext, args []string) error {
	nsID := c.NamespaceId

	nsName, err := c.Parent.Parent.getNSFromFlagOrArg0(cctx, args)
	if err != nil {
		return err
	}

	// Bleh, special case: if the nsName is "default", it may not have been
	// supplied at all, and nsID should take precedence. This won't catch the
	// user explicitly specifying "default" for the name AND a UUID, but we
	// can't help that without some way to know why nsName is "default".
	if nsName == "default" && nsID != "" {
		nsName = ""
	}

	if (nsID == "" && nsName == "") || (nsID != "" && nsName != "") {
		return fmt.Errorf("provide one of --namespace-id=<uuid> or -n name, but not both")
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
		nsNameOrID := nsName
		if nsNameOrID == "" {
			nsNameOrID = nsID
		}
		return fmt.Errorf("unable to describe namespace %s: %w", nsNameOrID, err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	return printNamespaceDescriptions(cctx, resp)
}

func (c *TemporalOperatorNamespaceListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}

	// This is a listing command subject to json vs jsonl rules
	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	var nextPageToken []byte
	for {
		resp, err := cl.WorkflowService().ListNamespaces(cctx, &workflowservice.ListNamespacesRequest{
			NextPageToken: nextPageToken,
			PageSize:      100,
		})
		if err != nil {
			return fmt.Errorf("failed listing namespaces: %w", err)
		}

		for _, ns := range resp.GetNamespaces() {
			if cctx.JSONOutput {
				_ = cctx.Printer.PrintStructured(ns, printer.StructuredOptions{})
			}
		}

		if !cctx.JSONOutput {
			_ = printNamespaceDescriptions(cctx, resp.Namespaces...)
		}

		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 {
			return nil
		}
	}
}

func (c *TemporalOperatorNamespaceUpdateCommand) run(cctx *CommandContext, args []string) error {
	nsName, err := c.Parent.Parent.getNSFromFlagOrArg0(cctx, args)
	if err != nil {
		return err
	}

	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	var updateRequest *workflowservice.UpdateNamespaceRequest

	if c.PromoteGlobal && len(c.ActiveCluster) > 0 {
		return fmt.Errorf("both --promote-global and --active-cluster flags cannot be set together")
	}

	if c.PromoteGlobal {
		cctx.Printer.Printlnf("Will promote local namespace to global namespace for:%s, other flag will be omitted. "+
			"If it is already global namespace, this will be no-op.\n", nsName)
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:        nsName,
			PromoteNamespace: true,
		}
	} else if len(c.ActiveCluster) > 0 {
		cctx.Printer.Printlnf("Will set active cluster name to: %s, other flag will be omitted.\n", c.ActiveCluster)
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
			return fmt.Errorf("namespace update failed: %w", err)
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
			retention = durationpb.New(c.Retention.Duration())
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
		return fmt.Errorf("namespace update failed: %w", err)
	}

	if cctx.JSONOutput {
		_ = cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString("Namespace %s update succeeded.", nsName))

	return nil
}

func printNamespaceDescriptions(cctx *CommandContext, responses ...*workflowservice.DescribeNamespaceResponse) error {
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
