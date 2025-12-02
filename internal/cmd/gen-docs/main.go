package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/temporalio/cli/internal/commandsgen"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		outputDir string
		inputFile string
	)

	flag.StringVar(&inputFile, "input", "", "Input YAML file (required)")
	flag.StringVar(&outputDir, "output", ".", "Output directory for docs")
	flag.Parse()

	// Read input from file
	if inputFile == "" {
		return fmt.Errorf("-input flag is required")
	}
	yamlBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed reading input: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed creating output directory: %w", err)
	}

	// Parse YAML
	cmds, err := commandsgen.ParseCommands(yamlBytes)
	if err != nil {
		return fmt.Errorf("failed parsing YAML: %w", err)
	}

	// Generate docs
	docs, err := commandsgen.GenerateDocsFiles(cmds)
	if err != nil {
		return fmt.Errorf("failed generating docs: %w", err)
	}

	// Write files
	for filename, content := range docs {
		filePath := filepath.Join(outputDir, filename+".mdx")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("failed writing %s: %w", filePath, err)
		}
	}

	return nil
}
