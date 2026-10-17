package temporalcli

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/temporalio/cli/internal/printer"
	commonpb "go.temporal.io/api/common/v1"
	computepb "go.temporal.io/api/compute/v1"
	deploymentpb "go.temporal.io/api/deployment/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func describeVersionComputeConfig(
	cctx *CommandContext,
	cl workflowservice.WorkflowServiceClient,
	namespace string,
	version *deploymentpb.WorkerDeploymentVersion,
) (*computepb.ComputeConfig, error) {
	resp, err := cl.DescribeWorkerDeploymentVersion(cctx, &workflowservice.DescribeWorkerDeploymentVersionRequest{
		Namespace:         namespace,
		DeploymentVersion: version,
	})
	if err != nil {
		return nil, fmt.Errorf("error describing worker deployment version: %w", err)
	}
	return resp.GetWorkerDeploymentVersionInfo().GetComputeConfig(), nil
}

func (o *DeploymentVersionOptions) deploymentVersion() *deploymentpb.WorkerDeploymentVersion {
	return &deploymentpb.WorkerDeploymentVersion{
		DeploymentName: o.DeploymentName,
		BuildId:        o.BuildId,
	}
}

func (c *TemporalWorkerDeploymentVersionCreateCommand) run(cctx *CommandContext, args []string) error {
	worker := c.Parent.Parent.Parent
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().CreateWorkerDeploymentVersion(cctx, &workflowservice.CreateWorkerDeploymentVersionRequest{
		Namespace:         worker.Namespace,
		DeploymentVersion: c.deploymentVersion(),
		Identity:          worker.Identity,
		RequestId:         uuid.NewString(),
	})
	if err != nil {
		return fmt.Errorf("error creating worker deployment version: %w", err)
	}
	cctx.Printer.Println("Successfully created worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentVersionDescribeCommand) run(cctx *CommandContext, args []string) error {
	worker := c.Parent.Parent.Parent
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	resp, err := cl.WorkflowService().DescribeWorkerDeploymentVersion(cctx, &workflowservice.DescribeWorkerDeploymentVersionRequest{
		Namespace:            worker.Namespace,
		DeploymentVersion:    c.deploymentVersion(),
		ReportTaskQueueStats: c.ReportTaskQueueStats,
	})
	if err != nil {
		return fmt.Errorf("error describing worker deployment version: %w", err)
	}
	return printVersionDescription(cctx, resp.GetWorkerDeploymentVersionInfo(), resp.GetVersionTaskQueues(), "Worker Deployment Version:", versionDescribeOptions{
		showStats: c.ReportTaskQueueStats,
	})
}

func (c *TemporalWorkerDeploymentVersionComputeConfigDescribeCommand) run(cctx *CommandContext, args []string) error {
	worker := c.Parent.Parent.Parent.Parent
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	cc, err := describeVersionComputeConfig(cctx, cl.WorkflowService(), worker.Namespace, c.deploymentVersion())
	if err != nil {
		return err
	}
	described := formatComputeConfigDescription(cc)
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(described, printer.StructuredOptions{})
	}
	if len(described.ScalingGroups) == 0 {
		cctx.Printer.Println("Worker deployment version has no compute config")
		return nil
	}

	// One card per scaling group, like the Version's own card, so that a group's
	// settings each get a line. Values are strings so that unset settings are left
	// out while a setting of 0 is still shown.
	type scalingGroupCard struct {
		ScalingGroup                        string
		TaskQueueTypes                      string `cli:",cardOmitEmpty"`
		Provider                            string `cli:",cardOmitEmpty"`
		Scaler                              string `cli:",cardOmitEmpty"`
		ScaleUpBacklogThreshold             string `cli:",cardOmitEmpty"`
		ScaleUpCooloffMs                    string `cli:",cardOmitEmpty"`
		MaxWorkerLifetimeMs                 string `cli:",cardOmitEmpty"`
		MetricsPollIntervalMs               string `cli:",cardOmitEmpty"`
		ScaleUpDispatchRateEpsilon          string `cli:",cardOmitEmpty"`
		ScaleUpDispatchRateEpsilonConfirmMs string `cli:",cardOmitEmpty"`
		SuppressScaleUpMs                   string `cli:",cardOmitEmpty"`
		SuppressPollIntervalMs              string `cli:",cardOmitEmpty"`
	}
	intStr := func(v *int64) string {
		if v == nil {
			return ""
		}
		return strconv.FormatInt(*v, 10)
	}
	keys := slices.Sorted(maps.Keys(described.ScalingGroups))
	cards := make([]scalingGroupCard, 0, len(keys))
	for _, key := range keys {
		sg := described.ScalingGroups[key]
		card := scalingGroupCard{
			ScalingGroup:   key,
			TaskQueueTypes: strings.Join(sg.TaskQueueTypes, ", "),
		}
		if sg.Provider != nil {
			card.Provider = sg.Provider.Type
		}
		if fs := sg.Scaler; fs != nil {
			card.Scaler = fs.Type
			card.ScaleUpBacklogThreshold = intStr(fs.ScaleUpBacklogThreshold)
			card.ScaleUpCooloffMs = intStr(fs.ScaleUpCooloffMs)
			card.MaxWorkerLifetimeMs = intStr(fs.MaxWorkerLifetimeMs)
			card.MetricsPollIntervalMs = intStr(fs.MetricsPollIntervalMs)
			if fs.ScaleUpDispatchRateEpsilon != nil {
				card.ScaleUpDispatchRateEpsilon = strconv.FormatFloat(*fs.ScaleUpDispatchRateEpsilon, 'g', -1, 64)
			}
			card.ScaleUpDispatchRateEpsilonConfirmMs = intStr(fs.ScaleUpDispatchRateEpsilonConfirmMs)
			card.SuppressScaleUpMs = intStr(fs.SuppressScaleUpMs)
			card.SuppressPollIntervalMs = intStr(fs.SuppressPollIntervalMs)
		}
		cards = append(cards, card)
	}
	cctx.Printer.Println(color.MagentaString("Compute Config Scaling Groups:"))
	return cctx.Printer.PrintStructured(cards, printer.StructuredOptions{})
}

