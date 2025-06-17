package temporalcli

import (
	"fmt"

	"go.temporal.io/api/enums/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func (c *TemporalWorkflowResetWithWorkflowUpdateOptionsCommand) run(cctx *CommandContext, args []string) error {
	validate, _ := c.Parent.getResetOperations()
	if err := validate(); err != nil {
		return err
	}

	if c.VersioningOverrideBehavior.Value == "pinned" && c.VersioningOverridePinnedVersion == "" {
		return fmt.Errorf("missing version with 'pinned' behavior")
	}
	if c.VersioningOverrideBehavior.Value != "pinned" && c.VersioningOverridePinnedVersion != "" {
		return fmt.Errorf("cannot set pinned version with %v behavior", c.VersioningOverrideBehavior)
	}

	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	var behavior enums.VersioningBehavior
	switch c.VersioningOverrideBehavior.Value {
	case "pinned":
		behavior = enums.VERSIONING_BEHAVIOR_PINNED
	case "auto_upgrade":
		behavior = enums.VERSIONING_BEHAVIOR_AUTO_UPGRADE
	default:
		return fmt.Errorf("invalid deployment behavior: %v, valid values are: 'pinned', and 'auto_upgrade'", c.VersioningOverrideBehavior)
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
					VersioningOverride: &workflowpb.VersioningOverride{
						Behavior:      behavior,
						PinnedVersion: c.VersioningOverridePinnedVersion,
					},
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
