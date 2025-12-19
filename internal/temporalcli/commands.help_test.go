package temporalcli_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelp_Root(t *testing.T) {
	h := NewCommandHarness(t)

	res := h.Execute("help")

	assert.Contains(t, res.Stdout.String(), "Available Commands:")
	assert.Contains(t, res.Stdout.String(), "workflow")
	assert.NoError(t, res.Err)
}

func TestHelp_Subcommand(t *testing.T) {
	h := NewCommandHarness(t)

	res := h.Execute("help", "workflow")

	assert.Contains(t, res.Stdout.String(), "Workflow commands")
	assert.NoError(t, res.Err)
}

func TestHelp_HelpShowsAllFlag(t *testing.T) {
	h := NewCommandHarness(t)
	res := h.Execute("help", "--help")

	assert.Contains(t, res.Stdout.String(), "-a, --all")
	assert.Contains(t, res.Stdout.String(), "extensions found in PATH")
}

func TestHelp_AllFlag_ShowsExtensions(t *testing.T) {
	h := newExtensionHarness(t)
	fooPath := h.createExtension("temporal-foo", codeEchoArgs)
	fooBarPath := h.createExtension("temporal-foo-bar", codeEchoArgs)
	h.createExtension("temporal-workflow-bar_baz", codeEchoArgs)

	// Without --all, no extensions are shown
	res := h.Execute("help")
	assert.NotContains(t, res.Stdout.String(), "foo")
	assert.NotContains(t, res.Stdout.String(), "bar-baz")
	assert.NoError(t, res.Err)

	// With --all, extensions on root level are shown in Available Commands (not Additional help topics)
	res = h.Execute("help", "--all")
	out := res.Stdout.String()
	assert.Contains(t, out, "foo")        // shown now!
	assert.NotContains(t, out, "bar-baz") // is under workflow

	// Verify foo appears in Available Commands section (between "Available Commands:" and "Flags:")
	availableIdx := strings.Index(out, "Available Commands:")
	fooIdx := strings.Index(out, "foo")
	flagsIdx := strings.Index(out, "Flags:")
	assert.Greater(t, fooIdx, availableIdx, "foo should appear after Available Commands:")
	assert.Less(t, fooIdx, flagsIdx, "foo should appear before Flags:")
	assert.NoError(t, res.Err)

	// Non-executable extensions are skipped
	// On Unix, remove executable permission; on Windows, rename to .bak extension
	if runtime.GOOS == "windows" {
		require.NoError(t, os.Rename(fooPath, fooPath+".bak"))
		require.NoError(t, os.Rename(fooBarPath, fooBarPath+".bak"))
	} else {
		require.NoError(t, os.Chmod(fooPath, 0644))
		require.NoError(t, os.Chmod(fooBarPath, 0644))
	}
	res = h.Execute("help", "--all")
	assert.NotContains(t, res.Stdout.String(), "foo")
	assert.NoError(t, res.Err)

	// With --all on built-in subcommand, shows nested extensions
	res = h.Execute("help", "workflow", "--all")
	assert.Contains(t, res.Stdout.String(), "bar-baz")
	assert.NoError(t, res.Err)
}

func TestHelp_AllFlag_FirstInPathWins(t *testing.T) {
	h := newExtensionHarness(t)
	binDir1 := h.binDir
	binDir2 := t.TempDir()

	// Set PATH with binDir1 before binDir2
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir1+string(os.PathListSeparator)+binDir2+string(os.PathListSeparator)+oldPath)
	t.Cleanup(func() { os.Setenv("PATH", oldPath) })

	// Create extension in binDir1 that outputs "first"
	h.createExtension("temporal-foo", `fmt.Println("first")`)

	// Create extension in binDir2 that outputs "second"
	h.binDir = binDir2
	h.createExtension("temporal-foo", `fmt.Println("second")`)

	// Should use the first one found in PATH
	res := h.Execute("foo")
	assert.Equal(t, "first\n", res.Stdout.String())
	assert.NoError(t, res.Err)
}
