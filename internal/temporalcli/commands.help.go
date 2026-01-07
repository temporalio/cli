package temporalcli

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// customizeHelpCommand adds the --all/-a flag to Cobra's built-in help command
// and customizes its behavior to include extensions when the flag is set.
func customizeHelpCommand(rootCmd *cobra.Command) {
	// Ensure the default help command is initialized
	rootCmd.InitDefaultHelpCmd()

	// Find the help command
	var helpCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "help" {
			helpCmd = c
			break
		}
	}
	if helpCmd == nil {
		return
	}

	// Add --all/-a flag
	var showAll bool
	helpCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all commands including extensions found in PATH.")

	// Store the original help function
	originalRun := helpCmd.Run

	// Override the run function
	helpCmd.Run = func(cmd *cobra.Command, args []string) {
		// Find target command
		targetCmd := rootCmd
		if len(args) > 0 {
			if found, _, err := rootCmd.Find(args); err == nil {
				targetCmd = found
			}
		}

		// If --all is set, register extensions as commands before showing help
		if showAll {
			registerExtensionCommands(targetCmd)
		}

		// Run original help
		originalRun(cmd, args)
	}
}

// registerExtensionCommands adds discovered extensions as placeholder commands
// so they appear in the default help output. It filters extensions based on
// the current command's path in the hierarchy.
func registerExtensionCommands(cmd *cobra.Command) {
	cmdPath := strings.Fields(cmd.CommandPath())
	seen := make(map[string]bool)

	for _, ext := range discoverExtensions() {
		// Extension must be deeper than current command and share the same prefix
		if len(ext) <= len(cmdPath) || !slices.Equal(ext[:len(cmdPath)], cmdPath) {
			continue
		}

		// Get the next level command name
		nextPart := ext[len(cmdPath)]

		// Skip if already added
		if seen[nextPart] {
			continue
		}

		// Skip if a built-in command exists
		if found, _, _ := cmd.Find([]string{nextPart}); found != cmd {
			continue
		}

		seen[nextPart] = true
		cmd.AddCommand(&cobra.Command{
			Use:                nextPart,
			DisableFlagParsing: true,
			Run:                func(*cobra.Command, []string) {},
		})
	}
}
