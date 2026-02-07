package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/temporalio/cli/internal/commandsgen"
)

// stringSlice implements flag.Value to support multiple -input flags
type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		outputDir  string
		inputFiles stringSlice
	)

	flag.Var(&inputFiles, "input", "Input YAML file (can be specified multiple times)")
	flag.StringVar(&outputDir, "output", ".", "Output directory for docs")
	flag.Parse()

	if len(inputFiles) == 0 {
		return fmt.Errorf("-input flag is required")
	}

	var yamlInputs [][]byte
	for _, inputFile := range inputFiles {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed reading input %s: %w", inputFile, err)
		}
		yamlInputs = append(yamlInputs, data)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed creating output directory: %w", err)
	}

	cmds, err := commandsgen.ParseCommands(yamlInputs...)
	if err != nil {
		return fmt.Errorf("failed parsing YAML: %w", err)
	}

	docs, err := commandsgen.GenerateDocsFiles(cmds)
	if err != nil {
		return fmt.Errorf("failed generating docs: %w", err)
	}

	for filename, content := range docs {
		filePath := filepath.Join(outputDir, filename+".mdx")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("failed writing %s: %w", filePath, err)
		}
	}

	return nil
}
