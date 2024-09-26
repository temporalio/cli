package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	// TODO: figure out how to structure generating multiple files
	b, err := commandsgen.GenerateDocsFiles(cmds)
	if err != nil {
		return err
	}

	// Write
	for filename, content := range b {
		filePath := filepath.Join(commandsDir, filename+".mdx")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("failed writing file: %w", err)
		}
	}

	// ATTEMPT: use cobra to parse commands and generate docs?
	// cctx, cancel, err := temporalcli.NewCommandContext(ctx, options)
	// if err != nil {
	// 	return err
	// }
	// defer cancel()

	// cctx := &temporalcli.CommandContext{}
	// cmd := temporalcli.NewTemporalCommand(cctx)
	// listFlags(&cmd.Command)

	return nil
}

// listFlags prints all flags of the TemporalActivityCompleteCommand.
func listFlags(cmd *cobra.Command) {
	fmt.Println("Listing all flags:")

	// Visit all local flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		fmt.Printf("Local Flag: --%s (Shorthand: -%s) - %s\n", flag.Name, flag.Shorthand, flag.Usage)
	})

	// Visit all persistent flags
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		fmt.Printf("Persistent Flag: %s\n\tshorthand: %s\n\tusage: %s\n", flag.Name, flag.Shorthand, flag.Usage)
	})
}
