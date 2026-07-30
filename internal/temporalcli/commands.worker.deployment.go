package temporalcli

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/api/common/v1"
	commonpb "go.temporal.io/api/common/v1"
	computepb "go.temporal.io/api/compute/v1"
	"go.temporal.io/api/deployment/v1"
	deploymentpb "go.temporal.io/api/deployment/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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
	Name               string                                 `json:"name"`
	Type               string                                 `json:"type"`
	Stats              *formattedVersionStatsRowType          `json:"stats,omitempty"`
	StatsByPriorityKey map[int32]formattedVersionStatsRowType `json:"statsByPriorityKey,omitempty"`
}

type formattedVersionStatsRowType struct {
	ApproximateBacklogCount int64         `json:"approximateBacklogCount"`
	ApproximateBacklogAge   time.Duration `json:"approximateBacklogAge"`
	BacklogIncreaseRate     float32       `json:"backlogIncreaseRate"`
	TasksAddRate            float32       `json:"tasksAddRate"`
	TasksDispatchRate       float32       `json:"tasksDispatchRate"`
}

// Text display types for task queue info (flattened stats)
type taskQueueDisplayRowBasic struct {
	Name string
	Type string
}

type taskQueueDisplayRowWithStats struct {
	Name                    string
	Type                    string
	ApproximateBacklogCount int64   `cli:",align=right"`
	ApproximateBacklogAge   string  `cli:",align=right"`
	BacklogIncreaseRate     float32 `cli:",align=right"`
	TasksAddRate            float32 `cli:",align=right"`
	TasksDispatchRate       float32 `cli:",align=right"`
}

type priorityStatsDisplayRow struct {
	Priority                int32   `cli:",align=right"`
	ApproximateBacklogCount int64   `cli:",align=right"`
	ApproximateBacklogAge   string  `cli:",align=right"`
	BacklogIncreaseRate     float32 `cli:",align=right"`
	TasksAddRate            float32 `cli:",align=right"`
	TasksDispatchRate       float32 `cli:",align=right"`
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
	ComputeConfig      *formattedComputeConfig         `json:"computeConfig,omitempty"`
}

type formattedComputeConfig struct {
	ScalingGroups map[string]formattedComputeConfigScalingGroup `json:"scalingGroups"`
}

type formattedComputeConfigScalingGroup struct {
	Provider *formattedComputeConfigProvider `json:"provider,omitempty"`
	Scaler   *formattedComputeConfigScaler   `json:"scaler,omitempty"`
}

type formattedComputeConfigProvider struct {
	Type string `json:"type"`
}

