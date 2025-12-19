package temporalcli

import (
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type versionSummariesRowType struct {
	DeploymentName string    `json:"deploymentName"`
	BuildID        string    `json:"BuildID"`
	DrainageStatus string    `json:"drainageStatus"`
	CreateTime     time.Time `json:"createTime"`
}

type formattedRoutingConfigType struct {
	CurrentVersionDeploymentName        string    `json:"currentVersionDeploymentName"`
	CurrentVersionBuildID               string    `json:"currentVersionBuildID"`
	RampingVersionDeploymentName        string    `json:"rampingVersionDeploymentName"`
	RampingVersionBuildID               string    `json:"rampingVersionBuildID"`
	RampingVersionPercentage            float32   `json:"rampingVersionPercentage"`
	CurrentVersionChangedTime           time.Time `json:"currentVersionChangedTime"`
	RampingVersionChangedTime           time.Time `json:"rampingVersionChangedTime"`
	RampingVersionPercentageChangedTime time.Time `json:"rampingVersionPercentageChangedTime"`
}

type formattedWorkerDeploymentInfoType struct {
	Name                 string                     `json:"name"`
	CreateTime           time.Time                  `json:"createTime"`
	LastModifierIdentity string                     `json:"lastModifierIdentity"`
	RoutingConfig        formattedRoutingConfigType `json:"routingConfig"`
	VersionSummaries     []versionSummariesRowType  `json:"versionSummaries"`
	ManagerIdentity      string                     `json:"managerIdentity"`
}

type formattedWorkerDeploymentListEntryType struct {
	Name                         string
	CreateTime                   time.Time
	CurrentVersionDeploymentName string  `cli:",cardOmitEmpty"`
	CurrentVersionBuildID        string  `cli:",cardOmitEmpty"`
	RampingVersionDeploymentName string  `cli:",cardOmitEmpty"`
	RampingVersionBuildID        string  `cli:",cardOmitEmpty"`
	RampingVersionPercentage     float32 `cli:",cardOmitEmpty"`
}

type formattedDrainageInfo struct {
	DrainageStatus  string    `json:"drainageStatus"`
	LastChangedTime time.Time `json:"lastChangedTime"`
	LastCheckedTime time.Time `json:"lastCheckedTime"`
}

type formattedTaskQueueInfoRowType struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type formattedWorkerDeploymentVersionInfoType struct {
	DeploymentName     string                          `json:"deploymentName"`
	BuildID            string                          `json:"BuildID"`
	CreateTime         time.Time                       `json:"createTime"`
	RoutingChangedTime time.Time                       `json:"routingChangedTime"`
	CurrentSinceTime   time.Time                       `json:"currentSinceTime"`
	RampingSinceTime   time.Time                       `json:"rampingSinceTime"`
	RampPercentage     float32                         `json:"rampPercentage"`
	DrainageInfo       formattedDrainageInfo           `json:"drainageInfo"`
	TaskQueuesInfos    []formattedTaskQueueInfoRowType `json:"taskQueuesInfos"`
	Metadata           map[string]*common.Payload      `json:"metadata"`
}

func drainageStatusToStr(drainage client.WorkerDeploymentVersionDrainageStatus) (string, error) {
	switch drainage {
	case client.WorkerDeploymentVersionDrainageStatusUnspecified:
		return "unspecified", nil
	case client.WorkerDeploymentVersionDrainageStatusDraining:
		return "draining", nil
	case client.WorkerDeploymentVersionDrainageStatusDrained:
		return "drained", nil
	default:
		return "", fmt.Errorf("unrecognized drainage status: %d", drainage)
	}
}

func formatVersionSummaries(vss []client.WorkerDeploymentVersionSummary) ([]versionSummariesRowType, error) {
	var vsRows []versionSummariesRowType
	for _, vs := range vss {
		drainageStr, err := drainageStatusToStr(vs.DrainageStatus)
		if err != nil {
			return vsRows, err
		}
		vsRows = append(vsRows, versionSummariesRowType{
			DeploymentName: vs.Version.DeploymentName,
			BuildID:        vs.Version.BuildID,
			CreateTime:     vs.CreateTime,
			DrainageStatus: drainageStr,
		})
	}
	return vsRows, nil
}

func formatRoutingConfig(rc client.WorkerDeploymentRoutingConfig) (formattedRoutingConfigType, error) {
	cvdn := ""
	cvbid := ""
	rvdn := ""
	rvbid := ""
	if rc.CurrentVersion != nil {
		cvdn = rc.CurrentVersion.DeploymentName
		cvbid = rc.CurrentVersion.BuildID
	}
	if rc.RampingVersion != nil {
		rvdn = rc.RampingVersion.DeploymentName
		rvbid = rc.RampingVersion.BuildID
	}
	return formattedRoutingConfigType{
		CurrentVersionDeploymentName:        cvdn,
		CurrentVersionBuildID:               cvbid,
		RampingVersionDeploymentName:        rvdn,
		RampingVersionBuildID:               rvbid,
		RampingVersionPercentage:            rc.RampingVersionPercentage,
		CurrentVersionChangedTime:           rc.CurrentVersionChangedTime,
		RampingVersionChangedTime:           rc.RampingVersionChangedTime,
		RampingVersionPercentageChangedTime: rc.RampingVersionPercentageChangedTime,
	}, nil
}

func workerDeploymentInfoToRows(deploymentInfo client.WorkerDeploymentInfo) (formattedWorkerDeploymentInfoType, error) {
	vs, err := formatVersionSummaries(deploymentInfo.VersionSummaries)
	if err != nil {
		return formattedWorkerDeploymentInfoType{}, err
	}

	rc, err := formatRoutingConfig(deploymentInfo.RoutingConfig)
	if err != nil {
		return formattedWorkerDeploymentInfoType{}, err
	}

	return formattedWorkerDeploymentInfoType{
		Name:                 deploymentInfo.Name,
		LastModifierIdentity: deploymentInfo.LastModifierIdentity,
		CreateTime:           deploymentInfo.CreateTime,
		RoutingConfig:        rc,
		VersionSummaries:     vs,
		ManagerIdentity:      deploymentInfo.ManagerIdentity,
	}, nil
}

func printWorkerDeploymentInfo(cctx *CommandContext, deploymentInfo client.WorkerDeploymentInfo, msg string) error {

	fDeploymentInfo, err := workerDeploymentInfoToRows(deploymentInfo)
	if err != nil {
		return err
	}

	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString(msg))
		curVerDepName := ""
		curVerBuildId := ""
		rampVerDepName := ""
		rampVerBuildId := ""
		if deploymentInfo.RoutingConfig.CurrentVersion != nil {
			curVerDepName = deploymentInfo.RoutingConfig.CurrentVersion.DeploymentName
			curVerBuildId = deploymentInfo.RoutingConfig.CurrentVersion.BuildID
		}
		if deploymentInfo.RoutingConfig.RampingVersion != nil {
			rampVerDepName = deploymentInfo.RoutingConfig.RampingVersion.DeploymentName
			rampVerBuildId = deploymentInfo.RoutingConfig.RampingVersion.BuildID
		}
		printMe := struct {
			Name                                string
			CreateTime                          time.Time
			LastModifierIdentity                string    `cli:",cardOmitEmpty"`
			ManagerIdentity                     string    `cli:",cardOmitEmpty"`
			CurrentVersionDeploymentName        string    `cli:",cardOmitEmpty"`
			CurrentVersionBuildID               string    `cli:",cardOmitEmpty"`
			RampingVersionDeploymentName        string    `cli:",cardOmitEmpty"`
			RampingVersionBuildID               string    `cli:",cardOmitEmpty"`
			RampingVersionPercentage            float32   `cli:",cardOmitEmpty"`
			CurrentVersionChangedTime           time.Time `cli:",cardOmitEmpty"`
			RampingVersionChangedTime           time.Time `cli:",cardOmitEmpty"`
			RampingVersionPercentageChangedTime time.Time `cli:",cardOmitEmpty"`
		}{
			Name:                                deploymentInfo.Name,
			CreateTime:                          deploymentInfo.CreateTime,
			LastModifierIdentity:                deploymentInfo.LastModifierIdentity,
			ManagerIdentity:                     deploymentInfo.ManagerIdentity,
			CurrentVersionDeploymentName:        curVerDepName,
			CurrentVersionBuildID:               curVerBuildId,
			RampingVersionDeploymentName:        rampVerDepName,
			RampingVersionBuildID:               rampVerBuildId,
			RampingVersionPercentage:            deploymentInfo.RoutingConfig.RampingVersionPercentage,
			CurrentVersionChangedTime:           deploymentInfo.RoutingConfig.CurrentVersionChangedTime,
			RampingVersionChangedTime:           deploymentInfo.RoutingConfig.RampingVersionChangedTime,
			RampingVersionPercentageChangedTime: deploymentInfo.RoutingConfig.RampingVersionPercentageChangedTime,
		}
		err := cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
		if err != nil {
			return fmt.Errorf("displaying worker deployment info failed: %w", err)
		}

		if len(deploymentInfo.VersionSummaries) > 0 {
			cctx.Printer.Println()
			cctx.Printer.Println(color.MagentaString("Version Summaries:"))
			err := cctx.Printer.PrintStructured(
				fDeploymentInfo.VersionSummaries,
				printer.StructuredOptions{Table: &printer.TableOptions{}},
			)
			if err != nil {
				return fmt.Errorf("displaying version summaries failed: %w", err)
			}
		}

		return nil
	}

	// json output
	return cctx.Printer.PrintStructured(fDeploymentInfo, printer.StructuredOptions{})
}

