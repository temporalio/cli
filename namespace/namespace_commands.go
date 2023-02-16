package namespace

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
	namespacepb "go.temporal.io/api/namespace/v1"
	"go.temporal.io/api/operatorservice/v1"
	replicationpb "go.temporal.io/api/replication/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/primitives/timestamp"
)

// createNamespace register a namespace
func createNamespace(c *cli.Context) error {
	ns, err := getNamespaceFromArgs(c)
	if err != nil {
		return err
	}

	description := c.String(common.FlagDescription)
	ownerEmail := c.String(common.FlagOwnerEmail)

	client := client.CFactory.FrontendClient(c)

	retention := common.DefaultNamespaceRetention
	if c.IsSet(common.FlagRetention) {
		retention, err = timestamp.ParseDurationDefaultDays(c.String(common.FlagRetention))
		if err != nil {
			return fmt.Errorf("option %s format is invalid: %w", common.FlagRetention, err)
		}
	}

	var isGlobalNamespace bool
	if c.IsSet(common.FlagIsGlobalNamespace) {
		isGlobalNamespace, err = strconv.ParseBool(c.String(common.FlagIsGlobalNamespace))
		if err != nil {
			return fmt.Errorf("option %s format is invalid: %w", common.FlagIsGlobalNamespace, err)
		}
	}

	data := map[string]string{}
	if c.IsSet(common.FlagNamespaceData) {
		datas := c.StringSlice(common.FlagNamespaceData)
		data, err = common.SplitKeyValuePairs(datas)
		if err != nil {
			return err
		}
	}
	if len(requiredNamespaceDataKeys) > 0 {
		err = validateNamespaceDataRequiredKeys(data)
		if err != nil {
			return err
		}
	}

	var activeCluster string
	if c.IsSet(common.FlagActiveCluster) {
		activeCluster = c.String(common.FlagActiveCluster)
	}

	var clusters []*replicationpb.ClusterReplicationConfig
	if c.IsSet(common.FlagCluster) {
		clusterNames := c.StringSlice(common.FlagCluster)
		for _, clusterName := range clusterNames {
			clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
				ClusterName: clusterName,
			})
		}
	}

	archState, err := archivalState(c, common.FlagHistoryArchivalState)
	if err != nil {
		return err
	}
	archVisState, err := archivalState(c, common.FlagVisibilityArchivalState)
	if err != nil {
		return err
	}

	request := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        ns,
		Description:                      description,
		OwnerEmail:                       ownerEmail,
		Data:                             data,
		WorkflowExecutionRetentionPeriod: &retention,
		Clusters:                         clusters,
		ActiveClusterName:                activeCluster,
		HistoryArchivalState:             archState,
		HistoryArchivalUri:               c.String(common.FlagHistoryArchivalURI),
		VisibilityArchivalState:          archVisState,
		VisibilityArchivalUri:            c.String(common.FlagVisibilityArchivalURI),
		IsGlobalNamespace:                isGlobalNamespace,
	}

	ctx, cancel := common.NewContext(c)
	defer cancel()
	_, err = client.RegisterNamespace(ctx, request)
	if err != nil {
		if _, ok := err.(*serviceerror.NamespaceAlreadyExists); !ok {
			return fmt.Errorf("namespace registration failed: %w", err)
		} else {
			return fmt.Errorf("namespace %s is already registered: %w", ns, err)
		}
	} else {
		fmt.Printf("Namespace %s successfully registered.\n", ns)
	}

	return nil
}

