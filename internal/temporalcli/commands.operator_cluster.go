package temporalcli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

func (c *TemporalOperatorClusterHealthCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	_, err = cl.CheckHealth(cctx, &client.CheckHealthRequest{})
	if err != nil {
		return fmt.Errorf("failed checking cluster health: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			struct {
				Status string `json:"status"`
			}{"SERVING"},
			printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString("SERVING"))
	return nil
}

func (c *TemporalOperatorClusterSystemCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	resp, err := cl.WorkflowService().GetSystemInfo(cctx, &workflowservice.GetSystemInfoRequest{})
	if err != nil {
		return fmt.Errorf("unable to get system information: %w", err)
	}
	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	return cctx.Printer.PrintStructured(
		struct {
			ServerVersion      string
			SupportsSchedules  bool
			UpsertMemo         bool
			EagerWorkflowStart bool
		}{
			ServerVersion:      resp.GetServerVersion(),
			SupportsSchedules:  resp.GetCapabilities().SupportsSchedules,
			UpsertMemo:         resp.GetCapabilities().UpsertMemo,
			EagerWorkflowStart: resp.GetCapabilities().EagerWorkflowStart,
		},
		printer.StructuredOptions{
			Table: &printer.TableOptions{},
		})
}

func (c *TemporalOperatorClusterDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	resp, err := cl.WorkflowService().GetClusterInfo(cctx, &workflowservice.GetClusterInfoRequest{})
	if err != nil {
		return fmt.Errorf("unable to get cluster information: %w", err)
	}
	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	fields := []string{"ClusterName", "PersistenceStore", "VisibilityStore"}
	if c.Detail {
		fields = append(fields, "HistoryShardCount", "VersionInfo")
	}

	return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{
		Fields: fields,
		Table:  &printer.TableOptions{},
	})
}

func (c *TemporalOperatorClusterListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// This is a listing command subject to json vs jsonl rules
	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	var nextPageToken []byte
	var execsProcessed int
	for pageIndex := 0; ; pageIndex++ {
		page, err := cl.OperatorService().ListClusters(cctx, &operatorservice.ListClustersRequest{
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return fmt.Errorf("failed listing clusters: %w", err)
		}
		var textTable []map[string]any
		for _, cluster := range page.GetClusters() {
			if c.Limit > 0 && execsProcessed >= c.Limit {
				break
			}
			execsProcessed++
			// For JSON we are going to dump one line of JSON per execution
			if cctx.JSONOutput {
				_ = cctx.Printer.PrintStructured(cluster, printer.StructuredOptions{})
			} else {
				// For non-JSON, we are doing a table for each page
				textTable = append(textTable, map[string]any{
					"Name":                   cluster.ClusterName,
					"ClusterId":              cluster.ClusterId,
					"Address":                cluster.Address,
					"HistoryShardCount":      cluster.HistoryShardCount,
					"InitialFailoverVersion": cluster.InitialFailoverVersion,
					"IsConnectionEnabled":    cluster.IsConnectionEnabled,
					"IsReplicationEnabled":   cluster.IsReplicationEnabled,
				})
			}
		}
		// Print table, headers only on first table
		if len(textTable) > 0 {
			_ = cctx.Printer.PrintStructured(textTable, printer.StructuredOptions{
				Fields: []string{"Name", "ClusterId", "Address", "HistoryShardCount", "InitialFailoverVersion", "IsConnectionEnabled", "IsReplicationEnabled"},
				Table:  &printer.TableOptions{NoHeader: pageIndex > 0},
			})
		}
		// Stop if next page token non-existing or executions reached limit
		nextPageToken = page.GetNextPageToken()
		if len(nextPageToken) == 0 || (c.Limit > 0 && execsProcessed >= c.Limit) {
			return nil
		}
	}
}

func (c *TemporalOperatorClusterUpsertCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	_, err = cl.OperatorService().AddOrUpdateRemoteCluster(cctx, &operatorservice.AddOrUpdateRemoteClusterRequest{
		FrontendAddress:               c.FrontendAddress,
		EnableRemoteClusterConnection: c.EnableConnection,
		EnableReplication:             c.EnableReplication,
	})
	if err != nil {
		return fmt.Errorf("unable to upsert cluster: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			struct {
				FrontendAddress string `json:"frontendAddress"`
			}{FrontendAddress: c.FrontendAddress},
			printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString(fmt.Sprintf("Upserted cluster %s", c.FrontendAddress)))
	return nil
}

func (c *TemporalOperatorClusterRemoveCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	_, err = cl.OperatorService().RemoveRemoteCluster(cctx, &operatorservice.RemoveRemoteClusterRequest{
		ClusterName: c.Name,
	})
	if err != nil {
		return fmt.Errorf("failed removing cluster: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			struct {
				ClusterName string `json:"clusterName"`
			}{ClusterName: c.Name},
			printer.StructuredOptions{})
	}
	cctx.Printer.Println(color.GreenString(fmt.Sprintf("Removed cluster %s", c.Name)))
	return nil
}
