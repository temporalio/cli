package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/common/tqid"
)

const taskQueueUnversioned = "UNVERSIONED"

type taskQueueReachabilityRowType struct {
	BuildID      string `json:"buildId"`
	Reachability string `json:"reachability"`
}

type pollerRowType struct {
	BuildID        string    `json:"buildId"`
	TaskQueueType  string    `json:"taskQueueType"`
	Identity       string    `json:"identity"`
	LastAccessTime time.Time `json:"lastAccessTime"`
	RatePerSecond  float64   `json:"ratePerSecond"`
}

type statsRowType struct {
	BuildID                 string  `json:"buildId"`
	TaskQueueType           string  `json:"taskQueueType"`
	ApproximateBacklogCount int64   `json:"approximateBacklogCount"`
	ApproximateBacklogAge   string  `json:"approximateBacklogAge"`
	BacklogIncreaseRate     float32 `json:"backlogIncreaseRate"`
	TasksAddRate            float32 `json:"tasksAddRate"`
	TasksDispatchRate       float32 `json:"tasksDispatchRate"`
}

type taskQueueDescriptionType struct {
	Reachability []taskQueueReachabilityRowType `json:"reachability"`
	Pollers      []pollerRowType                `json:"pollers"`
	Stats        []statsRowType                 `json:"stats"`
}

func reachabilityToStr(reachability client.BuildIDTaskReachability) (string, error) {
	switch reachability {
	case client.BuildIDTaskReachabilityUnspecified:
		return "unspecified", nil
	case client.BuildIDTaskReachabilityReachable:
		return "reachable", nil
	case client.BuildIDTaskReachabilityClosedWorkflowsOnly:
		return "closedWorkflowsOnly", nil
	case client.BuildIDTaskReachabilityUnreachable:
		return "unreachable", nil
	default:
		return "", fmt.Errorf("unrecognized reachability type: %d", reachability)
	}
}

func descriptionToReachabilityRows(taskQueueDescription client.TaskQueueDescription) ([]taskQueueReachabilityRowType, error) {
	var rRows []taskQueueReachabilityRowType
	// Unversioned queue first
	val, ok := taskQueueDescription.VersionsInfo[client.UnversionedBuildID]
	if ok {
		reachability, err := reachabilityToStr(val.TaskReachability)
		if err != nil {
			return nil, err
		}
		rRows = append(rRows, taskQueueReachabilityRowType{
			BuildID:      taskQueueUnversioned,
			Reachability: reachability,
		})
	}
	for k, val := range taskQueueDescription.VersionsInfo {
		if k != client.UnversionedBuildID {
			reachability, err := reachabilityToStr(val.TaskReachability)
			if err != nil {
				return nil, err
			}
			rRows = append(rRows, taskQueueReachabilityRowType{
				BuildID:      k,
				Reachability: reachability,
			})
		}
	}
	return rRows, nil
}

func taskQueueTypeToStr(taskQueueType client.TaskQueueType) (string, error) {
	switch taskQueueType {
	case client.TaskQueueTypeUnspecified:
		return "unspecified", nil
	case client.TaskQueueTypeWorkflow:
		return "workflow", nil
	case client.TaskQueueTypeActivity:
		return "activity", nil
	case client.TaskQueueTypeNexus:
		return "nexus", nil
	default:
		return "", fmt.Errorf("unrecognized task queue type: %d", taskQueueType)
	}
}

func buildIDToPollerRows(pRows []pollerRowType, buildID string, typesInfo map[client.TaskQueueType]client.TaskQueueTypeInfo) ([]pollerRowType, error) {
	for t, info := range typesInfo {
		taskQueueType, err := taskQueueTypeToStr(t)
		if err != nil {
			return pRows, err
		}
		for _, p := range info.Pollers {
			pRows = append(pRows, pollerRowType{
				BuildID:        buildID,
				TaskQueueType:  taskQueueType,
				Identity:       p.Identity,
				LastAccessTime: p.LastAccessTime,
				RatePerSecond:  p.RatePerSecond,
			})
		}
	}
	return pRows, nil
}

func descriptionToPollerRows(taskQueueDescription client.TaskQueueDescription) ([]pollerRowType, error) {
	var pRows []pollerRowType
	var err error
	// Unversioned queue first
	val, ok := taskQueueDescription.VersionsInfo[client.UnversionedBuildID]
	if ok {
		pRows, err = buildIDToPollerRows(pRows, taskQueueUnversioned, val.TypesInfo)
		if err != nil {
			return nil, err
		}
	}
	for k, val := range taskQueueDescription.VersionsInfo {
		if k != client.UnversionedBuildID {
			pRows, err = buildIDToPollerRows(pRows, k, val.TypesInfo)
			if err != nil {
				return nil, err
			}
		}
	}
	return pRows, nil
}