// UpdateNamespace updates a namespace
func UpdateNamespace(c *cli.Context) error {
	ns, err := getNamespaceFromArgs(c)
	if err != nil {
		return err
	}

	client := client.CFactory.FrontendClient(c)

	var updateRequest *workflowservice.UpdateNamespaceRequest
	ctx, cancel := common.NewContext(c)
	defer cancel()

	if c.IsSet(common.FlagPromoteNamespace) && c.Bool(common.FlagPromoteNamespace) {
		fmt.Printf("Will promote local namespace to global namespace for:%s, other flag will be omitted. "+
			"If it is already global namespace, this will be no-op.\n", ns)
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:        ns,
			PromoteNamespace: true,
		}
	} else if c.IsSet(common.FlagActiveCluster) {
		activeCluster := c.String(common.FlagActiveCluster)
		fmt.Printf("Will set active cluster name to: %s, other flag will be omitted.\n", activeCluster)
		replicationConfig := &replicationpb.NamespaceReplicationConfig{
			ActiveClusterName: activeCluster,
		}
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:         ns,
			ReplicationConfig: replicationConfig,
		}
	} else {
		resp, err := client.DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{
			Namespace: ns,
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
		retention := timestamp.DurationValue(resp.Config.GetWorkflowExecutionRetentionTtl())

		if c.IsSet(common.FlagDescription) {
			description = c.String(common.FlagDescription)
		}
		if c.IsSet(common.FlagOwnerEmail) {
			ownerEmail = c.String(common.FlagOwnerEmail)
		}
		data := map[string]string{}
		if c.IsSet(common.FlagNamespaceData) {
			datas := c.StringSlice(common.FlagNamespaceData)
			data, err = common.SplitKeyValuePairs(datas)
			if err != nil {
				return err
			}
		}
		if c.IsSet(common.FlagRetention) {
			retention, err = timestamp.ParseDurationDefaultDays(c.String(common.FlagRetention))
			if err != nil {
				return fmt.Errorf("option %s format is invalid: %w", common.FlagRetention, err)
			}
		}
		var clusters []*replicationpb.ClusterReplicationConfig
		if c.IsSet(common.FlagCluster) {
			clusterNames := c.StringSlice(common.FlagCluster)
			for _, clusterName := range clusterNames {
				clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
					ClusterName: clusterName,
				})
			}
		}

		updateInfo := &namespacepb.UpdateNamespaceInfo{
			Description: description,
			OwnerEmail:  ownerEmail,
			Data:        data,
		}

		archState, err := archivalState(c, common.FlagHistoryArchivalState)
		if err != nil {
			return err
		}
		archVisState, err := archivalState(c, common.FlagVisibilityArchivalState)
		if err != nil {
			return err
		}
		updateConfig := &namespacepb.NamespaceConfig{
			WorkflowExecutionRetentionTtl: &retention,
			HistoryArchivalState:          archState,
			HistoryArchivalUri:            c.String(common.FlagHistoryArchivalURI),
			VisibilityArchivalState:       archVisState,
			VisibilityArchivalUri:         c.String(common.FlagVisibilityArchivalURI),
		}
		replicationConfig := &replicationpb.NamespaceReplicationConfig{
			Clusters: clusters,
		}
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:         ns,
			UpdateInfo:        updateInfo,
			Config:            updateConfig,
			ReplicationConfig: replicationConfig,
		}
	}

	_, err = client.UpdateNamespace(ctx, updateRequest)
	if err != nil {
		switch err.(type) {
		case *serviceerror.NamespaceNotFound:
			return err
		default:
			return fmt.Errorf("namespace update failed: %w", err)
		}
	} else {
		fmt.Printf("Namespace %s successfully updated.\n", ns)
	}

	return nil
}

// DescribeNamespace updates a namespace
func DescribeNamespace(c *cli.Context) error {
	ns, nsID, err := getNamespaceFromIDArgs(c)
	if err != nil {
		return err
	}

	client := client.CFactory.FrontendClient(c)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	resp, err := client.DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{
		Namespace: ns,
		Id:        nsID,
	})
	if err != nil {
		switch err.(type) {
		case *serviceerror.NamespaceNotFound:
			return err
		default:
			return fmt.Errorf("namespace describe failed: %w", err)
		}
	}

	printNamespace(c, resp)

	return nil
}

// ListNamespaces list all namespaces
func ListNamespaces(c *cli.Context) error {
	client := client.CFactory.FrontendClient(c)

	namespaces, err := getAllNamespaces(c, client)
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		printNamespace(c, ns)
	}

	return nil
}

// DeleteNamespace deletes namespace.
func DeleteNamespace(c *cli.Context) error {
	ns, err := getNamespaceFromArgs(c)
	if err != nil {
		return err
	}

	promptMsg := color.Red(c, "Are you sure you want to delete namespace %s? Type namespace name to confirm:", ns)
	if !common.Prompt(promptMsg, c.Bool(common.FlagYes), ns) {
		return nil
	}

	client := client.CFactory.OperatorClient(c)
	ctx, cancel := common.NewContext(c)
	defer cancel()
	_, err = client.DeleteNamespace(ctx, &operatorservice.DeleteNamespaceRequest{
		Namespace: ns,
	})
	if err != nil {
		switch err.(type) {
		case *serviceerror.NamespaceNotFound:
			// Message is already good enough.
			return err
		default:
			return fmt.Errorf("unable to delete namespace: %w", err)
		}
	}

	fmt.Println(color.Green(c, "Namespace %s has been deleted", ns))
	return nil
}

