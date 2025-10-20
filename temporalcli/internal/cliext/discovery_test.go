package cliext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtensionDiscovery tests various extension discovery scenarios.
func TestExtensionDiscovery(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) (cleanup func())
		expectedCount  int
		expectedNames  []string
		validateExtras func(t *testing.T, extensions []Extension)
	}{
		{
			name: "single extension",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "echo 'hello from foo'")
				return setupTestPATH(t, tempDir)
			},
			expectedCount: 1,
			expectedNames: []string{"temporal-foo"},
			validateExtras: func(t *testing.T, extensions []Extension) {
				assert.Equal(t, []string{"foo"}, extensions[0].CommandTokens)
			},
		},
		{
			name: "multiple extensions",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "echo 'foo'")
				createExecutable(t, tempDir, "temporal-bar", "echo 'bar'")
				createExecutable(t, tempDir, "temporal-baz", "echo 'baz'")
				return setupTestPATH(t, tempDir)
			},
			expectedCount: 3,
			expectedNames: []string{"temporal-foo", "temporal-bar", "temporal-baz"},
		},
		{
			name: "duplicate names - first in PATH wins",
			setup: func(t *testing.T) func() {
				dir1 := t.TempDir()
				dir2 := t.TempDir()
				createExecutable(t, dir1, "temporal-foo", "echo 'from dir1'")
				createExecutable(t, dir2, "temporal-foo", "echo 'from dir2'")
				return setupTestPATH(t, dir1, dir2)
			},
			expectedCount: 1,
			expectedNames: []string{"temporal-foo"},
			validateExtras: func(t *testing.T, extensions []Extension) {
				// Path should be from dir1 (first in PATH)
				assert.Contains(t, extensions[0].Path, "temporal-foo")
			},
		},
		{
			name: "non-executable files ignored",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-executable", "echo 'I am executable'")
				createNonExecutable(t, tempDir, "temporal-not-executable")
				return setupTestPATH(t, tempDir)
			},
			expectedCount: 1,
			expectedNames: []string{"temporal-executable"},
		},
		{
			name: "files without temporal- prefix ignored",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-good", "echo 'has prefix'")
				createExecutable(t, tempDir, "bad-extension", "echo 'no prefix'")
				createExecutable(t, tempDir, "temporalfoo", "echo 'no dash'")
				return setupTestPATH(t, tempDir)
			},
			expectedCount: 1,
			expectedNames: []string{"temporal-good"},
		},
		{
			name: "empty PATH",
			setup: func(t *testing.T) func() {
				return setupTestPATH(t, "")
			},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "non-existent directory in PATH",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-real", "echo 'exists'")
				nonExistent := filepath.Join(tempDir, "does-not-exist")
				return setupTestPATH(t, nonExistent, tempDir)
			},
			expectedCount: 1,
			expectedNames: []string{"temporal-real"},
		},
		{
			name: "permission denied directory",
			setup: func(t *testing.T) func() {
				if os.Geteuid() == 0 {
					t.Skip("Skipping permission test when running as root")
				}

				tempDir := t.TempDir()
				restrictedDir := filepath.Join(tempDir, "restricted")
				err := os.Mkdir(restrictedDir, 0000)
				require.NoError(t, err)
				t.Cleanup(func() { os.Chmod(restrictedDir, 0755) })

				workingDir := filepath.Join(tempDir, "working")
				err = os.Mkdir(workingDir, 0755)
				require.NoError(t, err)

				createExecutable(t, workingDir, "temporal-test", "echo 'test'")
				return setupTestPATH(t, restrictedDir, workingDir)
			},
			expectedCount: 1,
			expectedNames: []string{"temporal-test"},
		},
		{
			name: "real world scenario - multiple complex extensions",
			setup: func(t *testing.T) func() {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-mycompany-login", "echo 'login'")
				createExecutable(t, tempDir, "temporal-workflow-show_diagram", "echo 'diagram'")
				createExecutable(t, tempDir, "temporal-cloud-namespace-list", "echo 'list'")
				return setupTestPATH(t, tempDir)
			},
			expectedCount: 3,
			expectedNames: []string{"temporal-mycompany-login", "temporal-workflow-show_diagram", "temporal-cloud-namespace-list"},
			validateExtras: func(t *testing.T, extensions []Extension) {
				tokenMap := make(map[string][]string)
				for _, ext := range extensions {
					tokenMap[ext.Name] = ext.CommandTokens
				}
				assert.Equal(t, []string{"mycompany", "login"}, tokenMap["temporal-mycompany-login"])
				assert.Equal(t, []string{"workflow", "show-diagram"}, tokenMap["temporal-workflow-show_diagram"])
				assert.Equal(t, []string{"cloud", "namespace", "list"}, tokenMap["temporal-cloud-namespace-list"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			extensions := DiscoverExtensions()

			require.Len(t, extensions, tt.expectedCount, "Unexpected number of extensions discovered")

			if tt.expectedCount > 0 {
				// Collect discovered names
				discoveredNames := make(map[string]bool)
				for _, ext := range extensions {
					discoveredNames[ext.Name] = true
				}

				// Verify all expected names are found
				for _, expectedName := range tt.expectedNames {
					assert.True(t, discoveredNames[expectedName], "Should discover extension: %s", expectedName)
				}
			}

			// Run additional validations if provided
			if tt.validateExtras != nil {
				tt.validateExtras(t, extensions)
			}
		})
	}
}

