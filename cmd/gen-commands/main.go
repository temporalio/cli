package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/temporalio/cli/internal/commandsgen"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		pkg         string
		contextType string
	)

	flag.StringVar(&pkg, "pkg", "main", "Package name for generated code")
	flag.StringVar(&contextType, "context", "*CommandContext", "Context type for generated code")
	flag.Parse()

	// Read input from stdin
	yamlBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed reading input: %w", err)
	}

	// Parse YAML
	cmds, err := commandsgen.ParseCommands(yamlBytes)
	if err != nil {
		return fmt.Errorf("failed parsing YAML: %w", err)
	}

	// Generate code
	b, err := commandsgen.GenerateCommandsCode(pkg, contextType, cmds)
	if err != nil {
		return fmt.Errorf("failed generating code: %w", err)
	}

	// Write output to stdout
	if _, err = os.Stdout.Write(b); err != nil {
		return fmt.Errorf("failed writing output: %w", err)
	}

	return nil
}
