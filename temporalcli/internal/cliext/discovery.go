package cliext

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// ExtensionPrefix is the required prefix for all Temporal CLI extensions.
	ExtensionPrefix = "temporal-"
)

// Extension represents a discovered CLI extension.
type Extension struct {
	// Name is the base name of the extension executable (e.g., "temporal-foo").
	Name string

	// Path is the full path to the extension executable.
	Path string

	// CommandTokens are the command parts extracted from the name.
	// For "temporal-foo-bar", this would be ["foo", "bar"].
	CommandTokens []string
}

// DiscoverExtensions scans the PATH environment variable and returns all
// discovered Temporal CLI extensions.
//
// Extensions are executables with the "temporal-" prefix. The function:
//   - Scans all directories in PATH
//   - Filters for files starting with "temporal-"
//   - Checks that files are executable
//   - Returns the first occurrence of each extension name (PATH order matters)
//
// Example:
//
//	extensions := cliext.DiscoverExtensions()
//	for _, ext := range extensions {
//	    fmt.Printf("Found extension: %s at %s\n", ext.Name, ext.Path)
//	}
func DiscoverExtensions() []Extension {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil
	}

	// Parse PATH based on OS
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}

	directories := strings.Split(pathEnv, pathSeparator)

	// Track discovered extension names to handle duplicates (first in PATH wins)
	seen := make(map[string]bool)
	var extensions []Extension

	for _, dir := range directories {
		if dir == "" {
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Skip directories we can't read (permission issues, doesn't exist, etc.)
			continue
		}

		for _, entry := range entries {
			name := entry.Name()

			// Check for temporal- prefix
			if !strings.HasPrefix(name, ExtensionPrefix) {
				continue
			}

			// Skip if we've already seen this extension name
			if seen[name] {
				continue
			}

			fullPath := filepath.Join(dir, name)

			// Check if file is executable
			if !isExecutable(fullPath) {
				continue
			}

			// Mark as seen and add to results
			seen[name] = true
			extensions = append(extensions, Extension{
				Name:          name,
				Path:          fullPath,
				CommandTokens: parseCommandTokens(name),
			})
		}
	}

	return extensions
}

// isExecutable checks if a file is executable.
// On Unix-like systems, this checks the executable permission bit.
// On Windows, this checks if the file exists (Windows uses file extensions).
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// On Windows, if the file exists and is not a directory, consider it potentially executable
	// Windows uses file extensions (.exe, .bat, .cmd) to determine executability
	if runtime.GOOS == "windows" {
		return !info.IsDir()
	}

	// On Unix-like systems, check the executable bit
	mode := info.Mode()
	return !mode.IsDir() && mode.Perm()&0111 != 0
}

// parseCommandTokens extracts command tokens from an extension name.
// For example:
//   - "temporal-foo" → ["foo"]
//   - "temporal-foo-bar" → ["foo", "bar"]
//   - "temporal-foo-bar_baz" → ["foo", "bar-baz"]
//
// Naming rules:
//   - Remove the "temporal-" prefix
//   - Split on "-" to get subcommands
//   - Convert "_" to "-" within each token (for commands with dashes)
func parseCommandTokens(name string) []string {
	// Remove "temporal-" prefix
	if !strings.HasPrefix(name, ExtensionPrefix) {
		return nil
	}

	withoutPrefix := strings.TrimPrefix(name, ExtensionPrefix)
	if withoutPrefix == "" {
		return nil
	}

	// Split on "-" to get tokens
	tokens := strings.Split(withoutPrefix, "-")

	// Convert "_" back to "-" in each token
	// This allows "temporal-foo-bar_baz" to represent "temporal foo bar-baz"
	for i, token := range tokens {
		tokens[i] = strings.ReplaceAll(token, "_", "-")
	}

	return tokens
}

// FindExtension searches for an extension by command tokens.
// It performs most-specific to least-specific matching.
//
// For example, given command tokens ["foo", "bar", "baz"], it will try to match:
//  1. temporal-foo-bar-baz
//  2. temporal-foo-bar
//  3. temporal-foo
//
// Returns the matched extension and remaining arguments, or nil if no match.
func FindExtension(extensions []Extension, commandTokens []string) (*Extension, []string) {
	if len(commandTokens) == 0 {
		return nil, nil
	}

	// Try matching from most specific to least specific
	for i := len(commandTokens); i > 0; i-- {
		tokensToMatch := commandTokens[:i]
		remaining := commandTokens[i:]

		// Try to find an extension matching these tokens
		for idx := range extensions {
			ext := &extensions[idx]
			if matchesTokens(ext.CommandTokens, tokensToMatch) {
				return ext, remaining
			}
		}
	}

	return nil, nil
}

// matchesTokens checks if extension tokens match the given command tokens.
// Also handles the underscore/dash equivalence.
func matchesTokens(extTokens, cmdTokens []string) bool {
	if len(extTokens) != len(cmdTokens) {
		return false
	}

	for i := range extTokens {
		// Normalize both for comparison (treat _ and - as equivalent)
		extNorm := strings.ReplaceAll(extTokens[i], "_", "-")
		cmdNorm := strings.ReplaceAll(cmdTokens[i], "_", "-")

		if extNorm != cmdNorm {
			return false
		}
	}

	return true
}
