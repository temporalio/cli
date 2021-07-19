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

	"github.com/temporalio/tctl/pkg/output"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
	namespacepb "go.temporal.io/api/namespace/v1"
	replicationpb "go.temporal.io/api/replication/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"

	"go.temporal.io/server/common/primitives/timestamp"
)

// RegisterNamespace register a namespace
func RegisterNamespace(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	description := c.String(FlagDescription)
	ownerEmail := c.String(FlagOwnerEmail)

	client := cFactory.FrontendClient(c)

	retention := defaultNamespaceRetention
	var err error
	if c.IsSet(FlagRetention) {
		retention, err = timestamp.ParseDurationDefaultDays(c.String(FlagRetention))
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", FlagRetention), err)
		}
	}

	var isGlobalNamespace bool
	if c.IsSet(FlagIsGlobalNamespace) {
		isGlobalNamespace, err = strconv.ParseBool(c.String(FlagIsGlobalNamespace))
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", FlagIsGlobalNamespace), err)
		}
	}

	namespaceData := map[string]string{}
	if c.IsSet(FlagNamespaceData) {
		namespaceDataStr := getRequiredOption(c, FlagNamespaceData)
		namespaceData, err = parseNamespaceDataKVs(namespaceDataStr)
		if err != nil {
			ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", FlagNamespaceData), err)
		}
	}
	if len(requiredNamespaceDataKeys) > 0 {
		err = checkRequiredNamespaceDataKVs(namespaceData)
		if err != nil {
			ErrorAndExit("Namespace data missed required data.", err)
		}
	}

	var activeClusterName string
	if c.IsSet(FlagActiveClusterName) {
		activeClusterName = c.String(FlagActiveClusterName)
	}

	var clusters []*replicationpb.ClusterReplicationConfig
	if c.IsSet(FlagClusters) {
		clusterStr := c.String(FlagClusters)
		clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
			ClusterName: clusterStr,
		})
		for _, clusterStr := range c.Args().Slice() {
			clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
				ClusterName: clusterStr,
			})
		}
	}

	request := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        namespace,
		Description:                      description,
		OwnerEmail:                       ownerEmail,
		Data:                             namespaceData,
		WorkflowExecutionRetentionPeriod: &retention,
		Clusters:                         clusters,
		ActiveClusterName:                activeClusterName,
		HistoryArchivalState:             archivalState(c, FlagHistoryArchivalState),
		HistoryArchivalUri:               c.String(FlagHistoryArchivalURI),
		VisibilityArchivalState:          archivalState(c, FlagVisibilityArchivalState),
		VisibilityArchivalUri:            c.String(FlagVisibilityArchivalURI),
		IsGlobalNamespace:                isGlobalNamespace,
	}

	ctx, cancel := newContext(c)
	defer cancel()
	_, err = client.RegisterNamespace(ctx, request)
	if err != nil {
		if _, ok := err.(*serviceerror.NamespaceAlreadyExists); !ok {
			ErrorAndExit("Register namespace operation failed.", err)
		} else {
			ErrorAndExit(fmt.Sprintf("Namespace %s already registered.", namespace), err)
		}
	} else {
		fmt.Printf("Namespace %s successfully registered.\n", namespace)
	}
}

