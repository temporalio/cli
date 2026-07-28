package temporalcli

import (
	"fmt"

	workflowpb "go.temporal.io/api/workflow/v1"
)

func (c *TemporalWorkflowResetWithWorkflowUpdateOptionsCommand) run(cctx *CommandContext, args []string) error {
	validate, _ := c.Parent.getResetOperations()
	if err := validate(); err != nil {
		return err
	}

	if c.VersioningOverrideBehavior.Value == "pinned" ||
		c.VersioningOverrideBehavior.Value == "one_time" {
		if c.VersioningOverrideDeploymentName == "" || c.VersioningOverrideBuildId == "" {
			return fmt.Errorf("deployment name and build id are required with '%s' behavior", c.VersioningOverrideBehavior.Value)
		}
	}
	if c.VersioningOverrideBehavior.Value != "pinned" &&
		c.VersioningOverrideBehavior.Value != "one_time" {
		if c.VersioningOverrideDeploymentName != "" || c.VersioningOverrideBuildId != "" {
			return fmt.Errorf("cannot set deployment name or build id with %v behavior", c.VersioningOverrideBehavior.Value)
		}
	}

	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	var versioningOverride *workflowpb.VersioningOverride
	switch c.VersioningOverrideBehavior.Value {
	case "pinned":
		versioningOverride = &workflowpb.VersioningOverride{
			Override: &workflowpb.VersioningOverride_Pinned{
				Pinned: &workflowpb.VersioningOverride_PinnedOverride{
					Behavior: workflowpb.VersioningOverride_PINNED_OVERRIDE_BEHAVIOR_PINNED,
					Version:  workerDeploymentVersionToProto(c.VersioningOverrideDeploymentName, c.VersioningOverrideBuildId),
				},
			},
		}
	case "one_time":
		versioningOverride = oneTimeVersioningOverrideToProto(c.VersioningOverrideDeploymentName, c.VersioningOverrideBuildId)
	case "auto_upgrade":
		versioningOverride = &workflowpb.VersioningOverride{
			Override: &workflowpb.VersioningOverride_AutoUpgrade{
				AutoUpgrade: true,
			},
		}
	default:
		return fmt.Errorf("invalid deployment behavior: %v, valid values are: 'pinned', 'auto_upgrade', and 'one_time'", c.VersioningOverrideBehavior.Value)
	}

	workflowExecutionOptions, protoMask, err := workflowExecutionOptionsForVersioningOverride(versioningOverride)
	if err != nil {
		return fmt.Errorf("invalid field mask: %w", err)
	}

	postOp := &workflowpb.PostResetOperation{
		Variant: &workflowpb.PostResetOperation_UpdateWorkflowOptions_{
			UpdateWorkflowOptions: &workflowpb.PostResetOperation_UpdateWorkflowOptions{
				WorkflowExecutionOptions: workflowExecutionOptions,
				UpdateMask:               protoMask,
			},
		},
	}

	if c.Parent.WorkflowId != "" {
		return c.Parent.doWorkflowResetWithPostOps(cctx, cl, []*workflowpb.PostResetOperation{postOp})
	}
	return c.Parent.runBatchResetWithPostOps(cctx, cl, []*workflowpb.PostResetOperation{postOp})
}
