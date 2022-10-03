// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"errors"
	"fmt"
	"strconv"

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

// RegisterNamespace register a namespace
func RegisterNamespace(c *cli.Context) error {
	ns, err := getNamespaceFromArgs(c)
	if err != nil {
		return err
	}

	description := c.String(FlagDescription)
	ownerEmail := c.String(FlagOwnerEmail)

	client := cFactory.FrontendClient(c)

	retention := defaultNamespaceRetention
	if c.IsSet(FlagRetention) {
		retention, err = timestamp.ParseDurationDefaultDays(c.String(FlagRetention))
		if err != nil {
			return fmt.Errorf("option %s format is invalid: %w", FlagRetention, err)
		}
	}

	var isGlobalNamespace bool
	if c.IsSet(FlagIsGlobalNamespace) {
		isGlobalNamespace, err = strconv.ParseBool(c.String(FlagIsGlobalNamespace))
		if err != nil {
			return fmt.Errorf("option %s format is invalid: %w", FlagIsGlobalNamespace, err)
		}
	}

	data := map[string]string{}
	if c.IsSet(FlagNamespaceData) {
		datas := c.StringSlice(FlagNamespaceData)
		data, err = SplitKeyValuePairs(datas)
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
	if c.IsSet(FlagActiveCluster) {
		activeCluster = c.String(FlagActiveCluster)
	}

	var clusters []*replicationpb.ClusterReplicationConfig
	if c.IsSet(FlagCluster) {
		clusterNames := c.StringSlice(FlagCluster)
		for _, clusterName := range clusterNames {
			clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
				ClusterName: clusterName,
			})
		}
	}

	archState, err := archivalState(c, FlagHistoryArchivalState)
	if err != nil {
		return err
	}
	archVisState, err := archivalState(c, FlagVisibilityArchivalState)
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
		HistoryArchivalUri:               c.String(FlagHistoryArchivalURI),
		VisibilityArchivalState:          archVisState,
		VisibilityArchivalUri:            c.String(FlagVisibilityArchivalURI),
		IsGlobalNamespace:                isGlobalNamespace,
	}

	ctx, cancel := newContext(c)
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

	client := cFactory.FrontendClient(c)

	var updateRequest *workflowservice.UpdateNamespaceRequest
	ctx, cancel := newContext(c)
	defer cancel()

	if c.IsSet(FlagPromoteNamespace) && c.Bool(FlagPromoteNamespace) {
		fmt.Printf("Will promote local namespace to global namespace for:%s, other flag will be omitted. "+
			"If it is already global namespace, this will be no-op.\n", ns)
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:        ns,
			PromoteNamespace: true,
		}
	} else if c.IsSet(FlagActiveCluster) {
		activeCluster := c.String(FlagActiveCluster)
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

		if c.IsSet(FlagDescription) {
			description = c.String(FlagDescription)
		}
		if c.IsSet(FlagOwnerEmail) {
			ownerEmail = c.String(FlagOwnerEmail)
		}
		data := map[string]string{}
		if c.IsSet(FlagNamespaceData) {
			datas := c.StringSlice(FlagNamespaceData)
			data, err = SplitKeyValuePairs(datas)
			if err != nil {
				return err
			}
		}
		if c.IsSet(FlagRetention) {
			retention, err = timestamp.ParseDurationDefaultDays(c.String(FlagRetention))
			if err != nil {
				return fmt.Errorf("option %s format is invalid: %w", FlagRetention, err)
			}
		}
		var clusters []*replicationpb.ClusterReplicationConfig
		if c.IsSet(FlagCluster) {
			clusterNames := c.StringSlice(FlagCluster)
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

		archState, err := archivalState(c, FlagHistoryArchivalState)
		if err != nil {
			return err
		}
		archVisState, err := archivalState(c, FlagVisibilityArchivalState)
		if err != nil {
			return err
		}
		updateConfig := &namespacepb.NamespaceConfig{
			WorkflowExecutionRetentionTtl: &retention,
			HistoryArchivalState:          archState,
			HistoryArchivalUri:            c.String(FlagHistoryArchivalURI),
			VisibilityArchivalState:       archVisState,
			VisibilityArchivalUri:         c.String(FlagVisibilityArchivalURI),
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

	client := cFactory.FrontendClient(c)

	ctx, cancel := newContext(c)
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
	client := cFactory.FrontendClient(c)

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
	if !prompt(promptMsg, c.Bool(FlagYes), ns) {
		return nil
	}

	client := cFactory.OperatorClient(c)
	ctx, cancel := newContext(c)
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
	ctx, cancel := newContext(c)
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
		if c.IsSet(FlagNamespace) {
			errMessage = fmt.Sprintf("%s. Global flag '%s' is not supported by namespace commands", errMessage, FlagNamespace)
		}
		return "", errors.New(errMessage)
	}
	return ns, nil
}

func getNamespaceFromIDArgs(c *cli.Context) (string, string, error) {
	ns := c.Args().First()
	nsID := c.String(FlagNamespaceID)

	if nsID == "" && ns == "" {
		errMessage := fmt.Sprintf("provide either %s flag or namespace as an argument", FlagNamespaceID)
		if c.IsSet(FlagNamespace) {
			errMessage = fmt.Sprintf("%s. Global flag '%s' is not supported by namespace commands", errMessage, FlagNamespace)
		}
		return "", "", errors.New(errMessage)
	}

	if nsID != "" && ns != "" {
		fmt.Println(color.Yellow(c, "Both %s flag and namespace are provided. Will use namespace Id to describe namespace", FlagNamespaceID))
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
