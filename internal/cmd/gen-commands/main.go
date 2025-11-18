package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/temporalio/cli/cliext/commandsgen"
	"github.com/temporalio/cli/internal/temporalcli"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Get commands dir
	_, file, _, _ := runtime.Caller(0)
	genCommandsDir := filepath.Dir(file)
	commandsDir := filepath.Join(genCommandsDir, "../../temporalcli")

	// Parse YAML
	cmds, err := commandsgen.ParseCommands(temporalcli.CommandsYAML())
	if err != nil {
		return fmt.Errorf("failed parsing YAML: %w", err)
	}

	// Generate code
	b, err := commandsgen.GenerateCommandsCode("temporalcli", "*CommandContext", cmds)
	if err != nil {
		return fmt.Errorf("failed generating code: %w", err)
	}

	// Write
	if err := os.WriteFile(filepath.Join(commandsDir, "commands.gen.go"), b, 0644); err != nil {
		return fmt.Errorf("failed writing file: %w", err)
	}
	return nil
}