func printNamespace(c *cli.Context, resp *workflowservice.DescribeNamespaceResponse) error {
	po := &output.PrintOptions{
		Fields:       []string{"NamespaceInfo.Name", "NamespaceInfo.Id", "NamespaceInfo.Description", "NamespaceInfo.OwnerEmail", "NamespaceInfo.State", "Config.WorkflowExecutionRetentionTtl", "ReplicationConfig.ActiveClusterName", "ReplicationConfig.Clusters", "Config.HistoryArchivalState", "Config.VisibilityArchivalState", "IsGlobalNamespace", "FailoverVersion", "FailoverHistory"},
		FieldsLong:   []string{"Config.HistoryArchivalUri", "Config.VisibilityArchivalUri"},
		OutputFormat: output.Card,
	}
	return output.PrintItems(c, []interface{}{resp}, po)
}

func getAllNamespaces(c *cli.Context, tClient workflowservice.WorkflowServiceClient) ([]*workflowservice.DescribeNamespaceResponse, error) {
	var res []*workflowservice.DescribeNamespaceResponse
	pagesize := int32(200)
	var token []byte
	ctx, cancel := common.NewContext(c)
	defer cancel()
	for more := true; more; more = len(token) > 0 {
		listRequest := &workflowservice.ListNamespacesRequest{
			PageSize:      pagesize,
			NextPageToken: token,
		}
		listResp, err := tClient.ListNamespaces(ctx, listRequest)
		if err != nil {
			return nil, fmt.Errorf("unable to list namespaces: %w", err)
		}
		token = listResp.GetNextPageToken()
		res = append(res, listResp.GetNamespaces()...)
	}
	return res, nil
}

func archivalState(c *cli.Context, stateFlagName string) (enumspb.ArchivalState, error) {
	if c.IsSet(stateFlagName) {
		switch c.String(stateFlagName) {
		case "disabled":
			return enumspb.ARCHIVAL_STATE_DISABLED, nil
		case "enabled":
			return enumspb.ARCHIVAL_STATE_ENABLED, nil
		default:
			return 0, fmt.Errorf("option %s format is invalid: invalid state, valid values are \"disabled\" and \"enabled\"", stateFlagName)
		}
	}
	return enumspb.ARCHIVAL_STATE_UNSPECIFIED, nil
}

func getNamespaceFromArgs(c *cli.Context) (string, error) {
	ns := c.Args().First()
	if ns == "" {
		errMessage := "provide namespace as an argument"
		if c.IsSet(common.FlagNamespace) {
			errMessage = fmt.Sprintf("%s. Command flag '--%s' is not supported by namespace commands", errMessage, common.FlagNamespace)
		}
		return "", errors.New(errMessage)
	}
	return ns, nil
}

func getNamespaceFromIDArgs(c *cli.Context) (string, string, error) {
	ns := c.Args().First()
	nsID := c.String(common.FlagNamespaceID)

	if nsID == "" && ns == "" {
		errMessage := fmt.Sprintf("provide either %s flag or namespace as an argument", common.FlagNamespaceID)
		if c.IsSet(common.FlagNamespace) {
			errMessage = fmt.Sprintf("%s. Global flag '%s' is not supported by namespace commands", errMessage, common.FlagNamespace)
		}
		return "", "", errors.New(errMessage)
	}

	if nsID != "" && ns != "" {
		fmt.Println(color.Yellow(c, "Both %s flag and namespace are provided. Will use namespace Id to describe namespace", common.FlagNamespaceID))
		ns = ""
	}

	return ns, nsID, nil
}

func validateNamespaceDataRequiredKeys(namespaceData map[string]string) error {
	for _, k := range requiredNamespaceDataKeys {
		_, ok := namespaceData[k]
		if !ok {
			return fmt.Errorf("missing namespace data key %v. Required keys: %v", k, requiredNamespaceDataKeys)
		}
	}
	return nil
}