// UpdateNamespace updates a namespace
func UpdateNamespace(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)

	client := cFactory.FrontendClient(c)

	var updateRequest *workflowservice.UpdateNamespaceRequest
	ctx, cancel := newContext(c)
	defer cancel()

	if c.IsSet(FlagActiveClusterName) {
		activeCluster := c.String(FlagActiveClusterName)
		fmt.Printf("Will set active cluster name to: %s, other flag will be omitted.\n", activeCluster)
		replicationConfig := &replicationpb.NamespaceReplicationConfig{
			ActiveClusterName: activeCluster,
		}
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:         namespace,
			ReplicationConfig: replicationConfig,
		}
	} else {
		resp, err := client.DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{
			Namespace: namespace,
		})
		if err != nil {
			if _, ok := err.(*serviceerror.NotFound); !ok {
				ErrorAndExit("Operation UpdateNamespace failed.", err)
			} else {
				ErrorAndExit(fmt.Sprintf("Namespace %s does not exist.", namespace), err)
			}
			return
		}

		description := resp.NamespaceInfo.GetDescription()
		ownerEmail := resp.NamespaceInfo.GetOwnerEmail()
		retention := timestamp.DurationValue(resp.Config.GetWorkflowExecutionRetentionTtl())
		var clusters []*replicationpb.ClusterReplicationConfig

		if c.IsSet(FlagDescription) {
			description = c.String(FlagDescription)
		}
		if c.IsSet(FlagOwnerEmail) {
			ownerEmail = c.String(FlagOwnerEmail)
		}
		namespaceData := map[string]string{}
		if c.IsSet(FlagNamespaceData) {
			namespaceDataStr := c.String(FlagNamespaceData)
			namespaceData, err = parseNamespaceDataKVs(namespaceDataStr)
			if err != nil {
				ErrorAndExit("Namespace data format is invalid.", err)
			}
		}
		if c.IsSet(FlagRetention) {
			retention, err = timestamp.ParseDurationDefaultDays(c.String(FlagRetention))
			if err != nil {
				ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", FlagRetention), err)
			}
		}
		if c.IsSet(FlagClusters) {
			clusterStr := c.String(FlagClusters)
			clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
				ClusterName: clusterStr,
			})
			for _, clusterStr := range c.Args().Slice() {
				clusters = append(clusters, &replicationpb.ClusterReplicationConfig{
					ClusterName: clusterStr,
				})
			}
		}

		var binBinaries *namespacepb.BadBinaries
		if c.IsSet(FlagAddBadBinary) {
			if !c.IsSet(FlagReason) {
				ErrorAndExit("Must provide a reason.", nil)
			}
			binChecksum := c.String(FlagAddBadBinary)
			reason := c.String(FlagReason)
			operator := getCurrentUserFromEnv()
			binBinaries = &namespacepb.BadBinaries{
				Binaries: map[string]*namespacepb.BadBinaryInfo{
					binChecksum: {
						Reason:   reason,
						Operator: operator,
					},
				},
			}
		}

		var badBinaryToDelete string
		if c.IsSet(FlagRemoveBadBinary) {
			badBinaryToDelete = c.String(FlagRemoveBadBinary)
		}

		updateInfo := &namespacepb.UpdateNamespaceInfo{
			Description: description,
			OwnerEmail:  ownerEmail,
			Data:        namespaceData,
		}
		updateConfig := &namespacepb.NamespaceConfig{
			WorkflowExecutionRetentionTtl: &retention,
			HistoryArchivalState:          archivalState(c, FlagHistoryArchivalState),
			HistoryArchivalUri:            c.String(FlagHistoryArchivalURI),
			VisibilityArchivalState:       archivalState(c, FlagVisibilityArchivalState),
			VisibilityArchivalUri:         c.String(FlagVisibilityArchivalURI),
			BadBinaries:                   binBinaries,
		}
		replicationConfig := &replicationpb.NamespaceReplicationConfig{
			Clusters: clusters,
		}
		updateRequest = &workflowservice.UpdateNamespaceRequest{
			Namespace:         namespace,
			UpdateInfo:        updateInfo,
			Config:            updateConfig,
			ReplicationConfig: replicationConfig,
			DeleteBadBinary:   badBinaryToDelete,
		}
	}

	_, err := client.UpdateNamespace(ctx, updateRequest)
	if err != nil {
		if _, ok := err.(*serviceerror.NotFound); !ok {
			ErrorAndExit("Operation UpdateNamespace failed.", err)
		} else {
			ErrorAndExit(fmt.Sprintf("Namespace %s does not exist.", namespace), err)
		}
	} else {
		fmt.Printf("Namespace %s successfully updated.\n", namespace)
	}
}

