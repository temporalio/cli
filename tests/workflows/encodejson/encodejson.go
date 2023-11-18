package encodejson

import (
	"encoding/json"
	"time"

	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Workflow is a Hello World workflow definition. (Ordinarily I would define
// this as a variadic function, but that's not supported currently--see
// https://github.com/temporalio/sdk-go/issues/1114)
func Workflow(ctx workflow.Context, a, b, c, d interface{}) (string, error) {
	args := []interface{}{a, b, c, d}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("EncodeJSON workflow started", a, b, c, d)

	result, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	logger.Info("EncodeJSON workflow completed", result)

	return string(result), nil
}
