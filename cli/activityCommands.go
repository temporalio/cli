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

	"github.com/temporalio/tctl/pkg/color"
	"github.com/urfave/cli/v2"
	failurepb "go.temporal.io/api/failure/v1"
	"go.temporal.io/api/workflowservice/v1"

	"go.temporal.io/server/common/payloads"
)

// CompleteActivity completes an activity
func CompleteActivity(c *cli.Context) error {
	namespace, err := getRequiredGlobalOption(c, FlagNamespace)
	if err != nil {
		return err
	}

	wid := c.String(FlagWorkflowID)
	rid := c.String(FlagRunID)
	if rid == "" {
		return fmt.Errorf("provide non-empty run id")
	}

	activityID := c.String(FlagActivityID)
	if len(activityID) == 0 {
		return fmt.Errorf("provide non-empty activity id")
	}
	result := c.String(FlagResult)
	identity := c.String(FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.FrontendClient(c)
	_, err = frontendClient.RespondActivityTaskCompletedById(ctx, &workflowservice.RespondActivityTaskCompletedByIdRequest{
		Namespace:  namespace,
		WorkflowId: wid,
		RunId:      rid,
		ActivityId: activityID,
		Result:     payloads.EncodeString(result),
		Identity:   identity,
	})
	if err != nil {
		return fmt.Errorf("unable to Complete activity.\n%s", err)
	} else {
		fmt.Println(color.Green(c, "activity was Completed"))
	}
	return nil
}

// FailActivity fails an activity
func FailActivity(c *cli.Context) error {
	namespace, err := getRequiredGlobalOption(c, FlagNamespace)
	if err != nil {
		return err
	}

	wid := c.String(FlagWorkflowID)
	rid := c.String(FlagRunID)
	if rid == "" {
		return fmt.Errorf("provide non-empty run id")
	}

	activityID := c.String(FlagActivityID)
	if len(activityID) == 0 {
		return fmt.Errorf("provide non-empty activity id")
	}
	reason := c.String(FlagReason)
	detail := c.String(FlagDetail)
	identity := c.String(FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

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
				Details:      payloads.EncodeString(detail),
			}},
		},
		Identity: identity,
	})
	if err != nil {
		return fmt.Errorf("unable to Fail activity.\n%s", err)
	} else {
		fmt.Println(color.Green(c, "activity was Failed"))

		return nil
	}
}