func (c *TemporalWorkerDeploymentVersionDeleteCommand) run(cctx *CommandContext, args []string) error {
	worker := c.Parent.Parent.Parent
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	dHandle := cl.WorkerDeploymentClient().GetHandle(c.DeploymentName)
	_, err = dHandle.DeleteVersion(cctx, client.WorkerDeploymentDeleteVersionOptions{
		BuildID:      c.BuildId,
		SkipDrainage: c.SkipDrainage,
		Identity:     worker.Identity,
	})
	if err != nil {
		return fmt.Errorf("error deleting worker deployment version: %w", err)
	}

	cctx.Printer.Println("Successfully deleted worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentVersionComputeConfigRemoveCommand) run(cctx *CommandContext, args []string) error {
	// An empty value is rejected rather than treated as unset, so that an empty
	// variable passed to --scaling-group does not remove every scaling group.
	removeOne := c.Command.Flags().Changed("scaling-group")
	if removeOne && c.ScalingGroup == "" {
		return fmt.Errorf("--scaling-group cannot be empty")
	}

	worker := c.Parent.Parent.Parent.Parent
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	// The scaling groups are read first because the API removes them by key and
	// silently ignores keys that do not exist.
	cc, err := describeVersionComputeConfig(cctx, cl.WorkflowService(), worker.Namespace, c.deploymentVersion())
	if err != nil {
		return err
	}
	var keys []string
	if removeOne {
		if _, ok := cc.GetScalingGroups()[c.ScalingGroup]; !ok {
			return fmt.Errorf("scaling group %q not found; run `temporal worker deployment version compute-config describe` to see the scaling groups of this worker deployment version", c.ScalingGroup)
		}
		keys = []string{c.ScalingGroup}
	} else {
		keys = slices.Sorted(maps.Keys(cc.GetScalingGroups()))
		if len(keys) == 0 {
			cctx.Printer.Println("Worker deployment version has no compute config to remove")
			return nil
		}
		prompt := fmt.Sprintf("Remove all %d scaling groups (%s) from Worker Deployment Version %q with Build ID %q? y/N",
			len(keys), strings.Join(keys, ", "), c.DeploymentName, c.BuildId)
		if len(keys) == 1 {
			prompt = fmt.Sprintf("Remove scaling group %q from Worker Deployment Version %q with Build ID %q? y/N",
				keys[0], c.DeploymentName, c.BuildId)
		}
		yes, err := cctx.promptYes(prompt, c.Yes)
		if err != nil {
			return err
		} else if !yes {
			return fmt.Errorf("user denied confirmation")
		}
	}
	_, err = cl.WorkflowService().UpdateWorkerDeploymentVersionComputeConfig(cctx, &workflowservice.UpdateWorkerDeploymentVersionComputeConfigRequest{
		Namespace:                        worker.Namespace,
		DeploymentVersion:                c.deploymentVersion(),
		RemoveComputeConfigScalingGroups: keys,
		Identity:                         worker.Identity,
		RequestId:                        uuid.NewString(),
	})
	if err != nil {
		return fmt.Errorf("error removing worker deployment version compute config: %w", err)
	}
	cctx.Printer.Println("Successfully removed worker deployment version compute config")
	return nil
}

// scalingGroup builds the provider-independent part of a scaling group from
// the --scaling-group and --task-queue-type flags. Task queue types are
// de-duplicated and sorted so that the request does not depend on flag order or
// repeated values.
func (o *ScalingGroupOptions) scalingGroup() (*computepb.ComputeConfigScalingGroup, error) {
	if o.ScalingGroup == "" {
		return nil, fmt.Errorf("--scaling-group cannot be empty")
	}
	sg := &computepb.ComputeConfigScalingGroup{}
	for _, t := range o.TaskQueueType.Values {
		var tqType enumspb.TaskQueueType
		switch t {
		case "workflow":
			tqType = enumspb.TASK_QUEUE_TYPE_WORKFLOW
		case "activity":
			tqType = enumspb.TASK_QUEUE_TYPE_ACTIVITY
		case "nexus":
			tqType = enumspb.TASK_QUEUE_TYPE_NEXUS
		default:
			return nil, fmt.Errorf("invalid --task-queue-type %q", t)
		}
		if !slices.Contains(sg.TaskQueueTypes, tqType) {
			sg.TaskQueueTypes = append(sg.TaskQueueTypes, tqType)
		}
	}
	slices.Sort(sg.TaskQueueTypes)
	return sg, nil
}

func (c *TemporalWorkerDeploymentVersionComputeConfigAwsLambdaCreateCommand) run(cctx *CommandContext, args []string) error {
	sg, err := awsLambdaScalingGroup(c.FunctionArn, &c.ScalingGroupOptions, &c.AwsAssumeRoleOptions, &c.NoSyncScalerOptions)
	if err != nil {
		return err
	}
	worker := c.Parent.Parent.Parent.Parent.Parent
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().CreateWorkerDeploymentVersion(cctx, &workflowservice.CreateWorkerDeploymentVersionRequest{
		Namespace:         worker.Namespace,
		DeploymentVersion: c.deploymentVersion(),
		ComputeConfig: &computepb.ComputeConfig{
			ScalingGroups: map[string]*computepb.ComputeConfigScalingGroup{c.ScalingGroup: sg},
		},
		Identity:  worker.Identity,
		RequestId: uuid.NewString(),
	})
	var alreadyExists *serviceerror.AlreadyExists
	if errors.As(err, &alreadyExists) {
		return fmt.Errorf("error creating worker deployment version: %w; to set its compute config, use `temporal worker deployment version compute-config aws-lambda update`", err)
	}
	if err != nil {
		return fmt.Errorf("error creating worker deployment version: %w", err)
	}
	cctx.Printer.Println("Successfully created worker deployment version")
	return nil
}

func (c *TemporalWorkerDeploymentVersionComputeConfigAwsLambdaUpdateCommand) run(cctx *CommandContext, args []string) error {
	update, err := c.scalingGroupUpdate()
	if err != nil {
		return err
	}
	worker := c.Parent.Parent.Parent.Parent.Parent
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().UpdateWorkerDeploymentVersionComputeConfig(cctx, &workflowservice.UpdateWorkerDeploymentVersionComputeConfigRequest{
		Namespace:         worker.Namespace,
		DeploymentVersion: c.deploymentVersion(),
		ComputeConfigScalingGroups: map[string]*computepb.ComputeConfigScalingGroupUpdate{
			c.ScalingGroup: update,
		},
		Identity:  worker.Identity,
		RequestId: uuid.NewString(),
	})
	var notFound *serviceerror.NotFound
	if errors.As(err, &notFound) {
		return fmt.Errorf("error updating worker deployment version compute config: %w; to create it, use `temporal worker deployment version compute-config aws-lambda create`", err)
	}
	if err != nil {
		return fmt.Errorf("error updating worker deployment version compute config: %w", err)
	}
	cctx.Printer.Println("Successfully updated worker deployment version compute config")
	return nil
}

// The AWS Lambda compute provider, and the worker controller's no-sync scaler
// that starts its Workers by invoking the function.
const (
	awsLambdaProviderType = "aws-lambda"
	noSyncScalerType      = "no-sync"
)

// awsLambdaScalingGroup builds an AWS Lambda scaling group from the command's
// flags.
func awsLambdaScalingGroup(
	functionArn string,
	group *ScalingGroupOptions,
	role *AwsAssumeRoleOptions,
	scaler *NoSyncScalerOptions,
) (*computepb.ComputeConfigScalingGroup, error) {
	provider, err := awsLambdaProvider(functionArn, role)
	if err != nil {
		return nil, err
	}
	scalerDetails, err := scaler.details()
	if err != nil {
		return nil, err
	}
	sg, err := group.scalingGroup()
	if err != nil {
		return nil, err
	}
	sg.Provider = provider
	sg.Scaler = &computepb.ComputeScaler{Type: noSyncScalerType, Details: scalerDetails}
	return sg, nil
}

// awsLambdaProvider builds the AWS Lambda compute provider from the provider
// flags. Provider config keys:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/compute_provider/aws_lambda.go
func awsLambdaProvider(functionArn string, role *AwsAssumeRoleOptions) (*computepb.ComputeProvider, error) {
	if functionArn == "" {
		return nil, fmt.Errorf("--function-arn cannot be empty")
	}
	details := map[string]any{"arn": functionArn}
	if err := role.setDetails(details); err != nil {
		return nil, err
	}
	payload, err := converter.GetDefaultDataConverter().ToPayload(details)
	if err != nil {
		return nil, err
	}
	return &computepb.ComputeProvider{Type: awsLambdaProviderType, Details: payload}, nil
}

// scalingGroupUpdate builds the update of one AWS Lambda scaling group from the
// flags that are passed. Only those settings change: the provider flags, which
// are passed together; the scaling flags on their own; and --task-queue-type.
// Setting the provider also resets the scaling settings to the ones passed with
// it, because the scaler of the previous provider may not accept them. A new
// scaling group is created from the flags that are passed, since the server
// ignores the update mask for it.
func (c *TemporalWorkerDeploymentVersionComputeConfigAwsLambdaUpdateCommand) scalingGroupUpdate() (*computepb.ComputeConfigScalingGroupUpdate, error) {
	flags := c.Command.Flags()
	providerSet := slices.ContainsFunc(
		[]string{"function-arn", "assume-role-arn", "assume-role-external-id", "skip-role-and-external-id"},
		flags.Changed)
	scalerDetails, err := c.NoSyncScalerOptions.details()
	if err != nil {
		return nil, err
	}
	sg, err := c.ScalingGroupOptions.scalingGroup()
	if err != nil {
		return nil, err
	}
	var paths []string
	if flags.Changed("task-queue-type") {
		paths = append(paths, "task_queue_types")
	} else {
		sg.TaskQueueTypes = nil
	}
	switch {
	case providerSet:
		if !flags.Changed("function-arn") {
			return nil, fmt.Errorf("--function-arn is required when changing the provider")
		}
		if sg.Provider, err = awsLambdaProvider(c.FunctionArn, &c.AwsAssumeRoleOptions); err != nil {
			return nil, err
		}
		sg.Scaler = &computepb.ComputeScaler{Type: noSyncScalerType, Details: scalerDetails}
		paths = append(paths, "provider.type", "provider.details", "scaler.type", "scaler.details")
	case scalerDetails != nil:
		sg.Scaler = &computepb.ComputeScaler{Details: scalerDetails}
		paths = append(paths, "scaler.details")
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no compute configuration provided to update")
	}
	return &computepb.ComputeConfigScalingGroupUpdate{
		ScalingGroup: sg,
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: paths},
	}, nil
}