// DescribeNamespace updates a namespace
func DescribeNamespace(c *cli.Context) {
	namespace := c.String(FlagNamespace)
	namespaceID := c.String(FlagNamespaceID)

	if namespaceID == "" && namespace == "" {
		ErrorAndExit("At least namespace_id or namespace must be provided.", nil)
	}
	if c.IsSet(FlagNamespace) && namespaceID != "" {
		ErrorAndExit("Only one of namespace_id or namespace must be provided.", nil)
	}
	if namespaceID != "" {
		namespace = ""
	}

	client := cFactory.FrontendClient(c)

	ctx, cancel := newContext(c)
	defer cancel()
	resp, err := client.DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{
		Namespace: namespace,
		Id:        namespaceID,
	})
	if err != nil {
		if _, ok := err.(*serviceerror.NotFound); !ok {
			ErrorAndExit("Operation DescribeNamespace failed.", err)
		}
		ErrorAndExit(fmt.Sprintf("Namespace %s does not exist.", namespace), err)
	}

	printNamespace(c, resp)
}

// ListNamespaces list all namespaces
func ListNamespaces(c *cli.Context) {

	client := cFactory.FrontendClient(c)
	for _, ns := range getAllNamespaces(c, client) {
		printNamespace(c, ns)
	}
}

func printNamespace(c *cli.Context, resp *workflowservice.DescribeNamespaceResponse) {
	opts := &output.PrintOptions{
		Fields:     []string{"NamespaceInfo.Name", "NamespaceInfo.Id", "NamespaceInfo.Description", "NamespaceInfo.OwnerEmail", "NamespaceInfo.State", "Config.WorkflowExecutionRetentionTtl", "ReplicationConfig.ActiveClusterName", "ReplicationConfig.Clusters", "Config.HistoryArchivalState", "Config.VisibilityArchivalState"},
		FieldsLong: []string{"Config.HistoryArchivalUri", "Config.VisibilityArchivalUri"},
		Output:     output.Card,
		NoPager:    true,
	}
	output.PrintItems(c, []interface{}{resp}, opts)

	type badBinary struct {
		Checksum   string
		Operator   string
		CreateTime string
		Reason     string
	}
	var badBinaries []interface{}

	for cs, bin := range resp.Config.BadBinaries.Binaries {
		badBinaries = append(badBinaries, badBinary{
			Checksum:   cs,
			Operator:   bin.GetOperator(),
			CreateTime: timestamp.TimeValue(bin.GetCreateTime()).String(),
			Reason:     bin.GetReason(),
		})
	}
	bOpts := &output.PrintOptions{
		Fields:      []string{"Checksum", "Operator", "CreateTime", "Reason"},
		IgnoreFlags: true,
		NoPager:     true,
	}
	output.PrintItems(c, badBinaries, bOpts)
}

func getAllNamespaces(c *cli.Context, tClient workflowservice.WorkflowServiceClient) []*workflowservice.DescribeNamespaceResponse {
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
			ErrorAndExit("Error when list namespaces info", err)
		}
		token = listResp.GetNextPageToken()
		res = append(res, listResp.GetNamespaces()...)
	}
	return res
}

func clustersToString(clusters []*replicationpb.ClusterReplicationConfig) string {
	var res string
	for i, cluster := range clusters {
		if i == 0 {
			res = res + cluster.GetClusterName()
		} else {
			res = res + ", " + cluster.GetClusterName()
		}
	}
	return res
}

func archivalState(c *cli.Context, stateFlagName string) enumspb.ArchivalState {
	if c.IsSet(stateFlagName) {
		switch c.String(stateFlagName) {
		case "disabled":
			return enumspb.ARCHIVAL_STATE_DISABLED
		case "enabled":
			return enumspb.ARCHIVAL_STATE_ENABLED
		default:
			ErrorAndExit(fmt.Sprintf("Option %s format is invalid.", stateFlagName), errors.New("invalid state, valid values are \"disabled\" and \"enabled\""))
		}
	}
	return enumspb.ARCHIVAL_STATE_UNSPECIFIED
}
