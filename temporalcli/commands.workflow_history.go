package temporalcli

import (
	"bytes"
	"fmt"
	"os"

	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	historyFormatJSON      = "json"
	historyFormatPB        = "pb"
	historyFormatPBTXT     = "pbtxt"
	historyFormatProto     = "proto"
	historyFormatProtoText = "prototext"
)

func (c *TemporalWorkflowHistoryConvertCommand) run(cctx *CommandContext, args []string) error {
	raw, err := os.ReadFile(c.SourceFile)
	if err != nil {
		return err
	}

	var history *historypb.History
	switch c.SourceFormat.Value {
	case historyFormatJSON:
		history, err = client.HistoryFromJSON(bytes.NewReader(raw), client.HistoryJSONOptions{})
		if err != nil {
			return err
		}

	case historyFormatProtoText:
		fallthrough
	case historyFormatPBTXT:
		history = new(historypb.History)
		err = prototext.Unmarshal(raw, history)
		if err != nil {
			return err
		}

	case historyFormatProto:
		fallthrough
	case historyFormatPB:
		history = new(historypb.History)
		err = history.Unmarshal(raw)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("BUG: --source-format=%q is not implemented yet", c.SourceFormat.Value)
	}

	switch c.TargetFormat.Value {
	case historyFormatJSON:
		mo := protojson.MarshalOptions{Indent: "  "}
		raw, err = mo.Marshal(history)
		if err != nil {
			return err
		}

	case historyFormatProtoText:
		fallthrough
	case historyFormatPBTXT:
		mo := prototext.MarshalOptions{Indent: "  "}
		raw, err = mo.Marshal(history)
		if err != nil {
			return err
		}

	case historyFormatProto:
		fallthrough
	case historyFormatPB:
		raw, err = history.Marshal()
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("BUG: --target-format=%q is not implemented yet", c.TargetFormat.Value)
	}

	switch c.TargetFile {
	case "":
		fallthrough
	case "-":
		_, err = cctx.Options.Stdout.Write(raw)
		if err != nil {
			return err
		}

	default:
		err = os.WriteFile(c.TargetFile, raw, 0o666)
		if err != nil {
			return err
		}
	}

	return nil
}
