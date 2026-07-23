package commandsgen

import (
	"bytes"
	"testing"
)

func TestGenerateRunnableCommandReturnsRunError(t *testing.T) {
	generated, err := GenerateCommandsCode("temporalcli", "*CommandContext", Commands{
		CommandList: []Command{
			{
				FullName:         "temporal",
				NamePath:         []string{"temporal"},
				Summary:          "Temporal commands",
				DescriptionPlain: "Temporal commands",
			},
			{
				FullName:         "temporal example",
				NamePath:         []string{"temporal", "example"},
				Summary:          "Example command",
				DescriptionPlain: "Example command",
			},
		},
	})
	if err != nil {
		t.Fatalf("GenerateCommandsCode() error = %v", err)
	}

	for _, want := range [][]byte{
		[]byte("s.Command.RunE = func(c *cobra.Command, args []string) error"),
		[]byte("cctx.commandRunStarted = true"),
		[]byte("return s.run(cctx, args)"),
	} {
		if !bytes.Contains(generated, want) {
			t.Errorf("generated code does not contain %q", want)
		}
	}
	for _, unwanted := range [][]byte{
		[]byte("Options.Fail"),
		[]byte("s.Command.Run = func"),
	} {
		if bytes.Contains(generated, unwanted) {
			t.Errorf("generated code unexpectedly contains %q", unwanted)
		}
	}
}