// setDetails sets the role that the server assumes in an AWS compute provider's
// details. The AWS compute providers read it from the same config keys.
func (o *AwsAssumeRoleOptions) setDetails(details map[string]any) error {
	if o.SkipRoleAndExternalId {
		if o.AssumeRoleArn != "" || o.AssumeRoleExternalId != "" {
			return fmt.Errorf("--assume-role-arn and --assume-role-external-id must be omitted when --skip-role-and-external-id is passed")
		}
		return nil
	}
	if o.AssumeRoleArn == "" || o.AssumeRoleExternalId == "" {
		return fmt.Errorf("--assume-role-arn and --assume-role-external-id are required unless --skip-role-and-external-id is passed")
	}
	details["role"] = o.AssumeRoleArn
	details["role_external_id"] = o.AssumeRoleExternalId
	return nil
}

// details builds the details of the worker controller's no-sync scaler, used by
// compute providers that start Workers by invoking them, such as AWS Lambda,
// from the scaling flags that are set. Each flag is named after the config key
// it sets, with dashes for underscores. It returns nil when none are set, so the
// server defaults apply. The settings are validated by the server. Config keys:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/scaling_algorithm/no_sync_match.go
func (o *NoSyncScalerOptions) details() (*commonpb.Payload, error) {
	details := map[string]any{}
	for flag, value := range map[string]any{
		"scale-up-backlog-threshold":                o.ScaleUpBacklogThreshold,
		"scale-up-cooloff-ms":                       o.ScaleUpCooloffMs,
		"max-worker-lifetime-ms":                    o.MaxWorkerLifetimeMs,
		"metrics-poll-interval-ms":                  o.MetricsPollIntervalMs,
		"scale-up-dispatch-rate-epsilon":            o.ScaleUpDispatchRateEpsilon,
		"scale-up-dispatch-rate-epsilon-confirm-ms": o.ScaleUpDispatchRateEpsilonConfirmMs,
		"suppress-scale-up-ms":                      o.SuppressScaleUpMs,
		"suppress-poll-interval-ms":                 o.SuppressPollIntervalMs,
	} {
		if o.FlagSet.Changed(flag) {
			details[strings.ReplaceAll(flag, "-", "_")] = value
		}
	}
	if len(details) == 0 {
		return nil, nil
	}
	return converter.GetDefaultDataConverter().ToPayload(details)
}

