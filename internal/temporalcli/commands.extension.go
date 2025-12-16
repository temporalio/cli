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

// tryExecuteExtension tries to execute an extension command if the command is not a built-in command.
// It returns an error if the extension command fails, and a boolean indicating whether an extension was executed.
func tryExecuteExtension(cctx *CommandContext, tcmd *TemporalCommand) (error, bool) {
	// Find the deepest matching built-in command and remaining args.
	//
	// E.g., "workflow diagram" -> foundCmd="workflow", remainingArgs=["diagram"].
	foundCmd, remainingArgs, findErr := tcmd.Command.Find(cctx.Options.Args)

	// Check if remaining args include positional args (not just flags).
	// If not, a built-in command fully handles this - no extension needed.
	hasPosArgs := slices.ContainsFunc(remainingArgs, isPosArg)
	if findErr == nil && !hasPosArgs {
		return nil, false
	}

	// Separate known CLI flags from extension args.
	//
	// E.g., ["--output", "json", "foo", "--bar"] -> cliFlags=["--output", "json"], extArgs=["foo", "--bar"]
	cliFlags, extArgs, splitErr := splitArgs(foundCmd, remainingArgs)

	// If there was an unknown flag before the extension name, reject it.
	if splitErr != nil {
		return splitErr, false
	}

	// If no positional args remain (potential extension names), let the built-in command handle it.
	if len(extArgs) == 0 {
		return nil, false
	}

	// Build command prefix from command path (excluding root "temporal").
	//
	// E.g., "temporal workflow" -> ["workflow"]
	//       "temporal workflow list" -> ["workflow", "list"]
	cmdPrefix := strings.Split(foundCmd.CommandPath(), " ")[1:]

	// Search for an extension executable.
	// Positional args are used to build the extension name.
	//
	// E.g., cmdPrefix=["workflow"], extArgs=["diagram", "arg1"] -> extPath="temporal-workflow-diagram", unmatchedArgs=["arg1"]
	extPath, unmatchedArgs := lookupExtension(cmdPrefix, extArgs)
	if extPath == "" {
		return nil, false
	}

	// Apply --command-timeout if set.
	// Parse errors are ignored since flag validation already happened earlier.
	_ = tcmd.Command.PersistentFlags().Parse(cctx.Options.Args)
	ctx := cctx.Context
	if timeout := tcmd.CommandTimeout.Duration(); timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Execute the extension command.
	cmd := exec.CommandContext(ctx, extPath, append(cliFlags, unmatchedArgs...)...)
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

func splitArgs(foundCmd *cobra.Command, args []string) (cliFlags, extArgs []string, err error) {
	// Build map of all flags and whether they take a value by walking up the command tree.
	flagTakesValue := map[string]bool{}
	for c := foundCmd; c != nil; c = c.Parent() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			// NoOptDefVal is only set for boolean flags and empty for flags requiring a value.
			takesValue := f.NoOptDefVal == ""
			if _, ok := flagTakesValue[f.Name]; !ok {
				flagTakesValue[f.Name] = takesValue
			}
			if f.Shorthand != "" {
				if _, ok := flagTakesValue[f.Shorthand]; !ok {
					flagTakesValue[f.Shorthand] = takesValue
				}
			}
		})
	}

	// Get flag normalizer to handle aliases.
	normalize := foundCmd.Flags().GetNormalizeFunc()

	// Split args into CLI flags (known flags) and extension args (positional args + unknown flags).
	// Unknown flags before the first positional arg are rejected.
	seenPosArg := false
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Positional arg: goes to extension args.
		if isPosArg(arg) {
			seenPosArg = true
			extArgs = append(extArgs, arg)
			continue
		}

		// Extract flag name, handling both --flag=value and --flag value forms.
		name, _, hasInline := strings.Cut(strings.TrimLeft(arg, "-"), "=")

		// Normalize the flag name to handle aliases.
		if normalize != nil {
			name = string(normalize(foundCmd.Flags(), name))
		}

		takesValue, isKnown := flagTakesValue[name]
		if !isKnown {
			if !seenPosArg {
				// Unknown flags before first positional arg are rejected.
				if err == nil {
					err = fmt.Errorf("unknown flag: --%s", name)
				}
				continue
			}
			// Unknown flags after first positional arg go to extension.
			extArgs = append(extArgs, arg)
			continue
		}

		// Known flag: goes to CLI flags.
		cliFlags = append(cliFlags, arg)

		// If flag takes a value and it's not inline (--flag=value), consume next arg.
		if takesValue && !hasInline && i+1 < len(args) {
			i++
			cliFlags = append(cliFlags, args[i])
		}
	}

	return cliFlags, extArgs, err
}

func lookupExtension(cmdPrefix, args []string) (extPath string, unmatchedArgs []string) {
	// Collect positional args for extension name lookup.
	// Dashes are converted to underscores so "foo bar-baz" finds "temporal-foo-bar_baz".
	// This avoids ambiguity since dashes separate command parts in the executable name.
	var posArgs []string
	for _, arg := range args {
		if isPosArg(arg) {
			posArgs = append(posArgs, strings.ReplaceAll(arg, extensionSeparator, argDashReplacement))
		}
	}

	parts := append(cmdPrefix, posArgs...)
	if len(parts) == 0 {
		return "", nil
	}

	// Try most-specific to least-specific.
	for n := len(parts); n > 0; n-- {
		path, err := exec.LookPath(extensionPrefix + strings.Join(parts[:n], extensionSeparator))
		if err != nil {
			continue
		}

		// Build unmatched args: all flags + positional args after matched count.
		matched := max(n-len(cmdPrefix), 0)
		posSeen := 0
		for _, arg := range args {
			if !isPosArg(arg) {
				unmatchedArgs = append(unmatchedArgs, arg)
			} else if posSeen++; posSeen > matched {
				unmatchedArgs = append(unmatchedArgs, arg)
			}
		}
		return path, unmatchedArgs
	}

	return "", nil
}

func isPosArg(arg string) bool {
	return !strings.HasPrefix(arg, "-")
}
