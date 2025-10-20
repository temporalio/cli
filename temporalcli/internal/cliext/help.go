package cliext

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// FormatExtensionsForHelp formats discovered extensions for display in help output.
//
// Each extension is formatted to show its command path with spaces instead of dashes/underscores.
// For example:
//   - temporal-foo-bar → "foo bar"
//   - temporal-workflow-show_diagram → "workflow show-diagram"
//   - temporal-cloud-namespace-list → "cloud namespace list"
//
// The returned strings are sorted alphabetically.
func FormatExtensionsForHelp(extensions []Extension) []string {
	if len(extensions) == 0 {
		return nil
	}

	formatted := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		formatted = append(formatted, FormatExtensionCommand(ext.CommandTokens))
	}

	sort.Strings(formatted)
	return formatted
}

// FormatExtensionCommand formats command tokens as a space-separated string.
//
// Example:
//   - ["foo", "bar"] → "foo bar"
//   - ["cloud", "namespace", "list"] → "cloud namespace list"
func FormatExtensionCommand(tokens []string) string {
	return strings.Join(tokens, " ")
}

// ShouldShowHelpForExtension determines if the given command tokens represent
// a help request that should be routed to an extension.
//
// Returns true if:
//   - First token is "help" and there are additional tokens
//   - The remaining tokens match an extension
//
// Example:
//   - ["help", "foo"] → should check if "foo" extension exists
//   - ["help", "foo", "bar"] → should check if "foo bar" extension exists
func ShouldShowHelpForExtension(commandTokens []string) bool {
	return len(commandTokens) >= 2 && commandTokens[0] == "help"
}

// ExtractHelpCommand extracts the extension command from a help invocation.
//
// If the command starts with "help", returns the remaining tokens.
// Otherwise, returns the original tokens.
//
// Examples:
//   - ["help", "foo", "bar"] → ["foo", "bar"]
//   - ["foo", "bar"] → ["foo", "bar"]
func ExtractHelpCommand(commandTokens []string) []string {
	if len(commandTokens) >= 2 && commandTokens[0] == "help" {
		return commandTokens[1:]
	}
	return commandTokens
}

// TryShowExtensionHelp attempts to show help for an extension.
//
// This is a helper function that:
// 1. Discovers available extensions
// 2. Checks if the command matches an extension
// 3. If found, executes the extension with --help flag
// 4. Returns true if help was shown, false otherwise
//
// Example usage:
//   if TryShowExtensionHelp(ctx, []string{"help", "foo", "bar"}) {
//       // Help was shown, exit
//       return
//   }
//   // No extension found, show built-in help
func TryShowExtensionHelp(ctx context.Context, commandTokens []string, timeout time.Duration) *ExecuteExtensionResult {
	if !ShouldShowHelpForExtension(commandTokens) {
		return &ExecuteExtensionResult{
			ExitCode: 1,
			Error:    fmt.Errorf("not a help command"),
		}
	}

	// Extract the extension command from "help <extension>"
	extensionTokens := ExtractHelpCommand(commandTokens)

	// Discover extensions
	extensions := DiscoverExtensions()
	if len(extensions) == 0 {
		return &ExecuteExtensionResult{
			ExitCode: 1,
			Error:    fmt.Errorf("no extensions found"),
		}
	}

	// Find matching extension
	ext, _ := FindExtension(extensions, extensionTokens)
	if ext == nil {
		return &ExecuteExtensionResult{
			ExitCode: 1,
			Error:    fmt.Errorf("no extension found for command: %v", extensionTokens),
		}
	}

	// Execute the extension with --help flag
	ioConfig := NewStdIOConfig()
	exitCode, err := ExecuteExtension(ctx, ext, []string{"--help"}, ioConfig, timeout)

	return &ExecuteExtensionResult{
		Extension: ext,
		ExitCode:  exitCode,
		Error:     err,
	}
}