func formatDrainageInfo(drainageInfo *client.WorkerDeploymentVersionDrainageInfo) (formattedDrainageInfo, error) {
	if drainageInfo == nil {
		return formattedDrainageInfo{}, nil
	}

	drainageStr, err := drainageStatusToStr(drainageInfo.DrainageStatus)
	if err != nil {
		return formattedDrainageInfo{}, err
	}

	return formattedDrainageInfo{
		DrainageStatus:  drainageStr,
		LastChangedTime: drainageInfo.LastChangedTime,
		LastCheckedTime: drainageInfo.LastCheckedTime,
	}, nil
}

func formatTaskQueuesInfos(tqis []client.WorkerDeploymentTaskQueueInfo) ([]formattedTaskQueueInfoRowType, error) {
	var tqiRows []formattedTaskQueueInfoRowType
	for _, tqi := range tqis {
		tqTypeStr, err := taskQueueTypeToStr(tqi.Type)
		if err != nil {
			return tqiRows, err
		}
		tqiRows = append(tqiRows, formattedTaskQueueInfoRowType{
			Name: tqi.Name,
			Type: tqTypeStr,
		})
	}
	return tqiRows, nil
}

func workerDeploymentVersionInfoToRows(deploymentInfo client.WorkerDeploymentVersionInfo) (formattedWorkerDeploymentVersionInfoType, error) {
	tqi, err := formatTaskQueuesInfos(deploymentInfo.TaskQueuesInfos)
	if err != nil {
		return formattedWorkerDeploymentVersionInfoType{}, err
	}

	drainage, err := formatDrainageInfo(deploymentInfo.DrainageInfo)
	if err != nil {
		return formattedWorkerDeploymentVersionInfoType{}, err
	}

	return formattedWorkerDeploymentVersionInfoType{
		DeploymentName:     deploymentInfo.Version.DeploymentName,
		BuildID:            deploymentInfo.Version.BuildID,
		CreateTime:         deploymentInfo.CreateTime,
		RoutingChangedTime: deploymentInfo.RoutingChangedTime,
		CurrentSinceTime:   deploymentInfo.CurrentSinceTime,
		RampingSinceTime:   deploymentInfo.RampingSinceTime,
		RampPercentage:     deploymentInfo.RampPercentage,
		DrainageInfo:       drainage,
		TaskQueuesInfos:    tqi,
		Metadata:           deploymentInfo.Metadata,
	}, nil
}

