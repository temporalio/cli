package awaitsignal

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	Done   = "done"
	input1 = "input1"
)

func Workflow(ctx workflow.Context) error {
	var v string
	_ = workflow.GetSignalChannel(ctx, Done).Receive(ctx, &v)

	if v == input1 {
		return nil
	}

	return fmt.Errorf("expected %s, received %s", input1, v)
}
