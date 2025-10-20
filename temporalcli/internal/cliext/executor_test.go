package cliext

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteExtension tests basic extension execution scenarios.
func TestExecuteExtension(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T) (*Extension, []string, func())
		timeout          time.Duration
		expectedExitCode int
		expectError      bool
	}{
		{
			name: "successful execution",
			setup: func(t *testing.T) (*Extension, []string, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-test", "exit 0")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				require.Len(t, extensions, 1)
				return &extensions[0], []string{}, cleanup
			},
			expectedExitCode: 0,
			expectError:      false,
		},
		{
			name: "non-zero exit code",
			setup: func(t *testing.T) (*Extension, []string, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-test", "exit 42")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				require.Len(t, extensions, 1)
				return &extensions[0], []string{}, cleanup
			},
			expectedExitCode: 42,
			expectError:      false,
		},
		{
			name: "extension with arguments",
			setup: func(t *testing.T) (*Extension, []string, func()) {
				tempDir := t.TempDir()
				// Script that exits with the number of arguments
				createExecutable(t, tempDir, "temporal-test", "exit $#")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				require.Len(t, extensions, 1)
				return &extensions[0], []string{"arg1", "arg2", "arg3"}, cleanup
			},
			expectedExitCode: 3,
			expectError:      false,
		},
		{
			name: "nil extension returns error",
			setup: func(t *testing.T) (*Extension, []string, func()) {
				return nil, []string{}, func() {}
			},
			expectedExitCode: 1,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, args, cleanup := tt.setup(t)
			defer cleanup()

			ctx := context.Background()
			ioConfig := NewStdIOConfig()
			exitCode, err := ExecuteExtension(ctx, ext, args, ioConfig, tt.timeout)

			assert.Equal(t, tt.expectedExitCode, exitCode, "Exit code mismatch")

			if tt.expectError {
				assert.Error(t, err, "Expected error")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}
		})
	}
}

// TestExecuteExtensionWithOutput tests extension execution with output capture.
func TestExecuteExtensionWithOutput(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T) (*Extension, []string, func())
		stdin            string
		expectedStdout   string
		expectedExitCode int
		expectError      bool
	}{
		{
			name: "capture stdout",
			setup: func(t *testing.T) (*Extension, []string, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-test", "echo 'Hello from extension'")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				require.Len(t, extensions, 1)
				return &extensions[0], []string{}, cleanup
			},
			expectedStdout:   "Hello from extension",
			expectedExitCode: 0,
			expectError:      false,
		},
		{
			name: "capture arguments in output",
			setup: func(t *testing.T) (*Extension, []string, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-test", "echo \"Args: $@\"")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				require.Len(t, extensions, 1)
				return &extensions[0], []string{"foo", "bar"}, cleanup
			},
			expectedStdout:   "Args: foo bar",
			expectedExitCode: 0,
			expectError:      false,
		},
		{
			name: "capture exit code with output",
			setup: func(t *testing.T) (*Extension, []string, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-test", "echo 'Error occurred'; exit 1")
				cleanup := setupTestPATH(t, tempDir)
				extensions := DiscoverExtensions()
				require.Len(t, extensions, 1)
				return &extensions[0], []string{}, cleanup
			},
			expectedStdout:   "Error occurred",
			expectedExitCode: 1,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, args, cleanup := tt.setup(t)
			defer cleanup()

			ctx := context.Background()
			var stdin io.Reader
			if tt.stdin != "" {
				stdin = strings.NewReader(tt.stdin)
			}

			ioConfig, stdout, _ := NewCaptureIOConfig(stdin)
			exitCode, err := ExecuteExtension(ctx, ext, args, ioConfig, 0)

			assert.Equal(t, tt.expectedExitCode, exitCode, "Exit code mismatch")
			assert.Contains(t, stdout.String(), tt.expectedStdout, "Stdout mismatch")

			if tt.expectError {
				assert.Error(t, err, "Expected error")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}
		})
	}
}

// TestExecuteExtensionTimeout tests timeout handling.
func TestExecuteExtensionTimeout(t *testing.T) {
	tempDir := t.TempDir()
	// Create extension that sleeps for 5 seconds
	createExecutable(t, tempDir, "temporal-sleep", "sleep 5")
	cleanup := setupTestPATH(t, tempDir)
	defer cleanup()

	extensions := DiscoverExtensions()
	require.Len(t, extensions, 1)

	ctx := context.Background()
	timeout := 100 * time.Millisecond

	start := time.Now()
	ioConfig := NewStdIOConfig()
	exitCode, err := ExecuteExtension(ctx, &extensions[0], []string{}, ioConfig, timeout)
	elapsed := time.Since(start)

	// Should timeout and not take the full 5 seconds
	assert.Less(t, elapsed, 2*time.Second, "Should timeout quickly")
	assert.NotEqual(t, 0, exitCode, "Should have non-zero exit code on timeout")
	assert.NoError(t, err, "Timeout is not an error, just non-zero exit")
}

