package cliext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatExtensionsForHelp tests formatting extensions for help display.
func TestFormatExtensionsForHelp(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(t *testing.T) ([]Extension, func())
		expectedFormatted  []string
	}{
		{
			name: "single simple extension",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "exit 0")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				return extensions, cleanup
			},
			expectedFormatted: []string{"foo"},
		},
		{
			name: "multiple extensions",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "exit 0")
				createExecutable(t, tempDir, "temporal-bar", "exit 0")
				createExecutable(t, tempDir, "temporal-baz", "exit 0")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				return extensions, cleanup
			},
			expectedFormatted: []string{"bar", "baz", "foo"}, // Sorted
		},
		{
			name: "multi-level extensions",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo-bar", "exit 0")
				createExecutable(t, tempDir, "temporal-cloud-namespace-list", "exit 0")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				return extensions, cleanup
			},
			expectedFormatted: []string{"cloud namespace list", "foo bar"},
		},
		{
			name: "extension with underscores",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-workflow-show_diagram", "exit 0")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				return extensions, cleanup
			},
			expectedFormatted: []string{"workflow show-diagram"},
		},
		{
			name: "complex real-world scenario",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-mycompany-login", "exit 0")
				createExecutable(t, tempDir, "temporal-workflow-show_diagram", "exit 0")
				createExecutable(t, tempDir, "temporal-cloud-namespace-list", "exit 0")
				createExecutable(t, tempDir, "temporal-cloud-namespace-create", "exit 0")
				createExecutable(t, tempDir, "temporal-foo", "exit 0")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				return extensions, cleanup
			},
			expectedFormatted: []string{
				"cloud namespace create",
				"cloud namespace list",
				"foo",
				"mycompany login",
				"workflow show-diagram",
			},
		},
		{
			name: "empty extension list",
			setup: func(t *testing.T) ([]Extension, func()) {
				return []Extension{}, func() {}
			},
			expectedFormatted: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extensions, cleanup := tt.setup(t)
			defer cleanup()

			formatted := FormatExtensionsForHelp(extensions)
			assert.Equal(t, tt.expectedFormatted, formatted)
		})
	}
}

