// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"fmt"

	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/urfave/cli/v2"
	failurepb "go.temporal.io/api/failure/v1"
	"go.temporal.io/api/workflowservice/v1"
)

// CompleteActivity completes an Activity
func CompleteActivity(c *cli.Context) error {
	namespace, err := requiredFlag(c, FlagNamespace)
	if err != nil {
		return err
	}

	wid := c.String(FlagWorkflowID)
	rid := c.String(FlagRunID)
	aid := c.String(FlagActivityID)
	result := c.String(FlagResult)
	identity := c.String(FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

	// TODO: This should use customDataConverter once the plugin interface
	// supports the full DataConverter API.
	resultPayloads, _ := defaultDataConverter().ToPayloads(result)

	frontendClient := cFactory.FrontendClient(c)
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
	namespace, err := requiredFlag(c, FlagNamespace)
	if err != nil {
		return err
	}

	wid := c.String(FlagWorkflowID)
	rid := c.String(FlagRunID)

	activityID := c.String(FlagActivityID)
	if len(activityID) == 0 {
		return fmt.Errorf("provide non-empty activity id")
	}
	reason := c.String(FlagReason)
	detail := c.String(FlagDetail)
	identity := c.String(FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

	// TODO: This should use customDataConverter once the plugin interface
	// supports the full DataConverter API.
	detailsPayloads, _ := defaultDataConverter().ToPayloads(detail)

	frontendClient := cFactory.FrontendClient(c)
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
