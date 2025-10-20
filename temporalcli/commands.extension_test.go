package temporalcli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtensionIntegration tests that extensions are only used when commands are not found
func TestExtensionIntegration(t *testing.T) {
	tests := []struct {
		name           string
		setupExtension func(t *testing.T, dir string) string
		args           []string
		expectError    bool
		expectOutput   string
		expectExtRan   bool
	}{
		{
			name: "built-in command runs normally",
			args: []string{"--version"},
			setupExtension: func(t *testing.T, dir string) string {
				// No extension needed
				return ""
			},
			expectError:  false,
			expectExtRan: false,
		},
		{
			name: "extension runs for unknown command",
			args: []string{"myext", "arg1"},
			setupExtension: func(t *testing.T, dir string) string {
				extPath := filepath.Join(dir, "temporal-myext")
				content := "#!/bin/sh\necho 'Extension ran'\n"
				err := os.WriteFile(extPath, []byte(content), 0755)
				require.NoError(t, err)
				return dir
			},
			expectError:  false,
			expectOutput: "Extension ran",
			expectExtRan: true,
		},
		{
			name: "extension cannot override built-in workflow",
			args: []string{"workflow"},
			setupExtension: func(t *testing.T, dir string) string {
				extPath := filepath.Join(dir, "temporal-workflow")
				content := "#!/bin/sh\necho 'Extension workflow'\n"
				err := os.WriteFile(extPath, []byte(content), 0755)
				require.NoError(t, err)
				return dir
			},
			expectError:  true, // workflow requires subcommand
			expectExtRan: false,
		},
		{
			name: "extension can add workflow subcommand",
			args: []string{"workflow", "diagram"},
			setupExtension: func(t *testing.T, dir string) string {
				extPath := filepath.Join(dir, "temporal-workflow-diagram")
				content := "#!/bin/sh\necho 'Workflow diagram'\n"
				err := os.WriteFile(extPath, []byte(content), 0755)
				require.NoError(t, err)
				return dir
			},
			expectError:  false,
			expectOutput: "Workflow diagram",
			expectExtRan: true,
		},
		{
			name: "extension cannot override activity",
			args: []string{"activity"},
			setupExtension: func(t *testing.T, dir string) string {
				extPath := filepath.Join(dir, "temporal-activity")
				content := "#!/bin/sh\necho 'Extension activity'\n"
				err := os.WriteFile(extPath, []byte(content), 0755)
				require.NoError(t, err)
				return dir
			},
			expectError:  true, // activity requires subcommand
			expectExtRan: false,
		},
		{
			name: "no extension found for unknown command",
			args: []string{"nonexistent", "command"},
			setupExtension: func(t *testing.T, dir string) string {
				return dir
			},
			expectError:  true,
			expectExtRan: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for extensions
			tmpDir := t.TempDir()
			extDir := tt.setupExtension(t, tmpDir)

			// Set PATH to include extension directory
			oldPath := os.Getenv("PATH")
			if extDir != "" {
				os.Setenv("PATH", extDir+string(os.PathListSeparator)+oldPath)
			}
			defer os.Setenv("PATH", oldPath)

			// Capture output
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			// Track if Fail was called
			var failErr error
			failCalled := false

			// Execute command
			ctx := context.Background()
			Execute(ctx, CommandOptions{
				Args:   tt.args,
				Stdin:  os.Stdin,
				Stdout: stdout,
				Stderr: stderr,
				Fail: func(err error) {
					failErr = err
					failCalled = true
				},
			})

			// Check results
			if tt.expectError {
				assert.True(t, failCalled, "Expected Fail() to be called")
			} else {
				assert.False(t, failCalled, "Expected Fail() not to be called, but got error: %v", failErr)
			}

			if tt.expectOutput != "" {
				output := stdout.String() + stderr.String()
				assert.Contains(t, output, tt.expectOutput)
			}
		})
	}
}

// TestExtensionExitCode tests that extension exit codes are properly handled
func TestExtensionExitCode(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		exitCode    int
		expectError bool
	}{
		{
			name:        "extension succeeds with exit 0",
			exitCode:    0,
			expectError: false,
		},
		{
			name:        "extension fails with exit 1",
			exitCode:    1,
			expectError: true,
		},
		{
			name:        "extension fails with exit 42",
			exitCode:    42,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create extension that exits with specific code
			extPath := filepath.Join(tmpDir, "temporal-exitext")
			content := "#!/bin/sh\nexit " + string(rune('0'+tt.exitCode)) + "\n"
			err := os.WriteFile(extPath, []byte(content), 0755)
			require.NoError(t, err)

			// Set PATH
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
			defer os.Setenv("PATH", oldPath)

			// Track if Fail was called
			failCalled := false

			// Execute command
			ctx := context.Background()
			Execute(ctx, CommandOptions{
				Args:   []string{"exitext"},
				Stdin:  os.Stdin,
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Fail: func(err error) {
					failCalled = true
				},
			})

			if tt.expectError {
				assert.True(t, failCalled, "Expected Fail() to be called")
			} else {
				assert.False(t, failCalled, "Expected Fail() not to be called")
			}

			// Clean up for next test
			os.Remove(extPath)
		})
	}
}
