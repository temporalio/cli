package temporalcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	extensionPrefix    = "temporal-"
	extensionSeparator = "-" // separates command parts in extension name
	argDashReplacement = "_" // dashes in args are replaced to avoid ambiguity
)

// cliArgsToParseForExtension lists CLI flags that should be parsed (validated).
var cliArgsToParseForExtension = map[string]bool{
	"command-timeout": true,
}

type ExtensionNonZeroExit struct {
	*exec.ExitError
}

func (err ExtensionNonZeroExit) Unwrap() error {
	return err.ExitError
}

// tryExecuteExtension tries to execute an extension command if the command is not a built-in command.
// It returns an error if the extension command fails, and a boolean indicating whether an extension was executed.
func tryExecuteExtension(cctx *CommandContext, tcmd *TemporalCommand) (error, bool) {
	// Special commands like "help" and "__complete" should be set aside and delegated to an extension command that matches
	// the rest of the given arguments. "temporal help my-extension" should be rewritten to "temporal-my-extension help"
	// Some of these commands, like "__complete" used for shell completion, don't actually get registered by Cobra until
	// just before they're invoked, so we need to split out the delegatable commands before trying to Find() a matching
	// subcommand.
	delegatableCommands, nonDelegatedArgs := splitDelegatedCommands(cctx.Options.Args)

	// If a completion command (used for generating the script that invokes "__complete") already exists or has been
	// explicitly disabled, this will do nothing, but if neither of those cases are true, we want to make sure
	// this command is registered so extensions can't shadow it.
	tcmd.Command.InitDefaultCompletionCmd()

	// Find the deepest matching built-in command and remaining args.
	foundCmd, remainingArgs, findErr := tcmd.Command.Find(nonDelegatedArgs)

	// Cobra normally adds --help/-h before parsing, but extension dispatch
	// pre-parses flags before Cobra's execution path runs. We Initialize it so that
	// help is treated as a normal CLI flag instead of surfacing as pflag.ErrHelp from flag parsing.
	foundCmd.InitDefaultHelpFlag()

	// Group args into these lists:
	// - cliParseArgs: args to validate (subset of cliPassArgs)
	// - cliPassArgs: known CLI args to pass to extension
	// - extArgs: args to pass to extension and use for extension lookup
	cliParseArgs, cliPassArgs, extArgs := groupArgs(foundCmd, remainingArgs)

	// Check if remaining args include positional args (not just flags).
	// If not, a built-in command fully handles this - no extension needed.
	if findErr == nil && !slices.ContainsFunc(extArgs, isPosArg) {
		return nil, false
	}

	// Search for an extension executable.
	cmdPrefix := strings.Fields(foundCmd.CommandPath())

	extPath, extArgs := lookupExtension(cmdPrefix, extArgs)

	// Parse CLI args that need validation.
	if len(cliParseArgs) > 0 {
		if err := foundCmd.Flags().Parse(cliParseArgs); err != nil {
			return err, false
		}
	}

	if extPath == "" {
		return nil, false
	}

	if len(delegatableCommands) > 0 && isCompletionCommand(delegatableCommands[0]) && len(extArgs) == 0 {
		// __complete always expects at least one argument, the last of which is the current subcommand
		// or argument to expand, with an empty string matching all possibilities.
		// ["temporal", "__complete", "activity"] means this cli should return any subcommands and extentions that
		// match "activity", whereas ["temporal", "__complete", "activity", ""] means we should show what's available
		// on the activity subcommand. The same logic applies to extension commands, so even if we matched an extension,
		// if there are no further args, it's still this cli's responsibility to respond to the completion request.
		return nil, false
	}

	// Apply --command-timeout if set.
	ctx := cctx.Context
	if timeout := tcmd.CommandTimeout.Duration(); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	rebuiltArgs := slices.Concat(delegatableCommands, cliPassArgs, extArgs)

	cmd := exec.CommandContext(ctx, extPath, rebuiltArgs...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = cctx.Options.Stdin, cctx.Options.Stdout, cctx.Options.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("program interrupted"), true
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			return ExtensionNonZeroExit{exitError}, true
		}
		return fmt.Errorf("extension %s failed: %w", extPath, err), true
	}

	return nil, true
}

// splitDelegatedCommands separates out commands that should be delegated to an extension
// from the rest of the args given. These commands are inherently position-dependent, so they're
// only treated specially when they're at the start of the list of arguments.
func splitDelegatedCommands(args []string) ([]string, []string) {
	if len(args) == 0 {
		return args, args
	}

	if args[0] == "help" {
		// "help __complete" never delegates, whatever comes after, so we can just mark "help" as delegatable and see what matches
		return args[:1], args[1:]
	}

	if isCompletionCommand(args[0]) {
		if len(args) > 1 && args[1] == "help" {
			// "__complete help" is what happens when a user types "temporal help<TAB>", so it should delegate both. This allows
			// shell completion to display the available help topics available from an extension
			return args[:2], args[2:]
		}

		return args[:1], args[1:]
	}

	return []string{}, args
}

func isCompletionCommand(arg string) bool {
	return arg == cobra.ShellCompRequestCmd || arg == cobra.ShellCompNoDescRequestCmd
}