type formattedComputeConfigScaler struct {
	Type              string   `json:"type"`
	MinInstances      *int64   `json:"minInstances,omitempty"`
	MaxInstances      *int64   `json:"maxInstances,omitempty"`
	InitialInstances  *int64   `json:"initialInstances,omitempty"`
	UtilizationTarget *float64 `json:"utilizationTarget,omitempty"`
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

// Proto-based conversion functions for describe-version command
// These functions convert gRPC proto types directly, avoiding SDK type dependencies.

func drainageStatusProtoToStr(status enumspb.VersionDrainageStatus) (string, error) {
	switch status {
	case enumspb.VERSION_DRAINAGE_STATUS_UNSPECIFIED:
		return "unspecified", nil
	case enumspb.VERSION_DRAINAGE_STATUS_DRAINING:
		return "draining", nil
	case enumspb.VERSION_DRAINAGE_STATUS_DRAINED:
		return "drained", nil
	default:
		return "", fmt.Errorf("unrecognized drainage status: %d", status)
	}
}

func taskQueueTypeProtoToStr(taskQueueType enumspb.TaskQueueType) (string, error) {
	switch taskQueueType {
	case enumspb.TASK_QUEUE_TYPE_UNSPECIFIED:
		return "unspecified", nil
	case enumspb.TASK_QUEUE_TYPE_WORKFLOW:
		return "workflow", nil
	case enumspb.TASK_QUEUE_TYPE_ACTIVITY:
		return "activity", nil
	case enumspb.TASK_QUEUE_TYPE_NEXUS:
		return "nexus", nil
	default:
		return "", fmt.Errorf("unrecognized task queue type: %d", taskQueueType)
	}
}

func formatVersionStatsProto(tqStats *taskqueuepb.TaskQueueStats) formattedVersionStatsRowType {
	if tqStats == nil {
		return formattedVersionStatsRowType{}
	}
	return formattedVersionStatsRowType{
		ApproximateBacklogCount: tqStats.ApproximateBacklogCount,
		ApproximateBacklogAge:   tqStats.ApproximateBacklogAge.AsDuration(),
		// BacklogIncreaseRate is computed as TasksAddRate - TasksDispatchRate (same as SDK)
		BacklogIncreaseRate: tqStats.TasksAddRate - tqStats.TasksDispatchRate,
		TasksAddRate:        tqStats.TasksAddRate,
		TasksDispatchRate:   tqStats.TasksDispatchRate,
	}
}

func formatTaskQueuesInfosProto(tqis []*workflowservice.DescribeWorkerDeploymentVersionResponse_VersionTaskQueue, includeStats bool) ([]formattedTaskQueueInfoRowType, error) {
	var tqiRows []formattedTaskQueueInfoRowType
	for _, tqi := range tqis {
		tqTypeStr, err := taskQueueTypeProtoToStr(tqi.GetType())
		if err != nil {
			return tqiRows, err
		}

		row := formattedTaskQueueInfoRowType{
			Name: tqi.GetName(),
			Type: tqTypeStr,
		}

		if includeStats {
			fVersionStats := formatVersionStatsProto(tqi.GetStats())
			row.Stats = &fVersionStats

			if len(tqi.GetStatsByPriorityKey()) > 0 {
				fVersionStatsByPriorityKey := map[int32]formattedVersionStatsRowType{}
				for k, v := range tqi.GetStatsByPriorityKey() {
					fVersionStatsByPriorityKey[k] = formatVersionStatsProto(v)
				}
				row.StatsByPriorityKey = fVersionStatsByPriorityKey
			}
		}

		tqiRows = append(tqiRows, row)
	}
	return tqiRows, nil
}

func formatDrainageInfoProto(drainageInfo *deploymentpb.VersionDrainageInfo) (formattedDrainageInfo, error) {
	if drainageInfo == nil {
		return formattedDrainageInfo{}, nil
	}

	drainageStr, err := drainageStatusProtoToStr(drainageInfo.GetStatus())
	if err != nil {
		return formattedDrainageInfo{}, err
	}

	return formattedDrainageInfo{
		DrainageStatus:  drainageStr,
		LastChangedTime: drainageInfo.GetLastChangedTime().AsTime(),
		LastCheckedTime: drainageInfo.GetLastCheckedTime().AsTime(),
	}, nil
}

// Config keys for the WCI rate-based scaler, shared by the build path
// (gcpCloudRunScalerDetails) and the read path (decodeScalerSettings) so the two
// never drift. Mirrors:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/scaling_algorithm/rate_based.go
const (
	scalerKeyMinCount          = "min_count"
	scalerKeyMaxCount          = "max_count"
	scalerKeyInitialCount      = "initial_count"
	scalerKeyUtilizationTarget = "utilization_target"
)

// scalerCountFromMap reads an integer worker-count value from a decoded scaler
// details map. The default data converter round-trips JSON numbers as float64,
// so that case is handled alongside the native int types.
func scalerCountFromMap(m map[string]any, key string) (int64, bool) {
	switch n := m[key].(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	default:
		return 0, false
	}
}

// scalerFloatFromMap reads a fractional value (e.g. utilization_target) from a
// decoded scaler details map. JSON numbers decode as float64.
func scalerFloatFromMap(m map[string]any, key string) (float64, bool) {
	switch n := m[key].(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

// scalerSettings holds the rate-based scaler settings surfaced for display. Each
// field is nil when the corresponding key is absent from the scaler details.
type scalerSettings struct {
	minInstances      *int64
	maxInstances      *int64
	initialInstances  *int64
	utilizationTarget *float64
}

// decodeScalerSettings extracts the rate-based scaler settings from a
// ComputeScaler's details payload for display. Best-effort: returns zero-value
// (all-nil) settings when the scaler, its details, or a key is absent, or when
// the payload cannot be decoded, so read paths never fail on an unexpected shape.
func decodeScalerSettings(s *computepb.ComputeScaler) scalerSettings {
	var out scalerSettings
	details := s.GetDetails()
	if details == nil {
		return out
	}
	var m map[string]any
	if err := converter.GetDefaultDataConverter().FromPayload(details, &m); err != nil {
		return out
	}
	if v, ok := scalerCountFromMap(m, scalerKeyMinCount); ok {
		out.minInstances = &v
	}
	if v, ok := scalerCountFromMap(m, scalerKeyMaxCount); ok {
		out.maxInstances = &v
	}
	if v, ok := scalerCountFromMap(m, scalerKeyInitialCount); ok {
		out.initialInstances = &v
	}
	if v, ok := scalerFloatFromMap(m, scalerKeyUtilizationTarget); ok {
		out.utilizationTarget = &v
	}
	return out
}

func formatComputeConfigProto(cc *computepb.ComputeConfig) *formattedComputeConfig {
	if cc == nil {
		return nil
	}
	msgSGs := cc.GetScalingGroups()
	if len(msgSGs) == 0 {
		return nil
	}
	sgs := make(map[string]formattedComputeConfigScalingGroup, len(msgSGs))
	for name, msgSG := range msgSGs {
		p := msgSG.GetProvider()
		s := msgSG.GetScaler()
		if p == nil && s == nil {
			continue
		}
		sg := formattedComputeConfigScalingGroup{}
		if p != nil {
			sg.Provider = &formattedComputeConfigProvider{
				Type: p.GetType(),
			}
		}
		if s != nil {
			fs := &formattedComputeConfigScaler{Type: s.GetType()}
			set := decodeScalerSettings(s)
			fs.MinInstances = set.minInstances
			fs.MaxInstances = set.maxInstances
			fs.InitialInstances = set.initialInstances
			fs.UtilizationTarget = set.utilizationTarget
			sg.Scaler = fs
		}
		sgs[name] = sg
	}
	return &formattedComputeConfig{
		ScalingGroups: sgs,
	}
}

// workerDeploymentVersionInfoProtoToRows converts gRPC proto types to formatted types for display.
func workerDeploymentVersionInfoProtoToRows(deploymentInfo *deploymentpb.WorkerDeploymentVersionInfo, taskQueueInfos []*workflowservice.DescribeWorkerDeploymentVersionResponse_VersionTaskQueue, includeStats bool) (formattedWorkerDeploymentVersionInfoType, error) {
	tqi, err := formatTaskQueuesInfosProto(taskQueueInfos, includeStats)
	if err != nil {
		return formattedWorkerDeploymentVersionInfoType{}, err
	}

	drainage, err := formatDrainageInfoProto(deploymentInfo.GetDrainageInfo())
	if err != nil {
		return formattedWorkerDeploymentVersionInfoType{}, err
	}

	computeConfig := formatComputeConfigProto(deploymentInfo.GetComputeConfig())

	return formattedWorkerDeploymentVersionInfoType{
		DeploymentName:     deploymentInfo.GetDeploymentVersion().GetDeploymentName(),
		BuildID:            deploymentInfo.GetDeploymentVersion().GetBuildId(),
		CreateTime:         deploymentInfo.GetCreateTime().AsTime(),
		RoutingChangedTime: deploymentInfo.GetRoutingChangedTime().AsTime(),
		CurrentSinceTime:   deploymentInfo.GetCurrentSinceTime().AsTime(),
		RampingSinceTime:   deploymentInfo.GetRampingSinceTime().AsTime(),
		RampPercentage:     deploymentInfo.GetRampPercentage(),
		DrainageInfo:       drainage,
		TaskQueuesInfos:    tqi,
		Metadata:           deploymentInfo.GetMetadata().GetEntries(),
		ComputeConfig:      computeConfig,
	}, nil
}

func computeConfigSummaryStr(cc *computepb.ComputeConfig) string {
	if cc == nil {
		return ""
	}
	summaries := []string{}
	for _, sg := range cc.GetScalingGroups() {
		p := sg.GetProvider()
		if p == nil {
			continue
		}
		summary := p.GetType()
		// Append whichever scaler settings are present so the one-line summary
		// reflects the configured limits (ordered min, initial, max, utilization).
		set := decodeScalerSettings(sg.GetScaler())
		parts := []string{}
		if set.minInstances != nil {
			parts = append(parts, fmt.Sprintf("min %d", *set.minInstances))
		}
		if set.initialInstances != nil {
			parts = append(parts, fmt.Sprintf("initial %d", *set.initialInstances))
		}
		if set.maxInstances != nil {
			parts = append(parts, fmt.Sprintf("max %d", *set.maxInstances))
		}
		if set.utilizationTarget != nil {
			parts = append(parts, fmt.Sprintf("utilization %g", *set.utilizationTarget))
		}
		if len(parts) > 0 {
			summary = fmt.Sprintf("%s (%s)", summary, strings.Join(parts, ", "))
		}
		if !slices.Contains(summaries, summary) {
			summaries = append(summaries, summary)
		}
	}
	return strings.Join(summaries, ",")
}

// printWorkerDeploymentVersionInfoProto prints worker deployment version info from proto types.
func printWorkerDeploymentVersionInfoProto(cctx *CommandContext, deploymentInfo *deploymentpb.WorkerDeploymentVersionInfo, taskQueueInfos []*workflowservice.DescribeWorkerDeploymentVersionResponse_VersionTaskQueue, msg string, opts printVersionInfoOptions) error {
	fDeploymentInfo, err := workerDeploymentVersionInfoProtoToRows(deploymentInfo, taskQueueInfos, opts.showStats)
	if err != nil {
		return err
	}

	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString(msg))
		var drainageStr string
		var drainageLastChangedTime time.Time
		var drainageLastCheckedTime time.Time
		if deploymentInfo.GetDrainageInfo() != nil {
			drainageStr, err = drainageStatusProtoToStr(deploymentInfo.GetDrainageInfo().GetStatus())
			if err != nil {
				return err
			}
			drainageLastChangedTime = deploymentInfo.GetDrainageInfo().GetLastChangedTime().AsTime()
			drainageLastCheckedTime = deploymentInfo.GetDrainageInfo().GetLastCheckedTime().AsTime()
		}
		computeConfigSummary := computeConfigSummaryStr(deploymentInfo.GetComputeConfig())

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
			ComputeConfigSummary    string                     `cli:",cardOmitEmpty"`
		}{
			DeploymentName:          deploymentInfo.GetDeploymentVersion().GetDeploymentName(),
			BuildID:                 deploymentInfo.GetDeploymentVersion().GetBuildId(),
			CreateTime:              deploymentInfo.GetCreateTime().AsTime(),
			RoutingChangedTime:      deploymentInfo.GetRoutingChangedTime().AsTime(),
			CurrentSinceTime:        deploymentInfo.GetCurrentSinceTime().AsTime(),
			RampingSinceTime:        deploymentInfo.GetRampingSinceTime().AsTime(),
			RampPercentage:          deploymentInfo.GetRampPercentage(),
			DrainageStatus:          drainageStr,
			DrainageLastChangedTime: drainageLastChangedTime,
			DrainageLastCheckedTime: drainageLastCheckedTime,
			Metadata:                deploymentInfo.GetMetadata().GetEntries(),
			ComputeConfigSummary:    computeConfigSummary,
		}
		err := cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
		if err != nil {
			return fmt.Errorf("displaying worker deployment version info failed: %w", err)
		}

		if len(taskQueueInfos) > 0 {
			if err := printTaskQueuesInfo(cctx, fDeploymentInfo.TaskQueuesInfos, opts); err != nil {
				return err
			}
		}

		return nil
	}

	// json output
	return cctx.Printer.PrintStructured(fDeploymentInfo, printer.StructuredOptions{})
}

