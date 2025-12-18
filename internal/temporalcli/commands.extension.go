package temporalcli

import (
	"context"
	"fmt"
	"os/exec"
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
	cmdPrefix := strings.Split(foundCmd.CommandPath(), " ")[1:]
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
		// Dashes are converted to underscores so "foo bar-baz" finds "temporal-foo-bar_baz".
		posArgs = append(posArgs, strings.ReplaceAll(arg, extensionSeparator, argDashReplacement))
	}

	// Try most-specific to least-specific.
	parts := append(cmdPrefix, posArgs...)
	for n := len(parts); n > len(cmdPrefix); n-- {
		path, err := exec.LookPath(extensionPrefix + strings.Join(parts[:n], extensionSeparator))
		if err != nil {
			continue
		}
		// Remove matched positionals from extArgs (they come first).
		matched := n - len(cmdPrefix)
		return path, extArgs[matched:]
	}

	return "", extArgs
}