func buildIDToStatsRows(statsRows []statsRowType, buildID string, typesInfo map[client.TaskQueueType]client.TaskQueueTypeInfo) ([]statsRowType, error) {
	for t, info := range typesInfo {
		taskQueueType, err := taskQueueTypeToStr(t)
		if err != nil {
			return statsRows, err
		}
		statsRows = append(statsRows, statsRowType{
			BuildID:                 buildID,
			TaskQueueType:           taskQueueType,
			ApproximateBacklogCount: info.Stats.ApproximateBacklogCount,
			ApproximateBacklogAge:   formatDuration(info.Stats.ApproximateBacklogAge),
			BacklogIncreaseRate:     info.Stats.BacklogIncreaseRate,
			TasksAddRate:            info.Stats.TasksAddRate,
			TasksDispatchRate:       info.Stats.TasksDispatchRate,
		})
	}
	return statsRows, nil
}

func descriptionToStatsRows(taskQueueDescription client.TaskQueueDescription) ([]statsRowType, error) {
	var statsRows []statsRowType
	var err error
	// Unversioned queue first
	val, ok := taskQueueDescription.VersionsInfo[client.UnversionedBuildID]
	if ok {
		statsRows, err = buildIDToStatsRows(statsRows, taskQueueUnversioned, val.TypesInfo)
		if err != nil {
			return nil, err
		}
	}
	// Versioned queues
	for k, val := range taskQueueDescription.VersionsInfo {
		if k != client.UnversionedBuildID {
			statsRows, err = buildIDToStatsRows(statsRows, k, val.TypesInfo)
			if err != nil {
				return nil, err
			}
		}
	}
	return statsRows, nil
}

func taskQueueDescriptionToRows(taskQueueDescription client.TaskQueueDescription, reportReachability bool, disableStats bool) (taskQueueDescriptionType, error) {
	var rRows []taskQueueReachabilityRowType
	var statsRows []statsRowType
	if reportReachability {
		var err error
		rRows, err = descriptionToReachabilityRows(taskQueueDescription)
		if err != nil {
			return taskQueueDescriptionType{}, err
		}
	}
	if !disableStats {
		var err error
		statsRows, err = descriptionToStatsRows(taskQueueDescription)
		if err != nil {
			return taskQueueDescriptionType{}, err
		}
	}
	pRows, err := descriptionToPollerRows(taskQueueDescription)
	if err != nil {
		return taskQueueDescriptionType{}, err
	}

	return taskQueueDescriptionType{
		Reachability: rRows,
		Pollers:      pRows,
		Stats:        statsRows,
	}, nil
}

func printTaskQueueDescription(cctx *CommandContext, taskQueueDescription client.TaskQueueDescription, reportReachability bool, disableStats bool) error {
	descRows, err := taskQueueDescriptionToRows(taskQueueDescription, reportReachability, disableStats)
	if err != nil {
		return fmt.Errorf("creating task queue description rows failed: %w", err)
	}

	if !cctx.JSONOutput {
		if reportReachability {
			cctx.Printer.Println(color.MagentaString("Task Reachability:"))
			err = cctx.Printer.PrintStructured(descRows.Reachability, printer.StructuredOptions{Table: &printer.TableOptions{}})
			if err != nil {
				return fmt.Errorf("displaying reachability failed: %w", err)
			}
		}

		if !cctx.JSONOutput {
			if !disableStats {
				cctx.Printer.Println(color.MagentaString("Task Queue Statistics:"))
				err = cctx.Printer.PrintStructured(descRows.Stats, printer.StructuredOptions{Table: &printer.TableOptions{}})
				if err != nil {
					return fmt.Errorf("displaying task queue statistics failed: %w", err)
				}
			}
		}

		cctx.Printer.Println(color.MagentaString("Pollers:"))
		return cctx.Printer.PrintStructured(descRows.Pollers, printer.StructuredOptions{Table: &printer.TableOptions{}})
	}

	// json output
	return cctx.Printer.PrintStructured(descRows, printer.StructuredOptions{})
}

