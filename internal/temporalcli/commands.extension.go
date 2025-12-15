package temporalcli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const (
	// extensionPrefix is the required prefix for all Temporal CLI extensions.
	extensionPrefix = "temporal-"
)

// lookupExtension searches for an extension executable from the given args.
// It performs most-specific to least-specific matching using exec.LookPath
// to find the executable on the PATH.
//
// Args starting with "-" are skipped when building extension names.
// Dashes within args are converted to underscores when building extension names.
// For example, given args ["foo", "--flag=value", "bar-baz"], it tries:
//
//	temporal-foo-bar_baz, temporal-foo
//
// If temporal-foo-bar_baz is found, returns (path, ["--flag=value"]).
// If temporal-foo is found, returns (path, ["--flag=value", "bar-baz"]).
func lookupExtension(args []string) (path string, extArgs []string) {
	// Collect indices of positional args (skip flags).
	var posIdx []int
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			posIdx = append(posIdx, i)
		}
	}

	if len(posIdx) == 0 {
		return "", nil
	}

	// Try matching from most specific to least specific.
	for n := len(posIdx); n > 0; n-- {
		parts := make([]string, n)
		for i := 0; i < n; i++ {
			parts[i] = strings.ReplaceAll(args[posIdx[i]], "-", "_")
		}

		path, err := exec.LookPath(extensionPrefix + strings.Join(parts, "-"))
		if err != nil {
			continue
		}

		// Return unmatched args: all flags + positional args after last match.
		lastMatch := posIdx[n-1]
		for i, arg := range args {
			if strings.HasPrefix(arg, "-") || i > lastMatch {
				extArgs = append(extArgs, arg)
			}
		}
		return path, extArgs
	}

	return "", nil
}

// tryExecuteExtension attempts to find and execute an extension for the given options.
// The bool return value indicates whether an extension was found and executed.
func tryExecuteExtension(ctx context.Context, opts *CommandOptions) (error, bool) {
	extPath, extArgs := lookupExtension(opts.Args)
	if extPath == "" {
		return nil, false
	}

	cmd := exec.CommandContext(ctx, extPath, extArgs...)
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	if err := cmd.Run(); err != nil {
		// Check for context cancellation or timeout first.
		if ctx.Err() != nil {
			return fmt.Errorf("program interrupted"), true
		}
		// No additional error handling here; the extension is expected to handle its own errors.
		if _, ok := err.(*exec.ExitError); ok {
			return nil, true
		}
		return fmt.Errorf("extension %s failed: %w", extPath, err), true
	}

	return nil, true
}
