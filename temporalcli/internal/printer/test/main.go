package main

import (
	"os"

	"github.com/temporalio/cli/temporalcli/internal/printer"
)

// This main function is used to test that the printer package don't panic if
// the CLI is run without a STDOUT. This is a tricky thing to test, as `go test`
// internally fix improper STDOUT.
func main() {
	p := &printer.Printer{
		Output: os.Stdout,
		JSON:   false,
	}
	p.Println("Test writing to stdout using Printer")
	os.Exit(0)
}
