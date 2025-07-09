package temporalcli

import (
	"fmt"

	deploymentpb "go.temporal.io/api/deployment/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func (c *TemporalWorkflowResetWithWorkflowUpdateOptionsCommand) run(cctx *CommandContext, args []string) error {
	validate, _ := c.Parent.getResetOperations()
	if err := validate(); err != nil {
		return err
	}

	if c.VersioningOverrideBehavior.Value == "pinned" {
		if c.VersioningOverrideDeploymentName == "" || c.VersioningOverrideBuildId == "" {
			return fmt.Errorf("deployment name and build id are required with 'pinned' behavior")
		}
	}
	if c.VersioningOverrideBehavior.Value != "pinned" {
		if c.VersioningOverrideDeploymentName != "" || c.VersioningOverrideBuildId != "" {
			return fmt.Errorf("cannot set deployment name or build id with %v behavior", c.VersioningOverrideBehavior.Value)
		}
	}

	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	VersioningOverride := &workflowpb.VersioningOverride{}
	switch c.VersioningOverrideBehavior.Value {
	case "pinned":
		VersioningOverride.Override = &workflowpb.VersioningOverride_Pinned{
			Pinned: &workflowpb.VersioningOverride_PinnedOverride{
				Behavior: workflowpb.VersioningOverride_PINNED_OVERRIDE_BEHAVIOR_PINNED,
				Version: &deploymentpb.WorkerDeploymentVersion{
					DeploymentName: c.VersioningOverrideDeploymentName,
					BuildId:        c.VersioningOverrideBuildId,
				},
			},
		}
	case "auto_upgrade":
		VersioningOverride.Override = &workflowpb.VersioningOverride_AutoUpgrade{
			AutoUpgrade: true,
		}
	default:
		return fmt.Errorf("invalid deployment behavior: %v, valid values are: 'pinned', and 'auto_upgrade'", c.VersioningOverrideBehavior.Value)
	}

	var workflowExecutionOptions *workflowpb.WorkflowExecutionOptions
	protoMask, err := fieldmaskpb.New(workflowExecutionOptions, "versioning_override")
	if err != nil {
		return fmt.Errorf("invalid field mask: %w", err)
	}

	postOp := &workflowpb.PostResetOperation{
		Variant: &workflowpb.PostResetOperation_UpdateWorkflowOptions_{
			UpdateWorkflowOptions: &workflowpb.PostResetOperation_UpdateWorkflowOptions{
				WorkflowExecutionOptions: &workflowpb.WorkflowExecutionOptions{
					VersioningOverride: VersioningOverride,
				},
				UpdateMask: protoMask,
			},
		},
	}

	if c.Parent.WorkflowId != "" {
		return c.Parent.doWorkflowResetWithPostOps(cctx, cl, []*workflowpb.PostResetOperation{postOp})
	}
	return c.Parent.runBatchResetWithPostOps(cctx, cl, []*workflowpb.PostResetOperation{postOp})
}