type formattedVersionDrainageInfo struct {
	DrainageStatus  string    `json:"drainageStatus"`
	LastChangedTime time.Time `json:"lastChangedTime"`
	LastCheckedTime time.Time `json:"lastCheckedTime"`
}

type formattedVersionTaskQueueRow struct {
	Name               string                                   `json:"name"`
	Type               string                                   `json:"type"`
	Stats              *formattedVersionTaskQueueStats          `json:"stats,omitempty"`
	StatsByPriorityKey map[int32]formattedVersionTaskQueueStats `json:"statsByPriorityKey,omitempty"`
}

type formattedVersionTaskQueueStats struct {
	ApproximateBacklogCount int64         `json:"approximateBacklogCount"`
	ApproximateBacklogAge   time.Duration `json:"approximateBacklogAge"`
	BacklogIncreaseRate     float32       `json:"backlogIncreaseRate"`
	TasksAddRate            float32       `json:"tasksAddRate"`
	TasksDispatchRate       float32       `json:"tasksDispatchRate"`
}

// Text display types for task queue info (flattened stats)
type versionTaskQueueDisplayRow struct {
	Name string
	Type string
}

type versionTaskQueueStatsDisplayRow struct {
	Name                    string
	Type                    string
	ApproximateBacklogCount int64   `cli:",align=right"`
	ApproximateBacklogAge   string  `cli:",align=right"`
	BacklogIncreaseRate     float32 `cli:",align=right"`
	TasksAddRate            float32 `cli:",align=right"`
	TasksDispatchRate       float32 `cli:",align=right"`
}