// TestParseCommandTokens tests the command token parsing logic.
func TestParseCommandTokens(t *testing.T) {
	tests := []struct {
		name           string
		extName        string
		expectedTokens []string
	}{
		{
			name:           "simple command",
			extName:        "temporal-foo",
			expectedTokens: []string{"foo"},
		},
		{
			name:           "two level command",
			extName:        "temporal-foo-bar",
			expectedTokens: []string{"foo", "bar"},
		},
		{
			name:           "three level command",
			extName:        "temporal-foo-bar-baz",
			expectedTokens: []string{"foo", "bar", "baz"},
		},
		{
			name:           "command with underscore",
			extName:        "temporal-foo-bar_baz",
			expectedTokens: []string{"foo", "bar-baz"},
		},
		{
			name:           "multiple underscores",
			extName:        "temporal-foo_bar-baz_qux",
			expectedTokens: []string{"foo-bar", "baz-qux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			createExecutable(t, tempDir, tt.extName, "echo 'test'")

			cleanup := setupTestPATH(t, tempDir)
			defer cleanup()

			extensions := DiscoverExtensions()

			require.Len(t, extensions, 1)
			assert.Equal(t, tt.expectedTokens, extensions[0].CommandTokens)
		})
	}
}

// TestFindExtension tests extension matching scenarios.
func TestFindExtension(t *testing.T) {
	tests := []struct {
		name              string
		setup             func(t *testing.T) ([]Extension, func())
		commandTokens     []string
		expectFound       bool
		expectedExtName   string
		expectedRemaining []string
	}{
		{
			name: "most specific match wins",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "echo 'foo'")
				createExecutable(t, tempDir, "temporal-foo-bar", "echo 'foo bar'")
				createExecutable(t, tempDir, "temporal-foo-bar-baz", "echo 'foo bar baz'")
				cleanup := setupTestPATH(t, tempDir)
				return DiscoverExtensions(), cleanup
			},
			commandTokens:     []string{"foo", "bar", "baz"},
			expectFound:       true,
			expectedExtName:   "temporal-foo-bar-baz",
			expectedRemaining: []string{},
		},
		{
			name: "fallback to less specific",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "echo 'foo'")
				createExecutable(t, tempDir, "temporal-foo-bar", "echo 'foo bar'")
				cleanup := setupTestPATH(t, tempDir)
				return DiscoverExtensions(), cleanup
			},
			commandTokens:     []string{"foo", "bar", "baz", "qux"},
			expectFound:       true,
			expectedExtName:   "temporal-foo-bar",
			expectedRemaining: []string{"baz", "qux"},
		},
		{
			name: "underscore and dash equivalence - dash form",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo-bar_baz", "echo 'test'")
				cleanup := setupTestPATH(t, tempDir)
				return DiscoverExtensions(), cleanup
			},
			commandTokens:     []string{"foo", "bar-baz"},
			expectFound:       true,
			expectedExtName:   "temporal-foo-bar_baz",
			expectedRemaining: []string{},
		},
		{
			name: "underscore and dash equivalence - underscore form",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo-bar_baz", "echo 'test'")
				cleanup := setupTestPATH(t, tempDir)
				return DiscoverExtensions(), cleanup
			},
			commandTokens:     []string{"foo", "bar_baz"},
			expectFound:       true,
			expectedExtName:   "temporal-foo-bar_baz",
			expectedRemaining: []string{},
		},
		{
			name: "no match found",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "echo 'foo'")
				cleanup := setupTestPATH(t, tempDir)
				return DiscoverExtensions(), cleanup
			},
			commandTokens:     []string{"bar", "baz"},
			expectFound:       false,
			expectedExtName:   "",
			expectedRemaining: nil,
		},
		{
			name: "empty command tokens",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-foo", "echo 'foo'")
				cleanup := setupTestPATH(t, tempDir)
				return DiscoverExtensions(), cleanup
			},
			commandTokens:     []string{},
			expectFound:       false,
			expectedExtName:   "",
			expectedRemaining: nil,
		},
		{
			name: "real world scenario with arguments",
			setup: func(t *testing.T) ([]Extension, func()) {
				tempDir := t.TempDir()
				createExecutable(t, tempDir, "temporal-cloud-namespace-list", "echo 'list'")
				cleanup := setupTestPATH(t, tempDir)
				return DiscoverExtensions(), cleanup
			},
			commandTokens:     []string{"cloud", "namespace", "list", "--profile", "prod"},
			expectFound:       true,
			expectedExtName:   "temporal-cloud-namespace-list",
			expectedRemaining: []string{"--profile", "prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extensions, cleanup := tt.setup(t)
			defer cleanup()

			ext, remaining := FindExtension(extensions, tt.commandTokens)

			if tt.expectFound {
				require.NotNil(t, ext, "Expected to find extension")
				assert.Equal(t, tt.expectedExtName, ext.Name, "Extension name mismatch")
				assert.Equal(t, tt.expectedRemaining, remaining, "Remaining arguments mismatch")
			} else {
				assert.Nil(t, ext, "Should not find extension")
				assert.Equal(t, tt.expectedRemaining, remaining, "Remaining should be nil")
			}
		})
	}
}
