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
	commandsDir := filepath.Join(file, "../../../../")

	// Parse markdown
	cmds, err := commandsgen.ParseCommands()
	if err != nil {
		return fmt.Errorf("failed parsing markdown: %w", err)
	}

	// Generate docs
	files, err := commandsgen.GenerateDocsFiles(cmds)
	if err != nil {
		return err
	}

	// Write
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(commandsDir, file.FileName), file.Data, 0644); err != nil {
			return fmt.Errorf("failed writing file: %w", err)
		}
	}

	return nil
}