// TestTryExecuteExtension tests the integrated discover-match-execute flow.
func TestTryExecuteExtension(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T) func()
		commandTokens    []string
		expectedExitCode int
		expectError      bool
		errorContains    string
	}{
		{
			name: "successful execution of matched extension",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "exit 0")
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"foo"},
			expectedExitCode: 0,
			expectError:      false,
		},
		{
			name: "execute with fallback matching",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "exit 0")
				createExecutable(t, tempDir, "temporal-foo-bar", "exit 10")
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"foo", "bar", "baz"},
			expectedExitCode: 10,
			expectError:      false,
		},
		{
			name: "no extensions found",
			setup: func(t *testing.T) func() {
				return setupTestPATH(t, "")
			},
			commandTokens:    []string{"foo"},
			expectedExitCode: 1,
			expectError:      true,
			errorContains:    "no extensions found",
		},
		{
			name: "no matching extension",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "exit 0")
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"bar"},
			expectedExitCode: 1,
			expectError:      true,
			errorContains:    "no extension found for command",
		},
		{
			name: "multi-level extension with args",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				// Echo the number of arguments
				createExecutable(t, tempDir, "temporal-cloud-namespace-list", "exit $#")
				return setupTestPATH(t, tempDir)
			},
			commandTokens:    []string{"cloud", "namespace", "list", "--profile", "prod"},
			expectedExitCode: 2, // Should have 2 remaining args
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			ctx := context.Background()
			result := tryExecuteExtension(ctx, tt.commandTokens, 0)

			assert.Equal(t, tt.expectedExitCode, result.ExitCode, "Exit code mismatch")

			if tt.expectError {
				require.Error(t, result.Error, "Expected error")
				if tt.errorContains != "" {
					assert.Contains(t, result.Error.Error(), tt.errorContains, "Error message mismatch")
				}
			} else {
				assert.NoError(t, result.Error, "Unexpected error")
				require.NotNil(t, result.Extension, "Extension should be set")
			}
		})
	}
}

// TestExecuteExtensionRealWorldScenarios tests realistic extension scenarios.
func TestExecuteExtensionRealWorldScenarios(t *testing.T) {
	tests := []struct {
		name             string
		scriptContent    string
		commandTokens    []string
		expectedOutput   string
		expectedExitCode int
	}{
		{
			name:             "help flag",
			scriptContent:    `if [ "$1" == "--help" ]; then echo "Usage: temporal demo [options]"; exit 0; fi`,
			commandTokens:    []string{"demo", "--help"},
			expectedOutput:   "Usage: temporal demo",
			expectedExitCode: 0,
		},
		{
			name:             "version flag",
			scriptContent:    `if [ "$1" == "--version" ]; then echo "demo v1.0.0"; exit 0; fi`,
			commandTokens:    []string{"demo", "--version"},
			expectedOutput:   "demo v1.0.0",
			expectedExitCode: 0,
		},
		{
			name:             "multi-arg command",
			scriptContent:    `echo "Running with args: $@"`,
			commandTokens:    []string{"demo", "--flag1", "value1", "--flag2", "value2"},
			expectedOutput:   "Running with args:",
			expectedExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			createExecutable(t, tempDir, "temporal-demo", tt.scriptContent)
			cleanup := setupTestPATH(t, tempDir)
			defer cleanup()

			ctx := context.Background()
			result := tryExecuteExtension(ctx, tt.commandTokens, 0)

			assert.Equal(t, tt.expectedExitCode, result.ExitCode, "Exit code mismatch")
			assert.NoError(t, result.Error, "Unexpected error")
			require.NotNil(t, result.Extension, "Extension should be found")

			// Also test with output capture to verify output
			extensions := DiscoverExtensions()
			ext, remaining := FindExtension(extensions, tt.commandTokens)
			require.NotNil(t, ext)

			ioConfig, stdout, _ := NewCaptureIOConfig(nil)
			exitCode, err := ExecuteExtension(ctx, ext, remaining, ioConfig, 0)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedExitCode, exitCode)
			assert.Contains(t, stdout.String(), tt.expectedOutput, "Output mismatch")
		})
	}
}
