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
		Row{"Workflow Id", info.GetExecution().GetWorkflowId()},
		Row{"Workflow Run Id", info.GetExecution().GetRunId()},
		Row{"Workflow Type", info.GetType().GetName()},
		Row{"Task Queue", info.GetTaskQueue()},
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