type versionPriorityStatsDisplayRow struct {
	Priority                int32   `cli:",align=right"`
	ApproximateBacklogCount int64   `cli:",align=right"`
	ApproximateBacklogAge   string  `cli:",align=right"`
	BacklogIncreaseRate     float32 `cli:",align=right"`
	TasksAddRate            float32 `cli:",align=right"`
	TasksDispatchRate       float32 `cli:",align=right"`
}

type formattedVersionDescription struct {
	DeploymentName     string                         `json:"deploymentName"`
	BuildID            string                         `json:"BuildID"`
	CreateTime         time.Time                      `json:"createTime"`
	RoutingChangedTime time.Time                      `json:"routingChangedTime"`
	CurrentSinceTime   time.Time                      `json:"currentSinceTime"`
	RampingSinceTime   time.Time                      `json:"rampingSinceTime"`
	RampPercentage     float32                        `json:"rampPercentage"`
	DrainageInfo       formattedVersionDrainageInfo   `json:"drainageInfo"`
	TaskQueuesInfos    []formattedVersionTaskQueueRow `json:"taskQueuesInfos"`
	Metadata           map[string]*commonpb.Payload   `json:"metadata"`
	ComputeConfig      *formattedVersionComputeConfig `json:"computeConfig,omitempty"`
}

type formattedVersionComputeConfig struct {
	ScalingGroups map[string]formattedVersionComputeConfigScalingGroup `json:"scalingGroups"`
}

type formattedVersionComputeConfigScalingGroup struct {
	Provider *formattedVersionComputeProvider `json:"provider,omitempty"`
	Scaler   *formattedVersionComputeScaler   `json:"scaler,omitempty"`
}

type formattedVersionComputeProvider struct {
	Type string `json:"type"`
}

type formattedVersionComputeScaler struct {
	Type                   string   `json:"type"`
	MinInstances           *int64   `json:"minInstances,omitempty"`
	MaxInstances           *int64   `json:"maxInstances,omitempty"`
	InitialInstances       *int64   `json:"initialInstances,omitempty"`
	UtilizationTarget      *float64 `json:"utilizationTarget,omitempty"`
	ScaleDownStabilization string   `json:"scaleDownStabilization,omitempty"`
}

// formattedComputeConfigDescription is the output of compute-config describe:
// every scaling group of the Version.
type formattedComputeConfigDescription struct {
	ScalingGroups map[string]formattedScalingGroupDescription `json:"scalingGroups"`
}

type formattedScalingGroupDescription struct {
	TaskQueueTypes []string                         `json:"taskQueueTypes,omitempty"`
	Provider       *formattedVersionComputeProvider `json:"provider,omitempty"`
	Scaler         *formattedScalerDescription      `json:"scaler,omitempty"`
}

// formattedScalerDescription is a scaler with the no-sync settings, which tune
// AWS Lambda scaling groups.
type formattedScalerDescription struct {
	Type                                string   `json:"type"`
	ScaleUpBacklogThreshold             *int64   `json:"scaleUpBacklogThreshold,omitempty"`
	ScaleUpCooloffMs                    *int64   `json:"scaleUpCooloffMs,omitempty"`
	MaxWorkerLifetimeMs                 *int64   `json:"maxWorkerLifetimeMs,omitempty"`
	MetricsPollIntervalMs               *int64   `json:"metricsPollIntervalMs,omitempty"`
	ScaleUpDispatchRateEpsilon          *float64 `json:"scaleUpDispatchRateEpsilon,omitempty"`
	ScaleUpDispatchRateEpsilonConfirmMs *int64   `json:"scaleUpDispatchRateEpsilonConfirmMs,omitempty"`
	SuppressScaleUpMs                   *int64   `json:"suppressScaleUpMs,omitempty"`
	SuppressPollIntervalMs              *int64   `json:"suppressPollIntervalMs,omitempty"`
}

func formatComputeConfigDescription(cc *computepb.ComputeConfig) formattedComputeConfigDescription {
	out := formattedComputeConfigDescription{ScalingGroups: map[string]formattedScalingGroupDescription{}}
	for name, msgSG := range cc.GetScalingGroups() {
		var sg formattedScalingGroupDescription
		for _, t := range msgSG.GetTaskQueueTypes() {
			str, err := versionTaskQueueTypeStr(t)
			if err != nil {
				str = fmt.Sprintf("%d", t)
			}
			sg.TaskQueueTypes = append(sg.TaskQueueTypes, str)
		}
		if p := msgSG.GetProvider(); p != nil {
			sg.Provider = &formattedVersionComputeProvider{Type: p.GetType()}
		}
		if s := msgSG.GetScaler(); s != nil {
			sg.Scaler = &formattedScalerDescription{Type: s.GetType()}
			decodeNoSyncScalerSettings(s, sg.Scaler)
		}
		out.ScalingGroups[name] = sg
	}
	return out
}