type printVersionInfoOptions struct {
	showStats bool
}

func formatDurationShort(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	return d.Truncate(time.Millisecond).String()
}

func printTaskQueuesInfo(cctx *CommandContext, taskQueues []formattedTaskQueueInfoRowType, opts printVersionInfoOptions) error {
	cctx.Printer.Println()
	cctx.Printer.Println(color.MagentaString("Task Queues:"))

	if opts.showStats {
		// Show flattened stats in the table
		rows := make([]taskQueueDisplayRowWithStats, 0, len(taskQueues))
		for _, tq := range taskQueues {
			row := taskQueueDisplayRowWithStats{
				Name: tq.Name,
				Type: tq.Type,
			}
			if tq.Stats != nil {
				row.ApproximateBacklogCount = tq.Stats.ApproximateBacklogCount
				row.ApproximateBacklogAge = formatDurationShort(tq.Stats.ApproximateBacklogAge)
				row.BacklogIncreaseRate = tq.Stats.BacklogIncreaseRate
				row.TasksAddRate = tq.Stats.TasksAddRate
				row.TasksDispatchRate = tq.Stats.TasksDispatchRate
			}
			rows = append(rows, row)
		}
		if err := cctx.Printer.PrintStructured(rows, printer.StructuredOptions{Table: &printer.TableOptions{}}); err != nil {
			return fmt.Errorf("displaying task queues failed: %w", err)
		}

		// Show per-priority stats automatically if any task queue has non-default priority data.
		// Skip if the only priority key is 3 (the default), as that would be redundant.
		for _, tq := range taskQueues {
			if !hasNonDefaultPriorityKeys(tq.StatsByPriorityKey) {
				continue
			}
			cctx.Printer.Println()
			cctx.Printer.Println(color.MagentaString(fmt.Sprintf("Stats by Priority (%s / %s):", tq.Name, tq.Type)))

			// Sort priority keys for consistent output
			priorities := make([]int32, 0, len(tq.StatsByPriorityKey))
			for p := range tq.StatsByPriorityKey {
				priorities = append(priorities, p)
			}
			sortInt32s(priorities)

			priorityRows := make([]priorityStatsDisplayRow, 0, len(priorities))
			for _, p := range priorities {
				stats := tq.StatsByPriorityKey[p]
				priorityRows = append(priorityRows, priorityStatsDisplayRow{
					Priority:                p,
					ApproximateBacklogCount: stats.ApproximateBacklogCount,
					ApproximateBacklogAge:   formatDurationShort(stats.ApproximateBacklogAge),
					BacklogIncreaseRate:     stats.BacklogIncreaseRate,
					TasksAddRate:            stats.TasksAddRate,
					TasksDispatchRate:       stats.TasksDispatchRate,
				})
			}
			if err := cctx.Printer.PrintStructured(priorityRows, printer.StructuredOptions{Table: &printer.TableOptions{}}); err != nil {
				return fmt.Errorf("displaying priority stats failed: %w", err)
			}
		}
	} else {
		// Show basic table without stats
		rows := make([]taskQueueDisplayRowBasic, 0, len(taskQueues))
		for _, tq := range taskQueues {
			rows = append(rows, taskQueueDisplayRowBasic{
				Name: tq.Name,
				Type: tq.Type,
			})
		}
		if err := cctx.Printer.PrintStructured(rows, printer.StructuredOptions{Table: &printer.TableOptions{}}); err != nil {
			return fmt.Errorf("displaying task queues failed: %w", err)
		}
	}

	return nil
}

