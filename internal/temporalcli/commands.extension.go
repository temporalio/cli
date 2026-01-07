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

// tryExecuteExtension tries to execute an extension command if the command is not a built-in command.
// It returns an error if the extension command fails, and a boolean indicating whether an extension was executed.
func tryExecuteExtension(cctx *CommandContext, tcmd *TemporalCommand) (error, bool) {
	// Find the deepest matching built-in command and remaining args.
	foundCmd, remainingArgs, findErr := tcmd.Command.Find(cctx.Options.Args)

	// Check if remaining args include positional args (not just flags).
	// If not, a built-in command fully handles this - no extension needed.
	hasPosArgs := slices.ContainsFunc(remainingArgs, isPosArg)
	if findErr == nil && !hasPosArgs {
		return nil, false
	}

	// Group args into these lists:
	// - cliParseArgs: args to validate (subset of cliPassArgs)
	// - cliPassArgs: known CLI args to pass to extension
	// - extArgs: args to pass to extension and use for extension lookup
	cliParseArgs, cliPassArgs, extArgs := groupArgs(foundCmd, remainingArgs)

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

	// Apply --command-timeout if set.
	ctx := cctx.Context
	if timeout := tcmd.CommandTimeout.Duration(); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, extPath, append(cliPassArgs, extArgs...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = cctx.Options.Stdin, cctx.Options.Stdout, cctx.Options.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("program interrupted"), true
		}
		if _, ok := err.(*exec.ExitError); ok {
			return nil, true
		}
		return fmt.Errorf("extension %s failed: %w", extPath, err), true
	}

	return nil, true
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
// and returns their command parts (without the prefix).
func discoverExtensions() [][]string {
	var extensions [][]string
	seen := make(map[string]bool)

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
			key := strings.Join(path, "/")
			if seen[key] {
				continue
			}

			seen[key] = true
			extensions = append(extensions, path)
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
