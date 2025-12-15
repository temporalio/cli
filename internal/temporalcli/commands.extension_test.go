package temporalcli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/imports"
)

// Go code snippets for cross-platform test extensions.
var (
	codeEchoArgs   = `fmt.Println("Args:", strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe"), strings.Join(os.Args[1:], " "))`
	codeEchoStderr = func(msg string) string {
		return fmt.Sprintf(`fmt.Fprintln(os.Stderr, %q)`, msg)
	}
	codeExit = func(code int) string {
		return fmt.Sprintf(`os.Exit(%d)`, code)
	}
	codeSleep = func(d time.Duration) string {
		return fmt.Sprintf(`time.Sleep(%d)`, d)
	}
	codeCat     = `io.Copy(os.Stdout, os.Stdin)`
	codeEchoEnv = func(name string) string {
		return fmt.Sprintf(`fmt.Println(os.Getenv(%q))`, name)
	}
)

func TestExtension_InvokesSingleLevelExtension(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs)

	res := h.Execute("foo")

	assert.Equal(t, "Args: temporal-foo \n", res.Stdout.String())
}

func TestExtension_InvokesNestedExtension(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo-bar", codeEchoArgs)

	res := h.Execute("foo", "bar")

	assert.Equal(t, "Args: temporal-foo-bar \n", res.Stdout.String())
}

func TestExtension_PrefersMostSpecificExtension(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs)
	h.createExtension("temporal-foo-bar", codeEchoArgs)

	res := h.Execute("foo", "bar")

	assert.Equal(t, "Args: temporal-foo-bar \n", res.Stdout.String())
}

func TestExtension_SkipsFlagsInLookup(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs)
	h.createExtension("temporal-foo-bar", codeEchoArgs)

	res := h.Execute("foo", "-x", "bar")

	// Should find temporal-foo-bar (skipping -x), not temporal-foo.
	assert.Equal(t, "Args: temporal-foo-bar -x\n", res.Stdout.String())
}

func TestExtension_ConvertsDashToUnderscoreInLookup(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo-bar_baz", codeEchoArgs)

	// Dash in arg is converted to underscore when looking up extension.
	res := h.Execute("foo", "bar-baz")
	assert.Equal(t, "Args: temporal-foo-bar_baz \n", res.Stdout.String())

	// Underscore in arg stays as underscore.
	res = h.Execute("foo", "bar_baz")
	assert.Equal(t, "Args: temporal-foo-bar_baz \n", res.Stdout.String())
}

func TestExtension_DoesNotOverrideBuiltinCommand(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-workflow", codeEchoArgs)

	res := h.Execute("workflow", "--help")

	assert.Contains(t, res.Stdout.String(), "Workflow commands perform operations on Workflow Executions")
}

func TestExtension_ExtendsBuiltinWithSubcommand(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-workflow-diagram", codeEchoArgs)

	res := h.Execute("workflow", "diagram")

	assert.Equal(t, "Args: temporal-workflow-diagram \n", res.Stdout.String())
}

func TestExtension_PassesFlagsAfterExtensionCommand(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs)

	res := h.Execute("foo", "bar", "--flag", "value")

	assert.Equal(t, "Args: temporal-foo bar --flag value\n", res.Stdout.String())
}

func TestExtension_PassesStdin(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeCat)
	h.Stdin.WriteString("hello from stdin")

	res := h.Execute("foo")

	assert.Equal(t, "hello from stdin", res.Stdout.String())
}

func TestExtension_InheritsEnvironmentVariables(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoEnv("TEST_EXT_VAR"))
	os.Setenv("TEST_EXT_VAR", "test_value_123")
	t.Cleanup(func() { os.Unsetenv("TEST_EXT_VAR") })

	res := h.Execute("foo")

	assert.Equal(t, "test_value_123\n", res.Stdout.String())
}

func TestExtension_RelaysStderr(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoStderr("stderr output"))

	res := h.Execute("foo")

	assert.Empty(t, res.Stdout.String())
	assert.Equal(t, "stderr output\n", res.Stderr.String())
}

func TestExtension_RelaysStdoutAndStderr(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", `
		fmt.Fprintln(os.Stdout, "stdout line")
		fmt.Fprintln(os.Stderr, "stderr line")
	`)

	res := h.Execute("foo")

	assert.Equal(t, "stdout line\n", res.Stdout.String())
	assert.Equal(t, "stderr line\n", res.Stderr.String())
}

func TestExtension_FailsOnNonExecutableCommand(t *testing.T) {
	h := newExtensionHarness(t)
	// Create file without execute permission.
	path := filepath.Join(h.binDir, "temporal-foo")
	err := os.WriteFile(path, []byte("a text file"), 0644)
	require.NoError(t, err)

	res := h.Execute("foo")

	assert.Contains(t, res.Stdout.String(), "Usage:") // help text is shown
	assert.EqualError(t, res.Err, "unknown command")
}

func TestExtension_PassesThroughNonZeroExit(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeEchoArgs, codeExit(42))

	res := h.Execute("foo")

	assert.Equal(t, "Args: temporal-foo \n", res.Stdout.String())
	assert.NoError(t, res.Err)
}

func TestExtension_FailsOnCommandTimeout(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeSleep(10*time.Second))

	res := h.Execute("foo", "--command-timeout", "100ms")

	assert.EqualError(t, res.Err, "program interrupted")
}

func TestExtension_FailsOnCommandCancellation(t *testing.T) {
	h := newExtensionHarness(t)
	h.createExtension("temporal-foo", codeSleep(10*time.Second))
	go func() {
		time.Sleep(100 * time.Millisecond)
		h.CancelContext()
	}()

	res := h.Execute("foo")

	assert.EqualError(t, res.Err, "program interrupted")
}

type extensionHarness struct {
	*CommandHarness
	binDir string
}

func newExtensionHarness(t *testing.T) *extensionHarness {
	t.Helper()

	binDir := t.TempDir()
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)
	t.Cleanup(func() { os.Setenv("PATH", oldPath) })

	return &extensionHarness{
		CommandHarness: NewCommandHarness(t),
		binDir:         binDir,
	}
}

func (h *extensionHarness) createExtension(name string, code ...string) string {
	h.t.Helper()

	// Wrap code in main function.
	source := fmt.Sprintf("package main\n\nfunc main() {\n%s\n}\n", strings.Join(code, "\n"))

	// Run goimports to resolve imports.
	formatted, err := imports.Process("main.go", []byte(source), nil)
	require.NoError(h.t, err, "Failed to process imports for %s:\n%s", name, source)

	// Write source file.
	srcPath := filepath.Join(h.binDir, name+".go")
	require.NoError(h.t, os.WriteFile(srcPath, formatted, 0644))

	// Build executable.
	binPath := filepath.Join(h.binDir, name)
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	output, err := cmd.CombinedOutput()
	require.NoError(h.t, err, "Failed to compile %s: %s\nSource:\n%s", name, output, formatted)

	return binPath
}