func sortInt32s(s []int32) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// defaultPriorityKey is the default priority key value. When this is the only
// priority key present, we skip showing per-priority stats as it would be redundant.
const defaultPriorityKey = 3

// hasNonDefaultPriorityKeys returns true if the map contains any priority keys
// other than the default (3), or contains multiple priority keys.
func hasNonDefaultPriorityKeys(statsByPriorityKey map[int32]formattedVersionStatsRowType) bool {
	if len(statsByPriorityKey) == 0 {
		return false
	}
	if len(statsByPriorityKey) > 1 {
		return true
	}
	// Exactly one key - check if it's the default
	_, hasDefault := statsByPriorityKey[defaultPriorityKey]
	return !hasDefault
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

func (c *TemporalWorkerDeploymentCreateCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	ns := c.Parent.Parent.Namespace
	identity := c.Parent.Parent.Identity
	deploymentName := c.Name
	requestID := uuid.NewString()

	request := &workflowservice.CreateWorkerDeploymentRequest{
		Namespace:      ns,
		DeploymentName: deploymentName,
		Identity:       identity,
		RequestId:      requestID,
	}

	_, err = cl.WorkflowService().CreateWorkerDeployment(cctx, request)
	if err != nil {
		return fmt.Errorf("error creating worker deployment: %w", err)
	}

	cctx.Printer.Println("Successfully created worker deployment")
	return nil
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

func validateAWSLambdaProviderDetails(details map[string]any, skipRoleAndExternalID bool) error {
	keys := []string{"arn"}
	if !skipRoleAndExternalID {
		// The server governs whether these are mandatory via its
		// require_role_and_external_id setting; --aws-lambda-skip-role-and-external-id
		// opts out of the client-side check for servers where it is disabled.
		keys = append(keys, "role", "role_external_id")
	}
	for _, key := range keys {
		if _, ok := details[key]; !ok {
			return fmt.Errorf("missing required AWS Lambda provider detail: %s", key)
		}
	}
	return nil
}

// awsLambdaProviderDetailsPayload returns the encoded Payload representing AWS
// Lambda compute provider details.
func awsLambdaProviderDetailsPayload(
	functionARN string,
	assumeRoleARN string,
	assumeRoleExternalID string,
	skipRoleAndExternalID bool,
) (*commonpb.Payload, error) {
	// Map keys from temporal-auto-scaled-workers:
	// https://github.com/temporalio/temporal-auto-scaled-workers/blob/c4a7e69b6504365d7e5326b0b8e6cd95e3293f96/wci/workflow/compute_provider/aws_lambda.go#L16-L20
	providerDetails := map[string]any{
		"arn": functionARN,
	}
	if assumeRoleARN != "" {
		providerDetails["role"] = assumeRoleARN
	}
	if assumeRoleExternalID != "" {
		providerDetails["role_external_id"] = assumeRoleExternalID
	}
	err := validateAWSLambdaProviderDetails(providerDetails, skipRoleAndExternalID)
	if err != nil {
		return nil, err
	}
	dc := converter.GetDefaultDataConverter()
	return dc.ToPayload(&providerDetails)
}

func validateGCPCloudRunProviderDetails(details map[string]any) error {
	for _, key := range []string{"project", "region", "worker_pool", "service_account"} {
		if v, ok := details[key].(string); !ok || v == "" {
			return fmt.Errorf("missing required GCP Cloud Run provider detail: %s", key)
		}
	}
	return nil
}

// gcpCloudRunProviderDetailsPayload returns the encoded Payload representing GCP
// Cloud Run compute provider details. All four keys are required: the first
// three name the worker-pool resource and service_account is the impersonation
// target the serverless chain depends on.
func gcpCloudRunProviderDetailsPayload(
	project string,
	region string,
	workerPool string,
	serviceAccount string,
) (*commonpb.Payload, error) {
	// Map keys from temporal-auto-scaled-workers:
	// https://github.com/temporalio/temporal-auto-scaled-workers/blob/d1390d11cb55b4450141ede559f7832e5620c1e4/wci/workflow/compute_provider/gcp_cloudrun.go#L21-L24
	providerDetails := map[string]any{
		"project":         project,
		"region":          region,
		"worker_pool":     workerPool,
		"service_account": serviceAccount,
	}
	err := validateGCPCloudRunProviderDetails(providerDetails)
	if err != nil {
		return nil, err
	}
	dc := converter.GetDefaultDataConverter()
	return dc.ToPayload(&providerDetails)
}

// computeProviderConfig selects the single compute provider for a Worker
// Deployment Version's "default" scaling group from the command's flags. It
// enforces that AWS Lambda and GCP Cloud Run flags are not mixed, then
// dispatches on the trigger flag (--aws-lambda-function-arn /
// --gcp-cloud-run-worker-pool). Returns an empty providerType when no provider
// flags are set, leaving the "no configuration" decision to the caller.
func computeProviderConfig(
	awsLambdaFunctionARN string,
	awsLambdaAssumeRoleARN string,
	awsLambdaAssumeRoleExternalID string,
	awsLambdaSkipRoleAndExternalID bool,
	gcpCloudRunProject string,
	gcpCloudRunRegion string,
	gcpCloudRunWorkerPool string,
	gcpCloudRunServiceAccount string,
) (providerType string, detailsPayload *commonpb.Payload, err error) {
	awsSet := awsLambdaFunctionARN != "" || awsLambdaAssumeRoleARN != "" || awsLambdaAssumeRoleExternalID != ""
	gcpSet := gcpCloudRunProject != "" || gcpCloudRunRegion != "" || gcpCloudRunWorkerPool != "" || gcpCloudRunServiceAccount != ""
	if awsSet && gcpSet {
		return "", nil, fmt.Errorf("cannot combine --aws-lambda-* and --gcp-cloud-run-* flags; a Worker Deployment Version supports a single compute provider")
	}

	switch {
	case awsLambdaFunctionARN != "":
		p, err := awsLambdaProviderDetailsPayload(
			awsLambdaFunctionARN,
			awsLambdaAssumeRoleARN,
			awsLambdaAssumeRoleExternalID,
			awsLambdaSkipRoleAndExternalID,
		)
		return "aws-lambda", p, err
	case gcpCloudRunWorkerPool != "":
		p, err := gcpCloudRunProviderDetailsPayload(
			gcpCloudRunProject,
			gcpCloudRunRegion,
			gcpCloudRunWorkerPool,
			gcpCloudRunServiceAccount,
		)
		return "gcp-cloud-run", p, err
	default:
		return "", nil, nil
	}
}

// scalerTypeByProvider maps each compute provider to the scaling algorithm
// compatible with its launch strategy in temporal-auto-scaled-workers: aws-lambda
// is invoke-based ("no-sync"); gcp-cloud-run is worker-set-based ("rate-based").
// WCI rejects an incompatible pairing at CreateWorkerDeploymentVersion.
var scalerTypeByProvider = map[string]string{
	"aws-lambda":    "no-sync",
	"gcp-cloud-run": "rate-based",
}

// scalerTypeForProvider returns the scaling algorithm for the given provider,
// erroring if the provider has no explicit mapping so an unknown or newly-added
// provider fails loudly here rather than silently getting an incompatible scaler.
func scalerTypeForProvider(providerType string) (string, error) {
	if scaler, ok := scalerTypeByProvider[providerType]; ok {
		return scaler, nil
	}
	return "", fmt.Errorf("no scaler mapping for compute provider %q", providerType)
}

// gcpCloudRunScalerDetails builds the ComputeScaler.Details payload from the GCP
// Cloud Run scaling flags. It carries two independent groups of rate-based scaler
// settings:
//   - the instance-count group (min_count/max_count/initial_count), which is
//     all-or-none and must satisfy min <= initial <= max, and
//   - utilization_target, a standalone fraction in (0, 1].
//
// The *Set booleans come from cobra's Flags().Changed, so an omitted flag stays
// distinct from an explicit 0. Returns a nil payload when nothing is set, leaving
// WCI's defaults (min 0, max 30, initial 0, utilization_target 0.8) in effect.
// Every setting is GCP Cloud Run only; any use with another provider is rejected.
// Config keys mirror the WCI rate-based scaler:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/scaling_algorithm/rate_based.go
// gcpScalerFlags holds the GCP Cloud Run scaling flag values together with
// whether each was actually set (from cobra's Flags().Changed). Pairing each
// value with its Set bool keeps an omitted flag distinct from an explicit 0 and
// removes the positional-argument risk of passing the raw values around.
type gcpScalerFlags struct {
	min            int
	minSet         bool
	max            int
	maxSet         bool
	initial        int
	initialSet     bool
	utilization    float32
	utilizationSet bool
}

func (f gcpScalerFlags) anySet() bool {
	return f.minSet || f.maxSet || f.initialSet || f.utilizationSet
}

func (f gcpScalerFlags) allSet() bool {
	return f.minSet && f.maxSet && f.initialSet && f.utilizationSet
}

// gcpCloudRunScalerDetails builds the ComputeScaler.Details payload from the GCP
// Cloud Run scaling flags (min/max/initial instance counts and utilization
// target). The four form a single all-or-none group: setting any one requires
// all four. That keeps the min<=initial<=max relationship self-contained and
// avoids comparing an explicit value against WCI's default for an unset sibling.
// Returns a nil payload when nothing is set, leaving WCI's defaults (min 0,
// max 30, initial 0, utilization_target 0.8) in effect. Every setting is GCP
// Cloud Run only; any use with another provider is rejected.
func gcpCloudRunScalerDetails(providerType string, f gcpScalerFlags) (*commonpb.Payload, error) {
	if !f.anySet() {
		return nil, nil
	}
	// These are GCP Cloud Run (rate-based) knobs only. Reject on any use with
	// another provider so the GCP-only nature is explicit, regardless of value.
	if providerType != "gcp-cloud-run" {
		return nil, fmt.Errorf("the Cloud Run scaling flags are only valid with --gcp-cloud-run-worker-pool")
	}
	if !f.allSet() {
		return nil, fmt.Errorf("--gcp-cloud-run-min-instances, --gcp-cloud-run-max-instances, --gcp-cloud-run-initial-instances, and --gcp-cloud-run-utilization-target must be set together")
	}
	if f.min < 0 {
		return nil, fmt.Errorf("--gcp-cloud-run-min-instances cannot be negative")
	}
	if f.max < 1 {
		return nil, fmt.Errorf("--gcp-cloud-run-max-instances must be at least 1")
	}
	if f.min > f.max {
		return nil, fmt.Errorf("--gcp-cloud-run-min-instances cannot exceed --gcp-cloud-run-max-instances")
	}
	if f.initial < f.min || f.initial > f.max {
		return nil, fmt.Errorf("--gcp-cloud-run-initial-instances must be between --gcp-cloud-run-min-instances and --gcp-cloud-run-max-instances")
	}
	// The (0, 1] range is intrinsic to utilization_target's meaning, not a
	// default that could drift, so mirroring the check here is safe.
	if f.utilization <= 0 || f.utilization > 1 {
		return nil, fmt.Errorf("--gcp-cloud-run-utilization-target must be greater than 0 and at most 1")
	}
	details := map[string]any{
		scalerKeyMinCount:          f.min,
		scalerKeyMaxCount:          f.max,
		scalerKeyInitialCount:      f.initial,
		scalerKeyUtilizationTarget: f.utilization,
	}
	dc := converter.GetDefaultDataConverter()
	return dc.ToPayload(&details)
}

func (c *TemporalWorkerDeploymentCreateVersionCommand) gcpScalerFlags() gcpScalerFlags {
	f := c.Command.Flags()
	return gcpScalerFlags{
		min: c.GcpCloudRunMinInstances, minSet: f.Changed("gcp-cloud-run-min-instances"),
		max: c.GcpCloudRunMaxInstances, maxSet: f.Changed("gcp-cloud-run-max-instances"),
		initial: c.GcpCloudRunInitialInstances, initialSet: f.Changed("gcp-cloud-run-initial-instances"),
		utilization: c.GcpCloudRunUtilizationTarget, utilizationSet: f.Changed("gcp-cloud-run-utilization-target"),
	}
}

func (c *TemporalWorkerDeploymentCreateVersionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	ns := c.Parent.Parent.Namespace
	buildID := c.BuildId
	identity := c.Parent.Parent.Identity
	deploymentName := c.DeploymentName
	requestID := uuid.NewString()

	providerType, detailsPayload, err := computeProviderConfig(
		c.AwsLambdaFunctionArn,
		c.AwsLambdaAssumeRoleArn,
		c.AwsLambdaAssumeRoleExternalId,
		c.AwsLambdaSkipRoleAndExternalId,
		c.GcpCloudRunProject,
		c.GcpCloudRunRegion,
		c.GcpCloudRunWorkerPool,
		c.GcpCloudRunServiceAccount,
	)
	if err != nil {
		return err
	}
	if providerType == "" {
		// We do not allow creation of an "empty" WDV.
		return fmt.Errorf("missing configuration for compute provider")
	}
	scalerType, err := scalerTypeForProvider(providerType)
	if err != nil {
		return err
	}
	scalerDetails, err := gcpCloudRunScalerDetails(providerType, c.gcpScalerFlags())
	if err != nil {
		return err
	}
	cc := &computepb.ComputeConfig{
		ScalingGroups: map[string]*computepb.ComputeConfigScalingGroup{
			"default": {
				Provider: &computepb.ComputeProvider{
					Type:    providerType,
					Details: detailsPayload,
				},
				Scaler: &computepb.ComputeScaler{
					Type:    scalerType,
					Details: scalerDetails,
				},
			},
		},
	}
	request := &workflowservice.CreateWorkerDeploymentVersionRequest{
		Namespace: ns,
		DeploymentVersion: &deployment.WorkerDeploymentVersion{
			DeploymentName: deploymentName,
			BuildId:        buildID,
		},
		Identity:      identity,
		ComputeConfig: cc,
		RequestId:     requestID,
	}

	_, err = cl.WorkflowService().CreateWorkerDeploymentVersion(cctx, request)
	if err != nil {
		return fmt.Errorf("error creating worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully created worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentUpdateVersionComputeConfigCommand) gcpScalerFlags() gcpScalerFlags {
	f := c.Command.Flags()
	return gcpScalerFlags{
		min: c.GcpCloudRunMinInstances, minSet: f.Changed("gcp-cloud-run-min-instances"),
		max: c.GcpCloudRunMaxInstances, maxSet: f.Changed("gcp-cloud-run-max-instances"),
		initial: c.GcpCloudRunInitialInstances, initialSet: f.Changed("gcp-cloud-run-initial-instances"),
		utilization: c.GcpCloudRunUtilizationTarget, utilizationSet: f.Changed("gcp-cloud-run-utilization-target"),
	}
}

func (c *TemporalWorkerDeploymentUpdateVersionComputeConfigCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	ns := c.Parent.Parent.Namespace
	buildID := c.BuildId
	identity := c.Parent.Parent.Identity
	deploymentName := c.DeploymentName
	requestID := uuid.NewString()

	request := &workflowservice.UpdateWorkerDeploymentVersionComputeConfigRequest{
		Namespace: ns,
		DeploymentVersion: &deployment.WorkerDeploymentVersion{
			DeploymentName: deploymentName,
			BuildId:        buildID,
		},
		Identity:  identity,
		RequestId: requestID,
	}

	if c.Remove {
		if c.AwsLambdaFunctionArn != "" || c.AwsLambdaAssumeRoleArn != "" || c.AwsLambdaAssumeRoleExternalId != "" ||
			c.GcpCloudRunProject != "" || c.GcpCloudRunRegion != "" || c.GcpCloudRunWorkerPool != "" || c.GcpCloudRunServiceAccount != "" ||
			c.gcpScalerFlags().anySet() {
			return fmt.Errorf("--remove cannot be combined with --aws-lambda-* or --gcp-cloud-run-* flags")
		}
		request.RemoveComputeConfigScalingGroups = []string{"default"}
	} else {
		providerType, detailsPayload, err := computeProviderConfig(
			c.AwsLambdaFunctionArn,
			c.AwsLambdaAssumeRoleArn,
			c.AwsLambdaAssumeRoleExternalId,
			c.AwsLambdaSkipRoleAndExternalId,
			c.GcpCloudRunProject,
			c.GcpCloudRunRegion,
			c.GcpCloudRunWorkerPool,
			c.GcpCloudRunServiceAccount,
		)
		if err != nil {
			return err
		}
		scalerFlags := c.gcpScalerFlags()

		var sg *computepb.ComputeConfigScalingGroup
		var updatePaths []string
		switch {
		case providerType != "":
			// Provider (re)configuration: rebuild the provider and scaler type,
			// and keep scaler.details consistent with the (possibly new) type.
			scalerType, err := scalerTypeForProvider(providerType)
			if err != nil {
				return err
			}
			scalerDetails, err := gcpCloudRunScalerDetails(providerType, scalerFlags)
			if err != nil {
				return err
			}
			sg = &computepb.ComputeConfigScalingGroup{
				Provider: &computepb.ComputeProvider{
					Type:    providerType,
					Details: detailsPayload,
				},
				Scaler: &computepb.ComputeScaler{
					Type:    scalerType,
					Details: scalerDetails,
				},
			}
			updatePaths = []string{"provider.type", "provider.details", "scaler.type"}
			// scaler.details must stay consistent with the scaler type. A
			// non-GCP scaler (no-sync) can't carry the rate-based details, so
			// clear them when switching away from GCP; for GCP, overwrite only
			// when new settings were supplied, otherwise preserve existing ones.
			if providerType != "gcp-cloud-run" || scalerDetails != nil {
				updatePaths = append(updatePaths, "scaler.details")
			}
		case scalerFlags.anySet():
			// Scaler-only update: change scaler.details without touching the
			// provider or scaler type. Targets the version's existing GCP Cloud
			// Run scaler; the server validates against the real config.
			scalerDetails, err := gcpCloudRunScalerDetails("gcp-cloud-run", scalerFlags)
			if err != nil {
				return err
			}
			sg = &computepb.ComputeConfigScalingGroup{
				Scaler: &computepb.ComputeScaler{Details: scalerDetails},
			}
			updatePaths = []string{"scaler.details"}
		default:
			return fmt.Errorf("no compute configuration provided to update")
		}

		request.ComputeConfigScalingGroups = map[string]*computepb.ComputeConfigScalingGroupUpdate{
			"default": {
				ScalingGroup: sg,
				UpdateMask:   &fieldmaskpb.FieldMask{Paths: updatePaths},
			},
		}
	}

	_, err = cl.WorkflowService().UpdateWorkerDeploymentVersionComputeConfig(cctx, request)
	if err != nil {
		return fmt.Errorf("error updating worker deployment version compute config: %w", err)
	}

	if c.Remove {
		cctx.Printer.Println("Successfully removed worker deployment version compute config")
	} else {
		cctx.Printer.Println("Successfully updated worker deployment version compute config")
	}
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

	// Use raw gRPC instead of SDK's DeploymentClient to avoid circular dependency
	// with SDK release that exposes TaskQueuesInfos in DescribeVersionResponse
	resp, err := cl.WorkflowService().DescribeWorkerDeploymentVersion(cctx, &workflowservice.DescribeWorkerDeploymentVersionRequest{
		Namespace: c.Parent.Parent.Namespace,
		DeploymentVersion: &deploymentpb.WorkerDeploymentVersion{
			DeploymentName: c.DeploymentName,
			BuildId:        c.BuildId,
		},
		ReportTaskQueueStats: c.ReportTaskQueueStats,
	})
	if err != nil {
		return fmt.Errorf("error describing worker deployment version: %w", err)
	}

	err = printWorkerDeploymentVersionInfoProto(cctx, resp.GetWorkerDeploymentVersionInfo(), resp.GetVersionTaskQueues(), "Worker Deployment Version:", printVersionInfoOptions{
		showStats: c.ReportTaskQueueStats,
	})
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

	if c.BuildId != "" && c.Unversioned {
		return fmt.Errorf("specify either --build-id or --unversioned, not both")
	}

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

func (c *TemporalWorkerDeploymentUpdateVersionMetadataCommand) run(cctx *CommandContext, args []string) error {
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
