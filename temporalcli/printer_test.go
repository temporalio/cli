package temporalcli_test

import (
	"bytes"
	"strings"
	"testing"
	"unicode"

	"github.com/temporalio/cli/temporalcli"
)

// TODO(cretz): Test:
// * Text printer specific fields
// * Text printer specific and non-specific fields and all sorts of table options
// * JSON printer

func TestTextPrinter(t *testing.T) {
	h := NewCommandHarness(t)
	defer h.Close()
	type MyStruct struct {
		Foo             string
		Bar             bool
		unexportedBaz   string
		ReallyLongField any
	}
	var buf bytes.Buffer
	p := temporalcli.NewTextPrinter(temporalcli.TextPrinterOptions{Output: &buf})

	// Simple struct non-table no fields set
	h.NoError(p.Print(temporalcli.PrintOptions{}, []*MyStruct{
		{
			Foo:           "1",
			unexportedBaz: "2",
			ReallyLongField: struct {
				Key any `json:"key"`
			}{Key: 123},
		},
		{
			Foo:             "not-a-number",
			Bar:             true,
			ReallyLongField: map[string]int{"": 0},
		},
	}))
	// Check
	h.Equal(normalizeMultiline(`
  Foo                        1
  Bar              false
  ReallyLongField  {"key":123}

  Foo              not-a-number
  Bar              true
  ReallyLongField  map[:0]`), normalizeMultiline(buf.String()))

	// TODO(cretz): more
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
