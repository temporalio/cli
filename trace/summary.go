package trace

import (
	"fmt"

	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
	sdkclient "go.temporal.io/sdk/client"
)

type Row struct {
	Key   string
	Value string
}

func PrintWorkflowSummary(c *cli.Context, sdkClient sdkclient.Client, wfId, runId string) error {
	tcCtx, cancel := common.NewIndefiniteContext(c)
	defer cancel()

	res, err := sdkClient.DescribeWorkflowExecution(tcCtx, wfId, runId)
	if err != nil {
		return err
	}

	info := res.GetWorkflowExecutionInfo()

	_, _ = title.Println("Execution summary:")
	rows := []Row{
		{"Workflow Id", info.GetExecution().GetWorkflowId()},
		{"Workflow Run Id", info.GetExecution().GetRunId()},
		{"Workflow Type", info.GetType().GetName()},
		{"Task Queue", info.GetTaskQueue()},
	}
	var i []interface{}
	for _, row := range rows {
		i = append(i, row)
	}

	err = output.PrintItems(c, i,
		&output.PrintOptions{
			Fields:   []string{"Key", "Value"},
			NoHeader: true,
		},
	)
	_, _ = fmt.Println()

	return err
}