// decodeNoSyncScalerSettings sets the no-sync scaler settings that are present
// in the scaler details. Absent settings, or details that cannot be decoded,
// leave the fields unset.
func decodeNoSyncScalerSettings(s *computepb.ComputeScaler, fs *formattedScalerDescription) {
	if s.GetType() != noSyncScalerType || s.GetDetails() == nil {
		return
	}
	var m map[string]any
	if err := converter.GetDefaultDataConverter().FromPayload(s.GetDetails(), &m); err != nil {
		return
	}
	count := func(key string) *int64 {
		if v, ok := versionScalerCount(m, key); ok {
			return &v
		}
		return nil
	}
	fs.ScaleUpBacklogThreshold = count("scale_up_backlog_threshold")
	fs.ScaleUpCooloffMs = count("scale_up_cooloff_ms")
	fs.MaxWorkerLifetimeMs = count("max_worker_lifetime_ms")
	fs.MetricsPollIntervalMs = count("metrics_poll_interval_ms")
	if v, ok := versionScalerFloat(m, "scale_up_dispatch_rate_epsilon"); ok {
		fs.ScaleUpDispatchRateEpsilon = &v
	}
	fs.ScaleUpDispatchRateEpsilonConfirmMs = count("scale_up_dispatch_rate_epsilon_confirm_ms")
	fs.SuppressScaleUpMs = count("suppress_scale_up_ms")
	fs.SuppressPollIntervalMs = count("suppress_poll_interval_ms")
}

// These functions convert gRPC proto types directly, avoiding SDK type dependencies.

