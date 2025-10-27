package temporalcli

import (
	"fmt"
	"strings"
	"time"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	deploymentpb "go.temporal.io/api/deployment/v1"
	enumspb "go.temporal.io/api/enums/v1"
	workerpb "go.temporal.io/api/worker/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

type workerHeartbeatListRow struct {
	WorkerInstanceKey string    `json:"workerInstanceKey"`
	Status            string    `json:"status"`
	TaskQueue         string    `json:"taskQueue"`
	WorkerIdentity    string    `json:"workerIdentity"`
	HostName          string    `json:"hostName"`
	Deployment        string    `json:"deployment,omitempty" cli:",cardOmitEmpty"`
	HeartbeatTime     time.Time `json:"heartbeatTime"`
	Elapsed           string    `json:"elapsedSinceLastHeartbeat"`
}

type formattedWorkerDeploymentVersionRef struct {
	DeploymentName string `json:"deploymentName"`
	BuildId        string `json:"buildId"`
}

type formattedWorkerHostInfo struct {
	HostName            string  `json:"hostName"`
	ProcessId           string  `json:"processId"`
	ProcessKey          string  `json:"processKey"`
	CurrentHostCPUUsage float32 `json:"currentHostCpuUsage"`
	CurrentHostMemUsage float32 `json:"currentHostMemUsage"`
}

type formattedWorkerSlotsInfo struct {
	CurrentAvailableSlots      int32  `json:"currentAvailableSlots"`
	CurrentUsedSlots           int32  `json:"currentUsedSlots"`
	SlotSupplierKind           string `json:"slotSupplierKind"`
	TotalProcessedTasks        int32  `json:"totalProcessedTasks"`
	TotalFailedTasks           int32  `json:"totalFailedTasks"`
	LastIntervalProcessedTasks int32  `json:"lastIntervalProcessedTasks"`
	LastIntervalFailureTasks   int32  `json:"lastIntervalFailureTasks"`
}

type formattedWorkerPollerInfo struct {
	CurrentPollers         int32     `json:"currentPollers"`
	LastSuccessfulPollTime time.Time `json:"lastSuccessfulPollTime"`
	IsAutoscaling          bool      `json:"isAutoscaling"`
}

type formattedPluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type formattedWorkerHeartbeatDetail struct {
	WorkerInstanceKey          string                               `json:"workerInstanceKey"`
	WorkerIdentity             string                               `json:"workerIdentity"`
	Status                     string                               `json:"status"`
	TaskQueue                  string                               `json:"taskQueue"`
	DeploymentVersion          *formattedWorkerDeploymentVersionRef `json:"deploymentVersion,omitempty" cli:",cardOmitEmpty"`
	SdkName                    string                               `json:"sdkName"`
	SdkVersion                 string                               `json:"sdkVersion"`
	StartTime                  time.Time                            `json:"startTime"`
	HeartbeatTime              time.Time                            `json:"heartbeatTime"`
	ElapsedSinceLastHeartbeat  string                               `json:"elapsedSinceLastHeartbeat"`
	HostInfo                   *formattedWorkerHostInfo             `json:"hostInfo"`
	WorkflowTaskSlotsInfo      *formattedWorkerSlotsInfo            `json:"workflowTaskSlotsInfo,omitempty" cli:",cardOmitEmpty"`
	ActivityTaskSlotsInfo      *formattedWorkerSlotsInfo            `json:"activityTaskSlotsInfo,omitempty" cli:",cardOmitEmpty"`
	NexusTaskSlotsInfo         *formattedWorkerSlotsInfo            `json:"nexusTaskSlotsInfo,omitempty" cli:",cardOmitEmpty"`
	LocalActivityTaskSlotsInfo *formattedWorkerSlotsInfo            `json:"localActivityTaskSlotsInfo,omitempty" cli:",cardOmitEmpty"`
	WorkflowPollerInfo         *formattedWorkerPollerInfo           `json:"workflowPollerInfo,omitempty" cli:",cardOmitEmpty"`
	WorkflowStickyPollerInfo   *formattedWorkerPollerInfo           `json:"workflowStickyPollerInfo,omitempty" cli:",cardOmitEmpty"`
	ActivityPollerInfo         *formattedWorkerPollerInfo           `json:"activityPollerInfo,omitempty" cli:",cardOmitEmpty"`
	NexusPollerInfo            *formattedWorkerPollerInfo           `json:"nexusPollerInfo,omitempty" cli:",cardOmitEmpty"`
	TotalStickyCacheHit        int32                                `json:"totalStickyCacheHit"`
	TotalStickyCacheMiss       int32                                `json:"totalStickyCacheMiss"`
	CurrentStickyCacheSize     int32                                `json:"currentStickyCacheSize"`
	Plugins                    []formattedPluginInfo                `json:"plugins,omitempty" cli:",cardOmitEmpty"`
}

func (c *TemporalWorkerDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().DescribeWorker(cctx, &workflowservice.DescribeWorkerRequest{
		Namespace:         c.Parent.Namespace,
		WorkerInstanceKey: c.WorkerInstanceKey,
	})
	if err != nil {
		return err
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	info := resp.GetWorkerInfo()
	if info == nil {
		return fmt.Errorf("worker info not found in response")
	}

	hb := info.GetWorkerHeartbeat()
	if hb == nil {
		return fmt.Errorf("worker heartbeat not found in response")
	}

	formatted := formatWorkerHeartbeatDetail(hb)
	return cctx.Printer.PrintStructured(formatted, printer.StructuredOptions{})
}

func (c *TemporalWorkerListCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	svc := cl.WorkflowService()

	limit := c.Limit

	cctx.Printer.StartList()
	defer cctx.Printer.EndList()

	printOpts := printer.StructuredOptions{Table: &printer.TableOptions{}}
	page := make([]*workerHeartbeatListRow, 0, 100)
	printed := 0
	var token []byte

	for {
		req := &workflowservice.ListWorkersRequest{
			Namespace:     c.Parent.Namespace,
			NextPageToken: token,
			Query:         c.Query,
		}

		resp, err := svc.ListWorkers(cctx, req)
		if err != nil {
			return err
		}

		workers := resp.GetWorkersInfo()
		if cctx.JSONOutput {
			for _, info := range workers {
				if limit > 0 && printed >= limit {
					break
				}
				if info == nil {
					continue
				}
				if err := cctx.Printer.PrintStructured(info, printer.StructuredOptions{}); err != nil {
					return err
				}
				printed++
			}
		} else {
			for _, info := range workers {
				if limit > 0 && printed >= limit {
					break
				}
				if info == nil {
					continue
				}
				hb := info.GetWorkerHeartbeat()
				if hb == nil {
					continue
				}
				row := formatWorkerHeartbeatListRow(hb)
				page = append(page, &row)
				printed++
				if len(page) == cap(page) {
					if err := cctx.Printer.PrintStructured(page, printOpts); err != nil {
						return err
					}
					page = page[:0]
					printOpts.Table.NoHeader = true
				}
			}
		}

		if limit > 0 && printed >= limit {
			break
		}

		token = resp.GetNextPageToken()
		if len(token) == 0 {
			break
		}
	}

	if !cctx.JSONOutput {
		if err := cctx.Printer.PrintStructured(page, printOpts); err != nil {
			return err
		}
	}

	return nil
}

func formatWorkerHeartbeatListRow(hb *workerpb.WorkerHeartbeat) workerHeartbeatListRow {
	if hb == nil {
		return workerHeartbeatListRow{}
	}

	row := workerHeartbeatListRow{
		WorkerInstanceKey: hb.GetWorkerInstanceKey(),
		Status:            workerStatusToString(hb.GetStatus()),
		TaskQueue:         hb.GetTaskQueue(),
		WorkerIdentity:    hb.GetWorkerIdentity(),
		HeartbeatTime:     timestampToTime(hb.GetHeartbeatTime()),
		Elapsed:           durationToString(hb.GetElapsedSinceLastHeartbeat()),
	}

	if host := hb.GetHostInfo(); host != nil {
		row.HostName = host.GetHostName()
	}
	if dv := hb.GetDeploymentVersion(); dv != nil {
		row.Deployment = formatDeploymentVersion(dv)
	}

	return row
}

func formatWorkerHeartbeatDetail(hb *workerpb.WorkerHeartbeat) formattedWorkerHeartbeatDetail {
	if hb == nil {
		return formattedWorkerHeartbeatDetail{}
	}

	detail := formattedWorkerHeartbeatDetail{
		WorkerInstanceKey:          hb.GetWorkerInstanceKey(),
		WorkerIdentity:             hb.GetWorkerIdentity(),
		Status:                     workerStatusToString(hb.GetStatus()),
		TaskQueue:                  hb.GetTaskQueue(),
		SdkName:                    hb.GetSdkName(),
		SdkVersion:                 hb.GetSdkVersion(),
		StartTime:                  timestampToTime(hb.GetStartTime()),
		HeartbeatTime:              timestampToTime(hb.GetHeartbeatTime()),
		ElapsedSinceLastHeartbeat:  durationToString(hb.GetElapsedSinceLastHeartbeat()),
		HostInfo:                   formatWorkerHostInfo(hb.GetHostInfo()),
		WorkflowTaskSlotsInfo:      formatWorkerSlots(hb.GetWorkflowTaskSlotsInfo()),
		ActivityTaskSlotsInfo:      formatWorkerSlots(hb.GetActivityTaskSlotsInfo()),
		NexusTaskSlotsInfo:         formatWorkerSlots(hb.GetNexusTaskSlotsInfo()),
		LocalActivityTaskSlotsInfo: formatWorkerSlots(hb.GetLocalActivitySlotsInfo()),
		WorkflowPollerInfo:         formatWorkerPoller(hb.GetWorkflowPollerInfo()),
		WorkflowStickyPollerInfo:   formatWorkerPoller(hb.GetWorkflowStickyPollerInfo()),
		ActivityPollerInfo:         formatWorkerPoller(hb.GetActivityPollerInfo()),
		NexusPollerInfo:            formatWorkerPoller(hb.GetNexusPollerInfo()),
		TotalStickyCacheHit:        hb.GetTotalStickyCacheHit(),
		TotalStickyCacheMiss:       hb.GetTotalStickyCacheMiss(),
		CurrentStickyCacheSize:     hb.GetCurrentStickyCacheSize(),
		Plugins:                    formatPlugins(hb.GetPlugins()),
	}

	if dv := hb.GetDeploymentVersion(); dv != nil {
		if dv.GetDeploymentName() != "" || dv.GetBuildId() != "" {
			detail.DeploymentVersion = &formattedWorkerDeploymentVersionRef{
				DeploymentName: dv.GetDeploymentName(),
				BuildId:        dv.GetBuildId(),
			}
		}
	}

	return detail
}

func workerStatusToString(status enumspb.WorkerStatus) string {
	statusStr := status.String()
	statusStr = strings.TrimPrefix(statusStr, "WORKER_STATUS_")
	if statusStr == "" {
		return "unspecified"
	}
	return statusStr
}

func formatDeploymentVersion(dv *deploymentpb.WorkerDeploymentVersion) string {
	if dv == nil {
		return ""
	}
	depName := dv.GetDeploymentName()
	buildID := dv.GetBuildId()
	switch {
	case depName != "" && buildID != "":
		return depName + "@" + buildID
	case depName != "":
		return depName
	case buildID != "":
		return buildID
	default:
		return ""
	}
}

func formatWorkerHostInfo(info *workerpb.WorkerHostInfo) *formattedWorkerHostInfo {
	if info == nil {
		return nil
	}
	formatted := &formattedWorkerHostInfo{
		HostName:            info.GetHostName(),
		ProcessId:           info.GetProcessId(),
		ProcessKey:          info.GetProcessKey(),
		CurrentHostCPUUsage: info.GetCurrentHostCpuUsage(),
		CurrentHostMemUsage: info.GetCurrentHostMemUsage(),
	}
	if formatted.HostName == "" && formatted.ProcessId == "" && formatted.ProcessKey == "" &&
		formatted.CurrentHostCPUUsage == 0 && formatted.CurrentHostMemUsage == 0 {
		return nil
	}
	return formatted
}

func formatWorkerSlots(info *workerpb.WorkerSlotsInfo) *formattedWorkerSlotsInfo {
	if info == nil {
		return nil
	}
	formatted := &formattedWorkerSlotsInfo{
		CurrentAvailableSlots:      info.GetCurrentAvailableSlots(),
		CurrentUsedSlots:           info.GetCurrentUsedSlots(),
		SlotSupplierKind:           info.GetSlotSupplierKind(),
		TotalProcessedTasks:        info.GetTotalProcessedTasks(),
		TotalFailedTasks:           info.GetTotalFailedTasks(),
		LastIntervalProcessedTasks: info.GetLastIntervalProcessedTasks(),
		LastIntervalFailureTasks:   info.GetLastIntervalFailureTasks(),
	}
	if formatted.CurrentAvailableSlots == 0 && formatted.CurrentUsedSlots == 0 && formatted.SlotSupplierKind == "Fixed" &&
		formatted.TotalProcessedTasks == 0 && formatted.TotalFailedTasks == 0 &&
		formatted.LastIntervalProcessedTasks == 0 && formatted.LastIntervalFailureTasks == 0 {
		return nil
	}
	return formatted
}

func formatWorkerPoller(info *workerpb.WorkerPollerInfo) *formattedWorkerPollerInfo {
	if info == nil {
		return nil
	}
	formatted := &formattedWorkerPollerInfo{
		CurrentPollers:         info.GetCurrentPollers(),
		LastSuccessfulPollTime: timestampToTime(info.GetLastSuccessfulPollTime()),
		IsAutoscaling:          info.GetIsAutoscaling(),
	}
	if formatted.CurrentPollers == 0 && formatted.LastSuccessfulPollTime.IsZero() && !formatted.IsAutoscaling {
		return nil
	}
	return formatted
}

func formatPlugins(plugins []*workerpb.PluginInfo) []formattedPluginInfo {
	if len(plugins) == 0 {
		return nil
	}
	formatted := make([]formattedPluginInfo, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil {
			continue
		}
		formatted = append(formatted, formattedPluginInfo{
			Name:    plugin.GetName(),
			Version: plugin.GetVersion(),
		})
	}
	if len(formatted) == 0 {
		return nil
	}
	return formatted
}

func durationToString(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}