func printWorkerDeploymentVersionInfo(cctx *CommandContext, deploymentInfo client.WorkerDeploymentVersionInfo, msg string) error {
	fDeploymentInfo, err := workerDeploymentVersionInfoToRows(deploymentInfo)
	if err != nil {
		return err
	}

	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString(msg))
		var drainageStr string
		var drainageLastChangedTime time.Time
		var drainageLastCheckedTime time.Time
		if deploymentInfo.DrainageInfo != nil {
			drainageStr, err = drainageStatusToStr(deploymentInfo.DrainageInfo.DrainageStatus)
			if err != nil {
				return err
			}
			drainageLastChangedTime = deploymentInfo.DrainageInfo.LastChangedTime
			drainageLastCheckedTime = deploymentInfo.DrainageInfo.LastCheckedTime
		}

		printMe := struct {
			DeploymentName          string
			BuildID                 string
			CreateTime              time.Time
			RoutingChangedTime      time.Time `cli:",cardOmitEmpty"`
			CurrentSinceTime        time.Time `cli:",cardOmitEmpty"`
			RampingSinceTime        time.Time `cli:",cardOmitEmpty"`
			RampPercentage          float32
			DrainageStatus          string                     `cli:",cardOmitEmpty"`
			DrainageLastChangedTime time.Time                  `cli:",cardOmitEmpty"`
			DrainageLastCheckedTime time.Time                  `cli:",cardOmitEmpty"`
			Metadata                map[string]*common.Payload `cli:",cardOmitEmpty"`
		}{
			DeploymentName:          deploymentInfo.Version.DeploymentName,
			BuildID:                 deploymentInfo.Version.BuildID,
			CreateTime:              deploymentInfo.CreateTime,
			RoutingChangedTime:      deploymentInfo.RoutingChangedTime,
			CurrentSinceTime:        deploymentInfo.CurrentSinceTime,
			RampingSinceTime:        deploymentInfo.RampingSinceTime,
			RampPercentage:          deploymentInfo.RampPercentage,
			DrainageStatus:          drainageStr,
			DrainageLastChangedTime: drainageLastChangedTime,
			DrainageLastCheckedTime: drainageLastCheckedTime,
			Metadata:                deploymentInfo.Metadata,
		}
		err := cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
		if err != nil {
			return fmt.Errorf("displaying worker deployment version info failed: %w", err)
		}

		if len(deploymentInfo.TaskQueuesInfos) > 0 {
			cctx.Printer.Println()
			cctx.Printer.Println(color.MagentaString("Task Queues:"))
			err := cctx.Printer.PrintStructured(
				fDeploymentInfo.TaskQueuesInfos,
				printer.StructuredOptions{Table: &printer.TableOptions{}},
			)
			if err != nil {
				return fmt.Errorf("displaying task queues failed: %w", err)
			}
		}

		return nil
	}

	// json output
	return cctx.Printer.PrintStructured(fDeploymentInfo, printer.StructuredOptions{})
}

