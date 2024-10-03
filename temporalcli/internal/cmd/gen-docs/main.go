package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/temporalio/cli/temporalcli/commandsgen"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Get commands dir
	_, file, _, _ := runtime.Caller(0)
	commandsDir := filepath.Join(file, "../../../../docs/")

	// Parse markdown
	cmds, err := commandsgen.ParseCommands()
	if err != nil {
		return fmt.Errorf("failed parsing markdown: %w", err)
	}

	// Generate docs
	b, err := commandsgen.GenerateDocsFiles(cmds)
	if err != nil {
		return err
	}

	// Write
	for filename, content := range b {
		filePath := filepath.Join(commandsDir, filename+".md")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("failed writing file: %w", err)
		}
	}

	return nil
}
