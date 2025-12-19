package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/temporalio/cli/internal/commandsgen"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func run() error {
	var (
		pkg         string
		contextType string
		inputFiles  stringSlice
	)

	flag.StringVar(&pkg, "pkg", "main", "Package name for generated code")
	flag.StringVar(&contextType, "context", "*CommandContext", "Context type for generated code")
	flag.Var(&inputFiles, "input", "Input YAML file (can be specified multiple times)")
	flag.Parse()

	// Read input from file
	if len(inputFiles) == 0 {
		return fmt.Errorf("-input flag is required")
	}

	yamlDataList := make([][]byte, len(inputFiles))
	for i, inputFile := range inputFiles {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed reading input %s: %w", inputFile, err)
		}
		yamlDataList[i] = data
	}

	// Parse YAML
	cmds, err := commandsgen.ParseCommands(yamlDataList...)
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
