package temporalcli

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/client"
)

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
	Version            string                          `json:"version"`
	CreateTime         time.Time                       `json:"createTime"`
	RoutingChangedTime time.Time                       `json:"routingChangedTime"`
	CurrentSinceTime   time.Time                       `json:"currentSinceTime"`
	RampingSinceTime   time.Time                       `json:"rampingSinceTime"`
	RampPercentage     float32                         `json:"rampPercentage"`
	DrainageInfo       formattedDrainageInfo           `json:"drainageInfo,omitempty"`
	TaskQueuesInfos    []formattedTaskQueueInfoRowType `json:"taskQueuesInfos,omitempty"`
	Metadata           map[string]*common.Payload      `json:"metadata,omitempty"`
}

func extractDeploymentName(version string, deploymentName string, failNonQualified bool) (string, error) {
	if version == "" || version == "__unversioned__" {
		if failNonQualified {
			return "", fmt.Errorf(
				"invalid deployment version type for this operation, use a fully-qualified version",
			)
		}
		if deploymentName == "" {
			return "", fmt.Errorf(
				"specify the deployment name with `--deployment-name` with a non-fully-qualified version",
			)
		}
		return deploymentName, nil
	}
	splitVersion := strings.SplitN(version, ".", 2)
	if len(splitVersion) != 2 {
		return "", fmt.Errorf(
			"invalid format for worker deployment version %v, not YourDeploymentName.YourBuildID",
			version,
		)
	}
	return splitVersion[0], nil
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
		Version:            deploymentInfo.Version,
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
			Version                 string
			CreateTime              time.Time
			RoutingChangedTime      time.Time `cli:",cardOmitEmpty"`
			CurrentSinceTime        time.Time `cli:",cardOmitEmpty"`
			RampingSinceTime        time.Time `cli:",cardOmitEmpty"`
			RampPercentage          float32
			DrainageStatus          string    `cli:",cardOmitEmpty"`
			DrainageLastChangedTime time.Time `cli:",cardOmitEmpty"`
			DrainageLastCheckedTime time.Time `cli:",cardOmitEmpty"`
		}{
			Version:                 deploymentInfo.Version,
			CreateTime:              deploymentInfo.CreateTime,
			RoutingChangedTime:      deploymentInfo.RoutingChangedTime,
			CurrentSinceTime:        deploymentInfo.CurrentSinceTime,
			RampingSinceTime:        deploymentInfo.RampingSinceTime,
			RampPercentage:          deploymentInfo.RampPercentage,
			DrainageStatus:          drainageStr,
			DrainageLastChangedTime: drainageLastChangedTime,
			DrainageLastCheckedTime: drainageLastCheckedTime,
		}
		err := cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
		if err != nil {
			return fmt.Errorf("displaying worker deployment version info failed: %w", err)
		}

		if len(deploymentInfo.TaskQueuesInfos) > 0 {
			cctx.Printer.Println()
			cctx.Printer.Println(color.MagentaString("Task Queues:"))
			err := cctx.Printer.PrintStructured(
				deploymentInfo.TaskQueuesInfos,
				printer.StructuredOptions{Table: &printer.TableOptions{}},
			)
			if err != nil {
				return fmt.Errorf("displaying task queues failed: %w", err)
			}
		}

		// TODO(antlai-temporal): Print metadata

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

func (c *TemporalWorkerDeploymentVersionCommand) getConflictToken(cctx *CommandContext, options *getDeploymentConflictTokenOptions) ([]byte, error) {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
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
		// duplicate `cctx.promptYes` check to avoid printing current rules with json
		if cctx.JSONOutput {
			return nil, fmt.Errorf("must bypass prompts when using JSON output")
		}
		err = printWorkerDeploymentInfo(cctx, resp.Info, "Worker Deployment Before Update:")
		if err != nil {
			return nil, fmt.Errorf("displaying deployment failed: %w", err)
		}

		yes, err := cctx.promptYes(
			fmt.Sprintf("Continue with set %v? y/N", options.safeModeMessage), false)
		if err != nil {
			return nil, err
		} else if !yes {
			return nil, fmt.Errorf("user denied confirmation")
		}
	}

	return resp.ConflictToken, nil
}

func (c *TemporalWorkerDeploymentVersionDeleteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	name, err := extractDeploymentName(c.Version, "", true)
	if err != nil {
		return err
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(name)
	_, err = dHandle.DeleteVersion(cctx, client.WorkerDeploymentDeleteVersionOptions{
		Version:      c.Version,
		SkipDrainage: c.SkipDrainage,
		Identity:     c.Identity,
	})
	if err != nil {
		return fmt.Errorf("error deleting worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully deleted worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentVersionDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	name, err := extractDeploymentName(c.Version, "", true)
	if err != nil {
		return err
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(name)

	resp, err := dHandle.DescribeVersion(cctx, client.WorkerDeploymentDescribeVersionOptions{
		Version: c.Version,
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

func (c *TemporalWorkerDeploymentVersionSetCurrentCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.Parent.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	name, err := extractDeploymentName(c.Version, c.DeploymentName, false)
	if err != nil {
		return err
	}

	token, err := c.Parent.getConflictToken(cctx, &getDeploymentConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "Current",
		deploymentName:  name,
	})
	if err != nil {
		return err
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(name)
	_, err = dHandle.SetCurrentVersion(cctx, client.WorkerDeploymentSetCurrentVersionOptions{
		Version:                 c.Version,
		Identity:                c.Identity,
		IgnoreMissingTaskQueues: c.IgnoreMissingTaskQueues,
		ConflictToken:           token,
	})
	if err != nil {
		return fmt.Errorf("error setting the current worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully setting the current worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentVersionSetRampingCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	name, err := extractDeploymentName(c.Version, c.DeploymentName, false)
	if err != nil {
		return err
	}

	token, err := c.Parent.getConflictToken(cctx, &getDeploymentConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "Ramping",
		deploymentName:  name,
	})
	if err != nil {
		return err
	}

	dHandle := cl.WorkerDeploymentClient().GetHandle(name)
	_, err = dHandle.SetRampingVersion(cctx, client.WorkerDeploymentSetRampingVersionOptions{
		Version:                 c.Version,
		Percentage:              c.Percentage,
		ConflictToken:           token,
		Identity:                c.Identity,
		IgnoreMissingTaskQueues: c.IgnoreMissingTaskQueues,
	})
	if err != nil {
		return fmt.Errorf("error  setting the ramping worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully setting the ramping worker deployment version")
	return nil
}
