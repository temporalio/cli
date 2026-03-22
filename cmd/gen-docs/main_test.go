package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenDocsMultipleInputs(t *testing.T) {
	outputDir := t.TempDir()

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"gen-docs",
		"-input", filepath.Join("..", "..", "internal", "temporalcli", "commands.yaml"),
		"-input", filepath.Join("..", "..", "cliext", "option-sets.yaml"),
		"-output", outputDir,
	}

	if err := run(); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no files were generated")
	}

	workflowPath := filepath.Join(outputDir, "workflow.mdx")
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		t.Fatal("workflow.mdx was not generated")
	}
}