func (c *TemporalTaskQueueDescribeCommand) run(cctx *CommandContext, args []string) error {
	if c.LegacyMode {
		return c.runLegacy(cctx, args)
	}
	// Call describeEnhanced
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}

	var selection *client.TaskQueueVersionSelection
	if len(c.SelectBuildId) > 0 || c.SelectUnversioned || c.SelectAllActive {
		selection = &client.TaskQueueVersionSelection{
			BuildIDs:    c.SelectBuildId,
			Unversioned: c.SelectUnversioned,
			AllActive:   c.SelectAllActive,
		}
	}

	var taskQueueTypes []client.TaskQueueType
	for _, t := range c.TaskQueueType.Values {
		var taskQueueType client.TaskQueueType
		switch t {
		case "workflow":
			taskQueueType = client.TaskQueueTypeWorkflow
		case "activity":
			taskQueueType = client.TaskQueueTypeActivity
		case "nexus":
			taskQueueType = client.TaskQueueTypeNexus
		default:
			return fmt.Errorf("unrecognized task queue type: %s", t)
		}
		taskQueueTypes = append(taskQueueTypes, taskQueueType)
	}

	resp, err := cl.DescribeTaskQueueEnhanced(cctx, client.DescribeTaskQueueEnhancedOptions{
		TaskQueue:              c.TaskQueue,
		Versions:               selection,
		TaskQueueTypes:         taskQueueTypes,
		ReportPollers:          true,
		ReportTaskReachability: c.ReportReachability,
		ReportStats:            !c.DisableStats,
	})
	if err != nil {
		return fmt.Errorf("unable to describe task queue: %w", err)
	}
	return printTaskQueueDescription(cctx, resp, c.ReportReachability, c.DisableStats)
}

func (c *TemporalTaskQueueDescribeCommand) runLegacy(cctx *CommandContext, args []string) error {
	// Call describe
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	var taskQueueType enums.TaskQueueType
	switch c.TaskQueueTypeLegacy.Value {
	case "workflow":
		taskQueueType = enums.TASK_QUEUE_TYPE_WORKFLOW
	case "activity":
		taskQueueType = enums.TASK_QUEUE_TYPE_ACTIVITY
	default:
		return fmt.Errorf("unrecognized task queue type: %q", c.TaskQueueTypeLegacy.Value)
	}

	taskQueue, err := tqid.NewTaskQueueFamily(c.Parent.Namespace, c.TaskQueue)
	if err != nil {
		return fmt.Errorf("failed to parse task queue name: %w", err)
	}
	partitions := c.PartitionsLegacy

	type statusWithPartition struct {
		Partition int `json:"partition"`
		taskqueue.TaskQueueStatus
	}
	type pollerWithPartition struct {
		Partition int `json:"partition"`
		taskqueue.PollerInfo
		// copy this out to display nicer in table or card, but not json
		Versioning *commonpb.WorkerVersionCapabilities `json:"-"`
	}

	var statuses []*statusWithPartition
	var pollers []*pollerWithPartition

	// TODO: remove this when the server does partition fan-out
	for p := 0; p < partitions; p++ {
		resp, err := cl.WorkflowService().DescribeTaskQueue(cctx, &workflowservice.DescribeTaskQueueRequest{
			Namespace: c.Parent.Namespace,
			TaskQueue: &taskqueue.TaskQueue{
				Name: taskQueue.TaskQueue(taskQueueType).NormalPartition(p).RpcName(),
				Kind: enums.TASK_QUEUE_KIND_NORMAL,
			},
			TaskQueueType:          taskQueueType,
			IncludeTaskQueueStatus: true,
		})
		if err != nil {
			return fmt.Errorf("unable to describe task queue: %w", err)
		}
		statuses = append(statuses, &statusWithPartition{
			Partition:       p,
			TaskQueueStatus: *resp.TaskQueueStatus,
		})
		for _, pi := range resp.Pollers {
			pollers = append(pollers, &pollerWithPartition{
				Partition:  p,
				PollerInfo: *pi,
				Versioning: pi.WorkerVersionCapabilities,
			})
		}
	}

	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(map[string]any{
			"taskQueues": statuses,
			"pollers":    pollers,
		}, printer.StructuredOptions{})
	}

	// For text, we will use a table for pollers
	cctx.Printer.Println(color.MagentaString("Pollers:"))
	items := make([]struct {
		Identity       string
		LastAccessTime time.Time
		RatePerSecond  float64
	}, len(pollers))
	for i, poller := range pollers {
		items[i].Identity = poller.Identity
		items[i].LastAccessTime = poller.LastAccessTime.AsTime()
		items[i].RatePerSecond = poller.RatePerSecond
	}
	return cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

func (c *TemporalTaskQueueListPartitionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	request := &workflowservice.ListTaskQueuePartitionsRequest{
		Namespace: c.Parent.Namespace,
		TaskQueue: &taskqueue.TaskQueue{
			Name: c.TaskQueue,
			Kind: enums.TASK_QUEUE_KIND_NORMAL,
		},
	}

	resp, err := cl.WorkflowService().ListTaskQueuePartitions(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to list task queues: %w", err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	var items []*taskqueue.TaskQueuePartitionMetadata
	cctx.Printer.Println(color.MagentaString("Workflow Task Queue Partitions\n"))
	for _, e := range resp.WorkflowTaskQueuePartitions {
		items = append(items, e)
	}
	_ = cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})

	items = items[:0]
	cctx.Printer.Println(color.MagentaString("\nActivity Task Queue Partitions\n"))
	for _, e := range resp.ActivityTaskQueuePartitions {
		items = append(items, e)
	}
	_ = cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})

	return nil
}