// TestFormatExtensionCommand tests formatting command tokens.
func TestFormatExtensionCommand(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		expected string
	}{
		{
			name:     "single token",
			tokens:   []string{"foo"},
			expected: "foo",
		},
		{
			name:     "two tokens",
			tokens:   []string{"foo", "bar"},
			expected: "foo bar",
		},
		{
			name:     "three tokens",
			tokens:   []string{"cloud", "namespace", "list"},
			expected: "cloud namespace list",
		},
		{
			name:     "token with dash",
			tokens:   []string{"workflow", "show-diagram"},
			expected: "workflow show-diagram",
		},
		{
			name:     "empty tokens",
			tokens:   []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatExtensionCommand(tt.tokens)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldShowHelpForExtension tests detecting help requests.
func TestShouldShowHelpForExtension(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		expected bool
	}{
		{
			name:     "help with extension",
			tokens:   []string{"help", "foo"},
			expected: true,
		},
		{
			name:     "help with multi-level extension",
			tokens:   []string{"help", "foo", "bar"},
			expected: true,
		},
		{
			name:     "just help",
			tokens:   []string{"help"},
			expected: false,
		},
		{
			name:     "not help",
			tokens:   []string{"foo", "bar"},
			expected: false,
		},
		{
			name:     "empty tokens",
			tokens:   []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldShowHelpForExtension(tt.tokens)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractHelpCommand tests extracting extension command from help invocation.
func TestExtractHelpCommand(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		expected []string
	}{
		{
			name:     "help with single extension",
			tokens:   []string{"help", "foo"},
			expected: []string{"foo"},
		},
		{
			name:     "help with multi-level extension",
			tokens:   []string{"help", "foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "not help command",
			tokens:   []string{"foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "just help",
			tokens:   []string{"help"},
			expected: []string{"help"},
		},
		{
			name:     "empty tokens",
			tokens:   []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractHelpCommand(tt.tokens)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTryShowExtensionHelp tests showing help for extensions.
func TestTryShowExtensionHelp(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T) func()
		commandTokens    []string
		expectedExitCode int
		expectError      bool
		expectedOutput   string // For output verification
	}{
		{
			name: "show help for existing extension",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				script := `if [ "$1" == "--help" ]; then echo "Usage: temporal demo"; exit 0; fi`
				createExecutable(t, tempDir, "temporal-demo", script)
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"help", "demo"},
			expectedExitCode: 0,
			expectError:      false,
		},
		{
			name: "show help for multi-level extension",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				script := `if [ "$1" == "--help" ]; then echo "Usage: temporal foo bar"; exit 0; fi`
				createExecutable(t, tempDir, "temporal-foo-bar", script)
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"help", "foo", "bar"},
			expectedExitCode: 0,
			expectError:      false,
		},
		{
			name: "non-existent extension",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "exit 0")
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"help", "nonexistent"},
			expectedExitCode: 1,
			expectError:      true,
		},
		{
			name: "not a help command",
			setup: func(t *testing.T) func() {
				return setupTestPATH(t, "")
			},
			commandTokens:    []string{"foo", "bar"},
			expectedExitCode: 1,
			expectError:      true,
		},
		{
			name: "help without extension name",
			setup: func(t *testing.T) func() {
				return setupTestPATH(t, "")
			},
			commandTokens:    []string{"help"},
			expectedExitCode: 1,
			expectError:      true,
		},
		{
			name: "extension without help support",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				// Extension that doesn't handle --help
				createExecutable(t, tempDir, "temporal-demo", "echo 'Running demo'")
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"help", "demo"},
			expectedExitCode: 0,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			ctx := context.Background()
			result := TryShowExtensionHelp(ctx, tt.commandTokens, 0)

			assert.Equal(t, tt.expectedExitCode, result.ExitCode, "Exit code mismatch")

			if tt.expectError {
				require.Error(t, result.Error, "Expected error")
			} else {
				assert.NoError(t, result.Error, "Unexpected error")
				require.NotNil(t, result.Extension, "Extension should be set")
			}
		})
	}
}

// TestTryShowExtensionHelpWithOutput tests help output capture.
func TestTryShowExtensionHelpWithOutput(t *testing.T) {
	tests := []struct {
		name           string
		scriptContent  string
		commandTokens  []string
		expectedOutput string
	}{
		{
			name: "basic help output",
			scriptContent: `if [ "$1" == "--help" ]; then
  echo "Usage: temporal demo [options]"
  echo "  --flag1   Description of flag1"
  exit 0
fi`,
			commandTokens:  []string{"help", "demo"},
			expectedOutput: "Usage: temporal demo",
		},
		{
			name: "multi-line help",
			scriptContent: `if [ "$1" == "--help" ]; then
  echo "temporal-foo - A demo extension"
  echo ""
  echo "Usage:"
  echo "  temporal foo [command]"
  exit 0
fi`,
			commandTokens:  []string{"help", "foo"},
			expectedOutput: "temporal-foo - A demo extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			extName := "temporal-" + tt.commandTokens[1]
			createExecutable(t, tempDir, extName, tt.scriptContent)
			cleanup := setupTestPATH(t, tempDir)
			defer cleanup()

			ctx := context.Background()
			extensions := DiscoverExtensions()
			ext, _ := FindExtension(extensions, tt.commandTokens[1:])
			require.NotNil(t, ext)

			ioConfig, stdout, _ := NewCaptureIOConfig(nil)
			exitCode, err := ExecuteExtension(ctx, ext, []string{"--help"}, ioConfig, 0)

			require.NoError(t, err)
			assert.Equal(t, 0, exitCode)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
		})
	}
}
