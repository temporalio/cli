package temporalcli

import (
	"bytes"
	"os"

	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/encoding/protojson"
)

func (c *TemporalWorkflowFixHistoryJsonCommand) run(cctx *CommandContext, args []string) error {
	raw, err := os.ReadFile(c.Source)
	if err != nil {
		return err
	}

	hjo := client.HistoryJSONOptions{}
	history, err := client.HistoryFromJSON(bytes.NewReader(raw), hjo)
	if err != nil {
		return err
	}

	mo := protojson.MarshalOptions{Indent: "  "}
	raw, err = mo.Marshal(history)
	if err != nil {
		return err
	}

	switch c.Target {
	case "", "-":
		_, err = cctx.Options.Stdout.Write(raw)
		return err

	default:
		return os.WriteFile(c.Target, raw, 0o666)
	}
}
