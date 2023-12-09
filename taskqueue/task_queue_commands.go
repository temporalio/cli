package taskqueue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/tqname"
)

// DescribeTaskQueue show pollers info of a given taskqueue
func DescribeTaskQueue(c *cli.Context) error {
	taskQueue := c.String(common.FlagTaskQueue)
	tqName, err := tqname.FromBaseName(taskQueue)
	if err != nil {
		return err
	}
	taskQueueType := strToTaskQueueType(c.String(common.FlagTaskQueueType))
	partitions := c.Int(common.FlagPartitions)

	ctx, cancel := common.NewContext(c)
	defer cancel()

	frontendClient := client.Factory(c.App).FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	type statusWithPartition struct {
		Partition int `json:"partition"`
		taskqueuepb.TaskQueueStatus
	}
	type pollerWithPartition struct {
		Partition int `json:"partition"`
		taskqueuepb.PollerInfo
		// copy this out to display nicer in table or card, but not json
		VersionCaps *commonpb.WorkerVersionCapabilities `json:"-"`
	}

	var statuses []any
	var pollers []any

	for p := 0; p < partitions; p++ {
		resp, err := frontendClient.DescribeTaskQueue(ctx, &workflowservice.DescribeTaskQueueRequest{
			Namespace: namespace,
			TaskQueue: &taskqueuepb.TaskQueue{
				Name: tqName.WithPartition(p).FullName(),
				Kind: enumspb.TASK_QUEUE_KIND_NORMAL,
			},
			TaskQueueType:          taskQueueType,
			IncludeTaskQueueStatus: true,
		})
		// note that even if it doesn't exist before this call, DescribeTaskQueue will return something
		if err != nil {
			return fmt.Errorf("unable to describe task queue: %w", err)
		}
		statuses = append(statuses, &statusWithPartition{
			Partition:       p,
			TaskQueueStatus: *resp.TaskQueueStatus,
		})
		for _, pi := range resp.Pollers {
			pollers = append(pollers, &pollerWithPartition{
				Partition:   p,
				PollerInfo:  *pi,
				VersionCaps: pi.WorkerVersionCapabilities,
			})
		}
	}

	if output.OutputOption(c.String(output.FlagOutput)) == output.JSON {
		// handle specially so we output a single object instead of two
		b, err := json.MarshalIndent(map[string]any{
			"taskQueues": statuses,
			"pollers":    pollers,
		}, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Println(string(b))
		return err
	}

	opts := &output.PrintOptions{
		Fields: []string{"Partition", "TaskQueueStatus.RatePerSecond", "TaskQueueStatus.BacklogCountHint", "TaskQueueStatus.ReadLevel", "TaskQueueStatus.AckLevel", "TaskQueueStatus.TaskIdBlock"},
	}
	err = output.PrintItems(c, statuses, opts)
	if err != nil {
		return err
	}

	opts = &output.PrintOptions{
		Fields: []string{"Partition", "PollerInfo.Identity", "PollerInfo.LastAccessTime", "PollerInfo.RatePerSecond", "VersionCaps.BuildId", "VersionCaps.UseVersioning"},
	}
	return output.PrintItems(c, pollers, opts)
}

// ListTaskQueuePartitions gets all the taskqueue partition and host information.
func ListTaskQueuePartitions(c *cli.Context) error {
	frontendClient := client.Factory(c.App).FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	taskQueue := c.String(common.FlagTaskQueue)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	request := &workflowservice.ListTaskQueuePartitionsRequest{
		Namespace: namespace,
		TaskQueue: &taskqueuepb.TaskQueue{
			Name: taskQueue,
			Kind: enumspb.TASK_QUEUE_KIND_NORMAL,
		},
	}

	resp, err := frontendClient.ListTaskQueuePartitions(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to list task queues: %w", err)
	}

	optsW := &output.PrintOptions{
		Fields: []string{"Key", "OwnerHostName"},
	}

	var items []interface{}
	fmt.Println(color.Magenta(c, "Workflow Task Queue Partitions\n"))
	for _, e := range resp.WorkflowTaskQueuePartitions {
		items = append(items, e)
	}
	err = output.PrintItems(c, items, optsW)
	if err != nil {
		return err
	}

	optsA := &output.PrintOptions{
		Fields: []string{"Key", "OwnerHostName"},
	}
	items = items[:0]
	fmt.Println(color.Magenta(c, "\nActivity Task Queue Partitions\n"))
	for _, e := range resp.ActivityTaskQueuePartitions {
		items = append(items, e)
	}
	return output.PrintItems(c, items, optsA)
}

// BuildIDAddNewDefault is implements the `update-build-ids add-new-default` subcommand
func BuildIDAddNewDefault(c *cli.Context) error {
	newBuildID := c.String(common.FlagBuildID)
	operation := workflowservice.UpdateWorkerBuildIdCompatibilityRequest{
		Operation: &workflowservice.UpdateWorkerBuildIdCompatibilityRequest_AddNewBuildIdInNewDefaultSet{
			AddNewBuildIdInNewDefaultSet: newBuildID,
		},
	}
	return updateBuildIDs(c, operation)
}

// BuildIDAddNewCompatible implements the `update-build-ids add-new-compatible` subcommand
func BuildIDAddNewCompatible(c *cli.Context) error {
	newBuildID := c.String(common.FlagBuildID)
	existingBuildID := c.String(common.FlagExistingCompatibleBuildID)
	setAsDefault := c.Bool(common.FlagSetBuildIDAsDefault)
	operation := workflowservice.UpdateWorkerBuildIdCompatibilityRequest{
		Operation: &workflowservice.UpdateWorkerBuildIdCompatibilityRequest_AddNewCompatibleBuildId{
			AddNewCompatibleBuildId: &workflowservice.UpdateWorkerBuildIdCompatibilityRequest_AddNewCompatibleVersion{
				NewBuildId:                newBuildID,
				ExistingCompatibleBuildId: existingBuildID,
				MakeSetDefault:            setAsDefault,
			},
		},
	}
	return updateBuildIDs(c, operation)
}

// BuildIDPromoteSet implements the `update-build-ids promote-set` subcommand
func BuildIDPromoteSet(c *cli.Context) error {
	buildID := c.String(common.FlagBuildID)
	operation := workflowservice.UpdateWorkerBuildIdCompatibilityRequest{
		Operation: &workflowservice.UpdateWorkerBuildIdCompatibilityRequest_PromoteSetByBuildId{
			PromoteSetByBuildId: buildID,
		},
	}
	return updateBuildIDs(c, operation)
}

// BuildIDPromoteInSet implements the `update-build-ids promote-id-in-set` subcommand
func BuildIDPromoteInSet(c *cli.Context) error {
	buildID := c.String(common.FlagBuildID)
	operation := workflowservice.UpdateWorkerBuildIdCompatibilityRequest{
		Operation: &workflowservice.UpdateWorkerBuildIdCompatibilityRequest_PromoteBuildIdWithinSet{
			PromoteBuildIdWithinSet: buildID,
		},
	}
	return updateBuildIDs(c, operation)
}

// GetBuildIDs is implements the `get-build-ids` subcommand
func GetBuildIDs(c *cli.Context) error {
	frontendClient := client.Factory(c.App).FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	taskQueue := c.String(common.FlagTaskQueue)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	request := &workflowservice.GetWorkerBuildIdCompatibilityRequest{
		Namespace: namespace,
		TaskQueue: taskQueue,
	}

	resp, err := frontendClient.GetWorkerBuildIdCompatibility(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to get task queue build ids: %w", err)
	}

	type rowtype struct {
		VersionSetId  string
		BuildIds      []string
		IsDefaultSet  bool
		DefaultForSet string
	}
	opts := &output.PrintOptions{
		Fields: []string{"BuildIds", "DefaultForSet", "IsDefaultSet"},
	}
	var items []interface{}
	for ix, e := range resp.GetMajorVersionSets() {
		row := rowtype{
			BuildIds:      e.GetBuildIds(),
			IsDefaultSet:  ix == len(resp.GetMajorVersionSets())-1,
			DefaultForSet: e.GetBuildIds()[len(e.GetBuildIds())-1],
		}
		items = append(items, row)
	}
	return output.PrintItems(c, items, opts)
}

// GetBuildIDReachability implements the `get-build-id-reachability` subcommand
func GetBuildIDReachability(c *cli.Context) error {
	frontendClient := client.Factory(c.App).FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	buildIDs := c.StringSlice(common.FlagBuildID)
	taskQueues := c.StringSlice(common.FlagTaskQueue)
	reachabilityType := strings.ToLower(c.String(common.FlagReachabilityType))
	reachability := enumspb.TASK_REACHABILITY_UNSPECIFIED
	if reachabilityType != "" {
		if reachabilityType == "open" {
			reachability = enumspb.TASK_REACHABILITY_OPEN_WORKFLOWS
		} else if reachabilityType == "closed" {
			reachability = enumspb.TASK_REACHABILITY_CLOSED_WORKFLOWS
		} else if reachabilityType == "existing" {
			reachability = enumspb.TASK_REACHABILITY_EXISTING_WORKFLOWS
		} else {
			return fmt.Errorf("invalid reachability type: %v", reachabilityType)
		}
	}

	ctx, cancel := common.NewContext(c)
	defer cancel()
	request := &workflowservice.GetWorkerTaskReachabilityRequest{
		Namespace:    namespace,
		BuildIds:     buildIDs,
		TaskQueues:   taskQueues,
		Reachability: reachability,
	}

	resp, err := frontendClient.GetWorkerTaskReachability(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to get Build ID reachability: %w", err)
	}

	type rowtype struct {
		BuildId      string
		TaskQueue    string
		Reachability []string
	}
	opts := &output.PrintOptions{
		Fields: []string{"BuildId", "TaskQueue", "Reachability"},
	}
	var items []interface{}
	for _, e := range resp.GetBuildIdReachability() {
		for _, r := range e.GetTaskQueueReachability() {
			reachability := make([]string, len(r.GetReachability()))
			for i, v := range r.GetReachability() {
				reachability[i] = v.String()
			}
			row := rowtype{
				BuildId:      e.GetBuildId(),
				TaskQueue:    r.GetTaskQueue(),
				Reachability: reachability,
			}
			items = append(items, row)
		}
	}
	return output.PrintItems(c, items, opts)
}

// updateBuildIDs manipulates the build ids of a given taskqueue. `partialReq` is a partial request
// containing only the operation field filled out.
func updateBuildIDs(c *cli.Context, partialReq workflowservice.UpdateWorkerBuildIdCompatibilityRequest) error {
	frontendClient := client.Factory(c.App).FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	taskQueue := c.String(common.FlagTaskQueue)

	ctx, cancel := common.NewContext(c)
	defer cancel()

	request := &workflowservice.UpdateWorkerBuildIdCompatibilityRequest{
		Namespace: namespace,
		TaskQueue: taskQueue,
		Operation: partialReq.Operation,
	}

	if _, err = frontendClient.UpdateWorkerBuildIdCompatibility(ctx, request); err != nil {
		return fmt.Errorf("error updating task queue build IDs: %w", err)
	}

	fmt.Println(color.Green(c, "Successfully updated task queue build IDs"))

	return nil
}

func strToTaskQueueType(str string) enumspb.TaskQueueType {
	if strings.ToLower(str) == "activity" {
		return enumspb.TASK_QUEUE_TYPE_ACTIVITY
	}
	return enumspb.TASK_QUEUE_TYPE_WORKFLOW
}
