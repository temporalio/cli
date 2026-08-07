package temporalcli

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
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
// so they appear in shell completion and the default help output. It filters extensions
// based on the current command's path in the hierarchy.
func registerExtensionCommands(cmd *cobra.Command) {
	cmdPath := strings.Fields(cmd.CommandPath())

	// When built-in subcommands are nested under other subcommands (e.g. `temporal activity cancel`),
	// they're guaranteed to have a defined parent (`temporal activity`, which has the parent `temporal`).
	// Extension subcommands can also be nested, but when `temporal foo bar` is created via an executable
	// named temporal-foo-bar, there's no guarantee that the `temporal foo` command exists. For the full
	// command to show up in help and completion, placeholders must exist at every level.
	extensionsAndExecutables := discoverExtensions()

	extensionKeys := maps.Keys(extensionsAndExecutables)

	// Shorter command paths first ensures the paths shown in the short description of placeholder commands
	// point to the closest match
	slices.SortFunc(extensionKeys, func(a, b string) int {
		return cmp.Compare(strings.Count(a, " "), strings.Count(b, " "))
	})
	for _, extKey := range extensionKeys {
		ext := strings.Split(extKey, " ")
		// Extension must be deeper than current command and share the same prefix
		if len(ext) <= len(cmdPath) || !slices.Equal(ext[:len(cmdPath)], cmdPath) {
			continue
		}

		extPath := ext[len(cmdPath):]

		parent := cmd
		executablePath := extensionsAndExecutables[extKey]

		for i, nextPart := range extPath {
			if found, _, _ := parent.Find([]string{nextPart}); found != parent {
				// Because we order extensions by depth, we can trust that any command that already exists already
				// has the most correct definition for its depth.
				parent = found
				continue
			}

			var short string

			if i == len(extPath)-1 {
				short = fmt.Sprintf("An extension command located at %s", executablePath)
			} else {
				short = fmt.Sprintf("Extension commands under %s", strings.Join(ext[:len(cmdPath)+i+1], " "))
			}

			newCmd := &cobra.Command{
				Use: nextPart,
				// Short descriptions must be unique at a given level because otherwise shell completion will
				// group all commands with the same description on the same line
				Short:              short,
				DisableFlagParsing: true,
				Run:                func(*cobra.Command, []string) {},
			}
			parent.AddCommand(newCmd)
			parent = newCmd
		}

	}
}
