package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/sdk/client"
	"go.temporal.io/api/common/v1"

)

type taskQueuesInfosRowType struct {
		Name string `json:"name"`
		Type string `json:"type"`
		FirstPollerTime time.Time `json:"firstPollerTime"`
}

type deploymentType struct {
	SeriesName string `json:"seriesName"`
	BuildID string `json:"buildId"`
}

type formattedDeploymentInfoType struct {
	Deployment deploymentType `json:"deployment"`
	CreateTime time.Time `json:"createTime"`
	IsCurrent bool `json:"isCurrent"`
	TaskQueuesInfos []taskQueuesInfosRowType `json:"taskQueuesInfos,omitempty"`
	Metadata map[string]*common.Payload `json:"metadata,omitempty"`
}

type formattedDeploymentReachabilityInfoType struct {
	DeploymentInfo formattedDeploymentInfoType `json:"deploymentInfo"`
	Reachability string `json:"reachability"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

type formattedDeploymentListEntryType struct {
	SeriesName string
	BuildID string
	CreateTime time.Time
	IsCurrent bool
}

type formattedDualDeploymentInfoType struct {
	Previous formattedDeploymentInfoType  `json:"previous"`
	Current formattedDeploymentInfoType  `json:"current"`
}


func formatTaskQueuesInfos(tqis []client.DeploymentTaskQueueInfo) ([]taskQueuesInfosRowType, error) {
	var tqiRows []taskQueuesInfosRowType
	for _, tqi := range tqis {
		tqTypeStr, err  := taskQueueTypeToStr(tqi.Type)
		if err != nil {
			return tqiRows, err
		}
		tqiRows = append(tqiRows, taskQueuesInfosRowType{
			Name: tqi.Name,
			Type: tqTypeStr,
			FirstPollerTime: tqi.FirstPollerTime,
		})
	}
	return tqiRows, nil
}

func deploymentInfoToRows(deploymentInfo client.DeploymentInfo) (formattedDeploymentInfoType, error) {
	tqi, err := formatTaskQueuesInfos(deploymentInfo.TaskQueuesInfos)
	if err != nil {
		return formattedDeploymentInfoType{}, err
	}

	return formattedDeploymentInfoType{
		Deployment: deploymentType{
			SeriesName: deploymentInfo.Deployment.SeriesName,
			BuildID: deploymentInfo.Deployment.BuildID,
		},
		CreateTime: deploymentInfo.CreateTime,
		IsCurrent: deploymentInfo.IsCurrent,
		TaskQueuesInfos: tqi,
		Metadata: deploymentInfo.Metadata,
	}, nil
}

func printDeploymentInfo(cctx *CommandContext, deploymentInfo client.DeploymentInfo, msg string) error {

	fDeploymentInfo, err := deploymentInfoToRows(deploymentInfo)
	if err != nil {
		return err
	}

	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString(msg))
		printMe := struct {
			SeriesName string
			BuildID string
			CreateTime time.Time
			IsCurrent bool
			Metadata map[string]*common.Payload `cli:",cardOmitEmpty"`
		}{
			SeriesName: deploymentInfo.Deployment.SeriesName,
			BuildID: deploymentInfo.Deployment.BuildID,
			CreateTime: deploymentInfo.CreateTime,
			IsCurrent: deploymentInfo.IsCurrent,
			Metadata: deploymentInfo.Metadata,
		}
		err := cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
		if err != nil {
			return fmt.Errorf("displaying deployment info failed: %w", err)
		}

		if len(deploymentInfo.TaskQueuesInfos) > 0 {
			cctx.Printer.Println()
			cctx.Printer.Println(color.MagentaString("Task Queues:"))
			err := cctx.Printer.PrintStructured(
				deploymentInfo.TaskQueuesInfos,
				printer.StructuredOptions{Table: &printer.TableOptions{}},
			)
			if err != nil {
				return fmt.Errorf("displaying task queues info failed: %w", err)
			}
		}

		return nil
	}

	// json output
	return cctx.Printer.PrintStructured(fDeploymentInfo, printer.StructuredOptions{})
}

func deploymentReachabilityTypeToStr(reachabilityType client.DeploymentReachability) (string, error) {
	switch reachabilityType {
	case client.DeploymentReachabilityUnspecified:
		return "unspecified", nil
	case client.DeploymentReachabilityReachable:
		return "reachable", nil
	case client.DeploymentReachabilityClosedWorkflows:
		return "closed", nil
	case client.DeploymentReachabilityUnreachable:
		return "unreachable", nil
	default:
		return "", fmt.Errorf("unrecognized deployment reachability type: %d", reachabilityType)
	}
}

func printDeploymentReachabilityInfo(cctx *CommandContext, reachability client.DeploymentReachabilityInfo) error {
	fDeploymentInfo, err := deploymentInfoToRows(reachability.DeploymentInfo)
	if err != nil {
		return err
	}

	rTypeStr, err := deploymentReachabilityTypeToStr(reachability.Reachability)
	if err != nil {
		return err
	}

	fReachabilityInfo := formattedDeploymentReachabilityInfoType{
		DeploymentInfo: fDeploymentInfo,
		LastUpdateTime: reachability.LastUpdateTime,
		Reachability: rTypeStr,
	}

	if !cctx.JSONOutput {
		err := printDeploymentInfo(cctx, reachability.DeploymentInfo, "Deployment:")
		if err != nil {
			return err
		}

		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Reachability:"))
		printMe := struct {
			LastUpdateTime time.Time
			Reachability string
		}{
			LastUpdateTime: fReachabilityInfo.LastUpdateTime,
			Reachability: fReachabilityInfo.Reachability,
		}
		return cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
	}

	// json output
	return cctx.Printer.PrintStructured(fReachabilityInfo, printer.StructuredOptions{})
}

func printDeploymentSetCurrentResponse(cctx *CommandContext, response client.DeploymentSetCurrentResponse) error {

	if !cctx.JSONOutput {
		err := printDeploymentInfo(cctx, response.Previous, "Previous Deployment:")
		if err != nil {
			 return fmt.Errorf("displaying previous deployment info failed: %w", err)
		}

		err = printDeploymentInfo(cctx, response.Current, "Current Deployment:")
		if err != nil {
			 return fmt.Errorf("displaying current deployment info failed: %w", err)
		}

		return nil
	}

	previous, err := deploymentInfoToRows(response.Previous)
	if err != nil {
		return fmt.Errorf("displaying previous deployment info failed: %w", err)
	}
	current, err := deploymentInfoToRows(response.Current)
	if err != nil {
		return fmt.Errorf("displaying current deployment info failed: %w", err)
	}

	return cctx.Printer.PrintStructured(formattedDualDeploymentInfoType{
		Previous: previous,
		Current: current,
	}, printer.StructuredOptions{})
}

func (c *TemporalDeploymentDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	if c.ReportReachability {
		// Expensive call, rate-limited by target method
		resp, err := cl.DeploymentClient().GetReachability(cctx, client.DeploymentGetReachabilityOptions{
			Deployment: client.Deployment{
				SeriesName: c.DeploymentSeriesName,
				BuildID: c.DeploymentBuildId,
			},
		})
		if err != nil {
			return fmt.Errorf("error describing deployment with reachability: %w", err)
		}

		err = printDeploymentReachabilityInfo(cctx, resp)
		if err != nil {
			return err
		}
	} else {
		resp, err := cl.DeploymentClient().Describe(cctx, client.DeploymentDescribeOptions{
			Deployment: client.Deployment{
				SeriesName: c.DeploymentSeriesName,
				BuildID: c.DeploymentBuildId,
			},
		})
		if err != nil {
			return fmt.Errorf("error describing deployment: %w", err)
		}
		err = printDeploymentInfo(cctx, resp.DeploymentInfo, "Deployment:")
		if err != nil {
			return err
		}

	}

	return nil
}


func (c *TemporalDeploymentGetCurrentCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.DeploymentClient().GetCurrent(cctx, client.DeploymentGetCurrentOptions{
		SeriesName: c.DeploymentSeriesName,
	})
	if err != nil {
		return fmt.Errorf("error getting the current deployment: %w", err)
	}

	err = printDeploymentInfo(cctx, resp.DeploymentInfo, "Current Deployment:")
	if err != nil {
		return err
	}

	return nil
}


func (c *TemporalDeploymentListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	res, err := cl.DeploymentClient().List(cctx, client.DeploymentListOptions{
		SeriesName: c.DeploymentSeriesName,
	})
	if err != nil {
		return err
	}

	// This is a listing command subject to json vs jsonl rules
	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	printTableOpts := printer.StructuredOptions{
		Table: &printer.TableOptions{},
	}

	// make artificial "pages" so we get better aligned columns
	page := make([]*formattedDeploymentListEntryType, 0, 100)

	for res.HasNext() {
		entry, err := res.Next()
		if err != nil {
			return err
		}
		listEntry := formattedDeploymentInfoType{
			Deployment: deploymentType{
				SeriesName: entry.Deployment.SeriesName,
				BuildID: entry.Deployment.BuildID,
			},
			CreateTime: entry.CreateTime,
			IsCurrent: entry.IsCurrent,
		}
		if cctx.JSONOutput {
			// For JSON dump one line of JSON per deployment
			_ = cctx.Printer.PrintStructured(listEntry, printer.StructuredOptions{})
		} else {
			// For non-JSON, we are doing a table for each page
			page = append(page, &formattedDeploymentListEntryType{
				SeriesName: listEntry.Deployment.SeriesName,
				BuildID: listEntry.Deployment.BuildID,
				CreateTime: listEntry.CreateTime,
				IsCurrent: listEntry.IsCurrent,
			})
			if len(page) == cap(page) {
				_ = cctx.Printer.PrintStructured(page, printTableOpts)
				page = page[:0]
				printTableOpts.Table.NoHeader = true
			}
		}
	}

	if !cctx.JSONOutput {
		// Last partial page for non-JSON
		_ = cctx.Printer.PrintStructured(page, printTableOpts)
	}

	return nil
}


func (c *TemporalDeploymentUpdateCurrentCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	metadata, err := stringKeysJSONValues(c.DeploymentMetadata, false);
	if err != nil {
		return fmt.Errorf("invalid metadata values: %w", err)
	}

	resp, err := cl.DeploymentClient().SetCurrent(cctx, client.DeploymentSetCurrentOptions{
		Deployment: client.Deployment{
			SeriesName: c.DeploymentSeriesName,
			BuildID: c.DeploymentBuildId,
		},
		MetadataUpdate: client.DeploymentMetadataUpdate{
			UpsertEntries: metadata,
		},
	})
	if err != nil {
		return fmt.Errorf("error updating the current deployment: %w", err)
	}

	err = printDeploymentSetCurrentResponse(cctx, resp)
	if err != nil {
		return err
	}

	cctx.Printer.Println("Successfully updating the current deployment")
	return nil
}
