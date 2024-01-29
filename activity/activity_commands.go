package activity

import (
	"fmt"

	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/urfave/cli/v2"
	failurepb "go.temporal.io/api/failure/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
)

// CompleteActivity completes an Activity
func CompleteActivity(c *cli.Context) error {
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)
	aid := c.String(common.FlagActivityID)
	result := c.String(common.FlagResult)
	identity := c.String(common.FlagIdentity)
	ctx, cancel := common.NewContext(c)
	defer cancel()

	resultPayloads, _ := converter.GetDefaultDataConverter().ToPayloads(result)

	frontendClient := client.Factory(c.App).FrontendClient(c)
	_, err = frontendClient.RespondActivityTaskCompletedById(ctx, &workflowservice.RespondActivityTaskCompletedByIdRequest{
		Namespace:  namespace,
		WorkflowId: wid,
		RunId:      rid,
		ActivityId: aid,
		Result:     resultPayloads,
		Identity:   identity,
	})
	if err != nil {
		return fmt.Errorf("unable to complete Activity: %w", err)
	} else {
		fmt.Println(color.Green(c, "Activity was completed"))
	}
	return nil
}

// FailActivity fails an activity
func FailActivity(c *cli.Context) error {
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)

	activityID := c.String(common.FlagActivityID)
	if len(activityID) == 0 {
		return fmt.Errorf("provide non-empty activity id")
	}
	reason := c.String(common.FlagReason)
	detail := c.String(common.FlagDetail)
	identity := c.String(common.FlagIdentity)
	ctx, cancel := common.NewContext(c)
	defer cancel()

	detailsPayloads, _ := converter.GetDefaultDataConverter().ToPayloads(detail)

	frontendClient := client.Factory(c.App).FrontendClient(c)
	_, err = frontendClient.RespondActivityTaskFailedById(ctx, &workflowservice.RespondActivityTaskFailedByIdRequest{
		Namespace:  namespace,
		WorkflowId: wid,
		RunId:      rid,
		ActivityId: activityID,
		Failure: &failurepb.Failure{
			Message: reason,
			Source:  "CLI",
			FailureInfo: &failurepb.Failure_ApplicationFailureInfo{ApplicationFailureInfo: &failurepb.ApplicationFailureInfo{
				NonRetryable: true,
				Details:      detailsPayloads,
			}},
		},
		Identity: identity,
	})
	if err != nil {
		return fmt.Errorf("unable to Fail activity: %w", err)
	} else {
		fmt.Println(color.Green(c, "activity was Failed"))

		return nil
	}
}
