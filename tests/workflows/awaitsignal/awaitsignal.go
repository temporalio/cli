package awaitsignal

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	Done   = "done"
	Input1 = "input1"
)

func Workflow(ctx workflow.Context) error {
	var v string
	_ = workflow.GetSignalChannel(ctx, Done).Receive(ctx, &v)

	if v == Input1 {
		return nil
	}

	return fmt.Errorf("expected %s, received %s", Input1, v)
}
