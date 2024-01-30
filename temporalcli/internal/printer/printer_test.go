package printer_test

import (
	"bytes"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/temporalcli/internal/printer"
)

// TODO(cretz): Test:
// * Text printer specific fields
// * Text printer specific and non-specific fields and all sorts of table options
// * JSON printer

func TestTextPrinter(t *testing.T) {
	type MyStruct struct {
		Foo              string
		Bar              bool
		unexportedBaz    string
		ReallyLongField  any
		Omitted          string `cli:",omit"`
		OmittedCardEmpty string `cli:",cardOmitEmpty"`
	}
	var buf bytes.Buffer
	p := printer.Printer{Output: &buf}
	// Simple struct non-table no fields set
	require.NoError(t, p.PrintStructured([]*MyStruct{
		{
			Foo:           "1",
			unexportedBaz: "2",
			ReallyLongField: struct {
				Key any `json:"key"`
			}{Key: 123},
			Omitted:          "value",
			OmittedCardEmpty: "value",
		},
		{
			Foo:             "not-a-number",
			Bar:             true,
			ReallyLongField: map[string]int{"": 0},
		},
	}, printer.StructuredOptions{}))
	// Check
	require.Equal(t, normalizeMultiline(`
  Foo               1
  Bar               false
  ReallyLongField   {"key":123}
  OmittedCardEmpty  value

  Foo              not-a-number
  Bar              true
  ReallyLongField  map[:0]`), normalizeMultiline(buf.String()))

	// TODO(cretz): Tables and more options
}

func normalizeMultiline(s string) string {
	// Split lines, trim trailing space on each (also removes \r), remove empty
	// lines, re-join
	var ret string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		// Only non-empty lines
		if line != "" {
			if ret != "" {
				ret += "\n"
			}
			ret += line
		}
	}
	return ret
}
