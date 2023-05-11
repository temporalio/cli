package awaitsignal

import (
	"go.temporal.io/sdk/workflow"
)

const (
	Done = "done"
)

func Workflow(ctx workflow.Context) error {
	_ = workflow.GetSignalChannel(ctx, Done).Receive(ctx, nil)
	return nil
}
