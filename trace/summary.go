package trace

import (
	"fmt"
	"github.com/temporalio/cli/common"
	"github.com/urfave/cli/v2"
	sdkclient "go.temporal.io/sdk/client"
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	bold = color.New(color.Bold)
)

// Table is a list of rows to be printed with the same width on the first column
type Table struct {
	rows []Row
	w    io.Writer
}

type Row struct {
	key   string
	value string
}

func NewTable(w io.Writer) *Table {
	t := new(Table)
	t.w = w
	return t
}

// padRight pads a text to the right with spaces to fill the padding width.
func padRight(text string, padding int) string {
	return fmt.Sprintf("%-*s", padding, text)
}

func (t *Table) Print() error {
	// Figure out which is the widest key
	maxWidth := 0
	for _, row := range t.rows {
		if l := len(row.key); l > maxWidth {
			maxWidth = l
		}
	}
	for _, row := range t.rows {
		_, err := fmt.Fprintf(t.w, "  %s : %s\n", bold.Sprintf(padRight(row.key, maxWidth)), row.value)
		if err != nil {
			return err
		}
	}
	return nil
}

// AppendRows appends a list of rows to the table
func (t *Table) AppendRows(rows ...Row) {
	t.rows = append(t.rows, rows...)
}

func PrintWorkflowSummary(c *cli.Context, sdkClient sdkclient.Client, wfId, runId string) error {
	tcCtx, cancel := common.NewIndefiniteContext(c)
	defer cancel()

	res, err := sdkClient.DescribeWorkflowExecution(tcCtx, wfId, runId)
	if err != nil {
		return err
	}

	info := res.GetWorkflowExecutionInfo()

	title.Println("Execution summary:")
	tb := NewTable(os.Stdout)
	tb.AppendRows(
		Row{"Workflow Id", info.GetExecution().GetWorkflowId()},
		Row{"Workflow Run Id", info.GetExecution().GetRunId()},
		Row{"Workflow Type", info.GetType().GetName()},
		Row{"Task Queue", info.GetTaskQueue()},
	)
	_ = tb.Print()
	fmt.Println()

	return nil
}
