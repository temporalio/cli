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

	"github.com/urfave/cli/v2"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/temporalio/tctl/pkg/color"
	"github.com/temporalio/tctl/pkg/output"
)

const (
	fullWorkflowServiceName = "temporal.api.workflowservice.v1.WorkflowService"
)

// HealthCheck check frontend health.
func HealthCheck(c *cli.Context) error {
	healthClient := cFactory.HealthClient(c)
	ctx, cancel := newContext(c)
	defer cancel()

	req := &healthpb.HealthCheckRequest{
		Service: fullWorkflowServiceName,
	}
	resp, err := healthClient.Check(ctx, req)

	if err != nil {
		return fmt.Errorf("unable to check health, service: %q.\n%s", req.GetService(), err)
	}

	fmt.Printf("%s: ", req.GetService())
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		fmt.Println(color.Red(c, "%v", resp.Status))
		return nil
	}

	fmt.Println(color.Green(c, "%v", resp.Status))
	return nil
}

// ListSearchAttributes lists search attributes
func ListSearchAttributes(c *cli.Context) error {
	wfClient, err := getSDKClient(c)
	if err != nil {
		return err
	}
	ctx, cancel := newContext(c)
	defer cancel()

	resp, err := wfClient.GetSearchAttributes(ctx)
	if err != nil {
		return fmt.Errorf("unable to get search attributes.\n%s", err)
	}

	searchAttributes := resp.GetKeys()
	var items []interface{}
	type sa struct {
		Name string
		Type string
	}
	for key, val := range searchAttributes {
		items = append(items, sa{Name: key, Type: val.String()})
	}

	opts := &output.PrintOptions{
		Fields:  []string{"Name", "Type"},
		NoPager: true,
	}
	output.PrintItems(c, items, opts)

	return nil
}
