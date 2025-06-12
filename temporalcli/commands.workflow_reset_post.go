package temporalcli

// import (
// 	"fmt"

// 	workflowpb "go.temporal.io/api/workflow/v1"
// 	"google.golang.org/protobuf/types/known/fieldmaskpb"
// )

// func (c *TemporalWorkflowResetWithWorkflowSignalCommand) run(cctx *CommandContext, _ []string) error {
// 	validate, _ := c.Parent.getResetOperations()
// 	if err := validate(); err != nil {
// 		return err
// 	}
// 	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer cl.Close()
// 	input, err := c.PayloadInputOptions.buildRawInputPayloads()
// 	if err != nil {
// 		return err
// 	}
// 	op := &workflowpb.PostResetOperation{
// 		Variant: &workflowpb.PostResetOperation_SignalWorkflow_{
// 			SignalWorkflow: &workflowpb.PostResetOperation_SignalWorkflow{
// 				SignalName: c.Name,
// 				Input:      input,
// 			},
// 		},
// 	}
// 	if c.Parent.WorkflowId != "" {
// 		return c.Parent.doWorkflowResetWithOperations(cctx, cl, []*workflowpb.PostResetOperation{op})
// 	}
// 	return c.Parent.runBatchResetWithOperations(cctx, cl, []*workflowpb.PostResetOperation{op})
// }

// func (c *TemporalWorkflowResetWithWorkflowUpdateOptionsCommand) run(cctx *CommandContext, _ []string) error {
// 	if c.VersioningOverrideBehavior.Value == "pinned" && c.VersioningOverridePinnedVersion == "" {
// 		return fmt.Errorf("missing version with 'pinned' behavior")
// 	}
// 	if (c.VersioningOverrideBehavior.Value == "unspecified" || c.VersioningOverrideBehavior.Value == "auto_upgrade") && c.VersioningOverridePinnedVersion != "" {
// 		return fmt.Errorf("cannot set pinned version with %v behavior", c.VersioningOverrideBehavior)
// 	}
// 	validate, _ := c.Parent.getResetOperations()
// 	if err := validate(); err != nil {
// 		return err
// 	}
// 	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer cl.Close()

// 	behavior := workflowpb.VersioningBehavior_VERSIONING_BEHAVIOR_UNSPECIFIED
// 	switch c.VersioningOverrideBehavior.Value {
// 	case "unspecified":
// 	case "pinned":
// 		behavior = workflowpb.VersioningBehavior_VERSIONING_BEHAVIOR_PINNED
// 	case "auto_upgrade":
// 		behavior = workflowpb.VersioningBehavior_VERSIONING_BEHAVIOR_AUTO_UPGRADE
// 	default:
// 		return fmt.Errorf("invalid deployment behavior: %v, valid values are: 'unspecified', 'pinned', and 'auto_upgrade'", c.VersioningOverrideBehavior)
// 	}
// 	mask, err := fieldmaskpb.New(&workflowpb.WorkflowExecutionOptions{}, "versioning_override")
// 	if err != nil {
// 		return fmt.Errorf("invalid field mask: %w", err)
// 	}
// 	op := &workflowpb.PostResetOperation{
// 		Variant: &workflowpb.PostResetOperation_UpdateWorkflowOptions_{
// 			UpdateWorkflowOptions: &workflowpb.PostResetOperation_UpdateWorkflowOptions{
// 				WorkflowExecutionOptions: &workflowpb.WorkflowExecutionOptions{
// 					VersioningOverride: &workflowpb.VersioningOverride{
// 						Behavior:      behavior,
// 						PinnedVersion: c.VersioningOverridePinnedVersion,
// 					},
// 				},
// 				UpdateMask: mask,
// 			},
// 		},
// 	}
// 	if c.Parent.WorkflowId != "" {
// 		return c.Parent.doWorkflowResetWithOperations(cctx, cl, []*workflowpb.PostResetOperation{op})
// 	}
// 	return c.Parent.runBatchResetWithOperations(cctx, cl, []*workflowpb.PostResetOperation{op})
// }
