package temporalcli

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/temporalio/cli/internal/printer"
	commonpb "go.temporal.io/api/common/v1"
	computepb "go.temporal.io/api/compute/v1"
	deploymentpb "go.temporal.io/api/deployment/v1"
	"go.temporal.io/api/serviceerror"
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

	// Use raw gRPC instead of SDK's DeploymentClient to avoid circular dependency
	// with SDK release that exposes TaskQueuesInfos in DescribeVersionResponse
	resp, err := cl.WorkflowService().DescribeWorkerDeploymentVersion(cctx, &workflowservice.DescribeWorkerDeploymentVersionRequest{
		Namespace:            worker.Namespace,
		DeploymentVersion:    c.deploymentVersion(),
		ReportTaskQueueStats: c.ReportTaskQueueStats,
	})
	if err != nil {
		return fmt.Errorf("error describing worker deployment version: %w", err)
	}
	return printWorkerDeploymentVersionInfoProto(cctx, resp.GetWorkerDeploymentVersionInfo(), resp.GetVersionTaskQueues(), "Worker Deployment Version:", printVersionInfoOptions{
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

	type scalingGroupCard struct {
		ScalingGroup string
		Provider     string `cli:",cardOmitEmpty"`
		Scaler       string `cli:",cardOmitEmpty"`
	}
	keys := slices.Sorted(maps.Keys(described.ScalingGroups))
	cards := make([]scalingGroupCard, 0, len(keys))
	for _, key := range keys {
		sg := described.ScalingGroups[key]
		card := scalingGroupCard{ScalingGroup: key}
		if sg.Provider != nil {
			card.Provider = sg.Provider.Type
		}
		if sg.Scaler != nil {
			card.Scaler = scalerSummary(sg.Scaler)
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
	removeOne := c.Command.Flags().Changed("scaling-group")
	if removeOne && c.ScalingGroup == "" {
		return fmt.Errorf("--scaling-group cannot be empty")
	}
	if removeOne && c.Yes {
		return fmt.Errorf("--yes cannot be used with --scaling-group")
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
// the --scaling-group flag. Scaling groups set from the CLI serve every Task
// Queue type.
func (o *ScalingGroupOptions) scalingGroup() (*computepb.ComputeConfigScalingGroup, error) {
	if o.ScalingGroup == "" {
		return nil, fmt.Errorf("--scaling-group cannot be empty")
	}
	return &computepb.ComputeConfigScalingGroup{}, nil
}

// update builds the provider-independent part of a scaling group update: the
// fields whose flags are passed, and their update mask paths.
func (o *ScalingGroupOptions) update() (*computepb.ComputeConfigScalingGroup, []string, error) {
	sg, err := o.scalingGroup()
	if err != nil {
		return nil, nil, err
	}
	return sg, nil, nil
}

// createVersionWithScalingGroup creates a Worker Deployment Version with one
// scaling group. provider is the compute-config subcommand of the scaling
// group's compute provider, used in the error hint.
func createVersionWithScalingGroup(
	cctx *CommandContext,
	worker *TemporalWorkerCommand,
	version *deploymentpb.WorkerDeploymentVersion,
	key string,
	sg *computepb.ComputeConfigScalingGroup,
	provider string,
) error {
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().CreateWorkerDeploymentVersion(cctx, &workflowservice.CreateWorkerDeploymentVersionRequest{
		Namespace:         worker.Namespace,
		DeploymentVersion: version,
		ComputeConfig: &computepb.ComputeConfig{
			ScalingGroups: map[string]*computepb.ComputeConfigScalingGroup{key: sg},
		},
		Identity:  worker.Identity,
		RequestId: uuid.NewString(),
	})
	var alreadyExists *serviceerror.AlreadyExists
	if errors.As(err, &alreadyExists) {
		return fmt.Errorf("error creating worker deployment version: %w; to set its compute config, use `temporal worker deployment version compute-config %s update`", err, provider)
	}
	if err != nil {
		return fmt.Errorf("error creating worker deployment version: %w", err)
	}
	cctx.Printer.Println("Successfully created worker deployment version")
	return nil
}

// updateVersionScalingGroup updates one scaling group of an existing Worker
// Deployment Version. provider is the compute-config subcommand of the scaling
// group's compute provider, used in the error hint.
func updateVersionScalingGroup(
	cctx *CommandContext,
	worker *TemporalWorkerCommand,
	version *deploymentpb.WorkerDeploymentVersion,
	key string,
	update *computepb.ComputeConfigScalingGroupUpdate,
	provider string,
) error {
	cl, err := dialClient(cctx, &worker.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	_, err = cl.WorkflowService().UpdateWorkerDeploymentVersionComputeConfig(cctx, &workflowservice.UpdateWorkerDeploymentVersionComputeConfigRequest{
		Namespace:         worker.Namespace,
		DeploymentVersion: version,
		ComputeConfigScalingGroups: map[string]*computepb.ComputeConfigScalingGroupUpdate{
			key: update,
		},
		Identity:  worker.Identity,
		RequestId: uuid.NewString(),
	})
	var notFound *serviceerror.NotFound
	if errors.As(err, &notFound) {
		return fmt.Errorf("error updating worker deployment version compute config: %w; to create it, use `temporal worker deployment version compute-config %s create`", err, provider)
	}
	if err != nil {
		return fmt.Errorf("error updating worker deployment version compute config: %w", err)
	}
	cctx.Printer.Println("Successfully updated worker deployment version compute config")
	return nil
}

// passed reports whether any of the role flags is passed.
func (o *AwsAssumeRoleOptions) passed() bool {
	return slices.ContainsFunc([]string{"assume-role-arn", "assume-role-external-id", "skip-role-and-external-id"}, o.FlagSet.Changed)
}

// The worker controller's no-sync scaler, which starts Workers by invoking them,
// such as AWS Lambda functions.
const noSyncScalerType = "no-sync"

// noSyncScalingGroup builds a scaling group for a compute provider that uses the
// no-sync scaler.
func noSyncScalingGroup(
	group *ScalingGroupOptions,
	provider *computepb.ComputeProvider,
	scalerDetails *commonpb.Payload,
) (*computepb.ComputeConfigScalingGroup, error) {
	sg, err := group.scalingGroup()
	if err != nil {
		return nil, err
	}
	sg.Provider = provider
	sg.Scaler = &computepb.ComputeScaler{Type: noSyncScalerType, Details: scalerDetails}
	return sg, nil
}

// noSyncScalingGroupUpdate builds an update of a scaling group for a compute
// provider that uses the no-sync scaler, from the flags that are passed. The
// provider is nil when its flags are not passed.
func noSyncScalingGroupUpdate(
	group *ScalingGroupOptions,
	provider *computepb.ComputeProvider,
	scalerDetails *commonpb.Payload,
) (*computepb.ComputeConfigScalingGroupUpdate, error) {
	sg, paths, err := group.update()
	if err != nil {
		return nil, err
	}
	switch {
	case provider != nil:
		// Provider (re)configuration: rebuild the provider and scaler type,
		// and keep scaler.details consistent with the (possibly new) type.
		sg.Provider = provider
		sg.Scaler = &computepb.ComputeScaler{Type: noSyncScalerType, Details: scalerDetails}
		paths = append(paths, "provider.type", "provider.details", "scaler.type", "scaler.details")
	case scalerDetails != nil:
		// Scaler-only update: change scaler.details without touching the
		// provider or scaler type.
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

type formattedComputeConfigDescription struct {
	ScalingGroups map[string]formattedScalingGroupDescription `json:"scalingGroups"`
}

type formattedScalingGroupDescription struct {
	Provider *formattedComputeConfigProvider `json:"provider,omitempty"`
	Scaler   *formattedScalerDescription     `json:"scaler,omitempty"`
}

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
		if p := msgSG.GetProvider(); p != nil {
			sg.Provider = &formattedComputeConfigProvider{Type: p.GetType()}
		}
		if s := msgSG.GetScaler(); s != nil {
			sg.Scaler = &formattedScalerDescription{Type: s.GetType()}
			decodeNoSyncScalerSettings(s, sg.Scaler)
		}
		out.ScalingGroups[name] = sg
	}
	return out
}

// scalerSummary returns a scaler's type followed by the settings that are set,
// such as "no-sync (scale_up_backlog_threshold 2, max_worker_lifetime_ms 300000)".
func scalerSummary(fs *formattedScalerDescription) string {
	var parts []string
	addInt := func(key string, v *int64) {
		if v != nil {
			parts = append(parts, fmt.Sprintf("%s %d", key, *v))
		}
	}
	addInt("scale_up_backlog_threshold", fs.ScaleUpBacklogThreshold)
	addInt("scale_up_cooloff_ms", fs.ScaleUpCooloffMs)
	addInt("max_worker_lifetime_ms", fs.MaxWorkerLifetimeMs)
	addInt("metrics_poll_interval_ms", fs.MetricsPollIntervalMs)
	if v := fs.ScaleUpDispatchRateEpsilon; v != nil {
		parts = append(parts, "scale_up_dispatch_rate_epsilon "+strconv.FormatFloat(*v, 'g', -1, 64))
	}
	addInt("scale_up_dispatch_rate_epsilon_confirm_ms", fs.ScaleUpDispatchRateEpsilonConfirmMs)
	addInt("suppress_scale_up_ms", fs.SuppressScaleUpMs)
	addInt("suppress_poll_interval_ms", fs.SuppressPollIntervalMs)
	if len(parts) == 0 {
		return fs.Type
	}
	return fs.Type + " (" + strings.Join(parts, ", ") + ")"
}

func decodeNoSyncScalerSettings(s *computepb.ComputeScaler, fs *formattedScalerDescription) {
	if s.GetType() != noSyncScalerType || s.GetDetails() == nil {
		return
	}
	var m map[string]any
	if err := converter.GetDefaultDataConverter().FromPayload(s.GetDetails(), &m); err != nil {
		return
	}
	count := func(key string) *int64 {
		if v, ok := scalerCountFromMap(m, key); ok {
			return &v
		}
		return nil
	}
	fs.ScaleUpBacklogThreshold = count("scale_up_backlog_threshold")
	fs.ScaleUpCooloffMs = count("scale_up_cooloff_ms")
	fs.MaxWorkerLifetimeMs = count("max_worker_lifetime_ms")
	fs.MetricsPollIntervalMs = count("metrics_poll_interval_ms")
	if v, ok := scalerFloatFromMap(m, "scale_up_dispatch_rate_epsilon"); ok {
		fs.ScaleUpDispatchRateEpsilon = &v
	}
	fs.ScaleUpDispatchRateEpsilonConfirmMs = count("scale_up_dispatch_rate_epsilon_confirm_ms")
	fs.SuppressScaleUpMs = count("suppress_scale_up_ms")
	fs.SuppressPollIntervalMs = count("suppress_poll_interval_ms")
}