type getDeploymentConflictTokenOptions struct {
	safeMode        bool
	safeModeMessage string
	deploymentName  string
}

func (c *TemporalWorkerDeploymentCommand) getConflictToken(cctx *CommandContext, options *getDeploymentConflictTokenOptions) ([]byte, error) {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return nil, err
	}
	defer cl.Close()

	dHandle := cl.WorkerDeploymentClient().GetHandle(options.deploymentName)

	resp, err := dHandle.Describe(cctx, client.WorkerDeploymentDescribeOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get deployment conflict token: %w", err)
	}

	if options.safeMode {
		// duplicate `cctx.promptYes` check to avoid printing deployment info with json
		if cctx.JSONOutput {
			return nil, fmt.Errorf("must bypass prompts when using JSON output")
		}
		err = printWorkerDeploymentInfo(cctx, resp.Info, "Worker Deployment Before Update:")
		if err != nil {
			return nil, fmt.Errorf("displaying deployment failed: %w", err)
		}

		yes, err := cctx.promptYes(
			fmt.Sprintf("Continue with set %v? y/N", options.safeModeMessage),
			false,
		)
		if err != nil {
			return nil, err
		} else if !yes {
			return nil, fmt.Errorf("user denied confirmation")
		}
	}

	return resp.ConflictToken, nil
}

func (c *TemporalWorkerDeploymentDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.Name)
	resp, err := dHandle.Describe(cctx, client.WorkerDeploymentDescribeOptions{})
	if err != nil {
		return fmt.Errorf("error describing worker deployment: %w", err)
	}
	err = printWorkerDeploymentInfo(cctx, resp.Info, "Worker Deployment:")
	if err != nil {
		return err
	}

	return nil
}

func (c *TemporalWorkerDeploymentDeleteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkerDeploymentClient().Delete(cctx, client.WorkerDeploymentDeleteOptions{
		Name:     c.Name,
		Identity: c.Parent.Parent.Identity,
	})
	if err != nil {
		return fmt.Errorf("error deleting worker deployment: %w", err)
	}

	cctx.Printer.Println("Successfully deleted worker deployment")
	return nil
}

func (c *TemporalWorkerDeploymentListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	res, err := cl.WorkerDeploymentClient().List(cctx, client.WorkerDeploymentListOptions{})
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
	page := make([]*formattedWorkerDeploymentListEntryType, 0, 100)

	for res.HasNext() {
		entry, err := res.Next()
		if err != nil {
			return err
		}
		rc, err := formatRoutingConfig(entry.RoutingConfig)
		if err != nil {
			return err
		}
		listEntry := formattedWorkerDeploymentInfoType{
			Name:          entry.Name,
			CreateTime:    entry.CreateTime,
			RoutingConfig: rc,
		}
		if cctx.JSONOutput {
			// For JSON dump one line of JSON per deployment
			_ = cctx.Printer.PrintStructured(listEntry, printer.StructuredOptions{})
		} else {
			// For non-JSON, we are doing a table for each page
			page = append(page, &formattedWorkerDeploymentListEntryType{
				Name:                         listEntry.Name,
				CreateTime:                   listEntry.CreateTime,
				CurrentVersionDeploymentName: listEntry.RoutingConfig.CurrentVersionDeploymentName,
				CurrentVersionBuildID:        listEntry.RoutingConfig.CurrentVersionBuildID,
				RampingVersionDeploymentName: listEntry.RoutingConfig.RampingVersionDeploymentName,
				RampingVersionBuildID:        listEntry.RoutingConfig.RampingVersionBuildID,
				RampingVersionPercentage:     listEntry.RoutingConfig.RampingVersionPercentage,
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

func (c *TemporalWorkerDeploymentManagerIdentitySetCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	token, err := c.Parent.Parent.getConflictToken(cctx, &getDeploymentConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "ManagerIdentity",
		deploymentName:  c.DeploymentName,
	})
	if err != nil {
		return err
	}

	newManagerIdentity := c.ManagerIdentity
	if c.Self {
		newManagerIdentity = c.Parent.Parent.Parent.Identity
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)
	resp, err := dHandle.SetManagerIdentity(cctx, client.WorkerDeploymentSetManagerIdentityOptions{
		Identity:        c.Parent.Parent.Parent.Identity,
		ConflictToken:   token,
		Self:            c.Self,
		ManagerIdentity: c.ManagerIdentity,
	})
	if err != nil {
		return fmt.Errorf("error setting the manager identity: %w", err)
	}

	cctx.Printer.Printlnf("Successfully set manager identity to '%s', was previously '%s'", newManagerIdentity, resp.PreviousManagerIdentity)
	return nil
}

func (c *TemporalWorkerDeploymentManagerIdentityUnsetCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	token, err := c.Parent.Parent.getConflictToken(cctx, &getDeploymentConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "ManagerIdentity",
		deploymentName:  c.DeploymentName,
	})
	if err != nil {
		return err
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)
	resp, err := dHandle.SetManagerIdentity(cctx, client.WorkerDeploymentSetManagerIdentityOptions{
		Identity:        c.Parent.Parent.Parent.Identity,
		ConflictToken:   token,
		ManagerIdentity: "",
	})
	if err != nil {
		return fmt.Errorf("error unsetting the manager identity: %w", err)
	}

	cctx.Printer.Printlnf("Successfully unset manager identity, was previously '%s'", resp.PreviousManagerIdentity)
	return nil
}

func (c *TemporalWorkerDeploymentDeleteVersionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)
	_, err = dHandle.DeleteVersion(cctx, client.WorkerDeploymentDeleteVersionOptions{
		BuildID:      c.BuildId,
		SkipDrainage: c.SkipDrainage,
		Identity:     c.Parent.Parent.Identity,
	})
	if err != nil {
		return fmt.Errorf("error deleting worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully deleted worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentDescribeVersionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)

	resp, err := dHandle.DescribeVersion(cctx, client.WorkerDeploymentDescribeVersionOptions{
		BuildID: c.BuildId,
	})
	if err != nil {
		return fmt.Errorf("error describing worker deployment version: %w", err)
	}

	err = printWorkerDeploymentVersionInfo(cctx, resp.Info, "Worker Deployment Version:")
	if err != nil {
		return err
	}

	return nil
}

func (c *TemporalWorkerDeploymentSetCurrentVersionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	token, err := c.Parent.getConflictToken(cctx, &getDeploymentConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "Current",
		deploymentName:  c.DeploymentName,
	})
	if err != nil && !(errors.As(err, new(*serviceerror.NotFound)) && c.AllowNoPollers) {
		return err
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)
	_, err = dHandle.SetCurrentVersion(cctx, client.WorkerDeploymentSetCurrentVersionOptions{
		BuildID:                 c.BuildId,
		Identity:                c.Parent.Parent.Identity,
		IgnoreMissingTaskQueues: c.IgnoreMissingTaskQueues,
		AllowNoPollers:          c.AllowNoPollers,
		ConflictToken:           token,
	})
	if err != nil {
		return fmt.Errorf("error setting the current worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully set the current worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentSetRampingVersionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	token, err := c.Parent.getConflictToken(cctx, &getDeploymentConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "Ramping",
		deploymentName:  c.DeploymentName,
	})
	if err != nil && !(errors.As(err, new(*serviceerror.NotFound)) && c.AllowNoPollers) {
		return err
	}

	percentage := c.Percentage
	if c.Delete {
		percentage = 0.0
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)
	_, err = dHandle.SetRampingVersion(cctx, client.WorkerDeploymentSetRampingVersionOptions{
		BuildID:                 c.BuildId,
		Percentage:              percentage,
		ConflictToken:           token,
		Identity:                c.Parent.Parent.Identity,
		IgnoreMissingTaskQueues: c.IgnoreMissingTaskQueues,
		AllowNoPollers:          c.AllowNoPollers,
	})
	if err != nil {
		return fmt.Errorf("error  setting the ramping worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully set the ramping worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentUpdateMetadataVersionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	metadata, err := stringKeysJSONValues(c.Metadata, false)
	if err != nil {
		return fmt.Errorf("invalid metadata values: %w", err)
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)
	response, err := dHandle.UpdateVersionMetadata(cctx, client.WorkerDeploymentUpdateVersionMetadataOptions{
		Version: worker.WorkerDeploymentVersion{
			BuildID:        c.BuildId,
			DeploymentName: c.DeploymentName,
		},
		MetadataUpdate: client.WorkerDeploymentMetadataUpdate{
			UpsertEntries: metadata,
			RemoveEntries: c.RemoveEntries,
		},
	})

	if err != nil {
		return err
	}

	cctx.Printer.Println(color.MagentaString("Metadata:"))
	printMe := struct {
		Metadata map[string]*common.Payload `cli:",cardOmitEmpty"`
	}{
		Metadata: response.Metadata,
	}

	err = cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
	if err != nil {
		return fmt.Errorf("displaying metadata failed: %w", err)
	}

	cctx.Printer.Println("Successfully updating metadata for worker deployment version")

	return nil
}