func versionDrainageStatusStr(status enumspb.VersionDrainageStatus) (string, error) {
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

func versionTaskQueueTypeStr(taskQueueType enumspb.TaskQueueType) (string, error) {
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

func formatVersionTaskQueueStats(tqStats *taskqueuepb.TaskQueueStats) formattedVersionTaskQueueStats {
	if tqStats == nil {
		return formattedVersionTaskQueueStats{}
	}
	return formattedVersionTaskQueueStats{
		ApproximateBacklogCount: tqStats.ApproximateBacklogCount,
		ApproximateBacklogAge:   tqStats.ApproximateBacklogAge.AsDuration(),
		// BacklogIncreaseRate is computed as TasksAddRate - TasksDispatchRate (same as SDK)
		BacklogIncreaseRate: tqStats.TasksAddRate - tqStats.TasksDispatchRate,
		TasksAddRate:        tqStats.TasksAddRate,
		TasksDispatchRate:   tqStats.TasksDispatchRate,
	}
}

func formatVersionTaskQueues(tqis []*workflowservice.DescribeWorkerDeploymentVersionResponse_VersionTaskQueue, includeStats bool) ([]formattedVersionTaskQueueRow, error) {
	var tqiRows []formattedVersionTaskQueueRow
	for _, tqi := range tqis {
		tqTypeStr, err := versionTaskQueueTypeStr(tqi.GetType())
		if err != nil {
			return tqiRows, err
		}

		row := formattedVersionTaskQueueRow{
			Name: tqi.GetName(),
			Type: tqTypeStr,
		}

		if includeStats {
			fVersionStats := formatVersionTaskQueueStats(tqi.GetStats())
			row.Stats = &fVersionStats

			if len(tqi.GetStatsByPriorityKey()) > 0 {
				fVersionStatsByPriorityKey := map[int32]formattedVersionTaskQueueStats{}
				for k, v := range tqi.GetStatsByPriorityKey() {
					fVersionStatsByPriorityKey[k] = formatVersionTaskQueueStats(v)
				}
				row.StatsByPriorityKey = fVersionStatsByPriorityKey
			}
		}

		tqiRows = append(tqiRows, row)
	}
	return tqiRows, nil
}

func formatVersionDrainageInfo(drainageInfo *deploymentpb.VersionDrainageInfo) (formattedVersionDrainageInfo, error) {
	if drainageInfo == nil {
		return formattedVersionDrainageInfo{}, nil
	}

	drainageStr, err := versionDrainageStatusStr(drainageInfo.GetStatus())
	if err != nil {
		return formattedVersionDrainageInfo{}, err
	}

	return formattedVersionDrainageInfo{
		DrainageStatus:  drainageStr,
		LastChangedTime: drainageInfo.GetLastChangedTime().AsTime(),
		LastCheckedTime: drainageInfo.GetLastCheckedTime().AsTime(),
	}, nil
}

// Config keys of the worker controller's rate-based scaler, decoded for display.
// Mirrors:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/scaling_algorithm/rate_based.go
const (
	versionScalerKeyMinCount          = "min_count"
	versionScalerKeyMaxCount          = "max_count"
	versionScalerKeyInitialCount      = "initial_count"
	versionScalerKeyUtilizationTarget = "utilization_target"
	versionScalerKeyNoSyncQuietMs     = "no_sync_quiet_ms"
)

// versionScalerCount reads an integer worker-count value from a decoded scaler
// details map. The default data converter round-trips JSON numbers as float64,
// so that case is handled alongside the native int types.
func versionScalerCount(m map[string]any, key string) (int64, bool) {
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

// versionScalerFloat reads a fractional value (e.g. utilization_target) from a
// decoded scaler details map. JSON numbers decode as float64.
func versionScalerFloat(m map[string]any, key string) (float64, bool) {
	switch n := m[key].(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

// versionScalerSettings holds the rate-based scaler settings surfaced for display. Each
// field is nil when the corresponding key is absent from the scaler details.
type versionScalerSettings struct {
	minInstances           *int64
	maxInstances           *int64
	initialInstances       *int64
	utilizationTarget      *float64
	scaleDownStabilization *time.Duration
}

// decodeVersionScalerSettings extracts the rate-based scaler settings from a
// ComputeScaler's details payload for display. Best-effort: returns zero-value
// (all-nil) settings when the scaler, its details, or a key is absent, or when
// the payload cannot be decoded, so read paths never fail on an unexpected shape.
func decodeVersionScalerSettings(s *computepb.ComputeScaler) versionScalerSettings {
	var out versionScalerSettings
	details := s.GetDetails()
	if details == nil {
		return out
	}
	var m map[string]any
	if err := converter.GetDefaultDataConverter().FromPayload(details, &m); err != nil {
		return out
	}
	if v, ok := versionScalerCount(m, versionScalerKeyMinCount); ok {
		out.minInstances = &v
	}
	if v, ok := versionScalerCount(m, versionScalerKeyMaxCount); ok {
		out.maxInstances = &v
	}
	if v, ok := versionScalerCount(m, versionScalerKeyInitialCount); ok {
		out.initialInstances = &v
	}
	if v, ok := versionScalerFloat(m, versionScalerKeyUtilizationTarget); ok {
		out.utilizationTarget = &v
	}
	if v, ok := versionScalerCount(m, versionScalerKeyNoSyncQuietMs); ok {
		// The WCI key is stored in milliseconds; surface it as a duration to match
		// how the CLI renders other duration fields (e.g. ApproximateBacklogAge).
		d := time.Duration(v) * time.Millisecond
		out.scaleDownStabilization = &d
	}
	return out
}

func formatVersionComputeConfig(cc *computepb.ComputeConfig) *formattedVersionComputeConfig {
	if cc == nil {
		return nil
	}
	msgSGs := cc.GetScalingGroups()
	if len(msgSGs) == 0 {
		return nil
	}
	sgs := make(map[string]formattedVersionComputeConfigScalingGroup, len(msgSGs))
	for name, msgSG := range msgSGs {
		p := msgSG.GetProvider()
		s := msgSG.GetScaler()
		if p == nil && s == nil {
			continue
		}
		sg := formattedVersionComputeConfigScalingGroup{}
		if p != nil {
			sg.Provider = &formattedVersionComputeProvider{
				Type: p.GetType(),
			}
		}
		if s != nil {
			fs := &formattedVersionComputeScaler{Type: s.GetType()}
			set := decodeVersionScalerSettings(s)
			fs.MinInstances = set.minInstances
			fs.MaxInstances = set.maxInstances
			fs.InitialInstances = set.initialInstances
			fs.UtilizationTarget = set.utilizationTarget
			if set.scaleDownStabilization != nil {
				fs.ScaleDownStabilization = formatDuration(*set.scaleDownStabilization)
			}
			sg.Scaler = fs
		}
		sgs[name] = sg
	}
	return &formattedVersionComputeConfig{
		ScalingGroups: sgs,
	}
}

// formatVersionDescription converts gRPC proto types to formatted types for display.
func formatVersionDescription(deploymentInfo *deploymentpb.WorkerDeploymentVersionInfo, taskQueueInfos []*workflowservice.DescribeWorkerDeploymentVersionResponse_VersionTaskQueue, includeStats bool) (formattedVersionDescription, error) {
	tqi, err := formatVersionTaskQueues(taskQueueInfos, includeStats)
	if err != nil {
		return formattedVersionDescription{}, err
	}

	drainage, err := formatVersionDrainageInfo(deploymentInfo.GetDrainageInfo())
	if err != nil {
		return formattedVersionDescription{}, err
	}

	computeConfig := formatVersionComputeConfig(deploymentInfo.GetComputeConfig())

	return formattedVersionDescription{
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

func versionComputeConfigSummary(cc *computepb.ComputeConfig) string {
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
		// reflects the configured limits (ordered min, initial, max, utilization,
		// scale-down-stabilization).
		set := decodeVersionScalerSettings(sg.GetScaler())
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
		if set.scaleDownStabilization != nil {
			parts = append(parts, fmt.Sprintf("scale-down-stabilization %s", formatDuration(*set.scaleDownStabilization)))
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

// printVersionDescription prints worker deployment version info from proto types.
func printVersionDescription(cctx *CommandContext, deploymentInfo *deploymentpb.WorkerDeploymentVersionInfo, taskQueueInfos []*workflowservice.DescribeWorkerDeploymentVersionResponse_VersionTaskQueue, msg string, opts versionDescribeOptions) error {
	fDeploymentInfo, err := formatVersionDescription(deploymentInfo, taskQueueInfos, opts.showStats)
	if err != nil {
		return err
	}

	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString(msg))
		var drainageStr string
		var drainageLastChangedTime time.Time
		var drainageLastCheckedTime time.Time
		if deploymentInfo.GetDrainageInfo() != nil {
			drainageStr, err = versionDrainageStatusStr(deploymentInfo.GetDrainageInfo().GetStatus())
			if err != nil {
				return err
			}
			drainageLastChangedTime = deploymentInfo.GetDrainageInfo().GetLastChangedTime().AsTime()
			drainageLastCheckedTime = deploymentInfo.GetDrainageInfo().GetLastCheckedTime().AsTime()
		}
		computeConfigSummary := versionComputeConfigSummary(deploymentInfo.GetComputeConfig())

		printMe := struct {
			DeploymentName          string
			BuildID                 string
			CreateTime              time.Time
			RoutingChangedTime      time.Time `cli:",cardOmitEmpty"`
			CurrentSinceTime        time.Time `cli:",cardOmitEmpty"`
			RampingSinceTime        time.Time `cli:",cardOmitEmpty"`
			RampPercentage          float32
			DrainageStatus          string                       `cli:",cardOmitEmpty"`
			DrainageLastChangedTime time.Time                    `cli:",cardOmitEmpty"`
			DrainageLastCheckedTime time.Time                    `cli:",cardOmitEmpty"`
			Metadata                map[string]*commonpb.Payload `cli:",cardOmitEmpty"`
			ComputeConfigSummary    string                       `cli:",cardOmitEmpty"`
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
			if err := printVersionTaskQueues(cctx, fDeploymentInfo.TaskQueuesInfos, opts); err != nil {
				return err
			}
		}

		return nil
	}

	// json output
	return cctx.Printer.PrintStructured(fDeploymentInfo, printer.StructuredOptions{})
}

type versionDescribeOptions struct {
	showStats bool
}

func formatVersionDurationShort(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	return d.Truncate(time.Millisecond).String()
}

func printVersionTaskQueues(cctx *CommandContext, taskQueues []formattedVersionTaskQueueRow, opts versionDescribeOptions) error {
	cctx.Printer.Println()
	cctx.Printer.Println(color.MagentaString("Task Queues:"))

	if opts.showStats {
		// Show flattened stats in the table
		rows := make([]versionTaskQueueStatsDisplayRow, 0, len(taskQueues))
		for _, tq := range taskQueues {
			row := versionTaskQueueStatsDisplayRow{
				Name: tq.Name,
				Type: tq.Type,
			}
			if tq.Stats != nil {
				row.ApproximateBacklogCount = tq.Stats.ApproximateBacklogCount
				row.ApproximateBacklogAge = formatVersionDurationShort(tq.Stats.ApproximateBacklogAge)
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
			if !versionHasNonDefaultPriorityKeys(tq.StatsByPriorityKey) {
				continue
			}
			cctx.Printer.Println()
			cctx.Printer.Println(color.MagentaString(fmt.Sprintf("Stats by Priority (%s / %s):", tq.Name, tq.Type)))

			// Sort priority keys for consistent output
			priorities := make([]int32, 0, len(tq.StatsByPriorityKey))
			for p := range tq.StatsByPriorityKey {
				priorities = append(priorities, p)
			}
			sortVersionPriorities(priorities)

			priorityRows := make([]versionPriorityStatsDisplayRow, 0, len(priorities))
			for _, p := range priorities {
				stats := tq.StatsByPriorityKey[p]
				priorityRows = append(priorityRows, versionPriorityStatsDisplayRow{
					Priority:                p,
					ApproximateBacklogCount: stats.ApproximateBacklogCount,
					ApproximateBacklogAge:   formatVersionDurationShort(stats.ApproximateBacklogAge),
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
		rows := make([]versionTaskQueueDisplayRow, 0, len(taskQueues))
		for _, tq := range taskQueues {
			rows = append(rows, versionTaskQueueDisplayRow{
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

func sortVersionPriorities(s []int32) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// versionDefaultPriorityKey is the default priority key value. When this is the only
// priority key present, we skip showing per-priority stats as it would be redundant.
const versionDefaultPriorityKey = 3

// versionHasNonDefaultPriorityKeys returns true if the map contains any priority keys
// other than the default (3), or contains multiple priority keys.
func versionHasNonDefaultPriorityKeys(statsByPriorityKey map[int32]formattedVersionTaskQueueStats) bool {
	if len(statsByPriorityKey) == 0 {
		return false
	}
	if len(statsByPriorityKey) > 1 {
		return true
	}
	// Exactly one key - check if it's the default
	_, hasDefault := statsByPriorityKey[versionDefaultPriorityKey]
	return !hasDefault
}