func groupArgs(foundCmd *cobra.Command, args []string) (cliParseArgs, cliPassArgs, extArgs []string) {
	seenPos := false
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if isPosArg(arg) {
			seenPos = true
			extArgs = append(extArgs, arg)
			continue
		}

		name, hasInline := parseFlagArg(arg)
		if f, takesValue := lookupFlag(foundCmd, name); f != nil {
			// Help flag after positional args should go to extArgs so it
			// gets forwarded to extensions (e.g. "temporal foo bar --help").
			// Before positional args it stays in cliPassArgs for Cobra to handle.
			if f.Name == "help" && seenPos {
				extArgs = append(extArgs, arg)
				continue
			}
			// Known CLI flag: goes to cliPassArgs.
			// Flags in cliArgsToParseForExtension also go to cliParseArgs.
			shouldParse := cliArgsToParseForExtension[f.Name]
			cliPassArgs = append(cliPassArgs, arg)
			if shouldParse {
				cliParseArgs = append(cliParseArgs, arg)
			}
			if takesValue && !hasInline && i+1 < len(args) {
				i++
				cliPassArgs = append(cliPassArgs, args[i])
				if shouldParse {
					cliParseArgs = append(cliParseArgs, args[i])
				}
			}
		} else {
			// Unknown flag: before first positional goes to cliParseArgs (to fail validation),
			// after first positional goes to extArgs (passed to extension).
			if seenPos {
				extArgs = append(extArgs, arg)
			} else {
				cliParseArgs = append(cliParseArgs, arg)
			}
		}
	}
	return
}

func isPosArg(arg string) bool {
	return !strings.HasPrefix(arg, "-")
}

// parseFlagArg extracts the flag name from a flag argument.
// Handles both --flag=value and --flag forms, returning the name and whether it has an inline value.
func parseFlagArg(arg string) (name string, hasInline bool) {
	name, _, hasInline = strings.Cut(strings.TrimLeft(arg, "-"), "=")
	return
}

// lookupFlag finds a flag by name on cmd and all parents.
// It resolves aliases and considers shorthand flags.
func lookupFlag(cmd *cobra.Command, name string) (*pflag.Flag, bool) {
	if normalize := cmd.Flags().GetNormalizeFunc(); normalize != nil {
		name = string(normalize(cmd.Flags(), name))
	}
	for c := cmd; c != nil; c = c.Parent() {
		if f := c.Flags().Lookup(name); f != nil {
			return f, f.NoOptDefVal == ""
		}
		if len(name) == 1 {
			if f := c.Flags().ShorthandLookup(name); f != nil {
				return f, f.NoOptDefVal == ""
			}
		}
	}
	return nil, false
}

// lookupExtension finds an extension executable and returns its path along with
// extArgs with matched positional args removed.
func lookupExtension(cmdPrefix, extArgs []string) (string, []string) {
	// Extract positional args from extArgs until we hit an unknown flag.
	// We stop at unknown flags because we can't tell if subsequent args are flag values or positionals.
	var posArgs []string
	for _, arg := range extArgs {
		if !isPosArg(arg) {
			break
		}
		posArgs = append(posArgs, arg)
	}

	// Try most-specific to least-specific.
	parts := append(cmdPrefix, posArgs...)
	for n := len(parts); n > len(cmdPrefix); n-- {
		binName := extensionCommandToBinary(parts[:n])
		if fullPath, _ := isExecutable(binName); fullPath != "" {
			// Remove matched positionals from extArgs (they come first).
			matched := n - len(cmdPrefix)
			return fullPath, extArgs[matched:]
		}
	}

	return "", extArgs
}

// discoverExtensions scans the PATH for executables with the "temporal-" prefix
// and returns their commands (without the prefix) mapped to the executable path
func discoverExtensions() map[string]string {
	extensions := make(map[string]string)

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()

			// Look for extensions.
			if !strings.HasPrefix(name, extensionPrefix) {
				continue
			}

			// Check if the file is executable.
			fullPath, baseName := isExecutable(filepath.Join(dir, name))
			if fullPath == "" {
				continue
			}

			path := extensionBinaryToCommandPath(baseName)
			key := strings.Join(path, " ")
			if extensions[key] != "" {
				continue
			}

			extensions[key] = filepath.Join(dir, entry.Name())
		}
	}
	return extensions
}

// isExecutable checks if a file or command is executable.
// On Windows, it validates PATHEXT suffix and strips it from the base name.
// Returns the full path and base name (without Windows extension suffix).
func isExecutable(name string) (fullPath, baseName string) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", ""
	}

	base := filepath.Base(path)
	if runtime.GOOS == "windows" {
		pathext := os.Getenv("PATHEXT")
		if pathext == "" {
			pathext = ".exe;.bat"
		}
		lower := strings.ToLower(base)
		for ext := range strings.SplitSeq(strings.ToLower(pathext), ";") {
			if ext != "" && strings.HasSuffix(lower, ext) {
				return path, base[:len(base)-len(ext)]
			}
		}
		return "", ""
	}
	return path, base
}

// extensionBinaryToCommandPath converts a binary name to command path.
// Underscores are converted to dashes.
// For example: "temporal-foo-bar_baz" -> ["temporal", "foo", "bar-baz"]
func extensionBinaryToCommandPath(binary string) []string {
	path := strings.Split(binary, extensionSeparator)
	for i, p := range path {
		path[i] = strings.ReplaceAll(p, argDashReplacement, extensionSeparator)
	}
	return path
}

// extensionCommandToBinary converts command path to a binary name.
// Dashes in path are converted to underscores.
// For example: ["temporal", "foo", "bar-baz"] -> "temporal-foo-bar_baz"
func extensionCommandToBinary(path []string) string {
	converted := make([]string, len(path))
	for i, p := range path {
		converted[i] = strings.ReplaceAll(p, extensionSeparator, argDashReplacement)
	}
	return strings.Join(converted, extensionSeparator)
}
