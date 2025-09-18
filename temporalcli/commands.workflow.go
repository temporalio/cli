package temporalcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/user"

	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/batch/v1"
	"go.temporal.io/api/common/v1"
	deploymentpb "go.temporal.io/api/deployment/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/query/v1"
	sdkpb "go.temporal.io/api/sdk/v1"
	"go.temporal.io/api/update/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const metadataQueryName = "__temporal_workflow_metadata"

func (c *TemporalWorkflowCancelCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})

	// Run single or batch
	if err != nil {
		return err
	} else if exec != nil {
		err = cl.CancelWorkflow(cctx, exec.WorkflowId, exec.RunId)
		if err != nil {
			return fmt.Errorf("failed to cancel workflow: %w", err)
		}
		cctx.Printer.Println("Canceled workflow")
	} else { // batchReq != nil
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_CancellationOperation{
			CancellationOperation: &batch.BatchOperationCancellation{
				Identity: c.Parent.Identity,
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}

func (c *TemporalWorkflowDeleteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})

	// Run single or batch
	if err != nil {
		return err
	} else if exec != nil {
		_, err := cl.WorkflowService().DeleteWorkflowExecution(cctx, &workflowservice.DeleteWorkflowExecutionRequest{
			Namespace:         c.Parent.Namespace,
			WorkflowExecution: &common.WorkflowExecution{WorkflowId: c.WorkflowId, RunId: c.RunId},
		})
		if err != nil {
			return fmt.Errorf("failed to delete workflow: %w", err)
		}
		cctx.Printer.Println("Delete workflow succeeded")
	} else { // batchReq != nil
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_DeletionOperation{
			DeletionOperation: &batch.BatchOperationDeletion{
				Identity: c.Parent.Identity,
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}
	return nil
}

func (c *TemporalWorkflowUpdateOptionsCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	if c.VersioningOverrideBehavior.Value == "unspecified" ||
		c.VersioningOverrideBehavior.Value == "auto_upgrade" {
		if c.VersioningOverrideDeploymentName != "" || c.VersioningOverrideBuildId != "" {
			return fmt.Errorf("cannot set pinned deployment name or build id with %v behavior",
				c.VersioningOverrideBehavior)
		}
	}

	if c.VersioningOverrideBehavior.Value == "pinned" {
		if c.VersioningOverrideDeploymentName == "" && c.VersioningOverrideBuildId == "" {
			return fmt.Errorf("missing deployment name and/or build id with 'pinned' behavior")
		}
	}

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})

	var overrideChange *client.VersioningOverrideChange
	switch c.VersioningOverrideBehavior.Value {
	case "unspecified":
		overrideChange = &client.VersioningOverrideChange{
			Value: nil,
		}
	case "pinned":
		overrideChange = &client.VersioningOverrideChange{
			Value: &client.PinnedVersioningOverride{
				Version: worker.WorkerDeploymentVersion{
					DeploymentName: c.VersioningOverrideDeploymentName,
					BuildId:        c.VersioningOverrideBuildId,
				},
			},
		}
	case "auto_upgrade":
		overrideChange = &client.VersioningOverrideChange{
			Value: &client.AutoUpgradeVersioningOverride{},
		}
	default:
		return fmt.Errorf(
			"invalid deployment behavior: %v, valid values are: 'unspecified', 'pinned', and 'auto_upgrade'",
			c.VersioningOverrideBehavior,
		)
	}

	// Run single or batch
	if err != nil {
		return err
	} else if exec != nil {

		_, err := cl.UpdateWorkflowExecutionOptions(cctx, client.UpdateWorkflowExecutionOptionsRequest{
			WorkflowId: exec.WorkflowId,
			RunId:      exec.RunId,
			WorkflowExecutionOptionsChanges: client.WorkflowExecutionOptionsChanges{
				VersioningOverride: overrideChange,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to update workflow options: %w", err)
		}
		cctx.Printer.Println("Update workflow options succeeded")
	} else { // Run batch
		var workflowExecutionOptions *workflowpb.WorkflowExecutionOptions
		protoMask, err := fieldmaskpb.New(workflowExecutionOptions, "versioning_override")
		if err != nil {
			return fmt.Errorf("invalid field mask: %w", err)
		}

		var protoVerOverride *workflowpb.VersioningOverride
		if overrideChange != nil {
			protoVerOverride = versioningOverrideToProto(overrideChange.Value)
		}

		batchReq.Operation = &workflowservice.StartBatchOperationRequest_UpdateWorkflowOptionsOperation{
			UpdateWorkflowOptionsOperation: &batch.BatchOperationUpdateWorkflowExecutionOptions{
				Identity: c.Parent.Identity,
				WorkflowExecutionOptions: &workflowpb.WorkflowExecutionOptions{
					VersioningOverride: protoVerOverride,
				},
				UpdateMask: protoMask,
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}
	return nil
}

func (c *TemporalWorkflowMetadataCommand) run(cctx *CommandContext, _ []string) error {
	return queryHelper(cctx, c.Parent, PayloadInputOptions{},
		metadataQueryName, c.RejectCondition, c.WorkflowReferenceOptions)
}

func (c *TemporalWorkflowQueryCommand) run(cctx *CommandContext, args []string) error {
	return queryHelper(cctx, c.Parent, c.PayloadInputOptions,
		c.Name, c.RejectCondition, c.WorkflowReferenceOptions)
}

func (c *TemporalWorkflowSignalCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	input, err := c.buildRawInputPayloads()
	if err != nil {
		return err
	}

	exec, batchReq, err := c.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{})

	// Run single or batch
	if err != nil {
		return err
	} else if exec != nil {
		// We have to use the raw signal service call here because the Go SDK's
		// signal call doesn't accept multiple arguments.
		_, err = cl.WorkflowService().SignalWorkflowExecution(cctx, &workflowservice.SignalWorkflowExecutionRequest{
			Namespace:         c.Parent.Namespace,
			WorkflowExecution: &common.WorkflowExecution{WorkflowId: c.WorkflowId, RunId: c.RunId},
			SignalName:        c.Name,
			Input:             input,
			Identity:          c.Parent.Identity,
		})
		if err != nil {
			return fmt.Errorf("failed signalling workflow: %w", err)
		}
		cctx.Printer.Println("Signal workflow succeeded")
	} else { // batchReq != nil
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_SignalOperation{
			SignalOperation: &batch.BatchOperationSignal{
				Signal:   c.Name,
				Input:    input,
				Identity: c.Parent.Identity,
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}
	return nil
}

func (c *TemporalWorkflowStackCommand) run(cctx *CommandContext, args []string) error {
	return queryHelper(cctx, c.Parent, PayloadInputOptions{},
		"__stack_trace", c.RejectCondition, c.WorkflowReferenceOptions)
}

func (c *TemporalWorkflowTerminateCommand) run(cctx *CommandContext, _ []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// We create a faux SingleWorkflowOrBatchOptions to use the shared logic
	opts := SingleWorkflowOrBatchOptions{
		WorkflowId: c.WorkflowId,
		RunId:      c.RunId,
		Query:      c.Query,
		Reason:     c.Reason,
		Yes:        c.Yes,
		Rps:        c.Rps,
	}

	exec, batchReq, err := opts.workflowExecOrBatch(cctx, c.Parent.Namespace, cl, singleOrBatchOverrides{
		// You're allowed to specify a reason when terminating a workflow
		AllowReasonWithWorkflowID: true,
	})

	// Run single or batch
	if err != nil {
		return err
	} else if exec != nil {
		reason := c.Reason
		if reason == "" {
			reason = defaultReason()
		}
		err = cl.TerminateWorkflow(cctx, exec.WorkflowId, exec.RunId, reason)
		if err != nil {
			return fmt.Errorf("failed to terminate workflow: %w", err)
		}
		cctx.Printer.Println("Workflow terminated")
	} else { // batchReq != nil
		batchReq.Operation = &workflowservice.StartBatchOperationRequest_TerminationOperation{
			TerminationOperation: &batch.BatchOperationTermination{
				Identity: c.Parent.Identity,
			},
		}
		if err := startBatchJob(cctx, cl, batchReq); err != nil {
			return err
		}
	}

	return nil
}

func (c *TemporalWorkflowUpdateStartCommand) run(cctx *CommandContext, args []string) error {
	waitForStage := client.WorkflowUpdateStageUnspecified
	switch c.WaitForStage.Value {
	case "accepted":
		waitForStage = client.WorkflowUpdateStageAccepted
	}
	if waitForStage != client.WorkflowUpdateStageAccepted {
		return fmt.Errorf("invalid wait for stage: %v, valid values are: 'accepted'", c.WaitForStage)
	}
	return workflowUpdateHelper(cctx, c.Parent.Parent.ClientOptions, c.PayloadInputOptions,
		UpdateTargetingOptions{
			WorkflowId: c.WorkflowId,
			UpdateId:   c.UpdateId,
			RunId:      c.RunId,
		}, c.UpdateStartingOptions, waitForStage)
}

func (c *TemporalWorkflowUpdateExecuteCommand) run(cctx *CommandContext, args []string) error {
	return workflowUpdateHelper(cctx, c.Parent.Parent.ClientOptions, c.PayloadInputOptions,
		UpdateTargetingOptions{
			WorkflowId: c.WorkflowId,
			UpdateId:   c.UpdateId,
			RunId:      c.RunId,
		}, c.UpdateStartingOptions, client.WorkflowUpdateStageCompleted)
}

func (c *TemporalWorkflowUpdateResultCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	updateHandle := cl.GetWorkflowUpdateHandle(client.GetWorkflowUpdateHandleOptions{
		WorkflowID: c.WorkflowId,
		RunID:      c.RunId,
		UpdateID:   c.UpdateId,
	})
	var valuePtr any
	err = updateHandle.Get(cctx, &valuePtr)
	printMe := struct {
		UpdateID string `json:"updateId"`
		Result   any    `json:"result,omitempty"`
		Failure  any    `json:"failure,omitempty"`
	}{UpdateID: updateHandle.UpdateID()}

	if err != nil {
		// Genuine update failure, so, include that in output rather than saying we couldn't fetch
		appErr := &temporal.ApplicationError{}
		if errors.As(err, &appErr) {
			if cctx.JSONOutput {
				fromAppErr, err := fromApplicationError(appErr)
				if err != nil {
					return fmt.Errorf("unable to fetch update result: %w", err)
				}
				printMe.Failure = fromAppErr
			} else {
				printMe.Failure = appErr.Error()
			}
			if err := cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{}); err != nil {
				return err
			}
			return errors.New("update is failed")
		}
		return fmt.Errorf("unable to fetch update result: %w", err)
	}

	printMe.Result = valuePtr
	return cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
}

func (c *TemporalWorkflowUpdateDescribeCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// TODO: Ideally workflow update handle's Get would allow a nonblocking option and we'd use that
	pollReq := &workflowservice.PollWorkflowExecutionUpdateRequest{
		Namespace: c.Parent.Parent.Namespace,
		UpdateRef: &update.UpdateRef{
			WorkflowExecution: &common.WorkflowExecution{
				WorkflowId: c.WorkflowId,
				RunId:      c.RunId,
			},
			UpdateId: c.UpdateId,
		},
		Identity: c.Parent.Parent.Identity,
		// WaitPolicy omitted intentionally for nonblocking
	}
	resp, err := cl.WorkflowService().PollWorkflowExecutionUpdate(cctx, pollReq)
	if err != nil {
		return fmt.Errorf("failed to describe update: %w", err)
	}

	printMe := struct {
		UpdateID string `json:"updateId"`
		Result   any    `json:"result,omitempty"`
		Failure  any    `json:"failure,omitempty"`
		Stage    string `json:"stage"`
	}{UpdateID: c.UpdateId, Stage: resp.GetStage().String()}

	switch v := resp.GetOutcome().GetValue().(type) {
	case *update.Outcome_Failure:
		// TODO: This doesn't exactly match update result, but it's hard to make it do so w/o higher
		//   level poll api
		if cctx.JSONOutput {
			printMe.Failure = v.Failure
		} else {
			printMe.Failure = v.Failure.GetMessage()
		}
	case *update.Outcome_Success:
		var value any
		if err := converter.GetDefaultDataConverter().FromPayloads(v.Success, &value); err != nil {
			value = fmt.Sprintf("<failed converting: %v>", err)
		}
		printMe.Result = value
	}

	return cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
}

func workflowUpdateHelper(cctx *CommandContext,
	clientOpts ClientOptions,
	inputOpts PayloadInputOptions,
	updateTargetOpts UpdateTargetingOptions,
	updateStartOpts UpdateStartingOptions,
	waitForStage client.WorkflowUpdateStage,
) error {
	cl, err := clientOpts.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	input, err := inputOpts.buildRawInput()
	if err != nil {
		return err
	}

	request := client.UpdateWorkflowOptions{
		WorkflowID:          updateTargetOpts.WorkflowId,
		RunID:               updateTargetOpts.RunId,
		UpdateName:          updateStartOpts.Name,
		UpdateID:            updateTargetOpts.UpdateId,
		FirstExecutionRunID: updateStartOpts.FirstExecutionRunId,
		Args:                input,
		WaitForStage:        waitForStage,
	}

	updateHandle, err := cl.UpdateWorkflow(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to update workflow: %w", err)
	}
	if waitForStage == client.WorkflowUpdateStageAccepted {
		// Use a canceled context to check whether the initial server response
		// shows that the update has _already_ failed, without issuing a second request.
		ctx, cancel := context.WithCancel(cctx)
		cancel()
		err = updateHandle.Get(ctx, nil)
		var timeoutOrCanceledErr *client.WorkflowUpdateServiceTimeoutOrCanceledError
		if err != nil && !errors.As(err, &timeoutOrCanceledErr) {
			return fmt.Errorf("unable to update workflow: %w", err)
		}
		return cctx.Printer.PrintStructured(
			struct {
				Name     string `json:"name"`
				UpdateID string `json:"updateId"`
			}{Name: updateStartOpts.Name, UpdateID: updateHandle.UpdateID()},
			printer.StructuredOptions{})
	}

	var valuePtr interface{}
	err = updateHandle.Get(cctx, &valuePtr)
	if err != nil {
		return fmt.Errorf("unable to update workflow: %w", err)
	}

	return cctx.Printer.PrintStructured(
		struct {
			Name     string      `json:"name"`
			UpdateID string      `json:"updateId"`
			Result   interface{} `json:"result"`
		}{Name: updateStartOpts.Name, UpdateID: updateHandle.UpdateID(), Result: valuePtr},
		printer.StructuredOptions{})
}

func username() string {
	username := "<unknown-user>"
	if u, err := user.Current(); err != nil && u.Username != "" {
		username = u.Username
	}
	return username
}

func defaultReason() string {
	return "Requested from CLI by " + username()
}

type singleOrBatchOverrides struct {
	AllowReasonWithWorkflowID bool
}

func (s *SingleWorkflowOrBatchOptions) workflowExecOrBatch(
	cctx *CommandContext,
	namespace string,
	cl client.Client,
	overrides singleOrBatchOverrides,
) (*common.WorkflowExecution, *workflowservice.StartBatchOperationRequest, error) {
	// If workflow is set, we return single execution
	if s.WorkflowId != "" {
		if s.Query != "" {
			return nil, nil, fmt.Errorf("cannot set query when workflow ID is set")
		} else if s.Reason != "" && !overrides.AllowReasonWithWorkflowID {
			return nil, nil, fmt.Errorf("cannot set reason when workflow ID is set")
		} else if s.Yes {
			return nil, nil, fmt.Errorf("cannot set 'yes' when workflow ID is set")
		} else if s.Rps != 0 {
			return nil, nil, fmt.Errorf("cannot set rps when workflow ID is set")
		}
		return &common.WorkflowExecution{WorkflowId: s.WorkflowId, RunId: s.RunId}, nil, nil
	}

	// Check query is set properly
	if s.Query == "" {
		return nil, nil, fmt.Errorf("must set either workflow ID or query")
	} else if s.WorkflowId != "" {
		return nil, nil, fmt.Errorf("cannot set workflow ID when query is set")
	} else if s.RunId != "" {
		return nil, nil, fmt.Errorf("cannot set run ID when query is set")
	}

	// Count the workflows that will be affected
	count, err := cl.CountWorkflow(cctx, &workflowservice.CountWorkflowExecutionsRequest{Query: s.Query})
	if err != nil {
		return nil, nil, fmt.Errorf("failed counting workflows from query: %w", err)
	}
	yes, err := cctx.promptYes(
		fmt.Sprintf("Start batch against approximately %v workflow(s)? y/N", count.Count), s.Yes)
	if err != nil {
		return nil, nil, err
	} else if !yes {
		// We consider this a command failure
		return nil, nil, fmt.Errorf("user denied confirmation")
	}

	// Default the reason if not set
	reason := s.Reason
	if reason == "" {
		reason = defaultReason()
	}

	return nil, &workflowservice.StartBatchOperationRequest{
		MaxOperationsPerSecond: s.Rps,
		Namespace:              namespace,
		JobId:                  uuid.NewString(),
		VisibilityQuery:        s.Query,
		Reason:                 reason,
	}, nil
}

func startBatchJob(cctx *CommandContext, cl client.Client, req *workflowservice.StartBatchOperationRequest) error {
	_, err := cl.WorkflowService().StartBatchOperation(cctx, req)
	if err != nil {
		return fmt.Errorf("failed starting batch operation: %w", err)
	}
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(
			struct {
				BatchJobID string `json:"batchJobId"`
			}{BatchJobID: req.JobId},
			printer.StructuredOptions{})
	}
	cctx.Printer.Printlnf("Started batch for job ID: %v", req.JobId)
	return nil
}

func queryHelper(cctx *CommandContext,
	parent *TemporalWorkflowCommand,
	inputOpts PayloadInputOptions,
	queryType string,
	rejectCondition StringEnum,
	execution WorkflowReferenceOptions,
) error {
	cl, err := parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	input, err := inputOpts.buildRawInputPayloads()
	if err != nil {
		return err
	}

	queryRejectCond := enums.QUERY_REJECT_CONDITION_UNSPECIFIED
	switch rejectCondition.Value {
	case "":
	case "not_open":
		queryRejectCond = enums.QUERY_REJECT_CONDITION_NOT_OPEN
	case "not_completed_cleanly":
		queryRejectCond = enums.QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
	default:
		return fmt.Errorf("invalid query reject condition: %v, valid values are: 'not_open', 'not_completed_cleanly'", rejectCondition)
	}

	result, err := cl.WorkflowService().QueryWorkflow(cctx, &workflowservice.QueryWorkflowRequest{
		Namespace: parent.Namespace,
		Execution: &common.WorkflowExecution{WorkflowId: execution.WorkflowId, RunId: execution.RunId},
		Query: &query.WorkflowQuery{
			QueryType: queryType,
			QueryArgs: input,
		},
		QueryRejectCondition: queryRejectCond,
	})

	if err != nil {
		return fmt.Errorf("querying workflow failed: %w", err)
	}

	if result.QueryRejected != nil {
		return fmt.Errorf("query was rejected, workflow has status: %v", result.QueryRejected.GetStatus())
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	if queryType == metadataQueryName {
		var metadata sdkpb.WorkflowMetadata
		err := UnmarshalProtoJSONWithOptions(result.QueryResult.Payloads[0].Data, &metadata, true)
		if err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		cctx.Printer.Println(color.MagentaString("Metadata:"))

		qDefs := metadata.GetDefinition().GetQueryDefinitions()
		if len(qDefs) > 0 {
			cctx.Printer.Println(printer.NonJSONIndent, color.MagentaString("Query Definitions:"))
			err := cctx.Printer.PrintStructured(qDefs, printer.StructuredOptions{
				Table:              &printer.TableOptions{NoHeader: true},
				NonJSONExtraIndent: 1,
			})
			if err != nil {
				return err
			}
		}
		sigDefs := metadata.GetDefinition().GetSignalDefinitions()
		if len(sigDefs) > 0 {
			cctx.Printer.Println(printer.NonJSONIndent, color.MagentaString("Signal Definitions:"))
			err := cctx.Printer.PrintStructured(sigDefs, printer.StructuredOptions{
				Table:              &printer.TableOptions{NoHeader: true},
				NonJSONExtraIndent: 1,
			})
			if err != nil {
				return err
			}
		}
		updDefs := metadata.GetDefinition().GetUpdateDefinitions()
		if len(updDefs) > 0 {
			cctx.Printer.Println(printer.NonJSONIndent, color.MagentaString("Update Definitions:"))
			err := cctx.Printer.PrintStructured(updDefs, printer.StructuredOptions{
				Table:              &printer.TableOptions{NoHeader: true},
				NonJSONExtraIndent: 1,
			})
			if err != nil {
				return err
			}
		}
		if metadata.GetCurrentDetails() != "" {
			cctx.Printer.Println(printer.NonJSONIndent, color.MagentaString("Current Details:"))
			cctx.Printer.Println(printer.NonJSONIndent, printer.NonJSONIndent,
				metadata.GetCurrentDetails())
		}
		return nil
	} else {
		cctx.Printer.Println(color.MagentaString("Query result:"))
		output := struct {
			QueryResult json.RawMessage `cli:",cardOmitEmpty"`
		}{}
		output.QueryResult, err = cctx.MarshalFriendlyJSONPayloads(result.QueryResult)
		if err != nil {
			return fmt.Errorf("failed to marshal query result: %w", err)
		}

		return cctx.Printer.PrintStructured(output, printer.StructuredOptions{})
	}
}

// This is (mostly) copy-pasted from the SDK since it's not exposed. Most of this will go away once
// the deprecated fields are no longer supported.
func versioningOverrideToProto(versioningOverride client.VersioningOverride) *workflowpb.VersioningOverride {
	if versioningOverride == nil {
		return nil
	}
	switch v := versioningOverride.(type) {
	case *client.PinnedVersioningOverride:
		return &workflowpb.VersioningOverride{
			Behavior:      enums.VERSIONING_BEHAVIOR_PINNED,
			PinnedVersion: fmt.Sprintf("%s.%s", v.Version.DeploymentName, v.Version.BuildId),
			Deployment: &deploymentpb.Deployment{
				SeriesName: v.Version.DeploymentName,
				BuildId:    v.Version.BuildId,
			},
			Override: &workflowpb.VersioningOverride_Pinned{
				Pinned: &workflowpb.VersioningOverride_PinnedOverride{
					Behavior: workflowpb.VersioningOverride_PINNED_OVERRIDE_BEHAVIOR_PINNED,
					Version: &deploymentpb.WorkerDeploymentVersion{
						DeploymentName: v.Version.DeploymentName,
						BuildId:        v.Version.BuildId,
					},
				},
			},
		}
	case *client.AutoUpgradeVersioningOverride:
		return &workflowpb.VersioningOverride{
			Behavior: enums.VERSIONING_BEHAVIOR_AUTO_UPGRADE,
			Override: &workflowpb.VersioningOverride_AutoUpgrade{AutoUpgrade: true},
		}
	default:
		return nil
	}
}
