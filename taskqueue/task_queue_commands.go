package taskqueue

import (
	"fmt"
	"strings"

	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
)

// DescribeTaskQueue show pollers info of a given taskqueue
func DescribeTaskQueue(c *cli.Context) error {
	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}
	taskQueue := c.String(common.FlagTaskQueue)
	taskQueueType := strToTaskQueueType(c.String(common.FlagTaskQueueType))

	ctx, cancel := common.NewContext(c)
	defer cancel()
	resp, err := sdkClient.DescribeTaskQueue(ctx, taskQueue, taskQueueType)
	if err != nil {
		return fmt.Errorf("unable to describe task queue: %w", err)
	}

	opts := &output.PrintOptions{
		// TODO enable when versioning feature is out
		// Fields: []string{"Identity", "LastAccessTime", "RatePerSecond", "WorkerVersioningId"},
		Fields: []string{"Identity", "LastAccessTime", "RatePerSecond"},
	}
	var items []interface{}
	for _, e := range resp.Pollers {
		items = append(items, e)
	}
	return output.PrintItems(c, items, opts)
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
	frontendClient := client.CFactory.FrontendClient(c)
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
		Fields: []string{"BuildId", "DefaultForSet", "IsDefaultSet"},
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
	frontendClient := client.CFactory.FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	buildIDs := c.StringSlice(common.FlagBuildID)
	taskQueues := c.StringSlice(common.FlagTaskQueue)
	ignoreClosed := c.Bool(common.FlagIgnoreClosedWorkflows)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	request := &workflowservice.GetWorkerTaskReachabilityRequest{
		Namespace:  namespace,
		BuildIds:   buildIDs,
		TaskQueues: taskQueues,
	}
	if ignoreClosed {
		request.Reachability = enumspb.TASK_REACHABILITY_OPEN_WORKFLOWS
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
	frontendClient := client.CFactory.FrontendClient(c)
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

	resp, err := frontendClient.UpdateWorkerBuildIdCompatibility(ctx, request)
	if err != nil {
		return fmt.Errorf("error updating task queue build ids: %w", err)
	}

	fmt.Println(color.Green(c, "Successfully updated task queue build ids. Set ID: %v", resp.GetVersionSetId()))

	return nil
}

func strToTaskQueueType(str string) enumspb.TaskQueueType {
	if strings.ToLower(str) == "activity" {
		return enumspb.TASK_QUEUE_TYPE_ACTIVITY
	}
	return enumspb.TASK_QUEUE_TYPE_WORKFLOW
}
