package main

import (
	"flag"
	"fmt"
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
		inputFile   string
	)

	flag.StringVar(&pkg, "pkg", "main", "Package name for generated code")
	flag.StringVar(&contextType, "context", "*CommandContext", "Context type for generated code")
	flag.StringVar(&inputFile, "input", "", "Input YAML file (required)")
	flag.Parse()

	// Read input from file
	if inputFile == "" {
		return fmt.Errorf("-input flag is required")
	}
	yamlBytes, err := os.ReadFile(inputFile)
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
